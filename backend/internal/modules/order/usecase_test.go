package order_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

// fakeRepo records the last ListQuery it received (already normalized by
// the usecase) and returns a fixed page or error, so tests can assert on
// exactly what the usecase delegates.
type fakeRepo struct {
	gotQuery order.ListQuery
	page     order.Page
	err      error

	gotParticipant   order.Participant
	gotParticipantID uuid.UUID
	gotOrderID       uuid.UUID
	orderDetailData  *order.OrderDetailData
	orderErr         error
	getOrderCalled   bool

	confirmArtistID uuid.UUID
	confirmOrderID  uuid.UUID
	confirmStatus   order.Status
	confirmErr      error
	confirmCalled   bool

	cancelParticipant   order.Participant
	cancelParticipantID uuid.UUID
	cancelOrderID       uuid.UUID
	cancelErr           error
	cancelCalled        bool

	expireCalled        bool
	expireNow           time.Time
	expirePendingCutoff time.Time
	expireIDs           []uuid.UUID
	expireErr           error

	created              *order.Order
	createdDeliverable   *order.Deliverable
	deliverableArtistID  uuid.UUID
	deliverableOrderID   uuid.UUID
	deliverableCreateErr error

	validateDeliverableArtistID uuid.UUID
	validateDeliverableOrderID  uuid.UUID
	validateDeliverableErr      error
	validateDeliverableCalled   bool

	delayRequestArtistID uuid.UUID
	delayRequestOrderID  uuid.UUID
	delayRequest         *order.OrderDelayRequest
	delayRequestErr      error
	delayRequestCalled   bool
}

func (f *fakeRepo) ListOrders(_ context.Context, query order.ListQuery) (order.Page, error) {
	f.gotQuery = query
	return f.page, f.err
}

func (f *fakeRepo) GetOrderByID(
	_ context.Context,
	participant order.Participant,
	participantID uuid.UUID,
	orderID uuid.UUID,
) (*order.OrderDetailData, error) {
	f.getOrderCalled = true
	f.gotParticipant = participant
	f.gotParticipantID = participantID
	f.gotOrderID = orderID

	return f.orderDetailData, f.orderErr
}

func (f *fakeRepo) ConfirmOrder(
	_ context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
	status order.Status,
) error {
	f.confirmCalled = true
	f.confirmArtistID = artistID
	f.confirmOrderID = orderID
	f.confirmStatus = status

	return f.confirmErr
}

func (f *fakeRepo) CancelOrder(
	_ context.Context,
	participant order.Participant,
	participantID uuid.UUID,
	orderID uuid.UUID,
) error {
	f.cancelCalled = true
	f.cancelParticipant = participant
	f.cancelParticipantID = participantID
	f.cancelOrderID = orderID

	return f.cancelErr
}

func (f *fakeRepo) CancelExpired(_ context.Context, pendingCutoff, now time.Time) ([]uuid.UUID, error) {
	f.expireCalled = true
	f.expireNow = now
	f.expirePendingCutoff = pendingCutoff
	return f.expireIDs, f.expireErr
}

func (repo *fakeRepo) Create(ctx context.Context, item *order.Order) error {
	repo.created = item
	return nil
}

func (repo *fakeRepo) CreateDeliverable(
	_ context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
	item *order.Deliverable,
) (*order.Deliverable, error) {
	repo.createdDeliverable = item
	repo.deliverableArtistID = artistID
	repo.deliverableOrderID = orderID
	if repo.deliverableCreateErr != nil {
		return nil, repo.deliverableCreateErr
	}

	created := *item
	created.Version = 1
	return &created, nil
}

func (repo *fakeRepo) ValidateOrderForDeliverable(
	_ context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
) error {
	repo.validateDeliverableCalled = true
	repo.validateDeliverableArtistID = artistID
	repo.validateDeliverableOrderID = orderID
	return repo.validateDeliverableErr
}

func (repo *fakeRepo) CreateDelayRequest(
	_ context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
	item *order.OrderDelayRequest,
) (*order.OrderDelayRequest, error) {
	repo.delayRequestCalled = true
	repo.delayRequestArtistID = artistID
	repo.delayRequestOrderID = orderID
	repo.delayRequest = item
	if repo.delayRequestErr != nil {
		return nil, repo.delayRequestErr
	}

	created := *item
	return &created, nil
}

var _ order.OrderRepository = (*fakeRepo)(nil)

// fakeStorage returns url+key for any GetPresignedURL call, or err if set —
// enough to prove ViewOrders resolves keys through the ObjectStorage port
// without a real S3 client.
type fakeStorage struct {
	url string
	err error
}

func (f fakeStorage) GetPresignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.url + key, nil
}

func (fakeStorage) Upload(_ context.Context, _ string, _ io.Reader, _ string) error {
	return nil
}

func (fakeStorage) Delete(_ context.Context, _ string) error {
	return nil
}

var _ order.ObjectStorage = fakeStorage{}

type storedObject struct {
	key         string
	contentType string
	content     []byte
}

type recordingStorage struct {
	uploads       []storedObject
	deletes       []string
	presignedKeys []string
	failUploadAt  int
	uploadErr     error
	presignErr    error
	deleteErr     error
}

func (storage *recordingStorage) GetPresignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	storage.presignedKeys = append(storage.presignedKeys, key)
	if storage.presignErr != nil {
		return "", storage.presignErr
	}
	return key, nil
}

func (storage *recordingStorage) Upload(_ context.Context, key string, body io.Reader, contentType string) error {
	content, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	storage.uploads = append(storage.uploads, storedObject{key: key, contentType: contentType, content: content})
	if storage.failUploadAt == len(storage.uploads) {
		return storage.uploadErr
	}
	return nil
}

func (storage *recordingStorage) Delete(_ context.Context, key string) error {
	storage.deletes = append(storage.deletes, key)
	return storage.deleteErr
}

var _ order.ObjectStorage = (*recordingStorage)(nil)

type fakeArtistGetter struct {
	name string
	err  error
}

func (getter fakeArtistGetter) GetArtistName(context.Context, uuid.UUID) (string, error) {
	return getter.name, getter.err
}

var _ order.ArtistGetter = fakeArtistGetter{}

func newTestOrderUsecase(repo order.OrderRepository, storage order.ObjectStorage) order.OrderUsecase {
	return order.NewOrderUsecase(repo, storage, fakeArtistGetter{})
}

func testDeliverablePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 640, 320))
	for y := 0; y < 320; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test image: %v", err)
	}
	return buf.Bytes()
}

func validQuery() order.ListQuery {
	return order.ListQuery{
		Participant:   order.ParticipantCustomer,
		ParticipantID: uuid.New(),
	}
}

func wantForbidden(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.Error, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeForbidden {
		t.Errorf("Code = %v, want %v", appErr.Code, apperror.CodeForbidden)
	}
}

func wantInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.Error, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeInvalidInput {
		t.Errorf("Code = %v, want %v", appErr.Code, apperror.CodeInvalidInput)
	}
}

func TestViewOrders_DefaultsSortOrderAndLimit(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	if _, err := usecase.ViewOrders(context.Background(), validQuery()); err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}

	if repo.gotQuery.Sort != order.DefaultSort {
		t.Errorf("Sort = %q, want default %q", repo.gotQuery.Sort, order.DefaultSort)
	}
	if repo.gotQuery.Order != order.DefaultOrder {
		t.Errorf("Order = %q, want default %q", repo.gotQuery.Order, order.DefaultOrder)
	}
	if repo.gotQuery.Limit != order.DefaultLimit {
		t.Errorf("Limit = %d, want default %d", repo.gotQuery.Limit, order.DefaultLimit)
	}
}

func TestViewOrders_ArtistParticipant_PassesThrough(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Participant = order.ParticipantArtist
	if _, err := usecase.ViewOrders(context.Background(), q); err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}

	if repo.gotQuery.Participant != order.ParticipantArtist {
		t.Errorf("Participant = %q, want %q", repo.gotQuery.Participant, order.ParticipantArtist)
	}
	if repo.gotQuery.ParticipantID != q.ParticipantID {
		t.Errorf("ParticipantID = %v, want %v", repo.gotQuery.ParticipantID, q.ParticipantID)
	}
}

func TestViewOrders_RejectsUnsupportedParticipant(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Participant = "admin"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantForbidden(t, err)
}

func TestViewOrders_RejectsMissingParticipantID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.ParticipantID = uuid.Nil
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidSort(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Sort = "created_at"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidOrder(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Order = "sideways"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsLimitOutOfRange(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	for _, limit := range []int{-1, order.MaxLimit + 1} {
		q := validQuery()
		q.Limit = limit
		_, err := usecase.ViewOrders(context.Background(), q)
		wantInvalidInput(t, err)
	}
}

func TestViewOrders_RejectsNegativeOffset(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Offset = -1
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidStatus(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Statuses = []order.Status{"NOT_A_STATUS"}
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_NormalizesStatuses_DedupesAndSorts(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Statuses = []order.Status{order.StatusSuccess, order.StatusPending, order.StatusPending}
	if _, err := usecase.ViewOrders(context.Background(), q); err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}

	want := []order.Status{order.StatusPending, order.StatusSuccess}
	got := repo.gotQuery.Statuses
	if len(got) != len(want) {
		t.Fatalf("Statuses = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Statuses[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestViewOrders_EmptyPageIsNotAnError(t *testing.T) {
	repo := &fakeRepo{page: order.Page{Orders: []order.Order{}}}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	page, err := usecase.ViewOrders(context.Background(), validQuery())
	if err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil for an empty page", err)
	}
	if len(page.Orders) != 0 {
		t.Errorf("Orders = %+v, want empty", page.Orders)
	}
}

func TestViewOrders_PropagatesRepositoryError(t *testing.T) {
	wantErr := apperror.Internal("boom", nil)
	repo := &fakeRepo{err: wantErr}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	_, err := usecase.ViewOrders(context.Background(), validQuery())
	if !errors.Is(err, wantErr) {
		t.Errorf("ViewOrders() error = %v, want %v", err, wantErr)
	}
}

func TestViewOrders_ResolvesDeliverablePreviewKeyToURL(t *testing.T) {
	key := "orders/abc/v2/preview.png"
	withKey := order.Order{ID: uuid.New(), DeliverablePreviewKey: &key}
	withoutKey := order.Order{ID: uuid.New()}
	repo := &fakeRepo{page: order.Page{Orders: []order.Order{withKey, withoutKey}}}
	usecase := newTestOrderUsecase(repo, fakeStorage{url: "https://cdn.test/"})

	page, err := usecase.ViewOrders(context.Background(), validQuery())
	if err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}
	want := "https://cdn.test/" + key
	if page.Orders[0].DeliverablePreviewURL == nil || *page.Orders[0].DeliverablePreviewURL != want {
		t.Errorf("Orders[0].DeliverablePreviewURL = %v, want %q", page.Orders[0].DeliverablePreviewURL, want)
	}
	if page.Orders[1].DeliverablePreviewURL != nil {
		t.Errorf("Orders[1].DeliverablePreviewURL = %v, want nil", *page.Orders[1].DeliverablePreviewURL)
	}
}

func TestViewOrders_PropagatesPresignError(t *testing.T) {
	key := "orders/abc/v1/preview.png"
	repo := &fakeRepo{page: order.Page{Orders: []order.Order{{ID: uuid.New(), DeliverablePreviewKey: &key}}}}
	wantErr := errors.New("s3 unreachable")
	usecase := newTestOrderUsecase(repo, fakeStorage{err: wantErr})

	_, err := usecase.ViewOrders(context.Background(), validQuery())
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("ViewOrders() error = %v, want *apperror.Error{Code: CodeInternal}", err)
	}
}

func TestGetOrder_RejectsUnsupportedParticipant(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	_, err := usecase.GetOrder(
		context.Background(),
		order.Participant("admin"),
		uuid.New(),
		uuid.New(),
	)

	wantForbidden(t, err)

	if repo.getOrderCalled {
		t.Error("GetOrderByID() should not be called for unsupported participant")
	}
}

func TestGetOrder_RejectsMissingParticipantID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	_, err := usecase.GetOrder(
		context.Background(),
		order.ParticipantCustomer,
		uuid.Nil,
		uuid.New(),
	)

	wantInvalidInput(t, err)

	if repo.getOrderCalled {
		t.Error("GetOrderByID() should not be called for missing participant ID")
	}
}

func TestGetOrder_RejectsMissingOrderID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	_, err := usecase.GetOrder(
		context.Background(),
		order.ParticipantCustomer,
		uuid.New(),
		uuid.Nil,
	)

	wantInvalidInput(t, err)

	if repo.getOrderCalled {
		t.Error("GetOrderByID() should not be called for missing order ID")
	}
}

func TestGetOrder_CustomerGetsArtistAsOtherParty(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	artistID := uuid.New()

	artistScore := 4.8

	wantOrder := order.Order{
		ID:         orderID,
		CustomerID: customerID,
		ArtistID:   artistID,
		Name:       "Test commission",
		Status:     order.StatusPending,
	}

	customer := order.OrderParty{
		ID:    customerID,
		Name:  "Customer",
		Email: "customer@example.com",
	}

	artist := order.OrderParty{
		ID:                artistID,
		Name:              "Artist",
		Email:             "artist@example.com",
		ArtistReviewScore: &artistScore,
	}

	repo := &fakeRepo{
		orderDetailData: &order.OrderDetailData{
			Order:    wantOrder,
			Customer: customer,
			Artist:   artist,
		},
	}

	usecase := newTestOrderUsecase(repo, fakeStorage{})

	got, err := usecase.GetOrder(
		context.Background(),
		order.ParticipantCustomer,
		customerID,
		orderID,
	)
	if err != nil {
		t.Fatalf("GetOrder() error = %v, want nil", err)
	}

	if repo.gotParticipant != order.ParticipantCustomer {
		t.Errorf(
			"Participant = %q, want %q",
			repo.gotParticipant,
			order.ParticipantCustomer,
		)
	}

	if repo.gotParticipantID != customerID {
		t.Errorf(
			"ParticipantID = %v, want %v",
			repo.gotParticipantID,
			customerID,
		)
	}

	if repo.gotOrderID != orderID {
		t.Errorf(
			"OrderID = %v, want %v",
			repo.gotOrderID,
			orderID,
		)
	}

	if got.Order.ID != orderID {
		t.Errorf("Order.ID = %v, want %v", got.Order.ID, orderID)
	}

	if got.OtherParty.ID != artistID {
		t.Errorf(
			"OtherParty.ID = %v, want %v",
			got.OtherParty.ID,
			artistID,
		)
	}

	if got.OtherParty.Name != artist.Name {
		t.Errorf(
			"OtherParty.Name = %q, want %q",
			got.OtherParty.Name,
			artist.Name,
		)
	}

	if got.OtherParty.Email != artist.Email {
		t.Errorf(
			"OtherParty.Email = %q, want %q",
			got.OtherParty.Email,
			artist.Email,
		)
	}

	if got.OtherParty.ArtistReviewScore == nil {
		t.Fatal("OtherParty.ArtistReviewScore = nil, want artist score")
	}

	if *got.OtherParty.ArtistReviewScore != artistScore {
		t.Errorf(
			"OtherParty.ArtistReviewScore = %v, want %v",
			*got.OtherParty.ArtistReviewScore,
			artistScore,
		)
	}
}

func TestGetOrder_ArtistGetsCustomerAsOtherParty(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	artistID := uuid.New()

	wantOrder := order.Order{
		ID:         orderID,
		CustomerID: customerID,
		ArtistID:   artistID,
		Name:       "Test commission",
		Status:     order.StatusPending,
	}

	customer := order.OrderParty{
		ID:    customerID,
		Name:  "Customer",
		Email: "customer@example.com",
	}

	artist := order.OrderParty{
		ID:    artistID,
		Name:  "Artist",
		Email: "artist@example.com",
	}

	repo := &fakeRepo{
		orderDetailData: &order.OrderDetailData{
			Order:    wantOrder,
			Customer: customer,
			Artist:   artist,
		},
	}

	usecase := newTestOrderUsecase(repo, fakeStorage{})

	got, err := usecase.GetOrder(
		context.Background(),
		order.ParticipantArtist,
		artistID,
		orderID,
	)
	if err != nil {
		t.Fatalf("GetOrder() error = %v, want nil", err)
	}

	if got.OtherParty.ID != customerID {
		t.Errorf(
			"OtherParty.ID = %v, want %v",
			got.OtherParty.ID,
			customerID,
		)
	}

	if got.OtherParty.Name != customer.Name {
		t.Errorf(
			"OtherParty.Name = %q, want %q",
			got.OtherParty.Name,
			customer.Name,
		)
	}

	if got.OtherParty.Email != customer.Email {
		t.Errorf(
			"OtherParty.Email = %q, want %q",
			got.OtherParty.Email,
			customer.Email,
		)
	}

	if got.OtherParty.ArtistReviewScore != nil {
		t.Errorf(
			"OtherParty.ArtistReviewScore = %v, want nil for customer",
			*got.OtherParty.ArtistReviewScore,
		)
	}
}

func TestGetOrder_PropagatesRepositoryError(t *testing.T) {
	wantErr := apperror.Internal("boom", nil)

	repo := &fakeRepo{
		orderErr: wantErr,
	}

	usecase := newTestOrderUsecase(repo, fakeStorage{})

	orderID := uuid.New()
	participantID := uuid.New()

	_, err := usecase.GetOrder(
		context.Background(),
		order.ParticipantCustomer,
		participantID,
		orderID,
	)
	if !errors.Is(err, wantErr) {
		t.Errorf("GetOrder() error = %v, want %v", err, wantErr)
	}
}

func TestGetOrder_RejectsNilRepositoryResult(t *testing.T) {
	usecase := newTestOrderUsecase(&fakeRepo{}, fakeStorage{})

	_, err := usecase.GetOrder(
		context.Background(),
		order.ParticipantCustomer,
		uuid.New(),
		uuid.New(),
	)

	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("GetOrder() error = %v, want internal application error", err)
	}
}

func TestConfirmOrder_AcceptsOrder(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	artistID := uuid.New()
	orderID := uuid.New()

	gotStatus, err := usecase.ConfirmOrder(
		context.Background(),
		artistID,
		orderID,
		order.ConfirmOrderInput{
			Accept: true,
		},
	)
	if err != nil {
		t.Fatalf("ConfirmOrder() error = %v, want nil", err)
	}
	if gotStatus != order.StatusNotPaid {
		t.Errorf("ConfirmOrder() status = %q, want %q", gotStatus, order.StatusNotPaid)
	}

	if !repo.confirmCalled {
		t.Fatalf("ConfirmOrder() repository method was not called")
	}

	if repo.confirmArtistID != artistID {
		t.Errorf("ArtistID = %v, want %v", repo.confirmArtistID, artistID)
	}

	if repo.confirmOrderID != orderID {
		t.Errorf("OrderID = %v, want %v", repo.confirmOrderID, orderID)
	}

	if repo.confirmStatus != order.StatusNotPaid {
		t.Errorf(
			"Status = %q, want %q",
			repo.confirmStatus,
			order.StatusNotPaid,
		)
	}
}

func TestConfirmOrder_RejectsOrder(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	artistID := uuid.New()
	orderID := uuid.New()

	gotStatus, err := usecase.ConfirmOrder(
		context.Background(),
		artistID,
		orderID,
		order.ConfirmOrderInput{
			Accept: false,
		},
	)
	if err != nil {
		t.Fatalf("ConfirmOrder() error = %v, want nil", err)
	}
	if gotStatus != order.StatusCancel {
		t.Errorf("ConfirmOrder() status = %q, want %q", gotStatus, order.StatusCancel)
	}

	if !repo.confirmCalled {
		t.Fatalf("ConfirmOrder() repository method was not called")
	}

	if repo.confirmArtistID != artistID {
		t.Errorf("ArtistID = %v, want %v", repo.confirmArtistID, artistID)
	}

	if repo.confirmOrderID != orderID {
		t.Errorf("OrderID = %v, want %v", repo.confirmOrderID, orderID)
	}

	if repo.confirmStatus != order.StatusCancel {
		t.Errorf(
			"Status = %q, want %q",
			repo.confirmStatus,
			order.StatusCancel,
		)
	}
}

func TestConfirmOrder_RejectsMissingArtistID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	gotStatus, err := usecase.ConfirmOrder(
		context.Background(),
		uuid.Nil,
		uuid.New(),
		order.ConfirmOrderInput{
			Accept: true,
		},
	)

	wantInvalidInput(t, err)
	if gotStatus != "" {
		t.Errorf("ConfirmOrder() status = %q, want empty status", gotStatus)
	}

	if repo.confirmCalled {
		t.Error("ConfirmOrder() repository method should not be called")
	}
}

func TestConfirmOrder_RejectsMissingOrderID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	gotStatus, err := usecase.ConfirmOrder(
		context.Background(),
		uuid.New(),
		uuid.Nil,
		order.ConfirmOrderInput{Accept: true},
	)

	wantInvalidInput(t, err)
	if gotStatus != "" {
		t.Errorf("ConfirmOrder() status = %q, want empty status", gotStatus)
	}

	if repo.confirmCalled {
		t.Error("ConfirmOrder() repository method should not be called")
	}
}

func TestConfirmOrder_PropagatesRepositoryError(t *testing.T) {
	wantErr := apperror.Internal("boom", nil)

	repo := &fakeRepo{
		confirmErr: wantErr,
	}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	gotStatus, err := usecase.ConfirmOrder(
		context.Background(),
		uuid.New(),
		uuid.New(),
		order.ConfirmOrderInput{Accept: true},
	)

	if gotStatus != "" {
		t.Errorf("ConfirmOrder() status = %q, want empty status", gotStatus)
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("ConfirmOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCreateDeliverableUploadsOriginalAndWatermarkedPreview(t *testing.T) {
	repo := &fakeRepo{}
	storage := &recordingStorage{}
	usecase := order.NewOrderUsecase(repo, storage, fakeArtistGetter{name: "Test Artist"})
	artistID, orderID := uuid.New(), uuid.New()
	original := testDeliverablePNG(t)

	created, err := usecase.CreateDeliverable(context.Background(), artistID, order.CreateDeliverableInput{
		OrderID:          orderID,
		DeliverableImage: bytes.NewReader(original),
	})
	if err != nil {
		t.Fatalf("CreateDeliverable() error = %v", err)
	}

	if len(storage.uploads) != 2 {
		t.Fatalf("uploads = %d, want 2", len(storage.uploads))
	}
	if !bytes.Equal(storage.uploads[0].content, original) || storage.uploads[0].contentType != "image/png" {
		t.Errorf("original upload = (%s, %q), want original PNG", storage.uploads[0].key, storage.uploads[0].contentType)
	}
	if storage.uploads[1].contentType != "image/jpeg" {
		t.Errorf("preview content type = %q, want image/jpeg", storage.uploads[1].contentType)
	}
	if !bytes.HasPrefix(storage.uploads[1].content, []byte{0xff, 0xd8, 0xff}) {
		t.Fatal("preview upload is not a JPEG")
	}
	preview, format, err := image.Decode(bytes.NewReader(storage.uploads[1].content))
	if err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if format != "jpeg" || preview.Bounds().Dx() != orderPreviewWidth || preview.Bounds().Dy() != orderPreviewHeight {
		t.Errorf("preview dimensions/format = %dx%d %q", preview.Bounds().Dx(), preview.Bounds().Dy(), format)
	}
	if created.OriginalImageKey != storage.uploads[0].key || created.PreviewImageKey != storage.uploads[1].key {
		t.Errorf("deliverable keys = (%q, %q), want uploaded keys", created.OriginalImageKey, created.PreviewImageKey)
	}
	if repo.deliverableArtistID != artistID || repo.deliverableOrderID != orderID {
		t.Error("repository did not receive the created deliverable with the authenticated artist and order IDs")
	}
	if !repo.validateDeliverableCalled ||
		repo.validateDeliverableArtistID != artistID ||
		repo.validateDeliverableOrderID != orderID {
		t.Error("order eligibility was not validated for the authenticated artist and order")
	}
	if repo.createdDeliverable == nil || repo.createdDeliverable.ID != created.ID {
		t.Error("repository did not receive the deliverable returned by CreateDeliverable")
	}
	if created.Version != 1 {
		t.Errorf("created version = %d, want repository-assigned version 1", created.Version)
	}
	if created.PreviewImageURL != created.PreviewImageKey {
		t.Errorf("preview URL = %q, want presigned URL for key %q", created.PreviewImageURL, created.PreviewImageKey)
	}
	if len(storage.presignedKeys) != 1 || storage.presignedKeys[0] != created.PreviewImageKey {
		t.Errorf("presigned keys = %v, want [%q]", storage.presignedKeys, created.PreviewImageKey)
	}
	if len(storage.deletes) != 0 {
		t.Errorf("deletes = %v, want none", storage.deletes)
	}
}

func TestCreateDeliverableRejectsInvalidOrOversizedImageBeforeUpload(t *testing.T) {
	tests := []struct {
		name  string
		image []byte
		want  error
	}{
		{name: "invalid image", image: []byte("not an image"), want: order.ErrInvalidDeliverableImage},
		{name: "oversized image", image: append(testDeliverablePNG(t), make([]byte, order.MaxDeliverableImageSize)...), want: order.ErrDeliverableImageTooLarge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := &recordingStorage{}
			usecase := order.NewOrderUsecase(&fakeRepo{}, storage, fakeArtistGetter{name: "Test Artist"})

			_, err := usecase.CreateDeliverable(context.Background(), uuid.New(), order.CreateDeliverableInput{
				OrderID:          uuid.New(),
				DeliverableImage: bytes.NewReader(test.image),
			})
			if !errors.Is(err, test.want) {
				t.Errorf("CreateDeliverable() error = %v, want %v", err, test.want)
			}
			if len(storage.uploads) != 0 {
				t.Errorf("uploads = %d, want 0", len(storage.uploads))
			}
		})
	}
}

func TestCreateDeliverableReturnsValidationErrorBeforeUpload(t *testing.T) {
	wantErr := order.ErrInvalidOrderStatus
	repo := &fakeRepo{validateDeliverableErr: wantErr}
	storage := &recordingStorage{}
	usecase := order.NewOrderUsecase(repo, storage, fakeArtistGetter{name: "Test Artist"})

	_, err := usecase.CreateDeliverable(context.Background(), uuid.New(), order.CreateDeliverableInput{
		OrderID:          uuid.New(),
		DeliverableImage: bytes.NewReader(testDeliverablePNG(t)),
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("CreateDeliverable() error = %v, want %v", err, wantErr)
	}
	if !repo.validateDeliverableCalled {
		t.Error("ValidateOrderForDeliverable() was not called")
	}
	if len(storage.uploads) != 0 {
		t.Errorf("uploads = %d, want 0 when the order is not eligible", len(storage.uploads))
	}
}

func TestCreateDeliverableCleansUpWhenPreviewUploadFails(t *testing.T) {
	storage := &recordingStorage{failUploadAt: 2, uploadErr: errors.New("storage unavailable")}
	usecase := order.NewOrderUsecase(&fakeRepo{}, storage, fakeArtistGetter{name: "Test Artist"})

	_, err := usecase.CreateDeliverable(context.Background(), uuid.New(), order.CreateDeliverableInput{
		OrderID:          uuid.New(),
		DeliverableImage: bytes.NewReader(testDeliverablePNG(t)),
	})
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("CreateDeliverable() error = %v, want internal application error", err)
	}
	if len(storage.deletes) != 1 || storage.deletes[0] != storage.uploads[0].key {
		t.Errorf("deletes = %v, want cleanup of uploaded original %q", storage.deletes, storage.uploads[0].key)
	}
}

func TestCreateDeliverableCleansUpWhenPersistenceFails(t *testing.T) {
	repoErr := apperror.Internal("database unavailable", nil)
	repo := &fakeRepo{deliverableCreateErr: repoErr}
	storage := &recordingStorage{}
	usecase := order.NewOrderUsecase(repo, storage, fakeArtistGetter{name: "Test Artist"})

	_, err := usecase.CreateDeliverable(context.Background(), uuid.New(), order.CreateDeliverableInput{
		OrderID:          uuid.New(),
		DeliverableImage: bytes.NewReader(testDeliverablePNG(t)),
	})
	if !errors.Is(err, repoErr) {
		t.Errorf("CreateDeliverable() error = %v, want %v", err, repoErr)
	}
	if len(storage.deletes) != 2 {
		t.Errorf("deletes = %v, want both uploaded images removed", storage.deletes)
	}
}

func TestCreateDeliverablePropagatesPreviewPresignFailure(t *testing.T) {
	storage := &recordingStorage{presignErr: errors.New("storage unavailable")}
	usecase := order.NewOrderUsecase(&fakeRepo{}, storage, fakeArtistGetter{name: "Test Artist"})

	created, err := usecase.CreateDeliverable(context.Background(), uuid.New(), order.CreateDeliverableInput{
		OrderID:          uuid.New(),
		DeliverableImage: bytes.NewReader(testDeliverablePNG(t)),
	})
	var appErr *apperror.Error
	if created != nil {
		t.Errorf("CreateDeliverable() result = %+v, want nil on presign failure", created)
	}
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("CreateDeliverable() error = %v, want internal application error", err)
	}
	if len(storage.presignedKeys) != 1 || storage.presignedKeys[0] != storage.uploads[1].key {
		t.Errorf("presigned keys = %v, want the persisted preview key %q", storage.presignedKeys, storage.uploads[1].key)
	}
	if len(storage.deletes) != 0 {
		t.Errorf("deletes = %v, want none after persistence succeeded", storage.deletes)
	}
}

func TestCreateDelayRequestPersistsAndReturnsInitializedRequest(t *testing.T) {
	artistID, orderID := uuid.New(), uuid.New()
	requestedDeadline := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	created, err := usecase.CreateDelayRequest(context.Background(), artistID, order.CreateDelayRequestInput{
		OrderID:           orderID,
		DeadlineRequestAt: requestedDeadline,
		Reason:            order.DelayReasonNeedMoreTime,
		Explanation:       "More time is needed to finish the order.",
	})
	if err != nil {
		t.Fatalf("CreateDelayRequest() error = %v, want nil", err)
	}
	if !repo.delayRequestCalled {
		t.Fatal("CreateDelayRequest() repository method was not called")
	}
	if repo.delayRequestArtistID != artistID || repo.delayRequestOrderID != orderID {
		t.Errorf(
			"repository received artistID=%s orderID=%s, want artistID=%s orderID=%s",
			repo.delayRequestArtistID,
			repo.delayRequestOrderID,
			artistID,
			orderID,
		)
	}
	if repo.delayRequest == nil {
		t.Fatal("repository received nil delay request")
	}
	if repo.delayRequest.ID == uuid.Nil ||
		repo.delayRequest.OrderID != orderID ||
		!repo.delayRequest.DeadlineRequestAt.Equal(requestedDeadline) ||
		repo.delayRequest.Reason != order.DelayReasonNeedMoreTime ||
		repo.delayRequest.Explanation != "More time is needed to finish the order." ||
		repo.delayRequest.CustomerDecision != order.DelayDecisionWait ||
		repo.delayRequest.CreatedAt.IsZero() ||
		!repo.delayRequest.CreatedAt.Equal(repo.delayRequest.UpdatedAt) {
		t.Errorf("repository received uninitialized delay request: %+v", repo.delayRequest)
	}
	if created == nil || created.ID != repo.delayRequest.ID {
		t.Errorf("CreateDelayRequest() result = %+v, want persisted request", created)
	}
}

func TestCreateDelayRequestRejectsInvalidInputBeforeRepositoryCall(t *testing.T) {
	tests := []struct {
		name     string
		artistID uuid.UUID
		input    order.CreateDelayRequestInput
	}{
		{
			name:  "missing artist ID",
			input: order.CreateDelayRequestInput{OrderID: uuid.New(), DeadlineRequestAt: time.Now(), Reason: order.DelayReasonOther},
		},
		{
			name:     "missing order ID",
			artistID: uuid.New(),
			input:    order.CreateDelayRequestInput{DeadlineRequestAt: time.Now(), Reason: order.DelayReasonOther},
		},
		{
			name:     "missing requested deadline",
			artistID: uuid.New(),
			input:    order.CreateDelayRequestInput{OrderID: uuid.New(), Reason: order.DelayReasonOther},
		},
		{
			name:     "invalid reason",
			artistID: uuid.New(),
			input:    order.CreateDelayRequestInput{OrderID: uuid.New(), DeadlineRequestAt: time.Now(), Reason: "INVALID"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepo{}
			usecase := newTestOrderUsecase(repo, fakeStorage{})

			created, err := usecase.CreateDelayRequest(context.Background(), test.artistID, test.input)
			wantInvalidInput(t, err)
			if created != nil {
				t.Errorf("CreateDelayRequest() result = %+v, want nil", created)
			}
			if repo.delayRequestCalled {
				t.Error("repository CreateDelayRequest() should not be called for invalid input")
			}
		})
	}
}

func TestCreateDelayRequestPropagatesRepositoryError(t *testing.T) {
	wantErr := apperror.Conflict("delay request already exists")
	repo := &fakeRepo{delayRequestErr: wantErr}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	created, err := usecase.CreateDelayRequest(context.Background(), uuid.New(), order.CreateDelayRequestInput{
		OrderID:           uuid.New(),
		DeadlineRequestAt: time.Now().Add(24 * time.Hour),
		Reason:            order.DelayReasonOther,
		Explanation:       "Other reason",
	})
	if created != nil {
		t.Errorf("CreateDelayRequest() result = %+v, want nil", created)
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("CreateDelayRequest() error = %v, want %v", err, wantErr)
	}
	if !repo.delayRequestCalled {
		t.Error("repository CreateDelayRequest() was not called")
	}
}

const (
	orderPreviewWidth  = 640
	orderPreviewHeight = 320
)

func TestCancelOrder_Success(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	participant := order.ParticipantCustomer
	participantID := uuid.New()
	orderID := uuid.New()

	err := usecase.CancelOrder(
		context.Background(),
		participant,
		participantID,
		orderID,
	)
	if err != nil {
		t.Fatalf("CancelOrder() error = %v, want nil", err)
	}

	if !repo.cancelCalled {
		t.Fatal("CancelOrder() repository method was not called")
	}

	if repo.cancelParticipant != participant {
		t.Errorf("Participant = %v, want %v", repo.cancelParticipant, participant)
	}

	if repo.cancelParticipantID != participantID {
		t.Errorf("ParticipantID = %v, want %v", repo.cancelParticipantID, participantID)
	}

	if repo.cancelOrderID != orderID {
		t.Errorf("OrderID = %v, want %v", repo.cancelOrderID, orderID)
	}
}

func TestCancelOrder_RejectsInvalidParticipant(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	err := usecase.CancelOrder(
		context.Background(),
		order.Participant("invalid_role"),
		uuid.New(),
		uuid.New(),
	)

	if !errors.Is(err, order.ErrUnsupportedParticipantRole) {
		t.Errorf("CancelOrder() error = %v, want %v", err, order.ErrUnsupportedParticipantRole)
	}

	if repo.cancelCalled {
		t.Error("CancelOrder() repository method should not be called")
	}
}

func TestCancelOrder_RejectsMissingParticipantID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	err := usecase.CancelOrder(
		context.Background(),
		order.ParticipantCustomer,
		uuid.Nil,
		uuid.New(),
	)

	if !errors.Is(err, order.ErrMissingParticipantID) {
		t.Errorf("CancelOrder() error = %v, want %v", err, order.ErrMissingParticipantID)
	}

	if repo.cancelCalled {
		t.Error("CancelOrder() repository method should not be called")
	}
}

func TestCancelOrder_RejectsMissingOrderID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	err := usecase.CancelOrder(
		context.Background(),
		order.ParticipantCustomer,
		uuid.New(),
		uuid.Nil,
	)

	if !errors.Is(err, order.ErrMissingOrderID) {
		t.Errorf("CancelOrder() error = %v, want %v", err, order.ErrMissingOrderID)
	}

	if repo.cancelCalled {
		t.Error("CancelOrder() repository method should not be called")
	}
}

func TestCancelOrder_PropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("database connection error")
	repo := &fakeRepo{
		cancelErr: wantErr,
	}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	err := usecase.CancelOrder(
		context.Background(),
		order.ParticipantCustomer,
		uuid.New(),
		uuid.New(),
	)

	if !errors.Is(err, wantErr) {
		t.Errorf("CancelOrder() error = %v, want %v", err, wantErr)
	}
}

func TestCancelExpiredOrders_UsesTTLRelativeToNow(t *testing.T) {
	repo := &fakeRepo{}
	usecase := newTestOrderUsecase(repo, fakeStorage{})

	before := time.Now()
	_, err := usecase.CancelExpiredOrders(context.Background())
	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.expireCalled {
		t.Fatal("expected repository CancelExpired to be called")
	}
	if repo.expireNow.Before(before) || repo.expireNow.After(after) {
		t.Fatalf("now %v not within [%v, %v]", repo.expireNow, before, after)
	}
	if want := repo.expireNow.Add(-order.PendingConfirmationTTL); !repo.expirePendingCutoff.Equal(want) {
		t.Fatalf("pending cutoff = %v, want %v", repo.expirePendingCutoff, want)
	}
}

func TestCancelExpiredOrders_ReturnsCancelledIDs(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	usecase := newTestOrderUsecase(&fakeRepo{expireIDs: ids}, fakeStorage{})

	got, err := usecase.CancelExpiredOrders(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(ids) || got[0] != ids[0] || got[1] != ids[1] {
		t.Fatalf("got %v, want %v", got, ids)
	}
}

func TestCancelExpiredOrders_PropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("db down")
	usecase := newTestOrderUsecase(&fakeRepo{expireErr: wantErr}, fakeStorage{})

	if _, err := usecase.CancelExpiredOrders(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}
