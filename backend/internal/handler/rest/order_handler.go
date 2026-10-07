package rest

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderUsecase order.OrderUsecase
	authUsecase  auth.AuthUsecase
}

// NewOrderHandler creates a handler for order read endpoints.
func NewOrderHandler(orderUsecase order.OrderUsecase, authUsecase auth.AuthUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase, authUsecase: authUsecase}
}

// Register adds the authenticated order endpoints to the API.
func (h *OrderHandler) Register(api huma.API) {
	huma.Get(api, "/orders", h.viewOrders,
		huma.OperationTags("orders"),
		func(o *huma.Operation) {
			o.OperationID = "view-orders"
			o.Summary = "ViewOrders"
			o.Description = "List the authenticated customer's or artist's orders, filtered by status, " +
				"sorted by deadline/price/updated_at, and paginated by limit/offset. " +
				"deliverable_preview_url is the artist's most recently submitted deliverable preview, presigned and time-limited, regardless of order status; null if none has been submitted yet."
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireAnyRole(api, user.RoleCustomer, user.RoleArtist))
		})

	huma.Get(api, "/orders/{order_id}", h.getOrder,
		huma.OperationTags("orders"),
		func(o *huma.Operation) {
			o.OperationID = "get-order"
			o.Summary = "GetOrder"
			o.Description = "Get the authenticated customer's or artist's order detail."
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireAnyRole(api, user.RoleCustomer, user.RoleArtist),
			)
		})
}

type orderSummaryView struct {
	ID                    string     `json:"id"`
	CustomerID            string     `json:"customer_id"`
	ArtistID              string     `json:"artist_id"`
	Name                  string     `json:"name"`
	PriceSatang           int64      `json:"price_satang"`
	DeadlineAt            time.Time  `json:"deadline_at"`
	Status                string     `json:"status"`
	DeliverablePreviewURL *string    `json:"deliverable_preview_url" doc:"Presigned URL of the most recently submitted deliverable's preview image, regardless of order status; null if none has been submitted yet."`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
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
		Orders []orderSummaryView `json:"orders"`
		Total  int                `json:"total"`
	}
}

// GetOrderInput contains the order ID requested by the authenticated
// participant.
type GetOrderInput struct {
	OrderID uuid.UUID `path:"order_id"`
}

// GetOrderOutput contains the detailed order information.
type GetOrderOutput struct {
	Body orderDetailView
}

// orderPartyView represents the participant on the opposite side of the
// authenticated user's order.
type orderPartyView struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	ArtistReviewScore *float64 `json:"artist_review_score,omitempty"`
}

// artworkSnapshotView represents the artwork information captured when the
// order was created.
type artworkSnapshotView struct {
	ArtworkName string   `json:"artwork_name"`
	CategoryID  string   `json:"category_id"`
	StyleIDs    []string `json:"style_ids"`
}

// deliverableView is the deliverable data exposed in an order detail response.
type deliverableView struct {
	ID              string                    `json:"id"`
	Version         int                       `json:"version"`
	Decision        order.DeliverableDecision `json:"decision"`
	Comment         *string                   `json:"comment"`
	PreviewImageKey string                    `json:"preview_image_key"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

// orderDetailView represents the complete order detail returned to the
// authenticated participant.
type orderDetailView struct {
	ID                  string              `json:"id"`
	CustomerID          string              `json:"customer_id"`
	ArtistID            string              `json:"artist_id"`
	Name                string              `json:"name"`
	ArtworkID           *string             `json:"artwork_id,omitempty"`
	ArtworkSnapshot     artworkSnapshotView `json:"artwork_snapshot"`
	PriceSatang         int64               `json:"price_satang"`
	CustomerDescription string              `json:"customer_description"`
	DeadlineAt          time.Time           `json:"deadline_at"`
	Status              string              `json:"status"`
	CompletedAt         *time.Time          `json:"completed_at,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	OtherParty          orderPartyView      `json:"other_party"`
	Deliverables        []deliverableView   `json:"deliverables"`
}

// viewOrders lists orders scoped to the authenticated customer or artist.
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
	out.Body.Orders = make([]orderSummaryView, len(page.Orders))
	for i, o := range page.Orders {
		out.Body.Orders[i] = toOrderSummaryView(&o)
	}
	out.Body.Total = page.Total
	return out, nil
}

// getOrder returns the order detail for the authenticated participant.
// The participant scope is derived from the authenticated user's role
// instead of being accepted from the request.
func (h *OrderHandler) getOrder(
	ctx context.Context,
	input *GetOrderInput,
) (*GetOrderOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	participant := participantForRole(info.Role)

	detail, err := h.orderUsecase.GetOrder(
		ctx,
		participant,
		info.UserID,
		input.OrderID,
	)
	if err != nil {
		return nil, mapAppError(err)
	}

	return &GetOrderOutput{
		Body: toOrderDetailView(detail),
	}, nil
}

// participantForRole converts an authenticated account role to its order scope.
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

// toOrderSummaryView converts a domain order to its list response view.
func toOrderSummaryView(o *order.Order) orderSummaryView {
	return orderSummaryView{
		ID:                    o.ID.String(),
		CustomerID:            o.CustomerID.String(),
		ArtistID:              o.ArtistID.String(),
		Name:                  o.Name,
		PriceSatang:           o.PriceSatangOrder,
		DeadlineAt:            o.DeadlineAt,
		Status:                string(o.Status),
		DeliverablePreviewURL: o.DeliverablePreviewURL,
		CompletedAt:           o.CompletedAt,
		CreatedAt:             o.CreatedAt,
		UpdatedAt:             o.UpdatedAt,
	}
}

// toOrderDetailView converts the domain order detail into the HTTP response
// representation. UUIDs are converted to strings to keep the API response
// consistent with the other order views.
func toOrderDetailView(o *order.OrderDetail) orderDetailView {
	var artworkID *string
	if o.ArtworkID != nil {
		id := o.ArtworkID.String()
		artworkID = &id
	}

	styleIDs := make([]string, len(o.ArtworkSnapshot.StyleIDs))
	for i, id := range o.ArtworkSnapshot.StyleIDs {
		styleIDs[i] = id.String()
	}

	deliverables := make([]deliverableView, len(o.Deliverables))
	for i, deliverable := range o.Deliverables {
		deliverables[i] = deliverableView{
			ID:              deliverable.ID.String(),
			Version:         deliverable.Version,
			Decision:        deliverable.Decision,
			Comment:         deliverable.Comment,
			PreviewImageKey: deliverable.PreviewImageKey,
			CreatedAt:       deliverable.CreatedAt,
			UpdatedAt:       deliverable.UpdatedAt,
		}
	}

	return orderDetailView{
		ID:         o.ID.String(),
		CustomerID: o.CustomerID.String(),
		ArtistID:   o.ArtistID.String(),
		Name:       o.Name,
		ArtworkID:  artworkID,
		ArtworkSnapshot: artworkSnapshotView{
			ArtworkName: o.ArtworkSnapshot.ArtworkName,
			CategoryID:  o.ArtworkSnapshot.CategoryID.String(),
			StyleIDs:    styleIDs,
		},
		PriceSatang:         o.PriceSatangOrder,
		CustomerDescription: o.CustomerDescription,
		DeadlineAt:          o.DeadlineAt,
		Status:              string(o.Status),
		CompletedAt:         o.CompletedAt,
		CreatedAt:           o.CreatedAt,
		UpdatedAt:           o.UpdatedAt,
		OtherParty: orderPartyView{
			ID:                o.OtherParty.ID.String(),
			Name:              o.OtherParty.Name,
			Email:             o.OtherParty.Email,
			ArtistReviewScore: o.OtherParty.ArtistReviewScore,
		},
		Deliverables: deliverables,
	}
}
