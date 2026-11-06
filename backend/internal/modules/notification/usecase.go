package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/google/uuid"
)

type notificationUsecase struct{ repo NotificationRepository }

// NewNotificationUsecase connects notification operations to their persistence dependency.
func NewNotificationUsecase(repo NotificationRepository) NotificationUsecase {
	return &notificationUsecase{repo: repo}
}

// Create validates the recipient and category, trims message text, and saves an unread notification.
func (u *notificationUsecase) Create(ctx context.Context, in CreateInput) (*Notification, error) {
	title, message := strings.TrimSpace(in.Title), strings.TrimSpace(in.Message)
	if in.UserID == uuid.Nil {
		return nil, ErrMissingUserID
	}
	if title == "" {
		return nil, ErrMissingTitle
	}
	if message == "" {
		return nil, ErrMissingMessage
	}
	if in.Type == "" {
		in.Type = TypeGeneral
	}
	if !in.Type.IsValid() {
		return nil, ErrInvalidType
	}
	n := &Notification{
		ID: uuid.New(), UserID: in.UserID, Title: title, Message: message, Type: in.Type, CreatedAt: time.Now(),
		OrderID: in.OrderID, DeliverableID: in.DeliverableID, DelayRequestID: in.DelayRequestID,
	}
	if err := u.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// ListMine validates pagination and returns an empty slice when the recipient has no results.
func (u *notificationUsecase) ListMine(ctx context.Context, id uuid.UUID, q ListQuery) (Page, error) {
	if id == uuid.Nil {
		return Page{}, ErrMissingUserID
	}
	if q.Limit < 1 || q.Limit > 100 {
		return Page{}, ErrInvalidLimit
	}
	if q.Offset < 0 {
		return Page{}, ErrInvalidOffset
	}
	page, err := u.repo.ListByUserID(ctx, id, q)
	if page.Items == nil {
		page.Items = []Notification{}
	}
	return page, err
}

// MarkRead delegates the ownership check and read update to one repository operation.
func (u *notificationUsecase) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrMissingUserID
	}
	if id == uuid.Nil {
		return ErrMissingNotificationID
	}
	return u.repo.MarkRead(ctx, id, userID)
}

// CountUnread returns the recipient's unread total independently of feed pagination.
func (u *notificationUsecase) CountUnread(ctx context.Context, id uuid.UUID) (int, error) {
	if id == uuid.Nil {
		return 0, ErrMissingUserID
	}
	return u.repo.CountUnread(ctx, id)
}

// NotificationCreator keeps order activity independent of feed operations.
type NotificationCreator interface {
	Create(context.Context, CreateInput) (*Notification, error)
}
type OrderNotificationReader interface {
	ReadNotificationOrder(context.Context, order.NotificationEvent) (*order.OrderDetailData, error)
}
type orderNotifier struct {
	reader        OrderNotificationReader
	notifications NotificationCreator
}

var _ order.OrderNotifier = (*orderNotifier)(nil)
var _ payment.Notifier = (*orderNotifier)(nil)

func NewPaymentNotifier(reader OrderNotificationReader, notifications NotificationCreator) payment.Notifier {
	return &orderNotifier{reader: reader, notifications: notifications}
}

func (n *orderNotifier) PaymentSecured(ctx context.Context, orderID uuid.UUID) error {
	return n.NotifyOrder(ctx, order.NotificationEvent{Kind: "PAYMENT_SECURED", OrderID: orderID})
}

func (n *orderNotifier) PaymentExpired(ctx context.Context, orderID uuid.UUID) error {
	return n.NotifyOrder(ctx, order.NotificationEvent{Kind: "PAYMENT_EXPIRED", OrderID: orderID})
}

func NewOrderNotifier(reader OrderNotificationReader, notifications NotificationCreator) order.OrderNotifier {
	return &orderNotifier{reader: reader, notifications: notifications}
}

func (n *orderNotifier) send(ctx context.Context, item order.Order, recipient uuid.UUID, event Type, title, message string, source order.NotificationEvent) error {
	var deliverableID, requestID *uuid.UUID
	if source.DeliverableID != uuid.Nil {
		deliverableID = &source.DeliverableID
	}
	if source.DelayRequestID != uuid.Nil {
		requestID = &source.DelayRequestID
	}
	_, err := n.notifications.Create(ctx, CreateInput{UserID: recipient, Type: event,
		Title: fmt.Sprintf("%s: %s", title, item.Name), Message: message, OrderID: &item.ID, DeliverableID: deliverableID, DelayRequestID: requestID})
	return err
}

func (n *orderNotifier) NotifyOrder(ctx context.Context, e order.NotificationEvent) error {
	detail, err := n.reader.ReadNotificationOrder(ctx, e)
	if err != nil {
		return err
	}
	item := detail.Order
	for _, d := range detail.Deliverables {
		if d.ID == e.DeliverableID {
			e.Version = d.Version
			break
		}
	}
	recipient := item.ArtistID
	event, title, message := TypeGeneral, "", ""
	switch e.Kind {
	case "ORDER_CREATED":
		event, title, message = TypeOrderCreated, "New Order", "The customer has placed a new order. Please check the order details to review the brief."
	case "ORDER_CONFIRMED":
		recipient = item.CustomerID
		if e.Decision == string(order.StatusNotPaid) {
			event, title, message = TypeOrderAccepted, "Order Accepted", "The artist has accepted your order! Please complete the payment to secure your commission."
		} else {
			event, title, message = TypeOrderDeclined, "Order Declined", "The artist declined this order. You are welcome to submit a new request anytime."
		}
	case "PAYMENT_SECURED":
		event, title, message = TypePaymentSecured, "Payment Secured", "The customer has deposited the payment. You can now start working on this commission."
	case "ORDER_CANCELED":
		actor := "customer"
		if e.Participant == order.ParticipantArtist {
			recipient = item.CustomerID
			actor = "artist"
		}
		event, title, message = TypeOrderCanceled, "Order Canceled", fmt.Sprintf("The %s canceled this order. Reason: %s. %s", actor, e.Reason, e.Explanation)
	case "DRAFT_SUBMITTED":
		recipient = item.CustomerID
		if e.Decision == string(order.DeliverableDecisionApproved) {
			event, title, message = TypeFinalArtworkReady, "Final Artwork Ready", "Your final artwork is complete and ready. Please check the order details to review and download it."
		} else {
			event, title, message = TypeDraftSubmitted, "Draft Submitted", fmt.Sprintf("The artist has submitted #%d Draft for your review. Please check the order details to accept the artwork or provide your feedback.", e.Version)
		}
	case "DELIVERABLE_CONFIRMED":
		if e.Decision == string(order.DeliverableDecisionApproved) {
			event, title, message = TypeOrderCompleted, "Order Completed", "The customer has accepted your artwork!"
		} else {
			message = "The customer has requested changes. Please review their feedback in the order details."
			for _, d := range detail.Deliverables {
				if d.ID == e.DeliverableID {
					message = fmt.Sprintf("The customer has requested changes for #%d Draft. Please review their feedback in the order details.", d.Version)
					break
				}
			}
			event, title = TypeRevisionRequested, "Revision Requested"
		}
	case "DEADLINE_EXTENSION_REQUESTED":
		recipient = item.CustomerID
		event, title, message = TypeDeadlineRequested, "Deadline Extension Requested", "The artist has requested a deadline extension. Please review the request in the order details."
	case "DEADLINE_EXTENSION_DECIDED":
		if e.Decision == string(order.DelayDecisionApproved) {
			event, title, message = TypeDeadlineApproved, "Deadline Extended", "Your postpone request was approved. Please check the order details for your new deadline."
		} else {
			event, title, message = TypeDeadlineRejected, "Postpone Request Declined", "Your postpone request was declined. Your deadline remains unchanged."
		}
	case "ORDER_EXPIRED":
		if err := n.send(ctx, item, item.CustomerID, TypeOrderCanceled, "Order Canceled", "This order was automatically canceled because its confirmation window or deadline expired.", e); err != nil {
			return err
		}
		event, title, message = TypeOrderCanceled, "Order Canceled", "This order was automatically canceled because its confirmation window or deadline expired."
	case "PAYMENT_EXPIRED":
		message = "This order was automatically canceled because payment was not completed within 24 hours."
		if err := n.send(ctx, item, item.CustomerID, TypeOrderCanceled, "Order Canceled", message, e); err != nil {
			return err
		}
		event, title = TypeOrderCanceled, "Order Canceled"
	case "DELIVERABLE_AUTO_APPROVED":
		if err := n.send(ctx, item, item.CustomerID, TypeFinalArtworkReady, "Final Artwork Ready", "The review window expired and your artwork was automatically accepted. Please check the order details to download it.", e); err != nil {
			return err
		}
		event, title, message = TypeOrderCompleted, "Order Completed", "The review window expired and your artwork was automatically accepted."
	case "DELAY_REQUEST_EXPIRED":
		if err := n.send(ctx, item, item.CustomerID, TypeDeadlineRejected, "Deadline Extension Expired", "The deadline extension request expired without a response and was automatically rejected. The deadline remains unchanged.", e); err != nil {
			return err
		}
		event, title, message = TypeDeadlineRejected, "Postpone Request Declined", "The customer did not respond in time. Your deadline extension request was automatically rejected and the deadline remains unchanged."
	default:
		return ErrInvalidType
	}
	if err := n.send(ctx, item, recipient, event, title, message, e); err != nil {
		return err
	}
	switch e.Kind {
	case "ORDER_CANCELED", "ORDER_EXPIRED", "DELIVERABLE_AUTO_APPROVED":
		return n.notifyWalletCredits(ctx, item)
	case "DRAFT_SUBMITTED", "DELIVERABLE_CONFIRMED":
		if e.Decision == string(order.DeliverableDecisionApproved) {
			return n.notifyWalletCredits(ctx, item)
		}
	}
	return nil
}

func (n *orderNotifier) notifyWalletCredits(ctx context.Context, item order.Order) error {
	p := item.Payment
	if p == nil || p.Status != string(payment.StatusSettled) || p.SettlementOutcome == nil {
		return nil
	}
	var reason string
	switch *p.SettlementOutcome {
	case payment.Completed:
		reason = "the order was completed"
	case payment.ArtistCancelled:
		reason = "the artist canceled the order"
	case payment.CustomerCancelled:
		reason = "the customer canceled the order"
	case payment.DeadlineExpired:
		reason = "the order deadline expired"
	default:
		return nil
	}
	for _, credit := range []struct {
		recipient     uuid.UUID
		typeName      Type
		title         string
		amount        int64
		transactionID *uuid.UUID
		status        *payment.TransferStatus
	}{
		{item.ArtistID, TypeWalletCredited, "Payment Credited", p.ArtistPayoutSatang, p.ArtistPayoutTransactionID, p.ArtistPayoutTransferStatus},
		{item.CustomerID, TypeRefundCredited, "Refund Credited", p.RefundSatang, p.RefundTransactionID, p.RefundTransferStatus},
	} {
		// An allocation is not proof of a credit; only committed successful transfers qualify.
		if credit.amount <= 0 || credit.transactionID == nil || *credit.transactionID == uuid.Nil || credit.status == nil || *credit.status != payment.TransferSuccess {
			continue
		}
		message := fmt.Sprintf("%d.%02d THB has been credited to your wallet because %s.", credit.amount/100, credit.amount%100, reason)
		if err := n.send(ctx, item, credit.recipient, credit.typeName, credit.title, message, order.NotificationEvent{}); err != nil {
			return err
		}
	}
	return nil
}
