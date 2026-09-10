package rest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

const validArtworkBody = `{
	"name":"Frontend draft",
	"category":"Illustration",
	"styles":[],
	"description":"Different from the dummy response",
	"artwork_samples":[],
	"minimum_deadline_days":1,
	"price_satang":0
}`

type artworkAuthStub struct {
	role user.Role
}

func (s artworkAuthStub) Login(context.Context, string, string) (*auth.AuthResult, error) {
	return nil, nil
}

func (s artworkAuthStub) Refresh(context.Context, string) (*auth.AuthResult, error) {
	return nil, nil
}

func (s artworkAuthStub) Logout(context.Context, uuid.UUID) error {
	return nil
}

func (s artworkAuthStub) Authenticate(context.Context, string) (*auth.TokenClaims, error) {
	return &auth.TokenClaims{
		UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		SessionID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		Role:      s.role,
	}, nil
}

func TestArtworkHandlerCreateReturnsFixedResponse(t *testing.T) {
	handler := newArtworkTestHandler(t, user.RoleArtist)
	rec := serveArtworkRequest(handler, http.MethodPost, "/artworks", validArtworkBody, true)

	if rec.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got artworkView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if want := dummyArtwork(); !reflect.DeepEqual(got, want) {
		t.Errorf("response body = %+v, want %+v", got, want)
	}
}

func TestArtworkHandlerDeleteReturnsNoContent(t *testing.T) {
	handler := newArtworkTestHandler(t, user.RoleArtist)
	rec := serveArtworkRequest(handler, http.MethodDelete, "/artworks/00000000-0000-0000-0000-000000000003", "", true)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("response body = %q, want empty", rec.Body.String())
	}
}

func TestArtworkHandlerRequiresAuthentication(t *testing.T) {
	handler := newArtworkTestHandler(t, user.RoleArtist)
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create", method: http.MethodPost, path: "/artworks", body: validArtworkBody},
		{name: "delete", method: http.MethodDelete, path: "/artworks/00000000-0000-0000-0000-000000000003"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveArtworkRequest(handler, tt.method, tt.path, tt.body, false)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("response status = %d, want %d: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestArtworkHandlerRequiresArtistRole(t *testing.T) {
	tests := []struct {
		name   string
		role   user.Role
		method string
		path   string
		body   string
	}{
		{name: "customer create", role: user.RoleCustomer, method: http.MethodPost, path: "/artworks", body: validArtworkBody},
		{name: "customer delete", role: user.RoleCustomer, method: http.MethodDelete, path: "/artworks/00000000-0000-0000-0000-000000000003"},
		{name: "admin create", role: user.RoleAdmin, method: http.MethodPost, path: "/artworks", body: validArtworkBody},
		{name: "admin delete", role: user.RoleAdmin, method: http.MethodDelete, path: "/artworks/00000000-0000-0000-0000-000000000003"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newArtworkTestHandler(t, tt.role)
			rec := serveArtworkRequest(handler, tt.method, tt.path, tt.body, true)
			if rec.Code != http.StatusForbidden {
				t.Errorf("response status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
			}
		})
	}
}

func newArtworkTestHandler(t *testing.T, role user.Role) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api, server := httpserver.New("", "/api/v1", nil, logger, nil)
	NewArtworkHandler(artworkAuthStub{role: role}).Register(api)
	return server.Handler()
}

func serveArtworkRequest(handler http.Handler, method, path, body string, authenticated bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authenticated {
		req.Header.Set("Authorization", "Bearer valid-token")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
