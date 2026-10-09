package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artwork"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderUsecase   order.OrderUsecase
	artworkUsecase artwork.Usecase
	authUsecase    auth.AuthUsecase
}

// NewOrderHandler creates a handler for order read endpoints.
func NewOrderHandler(orderUsecase order.OrderUsecase, artworkUsecase artwork.Usecase, authUsecase auth.AuthUsecase) *OrderHandler {
	return &OrderHandler{orderUsecase: orderUsecase, artworkUsecase: artworkUsecase, authUsecase: authUsecase}
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

	huma.Put(api, "/orders/{order_id}/confirm", h.confirmOrder,
		huma.OperationTags("orders"),
		func(o *huma.Operation) {
			o.OperationID = "confirm-order"
			o.Summary = "ConfirmOrder"
			o.Description = "Accept or reject a pending order as the authenticated artist."
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireRole(api, user.RoleArtist),
			)
		},
	)
	huma.Post(api, "/orders", h.createOrder,
		huma.OperationTags("orders"),
		func(o *huma.Operation) {
			o.OperationID = "create-order"
			o.Summary = "CreateOrder"
			o.Description = "Create a new order for the authenticated customer."
			o.DefaultStatus = http.StatusCreated
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireAnyRole(api, user.RoleCustomer),
			)
		},
	)
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

// ConfirmOrderInput contains the order ID and the artist's decision.
type ConfirmOrderInput struct {
	OrderID uuid.UUID `path:"order_id"`
	Body    struct {
		Accept bool `json:"accept"`
	}
}

// ConfirmOrderOutput is returned after the order confirmation is processed.
type ConfirmOrderOutput struct {
	Body struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
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
	Category    string   `json:"category"`
	Styles      []string `json:"styles"`
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

type CreateOrderInput struct {
	Body struct {
		ArtworkID           uuid.UUID `json:"artwork_id"`
		Name                string    `json:"name"`
		CustomerDescription string    `json:"customer_description"`
		DeadlineAt          time.Time `json:"deadline_at"`
	}
}
type CreateOrderOutput struct {
	Body orderSummaryView
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

	category, err := h.artworkUsecase.ListAllCategories(ctx)
	if err != nil {
		return nil, mapAppError(err)
	}

	styles, err := h.artworkUsecase.ListAllStyles(ctx)
	if err != nil {
		return nil, mapAppError(err)
	}

	categoryName := ""
	for _, cat := range category {
		if cat.ID == detail.ArtworkSnapshot.CategoryID {
			categoryName = cat.Label
			break
		}
	}

	styleMap := make(map[uuid.UUID]string, len(styles))
	for _, style := range styles {
		styleMap[style.ID] = style.Label
	}

	styleNames := make([]string, 0, len(detail.ArtworkSnapshot.StyleIDs))
	for _, styleID := range detail.ArtworkSnapshot.StyleIDs {
		if label, ok := styleMap[styleID]; ok {
			styleNames = append(styleNames, label)
		}
	}

	return &GetOrderOutput{
		Body: toOrderDetailView(detail, categoryName, styleNames),
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
func toOrderDetailView(o *order.OrderDetail,
	categoryName string,
	styleNames []string,
) orderDetailView {
	var artworkID *string
	if o.ArtworkID != nil {
		id := o.ArtworkID.String()
		artworkID = &id
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
			Category:    categoryName,
			Styles:      styleNames,
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

// confirmOrder accepts or rejects a pending order on behalf of the
// authenticated artist.
func (h *OrderHandler) confirmOrder(
	ctx context.Context,
	input *ConfirmOrderInput,
) (*ConfirmOrderOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	status, err := h.orderUsecase.ConfirmOrder(
		ctx,
		info.UserID,
		input.OrderID,
		order.ConfirmOrderInput{
			Accept: input.Body.Accept,
		},
	)
	if err != nil {
		return nil, mapAppError(err)
	}

	var message string
	switch input.Body.Accept {
	case true:
		message = "Order accepted"
	case false:
		message = "Order rejected"
	}

	out := &ConfirmOrderOutput{}
	out.Body.Message = message
	out.Body.Status = string(status)

	return out, nil
}

func (h *OrderHandler) createOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	artworkDetail, err := h.artworkUsecase.GetArtwork(ctx, input.Body.ArtworkID)
	if err != nil {
		return nil, mapAppError(err)
	}

	req := order.CreateInput{
		Name:                input.Body.Name,
		ArtworkID:           input.Body.ArtworkID,
		ArtworkDetail:       artworkDetail,
		CustomerDescription: input.Body.CustomerDescription,
		DeadlineAt:          input.Body.DeadlineAt,
	}

	createdOrder, err := h.orderUsecase.CreateOrder(ctx, info.UserID, req)
	if err != nil {
		return nil, mapAppError(err)
	}

	return &CreateOrderOutput{Body: toOrderSummaryView(createdOrder)}, nil

}
