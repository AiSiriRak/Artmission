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
	page      notification.Page
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
