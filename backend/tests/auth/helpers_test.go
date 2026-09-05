//go:build integration

package auth

import (
	"encoding/json"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
)

// Deliberately-wrong test data for negative login scenarios — not part of
// the register wire shape, so these stay local rather than living in
// apptest alongside the shared account/register fixtures.
const (
	wrongPassword = "wrong-password-1"
	unknownEmail  = "nobody@example.com"
)

// registerFields returns the same valid registration values as
// apptest.NewCustomerRegisterBody, keyed by wire field name (JSON key)
// instead of typed struct fields. Used only by the "register with invalid
// details" Scenario Outline, whose Examples table names the field to
// corrupt by its JSON key (e.g. "phone_number") and needs to overwrite
// exactly that one key without touching the rest.
func registerFields(username, email, password string) (map[string]any, error) {
	raw, err := json.Marshal(apptest.NewCustomerRegisterBody(username, email, password))
	if err != nil {
		return nil, err
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}
