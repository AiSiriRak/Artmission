//go:build integration

package users

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	accessToken, err := apptest.Login(u.client, u.account.Username, u.account.Password)
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
	sc.Step(`^the user updates their bank account with a blank bank name$`, func() error { return u.theUserUpdatesTheirBankAccountWithABlankBankName() })
	sc.Step(`^the system saves the updated bank account details$`, func() error { return u.theSystemSavesTheUpdatedBankAccountDetails() })
	sc.Step(`^the system requires the user to log in$`, func() error { return u.theSystemRequiresTheUserToLogIn() })
	sc.Step(`^the system rejects the bank account update$`, func() error { return u.theSystemRejectsTheBankAccountUpdate() })
}
