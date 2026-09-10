package baserepo

import (
	"context"

	"github.com/uptrace/bun"
)

// PaginationInput is one offset page's caller-supplied inputs.
type PaginationInput struct {
	Limit  int
	Offset int
}

// Page is one fetched, offset-paginated result: Items is at most
// PageRequest.Limit rows starting at PageRequest.Offset, and Total is the
// number of rows build's query matches across every page.
type Page[M any] struct {
	Items []M
	Total int
}

// Paginate applies build's query + pagination input LIMIT/OFFSET, and returns
// one page alongside the total row count under one database call.
func Paginate[M any](
	ctx context.Context,
	exec Executor,
	build func(*bun.SelectQuery) *bun.SelectQuery,
	input PaginationInput,
) (Page[M], error) {
	var models []M
	var total int
	err := exec.Run(ctx, func(idb bun.IDB) error {
		q := build(idb.NewSelect().Model(&models)).Limit(input.Limit).Offset(input.Offset)
		var err error
		total, err = q.ScanAndCount(ctx)
		return err
	})
	if err != nil {
		return Page[M]{}, err
	}
	return Page[M]{Items: models, Total: total}, nil
}
