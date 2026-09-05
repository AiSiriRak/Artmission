package rest

import (
	"context"
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
	huma.Get(api, "/orders/history", h.viewHiringHistory,
		huma.OperationTags("orders"),
		func(o *huma.Operation) {
			o.OperationID = "view-hiring-history"
			o.Summary = "ViewHiringHistory"
			o.Description = "View the authenticated customer's hiring history"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleCustomer))
		})
}

type ViewHiringHistoryInput struct{}

type orderView struct {
	ID                   string            `json:"id"`
	ArtistID             string            `json:"artist_id"`
	ArtworkID            *string           `json:"artwork_id,omitempty"`
	ArtworkName          string            `json:"artwork_name"`
	ArtworkDescription   string            `json:"artwork_description"`
	PriceSatang          int64             `json:"price_satang"`
	MinimumDeadlineDays  int               `json:"minimum_deadline_days"`
	PreviewImageURL      string            `json:"preview_image_url"`
	CustomerDescription  string            `json:"customer_description"`
	SelectedDeadlineDays int               `json:"selected_deadline_days"`
	DeadlineAt           *time.Time        `json:"deadline_at,omitempty"`
	Status               string            `json:"status"`
	Deliverables         []deliverableView `json:"deliverables"`
	CompletedAt          *time.Time        `json:"completed_at,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type deliverableView struct {
	ID               string    `json:"id"`
	OriginalImageURL string    `json:"original_image_url"`
	PreviewImageURL  string    `json:"preview_image_url"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
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
		out.Body.Orders[i] = toOrderView(&o)
	}
	return out, nil
}

func toOrderView(o *order.Order) orderView {
	var artworkID *string
	if o.ArtworkID != nil {
		id := o.ArtworkID.String()
		artworkID = &id
	}

	deliverables := make([]deliverableView, len(o.Deliverables))
	for i, deliverable := range o.Deliverables {
		deliverables[i] = deliverableView{
			ID:               deliverable.ID.String(),
			OriginalImageURL: deliverable.OriginalImageURL,
			PreviewImageURL:  deliverable.PreviewImageURL,
			SortOrder:        deliverable.SortOrder,
			CreatedAt:        deliverable.CreatedAt,
		}
	}

	return orderView{
		ID:                   o.ID.String(),
		ArtistID:             o.ArtistID.String(),
		ArtworkID:            artworkID,
		ArtworkName:          o.ArtworkNameSnapshot,
		ArtworkDescription:   o.ArtworkDescriptionSnapshot,
		PriceSatang:          o.PriceSatangSnapshot,
		MinimumDeadlineDays:  o.MinimumDeadlineDaysSnapshot,
		PreviewImageURL:      o.PreviewImageURLSnapshot,
		CustomerDescription:  o.CustomerDescription,
		SelectedDeadlineDays: o.SelectedDeadlineDays,
		DeadlineAt:           o.DeadlineAt,
		Status:               string(o.Status),
		Deliverables:         deliverables,
		CompletedAt:          o.CompletedAt,
		CreatedAt:            o.CreatedAt,
		UpdatedAt:            o.UpdatedAt,
	}
}
