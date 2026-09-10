package order

import "context"

type OrderUsecase interface {
	// ViewOrders validates/defaults query, then delegates to OrderRepository.
	ViewOrders(ctx context.Context, query ListQuery) (Page, error)
}

type OrderRepository interface {
	// ListOrders returns one page of orders for an already-validated query.
	ListOrders(ctx context.Context, query ListQuery) (Page, error)
}
