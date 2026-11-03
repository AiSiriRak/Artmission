package rest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/review"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

type reviewAuthStub struct {
	auth.AuthUsecase
	role user.Role
}

type reviewUsecaseStub struct {
	review.ReviewUsecase
	submitCalls  int
	reviews      []review.ReviewDetail
	listCalls    int
	listArtistID uuid.UUID
	listArtworkID uuid.UUID
	orderReview    *review.ReviewDetail
	orderReviewErr error
	getOrderCalls  int
	getOrderID     uuid.UUID
	getOrderUserID uuid.UUID
	listQuery review.ListQuery
}

func (s reviewAuthStub) Authenticate(
	_ context.Context,
	token string,
) (*auth.TokenClaims, error) {
	if token != "test-token" {
		return nil, auth.ErrInvalidToken
	}

	return &auth.TokenClaims{
		UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		SessionID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		Role:      s.role,
	}, nil
}

func (s *reviewUsecaseStub) SubmitReview(
	_ context.Context,
	customerID uuid.UUID,
	input review.SubmitReviewInput,
) (*review.Review, error) {
	s.submitCalls++

	return &review.Review{
		ID:         uuid.New(),
		OrderID:    input.OrderID,
		CustomerID: customerID,
		ArtistID:   uuid.New(),
		Rating:     input.Rating,
		Comment:    input.Comment,
	}, nil
}

func (s *reviewUsecaseStub) GetArtistReviews(
	_ context.Context,
	artistID uuid.UUID,
	query review.ListQuery,
) (*review.ReviewPage, error) {
	s.listCalls++
	s.listArtistID = artistID
	s.listQuery = query

	return &review.ReviewPage{
		Reviews: s.reviews,
		Total:   len(s.reviews),
	}, nil
}

func (s *reviewUsecaseStub) GetReviewsByArtwork(
	_ context.Context,
	artworkID uuid.UUID,
	query review.ListQuery,
) (*review.ReviewPage, error) {
	s.listCalls++
	s.listArtworkID = artworkID
	s.listQuery = query

	return &review.ReviewPage{
		Reviews: s.reviews,
		Total:   len(s.reviews),
	}, nil
}

func (s *reviewUsecaseStub) GetOrderReview(
	_ context.Context,
	userID uuid.UUID,
	orderID uuid.UUID,
) (*review.ReviewDetail, error) {
	s.getOrderCalls++
	s.getOrderUserID = userID
	s.getOrderID = orderID

	return s.orderReview, s.orderReviewErr
}

// TestReviewHandler_SubmitValidation verifies rating validation and access restrictions.
func TestReviewHandler_SubmitValidation(t *testing.T) {
	cases := []struct {
		name          string
		role          user.Role
		body          string
		authenticated bool
		wantStatus    int
		wantCalls     int
	}{
		{
			name:          "customer submits rating",
			role:          user.RoleCustomer,
			body:          `{"rating":5}`,
			authenticated: true,
			wantStatus:    http.StatusCreated,
			wantCalls:     1,
		},
		{
			name:          "missing rating",
			role:          user.RoleCustomer,
			body:          `{"comment":"Great work!"}`,
			authenticated: true,
			wantStatus:    http.StatusUnprocessableEntity,
		},
		{
			name:       "missing token",
			role:       user.RoleCustomer,
			body:       `{"rating":5}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:          "artist cannot submit",
			role:          user.RoleArtist,
			body:          `{"rating":5}`,
			authenticated: true,
			wantStatus:    http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := &reviewUsecaseStub{}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			api, server := httpserver.New("", "/api/v1", nil, logger, nil)

			NewReviewHandler(
				usecase,
				reviewAuthStub{role: tc.role},
			).Register(api)

			orderID := uuid.New()
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/orders/"+orderID.String()+"/review",
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")

			if tc.authenticated {
				req.Header.Set("Authorization", "Bearer test-token")
			}

			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf(
					"status = %d, want %d: %s",
					rec.Code,
					tc.wantStatus,
					rec.Body.String(),
				)
			}
			if usecase.submitCalls != tc.wantCalls {
				t.Fatalf(
					"submit calls = %d, want %d",
					usecase.submitCalls,
					tc.wantCalls,
				)
			}
		})
	}
}

// TestReviewHandler_GetArtistReviews verifies access, response fields, and the empty state.
func TestReviewHandler_GetArtistReviews(t *testing.T) {
	artistID := uuid.New()
	reviewID := uuid.New()
	comment := "Great work!"
	createdAt := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)

	items := []review.ReviewDetail{
		{
			Review: review.Review{
				ID:        reviewID,
				ArtistID:  artistID,
				Rating:    5,
				Comment:   &comment,
				CreatedAt: createdAt,
			},
			Username: "Customer",
			Order:    "Cat artwork",
		},
	}

	cases := []struct {
		name          string
		role          user.Role
		authenticated bool
		empty         bool
		wantStatus    int
	}{
		{
			name:          "customer can read",
			role:          user.RoleCustomer,
			authenticated: true,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "artist can read",
			role:          user.RoleArtist,
			authenticated: true,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "no reviews",
			role:          user.RoleCustomer,
			authenticated: true,
			empty:         true,
			wantStatus:    http.StatusOK,
		},
		{
			name:       "missing token",
			role:       user.RoleCustomer,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := &reviewUsecaseStub{reviews: items}
			if tc.empty {
				usecase.reviews = nil
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			api, server := httpserver.New("", "/api/v1", nil, logger, nil)
			NewReviewHandler(
				usecase,
				reviewAuthStub{role: tc.role},
			).Register(api)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/artists/"+artistID.String()+"/reviews",
				nil,
			)
			if tc.authenticated {
				req.Header.Set("Authorization", "Bearer test-token")
			}

			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf(
					"status = %d, want %d: %s",
					rec.Code, tc.wantStatus, rec.Body.String(),
				)
			}

			if tc.wantStatus != http.StatusOK {
				if usecase.listCalls != 0 {
					t.Fatal("an unauthorized request must not reach the usecase")
				}
				return
			}

			if usecase.listCalls != 1 || usecase.listArtistID != artistID {
				t.Fatal("expected one request for the selected artist")
			}

			if usecase.listQuery.Limit != 20 || usecase.listQuery.Offset != 0 {
				t.Fatal("expected default pagination: limit 20 and offset 0")
			}

			var body struct {
				Reviews []reviewDetailView `json:"reviews"`
				Total   *int               `json:"total"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Reviews == nil {
				t.Fatal("reviews must be an array, not null or missing")
			}
			if body.Total == nil || *body.Total != len(body.Reviews) {
				t.Fatal("total must be present and match the number of reviews")
			}

			if tc.empty {
				if len(body.Reviews) != 0 {
					t.Fatal("expected an empty reviews array")
				}
				return
			}

			if len(body.Reviews) != 1 {
				t.Fatalf("review count = %d, want 1", len(body.Reviews))
			}

			got := body.Reviews[0]
			if got.ID != reviewID ||
				got.Username != "Customer" ||
				got.Order != "Cat artwork" ||
				got.Rating != 5 ||
				got.Comment == nil ||
				*got.Comment != comment ||
				!got.CreatedAt.Equal(createdAt) {
				t.Fatal("response does not contain the expected review details")
			}
		})
	}
}

// TestReviewHandler_GetReviewsByArtwork verifies access, response fields, and the empty state.
func TestReviewHandler_GetReviewsByArtwork(t *testing.T) {
	artworkID := uuid.New()
	reviewID := uuid.New()
	comment := "Great work!"
	createdAt := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)

	items := []review.ReviewDetail{
		{
			Review: review.Review{
				ID:        reviewID,
				ArtistID:  uuid.New(),
				Rating:    5,
				Comment:   &comment,
				CreatedAt: createdAt,
			},
			Username: "Customer",
			Order:    "Cat artwork",
		},
	}

	cases := []struct {
		name          string
		role          user.Role
		authenticated bool
		empty         bool
		wantStatus    int
	}{
		{
			name:          "customer can read",
			role:          user.RoleCustomer,
			authenticated: true,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "artist can read",
			role:          user.RoleArtist,
			authenticated: true,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "no reviews",
			role:          user.RoleCustomer,
			authenticated: true,
			empty:         true,
			wantStatus:    http.StatusOK,
		},
		{
			name:       "missing token",
			role:       user.RoleCustomer,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := &reviewUsecaseStub{reviews: items}
			if tc.empty {
				usecase.reviews = nil
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			api, server := httpserver.New("", "/api/v1", nil, logger, nil)
			NewReviewHandler(
				usecase,
				reviewAuthStub{role: tc.role},
			).Register(api)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/artworks/"+artworkID.String()+"/reviews",
				nil,
			)
			if tc.authenticated {
				req.Header.Set("Authorization", "Bearer test-token")
			}

			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf(
					"status = %d, want %d: %s",
					rec.Code, tc.wantStatus, rec.Body.String(),
				)
			}

			if tc.wantStatus != http.StatusOK {
				if usecase.listCalls != 0 {
					t.Fatal("an unauthorized request must not reach the usecase")
				}
				return
			}

			if usecase.listCalls != 1 || usecase.listArtworkID != artworkID {
				t.Fatal("expected one request for the selected artwork")
			}

			if usecase.listQuery.Limit != 20 || usecase.listQuery.Offset != 0 {
				t.Fatal("expected default pagination: limit 20 and offset 0")
			}

			var body struct {
				Reviews []reviewDetailView `json:"reviews"`
				Total   *int               `json:"total"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Reviews == nil {
				t.Fatal("reviews must be an array, not null or missing")
			}
			if body.Total == nil || *body.Total != len(body.Reviews) {
				t.Fatal("total must be present and match the number of reviews")
			}

			if tc.empty {
				if len(body.Reviews) != 0 {
					t.Fatal("expected an empty reviews array")
				}
				return
			}

			if len(body.Reviews) != 1 {
				t.Fatalf("review count = %d, want 1", len(body.Reviews))
			}

			got := body.Reviews[0]
			if got.ID != reviewID ||
				got.Username != "Customer" ||
				got.Order != "Cat artwork" ||
				got.Rating != 5 ||
				got.Comment == nil ||
				*got.Comment != comment ||
				!got.CreatedAt.Equal(createdAt) {
				t.Fatal("response does not contain the expected review details")
			}
		})
	}
}

// TestReviewHandler_GetOrderReview verifies access, response fields, and error mapping.
func TestReviewHandler_GetOrderReview(t *testing.T) {
	orderID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	comment := "Great Work!"
	createdAt := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)

	item := &review.ReviewDetail{
		Review: review.Review{
			ID:         uuid.New(),
			OrderID:    orderID,
			CustomerID: userID,
			ArtistID:   uuid.New(),
			Rating:     5,
			Comment:    &comment,
			CreatedAt:  createdAt,
		},
	}

	cases := []struct {
		name          string
		role          user.Role
		authenticated bool
		usecaseErr    error
		wantStatus    int
	}{
		{
			name:          "customer can read",
			role:          user.RoleCustomer,
			authenticated: true,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "artist can read",
			role:          user.RoleArtist,
			authenticated: true,
			wantStatus:    http.StatusOK,
		},
		{
			name:       "missing token",
			role:       user.RoleCustomer,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:          "not an order participant",
			role:          user.RoleCustomer,
			authenticated: true,
			usecaseErr:    review.ErrNotOrderParticipant,
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "review not found",
			role:          user.RoleCustomer,
			authenticated: true,
			usecaseErr:    review.ErrReviewNotFound,
			wantStatus:    http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := &reviewUsecaseStub{
				orderReview:    item,
				orderReviewErr: tc.usecaseErr,
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			api, server := httpserver.New("", "/api/v1", nil, logger, nil)
			NewReviewHandler(
				usecase,
				reviewAuthStub{role: tc.role},
			).Register(api)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/orders/"+orderID.String()+"/review",
				nil,
			)
			if tc.authenticated {
				req.Header.Set("Authorization", "Bearer test-token")
			}

			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf(
					"status = %d, want %d: %s",
					rec.Code, tc.wantStatus, rec.Body.String(),
				)
			}

			if !tc.authenticated {
				if usecase.getOrderCalls != 0 {
					t.Fatal("an unauthorized request must not reach the usecase")
				}
				return
			}

			if usecase.getOrderCalls != 1 ||
				usecase.getOrderID != orderID ||
				usecase.getOrderUserID != userID {
				t.Fatal("expected one request with the order ID and authenticated user ID")
			}

			if tc.wantStatus != http.StatusOK {
				return
			}

			var got ReviewView
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.ID != item.ID ||
				got.OrderID != orderID ||
				got.CustomerID != item.CustomerID ||
				got.ArtistID != item.ArtistID ||
				got.Rating != 5 ||
				got.Comment == nil ||
				*got.Comment != comment ||
				!got.CreatedAt.Equal(createdAt) {
				t.Fatal("response does not contain the expected review")
			}
		})
	}
}
