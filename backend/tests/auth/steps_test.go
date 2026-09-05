//go:build integration

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

// authContext holds the state for exactly one scenario: the HTTP client
// (its own cookie jar, isolated from every other scenario), the last
// response, and whatever the scenario's "Given"/"When" steps produced for
// later steps to use.
type authContext struct {
	client *apptest.Client
	resp   *apptest.Response

	// credentials of "the registered account" set up by Given steps.
	username string
	email    string
	password string

	// captured across steps within one scenario.
	accessToken     string
	previousRefresh string
}

// --- given ---

// theUserDoesNotHaveAnAccount is a no-op: every scenario already starts
// with no account for its (unique, freshly generated) user. The step
// exists so the precondition is documented, even though nothing needs
// setting up for it to be true.
func (a *authContext) theUserDoesNotHaveAnAccount() error {
	return nil
}

func (a *authContext) theUserHasARegisteredAccount() error {
	suffix := apptest.UniqueSuffix()
	a.username = "cust-" + suffix
	a.email = "cust-" + suffix + "@example.com"
	a.password = apptest.FixturePassword

	if err := a.doRegister(apptest.NewCustomerRegisterBody(a.username, a.email, a.password)); err != nil {
		return err
	}
	if a.resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("fixture setup: expected registration to succeed with 201, got %d: %s", a.resp.StatusCode, a.resp.Body)
	}
	return nil
}

func (a *authContext) theUserHasLoggedIn() error {
	if err := a.login(a.email, a.password); err != nil {
		return err
	}
	if a.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fixture setup: expected login to succeed with 200, got %d: %s", a.resp.StatusCode, a.resp.Body)
	}
	return nil
}

// --- when: register ---

func (a *authContext) theUserRegistersWithValidDetails() error {
	username := "cust-" + apptest.UniqueSuffix()
	return a.doRegister(apptest.NewCustomerRegisterBody(username, username+"@example.com", apptest.FixturePassword))
}

func (a *authContext) theUserRegistersWithInvalidField(field, value string) error {
	username := "cust-" + apptest.UniqueSuffix()
	fields, err := registerFields(username, username+"@example.com", apptest.FixturePassword)
	if err != nil {
		return err
	}
	fields[field] = value
	return a.doRegister(fields)
}

func (a *authContext) theUserRegistersReusingTheUsername() error {
	email := "dup-" + apptest.UniqueSuffix() + "@example.com"
	return a.doRegister(apptest.NewCustomerRegisterBody(a.username, email, apptest.FixturePassword))
}

func (a *authContext) theUserRegistersReusingTheEmail() error {
	username := "dup-" + apptest.UniqueSuffix()
	return a.doRegister(apptest.NewCustomerRegisterBody(username, a.email, apptest.FixturePassword))
}

func (a *authContext) theUserRegistersAsAnArtistWithADescription() error {
	username := "artist-" + apptest.UniqueSuffix()
	return a.doRegister(apptest.NewArtistRegisterBody(username, username+"@example.com", apptest.FixturePassword, "I paint custom portraits."))
}

func (a *authContext) theUserRegistersAsAnArtistWithoutADescription() error {
	username := "artist-" + apptest.UniqueSuffix()
	return a.doRegister(apptest.NewArtistRegisterBody(username, username+"@example.com", apptest.FixturePassword, ""))
}

func (a *authContext) doRegister(payload any) error {
	resp, err := a.client.Do(http.MethodPost, "/auth/register", payload, nil)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

// --- when: login ---

func (a *authContext) theUserLogsInWithValidCredentials() error {
	return a.login(a.email, a.password)
}

func (a *authContext) theUserLogsInWithAnIncorrectPassword() error {
	return a.login(a.email, wrongPassword)
}

func (a *authContext) theUserLogsInWithAnUnknownEmail() error {
	return a.login(unknownEmail, apptest.FixturePassword)
}

func (a *authContext) login(email, password string) error {
	body := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{Email: email, Password: password}

	resp, err := a.client.Do(http.MethodPost, "/auth/login", body, nil)
	if err != nil {
		return err
	}
	a.resp = resp
	a.captureAccessToken()
	return nil
}

// --- when: refresh ---

const refreshPath = "/auth/refresh"

func (a *authContext) theUserRefreshesTheSession() error {
	if v, ok := a.client.CookieValue(a.client.BaseURL()+refreshPath, "refresh_token"); ok {
		a.previousRefresh = v
	}

	resp, err := a.client.Do(http.MethodPost, refreshPath, nil, nil)
	if err != nil {
		return err
	}
	a.resp = resp
	a.captureAccessToken()
	return nil
}

func (a *authContext) theUserRefreshesTheSessionAgainUsingThePreviousToken() error {
	if a.previousRefresh == "" {
		return fmt.Errorf("no previous refresh token captured yet")
	}
	resp, err := a.client.DoWithCookie(http.MethodPost, refreshPath, "refresh_token", a.previousRefresh)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func (a *authContext) theUserRefreshesTheSessionWithoutAToken() error {
	resp, err := a.client.DoNoCookies(http.MethodPost, refreshPath)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func (a *authContext) theUserRefreshesTheSessionWithAnInvalidToken() error {
	resp, err := a.client.DoWithCookie(http.MethodPost, refreshPath, "refresh_token", "not-a-real-token")
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

// --- when: logout ---

func (a *authContext) theUserLogsOut() error {
	// Capture the refresh token before logout.
	if v, ok := a.client.CookieValue(a.client.BaseURL()+refreshPath, "refresh_token"); ok {
		a.previousRefresh = v
	}
	return a.logout(a.accessToken)
}

func (a *authContext) theUserRefreshesUsingThePreviousRefreshToken() error {
	if a.previousRefresh == "" {
		return fmt.Errorf("no previous refresh token captured yet")
	}
	resp, err := a.client.DoWithCookie(http.MethodPost, refreshPath, "refresh_token", a.previousRefresh)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func (a *authContext) theUserLogsOutWithoutAnAccessToken() error {
	resp, err := a.client.Do(http.MethodPost, "/auth/logout", nil, nil)
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

func (a *authContext) theUserLogsOutWithAnInvalidAccessToken() error {
	return a.logout("not-a-real-token")
}

func (a *authContext) logout(bearerToken string) error {
	resp, err := a.client.Do(http.MethodPost, "/auth/logout", nil, map[string]string{
		"Authorization": "Bearer " + bearerToken,
	})
	if err != nil {
		return err
	}
	a.resp = resp
	return nil
}

// --- then ---

func (a *authContext) theSystemCreatesTheAccount() error {
	return a.expectStatus(http.StatusCreated)
}

func (a *authContext) theSystemDoesNotCreateTheAccount() error {
	return a.expectClientError()
}

func (a *authContext) theSystemAuthenticatesTheUser() error {
	return a.expectStatus(http.StatusOK)
}

func (a *authContext) theSystemDoesNotAuthenticateTheUser() error {
	return a.expectStatus(http.StatusUnauthorized)
}

func (a *authContext) theSystemStartsOrRenewsTheSession() error {
	if err := a.theResponseBodyShouldContainAnAccessToken(); err != nil {
		return err
	}
	return a.theResponseShouldSetARefreshTokenCookie()
}

func (a *authContext) theSystemRejectsTheRequest() error {
	return a.expectStatus(http.StatusUnauthorized)
}

func (a *authContext) theSystemTerminatesTheCurrentSession() error {
	return a.expectStatus(http.StatusNoContent)
}

func (a *authContext) theSystemDoesNotTerminateASession() error {
	return a.expectStatus(http.StatusUnauthorized)
}

func (a *authContext) theSystemDisplaysAnAppropriateErrorMessage() error {
	var body struct {
		Detail string `json:"detail"`
	}
	if err := a.resp.JSON(&body); err != nil {
		return fmt.Errorf("decode error body: %w (body: %s)", err, a.resp.Body)
	}
	if strings.TrimSpace(body.Detail) == "" {
		return fmt.Errorf("expected a non-empty error message, got body: %s", a.resp.Body)
	}
	return nil
}

// --- assertion helpers ---

func (a *authContext) expectStatus(code int) error {
	if a.resp.StatusCode != code {
		return fmt.Errorf("expected response status %d, got %d: %s", code, a.resp.StatusCode, a.resp.Body)
	}
	return nil
}

// expectClientError checks for any 4xx: the business requirement is "the
// system refused and said why," not which specific validation code fired
// (400 domain rule, 409 conflict, 422 schema validation are all valid
// reasons a registration can be refused).
func (a *authContext) expectClientError() error {
	if a.resp.StatusCode < 400 || a.resp.StatusCode >= 500 {
		return fmt.Errorf("expected a 4xx client error, got %d: %s", a.resp.StatusCode, a.resp.Body)
	}
	return nil
}

func (a *authContext) theResponseBodyShouldContainAnAccessToken() error {
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := a.resp.JSON(&body); err != nil {
		return fmt.Errorf("decode response body: %w (body: %s)", err, a.resp.Body)
	}
	if body.AccessToken == "" {
		return fmt.Errorf("expected a non-empty access_token, got body: %s", a.resp.Body)
	}
	return nil
}

func (a *authContext) theResponseShouldSetARefreshTokenCookie() error {
	for _, raw := range a.resp.Header.Values("Set-Cookie") {
		if strings.HasPrefix(raw, "refresh_token=") && !strings.HasPrefix(raw, "refresh_token=;") {
			return nil
		}
	}
	return fmt.Errorf("expected a refresh_token Set-Cookie header, got: %v", a.resp.Header.Values("Set-Cookie"))
}

func (a *authContext) captureAccessToken() {
	if a.resp.StatusCode != http.StatusOK {
		return
	}
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := a.resp.JSON(&body); err == nil {
		a.accessToken = body.AccessToken
	}
}

// InitializeScenario registers every auth step and resets the scenario
// context (a fresh client with an empty cookie jar) before each scenario,
// so scenarios never see each other's cookies or credentials.
func InitializeScenario(sc *godog.ScenarioContext) {
	var a *authContext

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		a = &authContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})

	// given
	sc.Step(`^the user does not have an account$`, func() error { return a.theUserDoesNotHaveAnAccount() })
	sc.Step(`^the user has a registered account$`, func() error { return a.theUserHasARegisteredAccount() })
	sc.Step(`^the user has logged in$`, func() error { return a.theUserHasLoggedIn() })

	// when: register
	sc.Step(`^the user registers with a valid username, password, and email$`, func() error {
		return a.theUserRegistersWithValidDetails()
	})
	sc.Step(`^the user registers with (\S+) "([^"]*)"$`, func(field, value string) error {
		return a.theUserRegistersWithInvalidField(field, value)
	})
	sc.Step(`^the user registers reusing that account's username$`, func() error { return a.theUserRegistersReusingTheUsername() })
	sc.Step(`^the user registers reusing that account's email$`, func() error { return a.theUserRegistersReusingTheEmail() })
	sc.Step(`^the user registers as an artist with a profile description$`, func() error {
		return a.theUserRegistersAsAnArtistWithADescription()
	})
	sc.Step(`^the user registers as an artist without a profile description$`, func() error {
		return a.theUserRegistersAsAnArtistWithoutADescription()
	})

	// when: login
	sc.Step(`^the user logs in with valid credentials$`, func() error { return a.theUserLogsInWithValidCredentials() })
	sc.Step(`^the user logs in with an incorrect password$`, func() error { return a.theUserLogsInWithAnIncorrectPassword() })
	sc.Step(`^the user logs in with an unknown email$`, func() error { return a.theUserLogsInWithAnUnknownEmail() })

	// when: refresh
	sc.Step(`^the user refreshes the session$`, func() error { return a.theUserRefreshesTheSession() })
	sc.Step(`^the user refreshes the session again using the previous refresh token$`, func() error {
		return a.theUserRefreshesTheSessionAgainUsingThePreviousToken()
	})
	sc.Step(`^the user refreshes the session without a refresh token$`, func() error { return a.theUserRefreshesTheSessionWithoutAToken() })
	sc.Step(`^the user refreshes the session with an invalid refresh token$`, func() error {
		return a.theUserRefreshesTheSessionWithAnInvalidToken()
	})

	// when: logout
	sc.Step(`^the user logs out$`, func() error { return a.theUserLogsOut() })
	sc.Step(`^the user logs out without an access token$`, func() error { return a.theUserLogsOutWithoutAnAccessToken() })
	sc.Step(`^the user logs out with an invalid access token$`, func() error { return a.theUserLogsOutWithAnInvalidAccessToken() })
	sc.Step(`^the user refreshes the session using the previous refresh token$`, func() error { return a.theUserRefreshesUsingThePreviousRefreshToken() })

	// then
	sc.Step(`^the system creates the account$`, func() error { return a.theSystemCreatesTheAccount() })
	sc.Step(`^the system does not create the account$`, func() error { return a.theSystemDoesNotCreateTheAccount() })
	sc.Step(`^the system displays an appropriate error message$`, func() error { return a.theSystemDisplaysAnAppropriateErrorMessage() })
	sc.Step(`^the system authenticates the user$`, func() error { return a.theSystemAuthenticatesTheUser() })
	sc.Step(`^the system does not authenticate the user$`, func() error { return a.theSystemDoesNotAuthenticateTheUser() })
	sc.Step(`^the system starts a new session for the user$`, func() error { return a.theSystemStartsOrRenewsTheSession() })
	sc.Step(`^the system renews the session for the user$`, func() error { return a.theSystemStartsOrRenewsTheSession() })
	sc.Step(`^the system rejects the request$`, func() error { return a.theSystemRejectsTheRequest() })
	sc.Step(`^the system terminates the current session$`, func() error { return a.theSystemTerminatesTheCurrentSession() })
	sc.Step(`^the system does not terminate a session$`, func() error { return a.theSystemDoesNotTerminateASession() })
}
