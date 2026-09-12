package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

const validArtworkBody = `{
	"name":"Frontend draft",
	"category":"Illustration",
	"styles":[],
	"description":"A custom editorial illustration",
	"artwork_samples":[],
	"minimum_deadline_days":1,
	"price_satang":0
}`

type artworkAuthStub struct {
	role user.Role
}

type artworkUsecaseStub struct {
	artworks       []artwork.Artwork
	created        *artwork.Artwork
	updated        *artwork.Artwork
	err            error
	createInput    artwork.CreateInput
	updateInput    artwork.UpdateInput
	deleteArtistID uuid.UUID
	deleteArtwork  uuid.UUID
}

func (stub *artworkUsecaseStub) ListByArtistID(context.Context, uuid.UUID) ([]artwork.Artwork, error) {
	return stub.artworks, stub.err
}

func (stub *artworkUsecaseStub) Create(_ context.Context, input artwork.CreateInput) (*artwork.Artwork, error) {
	stub.createInput = input
	return stub.created, stub.err
}

func (stub *artworkUsecaseStub) Update(_ context.Context, input artwork.UpdateInput) (*artwork.Artwork, error) {
	stub.updateInput = input
	return stub.updated, stub.err
}

func (stub *artworkUsecaseStub) Delete(_ context.Context, artistID, artworkID uuid.UUID) error {
	stub.deleteArtistID = artistID
	stub.deleteArtwork = artworkID
	return stub.err
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

func TestArtworkHandlerCreateReturnsCreatedArtwork(t *testing.T) {
	usecase := &artworkUsecaseStub{created: testArtwork()}
	handler := newArtworkTestHandlerWithUsecase(t, user.RoleArtist, usecase)
	rec := serveArtworkRequest(handler, http.MethodPost, "/artworks", validArtworkBody, true)

	if rec.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got artworkView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if want := newArtworkView(usecase.created); !reflect.DeepEqual(got, want) {
		t.Errorf("response body = %+v, want %+v", got, want)
	}
	if usecase.createInput.ArtistID == uuid.Nil || usecase.createInput.Name != "Frontend draft" || usecase.createInput.Category != "Illustration" || len(usecase.createInput.SampleFiles) != 1 {
		t.Errorf("create input = %+v", usecase.createInput)
	}
}

func TestArtworkHandlerDeleteReturnsNoContent(t *testing.T) {
	usecase := &artworkUsecaseStub{created: testArtwork()}
	handler := newArtworkTestHandlerWithUsecase(t, user.RoleArtist, usecase)
	artworkID := "00000000-0000-0000-0000-000000000003"
	rec := serveArtworkRequest(handler, http.MethodDelete, "/artworks/"+artworkID, "", true)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("response body = %q, want empty", rec.Body.String())
	}
	if usecase.deleteArtistID == uuid.Nil || usecase.deleteArtwork.String() != artworkID {
		t.Errorf("delete input = artist=%s artwork=%s", usecase.deleteArtistID, usecase.deleteArtwork)
	}
}

func TestArtworkHandlerUpdateReturnsUpdatedArtwork(t *testing.T) {
	updated := testArtwork()
	updated.Name = "Updated artwork"
	usecase := &artworkUsecaseStub{updated: updated}
	handler := newArtworkTestHandlerWithUsecase(t, user.RoleArtist, usecase)
	artworkID := "00000000-0000-0000-0000-000000000003"
	rec := serveArtworkRequest(handler, http.MethodPut, "/artworks/"+artworkID, validArtworkBody, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got artworkView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if want := newArtworkView(updated); !reflect.DeepEqual(got, want) {
		t.Errorf("response body = %+v, want %+v", got, want)
	}
	if usecase.updateInput.ArtworkID.String() != artworkID || usecase.updateInput.ArtistID == uuid.Nil || usecase.updateInput.Name != "Frontend draft" || len(usecase.updateInput.SampleFiles) != 1 || !reflect.DeepEqual(usecase.updateInput.DeletedSampleURLs, []string{"https://storage.example.com/old.png"}) {
		t.Errorf("update input = %+v", usecase.updateInput)
	}
}

func TestArtworkHandlerListsArtistArtworksWithoutAuthentication(t *testing.T) {
	artworkID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	handler := newArtworkTestHandlerWithUsecase(t, user.RoleArtist, &artworkUsecaseStub{artworks: []artwork.Artwork{{
		ID:                  artworkID,
		Name:                "Watercolor portrait",
		Category:            "Portrait",
		Styles:              []string{"Realism", "Watercolor"},
		Description:         "Painted portrait",
		Samples:             []artwork.Sample{{ImageURL: "https://storage.example.com/sample.webp"}},
		MinimumDeadlineDays: 7,
		PriceSatang:         50000,
	}}})
	rec := serveArtworkRequest(handler, http.MethodGet, "/artists/00000000-0000-0000-0000-000000000001/artworks", "", false)

	if rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Artworks []artistArtworkView `json:"artworks"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := artistArtworkView{
		ArtworkID:           artworkID,
		Name:                "Watercolor portrait",
		Category:            "Portrait",
		Styles:              []string{"Realism", "Watercolor"},
		Description:         "Painted portrait",
		ArtworkSamples:      []artworkSampleView{{ImageURL: "https://storage.example.com/sample.webp"}},
		MinimumDeadlineDays: 7,
		PriceSatang:         50000,
	}
	if len(got.Artworks) != 1 || !reflect.DeepEqual(got.Artworks[0], want) {
		t.Errorf("response body = %+v, want %+v", got.Artworks, []artistArtworkView{want})
	}
}

func TestArtworkHandlerReturnsNotFoundForMissingArtist(t *testing.T) {
	handler := newArtworkTestHandlerWithUsecase(t, user.RoleArtist, &artworkUsecaseStub{err: artwork.ErrArtistNotFound})
	rec := serveArtworkRequest(handler, http.MethodGet, "/artists/00000000-0000-0000-0000-000000000001/artworks", "", false)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusNotFound, rec.Body.String())
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
		{name: "update", method: http.MethodPut, path: "/artworks/00000000-0000-0000-0000-000000000003", body: validArtworkBody},
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
		{name: "customer update", role: user.RoleCustomer, method: http.MethodPut, path: "/artworks/00000000-0000-0000-0000-000000000003", body: validArtworkBody},
		{name: "customer delete", role: user.RoleCustomer, method: http.MethodDelete, path: "/artworks/00000000-0000-0000-0000-000000000003"},
		{name: "admin create", role: user.RoleAdmin, method: http.MethodPost, path: "/artworks", body: validArtworkBody},
		{name: "admin update", role: user.RoleAdmin, method: http.MethodPut, path: "/artworks/00000000-0000-0000-0000-000000000003", body: validArtworkBody},
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
	return newArtworkTestHandlerWithUsecase(t, role, &artworkUsecaseStub{created: testArtwork(), updated: testArtwork()})
}

func testArtwork() *artwork.Artwork {
	createdAt := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)
	return &artwork.Artwork{
		ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		ArtistID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:                "Frontend draft",
		Category:            "Illustration",
		Styles:              []string{},
		Description:         "A custom editorial illustration",
		Samples:             []artwork.Sample{},
		MinimumDeadlineDays: 1,
		PriceSatang:         0,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
	}
}

func newArtworkTestHandlerWithUsecase(t *testing.T, role user.Role, artworkUsecase artwork.Usecase) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api, server := httpserver.New("", "/api/v1", nil, logger, nil)
	NewArtworkHandler(artworkUsecase, artworkAuthStub{role: role}).Register(api)
	return server.Handler()
}

func serveArtworkRequest(handler http.Handler, method, path, body string, authenticated bool) *httptest.ResponseRecorder {
	var reader io.Reader = strings.NewReader(body)
	contentType := ""
	if body != "" && (method == http.MethodPost || method == http.MethodPut) {
		encoded, multipartContentType := validArtworkMultipartBody(method == http.MethodPut)
		reader = bytes.NewReader(encoded)
		contentType = multipartContentType
	}
	req := httptest.NewRequest(method, "/api/v1"+path, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if authenticated {
		req.Header.Set("Authorization", "Bearer valid-token")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func validArtworkMultipartBody(update bool) ([]byte, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range map[string]string{
		"name":                  "Frontend draft",
		"category":              "Illustration",
		"description":           "A custom editorial illustration",
		"minimum_deadline_days": "1",
		"price_satang":          "0",
	} {
		_ = writer.WriteField(name, value)
	}
	writeJSONPart(writer, "styles", `[]`)
	if update {
		writeJSONPart(writer, "deleted_sample_urls", `["https://storage.example.com/old.png"]`)
	}

	fileHeader := make(textproto.MIMEHeader)
	fileField := "artwork_samples"
	if update {
		fileField = "uploaded_samples"
	}
	fileHeader.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": fileField, "filename": "sample.png"}))
	fileHeader.Set("Content-Type", "image/png")
	file, _ := writer.CreatePart(fileHeader)
	_, _ = file.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00})
	_ = writer.Close()
	return body.Bytes(), writer.FormDataContentType()
}

func writeJSONPart(writer *multipart.Writer, name, value string) {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": name}))
	header.Set("Content-Type", "application/json")
	part, _ := writer.CreatePart(header)
	_, _ = part.Write([]byte(value))
}
