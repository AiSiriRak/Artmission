package rest

import (
	"context"
	"io"
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

// NewUserHandler configures account endpoints with their use cases and refresh-cookie settings.
func NewUserHandler(
	userUsecase user.UserUsecase,
	authUsecase auth.AuthUsecase,
	basePath string,
	isProduction bool,
	cookieDomain string,
) *UserHandler {
	return &UserHandler{
		userUsecase:  userUsecase,
		authUsecase:  authUsecase,
		basePath:     basePath,
		isProduction: isProduction,
		cookieDomain: cookieDomain,
	}
}

// Register adds the authenticated account, bank-account, deletion, and profile-image routes.
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

	huma.Get(api, "/users/me/bank-account", h.getBankAccount,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "get-bank-account"
			o.Summary = "GetBankAccount"
			o.Description = "Get the authenticated user's bank account"
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

	huma.Put(api, "/users/me/profile-image", h.updateProfileImage,
		huma.OperationTags("users"),
		func(o *huma.Operation) {
			o.OperationID = "update-profile-image"
			o.Summary = "UpdateProfileImage"
			o.Description = "Update or remove the authenticated user's profile image"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)
}

// GetAccountInput contains the request parameters for retrieving the authenticated user's account.
type GetAccountInput struct{}

// GetAccountOutput contains the authenticated user's account information.
type GetAccountOutput struct {
	Body accountView
}

// getAccount loads the authenticated user's account and converts it to the response view.
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

// updateAccountInput contains the username and optional password changes submitted by the user.
type updateAccountInput struct {
	Body struct {
		Username    string  `json:"username" minLength:"3" maxLength:"20"`
		OldPassword *string `json:"old_password,omitempty"`
		NewPassword *string `json:"new_password,omitempty" minLength:"8" maxLength:"16"`
	}
}

// UpdateAccountOutput contains the updated account information.
type UpdateAccountOutput struct {
	Body accountView
}

// updateAccount applies the submitted username and optional password change for the current user.
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

// getBankAccountInput contains the request parameters for retrieving the authenticated user's bank account.
type getBankAccountInput struct{}

// GetBankAccountOutput contains the authenticated user's bank account details.
type GetBankAccountOutput struct {
	Body bankAccountView
}

// getBankAccount returns the authenticated user's bank details with only the last four digits visible.
func (h *UserHandler) getBankAccount(ctx context.Context, _ *getBankAccountInput) (*GetBankAccountOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	bank, err := h.userUsecase.GetBankAccount(ctx, info.UserID, info.Role)
	if err != nil {
		return nil, mapAppError(err)
	}
	return &GetBankAccountOutput{Body: bankAccountView{
		BankName:          bank.BankName,
		AccountHolderName: bank.AccountHolderName,
		AccountLast4:      maskAccountNumber(bank.AccountNumber),
	}}, nil
}

// updateBankAccountInput contains the bank details submitted by the authenticated user.
type updateBankAccountInput struct {
	Body struct {
		BankName          string `json:"bank_name" minLength:"1"`
		AccountHolderName string `json:"account_holder_name" minLength:"1"`
		AccountNumber     string `json:"account_number" minLength:"1"`
	}
}

// UpdateBankAccountOutput contains the updated bank account details.
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

// DeleteAccountInput contains the request parameters for deleting the authenticated user's account.
type DeleteAccountInput struct{}

// DeleteAccountOutput contains the response header used to expire the refresh cookie.
type DeleteAccountOutput struct {
	SetCookie string `header:"Set-Cookie"`
}

// deleteAccount soft-deletes the authenticated user's account and expires the refresh cookie.
func (h *UserHandler) deleteAccount(
	ctx context.Context,
	_ *DeleteAccountInput,
) (*DeleteAccountOutput, error) {
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

// updateProfileImageForm contains the profile image upload and removal options.
type updateProfileImageForm struct {
	ProfileImage       huma.FormFile `form:"profile_image" contentType:"image/jpeg,image/png,image/webp" required:"false"`
	RemoveProfileImage bool          `form:"remove_profile_image" required:"false"`
}

// updateProfileImageInput contains the multipart form data for updating the user's profile image.
type updateProfileImageInput struct {
	RawBody huma.MultipartFormFiles[updateProfileImageForm]
}

// UpdateProfileImageOutput contains the updated account information after the profile image change.
type UpdateProfileImageOutput struct {
	Body accountView
}

// updateProfileImage validates multipart fields and updates or removes the current user's image.
func (h *UserHandler) updateProfileImage(ctx context.Context, in *updateProfileImageInput) (*UpdateProfileImageOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	form := in.RawBody.Data()
	for field := range in.RawBody.Form.Value {
		if field != "remove_profile_image" {
			return nil, huma.Error422UnprocessableEntity("unexpected multipart field: " + field)
		}
	}
	for field := range in.RawBody.Form.File {
		if field != "profile_image" {
			return nil, huma.Error422UnprocessableEntity("unexpected multipart field: " + field)
		}
	}

	var profileImage io.Reader
	if form.ProfileImage.IsSet {
		profileImage = form.ProfileImage.File
	}

	account, err := h.userUsecase.UpdateProfileImage(ctx, info.UserID, user.UpdateProfileImageInput{
		ProfileImage:       profileImage,
		RemoveProfileImage: form.RemoveProfileImage,
	})
	if err != nil {
		return nil, mapAppError(err)
	}

	return &UpdateProfileImageOutput{Body: newAccountView(account)}, nil
}

// accountView represents the account information exposed by the user endpoints.
type accountView struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	Role            user.Role `json:"role"`
	ProfileImageURL *string   `json:"profile_image_url"`
}

// bankAccountView represents the bank account information exposed by the API.
type bankAccountView struct {
	BankName          string `json:"bank_name"`
	AccountHolderName string `json:"account_holder_name"`
	AccountLast4      string `json:"account_last4"`
}

// newAccountView exposes the account fields returned by the account endpoints.
func newAccountView(account *user.User) accountView {
	return accountView{
		ID:              account.ID,
		Username:        account.Username,
		Email:           account.Email,
		Role:            account.Role,
		ProfileImageURL: account.ProfileImageURL,
	}
}

// maskAccountNumber hides all but the final four runes of an account number.
func maskAccountNumber(accountNumber string) string {
	const visibleCharacters = 4

	characters := []rune(accountNumber)
	if len(characters) <= visibleCharacters {
		return "••••"
	}

	return "••••" + string(characters[len(characters)-visibleCharacters:])
}
