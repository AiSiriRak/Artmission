package notification

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type fakeRepository struct{ rows []Notification }

func TestSpecificNotificationValidationErrors(t *testing.T) {
	// A nil repository ensures invalid input is rejected before persistence is attempted.
	uc := NewNotificationUsecase(nil)
	ctx := context.Background()
	owner := uuid.New()
	for _, tc := range []struct {
		name  string
		input CreateInput
		want  error
	}{
		{"missing recipient", CreateInput{Title: "Title", Message: "Message", Type: TypeSystem}, ErrMissingUserID},
		{"blank title", CreateInput{UserID: owner, Title: " \t", Message: "Message", Type: TypeSystem}, ErrMissingTitle},
		{"blank message", CreateInput{UserID: owner, Title: "Title", Message: " \n", Type: TypeSystem}, ErrMissingMessage},
		{"invalid type", CreateInput{UserID: owner, Title: "Title", Message: "Message", Type: "UNKNOWN"}, ErrInvalidType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Create(ctx, tc.input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			var appErr *apperror.Error
			if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
				t.Fatalf("expected invalid-input error, got %v", err)
			}
		})
	}
	for _, tc := range []struct {
		query ListQuery
		want  error
	}{
		{ListQuery{Limit: 0}, ErrInvalidLimit},
		{ListQuery{Limit: -1}, ErrInvalidLimit},
		{ListQuery{Limit: 101}, ErrInvalidLimit},
		{ListQuery{Limit: 20, Offset: -1}, ErrInvalidOffset},
	} {
		if _, err := uc.ListMine(ctx, owner, tc.query); !errors.Is(err, tc.want) {
			t.Fatalf("query %+v: got %v, want %v", tc.query, err, tc.want)
		}
	}
	if _, err := uc.ListMine(ctx, uuid.Nil, ListQuery{Limit: 20}); !errors.Is(err, ErrMissingUserID) {
		t.Fatalf("missing feed owner: %v", err)
	}
	if err := uc.MarkRead(ctx, uuid.New(), uuid.Nil); !errors.Is(err, ErrMissingUserID) {
		t.Fatalf("missing read owner: %v", err)
	}
	if err := uc.MarkRead(ctx, uuid.Nil, owner); !errors.Is(err, ErrMissingNotificationID) {
		t.Fatalf("missing notification ID: %v", err)
	}
	if _, err := uc.CountUnread(ctx, uuid.Nil); !errors.Is(err, ErrMissingUserID) {
		t.Fatalf("missing count owner: %v", err)
	}
}

func (r *fakeRepository) Create(_ context.Context, n *Notification) error {
	r.rows = append(r.rows, *n)
	return nil
}
func (r *fakeRepository) ListByUserID(_ context.Context, id uuid.UUID, q ListQuery) (Page, error) {
	items := make([]Notification, 0)
	for _, n := range r.rows {
		if n.UserID == id {
			items = append(items, n)
		}
	}
	total := len(items)
	if q.Offset >= total {
		return Page{Items: []Notification{}, Total: total}, nil
	}
	end := q.Offset + q.Limit
	if end > total {
		end = total
	}
	return Page{Items: items[q.Offset:end], Total: total}, nil
}
func (r *fakeRepository) MarkRead(_ context.Context, id, userID uuid.UUID) error {
	for i := range r.rows {
		if r.rows[i].ID == id && r.rows[i].UserID == userID {
			r.rows[i].IsRead = true
			return nil
		}
	}
	return ErrNotFound
}
func (r *fakeRepository) CountUnread(_ context.Context, id uuid.UUID) (int, error) {
	count := 0
	for _, n := range r.rows {
		if n.UserID == id && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func TestCreateAndReadOwnNotification(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepository{}
	uc := NewNotificationUsecase(repo)
	owner := uuid.New()
	n, err := uc.Create(ctx, CreateInput{UserID: owner, Title: " New order ", Message: " Review the brief ", Type: TypeOrderUpdate})
	if err != nil {
		t.Fatal(err)
	}
	if n.ID == uuid.Nil || n.CreatedAt.IsZero() || n.IsRead || n.Title != "New order" {
		t.Fatalf("unexpected notification: %+v", n)
	}
	if err := uc.MarkRead(ctx, n.ID, uuid.New()); err != ErrNotFound {
		t.Fatalf("another recipient: %v", err)
	}
	count, _ := uc.CountUnread(ctx, owner)
	if count != 1 {
		t.Fatalf("unread = %d", count)
	}
	for i := 0; i < 2; i++ {
		if err := uc.MarkRead(ctx, n.ID, owner); err != nil {
			t.Fatal(err)
		}
	}
	count, _ = uc.CountUnread(ctx, owner)
	if count != 0 {
		t.Fatalf("unread = %d", count)
	}
	page, err := uc.ListMine(ctx, uuid.New(), ListQuery{Limit: 20})
	if err != nil || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("empty feed: %+v, %v", page, err)
	}
}

func TestRejectInvalidInput(t *testing.T) {
	uc := NewNotificationUsecase(&fakeRepository{})
	ctx := context.Background()
	for _, in := range []CreateInput{
		{UserID: uuid.New(), Title: "title", Message: "message", Type: "INVALID"},
		{UserID: uuid.New(), Title: " ", Message: "message", Type: TypeSystem},
		{Title: "title", Message: "message", Type: TypeSystem},
	} {
		if _, err := uc.Create(ctx, in); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
	for _, q := range []ListQuery{{Limit: 0}, {Limit: 101}, {Limit: 20, Offset: -1}} {
		if _, err := uc.ListMine(ctx, uuid.New(), q); err == nil {
			t.Fatalf("accepted %+v", q)
		}
	}
}

type notificationRepoFake struct {
	order.OrderRepository
	item order.Order
	fail error
}

func (r *notificationRepoFake) Create(_ context.Context, item *order.Order) error {
	item.ArtistID = r.item.ArtistID
	return r.fail
}
func (r *notificationRepoFake) ConfirmOrder(_ context.Context, _, _ uuid.UUID, status order.Status) error {
	r.item.Status = status
	return r.fail
}
func (r *notificationRepoFake) GetOrderByID(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetailData, error) {
	return &order.OrderDetailData{Order: r.item}, nil
}

type notificationSinkFake struct {
	inputs []CreateInput
	fail   error
}

func (s *notificationSinkFake) Create(_ context.Context, in CreateInput) (*Notification, error) {
	if s.fail != nil {
		return nil, s.fail
	}
	s.inputs = append(s.inputs, in)
	return &Notification{}, nil
}

type notificationTxFake struct{ committed bool }

func (tx *notificationTxFake) Transaction(ctx context.Context, fn func(context.Context) error) error {
	err := fn(ctx)
	tx.committed = err == nil
	return err
}
func TestNewOrderNotifiesArtist(t *testing.T) {
	artist := uuid.New()
	sink := &notificationSinkFake{}
	tx := &notificationTxFake{}
	repo := NewNotifyingRepository(&notificationRepoFake{item: order.Order{ArtistID: artist}}, sink, tx)
	deadline := time.Now().Add(7 * 24 * time.Hour)
	err := repo.Create(context.Background(), &order.Order{ID: uuid.New(), Name: "Portrait", Status: order.StatusPending, DeadlineAt: deadline})
	if err != nil {
		t.Fatal(err)
	}
	if !tx.committed || len(sink.inputs) != 1 || sink.inputs[0].UserID != artist {
		t.Fatalf("notification: %+v", sink.inputs)
	}
	in := sink.inputs[0]
	if in.DeadlineAt == nil || !in.DeadlineAt.Equal(deadline) {
		t.Fatalf("missing initial order deadline: %+v", in)
	}
	if in.EventType != EventOrderCreated || in.OrderID == nil || in.OrderStatus == nil || *in.OrderStatus != order.StatusPending {
		t.Fatalf("missing card metadata: %+v", in)
	}
	if strings.Contains(in.Message, "Order:") || strings.Contains(in.Message, "Status:") {
		t.Fatalf("technical metadata leaked into display message: %s", in.Message)
	}
}

func TestCreatePreservesNotificationCardData(t *testing.T) {
	ctx := context.Background()
	orderID, recipient := uuid.New(), uuid.New()
	status := order.StatusNotPaid
	deadline := time.Now().UTC()
	uc := NewNotificationUsecase(&fakeRepository{})
	n, err := uc.Create(ctx, CreateInput{
		UserID: recipient, Title: "Order Accepted: Portrait", Message: "Please complete payment.", Type: TypeOrderUpdate,
		EventType: EventOrderAccepted, OrderID: &orderID, OrderStatus: &status, DeadlineAt: &deadline,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n.EventType != EventOrderAccepted || n.OrderID == nil || *n.OrderID != orderID ||
		n.OrderStatus == nil || *n.OrderStatus != status || n.DeadlineAt == nil || !n.DeadlineAt.Equal(deadline) {
		t.Fatalf("metadata lost: %+v", n)
	}
}

func TestRejectInvalidCardMetadata(t *testing.T) {
	status := order.Status("UNKNOWN")
	for _, tc := range []struct {
		name   string
		mutate func(*CreateInput)
		want   error
	}{
		{"event", func(in *CreateInput) { in.EventType = "UNKNOWN" }, ErrInvalidEventType},
		{"status", func(in *CreateInput) { in.OrderStatus = &status }, ErrInvalidOrderStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := CreateInput{UserID: uuid.New(), Title: "Title", Message: "Message", Type: TypeOrderUpdate}
			tc.mutate(&in)
			if _, err := NewNotificationUsecase(nil).Create(context.Background(), in); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestOrderAcceptanceNotifiesCustomer(t *testing.T) {
	customer := uuid.New()
	artist := uuid.New()
	sink := &notificationSinkFake{}
	repo := NewNotifyingRepository(&notificationRepoFake{item: order.Order{CustomerID: customer, ArtistID: artist, Name: "Portrait"}}, sink, &notificationTxFake{})
	if err := repo.ConfirmOrder(context.Background(), artist, uuid.New(), order.StatusNotPaid); err != nil {
		t.Fatal(err)
	}
	if len(sink.inputs) != 1 || sink.inputs[0].UserID != customer {
		t.Fatalf("notification: %+v", sink.inputs)
	}
}
func TestFailedOrderDoesNotNotify(t *testing.T) {
	sink := &notificationSinkFake{}
	tx := &notificationTxFake{}
	repo := NewNotifyingRepository(&notificationRepoFake{fail: errors.New("write failed")}, sink, tx)
	if err := repo.Create(context.Background(), &order.Order{}); err == nil {
		t.Fatal("expected error")
	}
	if len(sink.inputs) != 0 || tx.committed {
		t.Fatal("failed write must not notify or commit")
	}
}
func TestNotificationFailureAbortsTransaction(t *testing.T) {
	tx := &notificationTxFake{}
	repo := NewNotifyingRepository(&notificationRepoFake{}, &notificationSinkFake{fail: errors.New("notification failed")}, tx)
	if err := repo.Create(context.Background(), &order.Order{Status: order.StatusPending}); err == nil || tx.committed {
		t.Fatal("notification failure must abort the order transaction")
	}
}
