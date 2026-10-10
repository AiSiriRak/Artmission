package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

type fakeOrderUsecase struct {
	viewOrdersFunc          func(context.Context, order.ListQuery) (order.Page, error)
	getOrderFunc            func(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetail, error)
	confirmOrderFunc        func(context.Context, uuid.UUID, uuid.UUID, order.ConfirmOrderInput) (order.Status, error)
	cancelOrderFunc         func(context.Context, order.Participant, uuid.UUID, uuid.UUID) error
	cancelExpiredOrdersFunc func(context.Context) ([]uuid.UUID, error)
	createOrderFunc         func(ctx context.Context, customerID uuid.UUID, input order.CreateInput) (*order.Order, error)
	gotListQuery            order.ListQuery
	gotParticipant          order.Participant
	gotUserID               uuid.UUID
	gotOrderID              uuid.UUID
	gotConfirmInput         order.ConfirmOrderInput
	confirmCalls            int
	cancelCalls             int
}

type fakeArtworkUsecase struct {
	artwork.Usecase
	getArtworkFunc func(ctx context.Context, artworkID uuid.UUID) (*artwork.Artwork, error)
	categories     []artwork.Category
	styles         []artwork.Style
}

func (f *fakeArtworkUsecase) GetArtwork(ctx context.Context, artworkID uuid.UUID) (*artwork.Artwork, error) {
	if f.getArtworkFunc != nil {
		return f.getArtworkFunc(ctx, artworkID)
	}
	return &artwork.Artwork{
		ID:          artworkID,
		Name:        "Test Artwork",
		PriceSatang: 10000,
	}, nil
}

func (f *fakeArtworkUsecase) ListAllCategories(context.Context) ([]artwork.Category, error) {
	return f.categories, nil
}

func (f *fakeArtworkUsecase) ListAllStyles(context.Context) ([]artwork.Style, error) {
	return f.styles, nil
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

func (f *fakeOrderUsecase) CreateOrder(ctx context.Context, customerID uuid.UUID, input order.CreateInput) (*order.Order, error) {
	if f.createOrderFunc != nil {
		return f.createOrderFunc(ctx, customerID, input)
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

func (f *fakeOrderUsecase) ConfirmOrder(
	ctx context.Context,
	artistID uuid.UUID,
	orderID uuid.UUID,
	input order.ConfirmOrderInput,
) (order.Status, error) {
	f.confirmCalls++
	f.gotUserID = artistID
	f.gotOrderID = orderID
	f.gotConfirmInput = input
	if f.confirmOrderFunc != nil {
		return f.confirmOrderFunc(ctx, artistID, orderID, input)
	}
	if input.Accept {
		return order.StatusNotPaid, nil
	}
	return order.StatusCancel, nil
}

func (f *fakeOrderUsecase) CancelOrder(
	ctx context.Context,
	participant order.Participant,
	participantID uuid.UUID,
	orderID uuid.UUID,
) error {
	f.cancelCalls++
	f.gotParticipant = participant
	f.gotUserID = participantID
	f.gotOrderID = orderID
	if f.cancelOrderFunc != nil {
		return f.cancelOrderFunc(ctx, participant, participantID, orderID)
	}
	return nil
}

func (f *fakeOrderUsecase) CancelExpiredOrders(ctx context.Context) ([]uuid.UUID, error) {
	f.cancelCalls++
	if f.cancelExpiredOrdersFunc != nil {
		return f.cancelExpiredOrdersFunc(ctx)
	}
	return nil, nil
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
	deliverableUpdatedAt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	deliverableComment := "Please revise this version."
	deliverablePreviewURL := "https://storage.example.com/orders/order-10/v1/preview.png"
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
				Deliverables: []order.Deliverable{{
					ID:              uuid.MustParse("00000000-0000-0000-0000-000000000030"),
					Version:         1,
					Decision:        order.DeliverableDecisionWait,
					Comment:         &deliverableComment,
					PreviewImageURL: deliverablePreviewURL,
					CreatedAt:       createdAt,
					UpdatedAt:       deliverableUpdatedAt,
				}},
				OtherParty: otherParty,
			}
			usecase := &fakeOrderUsecase{
				getOrderFunc: func(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetail, error) {
					return detail, nil
				},
			}
			artworkUC := &fakeArtworkUsecase{
				categories: []artwork.Category{{
					ID:    uuid.MustParse("00000000-0000-0000-0000-000000000020"),
					Label: "Portrait",
				}},
				styles: []artwork.Style{
					{ID: uuid.MustParse("00000000-0000-0000-0000-000000000021"), Label: "Watercolor"},
					{ID: uuid.MustParse("00000000-0000-0000-0000-000000000022"), Label: "Digital"},
				},
			}
			handler := newOrderTestHandlerWithArtwork(t, tt.role, usecase, artworkUC)
			rec := serveOrderRequest(handler, http.MethodGet, "/orders/"+orderID, true)

			if rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got orderDetailView
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			styleNames := []string{
				"Watercolor",
				"Digital",
			}
			artworkIDString := artworkUUID.String()
			want := orderDetailView{
				ID:                  orderID,
				CustomerID:          customerUUID.String(),
				ArtistID:            artistUUID.String(),
				Name:                "Portrait Commission",
				ArtworkID:           &artworkIDString,
				ArtworkSnapshot:     artworkSnapshotView{ArtworkName: "Portrait Example", Category: "Portrait", Styles: styleNames},
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
				Deliverables: []deliverableView{{
					ID:              "00000000-0000-0000-0000-000000000030",
					Version:         1,
					Decision:        order.DeliverableDecisionWait,
					Comment:         &deliverableComment,
					PreviewImageURL: deliverablePreviewURL,
					CreatedAt:       createdAt,
					UpdatedAt:       deliverableUpdatedAt,
				}},
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

func TestOrderHandlerConfirmOrderAcceptsOrRejectsOrder(t *testing.T) {
	const (
		orderID = "00000000-0000-0000-0000-000000000010"
		userID  = "00000000-0000-0000-0000-000000000001"
	)

	for _, tt := range []struct {
		name        string
		body        string
		accept      bool
		wantMessage string
		wantStatus  order.Status
	}{
		{name: "accept", body: `{"accept":true}`, accept: true, wantMessage: "Order accepted", wantStatus: order.StatusNotPaid},
		{name: "reject", body: `{"accept":false}`, accept: false, wantMessage: "Order rejected", wantStatus: order.StatusCancel},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeOrderUsecase{}
			handler := newOrderTestHandler(t, user.RoleArtist, usecase)
			rec := serveOrderRequest(
				handler,
				http.MethodPut,
				"/orders/"+orderID+"/confirm",
				true,
				tt.body,
			)

			if rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got struct {
				Message string `json:"message"`
				Status  string `json:"status"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("response message = %q, want %q", got.Message, tt.wantMessage)
			}
			if got.Status != string(tt.wantStatus) {
				t.Errorf("response status = %q, want %q", got.Status, tt.wantStatus)
			}
			if usecase.confirmCalls != 1 ||
				usecase.gotUserID != uuid.MustParse(userID) ||
				usecase.gotOrderID != uuid.MustParse(orderID) ||
				usecase.gotConfirmInput.Accept != tt.accept {
				t.Errorf(
					"confirm input = calls %d, artistID %s, orderID %s, accept %t",
					usecase.confirmCalls,
					usecase.gotUserID,
					usecase.gotOrderID,
					usecase.gotConfirmInput.Accept,
				)
			}
		})
	}
}

func TestOrderHandlerCreateOrder(t *testing.T) {
	const (
		customerID = "00000000-0000-0000-0000-000000000001"
		artworkID  = "00000000-0000-0000-0000-000000000004"
		orderID    = "00000000-0000-0000-0000-000000000010"
	)
	deadline := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success creating an order as customer", func(t *testing.T) {
		artworkUUID := uuid.MustParse(artworkID)
		createdOrder := &order.Order{
			ID:                  uuid.MustParse(orderID),
			CustomerID:          uuid.MustParse(customerID),
			ArtistID:            uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			ArtworkID:           &artworkUUID,
			Name:                "Portrait Commission",
			PriceSatangOrder:    150000,
			CustomerDescription: "Please draw me with blue background",
			DeadlineAt:          deadline,
			Status:              order.StatusPending,
			CreatedAt:           createdAt,
			UpdatedAt:           createdAt,
		}

		usecase := &fakeOrderUsecase{
			createOrderFunc: func(ctx context.Context, custID uuid.UUID, input order.CreateInput) (*order.Order, error) {
				if custID != uuid.MustParse(customerID) {
					t.Errorf("customerID = %s, want %s", custID, customerID)
				}
				if input.ArtworkID != artworkUUID {
					t.Errorf("artworkID = %s, want %s", input.ArtworkID, artworkUUID)
				}
				return createdOrder, nil
			},
		}

		handler := newOrderTestHandler(t, user.RoleCustomer, usecase)

		body := map[string]any{
			"artwork_id":           artworkID,
			"name":                 "Portrait Commission",
			"customer_description": "Please draw me with blue background",
			"deadline_at":          deadline.Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var got orderDetailView
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		if got.ID != orderID || got.Status != string(order.StatusPending) {
			t.Errorf("got order ID = %s, status = %s; want ID = %s, status = PENDING", got.ID, got.Status, orderID)
		}
	})

	t.Run("rejected when user is an artist", func(t *testing.T) {
		usecase := &fakeOrderUsecase{}
		handler := newOrderTestHandler(t, user.RoleArtist, usecase)

		body := map[string]any{
			"ArtworkID":           artworkID,
			"Name":                "Portrait Commission",
			"CustomerDescription": "Test",
			"DeadlineAt":          deadline.Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("response status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("unauthenticated request rejected", func(t *testing.T) {
		usecase := &fakeOrderUsecase{}
		handler := newOrderTestHandler(t, user.RoleCustomer, usecase)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("response status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func TestOrderHandlerConfirmOrderRequiresArtistRole(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/confirm",
		true,
		`{"accept":true}`,
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if usecase.confirmCalls != 0 {
		t.Errorf("ConfirmOrder() calls = %d, want 0", usecase.confirmCalls)
	}
}

func TestOrderHandlerCancelOrder(t *testing.T) {
	const (
		orderID = "00000000-0000-0000-0000-000000000010"
		userID  = "00000000-0000-0000-0000-000000000001"
	)

	for _, tt := range []struct {
		name        string
		role        user.Role
		participant order.Participant
	}{
		{name: "customer cancels own order", role: user.RoleCustomer, participant: order.ParticipantCustomer},
		{name: "artist cancels own order", role: user.RoleArtist, participant: order.ParticipantArtist},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeOrderUsecase{}
			handler := newOrderTestHandler(t, tt.role, usecase)
			rec := serveOrderRequest(
				handler,
				http.MethodPut,
				"/orders/"+orderID+"/cancel",
				true,
			)

			if rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got struct {
				Message string `json:"message"`
				Status  string `json:"status"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Message != "Order cancelled successfully" {
				t.Errorf("response message = %q, want %q", got.Message, "Order cancelled successfully")
			}
			if got.Status != string(order.StatusCancel) {
				t.Errorf("response status = %q, want %q", got.Status, order.StatusCancel)
			}
			if usecase.cancelCalls != 1 ||
				usecase.gotParticipant != tt.participant ||
				usecase.gotUserID != uuid.MustParse(userID) ||
				usecase.gotOrderID != uuid.MustParse(orderID) {
				t.Errorf(
					"cancel input = calls %d, participant %q, userID %s, orderID %s; want 1, %q, %s, %s",
					usecase.cancelCalls,
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

func TestOrderHandlerCancelOrderRequiresAuthentication(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/cancel",
		false,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if usecase.cancelCalls != 0 {
		t.Errorf("CancelOrder() calls = %d, want 0", usecase.cancelCalls)
	}
}

func TestOrderHandlerCancelOrderRequiresCustomerOrArtistRole(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleAdmin, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/cancel",
		true,
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if usecase.cancelCalls != 0 {
		t.Errorf("CancelOrder() calls = %d, want 0", usecase.cancelCalls)
	}
}

func TestOrderHandlerCancelOrderPropagatesUsecaseError(t *testing.T) {
	usecase := &fakeOrderUsecase{
		cancelOrderFunc: func(context.Context, order.Participant, uuid.UUID, uuid.UUID) error {
			return order.ErrInvalidOrderStatus
		},
	}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/cancel",
		true,
	)

	if rec.Code != http.StatusConflict {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	if usecase.cancelCalls != 1 {
		t.Errorf("CancelOrder() calls = %d, want 1", usecase.cancelCalls)
	}
}

func TestOrderHandlerCancelOrderRejectsInvalidOrderID(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/not-a-uuid/cancel",
		true,
	)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	if usecase.cancelCalls != 0 {
		t.Errorf("CancelOrder() calls = %d, want 0", usecase.cancelCalls)
	}
}

func newOrderTestHandler(
	t *testing.T,
	role user.Role,
	usecase order.OrderUsecase,
) http.Handler {
	return newOrderTestHandlerWithArtwork(t, role, usecase, &fakeArtworkUsecase{})
}

func newOrderTestHandlerWithArtwork(
	t *testing.T,
	role user.Role,
	usecase order.OrderUsecase,
	artworkUsecase artwork.Usecase,
) http.Handler {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api, server := httpserver.New("", "/api/v1", nil, logger, nil)
	NewOrderHandler(usecase, artworkUsecase, orderAuthStub{role: role}).Register(api)
	return server.Handler()
}

func serveOrderRequest(
	handler http.Handler,
	method,
	path string,
	authenticated bool,
	body ...string,
) *httptest.ResponseRecorder {
	var requestBody io.Reader
	if len(body) > 0 {
		requestBody = strings.NewReader(body[0])
	}
	req := httptest.NewRequest(method, "/api/v1"+path, requestBody)
	req.Header.Set("Content-Type", "application/json")
	if authenticated {
		req.Header.Set("Authorization", "Bearer test-token")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func serveOrderRequestWithBody(
	handler http.Handler,
	method string,
	path string,
	body string,
	authenticated bool,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authenticated {
		req.Header.Set("Authorization", "******")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
