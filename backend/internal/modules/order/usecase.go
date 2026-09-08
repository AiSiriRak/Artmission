package order

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type orderUsecase struct {
	repo    OrderRepository
	storage ObjectStorage
}

func NewOrderUsecase(repo OrderRepository, storage ObjectStorage) OrderUsecase {
	return &orderUsecase{repo: repo, storage: storage}
}

const orderDeliverableTTL = 15 * time.Minute

func (u *orderUsecase) ViewOrders(ctx context.Context, query ListQuery) (Page, error) {
	normalized, err := normalizeListQuery(query)
	if err != nil {
		return Page{}, err
	}

	page, err := u.repo.ListOrders(ctx, normalized)
	if err != nil {
		return Page{}, err
	}

	for i := range page.Orders {
		key := page.Orders[i].DeliverablePreviewKey
		if key == nil {
			continue
		}
		// TODO: Fix this N+1 query problem here later
		url, err := u.storage.GetPresignedURL(ctx, *key, orderDeliverableTTL)
		if err != nil {
			return Page{}, apperror.Internal("failed to presign deliverable preview image", err)
		}
		page.Orders[i].DeliverablePreviewURL = &url
	}

	return page, nil
}

// normalizeListQuery validates query and defaults every optional field.
func normalizeListQuery(q ListQuery) (ListQuery, error) {
	if !q.Participant.IsValid() {
		return ListQuery{}, apperror.Forbidden("unsupported participant role")
	}
	if q.ParticipantID == uuid.Nil {
		return ListQuery{}, apperror.InvalidInput("missing participant id", nil)
	}

	statuses, err := normalizeStatuses(q.Statuses)
	if err != nil {
		return ListQuery{}, err
	}
	q.Statuses = statuses

	if q.Sort == "" {
		q.Sort = DefaultSort
	}
	if !q.Sort.IsValid() {
		return ListQuery{}, apperror.InvalidInput(fmt.Sprintf("invalid sort field %q", q.Sort), nil)
	}

	if q.Order == "" {
		q.Order = DefaultOrder
	}
	if !q.Order.IsValid() {
		return ListQuery{}, apperror.InvalidInput(fmt.Sprintf("invalid sort order %q", q.Order), nil)
	}

	if q.Limit == 0 {
		q.Limit = DefaultLimit
	}
	if q.Limit < 1 || q.Limit > MaxLimit {
		return ListQuery{}, apperror.InvalidInput(fmt.Sprintf("limit must be between 1 and %d", MaxLimit), nil)
	}

	if q.Offset < 0 {
		return ListQuery{}, apperror.InvalidInput("offset must not be negative", nil)
	}

	return q, nil
}

// normalizeStatuses validates every status, drops duplicates, and sorts
// the result into a deterministic order.
func normalizeStatuses(in []Status) ([]Status, error) {
	if len(in) == 0 {
		return nil, nil
	}

	seen := make(map[Status]bool, len(in))
	out := make([]Status, 0, len(in))
	for _, s := range in {
		if !s.IsValid() {
			return nil, apperror.InvalidInput(fmt.Sprintf("invalid status %q", s), nil)
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}

	slices.Sort(out)
	return out, nil
}
