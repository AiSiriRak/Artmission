//go:build integration

package reviews

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func InitializeScenario(sc *godog.ScenarioContext) {
	var c *reviewsContext
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		c = &reviewsContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})
	sc.Step(`^a commission in status "([^"]+)"$`, func(status string) error {
		return c.prepareCommission(status)
	})
	sc.Step(`^the customer submits this review:$`, func(body *godog.DocString) error {
		var payload any
		if err := json.Unmarshal([]byte(body.Content), &payload); err != nil {
			return err
		}
		return c.submit(c.orderID, payload, c.customerToken)
	})
	sc.Step(`^the review is saved with rating (\d+) and comment "([^"]*)"$`, func(rating int, comment string) error {
		if err := c.requireStatus(http.StatusCreated); err != nil {
			return err
		}
		var result reviewResponse
		if err := c.response.JSON(&result); err != nil {
			return err
		}
		if result.ID == uuid.Nil || result.OrderID != c.orderID || result.CustomerID.String() != c.customer.ID || result.ArtistID.String() != c.artist.ID || result.Rating != rating || result.CreatedAt.IsZero() {
			return fmt.Errorf("unexpected created review: %+v", result)
		}
		if err := checkComment(result.Comment, comment); err != nil {
			return err
		}
		var stored reviewResponse
		if err := app.DB.NewSelect().Table("reviews").Where("order_id = ?", c.orderID).Scan(context.Background(), &stored); err != nil {
			return err
		}
		if stored.ID != result.ID || stored.OrderID != c.orderID || stored.CustomerID != result.CustomerID || stored.ArtistID != result.ArtistID || stored.Rating != rating || stored.CreatedAt.Sub(result.CreatedAt).Abs() > time.Microsecond {
			return fmt.Errorf("database review does not match response: %+v", stored)
		}
		return checkComment(stored.Comment, comment)
	})
	sc.Step(`^the artist average score is ([0-9.]+)$`, func(want float64) error {
		return c.requireAverage(&want)
	})
	sc.Step(`^the review is rejected as "([^"]+)"$`, func(reason string) error {
		statuses := map[string]int{
			"invalid request":       http.StatusUnprocessableEntity,
			"unfinished commission": http.StatusConflict,
			"duplicate review":      http.StatusConflict,
			"forbidden":             http.StatusForbidden,
			"unauthenticated":       http.StatusUnauthorized,
			"missing order":         http.StatusNotFound,
		}
		status, ok := statuses[reason]
		if !ok {
			return fmt.Errorf("unknown rejection reason %q", reason)
		}
		if err := c.requireStatus(status); err != nil {
			return err
		}
		count, err := c.countOrderReviews()
		if err != nil {
			return err
		}
		if count != c.existingReviews {
			return fmt.Errorf("a rejected request changed the review count: got %d, want %d", count, c.existingReviews)
		}
		return nil
	})
	sc.Step(`^"([^"]+)" submits a review$`, func(caller string) error {
		token, err := c.tokenFor(caller)
		if err != nil {
			return err
		}
		return c.submit(c.orderID, map[string]any{"rating": 5}, token)
	})
	sc.Step(`^the customer reviews an order that does not exist$`, func() error {
		return c.submit(uuid.New(), map[string]any{"rating": 5}, c.customerToken)
	})
	sc.Step(`^the customer has already reviewed the commission$`, func() error {
		if err := c.submit(c.orderID, map[string]any{"rating": 5, "comment": "Original review"}, c.customerToken); err != nil {
			return err
		}
		if err := c.requireStatus(http.StatusCreated); err != nil {
			return err
		}
		c.existingReviews = 1
		return nil
	})
	sc.Step(`^the original review remains unchanged$`, func() error {
		var row reviewResponse
		if err := app.DB.NewSelect().Table("reviews").Where("order_id = ?", c.orderID).Scan(context.Background(), &row); err != nil {
			return err
		}
		if row.Rating != 5 {
			return fmt.Errorf("duplicate submission replaced the rating")
		}
		return checkComment(row.Comment, "Original review")
	})
	sc.Step(`^the customer submits the same review concurrently$`, func() error {
		results := make(chan submissionResult, 2)
		for range 2 {
			go func() {
				client := apptest.NewClient(app.BaseURL())
				response, err := client.Do(http.MethodPost, "/orders/"+c.orderID.String()+"/review", map[string]any{"rating": 5}, map[string]string{"Authorization": "Bearer " + c.customerToken})
				results <- submissionResult{response: response, err: err}
			}()
		}
		for range 2 {
			c.concurrent = append(c.concurrent, <-results)
		}
		return nil
	})
	sc.Step(`^only one review is saved$`, func() error {
		created, conflicted := 0, 0
		for _, result := range c.concurrent {
			if result.err != nil {
				return result.err
			}
			switch result.response.StatusCode {
			case http.StatusCreated:
				created++
			case http.StatusConflict:
				conflicted++
			default:
				return fmt.Errorf("unexpected concurrent response %d: %s", result.response.StatusCode, result.response.Body)
			}
		}
		count, err := c.countOrderReviews()
		if err != nil {
			return err
		}
		if created != 1 || conflicted != 1 || count != 1 {
			return fmt.Errorf("expected one created review and one conflict, got created=%d conflicts=%d rows=%d", created, conflicted, count)
		}
		want := 5.0
		return c.requireAverage(&want)
	})
	sc.Step(`^the customer reviews three completed commissions with ratings 5, 4, and 4$`, func() error {
		for _, rating := range []int{5, 4, 4} {
			orderID, err := c.createOrder("SUCCESS", time.Now().UTC())
			if err != nil {
				return err
			}
			if err := c.submit(orderID, map[string]any{"rating": rating}, c.customerToken); err != nil {
				return err
			}
			if err := c.requireStatus(http.StatusCreated); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^the artist has reviews whose order dates differ from their review dates$`, func() error {
		return c.prepareReviewHistory(false)
	})
	sc.Step(`^the artist has reviews with identical timestamps$`, func() error {
		return c.prepareReviewHistory(true)
	})
	sc.Step(`^"([^"]+)" requests the artist reviews$`, func(caller string) error {
		token, err := c.tokenFor(caller)
		if err != nil {
			return err
		}
		return c.listReviews(token)
	})
	sc.Step(`^the artist reviews contain the customer and purchased artwork names$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("expected review details")
		}
		for _, item := range items {
			if item.Username != c.customer.Username || item.Order != "Purchased artwork name" || item.Rating != 5 || item.Comment == nil || *item.Comment != "Original review" || item.CreatedAt.IsZero() || item.ArtistID.String() != c.artist.ID {
				return fmt.Errorf("incorrect review details: %+v", item)
			}
		}
		return nil
	})
	sc.Step(`^the artist reviews are returned newest first$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if len(items) != len(c.expectedIDs) {
			return fmt.Errorf("expected %d reviews, got %d", len(c.expectedIDs), len(items))
		}
		for i, item := range items {
			if item.ID != c.expectedIDs[i] || !item.CreatedAt.Equal(c.expectedDates[i]) {
				return fmt.Errorf("review %d is out of order: got %s at %s, want %s at %s", i, item.ID, item.CreatedAt, c.expectedIDs[i], c.expectedDates[i])
			}
		}
		return nil
	})
	sc.Step(`^reviews from other artists are excluded$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if len(items) != 3 {
			return fmt.Errorf("expected only this artist's three reviews, got %d", len(items))
		}
		for _, item := range items {
			if item.ArtistID.String() != c.artist.ID {
				return fmt.Errorf("a review belonging to another artist was returned")
			}
		}
		return nil
	})
	sc.Step(`^the reviewed artwork has been deleted$`, func() error {
		_, err := app.DB.ExecContext(context.Background(), "DELETE FROM artworks WHERE id = ?", c.artworkID)
		return err
	})
	sc.Step(`^the system returns an empty review list and no average score$`, func() error {
		items, err := c.readReviewList()
		if err != nil {
			return err
		}
		if items == nil || len(items) != 0 {
			return fmt.Errorf("expected reviews: [], got %+v", items)
		}
		return c.requireAverage(nil)
	})
	sc.Step(`^the artist review list requires authentication$`, func() error {
		return c.requireStatus(http.StatusUnauthorized)
	})
}

func checkComment(got *string, want string) error {
	if want == "NULL" {
		if got != nil {
			return fmt.Errorf("expected a NULL comment, got %q", *got)
		}
	} else if got == nil || *got != want {
		return fmt.Errorf("expected comment %q, got %v", want, got)
	}
	return nil
}

func (c *reviewsContext) tokenFor(caller string) (string, error) {
	switch caller {
	case "the customer":
		return c.customerToken, nil
	case "the artist":
		return c.artistToken, nil
	case "a visitor":
		return "", nil
	case "an expired token":
		return "invalid-or-expired-token", nil
	case "another customer":
		other, err := apptest.RegisterCustomer(app, c.client)
		if err != nil {
			return "", err
		}
		return apptest.Login(c.client, other.Email, other.Password)
	default:
		return "", fmt.Errorf("unknown caller %q", caller)
	}
}

func (c *reviewsContext) readReviewList() ([]reviewResponse, error) {
	if err := c.requireStatus(http.StatusOK); err != nil {
		return nil, err
	}
	var body struct {
		Reviews []reviewResponse `json:"reviews"`
	}
	if err := c.response.JSON(&body); err != nil {
		return nil, err
	}
	return body.Reviews, nil
}

// prepareReviewHistory makes order chronology disagree with review chronology.
func (c *reviewsContext) prepareReviewHistory(equalTimes bool) error {
	base := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	dates := []time.Time{base, base.Add(2 * time.Hour), base.Add(time.Hour)}
	if equalTimes {
		ids = []uuid.UUID{
			uuid.MustParse("00000000-0000-0000-0000-000000000010"),
			uuid.MustParse("00000000-0000-0000-0000-000000000030"),
			uuid.MustParse("00000000-0000-0000-0000-000000000020"),
		}
		dates = []time.Time{base, base, base}
	}
	for i := range ids {
		orderID, err := c.createOrder("SUCCESS", base.Add(time.Duration(3-i)*time.Hour))
		if err != nil {
			return err
		}
		if err := c.submit(orderID, map[string]any{"rating": 5, "comment": "Original review"}, c.customerToken); err != nil {
			return err
		}
		if err := c.requireStatus(http.StatusCreated); err != nil {
			return err
		}
		// Fixed review times avoid sleeps and make timestamp ties reproducible.
		if _, err := app.DB.ExecContext(context.Background(), "UPDATE reviews SET id = ?, created_at = ? WHERE order_id = ?", ids[i], dates[i], orderID); err != nil {
			return err
		}
	}
	c.expectedIDs = []uuid.UUID{ids[1], ids[2], ids[0]}
	c.expectedDates = []time.Time{dates[1], dates[2], dates[0]}

	other := &reviewsContext{client: apptest.NewClient(app.BaseURL())}
	if err := other.prepareCommission("SUCCESS"); err != nil {
		return err
	}
	if err := other.submit(other.orderID, map[string]any{"rating": 1}, other.customerToken); err != nil {
		return err
	}
	return other.requireStatus(http.StatusCreated)
}
