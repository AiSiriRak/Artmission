package order

import (
	"context"

	"github.com/google/uuid"
)

type OrderUsecase interface {
	// ViewHiringHistory lists every order belonging to customerID, most
	// recent first. An empty slice (no orders yet) is a valid result, not
	// an error.
	ViewHiringHistory(ctx context.Context, customerID uuid.UUID) ([]Order, error)
}

type orderUsecase struct {
	repo OrderRepository
}

func NewOrderUsecase(repo OrderRepository) OrderUsecase {
	return &orderUsecase{repo: repo}
}

func (u *orderUsecase) ViewHiringHistory(ctx context.Context, customerID uuid.UUID) ([]Order, error) {
	return u.repo.ListByCustomerID(ctx, customerID)
}
