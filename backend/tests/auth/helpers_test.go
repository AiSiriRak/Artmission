//go:build integration

package auth

// Request payload shapes, mirrored from
// internal/handler/rest/auth_handler.go's registerInputBody — kept
// independent of the handler's unexported types since tests only care
// about the wire shape, not the Go type used to produce it.

const (
	validPassword    = "correct1" // 8 chars: satisfies register's 8-16 length bound
	validFirstName   = "Test"
	validLastName    = "User"
	validPhoneNumber = "0800000000"
	validBankName    = "Test Bank"
	validAccountNo   = "1234567890"

	wrongPassword   = "wrong-password-1"
	unknownUsername = "no-such-user"
)

type bankAccountPayload struct {
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
}

type artistPayload struct {
	Description string `json:"description"`
}

type registerPayload struct {
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	Password    string             `json:"password"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	PhoneNumber string             `json:"phone_number"`
	Role        string             `json:"role"`
	BankAccount bankAccountPayload `json:"bank_account"`
	Artist      *artistPayload     `json:"artist,omitempty"`
}

// newCustomerRegisterPayload builds an otherwise-valid customer
// registration body for username/email.
func newCustomerRegisterPayload(username, email, password string) registerPayload {
	return registerPayload{
		Username:    username,
		Email:       email,
		Password:    password,
		FirstName:   validFirstName,
		LastName:    validLastName,
		PhoneNumber: validPhoneNumber,
		Role:        "customer",
		BankAccount: bankAccountPayload{BankName: validBankName, AccountNumber: validAccountNo},
	}
}

// newArtistRegisterPayload builds an artist registration body. Passing an
// empty description omits the artist object entirely (rather than sending
// an empty description), matching the case the domain rejects with 400.
func newArtistRegisterPayload(username, email, password, description string) registerPayload {
	payload := newCustomerRegisterPayload(username, email, password)
	payload.Role = "artist"
	if description != "" {
		payload.Artist = &artistPayload{Description: description}
	}
	return payload
}

// registerFields returns the same valid registration values as
// newCustomerRegisterPayload, keyed by wire field name (JSON key) instead
// of typed struct fields. Used only by the "register with invalid
// details" Scenario Outline, whose Examples table names the field to
// corrupt by its JSON key (e.g. "phone_number") and needs to overwrite
// exactly that one key without touching the rest.
func registerFields(username, email, password string) map[string]any {
	return map[string]any{
		"username":     username,
		"email":        email,
		"password":     password,
		"first_name":   validFirstName,
		"last_name":    validLastName,
		"phone_number": validPhoneNumber,
		"role":         "customer",
		"bank_account": map[string]any{
			"bank_name":      validBankName,
			"account_number": validAccountNo,
		},
	}
}
