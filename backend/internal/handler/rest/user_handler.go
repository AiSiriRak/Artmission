package rest

import (
	"context"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
)

type UserHandler struct {
	userUsecase user.UserUsecase
	authUsecase auth.AuthUsecase
}

func NewUserHandler(userUsecase user.UserUsecase, authUsecase auth.AuthUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, authUsecase: authUsecase}
}

func (h *UserHandler) Register(api huma.API) {
	huma.Put(api, "/users/me/bank-account", h.updateBankAccount,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "update-bank-account"
			o.Summary = "UpdateBankAccount"
			o.Description = "Update the authenticated user's bank account"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)
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
