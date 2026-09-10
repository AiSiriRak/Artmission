//go:build integration

package orders

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

// ordersContext holds the state for exactly one scenario: the HTTP client
// (its own cookie jar, isolated from every other scenario), the account
// under test, and whatever fixtures its Given steps produced.
type ordersContext struct {
	client *apptest.Client

	account     apptest.Account
	accountRole string // "customer" or "artist" — which side of orders o.account sits on
	accessToken string

	counterpart apptest.Account // shared, lazily-created FK backing every order seeded for o.account

	seededOrders map[string]string // order ID → seeded status, for o.account
	traversed    map[string]bool   // order IDs seen while paging through every page
	firstPageIDs []string          // remembered first page, for later-offset assertions

	resp *apptest.Response
	page orderViewBody // last response decoded while resp.StatusCode == 200
}

// --- given ---

func (o *ordersContext) theUserHasARegisteredCustomerAccount() error {
	account, err := apptest.RegisterCustomer(app, o.client)
	if err != nil {
		return err
	}
	o.account = account
	o.accountRole = "customer"
	return nil
}

func (o *ordersContext) theUserHasARegisteredArtistAccount() error {
	account, err := apptest.RegisterArtist(app, o.client, "I paint custom portraits.")
	if err != nil {
		return err
	}
	o.account = account
	o.accountRole = "artist"
	return nil
}

func (o *ordersContext) theUserHasLoggedIn() error {
	token, err := apptest.Login(o.client, o.account.Email, o.account.Password)
	if err != nil {
		return err
	}
	o.accessToken = token
	return nil
}

func (o *ordersContext) theUserHasOneOrMoreOrders() error {
	counterpart, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	for range 2 {
		id, err := o.seedForAccount(orderSeed{}, counterpart)
		if err != nil {
			return err
		}
		o.seededOrders[id] = "PENDING"
	}
	return nil
}

func (o *ordersContext) theUserHasAnOrder() error {
	counterpart, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	id, err := o.seedForAccount(orderSeed{}, counterpart)
	if err != nil {
		return err
	}
	o.seededOrders[id] = "PENDING"
	return nil
}

func (o *ordersContext) theUserHasAnOrderWithStatus(status string) error {
	counterpart, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	id, err := o.seedForAccount(orderSeed{Status: status}, counterpart)
	if err != nil {
		return err
	}
	o.seededOrders[id] = status
	return nil
}

// theUserHasNOrders seeds n orders for o.account with strictly increasing
// UpdatedAt values — deterministic under ViewOrders' default
// updated_at-desc sort, so pagination assertions can rely on exact order.
func (o *ordersContext) theUserHasNOrders(n int) error {
	counterpart, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	base := time.Now().Add(-time.Duration(n+1) * time.Minute)
	for i := range n {
		id, err := o.seedForAccount(orderSeed{UpdatedAt: base.Add(time.Duration(i) * time.Minute)}, counterpart)
		if err != nil {
			return err
		}
		o.seededOrders[id] = "PENDING"
	}
	return nil
}

// theUserHasAnOrderWithDeadline seeds an order whose deadline is parsed
// from an RFC3339 string, or nil when deadline is "" — used to prove the
// deadline sort's null-bucket placement/tie-breaking against real Postgres.
func (o *ordersContext) theUserHasAnOrderWithDeadline(deadline string) error {
	counterpart, err := o.sharedCounterpart()
	if err != nil {
		return err
	}

	var parsed *time.Time
	if deadline != "" {
		t, err := time.Parse(time.RFC3339, deadline)
		if err != nil {
			return fmt.Errorf("parse deadline %q: %w", deadline, err)
		}
		parsed = &t
	}

	id, err := o.seedForAccount(orderSeed{DeadlineAt: parsed}, counterpart)
	if err != nil {
		return err
	}
	o.seededOrders[id] = "PENDING"
	return nil
}

func (o *ordersContext) anotherCustomerHasAnOrder() error {
	other, err := apptest.RegisterCustomer(app, apptest.NewClient(app.BaseURL()))
	if err != nil {
		return err
	}
	artist, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	_, err = seedOrder(orderSeed{CustomerID: other.ID, ArtistID: artist.ID})
	return err
}

func (o *ordersContext) anotherArtistHasAnOrder() error {
	other, err := apptest.RegisterArtist(app, apptest.NewClient(app.BaseURL()), "Another artist.")
	if err != nil {
		return err
	}
	customer, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	_, err = seedOrder(orderSeed{CustomerID: customer.ID, ArtistID: other.ID})
	return err
}

// --- when ---

// getOrders sends GET /orders?<values> as the logged-in user and stores the
// raw response. It decodes into o.page only on a 200: assertions on a
// non-200 response (401/400/...) read o.resp directly.
func (o *ordersContext) getOrders(values url.Values) error {
	path := "/orders"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	resp, err := o.client.Do(http.MethodGet, path, nil, map[string]string{
		"Authorization": "Bearer " + o.accessToken,
	})
	if err != nil {
		return err
	}
	o.resp = resp
	o.page = orderViewBody{}
	if resp.StatusCode == http.StatusOK {
		if err := resp.JSON(&o.page); err != nil {
			return fmt.Errorf("decode orders response: %w (body: %s)", err, resp.Body)
		}
	}
	return nil
}

func (o *ordersContext) theUserViewsTheirOrders() error {
	return o.getOrders(url.Values{})
}

func (o *ordersContext) theUserViewsTheirOrdersWithoutLoggingIn() error {
	resp, err := o.client.Do(http.MethodGet, "/orders", nil, nil)
	if err != nil {
		return err
	}
	o.resp = resp
	return nil
}

func (o *ordersContext) theUserViewsTheirOrdersFilteredByStatus(status string) error {
	return o.getOrders(url.Values{"status": {status}})
}

func (o *ordersContext) theUserViewsTheirOrdersWithALimitAndOffsetOf(limit, offset int) error {
	return o.getOrders(url.Values{"limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}})
}

func (o *ordersContext) theUserViewsTheirOrdersSortedByDeadline(order string) error {
	return o.getOrders(url.Values{"sort": {"deadline"}, "order": {order}})
}

func (o *ordersContext) theUserViewsTheirOrdersWithAnInvalidOffset() error {
	return o.getOrders(url.Values{"offset": {"-1"}})
}

func (o *ordersContext) theUserRemembersTheCurrentPageAsTheFirstPage() error {
	o.firstPageIDs = idsOf(o.page.Orders)
	return nil
}

// theUserPagesThroughAllOfTheirOrdersUsingALimit walks forward through
// every page via increasing offsets, using the running total of orders
// seen so far as the next offset, until a page comes back short of a
// full limit — the end-to-end proof that forward traversal with this
// limit visits every seeded order without duplication or omission.
func (o *ordersContext) theUserPagesThroughAllOfTheirOrdersUsingALimit(limit int) error {
	seen := map[string]bool{}
	offset := 0

	for {
		values := url.Values{"limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}
		if err := o.getOrders(values); err != nil {
			return err
		}
		if o.resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected response status 200 while paging, got %d: %s", o.resp.StatusCode, o.resp.Body)
		}
		for _, ord := range o.page.Orders {
			if seen[ord.ID] {
				return fmt.Errorf("order %s was returned more than once while paging forward", ord.ID)
			}
			seen[ord.ID] = true
		}
		offset += len(o.page.Orders)
		if len(o.page.Orders) < limit || offset >= o.page.Total {
			break
		}
	}

	o.traversed = seen
	return nil
}

// --- then ---

type orderViewItem struct {
	ID                   string  `json:"id"`
	CustomerID           string  `json:"customer_id"`
	ArtistID             string  `json:"artist_id"`
	ArtworkName          string  `json:"artwork_name"`
	ArtworkDescription   string  `json:"artwork_description"`
	PriceSatang          int64   `json:"price_satang"`
	MinimumDeadlineDays  int     `json:"minimum_deadline_days"`
	PreviewImageURL      string  `json:"preview_image_url"`
	CustomerDescription  string  `json:"customer_description"`
	SelectedDeadlineDays int     `json:"selected_deadline_days"`
	Status               string  `json:"status"`
	DeadlineAt           *string `json:"deadline_at"`
	Deliverables         []any   `json:"deliverables"`
}

type orderViewBody struct {
	Orders []orderViewItem `json:"orders"`
	Total  int             `json:"total"`
}

func idsOf(orders []orderViewItem) []string {
	ids := make([]string, len(orders))
	for i, o := range orders {
		ids[i] = o.ID
	}
	return ids
}

func (o *ordersContext) theSystemShowsAllOfTheUsersOrdersWithTheirCurrentStatus() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}

	got := map[string]string{}
	for _, ord := range body.Orders {
		if ord.Status == "" {
			return fmt.Errorf("expected every order to have a status, got: %+v", ord)
		}
		if ord.ArtworkName != seedArtworkName ||
			ord.ArtworkDescription != seedArtworkDescription ||
			ord.PriceSatang != seedPriceSatang ||
			ord.MinimumDeadlineDays != seedMinimumDeadlineDays ||
			ord.PreviewImageURL != seedPreviewImageURL ||
			ord.CustomerDescription != seedCustomerDescription ||
			ord.SelectedDeadlineDays != seedSelectedDeadlineDays {
			return fmt.Errorf("order %s did not preserve its artwork and customer snapshots: %+v", ord.ID, ord)
		}
		if ord.Deliverables == nil {
			return fmt.Errorf("order %s: expected deliverables to be an array", ord.ID)
		}
		got[ord.ID] = ord.Status
	}
	for wantID, wantStatus := range o.seededOrders {
		gotStatus, ok := got[wantID]
		if !ok {
			return fmt.Errorf("expected seeded order %s in the user's orders, got orders: %+v", wantID, body.Orders)
		}
		if gotStatus != wantStatus {
			return fmt.Errorf("order %s: expected status %q, got %q", wantID, wantStatus, gotStatus)
		}
	}
	return nil
}

func (o *ordersContext) theSystemDoesNotShowAnyOtherUsersOrders() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}

	if len(body.Orders) != len(o.seededOrders) {
		return fmt.Errorf("expected exactly the user's own %d order(s), got %d: %+v", len(o.seededOrders), len(body.Orders), body.Orders)
	}
	for _, ord := range body.Orders {
		if _, ok := o.seededOrders[ord.ID]; !ok {
			return fmt.Errorf("orders returned an order %s that does not belong to the user", ord.ID)
		}
	}
	return nil
}

func (o *ordersContext) theSystemShowsAnEmptyOrderList() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	if len(body.Orders) != 0 {
		return fmt.Errorf("expected an empty order list, got %d orders", len(body.Orders))
	}
	if body.Total != 0 {
		return fmt.Errorf("expected a total of 0 on an empty page, got %d", body.Total)
	}
	return nil
}

func (o *ordersContext) theSystemReturnsOrdersSortedByDeadline(order string) error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	return assertDeadlineOrder(body.Orders, order == "asc")
}

// assertDeadlineOrder proves Postgres's default null placement for a
// plain, un-COALESCE'd o.deadline_at column (see
// internal/adapters/postgres/order_repository.go's orderSortColumns)
// against real Postgres: values monotonic in the requested direction, and
// every null-deadline order grouped at the correct end (last for asc,
// first for desc) rather than interleaved or silently dropped.
func assertDeadlineOrder(orders []orderViewItem, ascending bool) error {
	nullsFirst := !ascending
	sawOppositeBucket := false
	for i, ord := range orders {
		isNull := ord.DeadlineAt == nil
		if isNull == nullsFirst {
			if sawOppositeBucket {
				return fmt.Errorf("order at index %d (deadline_at=%v) is out of null-bucket order: %+v", i, ord.DeadlineAt, orders)
			}
			continue
		}
		sawOppositeBucket = true
	}

	var prev *time.Time
	for _, ord := range orders {
		if ord.DeadlineAt == nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, *ord.DeadlineAt)
		if err != nil {
			return fmt.Errorf("parse returned deadline_at %q: %w", *ord.DeadlineAt, err)
		}
		if prev != nil {
			if ascending && prev.After(t) {
				return fmt.Errorf("expected ascending deadline order, got %s before %s", prev, t)
			}
			if !ascending && prev.Before(t) {
				return fmt.Errorf("expected descending deadline order, got %s before %s", prev, t)
			}
		}
		prev = &t
	}
	return nil
}

func (o *ordersContext) theSystemShowsOnlyOrdersWithStatus(status string) error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	if len(body.Orders) == 0 {
		return fmt.Errorf("expected at least one order with status %q, got none", status)
	}
	for _, ord := range body.Orders {
		if ord.Status != status {
			return fmt.Errorf("expected only status %q, got order %s with status %q", status, ord.ID, ord.Status)
		}
	}
	return nil
}

func (o *ordersContext) theSystemReturnsEveryOrderExactlyOnce() error {
	if len(o.traversed) != len(o.seededOrders) {
		return fmt.Errorf("traversed %d distinct orders, want exactly the %d seeded", len(o.traversed), len(o.seededOrders))
	}
	for id := range o.seededOrders {
		if !o.traversed[id] {
			return fmt.Errorf("expected seeded order %s while paging through every order, it never appeared", id)
		}
	}
	return nil
}

func (o *ordersContext) theSystemReturnsTheFirstPageOfOrdersAgain() error {
	got := idsOf(o.page.Orders)
	if len(got) != len(o.firstPageIDs) {
		return fmt.Errorf("got %d orders, want %d matching the remembered first page", len(got), len(o.firstPageIDs))
	}
	for i := range got {
		if got[i] != o.firstPageIDs[i] {
			return fmt.Errorf("order[%d] = %s, want %s (remembered first-page order)", i, got[i], o.firstPageIDs[i])
		}
	}
	return nil
}

func (o *ordersContext) theSystemReportsATotalOfNOrders(n int) error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	if body.Total != n {
		return fmt.Errorf("expected total %d, got %d", n, body.Total)
	}
	return nil
}

func (o *ordersContext) theSystemRequiresTheUserToLogIn() error {
	return o.expectStatus(http.StatusUnauthorized)
}

func (o *ordersContext) decodeOrders() (orderViewBody, error) {
	if o.resp.StatusCode != http.StatusOK {
		return orderViewBody{}, fmt.Errorf("expected response status 200, got %d: %s", o.resp.StatusCode, o.resp.Body)
	}
	return o.page, nil
}

func (o *ordersContext) expectStatus(code int) error {
	if o.resp.StatusCode != code {
		return fmt.Errorf("expected response status %d, got %d: %s", code, o.resp.StatusCode, o.resp.Body)
	}
	return nil
}

// expectClientError checks for any 4xx: the business requirement is "the
// system refused and said why," not which specific validation layer fired
// (huma's own schema/enum validation vs. mapAppError's apperror.CodeInvalidInput
// are both valid reasons a request can be refused) — mirrors
// tests/auth's authContext.expectClientError.
func (o *ordersContext) expectClientError() error {
	if o.resp.StatusCode < 400 || o.resp.StatusCode >= 500 {
		return fmt.Errorf("expected a 4xx client error, got %d: %s", o.resp.StatusCode, o.resp.Body)
	}
	return nil
}

// InitializeScenario registers every ViewOrders step and resets the
// scenario context (a fresh client with an empty cookie jar) before each
// scenario, so scenarios never see each other's cookies or fixtures.
func InitializeScenario(sc *godog.ScenarioContext) {
	var o *ordersContext

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		o = &ordersContext{
			client:       apptest.NewClient(app.BaseURL()),
			seededOrders: map[string]string{},
		}
		return ctx, nil
	})

	// given
	sc.Step(`^the user has a registered customer account$`, func() error { return o.theUserHasARegisteredCustomerAccount() })
	sc.Step(`^the user has a registered artist account$`, func() error { return o.theUserHasARegisteredArtistAccount() })
	sc.Step(`^the user has logged in$`, func() error { return o.theUserHasLoggedIn() })
	sc.Step(`^the user has one or more orders$`, func() error { return o.theUserHasOneOrMoreOrders() })
	sc.Step(`^the user has an order$`, func() error { return o.theUserHasAnOrder() })
	sc.Step(`^the user has an order with status "([^"]*)"$`, func(status string) error { return o.theUserHasAnOrderWithStatus(status) })
	sc.Step(`^the user has an order with deadline "([^"]*)"$`, func(deadline string) error { return o.theUserHasAnOrderWithDeadline(deadline) })
	sc.Step(`^the user has an order with no deadline$`, func() error { return o.theUserHasAnOrderWithDeadline("") })
	sc.Step(`^the user has (\d+) orders$`, func(n int) error { return o.theUserHasNOrders(n) })
	sc.Step(`^another customer has an order$`, func() error { return o.anotherCustomerHasAnOrder() })
	sc.Step(`^another artist has an order$`, func() error { return o.anotherArtistHasAnOrder() })

	// when
	sc.Step(`^the user views their orders$`, func() error { return o.theUserViewsTheirOrders() })
	sc.Step(`^the user views their orders without logging in$`, func() error { return o.theUserViewsTheirOrdersWithoutLoggingIn() })
	sc.Step(`^the user views their orders filtered by status "([^"]*)"$`, func(status string) error {
		return o.theUserViewsTheirOrdersFilteredByStatus(status)
	})
	sc.Step(`^the user views their orders with a limit of (\d+) and an offset of (\d+)$`, func(limit, offset int) error {
		return o.theUserViewsTheirOrdersWithALimitAndOffsetOf(limit, offset)
	})
	sc.Step(`^the user views their orders sorted by deadline in "([^"]*)" order$`, func(order string) error {
		return o.theUserViewsTheirOrdersSortedByDeadline(order)
	})
	sc.Step(`^the user views their orders with an invalid offset$`, func() error { return o.theUserViewsTheirOrdersWithAnInvalidOffset() })
	sc.Step(`^the user remembers the current page as the first page$`, func() error {
		return o.theUserRemembersTheCurrentPageAsTheFirstPage()
	})
	sc.Step(`^the user pages through all of their orders using a limit of (\d+)$`, func(limit int) error {
		return o.theUserPagesThroughAllOfTheirOrdersUsingALimit(limit)
	})

	// then
	sc.Step(`^the system shows all of the user's orders with their current status$`, func() error {
		return o.theSystemShowsAllOfTheUsersOrdersWithTheirCurrentStatus()
	})
	sc.Step(`^the system does not show any other user's orders$`, func() error { return o.theSystemDoesNotShowAnyOtherUsersOrders() })
	sc.Step(`^the system shows an empty order list$`, func() error { return o.theSystemShowsAnEmptyOrderList() })
	sc.Step(`^the system shows only orders with status "([^"]*)"$`, func(status string) error {
		return o.theSystemShowsOnlyOrdersWithStatus(status)
	})
	sc.Step(`^the system returns every seeded order exactly once$`, func() error { return o.theSystemReturnsEveryOrderExactlyOnce() })
	sc.Step(`^the system returns orders sorted by deadline in "([^"]*)" order$`, func(order string) error {
		return o.theSystemReturnsOrdersSortedByDeadline(order)
	})
	sc.Step(`^the system returns the first page of orders again$`, func() error { return o.theSystemReturnsTheFirstPageOfOrdersAgain() })
	sc.Step(`^the system reports a total of (\d+) orders$`, func(n int) error { return o.theSystemReportsATotalOfNOrders(n) })
	sc.Step(`^the system requires the user to log in$`, func() error { return o.theSystemRequiresTheUserToLogIn() })
	sc.Step(`^the system rejects the request due to an invalid offset$`, func() error { return o.expectClientError() })
	sc.Step(`^the system rejects the request due to an invalid status$`, func() error { return o.expectClientError() })
}
