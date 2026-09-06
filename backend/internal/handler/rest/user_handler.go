package rest

import (
	"context"
	"net/http"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	userUsecase  user.UserUsecase
	authUsecase  auth.AuthUsecase
	basePath     string
	isProduction bool
	cookieDomain string
}

func NewUserHandler(userUsecase user.UserUsecase, authUsecase auth.AuthUsecase, basePath string, isProduction bool, cookieDomain string) *UserHandler {
	return &UserHandler{
		userUsecase:  userUsecase,
		authUsecase:  authUsecase,
		basePath:     basePath,
		isProduction: isProduction,
		cookieDomain: cookieDomain,
	}
}

func (h *UserHandler) Register(api huma.API) {
	huma.Get(api, "/users/me", h.getAccount,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "get-account"
			o.Summary = "GetAccount"
			o.Description = "Get the authenticated user's account information"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)

	huma.Put(api, "/users/me", h.updateAccount,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "update-account"
			o.Summary = "UpdateAccount"
			o.Description = "Update the authenticated user's username and optionally change their password"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)

	huma.Put(api, "/users/me/bank-account", h.updateBankAccount,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "update-bank-account"
			o.Summary = "UpdateBankAccount"
			o.Description = "Update the authenticated user's bank account"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)

	huma.Delete(api, "/users/me", h.deleteAccount,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "delete-account"
			o.Summary = "DeleteAccount"
			o.Description = "Delete the authenticated user's account when they have no active orders"
			o.DefaultStatus = http.StatusNoContent
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)
}

type DeleteAccountInput struct{}

type DeleteAccountOutput struct {
	SetCookie string `header:"Set-Cookie"`
}

func (h *UserHandler) deleteAccount(ctx context.Context, _ *DeleteAccountInput) (*DeleteAccountOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	if err := h.userUsecase.DeleteAccount(ctx, info.UserID); err != nil {
		return nil, mapAppError(err)
	}

	return &DeleteAccountOutput{
		SetCookie: clearRefreshCookie(h.basePath, h.isProduction, h.cookieDomain),
	}, nil
}

type accountView struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     user.Role `json:"role"`
}

type GetAccountInput struct{}

type GetAccountOutput struct {
	Body accountView
}

func (h *UserHandler) getAccount(ctx context.Context, _ *GetAccountInput) (*GetAccountOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	account, err := h.userUsecase.GetByID(ctx, info.UserID)
	if err != nil {
		return nil, mapAppError(err)
	}
	return &GetAccountOutput{Body: newAccountView(account)}, nil
}

type updateAccountInput struct {
	Body struct {
		Username    string  `json:"username" minLength:"3" maxLength:"20"`
		OldPassword *string `json:"old_password,omitempty"`
		NewPassword *string `json:"new_password,omitempty" minLength:"8" maxLength:"16"`
	}
}

type UpdateAccountOutput struct {
	Body accountView
}

func (h *UserHandler) updateAccount(ctx context.Context, in *updateAccountInput) (*UpdateAccountOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	account, err := h.userUsecase.UpdateAccount(ctx, info.UserID, user.UpdateAccountInput{
		Username:    in.Body.Username,
		OldPassword: in.Body.OldPassword,
		NewPassword: in.Body.NewPassword,
	})
	if err != nil {
		return nil, mapAppError(err)
	}
	return &UpdateAccountOutput{Body: newAccountView(account)}, nil
}

func newAccountView(account *user.User) accountView {
	return accountView{
		ID:       account.ID,
		Username: account.Username,
		Email:    account.Email,
		Role:     account.Role,
	}
}

type updateBankAccountInput struct {
	Body struct {
		BankName          string `json:"bank_name" minLength:"1"`
		AccountHolderName string `json:"account_holder_name" minLength:"1"`
		AccountNumber     string `json:"account_number" minLength:"1"`
	}
}

type bankAccountView struct {
	BankName          string `json:"bank_name"`
	AccountHolderName string `json:"account_holder_name"`
	AccountLast4      string `json:"account_last4"`
}

type UpdateBankAccountOutput struct {
	Body bankAccountView
}

// TODO(payment): In the payment-gateway sprint, submit this destination to the
// selected provider and expose its verification status. This settings endpoint
// intentionally only stores user-entered details today because provider
// selection, recipient verification, and payout processing are not in scope.
func (h *UserHandler) updateBankAccount(ctx context.Context, in *updateBankAccountInput) (*UpdateBankAccountOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	bank, err := h.userUsecase.UpdateBankAccount(ctx, info.UserID, info.Role, user.BankAccountInput{
		BankName:          in.Body.BankName,
		AccountHolderName: in.Body.AccountHolderName,
		AccountNumber:     in.Body.AccountNumber,
	})
	if err != nil {
		return nil, mapAppError(err)
	}

	return &UpdateBankAccountOutput{Body: bankAccountView{
		BankName:          bank.BankName,
		AccountHolderName: bank.AccountHolderName,
		AccountLast4:      maskAccountNumber(bank.AccountNumber),
	}}, nil
}

func maskAccountNumber(accountNumber string) string {
	const visibleCharacters = 4

	characters := []rune(accountNumber)
	if len(characters) <= visibleCharacters {
		return "••••"
	}

	return "••••" + string(characters[len(characters)-visibleCharacters:])
}
