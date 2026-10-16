package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

type fakeOrderUsecase struct {
	viewOrdersFunc          func(context.Context, order.ListQuery) (order.Page, error)
	getOrderFunc            func(context.Context, order.Participant, uuid.UUID, uuid.UUID) (*order.OrderDetail, error)
	confirmOrderFunc        func(context.Context, uuid.UUID, uuid.UUID, order.ConfirmOrderInput) (order.Status, error)
	confirmDeliverableFunc  func(context.Context, uuid.UUID, order.ConfirmDeliverableInput) (order.DeliverableDecision, error)
	cancelOrderFunc         func(context.Context, order.Participant, uuid.UUID, order.CancelInput) (*order.OrderCancellation, error)
	cancelExpiredOrdersFunc func(context.Context) ([]uuid.UUID, error)
	createOrderFunc         func(context.Context, uuid.UUID, order.CreateInput) (*order.Order, error)
	createDeliverableFunc   func(context.Context, uuid.UUID, order.CreateDeliverableInput) (*order.Deliverable, error)
	createDelayRequestFunc  func(context.Context, uuid.UUID, order.CreateDelayRequestInput) (*order.OrderDelayRequest, error)
	confirmDelayRequestFunc func(context.Context, uuid.UUID, uuid.UUID, order.ConfirmDelayRequestInput) (*order.ConfirmDelayRequestOutput, error)

	gotListQuery               order.ListQuery
	gotParticipant             order.Participant
	gotUserID                  uuid.UUID
	gotOrderID                 uuid.UUID
	gotConfirmInput            order.ConfirmOrderInput
	gotConfirmDeliverableInput order.ConfirmDeliverableInput
	gotCancelInput             order.CancelInput
	confirmDeliverableCalls    int
	confirmCalls               int
	cancelCalls                int
	gotDeliverableInput        order.CreateDeliverableInput
	gotDeliverableArtistID     uuid.UUID
	gotDelayRequestInput       order.CreateDelayRequestInput
	gotDelayRequestArtist      uuid.UUID
	delayRequestCalls          int
	gotConfirmDelayInput       order.ConfirmDelayRequestInput
	gotConfirmDelayUserID      uuid.UUID
	gotConfirmDelayOrderID     uuid.UUID
	confirmDelayCalls          int
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

func (f *fakeOrderUsecase) ConfirmDeliverable(
	ctx context.Context,
	customerID uuid.UUID,
	input order.ConfirmDeliverableInput,
) (order.DeliverableDecision, error) {
	f.confirmDeliverableCalls++
	f.gotUserID = customerID
	f.gotConfirmDeliverableInput = input
	if f.confirmDeliverableFunc != nil {
		return f.confirmDeliverableFunc(ctx, customerID, input)
	}
	if input.Accept {
		return order.DeliverableDecisionApproved, nil
	}
	return order.DeliverableDecisionRejected, nil
}

func (f *fakeOrderUsecase) CancelOrder(
	ctx context.Context,
	participant order.Participant,
	participantID uuid.UUID,
	input order.CancelInput,
) (*order.OrderCancellation, error) {
	f.cancelCalls++
	f.gotParticipant = participant
	f.gotUserID = participantID
	f.gotOrderID = input.OrderID
	f.gotCancelInput = input
	if f.cancelOrderFunc != nil {
		return f.cancelOrderFunc(ctx, participant, participantID, input)
	}
	return &order.OrderCancellation{
		ID:          uuid.New(),
		OrderID:     input.OrderID,
		Reason:      string(input.Reason),
		Explanation: input.Explanation,
		CreatedAt:   time.Now(),
	}, nil
}

func (f *fakeOrderUsecase) CancelExpiredOrders(ctx context.Context) ([]uuid.UUID, error) {
	f.cancelCalls++
	if f.cancelExpiredOrdersFunc != nil {
		return f.cancelExpiredOrdersFunc(ctx)
	}
	return nil, nil
}

func (f *fakeOrderUsecase) CancelExpiredDelayRequest(context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *fakeOrderUsecase) AutoApproveExpiredDeliverables(context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *fakeOrderUsecase) CreateDeliverable(
	ctx context.Context,
	artistID uuid.UUID,
	input order.CreateDeliverableInput,
) (*order.Deliverable, error) {
	f.gotDeliverableArtistID = artistID
	f.gotDeliverableInput = input
	if f.createDeliverableFunc != nil {
		return f.createDeliverableFunc(ctx, artistID, input)
	}
	return nil, nil
}

func (f *fakeOrderUsecase) CreateDelayRequest(
	ctx context.Context,
	artistID uuid.UUID,
	input order.CreateDelayRequestInput,
) (*order.OrderDelayRequest, error) {
	f.delayRequestCalls++
	f.gotDelayRequestArtist = artistID
	f.gotDelayRequestInput = input
	if f.createDelayRequestFunc != nil {
		return f.createDelayRequestFunc(ctx, artistID, input)
	}
	return nil, nil
}

func (f *fakeOrderUsecase) ConfirmDelayRequest(
	ctx context.Context,
	customerID uuid.UUID,
	orderID uuid.UUID,
	input order.ConfirmDelayRequestInput,
) (*order.ConfirmDelayRequestOutput, error) {
	f.confirmDelayCalls++
	f.gotConfirmDelayUserID = customerID
	f.gotConfirmDelayOrderID = orderID
	f.gotConfirmDelayInput = input
	if f.confirmDelayRequestFunc != nil {
		return f.confirmDelayRequestFunc(ctx, customerID, orderID, input)
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
	originalImageKey := "orders/order-10/v1/original.png"
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
				OtherParty:       otherParty,
				OriginalImageKey: &originalImageKey,
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
				OriginalImageKey: &originalImageKey,
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

func TestOrderHandlerConfirmDeliverableAcceptsOrRejectsDeliverable(t *testing.T) {
	const (
		orderID       = "00000000-0000-0000-0000-000000000010"
		deliverableID = "00000000-0000-0000-0000-000000000020"
		userID        = "00000000-0000-0000-0000-000000000001"
	)

	for _, tt := range []struct {
		name        string
		body        string
		accept      bool
		comment     string
		wantMessage string
		wantStatus  order.DeliverableDecision
	}{
		{
			name:        "accept",
			body:        `{"accept":true}`,
			accept:      true,
			wantMessage: "Deliverable accepted",
			wantStatus:  order.DeliverableDecisionApproved,
		},
		{
			name:        "reject with comment",
			body:        `{"accept":false,"comment":"Please revise the colors."}`,
			comment:     "Please revise the colors.",
			wantMessage: "Deliverable rejected",
			wantStatus:  order.DeliverableDecisionRejected,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeOrderUsecase{}
			handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
			rec := serveOrderRequest(
				handler,
				http.MethodPut,
				"/orders/"+orderID+"/deliverables/"+deliverableID+"/confirm",
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
			if got.Message != tt.wantMessage || got.Status != string(tt.wantStatus) {
				t.Errorf("response = %+v, want message %q and status %q", got, tt.wantMessage, tt.wantStatus)
			}
			if usecase.confirmDeliverableCalls != 1 ||
				usecase.gotUserID != uuid.MustParse(userID) ||
				usecase.gotConfirmDeliverableInput != (order.ConfirmDeliverableInput{
					OrderID:       uuid.MustParse(orderID),
					DeliverableID: uuid.MustParse(deliverableID),
					Accept:        tt.accept,
					Comment:       tt.comment,
				}) {
				t.Errorf(
					"ConfirmDeliverable() received calls=%d customerID=%s input=%+v",
					usecase.confirmDeliverableCalls,
					usecase.gotUserID,
					usecase.gotConfirmDeliverableInput,
				)
			}
		})
	}
}

func TestOrderHandlerConfirmDeliverablePropagatesUsecaseError(t *testing.T) {
	usecase := &fakeOrderUsecase{
		confirmDeliverableFunc: func(context.Context, uuid.UUID, order.ConfirmDeliverableInput) (order.DeliverableDecision, error) {
			return "", order.ErrInvalidDeliverableStatus
		},
	}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/deliverables/00000000-0000-0000-0000-000000000020/confirm",
		true,
		`{"accept":true}`,
	)

	if rec.Code != http.StatusConflict {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	if usecase.confirmDeliverableCalls != 1 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 1", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerConfirmDeliverableMapsUsecaseErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "invalid input", err: apperror.InvalidInput("invalid input", nil), wantStatus: http.StatusBadRequest},
		{name: "unauthorized", err: apperror.Unauthorized("unauthorized"), wantStatus: http.StatusUnauthorized},
		{name: "forbidden", err: apperror.Forbidden("forbidden"), wantStatus: http.StatusForbidden},
		{name: "not found", err: apperror.NotFound("not found"), wantStatus: http.StatusNotFound},
		{name: "conflict", err: apperror.Conflict("conflict"), wantStatus: http.StatusConflict},
		{name: "internal", err: apperror.Internal("internal", nil), wantStatus: http.StatusInternalServerError},
		{name: "unrecognized", err: errors.New("unexpected error"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &fakeOrderUsecase{
				confirmDeliverableFunc: func(context.Context, uuid.UUID, order.ConfirmDeliverableInput) (order.DeliverableDecision, error) {
					return "", tt.err
				},
			}
			handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
			rec := serveOrderRequest(
				handler,
				http.MethodPut,
				"/orders/00000000-0000-0000-0000-000000000010/deliverables/00000000-0000-0000-0000-000000000020/confirm",
				true,
				`{"accept":true}`,
			)

			if rec.Code != tt.wantStatus {
				t.Errorf("response status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if usecase.confirmDeliverableCalls != 1 {
				t.Errorf("ConfirmDeliverable() calls = %d, want 1", usecase.confirmDeliverableCalls)
			}
		})
	}
}

func TestOrderHandlerConfirmDeliverableRejectsMissingAuthenticationContext(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := &OrderHandler{orderUsecase: usecase}
	input := &confirmDeliverableInput{
		OrderID:       uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		DeliverableID: uuid.MustParse("00000000-0000-0000-0000-000000000020"),
	}
	input.Body.Accept = true

	output, err := handler.confirmDeliverable(context.Background(), input)
	if err == nil {
		t.Fatal("confirmDeliverable() error = nil, want missing-authentication error")
	}
	if output != nil {
		t.Errorf("confirmDeliverable() output = %+v, want nil", output)
	}
	if !strings.Contains(err.Error(), "missing authentication") {
		t.Errorf("confirmDeliverable() error = %q, want missing-authentication error", err)
	}
	if usecase.confirmDeliverableCalls != 0 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 0", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerConfirmDeliverableRequiresCustomerRole(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleArtist, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/deliverables/00000000-0000-0000-0000-000000000020/confirm",
		true,
		`{"accept":true}`,
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if usecase.confirmDeliverableCalls != 0 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 0", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerConfirmDeliverableRequiresAuthentication(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/deliverables/00000000-0000-0000-0000-000000000020/confirm",
		false,
		`{"accept":true}`,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if usecase.confirmDeliverableCalls != 0 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 0", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerConfirmDeliverableRejectsInvalidDeliverableID(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/deliverables/not-a-uuid/confirm",
		true,
		`{"accept":true}`,
	)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	if usecase.confirmDeliverableCalls != 0 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 0", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerConfirmDeliverableRejectsInvalidOrderID(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/not-a-uuid/deliverables/00000000-0000-0000-0000-000000000020/confirm",
		true,
		`{"accept":true}`,
	)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	if usecase.confirmDeliverableCalls != 0 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 0", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerConfirmDeliverableRejectsMalformedBody(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/deliverables/00000000-0000-0000-0000-000000000020/confirm",
		true,
		`{"accept":`,
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if usecase.confirmDeliverableCalls != 0 {
		t.Errorf("ConfirmDeliverable() calls = %d, want 0", usecase.confirmDeliverableCalls)
	}
}

func TestOrderHandlerCreateDeliverableUploadsMultipartImage(t *testing.T) {
	const (
		artistID = "00000000-0000-0000-0000-000000000001"
		orderID  = "00000000-0000-0000-0000-000000000010"
	)
	createdAt := time.Date(2026, time.October, 10, 12, 0, 0, 0, time.UTC)
	imageContent := []byte("deliverable image bytes")
	created := &order.Deliverable{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000020"),
		Version:          1,
		Decision:         order.DeliverableDecisionWait,
		OriginalImageKey: "orders/" + orderID + "/original.png",
		PreviewImageKey:  "orders/" + orderID + "/preview.jpg",
		PreviewImageURL:  "https://storage.example.com/preview",
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}
	var gotImage []byte
	usecase := &fakeOrderUsecase{
		createDeliverableFunc: func(_ context.Context, gotArtistID uuid.UUID, input order.CreateDeliverableInput) (*order.Deliverable, error) {
			if gotArtistID != uuid.MustParse(artistID) {
				t.Errorf("artistID = %s, want %s", gotArtistID, artistID)
			}
			if input.OrderID != uuid.MustParse(orderID) {
				t.Errorf("orderID = %s, want %s", input.OrderID, orderID)
			}
			var err error
			gotImage, err = io.ReadAll(input.DeliverableImage)
			if err != nil {
				t.Errorf("read deliverable image: %v", err)
			}
			return created, nil
		},
	}
	handler := newOrderTestHandler(t, user.RoleArtist, usecase)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="deliverable_image"; filename="deliverable.png"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create image form part: %v", err)
	}
	if _, err := part.Write(imageContent); err != nil {
		t.Fatalf("write image form part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/orders/"+orderID+"/deliverables",
		&body,
	)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !bytes.Equal(gotImage, imageContent) {
		t.Errorf("uploaded image = %q, want %q", gotImage, imageContent)
	}
	if usecase.gotDeliverableArtistID != uuid.MustParse(artistID) ||
		usecase.gotDeliverableInput.OrderID != uuid.MustParse(orderID) {
		t.Errorf("usecase received artistID=%s orderID=%s", usecase.gotDeliverableArtistID, usecase.gotDeliverableInput.OrderID)
	}

	var got deliverableView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != created.ID.String() ||
		got.Version != created.Version ||
		got.Decision != created.Decision ||
		got.PreviewImageURL != created.PreviewImageURL ||
		!got.CreatedAt.Equal(createdAt) ||
		!got.UpdatedAt.Equal(createdAt) {
		t.Errorf("response deliverable = %+v, want fields from %+v", got, created)
	}
}

func TestOrderHandlerCreateDeliverableRequiresArtistRole(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/orders/00000000-0000-0000-0000-000000000010/deliverables",
		nil,
	)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if usecase.gotDeliverableArtistID != uuid.Nil {
		t.Error("CreateDeliverable() should not be called for a customer")
	}
}

func TestOrderHandlerCreateDelayRequest(t *testing.T) {
	const (
		artistID  = "00000000-0000-0000-0000-000000000001"
		orderID   = "00000000-0000-0000-0000-000000000010"
		requestID = "00000000-0000-0000-0000-000000000020"
	)
	originalDeadline := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	requestedDeadline := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, time.October, 11, 12, 0, 0, 0, time.UTC)
	created := &order.OrderDelayRequest{
		ID:                 uuid.MustParse(requestID),
		OrderID:            uuid.MustParse(orderID),
		OriginalDeadlineAt: originalDeadline,
		DeadlineRequestAt:  requestedDeadline,
		Reason:             order.DelayReasonNeedMoreTime,
		Explanation:        "I need more time to finish the details.",
		CustomerDecision:   order.DelayDecisionWait,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	}
	usecase := &fakeOrderUsecase{
		createDelayRequestFunc: func(_ context.Context, gotArtistID uuid.UUID, input order.CreateDelayRequestInput) (*order.OrderDelayRequest, error) {
			if gotArtistID != uuid.MustParse(artistID) {
				t.Errorf("artistID = %s, want %s", gotArtistID, artistID)
			}
			if input.OrderID != uuid.MustParse(orderID) ||
				!input.DeadlineRequestAt.Equal(requestedDeadline) ||
				input.Reason != order.DelayReasonNeedMoreTime ||
				input.Explanation != created.Explanation {
				t.Errorf("CreateDelayRequest() input = %+v", input)
			}
			return created, nil
		},
	}
	handler := newOrderTestHandler(t, user.RoleArtist, usecase)
	body := `{
		"deadline_request_at":"2026-10-27T12:00:00Z",
		"reason":"NEED_MORE_TIME",
		"explanation":"I need more time to finish the details."
	}`
	rec := serveOrderRequestWithBody(
		handler,
		http.MethodPost,
		"/orders/"+orderID+"/delay-requests",
		body,
		true,
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var responseFields map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &responseFields); err != nil {
		t.Fatalf("decode response fields: %v", err)
	}
	if _, ok := responseFields["expire_at"]; ok {
		t.Error("response should not include the removed expire_at field")
	}
	if usecase.delayRequestCalls != 1 ||
		usecase.gotDelayRequestArtist != uuid.MustParse(artistID) ||
		usecase.gotDelayRequestInput.OrderID != uuid.MustParse(orderID) {
		t.Errorf(
			"CreateDelayRequest() calls = %d, artistID = %s, orderID = %s",
			usecase.delayRequestCalls,
			usecase.gotDelayRequestArtist,
			usecase.gotDelayRequestInput.OrderID,
		)
	}

	var got delayRequestView
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := toDelayRequestView(created)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("response delay request = %+v, want %+v", got, want)
	}
}

func TestOrderHandlerCreateDelayRequestRequiresArtistRole(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequestWithBody(
		handler,
		http.MethodPost,
		"/orders/00000000-0000-0000-0000-000000000010/delay-requests",
		`{"deadline_request_at":"2026-10-27T12:00:00Z","reason":"OTHER","explanation":"Details"}`,
		true,
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if usecase.delayRequestCalls != 0 {
		t.Errorf("CreateDelayRequest() calls = %d, want 0", usecase.delayRequestCalls)
	}
}

func TestOrderHandlerCreateDelayRequestRequiresAuthentication(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleArtist, usecase)
	rec := serveOrderRequestWithBody(
		handler,
		http.MethodPost,
		"/orders/00000000-0000-0000-0000-000000000010/delay-requests",
		`{"deadline_request_at":"2026-10-27T12:00:00Z","reason":"OTHER","explanation":"Details"}`,
		false,
	)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if usecase.delayRequestCalls != 0 {
		t.Errorf("CreateDelayRequest() calls = %d, want 0", usecase.delayRequestCalls)
	}
}

func TestOrderHandlerCreateDelayRequestRejectsInvalidReason(t *testing.T) {
	usecase := &fakeOrderUsecase{}
	handler := newOrderTestHandler(t, user.RoleArtist, usecase)
	rec := serveOrderRequestWithBody(
		handler,
		http.MethodPost,
		"/orders/00000000-0000-0000-0000-000000000010/delay-requests",
		`{"deadline_request_at":"2026-10-27T12:00:00Z","reason":"INVALID","explanation":"Details"}`,
		true,
	)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusUnprocessableEntity, rec.Body.String())
	}
	if usecase.delayRequestCalls != 0 {
		t.Errorf("CreateDelayRequest() calls = %d, want 0", usecase.delayRequestCalls)
	}
}

func TestOrderHandlerCreateDelayRequestPropagatesUsecaseError(t *testing.T) {
	usecase := &fakeOrderUsecase{
		createDelayRequestFunc: func(context.Context, uuid.UUID, order.CreateDelayRequestInput) (*order.OrderDelayRequest, error) {
			return nil, order.ErrDelayRequestPending
		},
	}
	handler := newOrderTestHandler(t, user.RoleArtist, usecase)
	rec := serveOrderRequestWithBody(
		handler,
		http.MethodPost,
		"/orders/00000000-0000-0000-0000-000000000010/delay-requests",
		`{"deadline_request_at":"2026-10-27T12:00:00Z","reason":"OTHER","explanation":"Details"}`,
		true,
	)

	if rec.Code != http.StatusConflict {
		t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	if usecase.delayRequestCalls != 1 {
		t.Errorf("CreateDelayRequest() calls = %d, want 1", usecase.delayRequestCalls)
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
				`{"reason":"OTHER","explanation":"Circumstances changed"}`,
			)

			if rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var got struct {
				OrderID     uuid.UUID `json:"order_id"`
				Reason      string    `json:"reason"`
				Explanation string    `json:"explanation"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.OrderID != uuid.MustParse(orderID) {
				t.Errorf("response order_id = %s, want %s", got.OrderID, orderID)
			}
			if got.Reason != string(order.CancelReasonOther) {
				t.Errorf("response reason = %q, want %q", got.Reason, order.CancelReasonOther)
			}
			if got.Explanation != "Circumstances changed" {
				t.Errorf("response explanation = %q, want %q", got.Explanation, "Circumstances changed")
			}
			if usecase.cancelCalls != 1 ||
				usecase.gotParticipant != tt.participant ||
				usecase.gotUserID != uuid.MustParse(userID) ||
				usecase.gotOrderID != uuid.MustParse(orderID) ||
				usecase.gotCancelInput != (order.CancelInput{
					OrderID:     uuid.MustParse(orderID),
					Reason:      order.CancelReasonOther,
					Explanation: "Circumstances changed",
				}) {
				t.Errorf(
					"cancel input = calls %d, participant %q, userID %s, input %+v; want 1, %q, %s, expected cancellation input",
					usecase.cancelCalls,
					usecase.gotParticipant,
					usecase.gotUserID,
					usecase.gotCancelInput,
					tt.participant,
					userID,
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
		cancelOrderFunc: func(context.Context, order.Participant, uuid.UUID, order.CancelInput) (*order.OrderCancellation, error) {
			return nil, order.ErrInvalidOrderStatus
		},
	}
	handler := newOrderTestHandler(t, user.RoleCustomer, usecase)
	rec := serveOrderRequest(
		handler,
		http.MethodPut,
		"/orders/00000000-0000-0000-0000-000000000010/cancel",
		true,
		`{"reason":"OTHER","explanation":"Circumstances changed"}`,
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
		`{"reason":"OTHER","explanation":"Circumstances changed"}`,
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
		req.Header.Set("Authorization", "Bearer test-token")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
