package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
)

type OrderHandler struct {
	orderUsecase order.OrderUsecase
	authUsecase  auth.AuthUsecase
}

func NewOrderHandler(orderUsecase order.OrderUsecase, authUsecase auth.AuthUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase, authUsecase: authUsecase}
}

func (h *OrderHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "view-hiring-history",
		Method:      http.MethodGet,
		Path:        "/orders/history",
		Summary:     "View the authenticated customer's hiring history",
		Middlewares: huma.Middlewares{
			requireAuth(api, h.authUsecase),
			requireRole(api, user.RoleCustomer),
		},
	}, h.viewHiringHistory)
}

type ViewHiringHistoryInput struct{}

type orderView struct {
	ID          string     `json:"id"`
	ArtistID    string     `json:"artist_id"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Style       string     `json:"style"`
	Price       *float64   `json:"price,omitempty"`
	Status      string     `json:"status"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ViewHiringHistoryOutput struct {
	Body struct {
		Orders []orderView `json:"orders"`
	}
}

func (h *OrderHandler) viewHiringHistory(ctx context.Context, _ *ViewHiringHistoryInput) (*ViewHiringHistoryOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	orders, err := h.orderUsecase.ViewHiringHistory(ctx, info.UserID)
	if err != nil {
		return nil, mapAppError(err)
	}

	out := &ViewHiringHistoryOutput{}
	out.Body.Orders = make([]orderView, len(orders))
	for i, o := range orders {
		out.Body.Orders[i] = orderView{
			ID:          o.ID.String(),
			ArtistID:    o.ArtistID.String(),
			Description: o.Description,
			Category:    o.Category,
			Style:       o.Style,
			Price:       o.Price,
			Status:      string(o.Status),
			Deadline:    o.Deadline,
			CompletedAt: o.CompletedAt,
			CreatedAt:   o.CreatedAt,
		}
	}
	return out, nil
}
