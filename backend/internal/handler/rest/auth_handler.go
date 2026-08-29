package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
)

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	userUsecase  user.UserUsecase
	authUsecase  auth.AuthUsecase
	basePath     string
	isProduction bool
	cookieDomain string
}

func NewAuthHandler(userUsecase user.UserUsecase, authUsecase auth.AuthUsecase, basePath string, isProduction bool, cookieDomain string) *AuthHandler {
	return &AuthHandler{
		userUsecase:  userUsecase,
		authUsecase:  authUsecase,
		basePath:     basePath,
		isProduction: isProduction,
		cookieDomain: cookieDomain,
	}
}

func (h *AuthHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register a new account",
	}, h.register)

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Log in with username and password",
	}, h.login)

	huma.Register(api, huma.Operation{
		OperationID: "refresh",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Exchange a refresh token (cookie) for a new access token",
	}, h.refresh)

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/auth/logout",
		Summary:     "Log out and terminate the current session",
		Middlewares: huma.Middlewares{requireAuth(api, h.authUsecase)},
	}, h.logout)
}

// --- register ---

type registerBankAccountBody struct {
	BankName      string `json:"bank_name" minLength:"1"`
	AccountNumber string `json:"account_number" minLength:"1"`
}

type registerArtistBody struct {
	Description string `json:"description" minLength:"1"`
}

type registerInputBody struct {
	Username    string                  `json:"username" minLength:"3" maxLength:"20"`
	Email       string                  `json:"email" format:"email"`
	Password    string                  `json:"password" minLength:"8" maxLength:"16"`
	FirstName   string                  `json:"first_name" minLength:"1"`
	LastName    string                  `json:"last_name" minLength:"1"`
	PhoneNumber string                  `json:"phone_number" minLength:"1"`
	Role        string                  `json:"role" enum:"customer,artist"`
	BankAccount registerBankAccountBody `json:"bank_account"`
	Artist      *registerArtistBody     `json:"artist,omitempty"`
}

type RegisterInput struct {
	Body registerInputBody
}

type RegisterOutput struct {
	Status int
}

func (h *AuthHandler) register(ctx context.Context, in *RegisterInput) (*RegisterOutput, error) {
	input := user.RegisterInput{
		Username:    in.Body.Username,
		Email:       in.Body.Email,
		FirstName:   in.Body.FirstName,
		LastName:    in.Body.LastName,
		PhoneNumber: in.Body.PhoneNumber,
		Password:    in.Body.Password,
		Role:        user.Role(in.Body.Role),
		BankAccount: user.BankAccountInput{
			BankName:      in.Body.BankAccount.BankName,
			AccountNumber: in.Body.BankAccount.AccountNumber,
		},
	}
	if in.Body.Artist != nil {
		input.Artist = &user.ArtistProfileInput{Description: in.Body.Artist.Description}
	}

	if _, err := h.userUsecase.Register(ctx, input); err != nil {
		return nil, mapAppError(err)
	}

	return &RegisterOutput{Status: http.StatusCreated}, nil
}

// --- login ---

type LoginInput struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
}

type authResultBody struct {
	AccessToken string `json:"access_token"`
}

type LoginOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      authResultBody
}

func (h *AuthHandler) login(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
	result, err := h.authUsecase.Login(ctx, in.Body.Username, in.Body.Password)
	if err != nil {
		return nil, mapAppError(err)
	}

	return &LoginOutput{
		SetCookie: h.refreshCookie(result.RefreshToken, result.RefreshTokenExpiresAt),
		Body:      authResultBody{AccessToken: result.AccessToken},
	}, nil
}

// --- refresh ---

type RefreshInput struct {
	RefreshToken string `cookie:"refresh_token"`
}

type RefreshOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      authResultBody
}

func (h *AuthHandler) refresh(ctx context.Context, in *RefreshInput) (*RefreshOutput, error) {
	if in.RefreshToken == "" {
		return nil, huma.Error401Unauthorized("missing refresh token cookie")
	}

	result, err := h.authUsecase.Refresh(ctx, in.RefreshToken)
	if err != nil {
		return nil, mapAppError(err)
	}

	return &RefreshOutput{
		SetCookie: h.refreshCookie(result.RefreshToken, result.RefreshTokenExpiresAt),
		Body:      authResultBody{AccessToken: result.AccessToken},
	}, nil
}

// --- logout ---

type LogoutInput struct{}

type LogoutOutput struct {
	Status    int
	SetCookie string `header:"Set-Cookie"`
}

func (h *AuthHandler) logout(ctx context.Context, _ *LogoutInput) (*LogoutOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	if err := h.authUsecase.Logout(ctx, info.SessionID); err != nil {
		return nil, mapAppError(err)
	}

	return &LogoutOutput{
		Status:    http.StatusNoContent,
		SetCookie: h.clearRefreshCookie(),
	}, nil
}

// --- helpers ---

func (h *AuthHandler) refreshCookie(value string, expiresAt time.Time) string {
	c := &http.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     h.basePath + "/auth/refresh",
		Domain:   h.cookieDomain,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
	}
	return c.String()
}

func (h *AuthHandler) clearRefreshCookie() string {
	c := &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     h.basePath + "/auth/refresh",
		Domain:   h.cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteStrictMode,
	}
	return c.String()
}


