package order

import (
	"context"

	"github.com/google/uuid"
)

type orderUsecase struct {
	repo OrderRepository
}

func NewOrderUsecase(repo OrderRepository) OrderUsecase {
	return &orderUsecase{repo: repo}
}

func (u *orderUsecase) ViewHiringHistory(ctx context.Context, customerID uuid.UUID) ([]Order, error) {
	return u.repo.ListByCustomerID(ctx, customerID)
}
