package order_test

import (
	"context"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/google/uuid"
)

type fakeRepo struct {
	byCustomer map[uuid.UUID][]order.Order
}

func (f *fakeRepo) ListByCustomerID(_ context.Context, customerID uuid.UUID) ([]order.Order, error) {
	return f.byCustomer[customerID], nil
}

var _ order.OrderRepository = (*fakeRepo)(nil)

func TestViewHiringHistory_ReturnsOnlyThatCustomersOrders(t *testing.T) {
	customerID := uuid.New()
	otherCustomerID := uuid.New()

	repo := &fakeRepo{byCustomer: map[uuid.UUID][]order.Order{
		customerID:      {{ID: uuid.New(), CustomerID: customerID, Status: order.StatusPending}},
		otherCustomerID: {{ID: uuid.New(), CustomerID: otherCustomerID, Status: order.StatusSuccess}},
	}}
	usecase := order.NewOrderUsecase(repo)

	got, err := usecase.ViewHiringHistory(context.Background(), customerID)
	if err != nil {
		t.Fatalf("ViewHiringHistory() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].CustomerID != customerID {
		t.Errorf("ViewHiringHistory() = %+v, want exactly the requesting customer's one order", got)
	}
}

func TestViewHiringHistory_EmptyIsNotAnError(t *testing.T) {
	repo := &fakeRepo{byCustomer: map[uuid.UUID][]order.Order{}}
	usecase := order.NewOrderUsecase(repo)

	got, err := usecase.ViewHiringHistory(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ViewHiringHistory() error = %v, want nil for a customer with no orders", err)
	}
	if len(got) != 0 {
		t.Errorf("ViewHiringHistory() = %+v, want empty slice", got)
	}
}
