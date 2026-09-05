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
	huma.Post(api, "/auth/register", h.register,
		huma.OperationTags("auth"),
		func(o *huma.Operation) {
			o.OperationID = "register"
			o.Summary = "Register"
			o.Description = "Register a new account"
			o.DefaultStatus = http.StatusCreated
		},
	)

	huma.Post(api, "/auth/login", h.login,
		huma.OperationTags("auth"),
		func(o *huma.Operation) {
			o.OperationID = "login"
			o.Summary = "Login"
			o.Description = "Log in with email and password"
		},
	)

	huma.Post(api, "/auth/refresh", h.refresh,
		huma.OperationTags("auth"),
		func(o *huma.Operation) {
			o.OperationID = "refresh"
			o.Summary = "Refresh"
			o.Description = "Exchange a refresh token (cookie) for a new access token"
		},
	)

	huma.Post(api, "/auth/logout", h.logout,
		huma.OperationTags("auth"),
		func(o *huma.Operation) {
			o.OperationID = "logout"
			o.Summary = "Logout"
			o.Description = "Log out and terminate the current session"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase))
		},
	)
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
	Role        string                  `json:"role" enum:"customer,artist"`
	BankAccount registerBankAccountBody `json:"bank_account"`
	Artist      *registerArtistBody     `json:"artist,omitempty"`
}

type RegisterInput struct {
	Body registerInputBody
}

type RegisterOutput struct{}

func (h *AuthHandler) register(ctx context.Context, in *RegisterInput) (*RegisterOutput, error) {
	input := user.RegisterInput{
		Username: in.Body.Username,
		Email:    in.Body.Email,
		Password: in.Body.Password,
		Role:     user.Role(in.Body.Role),
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

	return &RegisterOutput{}, nil
}

// --- login ---

type LoginInput struct {
	Body struct {
		Email    string `json:"email" format:"email"`
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
	result, err := h.authUsecase.Login(ctx, in.Body.Email, in.Body.Password)
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
