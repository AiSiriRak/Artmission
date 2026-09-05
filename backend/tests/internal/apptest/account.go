//go:build integration

package apptest

import (
	"context"
	"fmt"
	"net/http"
)

// Account is not a backend type — it's test-fixture bookkeeping: the
// credentials a suite used to create an account through the real HTTP
// API, plus the ID looked up afterward (registration's response carries
// no body). It exists so a scenario can remember "who did I just create"
// and log back in or reference that ID later. It has no relationship to
// any request/response wire shape or domain entity — in particular, it
// carries a plaintext Password no domain type could ever hold.
type Account struct {
	ID       string
	Username string
	Email    string
	Password string
}

// BankAccountBody, ArtistBody and RegisterBody mirror the wire shape of
// POST /auth/register's request body (see
// internal/handler/rest/auth_handler.go's registerInputBody) — kept
// independent of that unexported type since tests only care about the
// wire shape, not the Go type used to produce it. Exported so any
// domain's suite that needs to test register's own behavior (not just get
// a working fixture account) can build arbitrary — including invalid —
// payloads from the same shape, instead of keeping a second copy.
type BankAccountBody struct {
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
}

type ArtistBody struct {
	Description string `json:"description"`
}

type RegisterBody struct {
	Username    string          `json:"username"`
	Email       string          `json:"email"`
	Password    string          `json:"password"`
	Role        string          `json:"role"`
	BankAccount BankAccountBody `json:"bank_account"`
	Artist      *ArtistBody     `json:"artist,omitempty"`
}

// FixturePassword satisfies register's 8-16 character length bound.
const FixturePassword = "correct1"

// NewCustomerRegisterBody builds an otherwise-valid customer registration
// body for username/email/password.
func NewCustomerRegisterBody(username, email, password string) RegisterBody {
	return RegisterBody{
		Username:    username,
		Email:       email,
		Password:    password,
		Role:        "customer",
		BankAccount: BankAccountBody{BankName: "Test Bank", AccountNumber: "1234567890"},
	}
}

// NewArtistRegisterBody builds an artist registration body. An empty
// description omits the artist object entirely (rather than sending an
// empty description) — the case the domain rejects with 400.
func NewArtistRegisterBody(username, email, password, description string) RegisterBody {
	body := NewCustomerRegisterBody(username, email, password)
	body.Role = "artist"
	if description != "" {
		body.Artist = &ArtistBody{Description: description}
	}
	return body
}

// RegisterCustomer registers a new customer account through the real HTTP
// API and returns it with its ID looked up from the database — the
// register response carries no body (see RegisterOutput).
func RegisterCustomer(app *App, client *Client) (Account, error) {
	username := "cust-" + UniqueSuffix()
	body := NewCustomerRegisterBody(username, username+"@example.com", FixturePassword)
	return registerAccount(app, client, body)
}

// RegisterArtist registers a new artist account (with a profile
// description) through the real HTTP API.
func RegisterArtist(app *App, client *Client, description string) (Account, error) {
	username := "artist-" + UniqueSuffix()
	body := NewArtistRegisterBody(username, username+"@example.com", FixturePassword, description)
	return registerAccount(app, client, body)
}

func registerAccount(app *App, client *Client, body RegisterBody) (Account, error) {
	resp, err := client.Do(http.MethodPost, "/auth/register", body, nil)
	if err != nil {
		return Account{}, err
	}
	if resp.StatusCode != http.StatusCreated {
		return Account{}, fmt.Errorf("apptest: expected registration to succeed with 201, got %d: %s", resp.StatusCode, resp.Body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	id, err := app.UserIDByEmail(ctx, body.Email)
	if err != nil {
		return Account{}, fmt.Errorf("apptest: look up registered user id: %w", err)
	}

	return Account{ID: id.String(), Username: body.Username, Email: body.Email, Password: body.Password}, nil
}

// Login logs in with an existing account through the real HTTP API and
// returns the access token; the refresh_token cookie is left set on
// client's own jar.
func Login(client *Client, email, password string) (string, error) {
	body := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{Email: email, Password: password}

	resp, err := client.Do(http.MethodPost, "/auth/login", body, nil)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("apptest: expected login to succeed with 200, got %d: %s", resp.StatusCode, resp.Body)
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
	}
	if err := resp.JSON(&parsed); err != nil {
		return "", fmt.Errorf("apptest: decode login response: %w", err)
	}
	return parsed.AccessToken, nil
}
