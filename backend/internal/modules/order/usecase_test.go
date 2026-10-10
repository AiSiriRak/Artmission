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

	created              *order.Order
	createdDeliverable   *order.Deliverable
	deliverableArtistID  uuid.UUID
	deliverableOrderID   uuid.UUID
	deliverableCreateErr error
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

func (repo *fakeRepo) Create(ctx context.Context, item *order.Order) error {
	repo.created = item
	return nil
}

func (repo *fakeRepo) CreateDeliverable(
	_ context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
	item *order.Deliverable,
) error {
	repo.createdDeliverable = item
	repo.deliverableArtistID = artistID
	repo.deliverableOrderID = orderID
	return repo.deliverableCreateErr
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
	uploads      []storedObject
	deletes      []string
	failUploadAt int
	uploadErr    error
	deleteErr    error
}

func (storage *recordingStorage) GetPresignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Participant = "admin"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantForbidden(t, err)
}

func TestViewOrders_RejectsMissingParticipantID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.ParticipantID = uuid.Nil
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidSort(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Sort = "created_at"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidOrder(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Order = "sideways"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsLimitOutOfRange(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	for _, limit := range []int{-1, order.MaxLimit + 1} {
		q := validQuery()
		q.Limit = limit
		_, err := usecase.ViewOrders(context.Background(), q)
		wantInvalidInput(t, err)
	}
}

func TestViewOrders_RejectsNegativeOffset(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Offset = -1
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidStatus(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

	q := validQuery()
	q.Statuses = []order.Status{"NOT_A_STATUS"}
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_NormalizesStatuses_DedupesAndSorts(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{url: "https://cdn.test/"})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{err: wantErr})

	_, err := usecase.ViewOrders(context.Background(), validQuery())
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("ViewOrders() error = %v, want *apperror.Error{Code: CodeInternal}", err)
	}
}

func TestGetOrder_RejectsUnsupportedParticipant(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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

	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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

	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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

	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(&fakeRepo{}, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	usecase := order.NewOrderUsecase(repo, fakeStorage{})

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
	if format != "jpeg" || preview.Bounds().Dx() != orderPreviewSize || preview.Bounds().Dy() != orderPreviewHeight {
		t.Errorf("preview dimensions/format = %dx%d %q", preview.Bounds().Dx(), preview.Bounds().Dy(), format)
	}
	if created.OriginalImageKey != storage.uploads[0].key || created.PreviewImageKey != storage.uploads[1].key {
		t.Errorf("deliverable keys = (%q, %q), want uploaded keys", created.OriginalImageKey, created.PreviewImageKey)
	}
	if repo.deliverableArtistID != artistID || repo.deliverableOrderID != orderID || repo.createdDeliverable != created {
		t.Error("repository did not receive the created deliverable with the authenticated artist and order IDs")
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

const (
	orderPreviewSize   = 400
	orderPreviewHeight = 200
)
