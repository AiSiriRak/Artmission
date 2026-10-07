package rest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

type fakeOrderUsecase struct {
	viewOrdersFunc func(context.Context, order.ListQuery) (order.Page, error)
	getOrderFunc   func(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetail, error)
	gotListQuery   order.ListQuery
	gotParticipant order.Participant
	gotUserID      uuid.UUID
	gotOrderID     uuid.UUID
}

func (f *fakeOrderUsecase) GetOrder(
	ctx context.Context,
	participant order.Participant,
	userID uuid.UUID,
	orderID uuid.UUID,
) (*order.OrderDetail, error) {
	f.gotParticipant = participant
	f.gotUserID = userID
	f.gotOrderID = orderID
	if f.getOrderFunc != nil {
		return f.getOrderFunc(ctx, participant, userID, orderID)
	}
	return nil, nil
}

func (f *fakeOrderUsecase) ViewOrders(ctx context.Context, query order.ListQuery) (order.Page, error) {
	f.gotListQuery = query
	if f.viewOrdersFunc != nil {
		return f.viewOrdersFunc(ctx, query)
	}
	return order.Page{}, nil
}

type orderAuthStub struct {
	role user.Role
}

func (s orderAuthStub) Login(context.Context, string, string) (*auth.AuthResult, error) {
	return nil, nil
}

func (s orderAuthStub) Refresh(context.Context, string) (*auth.AuthResult, error) {
	return nil, nil
}

func (s orderAuthStub) Logout(context.Context, uuid.UUID) error {
	return nil
}

func (s orderAuthStub) Authenticate(context.Context, string) (*auth.TokenClaims, error) {
	return &auth.TokenClaims{
		UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		SessionID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		Role:      s.role,
	}, nil
}

func TestOrderHandlerViewOrdersReturnsFilteredPage(t *testing.T) {
	const userID = "00000000-0000-0000-0000-000000000001"
	deadline := time.Date(2026, time.October, 10, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	previewURL := "https://storage.example.com/deliverable-preview"

	for _, tt := range []struct {
		name        string
		role        user.Role
		participant order.Participant
	}{
		{name: "customer", role: user.RoleCustomer, participant: order.ParticipantCustomer},
		{name: "artist", role: user.RoleArtist, participant: order.ParticipantArtist},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeOrderUsecase{
				viewOrdersFunc: func(context.Context, order.ListQuery) (order.Page, error) {
					return order.Page{
						Orders: []order.Order{{
							ID:                    uuid.MustParse("00000000-0000-0000-0000-000000000010"),
							CustomerID:            uuid.MustParse(userID),
							ArtistID:              uuid.MustParse("00000000-0000-0000-0000-000000000003"),
							Name:                  "Portrait Commission",
							PriceSatangOrder:      150000,
							DeadlineAt:            deadline,
							Status:                order.StatusInProcess,
							DeliverablePreviewURL: &previewURL,
							CreatedAt:             createdAt,
							UpdatedAt:             createdAt,
						}},
						Total: 3,
					}, nil
				},
			}
			handler := newOrderTestHandler(t, tt.role, usecase)
			rec := serveOrderRequest(
				handler,
				http.MethodGet,
				"/orders?status=PENDING&status=SUCCESS&sort=price&order=asc&limit=10&offset=20",
				true,
			)

			if rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got struct {
				Orders []orderSummaryView `json:"orders"`
				Total  int                `json:"total"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Total != 3 || len(got.Orders) != 1 {
				t.Fatalf("response page = total %d, orders %d; want total 3, orders 1", got.Total, len(got.Orders))
			}

			gotOrder := got.Orders[0]
			if gotOrder.ID != "00000000-0000-0000-0000-000000000010" ||
				gotOrder.CustomerID != userID ||
				gotOrder.ArtistID != "00000000-0000-0000-0000-000000000003" ||
				gotOrder.Name != "Portrait Commission" ||
				gotOrder.PriceSatang != 150000 ||
				gotOrder.Status != string(order.StatusInProcess) ||
				!gotOrder.DeadlineAt.Equal(deadline) ||
				gotOrder.DeliverablePreviewURL == nil ||
				*gotOrder.DeliverablePreviewURL != previewURL {
				t.Errorf("response order = %+v", gotOrder)
			}

			wantQuery := order.ListQuery{
				Participant:   tt.participant,
				ParticipantID: uuid.MustParse(userID),
				Statuses:      []order.Status{order.StatusPending, order.StatusSuccess},
				Sort:          order.SortFieldPrice,
				Order:         order.SortOrderAsc,
				Limit:         10,
				Offset:        20,
			}
			if !reflect.DeepEqual(usecase.gotListQuery, wantQuery) {
				t.Errorf("list query = %+v, want %+v", usecase.gotListQuery, wantQuery)
			}
		})
	}
}

func TestOrderHandlerGetOrderReturnsOrderDetail(t *testing.T) {
	const (
		userID     = "00000000-0000-0000-0000-000000000001"
		orderID    = "00000000-0000-0000-0000-000000000010"
		customerID = "00000000-0000-0000-0000-000000000002"
		artistID   = "00000000-0000-0000-0000-000000000003"
		artworkID  = "00000000-0000-0000-0000-000000000004"
	)
	deadline := time.Date(2026, time.October, 10, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, time.October, 9, 12, 0, 0, 0, time.UTC)
	reviewScore := 4.5

	for _, tt := range []struct {
		name        string
		role        user.Role
		participant order.Participant
	}{
		{name: "customer", role: user.RoleCustomer, participant: order.ParticipantCustomer},
		{name: "artist", role: user.RoleArtist, participant: order.ParticipantArtist},
	} {
		t.Run(tt.name, func(t *testing.T) {
			customerUUID := uuid.MustParse(customerID)
			artistUUID := uuid.MustParse(artistID)
			otherParty := order.OrderParty{
				ID:                artistUUID,
				Name:              "Artist",
				Email:             "artist@example.com",
				ArtistReviewScore: &reviewScore,
			}
			if tt.role == user.RoleCustomer {
				customerUUID = uuid.MustParse(userID)
			} else {
				artistUUID = uuid.MustParse(userID)
				otherParty = order.OrderParty{
					ID:    customerUUID,
					Name:  "Customer",
					Email: "customer@example.com",
				}
			}
			artworkUUID := uuid.MustParse(artworkID)

			detail := &order.OrderDetail{
				Order: order.Order{
					ID:                  uuid.MustParse(orderID),
					CustomerID:          customerUUID,
					ArtistID:            artistUUID,
					Name:                "Portrait Commission",
					ArtworkID:           &artworkUUID,
					PriceSatangOrder:    150000,
					CustomerDescription: "Draw a portrait",
					DeadlineAt:          deadline,
					Status:              order.StatusSuccess,
					CompletedAt:         &completedAt,
					CreatedAt:           createdAt,
					UpdatedAt:           completedAt,
				},
				ArtworkSnapshot: order.ArtworkSnapshot{
					ArtworkName: "Portrait Example",
					CategoryID:  uuid.MustParse("00000000-0000-0000-0000-000000000020"),
					StyleIDs: []uuid.UUID{
						uuid.MustParse("00000000-0000-0000-0000-000000000021"),
						uuid.MustParse("00000000-0000-0000-0000-000000000022"),
					},
				},
				OtherParty: otherParty,
			}
			usecase := &fakeOrderUsecase{
				getOrderFunc: func(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetail, error) {
					return detail, nil
				},
			}
			handler := newOrderTestHandler(t, tt.role, usecase)
			rec := serveOrderRequest(handler, http.MethodGet, "/orders/"+orderID, true)

			if rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got orderDetailView
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			categoryID := "00000000-0000-0000-0000-000000000020"
			styleIDs := []string{
				"00000000-0000-0000-0000-000000000021",
				"00000000-0000-0000-0000-000000000022",
			}
			artworkIDString := artworkUUID.String()
			want := orderDetailView{
				ID:                  orderID,
				CustomerID:          customerUUID.String(),
				ArtistID:            artistUUID.String(),
				Name:                "Portrait Commission",
				ArtworkID:           &artworkIDString,
				ArtworkSnapshot:     artworkSnapshotView{ArtworkName: "Portrait Example", CategoryID: categoryID, StyleIDs: styleIDs},
				PriceSatang:         150000,
				CustomerDescription: "Draw a portrait",
				DeadlineAt:          deadline,
				Status:              string(order.StatusSuccess),
				CompletedAt:         &completedAt,
				CreatedAt:           createdAt,
				UpdatedAt:           completedAt,
				OtherParty: orderPartyView{
					ID:                otherParty.ID.String(),
					Name:              otherParty.Name,
					Email:             otherParty.Email,
					ArtistReviewScore: otherParty.ArtistReviewScore,
				},
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response body = %+v, want %+v", got, want)
			}

			if usecase.gotParticipant != tt.participant ||
				usecase.gotUserID != uuid.MustParse(userID) ||
				usecase.gotOrderID != uuid.MustParse(orderID) {
				t.Errorf(
					"get order input = participant %q, userID %s, orderID %s; want %q, %s, %s",
					usecase.gotParticipant,
					usecase.gotUserID,
					usecase.gotOrderID,
					tt.participant,
					userID,
					orderID,
				)
			}
		})
	}
}

func newOrderTestHandler(t *testing.T, role user.Role, usecase order.OrderUsecase) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api, server := httpserver.New("", "/api/v1", nil, logger, nil)
	NewOrderHandler(usecase, orderAuthStub{role: role}).Register(api)
	return server.Handler()
}

func serveOrderRequest(handler http.Handler, method, path string, authenticated bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1"+path, nil)
	if authenticated {
		req.Header.Set("Authorization", "Bearer test-token")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
