//go:build integration

package orders

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
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

	lastOrderID               string // most recently seeded order's ID, for single-order deliverable-preview assertions
	otherOrderID              string // order ID seeded for an account other than o.account
	lastDeliverablePreviewKey string // most recently seeded deliverable version's preview_image_key
	targetArtworkID           string

	resp        *apptest.Response
	page        viewOrdersOutput // last response decoded while resp.StatusCode == 200
	orderDetail orderDetailOutput
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
	o.lastOrderID = id
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
	o.lastOrderID = id
	return nil
}

// theUserHasAnOrderWithStatusAndNSubmittedDeliverableVersions seeds one
// order at status, then n order_deliverables rows for it at versions
// 1..n with deterministic preview keys — proving ViewOrders resolves the
// highest version's preview regardless of the order's status (including
// terminal ones like CANCEL). Every version before the last is seeded
// REJECTED (the order_deliverables_order_id_pending_key unique index
// allows at most one WAIT row per order); only the last version is
// left pending (decision WAIT).
func (o *ordersContext) theUserHasAnOrderWithStatusAndNSubmittedDeliverableVersions(status string, n int) error {
	counterpart, err := o.sharedCounterpart()
	if err != nil {
		return err
	}
	id, err := o.seedForAccount(orderSeed{Status: status}, counterpart)
	if err != nil {
		return err
	}
	o.seededOrders[id] = status
	o.lastOrderID = id

	rejected := "REJECTED"
	for v := 1; v <= n; v++ {
		key := fmt.Sprintf("orders/%s/v%d/preview.png", id, v)
		decision := "WAIT"
		if v < n {
			decision = rejected
		}
		if err := seedDeliverable(id, v, key, decision); err != nil {
			return err
		}
		o.lastDeliverablePreviewKey = key
	}
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
	o.otherOrderID, err = seedOrder(orderSeed{CustomerID: other.ID, ArtistID: artist.ID})
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
	o.otherOrderID, err = seedOrder(orderSeed{CustomerID: customer.ID, ArtistID: other.ID})
	return err
}

func (o *ordersContext) anArtworkExistsForCommission() error {
	artist, err := apptest.RegisterArtist(app, apptest.NewClient(app.BaseURL()), "Bio")
	if err != nil {
		return err
	}
	categoryID := uuid.New()
	artworkID := uuid.New()

	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO categories (id, label) VALUES (?, ?)
	`, categoryID, "Category "+categoryID.String()); err != nil {
		return err
	}

	if _, err := app.DB.ExecContext(context.Background(), `
		INSERT INTO artworks (id, artist_id, category_id, name, description, price_satang, minimum_deadline_days)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, artworkID, uuid.MustParse(artist.ID), categoryID, "Sample Artwork", "Desc", 10000, 7); err != nil {
		return err
	}

	o.targetArtworkID = artworkID.String()
	return nil
}

func (o *ordersContext) theUserSubmitsANewOrderForTheArtwork() error {
	return o.submitNewOrder(time.Now().AddDate(0, 0, 10))
}

func (o *ordersContext) theUserSubmitsANewOrderWithDeadlineIn(days int) error {
	return o.submitNewOrder(time.Now().AddDate(0, 0, days))
}

func (o *ordersContext) theUserSubmitsANewOrderWithDeadlineInThePast() error {
	return o.submitNewOrder(time.Now().AddDate(0, 0, -1))
}

func (o *ordersContext) submitNewOrder(deadline time.Time) error {
	payload := map[string]any{
		"artwork_id":           o.targetArtworkID,
		"name":                 "My Custom Portrait",
		"customer_description": "Blue background please",
		"deadline_at":          deadline.Format(time.RFC3339),
	}

	resp, err := o.client.Do(http.MethodPost, "/orders", payload, map[string]string{
		"Authorization": "Bearer " + o.accessToken,
	})
	if err != nil {
		return err
	}
	o.resp = resp
	return nil
}

func (o *ordersContext) theSystemRejectsTheRequestDueToForbiddenRole() error {
	return o.expectClientError()
}

func (o *ordersContext) theSystemRejectsTheOrderDueToAnInvalidDeadline() error {
	return o.expectClientError()
}

func (o *ordersContext) theUserSubmitsANewOrderWithoutLoggingIn() error {
	deadline := time.Now().AddDate(0, 0, 10).Format(time.RFC3339)
	payload := map[string]any{
		"artwork_id":           o.targetArtworkID,
		"name":                 "My Custom Portrait",
		"customer_description": "Blue background please",
		"deadline_at":          deadline,
	}

	resp, err := o.client.Do(http.MethodPost, "/orders", payload, nil)
	if err != nil {
		return err
	}
	o.resp = resp
	return nil
}

func (o *ordersContext) theSystemCreatesTheOrderSuccessfullyWithStatus(status string) error {
	if err := o.expectStatus(http.StatusCreated); err != nil {
		return err
	}
	var res struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := o.resp.JSON(&res); err != nil {
		return err
	}
	if res.Status != status {
		return fmt.Errorf("expected status %s, got %s", status, res.Status)
	}
	return nil
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
	o.page = viewOrdersOutput{}
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

// theUserViewsTheirLastOrder fetches the most recently seeded order detail as the authenticated user.
func (o *ordersContext) theUserViewsTheirLastOrder() error {
	return o.getOrder(o.lastOrderID, true)
}

// theUserViewsTheirLastOrderWithoutLoggingIn fetches order detail without an access token.
func (o *ordersContext) theUserViewsTheirLastOrderWithoutLoggingIn() error {
	return o.getOrder(o.lastOrderID, false)
}

// theUserViewsAnotherUsersOrder fetches an order that belongs to a different account.
func (o *ordersContext) theUserViewsAnotherUsersOrder() error {
	return o.getOrder(o.otherOrderID, true)
}

func (o *ordersContext) confirmLastOrder(accept bool) error {
	resp, err := o.client.Do(
		http.MethodPut,
		"/orders/"+o.lastOrderID+"/confirm",
		map[string]bool{"accept": accept},
		map[string]string{"Authorization": "Bearer " + o.accessToken},
	)
	if err != nil {
		return err
	}
	o.resp = resp
	return nil
}

func (o *ordersContext) theArtistAcceptsTheirLastOrder() error {
	return o.confirmLastOrder(true)
}

func (o *ordersContext) theArtistRejectsTheirLastOrder() error {
	return o.confirmLastOrder(false)
}

func (o *ordersContext) theUserAttemptsToAcceptTheirLastOrder() error {
	return o.confirmLastOrder(true)
}

// getOrder sends GET /orders/{id} and decodes a successful detail response.
func (o *ordersContext) getOrder(orderID string, authenticated bool) error {
	headers := map[string]string(nil)
	if authenticated {
		headers = map[string]string{"Authorization": "Bearer " + o.accessToken}
	}
	resp, err := o.client.Do(http.MethodGet, "/orders/"+orderID, nil, headers)
	if err != nil {
		return err
	}
	o.resp = resp
	o.orderDetail = orderDetailOutput{}
	if resp.StatusCode == http.StatusOK {
		if err := resp.JSON(&o.orderDetail); err != nil {
			return fmt.Errorf("decode order detail response: %w (body: %s)", err, resp.Body)
		}
	}
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

type orderSummaryView struct {
	ID                    string  `json:"id"`
	CustomerID            string  `json:"customer_id"`
	ArtistID              string  `json:"artist_id"`
	Name                  string  `json:"name"`
	PriceSatang           int64   `json:"price_satang"`
	Status                string  `json:"status"`
	DeadlineAt            *string `json:"deadline_at"`
	DeliverablePreviewURL *string `json:"deliverable_preview_url"`
}

type viewOrdersOutput struct {
	Orders []orderSummaryView `json:"orders"`
	Total  int                `json:"total"`
}

type orderDetailOutput struct {
	ID                  string                `json:"id"`
	CustomerID          string                `json:"customer_id"`
	ArtistID            string                `json:"artist_id"`
	Name                string                `json:"name"`
	ArtworkID           *string               `json:"artwork_id"`
	ArtworkSnapshot     artworkSnapshotOutput `json:"artwork_snapshot"`
	PriceSatang         int64                 `json:"price_satang"`
	CustomerDescription string                `json:"customer_description"`
	DeadlineAt          time.Time             `json:"deadline_at"`
	Status              string                `json:"status"`
	OtherParty          orderPartyOutput      `json:"other_party"`
}

type artworkSnapshotOutput struct {
	ArtworkName string   `json:"artwork_name"`
	Category    string   `json:"category"`
	Styles      []string `json:"styles"`
}

type orderPartyOutput struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	ArtistReviewScore *float64 `json:"artist_review_score"`
}

func idsOf(orders []orderSummaryView) []string {
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
		if ord.PriceSatang != seedPriceSatang ||
			ord.Name != seedOrderName {
			return fmt.Errorf("order %s did not preserve its name/price snapshot: %+v", ord.ID, ord)
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
func assertDeadlineOrder(orders []orderSummaryView, ascending bool) error {
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

// theSystemShowsTheOrdersLatestDeliverablePreview asserts the response's
// deliverable_preview_url for o.lastOrderID is non-null and references
// o.lastDeliverablePreviewKey — the highest version seeded, proving the
// repository's DISTINCT ON ... ORDER BY version DESC picks the latest
// row and the usecase presigns it correctly.
func (o *ordersContext) theSystemShowsTheOrdersLatestDeliverablePreview() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	for _, ord := range body.Orders {
		if ord.ID != o.lastOrderID {
			continue
		}
		if ord.DeliverablePreviewURL == nil {
			return fmt.Errorf("order %s: expected a deliverable preview url, got null", ord.ID)
		}
		if !strings.Contains(*ord.DeliverablePreviewURL, o.lastDeliverablePreviewKey) {
			return fmt.Errorf("order %s: expected preview url to reference key %q, got %q", ord.ID, o.lastDeliverablePreviewKey, *ord.DeliverablePreviewURL)
		}
		return nil
	}
	return fmt.Errorf("order %s not found in response", o.lastOrderID)
}

// theSystemShowsNoDeliverablePreviewForTheOrder asserts every order in the
// response has a null deliverable_preview_url — used by the "no
// deliverable submitted yet" scenario, where exactly one order exists.
func (o *ordersContext) theSystemShowsNoDeliverablePreviewForTheOrder() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	for _, ord := range body.Orders {
		if ord.DeliverablePreviewURL != nil {
			return fmt.Errorf("order %s: expected no deliverable preview url, got %q", ord.ID, *ord.DeliverablePreviewURL)
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

// theSystemShowsTheOrderDetailsForTheUser checks persisted details and the caller's opposite party.
func (o *ordersContext) theSystemShowsTheOrderDetailsForTheUser() error {
	if err := o.expectStatus(http.StatusOK); err != nil {
		return err
	}
	got := o.orderDetail
	if got.ID != o.lastOrderID {
		return fmt.Errorf("order detail ID = %q, want %q", got.ID, o.lastOrderID)
	}
	if got.Name != seedOrderName ||
		got.PriceSatang != seedPriceSatang ||
		got.CustomerDescription != seedCustomerDescription ||
		got.Status != "PENDING" {
		return fmt.Errorf("order detail does not match seeded order: %+v", got)
	}
	if got.ArtworkID == nil || *got.ArtworkID == "" {
		return fmt.Errorf("expected order detail to include artwork_id: %+v", got)
	}
	if got.ArtworkSnapshot.ArtworkName != seedArtworkName ||
		!strings.HasPrefix(got.ArtworkSnapshot.Category, "Order fixture ") ||
		len(got.ArtworkSnapshot.Styles) != 0 {
		return fmt.Errorf("unexpected artwork snapshot: %+v", got.ArtworkSnapshot)
	}
	if got.DeadlineAt.IsZero() {
		return fmt.Errorf("expected order detail to include a deadline: %+v", got)
	}

	var expectedParty apptest.Account
	if o.accountRole == "artist" {
		if got.CustomerID != o.counterpart.ID || got.ArtistID != o.account.ID {
			return fmt.Errorf("order participants = customer %q artist %q; want customer %q artist %q", got.CustomerID, got.ArtistID, o.counterpart.ID, o.account.ID)
		}
		expectedParty = o.counterpart
	} else {
		if got.CustomerID != o.account.ID || got.ArtistID != o.counterpart.ID {
			return fmt.Errorf("order participants = customer %q artist %q; want customer %q artist %q", got.CustomerID, got.ArtistID, o.account.ID, o.counterpart.ID)
		}
		expectedParty = o.counterpart
	}
	if got.OtherParty.ID != expectedParty.ID ||
		got.OtherParty.Name != expectedParty.Username ||
		got.OtherParty.Email != expectedParty.Email {
		return fmt.Errorf("other_party = %+v; want account ID %q, username %q, email %q", got.OtherParty, expectedParty.ID, expectedParty.Username, expectedParty.Email)
	}
	return nil
}

func (o *ordersContext) theSystemConfirmsTheOrderWithStatus(status string) error {
	if err := o.expectStatus(http.StatusOK); err != nil {
		return err
	}

	var confirmation struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	if err := o.resp.JSON(&confirmation); err != nil {
		return fmt.Errorf("decode confirmation response: %w (body: %s)", err, o.resp.Body)
	}

	wantMessage := "Order accepted"
	if status == "CANCEL" {
		wantMessage = "Order rejected"
	}
	if confirmation.Message != wantMessage {
		return fmt.Errorf("confirmation message = %q, want %q", confirmation.Message, wantMessage)
	}
	if confirmation.Status != status {
		return fmt.Errorf("confirmation status = %q, want %q", confirmation.Status, status)
	}

	if err := o.getOrder(o.lastOrderID, true); err != nil {
		return err
	}
	if err := o.expectStatus(http.StatusOK); err != nil {
		return err
	}
	if o.orderDetail.Status != status {
		return fmt.Errorf("order status = %q, want %q", o.orderDetail.Status, status)
	}
	return nil
}

func (o *ordersContext) theSystemForbidsTheOrderConfirmation() error {
	return o.expectStatus(http.StatusForbidden)
}

func (o *ordersContext) theSystemRejectsConfirmationForInvalidOrderStatus() error {
	return o.expectStatus(http.StatusConflict)
}

// theSystemHidesTheOrderFromTheUser verifies another participant's order is reported as not found.
func (o *ordersContext) theSystemHidesTheOrderFromTheUser() error {
	return o.expectStatus(http.StatusNotFound)
}

func (o *ordersContext) decodeOrders() (viewOrdersOutput, error) {
	if o.resp.StatusCode != http.StatusOK {
		return viewOrdersOutput{}, fmt.Errorf("expected response status 200, got %d: %s", o.resp.StatusCode, o.resp.Body)
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

// InitializeScenario registers order endpoint steps and resets the
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
	sc.Step(`^the user has an order with status "([^"]*)" and (\d+) submitted deliverable versions$`, func(status string, n int) error {
		return o.theUserHasAnOrderWithStatusAndNSubmittedDeliverableVersions(status, n)
	})
	sc.Step(`^the user has an order with deadline "([^"]*)"$`, func(deadline string) error { return o.theUserHasAnOrderWithDeadline(deadline) })
	sc.Step(`^the user has an order with no deadline$`, func() error { return o.theUserHasAnOrderWithDeadline("") })
	sc.Step(`^the user has (\d+) orders$`, func(n int) error { return o.theUserHasNOrders(n) })
	sc.Step(`^another customer has an order$`, func() error { return o.anotherCustomerHasAnOrder() })
	sc.Step(`^another artist has an order$`, func() error { return o.anotherArtistHasAnOrder() })
	sc.Step(`^an artwork exists for commission$`, func() error { return o.anArtworkExistsForCommission() })

	// when
	sc.Step(`^the user views their orders$`, func() error { return o.theUserViewsTheirOrders() })
	sc.Step(`^the user views their orders without logging in$`, func() error { return o.theUserViewsTheirOrdersWithoutLoggingIn() })
	sc.Step(`^the user views their last order$`, func() error { return o.theUserViewsTheirLastOrder() })
	sc.Step(`^the user views their last order without logging in$`, func() error { return o.theUserViewsTheirLastOrderWithoutLoggingIn() })
	sc.Step(`^the user views another user's order$`, func() error { return o.theUserViewsAnotherUsersOrder() })
	sc.Step(`^the artist accepts their last order$`, func() error { return o.theArtistAcceptsTheirLastOrder() })
	sc.Step(`^the artist rejects their last order$`, func() error { return o.theArtistRejectsTheirLastOrder() })
	sc.Step(`^the user attempts to accept their last order$`, func() error { return o.theUserAttemptsToAcceptTheirLastOrder() })
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
	sc.Step(`^the user submits a new order for the artwork$`, func() error { return o.theUserSubmitsANewOrderForTheArtwork() })
	sc.Step(`^the user submits a new order with a deadline (-?\d+) days from now$`, func(days int) error {
		return o.theUserSubmitsANewOrderWithDeadlineIn(days)
	})
	sc.Step(`^the user submits a new order with a deadline in the past$`, func() error {
		return o.theUserSubmitsANewOrderWithDeadlineInThePast()
	})
	sc.Step(`^the user submits a new order without logging in$`, func() error { return o.theUserSubmitsANewOrderWithoutLoggingIn() })

	// then
	sc.Step(`^the system shows all of the user's orders with their current status$`, func() error {
		return o.theSystemShowsAllOfTheUsersOrdersWithTheirCurrentStatus()
	})
	sc.Step(`^the system does not show any other user's orders$`, func() error { return o.theSystemDoesNotShowAnyOtherUsersOrders() })
	sc.Step(`^the system shows an empty order list$`, func() error { return o.theSystemShowsAnEmptyOrderList() })
	sc.Step(`^the system shows only orders with status "([^"]*)"$`, func(status string) error {
		return o.theSystemShowsOnlyOrdersWithStatus(status)
	})
	sc.Step(`^the system shows the order's latest deliverable preview$`, func() error {
		return o.theSystemShowsTheOrdersLatestDeliverablePreview()
	})
	sc.Step(`^the system shows no deliverable preview for the order$`, func() error {
		return o.theSystemShowsNoDeliverablePreviewForTheOrder()
	})
	sc.Step(`^the system returns every seeded order exactly once$`, func() error { return o.theSystemReturnsEveryOrderExactlyOnce() })
	sc.Step(`^the system returns orders sorted by deadline in "([^"]*)" order$`, func(order string) error {
		return o.theSystemReturnsOrdersSortedByDeadline(order)
	})
	sc.Step(`^the system returns the first page of orders again$`, func() error { return o.theSystemReturnsTheFirstPageOfOrdersAgain() })
	sc.Step(`^the system reports a total of (\d+) orders$`, func(n int) error { return o.theSystemReportsATotalOfNOrders(n) })
	sc.Step(`^the system requires the user to log in$`, func() error { return o.theSystemRequiresTheUserToLogIn() })
	sc.Step(`^the system shows the order details for the user$`, func() error {
		return o.theSystemShowsTheOrderDetailsForTheUser()
	})
	sc.Step(`^the system confirms the order with status "([^"]*)"$`, func(status string) error {
		return o.theSystemConfirmsTheOrderWithStatus(status)
	})
	sc.Step(`^the system forbids the order confirmation$`, func() error {
		return o.theSystemForbidsTheOrderConfirmation()
	})
	sc.Step(`^the system rejects the confirmation because the order status does not allow it$`, func() error {
		return o.theSystemRejectsConfirmationForInvalidOrderStatus()
	})
	sc.Step(`^the system hides the order from the user$`, func() error {
		return o.theSystemHidesTheOrderFromTheUser()
	})
	sc.Step(`^the system rejects the request due to an invalid offset$`, func() error { return o.expectClientError() })
	sc.Step(`^the system rejects the request due to an invalid status$`, func() error { return o.expectClientError() })
	sc.Step(`^the system creates the order successfully with status "([^"]*)"$`, func(status string) error {
		return o.theSystemCreatesTheOrderSuccessfullyWithStatus(status)
	})
	sc.Step(`^the system rejects the request due to forbidden role$`, func() error {
		return o.theSystemRejectsTheRequestDueToForbiddenRole()
	})
	sc.Step(`^the system rejects the order due to an invalid deadline$`, func() error {
		return o.theSystemRejectsTheOrderDueToAnInvalidDeadline()
	})
}
