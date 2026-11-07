//go:build integration

package notifications

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/adapters/postgres"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/baserepo"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/google/uuid"
)

type notificationCard struct {
	ID                 string     `json:"id"`
	Type               string     `json:"type"`
	OrderID            *string    `json:"order_id"`
	DeliverableID      *string    `json:"deliverable_id"`
	DeliverableVersion *int       `json:"deliverable_version"`
	DelayRequestID     *string    `json:"delay_request_id"`
	DeadlineAt         *time.Time `json:"deadline_at"`
	IsRead             bool       `json:"is_read"`
}
type notificationFeed struct {
	Items []notificationCard `json:"items"`
	Total int                `json:"total"`
}
type notificationsContext struct {
	clients                                  map[string]*apptest.Client
	accounts                                 map[string]apptest.Account
	tokens                                   map[string]string
	orderID, draftID, requestID, expiredKind string
	version                                  int
	deadline, proposed                       time.Time
	approvedDeadlines                        []time.Time
	resp                                     *apptest.Response
	page                                     notificationFeed
	card                                     notificationCard
	beforeArtist, beforeCustomer             int
}

func (n *notificationsContext) registerAccounts() error {
	n.clients = map[string]*apptest.Client{}
	n.accounts = map[string]apptest.Account{}
	n.tokens = map[string]string{}
	for _, role := range []string{"artist", "customer"} {
		c := apptest.NewClient(app.BaseURL())
		n.clients[role] = c
		var account apptest.Account
		var err error
		if role == "artist" {
			account, err = apptest.RegisterArtist(app, c, "Commission artist")
		} else {
			account, err = apptest.RegisterCustomer(app, c)
		}
		if err != nil {
			return err
		}
		n.accounts[role] = account
		n.tokens[role], err = apptest.Login(c, account.Email, account.Password)
		if err != nil {
			return err
		}
	}
	return nil
}
func (n *notificationsContext) headers(role string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + n.tokens[role]}
}
func (n *notificationsContext) request(role, method, path string, body any, want int) error {
	resp, err := n.clients[role].Do(method, path, body, n.headers(role))
	if err != nil {
		return err
	}
	n.resp = resp
	if want != 0 && resp.StatusCode != want {
		return fmt.Errorf("%s %s: got %d, want %d: %s", method, path, resp.StatusCode, want, resp.Body)
	}
	return nil
}
func (n *notificationsContext) feed(role, query string) (notificationFeed, error) {
	resp, err := n.clients[role].Do(http.MethodGet, "/notifications"+query, nil, n.headers(role))
	if err != nil {
		return notificationFeed{}, err
	}
	if resp.StatusCode != 200 {
		return notificationFeed{}, fmt.Errorf("feed %s: %d %s", role, resp.StatusCode, resp.Body)
	}
	var page notificationFeed
	err = resp.JSON(&page)
	return page, err
}
func (n *notificationsContext) rememberTotals() error {
	a, err := n.feed("artist", "")
	if err != nil {
		return err
	}
	c, err := n.feed("customer", "")
	if err != nil {
		return err
	}
	n.beforeArtist, n.beforeCustomer = a.Total, c.Total
	return nil
}
func (n *notificationsContext) unchanged() error {
	a, err := n.feed("artist", "")
	if err != nil {
		return err
	}
	c, err := n.feed("customer", "")
	if err != nil {
		return err
	}
	if a.Total != n.beforeArtist || c.Total != n.beforeCustomer {
		return fmt.Errorf("unexpected new notifications: artist %d -> %d, customer %d -> %d", n.beforeArtist, a.Total, n.beforeCustomer, c.Total)
	}
	return nil
}
func (n *notificationsContext) receives(role, kind string) error {
	page, err := n.feed(role, "")
	if err != nil {
		return err
	}
	for _, card := range page.Items {
		if card.Type == kind && card.OrderID != nil && *card.OrderID == n.orderID {
			n.card = card
			return nil
		}
	}
	return fmt.Errorf("%s has no %s notification for order %s", role, kind, n.orderID)
}
func (n *notificationsContext) exec(q string, args ...any) error {
	_, err := app.DB.ExecContext(context.Background(), q, args...)
	return err
}
func (n *notificationsContext) createOrder() error {
	artwork := uuid.New()
	if err := n.exec("INSERT INTO artworks(id,artist_id,category_id,name,description,price_satang,minimum_deadline_days) SELECT ?,?,id,'Portrait','Commission',10000,1 FROM categories LIMIT 1", artwork, n.accounts["artist"].ID); err != nil {
		return err
	}
	n.deadline = time.Now().UTC().Add(7 * 24 * time.Hour)
	if err := n.request("customer", "POST", "/orders", map[string]any{"artwork_id": artwork, "name": "Portrait", "customer_description": "Draw a portrait", "deadline_at": n.deadline}, 201); err != nil {
		return err
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := n.resp.JSON(&item); err != nil {
		return err
	}
	n.orderID = item.ID
	return nil
}
func (n *notificationsContext) orderWithStatus(status string) error {
	if err := n.createOrder(); err != nil {
		return err
	}
	if status == "PENDING" {
		return nil
	}
	if err := n.confirmOrder("accepts"); err != nil {
		return err
	}
	if status == "NOT_PAID" {
		return nil
	}
	if status != "IN_PROCESS" {
		return fmt.Errorf("unsupported fixture status %s", status)
	}
	return n.pay()
}
func (n *notificationsContext) confirmOrder(decision string) error {
	return n.request("artist", "PUT", "/orders/"+n.orderID+"/confirm", map[string]bool{"accept": decision == "accepts"}, 200)
}

// The application still handles QR/slip requests normally; only the external provider is simulated.
type notificationPaymentProvider struct{}

func (notificationPaymentProvider) GeneratePromptPayQR(_ context.Context, amount int64) (payment.QRImage, error) {
	return payment.QRImage{Payload: "bdd-test-qr", AmountSatang: amount}, nil
}
func (notificationPaymentProvider) VerifySlipImage(_ context.Context, image []byte, _ string, amount int64, _ time.Time) (payment.SlipVerification, error) {
	valid := !bytes.Contains(image, []byte("invalid-test-slip"))
	code := "200000"
	if !valid {
		code = "200500"
	}
	return payment.SlipVerification{Valid: valid, ResultCode: code, AmountSatang: amount, TransRef: uuid.NewString(), DateTime: time.Now()}, nil
}
func (n *notificationsContext) sendSlip(valid bool) error {
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 32, 32))); err != nil {
		return err
	}
	if !valid {
		imageData.WriteString("invalid-test-slip")
	}
	resp, err := n.clients["customer"].DoMultipart("POST", "/orders/"+n.orderID+"/payment/verify", nil, []apptest.MultipartFile{{FieldName: "file", Filename: "slip.png", ContentType: "image/png", Data: imageData.Bytes()}}, n.headers("customer"))
	n.resp = resp
	return err
}
func (n *notificationsContext) pay() error {
	if err := n.request("customer", "POST", "/orders/"+n.orderID+"/payment/qr", nil, 200); err != nil {
		return err
	}
	if err := n.sendSlip(true); err != nil {
		return err
	}
	var result struct {
		Success bool `json:"success"`
	}
	if err := n.resp.JSON(&result); err != nil {
		return err
	}
	if n.resp.StatusCode != 200 || !result.Success {
		return fmt.Errorf("slip verification failed: %d %s", n.resp.StatusCode, n.resp.Body)
	}
	return nil
}

func (n *notificationsContext) submitDraft() error {
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 32, 32))); err != nil {
		return err
	}
	resp, err := n.clients["artist"].DoMultipart("POST", "/orders/"+n.orderID+"/deliverables", nil, []apptest.MultipartFile{{FieldName: "deliverable_image", Filename: "draft.png", ContentType: "image/png", Data: b.Bytes()}}, n.headers("artist"))
	if err != nil {
		return err
	}
	n.resp = resp
	if resp.StatusCode != 201 {
		return fmt.Errorf("submit draft: %d %s", resp.StatusCode, resp.Body)
	}
	var d struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
	}
	if err := resp.JSON(&d); err != nil {
		return err
	}
	n.draftID, n.version = d.ID, d.Version
	return nil
}
func (n *notificationsContext) confirmDraft(decision string) error {
	comment := ""
	if decision != "accepts" {
		comment = "Please adjust colors"
	}
	return n.request("customer", "PUT", "/orders/"+n.orderID+"/deliverables/"+n.draftID+"/confirm", map[string]any{"accept": decision == "accepts", "comment": comment}, 200)
}
func (n *notificationsContext) requestExtension() error {
	n.proposed = n.deadline.Add(24 * time.Hour)
	if err := n.request("artist", "POST", "/orders/"+n.orderID+"/delay-requests", map[string]any{"deadline_request_at": n.proposed, "reason": "NEED_MORE_TIME", "explanation": "More detail required"}, 201); err != nil {
		return err
	}
	var r struct {
		ID string `json:"id"`
	}
	if err := n.resp.JSON(&r); err != nil {
		return err
	}
	n.requestID = r.ID
	return nil
}
func (n *notificationsContext) confirmExtension(decision string) error {
	if err := n.request("customer", "PUT", "/orders/"+n.orderID+"/delay-requests/"+n.requestID+"/confirm", map[string]bool{"accept": decision == "approves"}, 200); err != nil {
		return err
	}
	if decision == "approves" {
		n.deadline = n.proposed
	}
	return nil
}

// Workers have no public endpoint; invoke the same usecase as the production worker.
func orderActions(creator notification.NotificationCreator) order.OrderUsecase {
	tx := baserepo.NewTransactioner(app.DB)
	return order.NewOrderUsecase(postgres.NewOrderRepository(app.DB), storage.Private, order.WithNotifications(notification.NewOrderNotifier(postgres.NewOrderNotificationReader(app.DB), creator), tx))
}
func notificationActions() notification.NotificationUsecase {
	return notification.NewNotificationUsecase(postgres.NewNotificationRepository(app.DB))
}
func (n *notificationsContext) runExpiry() error {
	actions := orderActions(notificationActions())
	ctx := context.Background()
	if _, err := actions.CancelExpiredOrders(ctx); err != nil {
		return err
	}
	if _, err := actions.AutoApproveExpiredDeliverables(ctx); err != nil {
		return err
	}
	_, err := actions.CancelExpiredDelayRequest(ctx)
	return err
}

type failingNotifications struct{}

type failSecondNotification struct {
	creator notification.NotificationUsecase
	calls   int
}

func (f *failSecondNotification) Create(ctx context.Context, in notification.CreateInput) (*notification.Notification, error) {
	f.calls++
	if f.calls == 2 {
		return nil, errors.New("second notification failed")
	}
	return f.creator.Create(ctx, in)
}

func (failingNotifications) Create(context.Context, notification.CreateInput) (*notification.Notification, error) {
	return nil, errors.New("notification storage unavailable")
}
