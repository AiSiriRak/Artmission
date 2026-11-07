package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type fakeRepository struct{ rows []Notification }

func TestNotificationTypeStoresSpecificEvent(t *testing.T) {
	uc := NewNotificationUsecase(&fakeRepository{})
	n, err := uc.Create(context.Background(), CreateInput{UserID: uuid.New(), Title: "Draft Submitted", Message: "Review draft #2", Type: Type("DRAFT_SUBMITTED")})
	if err != nil || n.Type != Type("DRAFT_SUBMITTED") {
		t.Fatalf("specific event type: %+v %v", n, err)
	}
	if _, err := uc.Create(context.Background(), CreateInput{UserID: uuid.New(), Title: "Title", Message: "Message", Type: Type("ORDER_UPDATE")}); !errors.Is(err, ErrInvalidType) {
		t.Fatalf("legacy category must be rejected: %v", err)
	}
}

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
		{"missing recipient", CreateInput{Title: "Title", Message: "Message", Type: TypeGeneral}, ErrMissingUserID},
		{"blank title", CreateInput{UserID: owner, Title: " \t", Message: "Message", Type: TypeGeneral}, ErrMissingTitle},
		{"blank message", CreateInput{UserID: owner, Title: "Title", Message: " \n", Type: TypeGeneral}, ErrMissingMessage},
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
	n, err := uc.Create(ctx, CreateInput{UserID: owner, Title: " New order ", Message: " Review the brief ", Type: TypeOrderCreated})
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
		{UserID: uuid.New(), Title: " ", Message: "message", Type: TypeGeneral},
		{Title: "title", Message: "message", Type: TypeGeneral},
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

func TestCreatePreservesNotificationCardData(t *testing.T) {
	ctx := context.Background()
	orderID, recipient := uuid.New(), uuid.New()
	deliverableID := uuid.New()
	uc := NewNotificationUsecase(&fakeRepository{})
	n, err := uc.Create(ctx, CreateInput{
		UserID: recipient, Title: "Order Accepted: Portrait", Message: "Please complete payment.", Type: TypeOrderAccepted, OrderID: &orderID, DeliverableID: &deliverableID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n.Type != TypeOrderAccepted || n.OrderID == nil || *n.OrderID != orderID ||
		n.DeliverableID == nil || *n.DeliverableID != deliverableID {
		t.Fatalf("metadata lost: %+v", n)
	}
}

func TestRejectInvalidCardMetadata(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*CreateInput)
		want   error
	}{
		{"event", func(in *CreateInput) { in.Type = "UNKNOWN" }, ErrInvalidType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := CreateInput{UserID: uuid.New(), Title: "Title", Message: "Message", Type: TypeOrderCreated}
			tc.mutate(&in)
			if _, err := NewNotificationUsecase(nil).Create(context.Background(), in); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

type orderReaderFake struct {
	detail order.OrderDetailData
	err    error
}

func (r orderReaderFake) ReadNotificationOrder(context.Context, order.NotificationEvent) (*order.OrderDetailData, error) {
	return &r.detail, r.err
}

func TestOrderEventCards(t *testing.T) {
	artist, customer, id := uuid.New(), uuid.New(), uuid.New()
	for _, tc := range []struct {
		kind, decision string
		event          Type
		customer       bool
		count          int
	}{
		{"ORDER_CREATED", "", TypeOrderCreated, false, 1},
		{"ORDER_CONFIRMED", "NOT_PAID", TypeOrderAccepted, true, 1},
		{"ORDER_CONFIRMED", "CANCEL", TypeOrderDeclined, true, 1},
		{"PAYMENT_SECURED", "", TypePaymentSecured, false, 1},
		{"DRAFT_SUBMITTED", "WAIT", TypeDraftSubmitted, true, 1},
		{"DRAFT_SUBMITTED", "APPROVED", TypeFinalArtworkReady, true, 1},
		{"DELIVERABLE_CONFIRMED", "APPROVED", TypeOrderCompleted, false, 1},
		{"DELIVERABLE_CONFIRMED", "REJECTED", TypeRevisionRequested, false, 1},
		{"DEADLINE_EXTENSION_REQUESTED", "", TypeDeadlineRequested, true, 1},
		{"DEADLINE_EXTENSION_DECIDED", "APPROVED", TypeDeadlineApproved, false, 1},
		{"DEADLINE_EXTENSION_DECIDED", "REJECTED", TypeDeadlineRejected, false, 1},
		{"ORDER_EXPIRED", "", TypeOrderCanceled, false, 2},
		{"PAYMENT_EXPIRED", "", TypeOrderCanceled, false, 2},
		{"DELIVERABLE_AUTO_APPROVED", "", TypeOrderCompleted, false, 2},
		{"DELAY_REQUEST_EXPIRED", "", TypeDeadlineRejected, false, 2},
		{"ORDER_CANCELED", "", TypeOrderCanceled, false, 1},
	} {
		t.Run(tc.kind+tc.decision, func(t *testing.T) {
			deadline := time.Now().UTC()
			repo := &fakeRepository{}
			reader := orderReaderFake{detail: order.OrderDetailData{Order: order.Order{ID: id, ArtistID: artist, CustomerID: customer, Name: "Portrait", Status: order.StatusInProcess}}}
			notifier := NewOrderNotifier(reader, NewNotificationUsecase(repo))
			err := notifier.NotifyOrder(context.Background(), order.NotificationEvent{Kind: tc.kind, OrderID: id, Decision: tc.decision, Version: 2, DeadlineAt: &deadline})
			if err != nil {
				t.Fatal(err)
			}
			uc := NewNotificationUsecase(repo)
			owner := artist
			if tc.customer {
				owner = customer
			}
			page, err := uc.ListMine(context.Background(), owner, ListQuery{Limit: 20})
			if err != nil || len(page.Items) != 1 {
				t.Fatalf("missing card: %+v %v", page, err)
			}
			n := page.Items[0]
			if n.Type != tc.event || n.OrderID == nil || *n.OrderID != id {
				t.Fatalf("wrong card: %+v", n)
			}
			if tc.count == 2 {
				page, _ = uc.ListMine(context.Background(), customer, ListQuery{Limit: 20})
				if len(page.Items) != 1 {
					t.Fatal("worker event must notify both participants")
				}
			}
		})
	}
}

func TestOrderNotificationReaderFailure(t *testing.T) {
	want := errors.New("lookup failed")
	notifier := NewOrderNotifier(orderReaderFake{err: want}, NewNotificationUsecase(&fakeRepository{}))
	if err := notifier.NotifyOrder(context.Background(), order.NotificationEvent{}); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestRevisionMessageUsesMatchedDeliverableOrFallback(t *testing.T) {
	id := uuid.New()
	for _, tc := range []struct {
		name         string
		deliverables []order.Deliverable
		want         string
	}{
		{"matched", []order.Deliverable{{ID: uuid.New(), Version: 1}, {ID: id, Version: 2}}, "The customer has requested changes for #2 Draft. Please review their feedback in the order details."},
		{"not found", []order.Deliverable{{ID: uuid.New(), Version: 1}}, "The customer has requested changes. Please review their feedback in the order details."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{}
			notifier := NewOrderNotifier(orderReaderFake{detail: order.OrderDetailData{
				Order:        order.Order{ID: uuid.New(), ArtistID: uuid.New(), CustomerID: uuid.New(), Name: "Portrait"},
				Deliverables: tc.deliverables,
			}}, NewNotificationUsecase(repo))
			if err := notifier.NotifyOrder(context.Background(), order.NotificationEvent{Kind: "DELIVERABLE_CONFIRMED", DeliverableID: id, Version: 99, Decision: "REJECTED"}); err != nil {
				t.Fatal(err)
			}
			if len(repo.rows) != 1 || repo.rows[0].Message != tc.want {
				t.Fatalf("revision fallback changed: %+v", repo.rows)
			}
		})
	}
}

func TestWalletSettlementCards(t *testing.T) {
	for _, tc := range []struct {
		name, kind, decision string
		outcome              payment.SettlementOutcome
		refund, payout       int64
	}{
		{"artist cancellation", "ORDER_CANCELED", "", payment.ArtistCancelled, 10001, 0},
		{"customer cancellation", "ORDER_CANCELED", "", payment.CustomerCancelled, 5000, 5001},
		{"deadline expiry", "ORDER_EXPIRED", "", payment.DeadlineExpired, 10001, 0},
		{"customer acceptance", "DELIVERABLE_CONFIRMED", "APPROVED", payment.Completed, 0, 10001},
		{"final draft", "DRAFT_SUBMITTED", "APPROVED", payment.Completed, 0, 10001},
		{"review expiry", "DELIVERABLE_AUTO_APPROVED", "", payment.Completed, 0, 10001},
	} {
		t.Run(tc.name, func(t *testing.T) {
			artist, customer, id := uuid.New(), uuid.New(), uuid.New()
			success := payment.TransferSuccess
			summary := &payment.Summary{Status: "SETTLED", SettlementOutcome: &tc.outcome,
				RefundSatang: tc.refund, ArtistPayoutSatang: tc.payout}
			if tc.refund > 0 {
				transactionID := uuid.New()
				summary.RefundTransactionID = &transactionID
				summary.RefundTransferStatus = &success
			}
			if tc.payout > 0 {
				transactionID := uuid.New()
				summary.ArtistPayoutTransactionID = &transactionID
				summary.ArtistPayoutTransferStatus = &success
			}
			repo := &fakeRepository{}
			n := NewOrderNotifier(orderReaderFake{detail: order.OrderDetailData{Order: order.Order{ID: id, Name: "Portrait", ArtistID: artist, CustomerID: customer, Payment: summary}}}, NewNotificationUsecase(repo))
			if err := n.NotifyOrder(context.Background(), order.NotificationEvent{Kind: tc.kind, Decision: tc.decision}); err != nil {
				t.Fatal(err)
			}
			for _, credit := range []struct {
				kind   Type
				owner  uuid.UUID
				amount int64
			}{{"REFUND_CREDITED", customer, tc.refund}, {"WALLET_CREDITED", artist, tc.payout}} {
				count := 0
				for _, card := range repo.rows {
					if card.Type != credit.kind {
						continue
					}
					count++
					amount := fmt.Sprintf("%d.%02d THB", credit.amount/100, credit.amount%100)
					if card.UserID != credit.owner || card.OrderID == nil || *card.OrderID != id || !strings.Contains(card.Message, amount) || !strings.Contains(card.Message, "wallet") {
						t.Fatalf("incorrect credit notification: %+v", card)
					}
				}
				want := 0
				if credit.amount > 0 {
					want = 1
				}
				if count != want {
					t.Fatalf("%s count %d want %d", credit.kind, count, want)
				}
			}
		})
	}
}

func TestWalletCardsRequireSuccessfulRecordedCredit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*payment.Summary)
	}{
		{"unpaid", func(p *payment.Summary) { p.Status = "PENDING" }},
		{"held", func(p *payment.Summary) { p.Status = "HELD" }},
		{"missing outcome", func(p *payment.Summary) { p.SettlementOutcome = nil }},
		{"missing transaction", func(p *payment.Summary) { p.RefundTransactionID = nil }},
		{"zero transaction", func(p *payment.Summary) { p.RefundTransactionID = &uuid.Nil }},
		{"missing transfer status", func(p *payment.Summary) { p.RefundTransferStatus = nil }},
		{"pending transfer", func(p *payment.Summary) { status := payment.TransferPending; p.RefundTransferStatus = &status }},
		{"failed transfer", func(p *payment.Summary) { status := payment.TransferFailed; p.RefundTransferStatus = &status }},
		{"zero amount", func(p *payment.Summary) { p.RefundSatang = 0 }},
		{"negative amount", func(p *payment.Summary) { p.RefundSatang = -1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outcome, status, transactionID := payment.ArtistCancelled, payment.TransferSuccess, uuid.New()
			p := &payment.Summary{Status: "SETTLED", SettlementOutcome: &outcome, RefundSatang: 10001, RefundTransactionID: &transactionID, RefundTransferStatus: &status}
			tc.mutate(p)
			repo := &fakeRepository{}
			n := NewOrderNotifier(orderReaderFake{detail: order.OrderDetailData{Order: order.Order{ID: uuid.New(), ArtistID: uuid.New(), CustomerID: uuid.New(), Name: "Portrait", Payment: p}}}, NewNotificationUsecase(repo))
			if err := n.NotifyOrder(context.Background(), order.NotificationEvent{Kind: "ORDER_CANCELED"}); err != nil {
				t.Fatal(err)
			}
			if len(repo.rows) != 1 || repo.rows[0].Type != TypeOrderCanceled {
				t.Fatalf("unexpected wallet card: %+v", repo.rows)
			}
		})
	}
}
