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

type registerInputBody struct {
	Username string `json:"username" minLength:"3" maxLength:"32"`
	Email    string `json:"email" format:"email"`
	Phone    string `json:"phone" minLength:"1"`
	Password string `json:"password" minLength:"8"`
	Role     string `json:"role" enum:"customer,artist"`
}

type RegisterInput struct {
	Body registerInputBody
}

type userView struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterOutput struct {
	Status int
	Body   userView
}

func (h *AuthHandler) register(ctx context.Context, in *RegisterInput) (*RegisterOutput, error) {
	created, err := h.userUsecase.Register(ctx, user.RegisterInput{
		Username: in.Body.Username,
		Email:    in.Body.Email,
		Phone:    in.Body.Phone,
		Password: in.Body.Password,
		Role:     user.Role(in.Body.Role),
	})
	if err != nil {
		return nil, mapAppError(err)
	}

	return &RegisterOutput{Status: http.StatusCreated, Body: toUserView(created)}, nil
}

// --- login ---

type LoginInput struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
}

type authResultBody struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	User                 userView  `json:"user"`
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
		Body:      toAuthResultBody(result),
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
		Body:      toAuthResultBody(result),
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

func toUserView(u *user.User) userView {
	return userView{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
}

func toAuthResultBody(r *auth.AuthResult) authResultBody {
	return authResultBody{
		AccessToken:          r.AccessToken,
		AccessTokenExpiresAt: r.AccessTokenExpiresAt,
		User:                 toUserView(r.User),
	}
}
