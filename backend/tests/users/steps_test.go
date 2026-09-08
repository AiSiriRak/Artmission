//go:build integration

package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/security"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

type usersContext struct {
	client       *apptest.Client
	account      apptest.Account
	accessToken  string
	originalBank *bankAccountRecord
	resp         *apptest.Response
}

type updateAccountBody struct {
	Username    string  `json:"username"`
	OldPassword *string `json:"old_password,omitempty"`
	NewPassword *string `json:"new_password,omitempty"`
}

type accountResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type accountRecord struct {
	Username     string `bun:"username"`
	Email        string `bun:"email"`
	PasswordHash string `bun:"password_hash"`
	Role         string `bun:"role"`
}

type bankAccountBody struct {
	BankName          string `json:"bank_name"`
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
}

type bankAccountResponse struct {
	BankName          string `json:"bank_name"`
	AccountHolderName string `json:"account_holder_name"`
	AccountLast4      string `json:"account_last4"`
}

type bankAccountRecord struct {
	BankName          string    `bun:"bank_name"`
	AccountHolderName string    `bun:"account_holder_name"`
	AccountNumber     string    `bun:"account_number"`
	CreatedAt         time.Time `bun:"created_at"`
	UpdatedAt         time.Time `bun:"updated_at"`
}

func (u *usersContext) theUserHasARegisteredAccount() error {
	account, err := apptest.RegisterCustomer(app, u.client)
	if err != nil {
		return err
	}
	u.account = account
	bank, err := u.loadBankAccount()
	if err != nil {
		return err
	}
	u.originalBank = bank
	return nil
}

func (u *usersContext) theUserHasLoggedIn() error {
	accessToken, err := apptest.Login(u.client, u.account.Email, u.account.Password)
	if err != nil {
		return err
	}
	u.accessToken = accessToken
	return nil
}

func (u *usersContext) theUserHasNoSavedBankAccount() error {
	userID, err := uuid.Parse(u.account.ID)
	if err != nil {
		return fmt.Errorf("parse fixture user ID: %w", err)
	}
	_, err = app.DB.NewDelete().Table("bank_accounts").Where("user_id = ?", userID).Exec(context.Background())
	if err == nil {
		u.originalBank = nil
	}
	return err
}

func (u *usersContext) loadBankAccount() (*bankAccountRecord, error) {
	userID, err := uuid.Parse(u.account.ID)
	if err != nil {
		return nil, fmt.Errorf("parse fixture user ID: %w", err)
	}
	bank := new(bankAccountRecord)
	err = app.DB.NewSelect().Table("bank_accounts").Column("bank_name", "account_holder_name", "account_number", "created_at", "updated_at").Where("user_id = ?", userID).Scan(context.Background(), bank)
	if err != nil {
		return nil, err
	}
	return bank, nil
}

func (u *usersContext) theUserUpdatesTheirBankAccountWithValidDetails() error {
	return u.updateBankAccount(bankAccountBody{BankName: "Kasikorn", AccountHolderName: "Test User", AccountNumber: "1234567890"}, u.accessToken)
}

func (u *usersContext) theUserUpdatesABankAccountWithoutLoggingIn() error {
	return u.updateBankAccount(bankAccountBody{BankName: "Kasikorn", AccountHolderName: "Test User", AccountNumber: "1234567890"}, "")
}

func (u *usersContext) theUserViewsTheirBankAccount() error {
	return u.getBankAccount(u.accessToken)
}

func (u *usersContext) theUserViewsABankAccountWithoutLoggingIn() error {
	return u.getBankAccount("")
}

func (u *usersContext) theUserUpdatesTheirBankAccountWithABlankBankName() error {
	return u.updateBankAccount(bankAccountBody{BankName: "", AccountHolderName: "Test User", AccountNumber: "1234567890"}, u.accessToken)
}

func (u *usersContext) updateBankAccount(body bankAccountBody, accessToken string) error {
	headers := map[string]string{}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	resp, err := u.client.Do(http.MethodPut, "/users/me/bank-account", body, headers)
	if err != nil {
		return err
	}
	u.resp = resp
	return nil
}

func (u *usersContext) theUserRequestsTheirAccountInformation() error {
	return u.getAccount(u.accessToken)
}

func (u *usersContext) theUserRequestsAccountInformationWithoutLoggingIn() error {
	return u.getAccount("")
}

func (u *usersContext) getAccount(accessToken string) error {
	headers := map[string]string{}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	resp, err := u.client.Do(http.MethodGet, "/users/me", nil, headers)
	if err != nil {
		return err
	}
	u.resp = resp
	return nil
}

func (u *usersContext) theSystemReturnsTheirSafeAccountInformation() error {
	if u.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	var body accountResponse
	if err := u.resp.JSON(&body); err != nil {
		return fmt.Errorf("decode account response: %w", err)
	}
	if body.ID != u.account.ID || body.Username != u.account.Username || body.Email != u.account.Email || body.Role != "customer" {
		return fmt.Errorf("unexpected account response: %+v", body)
	}
	var fields map[string]json.RawMessage
	if err := u.resp.JSON(&fields); err != nil {
		return fmt.Errorf("decode account response fields: %w", err)
	}
	for _, field := range []string{"password", "password_hash", "old_password", "new_password"} {
		if _, exists := fields[field]; exists {
			return fmt.Errorf("account response exposed %q", field)
		}
	}
	return nil
}

func (u *usersContext) theUserUpdatesTheirUsername() error {
	return u.updateAccount(updateAccountBody{Username: "  updated-name  "})
}

func (u *usersContext) theUserChangesTheirPasswordWithTheCorrectOldPassword() error {
	oldPassword := u.account.Password
	newPassword := "different2"
	return u.updateAccount(updateAccountBody{Username: u.account.Username, OldPassword: &oldPassword, NewPassword: &newPassword})
}

func (u *usersContext) theUserChangesTheirPasswordWithAnIncorrectOldPassword() error {
	oldPassword := "incorrect"
	newPassword := "different2"
	return u.updateAccount(updateAccountBody{Username: "should-not-save", OldPassword: &oldPassword, NewPassword: &newPassword})
}

func (u *usersContext) theUserSendsOnlyANewPassword() error {
	newPassword := "different2"
	return u.updateAccount(updateAccountBody{Username: u.account.Username, NewPassword: &newPassword})
}

func (u *usersContext) updateAccount(body updateAccountBody) error {
	resp, err := u.client.Do(http.MethodPut, "/users/me", body, map[string]string{"Authorization": "Bearer " + u.accessToken})
	if err != nil {
		return err
	}
	u.resp = resp
	return nil
}

func (u *usersContext) loadAccount() (*accountRecord, error) {
	userID, err := uuid.Parse(u.account.ID)
	if err != nil {
		return nil, fmt.Errorf("parse fixture user ID: %w", err)
	}
	record := new(accountRecord)
	err = app.DB.NewSelect().Table("users").Column("username", "email", "password_hash", "role").Where("id = ?", userID).Scan(context.Background(), record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (u *usersContext) theSystemSavesAndReturnsTheNewUsername() error {
	if u.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	var body accountResponse
	if err := u.resp.JSON(&body); err != nil {
		return fmt.Errorf("decode account response: %w", err)
	}
	if body.Username != "updated-name" || body.Email != u.account.Email || body.Role != "customer" {
		return fmt.Errorf("unexpected updated account response: %+v", body)
	}
	record, err := u.loadAccount()
	if err != nil {
		return err
	}
	if record.Username != "updated-name" || record.Email != u.account.Email || record.Role != "customer" {
		return fmt.Errorf("unexpected persisted account: %+v", record)
	}
	if !security.VerifyPassword(record.PasswordHash, u.account.Password) {
		return fmt.Errorf("username update changed the password")
	}
	return nil
}

func (u *usersContext) theNewPasswordWorksAndTheCurrentSessionRemainsValid() error {
	if u.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	if err := u.getAccount(u.accessToken); err != nil {
		return err
	}
	if u.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected current session to remain valid, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	if _, err := apptest.Login(apptest.NewClient(app.BaseURL()), u.account.Email, "different2"); err != nil {
		return fmt.Errorf("log in with new password: %w", err)
	}
	oldLogin, err := apptest.NewClient(app.BaseURL()).Do(http.MethodPost, "/auth/login", map[string]string{
		"email": u.account.Email, "password": u.account.Password,
	}, nil)
	if err != nil {
		return err
	}
	if oldLogin.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected old password login to return 401, got %d: %s", oldLogin.StatusCode, oldLogin.Body)
	}
	return nil
}

func (u *usersContext) theAccountInformationUpdateIsUnauthorizedAndNoChangesAreSaved() error {
	if u.resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected status 401, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	record, err := u.loadAccount()
	if err != nil {
		return err
	}
	if record.Username != u.account.Username || !security.VerifyPassword(record.PasswordHash, u.account.Password) {
		return fmt.Errorf("account changed after incorrect old password: %+v", record)
	}
	return nil
}

func (u *usersContext) theSystemRejectsTheAccountInformationUpdate() error {
	if u.resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("expected status 400, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	return nil
}

func (u *usersContext) getBankAccount(accessToken string) error {
	headers := map[string]string{}
	if accessToken != "" {
		headers["Authorization"] = "Bearer " + accessToken
	}
	resp, err := u.client.Do(http.MethodGet, "/users/me/bank-account", nil, headers)
	if err != nil {
		return err
	}
	u.resp = resp
	return nil
}

func (u *usersContext) theSystemSavesTheUpdatedBankAccountDetails() error {
	if u.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	var body bankAccountResponse
	if err := u.resp.JSON(&body); err != nil {
		return fmt.Errorf("decode bank account response: %w", err)
	}
	if body.BankName != "Kasikorn" || body.AccountHolderName != "Test User" || body.AccountLast4 != "••••7890" {
		return fmt.Errorf("unexpected updated bank account: %+v", body)
	}
	bank, err := u.loadBankAccount()
	if err != nil {
		return fmt.Errorf("read persisted bank account: %w", err)
	}
	if bank.BankName != "Kasikorn" || bank.AccountHolderName != "Test User" || bank.AccountNumber != "1234567890" {
		return fmt.Errorf("unexpected persisted bank account: %+v", bank)
	}
	if bank.CreatedAt.IsZero() || bank.UpdatedAt.IsZero() {
		return fmt.Errorf("expected persisted timestamps, got %+v", bank)
	}
	if u.originalBank != nil {
		if !bank.CreatedAt.Equal(u.originalBank.CreatedAt) {
			return fmt.Errorf("created_at changed from %v to %v", u.originalBank.CreatedAt, bank.CreatedAt)
		}
		if !bank.UpdatedAt.After(u.originalBank.UpdatedAt) {
			return fmt.Errorf("updated_at = %v, want after %v", bank.UpdatedAt, u.originalBank.UpdatedAt)
		}
	}
	return nil
}

func (u *usersContext) theSystemReturnsTheirMaskedBankAccountDetails() error {
	if u.resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	var body bankAccountResponse
	if err := u.resp.JSON(&body); err != nil {
		return fmt.Errorf("decode bank account response: %w", err)
	}
	if body.BankName != "Test Bank" || body.AccountHolderName != "Test User" || body.AccountLast4 != "••••7890" {
		return fmt.Errorf("unexpected bank account response: %+v", body)
	}
	return nil
}

func (u *usersContext) theSystemRequiresTheUserToLogIn() error {
	if u.resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected status 401, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	return nil
}

func (u *usersContext) theSystemRejectsTheBankAccountUpdate() error {
	if u.resp.StatusCode < 400 || u.resp.StatusCode >= 500 {
		return fmt.Errorf("expected a 4xx response, got %d: %s", u.resp.StatusCode, u.resp.Body)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var u *usersContext

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		u = &usersContext{client: apptest.NewClient(app.BaseURL())}
		return ctx, nil
	})

	sc.Step(`^the user has a registered account$`, func() error { return u.theUserHasARegisteredAccount() })
	sc.Step(`^the user has logged in$`, func() error { return u.theUserHasLoggedIn() })
	sc.Step(`^the user has no saved bank account$`, func() error { return u.theUserHasNoSavedBankAccount() })
	sc.Step(`^the user updates their bank account with valid details$`, func() error { return u.theUserUpdatesTheirBankAccountWithValidDetails() })
	sc.Step(`^the user updates a bank account without logging in$`, func() error { return u.theUserUpdatesABankAccountWithoutLoggingIn() })
	sc.Step(`^the user views their bank account$`, func() error { return u.theUserViewsTheirBankAccount() })
	sc.Step(`^the user views a bank account without logging in$`, func() error { return u.theUserViewsABankAccountWithoutLoggingIn() })
	sc.Step(`^the user updates their bank account with a blank bank name$`, func() error { return u.theUserUpdatesTheirBankAccountWithABlankBankName() })
	sc.Step(`^the user requests their account information$`, func() error { return u.theUserRequestsTheirAccountInformation() })
	sc.Step(`^the user requests account information without logging in$`, func() error { return u.theUserRequestsAccountInformationWithoutLoggingIn() })
	sc.Step(`^the system returns their safe account information$`, func() error { return u.theSystemReturnsTheirSafeAccountInformation() })
	sc.Step(`^the user updates their username$`, func() error { return u.theUserUpdatesTheirUsername() })
	sc.Step(`^the system saves and returns the new username$`, func() error { return u.theSystemSavesAndReturnsTheNewUsername() })
	sc.Step(`^the user changes their password with the correct old password$`, func() error { return u.theUserChangesTheirPasswordWithTheCorrectOldPassword() })
	sc.Step(`^the new password works and the current session remains valid$`, func() error { return u.theNewPasswordWorksAndTheCurrentSessionRemainsValid() })
	sc.Step(`^the user changes their password with an incorrect old password$`, func() error { return u.theUserChangesTheirPasswordWithAnIncorrectOldPassword() })
	sc.Step(`^the account information update is unauthorized and no changes are saved$`, func() error { return u.theAccountInformationUpdateIsUnauthorizedAndNoChangesAreSaved() })
	sc.Step(`^the user sends only a new password$`, func() error { return u.theUserSendsOnlyANewPassword() })
	sc.Step(`^the system rejects the account information update$`, func() error { return u.theSystemRejectsTheAccountInformationUpdate() })
	sc.Step(`^the system saves the updated bank account details$`, func() error { return u.theSystemSavesTheUpdatedBankAccountDetails() })
	sc.Step(`^the system returns their masked bank account details$`, func() error { return u.theSystemReturnsTheirMaskedBankAccountDetails() })
	sc.Step(`^the system requires the user to log in$`, func() error { return u.theSystemRequiresTheUserToLogIn() })
	sc.Step(`^the system rejects the bank account update$`, func() error { return u.theSystemRejectsTheBankAccountUpdate() })
}
