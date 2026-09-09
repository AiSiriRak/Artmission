//go:build integration

package orders

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

// ordersContext holds the state for exactly one scenario: the HTTP client
// (its own cookie jar, isolated from every other scenario), the account
// under test, and whatever fixtures its Given steps produced.
type ordersContext struct {
	client *apptest.Client

	account     apptest.Account
	accessToken string
	artist      apptest.Account // shared, lazily-created FK backing every seeded order

	seededOrders map[string]string // order ID → seeded status, for the account under test

	resp *apptest.Response
}

// --- given ---

func (o *ordersContext) theUserHasARegisteredCustomerAccount() error {
	account, err := apptest.RegisterCustomer(app, o.client)
	if err != nil {
		return err
	}
	o.account = account
	return nil
}

func (o *ordersContext) theUserHasARegisteredArtistAccount() error {
	account, err := apptest.RegisterArtist(app, o.client, "I paint custom portraits.")
	if err != nil {
		return err
	}
	o.account = account
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
	artist, err := o.sharedArtist()
	if err != nil {
		return err
	}
	for range 2 {
		id, status, err := seedOrder(o.account.ID, artist.ID)
		if err != nil {
			return err
		}
		o.seededOrders[id] = status
	}
	return nil
}

func (o *ordersContext) theUserHasAnOrder() error {
	artist, err := o.sharedArtist()
	if err != nil {
		return err
	}
	id, status, err := seedOrder(o.account.ID, artist.ID)
	if err != nil {
		return err
	}
	o.seededOrders[id] = status
	return nil
}

func (o *ordersContext) anotherCustomerHasAnOrder() error {
	other, err := apptest.RegisterCustomer(app, apptest.NewClient(app.BaseURL()))
	if err != nil {
		return err
	}
	artist, err := o.sharedArtist()
	if err != nil {
		return err
	}
	_, _, err = seedOrder(other.ID, artist.ID)
	return err
}

// --- when ---

func (o *ordersContext) theUserViewsTheirHiringHistory() error {
	resp, err := o.client.Do(http.MethodGet, "/orders/history", nil, map[string]string{
		"Authorization": "Bearer " + o.accessToken,
	})
	if err != nil {
		return err
	}
	o.resp = resp
	return nil
}

func (o *ordersContext) theUserViewsTheirHiringHistoryWithoutLoggingIn() error {
	resp, err := o.client.Do(http.MethodGet, "/orders/history", nil, nil)
	if err != nil {
		return err
	}
	o.resp = resp
	return nil
}

// --- then ---

type orderViewBody struct {
	Orders []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"orders"`
}

func (o *ordersContext) theSystemShowsAllOfTheUsersOrdersWithTheirCurrentStatus() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}

	got := map[string]string{}
	for _, order := range body.Orders {
		if order.Status == "" {
			return fmt.Errorf("expected every order to have a status, got: %+v", order)
		}
		got[order.ID] = order.Status
	}
	for wantID, wantStatus := range o.seededOrders {
		gotStatus, ok := got[wantID]
		if !ok {
			return fmt.Errorf("expected seeded order %s in hiring history, got orders: %+v", wantID, body.Orders)
		}
		if gotStatus != wantStatus {
			return fmt.Errorf("order %s: expected status %q, got %q", wantID, wantStatus, gotStatus)
		}
	}
	return nil
}

func (o *ordersContext) theSystemDoesNotShowAnyOtherCustomersOrders() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}

	if len(body.Orders) != len(o.seededOrders) {
		return fmt.Errorf("expected exactly the user's own %d order(s), got %d: %+v", len(o.seededOrders), len(body.Orders), body.Orders)
	}
	want := map[string]bool{}
	for id := range o.seededOrders {
		want[id] = true
	}
	for _, order := range body.Orders {
		if !want[order.ID] {
			return fmt.Errorf("hiring history returned an order %s that does not belong to the user", order.ID)
		}
	}
	return nil
}

func (o *ordersContext) theSystemShowsAnEmptyHiringHistory() error {
	body, err := o.decodeOrders()
	if err != nil {
		return err
	}
	if len(body.Orders) != 0 {
		return fmt.Errorf("expected an empty hiring history, got %d orders", len(body.Orders))
	}
	return nil
}

func (o *ordersContext) theSystemRequiresTheUserToLogIn() error {
	return o.expectStatus(http.StatusUnauthorized)
}

func (o *ordersContext) theSystemForbidsTheRequest() error {
	return o.expectStatus(http.StatusForbidden)
}

func (o *ordersContext) decodeOrders() (orderViewBody, error) {
	if o.resp.StatusCode != http.StatusOK {
		return orderViewBody{}, fmt.Errorf("expected response status 200, got %d: %s", o.resp.StatusCode, o.resp.Body)
	}
	var body orderViewBody
	if err := o.resp.JSON(&body); err != nil {
		return orderViewBody{}, fmt.Errorf("decode response body: %w (body: %s)", err, o.resp.Body)
	}
	return body, nil
}

func (o *ordersContext) expectStatus(code int) error {
	if o.resp.StatusCode != code {
		return fmt.Errorf("expected response status %d, got %d: %s", code, o.resp.StatusCode, o.resp.Body)
	}
	return nil
}

// InitializeScenario registers every hiring-history step and resets the
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
	sc.Step(`^another customer has an order$`, func() error { return o.anotherCustomerHasAnOrder() })

	// when
	sc.Step(`^the user views their hiring history$`, func() error { return o.theUserViewsTheirHiringHistory() })
	sc.Step(`^the user views their hiring history without logging in$`, func() error {
		return o.theUserViewsTheirHiringHistoryWithoutLoggingIn()
	})

	// then
	sc.Step(`^the system shows all of the user's orders with their current status$`, func() error {
		return o.theSystemShowsAllOfTheUsersOrdersWithTheirCurrentStatus()
	})
	sc.Step(`^the system does not show any other customer's orders$`, func() error { return o.theSystemDoesNotShowAnyOtherCustomersOrders() })
	sc.Step(`^the system shows an empty hiring history$`, func() error { return o.theSystemShowsAnEmptyHiringHistory() })
	sc.Step(`^the system requires the user to log in$`, func() error { return o.theSystemRequiresTheUserToLogIn() })
	sc.Step(`^the system forbids the request$`, func() error { return o.theSystemForbidsTheRequest() })
}
