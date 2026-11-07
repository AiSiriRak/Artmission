package rest

import (
	"context"
	"encoding/json"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/notification"
	"github.com/google/uuid"
	"testing"
	"time"
)

type notificationTestUsecase struct {
	notification.NotificationUsecase
	recipient uuid.UUID
	markedID  uuid.UUID
	markError error
	page      notification.Page
}

func (u *notificationTestUsecase) MarkRead(_ context.Context, id, recipient uuid.UUID) error {
	u.markedID, u.recipient = id, recipient
	return u.markError
}

func TestMarkNotificationReadReturnsJSON(t *testing.T) {
	id, owner := uuid.New(), uuid.New()
	uc := &notificationTestUsecase{}
	out, err := NewNotificationHandler(uc, nil).markRead(withAuthInfo(context.Background(), AuthInfo{UserID: owner}), &markNotificationReadInput{ID: id})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Body struct {
			ID      uuid.UUID `json:"id"`
			IsRead  bool      `json:"is_read"`
			Message string    `json:"message"`
		}
	}
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	if result.Body.ID != id || !result.Body.IsRead || result.Body.Message != "Notification marked as read." {
		t.Fatalf("missing read confirmation: %s", encoded)
	}
	if uc.markedID != id || uc.recipient != owner {
		t.Fatal("must mark the requested notification for the authenticated recipient")
	}
}

func TestMarkNotificationReadDoesNotReturnSuccessForMissingNotification(t *testing.T) {
	uc := &notificationTestUsecase{markError: notification.ErrNotFound}
	out, err := NewNotificationHandler(uc, nil).markRead(withAuthInfo(context.Background(), AuthInfo{UserID: uuid.New()}), &markNotificationReadInput{ID: uuid.New()})
	if err == nil || out != nil {
		t.Fatal("missing notification must return an error without success JSON")
	}
}

func (u *notificationTestUsecase) ListMine(_ context.Context, id uuid.UUID, _ notification.ListQuery) (notification.Page, error) {
	u.recipient = id
	return u.page, nil
}

func TestNotificationHandlerReturnsCardMetadata(t *testing.T) {
	orderID, owner := uuid.New(), uuid.New()
	deadline := time.Now().UTC()
	uc := &notificationTestUsecase{page: notification.Page{Items: []notification.Notification{{
		ID: uuid.New(), Type: notification.TypeOrderAccepted, OrderID: &orderID,
		DeadlineAt: &deadline,
	}}, Total: 1}}
	out, err := NewNotificationHandler(uc, nil).listMine(withAuthInfo(context.Background(), AuthInfo{UserID: owner}), &listNotificationsInput{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	n := out.Body.Items[0]
	if n.Type != notification.TypeOrderAccepted || n.OrderID == nil || *n.OrderID != orderID ||
		n.DeadlineAt == nil || !n.DeadlineAt.Equal(deadline) {
		t.Fatalf("response lost card data: %+v", n)
	}
	encoded, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["type"] != "ORDER_ACCEPTED" {
		t.Fatalf("type must be specific event: %s", encoded)
	}
	for _, removed := range []string{"event_type", "order_status"} {
		if _, ok := fields[removed]; ok {
			t.Fatalf("obsolete field %s remains", removed)
		}
	}
}
func TestNotificationHandlerUsesAuthenticatedRecipient(t *testing.T) {
	uc := &notificationTestUsecase{}
	h := NewNotificationHandler(uc, nil)
	owner := uuid.New()
	out, err := h.listMine(withAuthInfo(context.Background(), AuthInfo{UserID: owner}), &listNotificationsInput{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if uc.recipient != owner {
		t.Fatal("recipient must come from authentication")
	}
	if out.Body.Items == nil {
		t.Fatal("empty feed must serialize as an array")
	}
}
func TestNotificationHandlersRequireAuthentication(t *testing.T) {
	h := NewNotificationHandler(nil, nil)
	ctx := context.Background()
	if _, err := h.listMine(ctx, &listNotificationsInput{Limit: 20}); err == nil {
		t.Fatal("list accepted missing authentication")
	}
	if _, err := h.markRead(ctx, &markNotificationReadInput{ID: uuid.New()}); err == nil {
		t.Fatal("mark read accepted missing authentication")
	}
	if _, err := h.countUnread(ctx, &struct{}{}); err == nil {
		t.Fatal("count accepted missing authentication")
	}
}
