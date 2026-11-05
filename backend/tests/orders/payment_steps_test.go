//go:build integration

package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func TestPaymentSettlementFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	s3 := apptest.StartObjectStorage(ctx, t)
	app = apptest.NewAppWithS3(t, pg.DSN, s3)
	suite := godog.TestSuite{ScenarioInitializer: InitializeScenario, Options: &godog.Options{
		Format: "pretty,cucumber:../../reports/json/cucumber-payment-settlement.json",
		Paths:  []string{"features/payment_settlement.feature"}, TestingT: t,
	}}
	if suite.Run() != 0 {
		t.Fatal("payment settlement BDD scenarios failed")
	}
}

func registerPaymentSteps(sc *godog.ScenarioContext, current func() *ordersContext) {
	sc.Step(`^the last order has a "(HELD|PENDING)" payment of (\d+) satang$`, func(status string, amount int64) error {
		o := current()
		_, err := app.DB.NewRaw(`INSERT INTO payments (id, order_id, amount_satang, status) VALUES (?, ?, ?, ?)`, uuid.New(), o.lastOrderID, amount, status).Exec(context.Background())
		return err
	})
	sc.Step(`^the last order summary shows "([^"]*)" with refund (\d+) and artist payout (\d+) satang$`, func(outcome string, refund, artist int64) error {
		o := current()
		resp, err := o.client.Do(http.MethodGet, "/orders", nil, map[string]string{"Authorization": "Bearer " + o.accessToken})
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("summary status %d: %s", resp.StatusCode, resp.Body)
		}
		var body struct {
			Orders []struct {
				ID      string `json:"id"`
				Payment *struct {
					Status       string  `json:"payment_status"`
					Outcome      *string `json:"settlement_outcome"`
					Refund       int64   `json:"refund_satang"`
					Artist       int64   `json:"artist_payout_satang"`
					RefundStatus *string `json:"refund_transfer_status"`
					ArtistStatus *string `json:"artist_payout_transfer_status"`
				} `json:"payment"`
			} `json:"orders"`
		}
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			return err
		}
		for _, item := range body.Orders {
			if item.ID != o.lastOrderID {
				continue
			}
			p := item.Payment
			if p == nil || p.Refund != refund || p.Artist != artist {
				return fmt.Errorf("wrong allocation: %s", resp.Body)
			}
			if outcome == "UNPAID" {
				if p.Status != "CANCELLED" || p.Outcome != nil || p.RefundStatus != nil || p.ArtistStatus != nil {
					return fmt.Errorf("pending payment was settled: %+v", p)
				}
				return nil
			}
			if p.Status != "SETTLED" || p.Outcome == nil || *p.Outcome != outcome {
				return fmt.Errorf("wrong settlement: %+v", p)
			}
			if refund > 0 && (p.RefundStatus == nil || *p.RefundStatus != "SUCCESS") {
				return fmt.Errorf("refund was not credited")
			}
			if artist > 0 && (p.ArtistStatus == nil || *p.ArtistStatus != "SUCCESS") {
				return fmt.Errorf("payout was not credited")
			}
			return nil
		}
		return fmt.Errorf("order missing from summary")
	})
	sc.Step(`^the user's wallet shows (\d+) satang and (\d+) incoming transactions$`, func(balance int64, count int) error {
		o := current()
		headers := map[string]string{"Authorization": "Bearer " + o.accessToken}
		resp, err := o.client.Do(http.MethodGet, "/wallet/me", nil, headers)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("wallet: %d %s", resp.StatusCode, resp.Body)
		}
		var body struct {
			Balance int64 `json:"balance_satang"`
		}
		if err := resp.JSON(&body); err != nil {
			return err
		}
		if body.Balance != balance {
			return fmt.Errorf("unexpected wallet: %s", resp.Body)
		}
		resp, err = o.client.Do(http.MethodGet, "/transaction/me", nil, headers)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("transactions: %d %s", resp.StatusCode, resp.Body)
		}
		var history struct {
			Total        int `json:"total"`
			Transactions []struct {
				OrderID      string `json:"order_id"`
				AmountSatang int64  `json:"amount_satang"`
			} `json:"transactions"`
		}
		if err := resp.JSON(&history); err != nil {
			return err
		}
		if history.Total != count || len(history.Transactions) != count {
			return fmt.Errorf("unexpected history: %s", resp.Body)
		}
		if count == 1 && (history.Transactions[0].OrderID != o.lastOrderID || history.Transactions[0].AmountSatang != balance) {
			return fmt.Errorf("wrong wallet recipient: %s", resp.Body)
		}
		return nil
	})
}
