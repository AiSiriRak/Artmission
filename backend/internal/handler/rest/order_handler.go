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
	huma.Get(api, "/orders", h.viewOrders,
		huma.OperationTags("orders"),
		func(o *huma.Operation) {
			o.OperationID = "view-orders"
			o.Summary = "ViewOrders"
			o.Description = "List the authenticated customer's or artist's orders, filtered by status, " +
				"sorted by deadline/price/updated_at, and paginated by limit/offset"
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireAnyRole(api, user.RoleCustomer, user.RoleArtist))
		})
}

type deliverableView struct {
	ID               string    `json:"id"`
	OriginalImageURL string    `json:"original_image_url"`
	PreviewImageURL  string    `json:"preview_image_url"`
	SortOrder        int       `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
}

type orderView struct {
	ID                   string            `json:"id"`
	CustomerID           string            `json:"customer_id"`
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

type ViewOrdersInput struct {
	Status []string `query:"status,explode" enum:"PENDING,NOT_PAID,IN_PROCESS,SUCCESS,CANCEL" doc:"Filter by one or more order statuses. Repeated values are OR'd. Omit for every status."`
	Sort   string   `query:"sort" enum:"deadline,price,updated_at" default:"updated_at" doc:"Field to sort by."`
	Order  string   `query:"order" enum:"asc,desc" default:"desc" doc:"Sort direction."`
	Limit  int      `query:"limit" minimum:"1" maximum:"100" default:"20" doc:"Maximum number of orders to return."`
	Offset int      `query:"offset" minimum:"0" default:"0" doc:"Number of matching orders to skip before the first returned row."`
}

type ViewOrdersOutput struct {
	Body struct {
		Orders []orderView `json:"orders"`
		Total  int         `json:"total"`
	}
}

func (h *OrderHandler) viewOrders(ctx context.Context, in *ViewOrdersInput) (*ViewOrdersOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	statuses := make([]order.Status, len(in.Status))
	for i, s := range in.Status {
		statuses[i] = order.Status(s)
	}

	query := order.ListQuery{
		Participant:   participantForRole(info.Role),
		ParticipantID: info.UserID,
		Statuses:      statuses,
		Sort:          order.SortField(in.Sort),
		Order:         order.SortOrder(in.Order),
		Limit:         in.Limit,
		Offset:        in.Offset,
	}

	page, err := h.orderUsecase.ViewOrders(ctx, query)
	if err != nil {
		return nil, mapAppError(err)
	}

	out := &ViewOrdersOutput{}
	out.Body.Orders = make([]orderView, len(page.Orders))
	for i, o := range page.Orders {
		out.Body.Orders[i] = toOrderView(&o)
	}
	out.Body.Total = page.Total
	return out, nil
}

func participantForRole(role user.Role) order.Participant {
	switch role {
	case user.RoleCustomer:
		return order.ParticipantCustomer
	case user.RoleArtist:
		return order.ParticipantArtist
	default:
		return ""
	}
}

func toOrderView(o *order.Order) orderView {
	var artworkID *string
	if o.ArtworkID != nil {
		id := o.ArtworkID.String()
		artworkID = &id
	}

	deliverables := make([]deliverableView, len(o.Deliverables))
	for i, d := range o.Deliverables {
		deliverables[i] = deliverableView{
			ID:               d.ID.String(),
			OriginalImageURL: d.OriginalImageURL,
			PreviewImageURL:  d.PreviewImageURL,
			SortOrder:        d.SortOrder,
			CreatedAt:        d.CreatedAt,
		}
	}

	return orderView{
		ID:                   o.ID.String(),
		CustomerID:           o.CustomerID.String(),
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
