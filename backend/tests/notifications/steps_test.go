//go:build integration

package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func InitializeScenario(sc *godog.ScenarioContext) {
	var n *notificationsContext
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		n = &notificationsContext{}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		// Failed worker scenarios leave an expired PENDING row; restore its clock for isolation.
		if n.expiredKind == "order" && n.orderID != "" {
			return ctx, n.exec("UPDATE orders SET created_at=? WHERE id=? AND status='PENDING'", time.Now(), n.orderID)
		}
		return ctx, nil
	})

	sc.Step(`^an artist and a customer are registered and logged in$`, func() error { return n.registerAccounts() })
	sc.Step(`^the customer creates an order$`, func() error { return n.createOrder() })
	sc.Step(`^an order with status "([^"]*)" exists$`, func(status string) error { return n.orderWithStatus(status) })
	sc.Step(`^the artist (accepts|declines) the order$`, func(decision string) error { return n.confirmOrder(decision) })
	sc.Step(`^the (artist|customer) receives an? "([^"]*)" notification$`, func(role, kind string) error { return n.receives(role, kind) })
	sc.Step(`^the customer has no notifications$`, func() error {
		p, err := n.feed("customer", "")
		if err != nil {
			return err
		}
		if p.Total != 0 || p.Items == nil {
			return fmt.Errorf("expected empty array, got %+v", p)
		}
		return nil
	})
	sc.Step(`^the customer bypasses payment$`, func() error { return n.bypass() })
	sc.Step(`^the customer repeats payment bypass$`, func() error {
		if err := n.rememberTotals(); err != nil {
			return err
		}
		return n.request("customer", "PUT", "/orders/"+n.orderID+"/bypass-payment", nil, 0)
	})
	sc.Step(`^the request fails without adding notifications$`, func() error {
		if n.resp.StatusCode < 400 {
			return fmt.Errorf("expected failed request, got %d", n.resp.StatusCode)
		}
		return n.unchanged()
	})
	sc.Step(`^the artist submits a draft$`, func() error { return n.submitDraft() })
	sc.Step(`^a draft is awaiting review$`, func() error {
		if err := n.orderWithStatus("IN_PROCESS"); err != nil {
			return err
		}
		return n.submitDraft()
	})
	sc.Step(`^the notification references the submitted draft and its version$`, func() error {
		if n.card.DeliverableID == nil || *n.card.DeliverableID != n.draftID || n.card.DeliverableVersion == nil || *n.card.DeliverableVersion != n.version {
			return fmt.Errorf("draft link/version lost: %+v", n.card)
		}
		return nil
	})
	sc.Step(`^the customer (accepts|rejects) the draft$`, func(decision string) error { return n.confirmDraft(decision) })
	sc.Step(`^the (artist|customer) cancels the order$`, func(role string) error {
		reason := "PERSONAL_OR_TECHNICAL_ISSUE"
		if role == "customer" {
			reason = "NO_LONGER_NEEDED"
		}
		return n.request(role, "PUT", "/orders/"+n.orderID+"/cancel", map[string]string{"reason": reason, "explanation": "Unable to continue"}, 200)
	})
	sc.Step(`^two drafts have been rejected$`, func() error {
		if err := n.orderWithStatus("IN_PROCESS"); err != nil {
			return err
		}
		for i := 0; i < 2; i++ {
			if err := n.submitDraft(); err != nil {
				return err
			}
			if err := n.confirmDraft("rejects"); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^the customer requests their notifications$`, func() error { var err error; n.page, err = n.feed("customer", ""); return err })
	sc.Step(`^the feed contains (\d+) items and total (\d+)$`, func(items, total int) error {
		if n.page.Items == nil || len(n.page.Items) != items || n.page.Total != total {
			return fmt.Errorf("feed: %+v, want %d items and total %d", n.page, items, total)
		}
		return nil
	})
	sc.Step(`^the customer marks the artist notification read$`, func() error {
		if err := n.receives("artist", "ORDER_CREATED"); err != nil {
			return err
		}
		return n.request("customer", "PATCH", "/notifications/"+n.card.ID+"/read", nil, 0)
	})
	sc.Step(`^the response status is (\d+)$`, func(want int) error {
		if n.resp == nil || n.resp.StatusCode != want {
			return fmt.Errorf("expected status %d, got %+v", want, n.resp)
		}
		return nil
	})
	sc.Step(`^the artist marks their notification read twice$`, func() error {
		if err := n.receives("artist", "ORDER_CREATED"); err != nil {
			return err
		}
		for i := 0; i < 2; i++ {
			if err := n.request("artist", "PATCH", "/notifications/"+n.card.ID+"/read", nil, 204); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^the artist has (\d+) unread notifications$`, func(want int) error {
		if err := n.request("artist", "GET", "/notifications/unread-count", nil, 200); err != nil {
			return err
		}
		var result struct {
			Count int `json:"count"`
		}
		if err := n.resp.JSON(&result); err != nil {
			return err
		}
		if result.Count != want {
			return fmt.Errorf("unread %d, want %d", result.Count, want)
		}
		return nil
	})
	sc.Step(`^the customer has created (\d+) orders$`, func(count int) error {
		for i := 0; i < count; i++ {
			if err := n.createOrder(); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^the artist requests notifications with limit (\d+) and offset (\d+)$`, func(limit, offset int) error {
		var err error
		n.page, err = n.feed("artist", fmt.Sprintf("?limit=%d&offset=%d", limit, offset))
		return err
	})
	sc.Step(`^a visitor requests "([^"]*)" using "([^"]*)"$`, func(path, method string) error {
		resp, err := apptest.NewClient(app.BaseURL()).Do(method, path, nil, nil)
		n.resp = resp
		return err
	})
	sc.Step(`^the artist requests a deadline extension$`, func() error { return n.requestExtension() })
	sc.Step(`^a deadline extension is awaiting a response$`, func() error { return n.requestExtension() })
	sc.Step(`^the notification references that extension and its proposed deadline$`, func() error {
		if n.card.DelayRequestID == nil || *n.card.DelayRequestID != n.requestID || n.card.DeadlineAt == nil || n.card.DeadlineAt.Sub(n.proposed).Abs() > time.Microsecond {
			return fmt.Errorf("extension link/date lost: %+v", n.card)
		}
		return nil
	})
	sc.Step(`^the customer (approves|rejects) the deadline extension$`, func(decision string) error { return n.confirmExtension(decision) })
	sc.Step(`^the artist and customer approve two extension rounds$`, func() error {
		for i := 0; i < 2; i++ {
			if err := n.requestExtension(); err != nil {
				return err
			}
			n.approvedDeadlines = append(n.approvedDeadlines, n.proposed)
			if err := n.confirmExtension("approves"); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^both approval notifications retain their own deadline$`, func() error {
		p, err := n.feed("artist", "")
		if err != nil {
			return err
		}
		var dates []time.Time
		for _, c := range p.Items {
			if c.Type == "DEADLINE_EXTENSION_APPROVED" && c.OrderID != nil && *c.OrderID == n.orderID && c.DeadlineAt != nil {
				dates = append(dates, *c.DeadlineAt)
			}
		}
		if len(dates) != 2 {
			return fmt.Errorf("expected two approval cards, got %v", dates)
		}
		for i := 0; i < 2; i++ {
			if dates[i].Sub(n.approvedDeadlines[1-i]).Abs() > time.Microsecond {
				return fmt.Errorf("historical deadlines changed: %v", dates)
			}
		}
		return nil
	})
	sc.Step(`^an expired "([^"]*)" exists$`, func(kind string) error {
		n.expiredKind = kind
		if kind == "order" {
			if err := n.orderWithStatus("PENDING"); err != nil {
				return err
			}
			return n.exec("UPDATE orders SET created_at=? WHERE id=?", time.Now().Add(-order.PendingConfirmationTTL-time.Hour), n.orderID)
		}
		if err := n.orderWithStatus("IN_PROCESS"); err != nil {
			return err
		}
		if kind == "draft" {
			if err := n.submitDraft(); err != nil {
				return err
			}
			return n.exec("UPDATE order_deliverables SET created_at=? WHERE id=?", time.Now().Add(-order.DeliverableDecisionTTL-time.Hour), n.draftID)
		}
		if kind != "delay request" {
			return fmt.Errorf("unknown expired fixture %s", kind)
		}
		if err := n.requestExtension(); err != nil {
			return err
		}
		return n.exec("UPDATE order_delay_requests SET created_at=? WHERE id=?", time.Now().Add(-order.DelayRequestTTL-time.Hour), n.requestID)
	})
	sc.Step(`^the expiry process runs$`, func() error { return n.runExpiry() })
	sc.Step(`^the expiry process runs again$`, func() error {
		if err := n.rememberTotals(); err != nil {
			return err
		}
		return n.runExpiry()
	})
	sc.Step(`^no extra notifications are created$`, func() error { return n.unchanged() })
	sc.Step(`^notification storage fails during order acceptance$`, func() error {
		if err := n.rememberTotals(); err != nil {
			return err
		}
		_, err := orderActions(failingNotifications{}).ConfirmOrder(context.Background(), uuid.MustParse(n.accounts["artist"].ID), uuid.MustParse(n.orderID), order.ConfirmOrderInput{Accept: true})
		if err == nil {
			return fmt.Errorf("expected notification failure")
		}
		return nil
	})
	sc.Step(`^notification storage fails for the second worker recipient$`, func() error {
		if err := n.rememberTotals(); err != nil {
			return err
		}
		_, err := orderActions(&failSecondNotification{creator: notificationActions()}).CancelExpiredOrders(context.Background())
		if err == nil {
			return fmt.Errorf("expected second recipient failure")
		}
		return nil
	})
	sc.Step(`^the order remains pending without new notifications$`, func() error {
		var status string
		if err := app.DB.NewSelect().TableExpr("orders").Column("status").Where("id=?", n.orderID).Scan(context.Background(), &status); err != nil {
			return err
		}
		if status != "PENDING" {
			return fmt.Errorf("order status %s, expected rollback to PENDING", status)
		}
		return n.unchanged()
	})
	sc.Step(`^a payment ledger entry is recorded$`, func() error {
		if err := n.rememberTotals(); err != nil {
			return err
		}
		paymentID := uuid.New()
		if err := n.exec("INSERT INTO payments(id,order_id,amount_satang,status) VALUES (?,?,10000,'HELD')", paymentID, n.orderID); err != nil {
			return err
		}
		return n.exec("INSERT INTO payment_transactions(id,payment_id,type,amount_satang) VALUES (?,?,'CHARGE',10000)", uuid.New(), paymentID)
	})
}
