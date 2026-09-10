package order_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/order"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

// fakeRepo records the last ListQuery it received (already normalized by
// the usecase) and returns a fixed page or error, so tests can assert on
// exactly what the usecase delegates.
type fakeRepo struct {
	gotQuery order.ListQuery
	page     order.Page
	err      error
}

func (f *fakeRepo) ListOrders(_ context.Context, query order.ListQuery) (order.Page, error) {
	f.gotQuery = query
	return f.page, f.err
}

var _ order.OrderRepository = (*fakeRepo)(nil)

func validQuery() order.ListQuery {
	return order.ListQuery{
		Participant:   order.ParticipantCustomer,
		ParticipantID: uuid.New(),
	}
}

func wantForbidden(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.Error, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeForbidden {
		t.Errorf("Code = %v, want %v", appErr.Code, apperror.CodeForbidden)
	}
}

func wantInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.Error, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeInvalidInput {
		t.Errorf("Code = %v, want %v", appErr.Code, apperror.CodeInvalidInput)
	}
}

func TestViewOrders_DefaultsSortOrderAndLimit(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	if _, err := usecase.ViewOrders(context.Background(), validQuery()); err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}

	if repo.gotQuery.Sort != order.DefaultSort {
		t.Errorf("Sort = %q, want default %q", repo.gotQuery.Sort, order.DefaultSort)
	}
	if repo.gotQuery.Order != order.DefaultOrder {
		t.Errorf("Order = %q, want default %q", repo.gotQuery.Order, order.DefaultOrder)
	}
	if repo.gotQuery.Limit != order.DefaultLimit {
		t.Errorf("Limit = %d, want default %d", repo.gotQuery.Limit, order.DefaultLimit)
	}
}

func TestViewOrders_ArtistParticipant_PassesThrough(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Participant = order.ParticipantArtist
	if _, err := usecase.ViewOrders(context.Background(), q); err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}

	if repo.gotQuery.Participant != order.ParticipantArtist {
		t.Errorf("Participant = %q, want %q", repo.gotQuery.Participant, order.ParticipantArtist)
	}
	if repo.gotQuery.ParticipantID != q.ParticipantID {
		t.Errorf("ParticipantID = %v, want %v", repo.gotQuery.ParticipantID, q.ParticipantID)
	}
}

func TestViewOrders_RejectsUnsupportedParticipant(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Participant = "admin"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantForbidden(t, err)
}

func TestViewOrders_RejectsMissingParticipantID(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.ParticipantID = uuid.Nil
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidSort(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Sort = "created_at"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidOrder(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Order = "sideways"
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsLimitOutOfRange(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	for _, limit := range []int{-1, order.MaxLimit + 1} {
		q := validQuery()
		q.Limit = limit
		_, err := usecase.ViewOrders(context.Background(), q)
		wantInvalidInput(t, err)
	}
}

func TestViewOrders_RejectsNegativeOffset(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Offset = -1
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_RejectsInvalidStatus(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Statuses = []order.Status{"NOT_A_STATUS"}
	_, err := usecase.ViewOrders(context.Background(), q)
	wantInvalidInput(t, err)
}

func TestViewOrders_NormalizesStatuses_DedupesAndSorts(t *testing.T) {
	repo := &fakeRepo{}
	usecase := order.NewOrderUsecase(repo)

	q := validQuery()
	q.Statuses = []order.Status{order.StatusSuccess, order.StatusPending, order.StatusPending}
	if _, err := usecase.ViewOrders(context.Background(), q); err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil", err)
	}

	want := []order.Status{order.StatusPending, order.StatusSuccess}
	got := repo.gotQuery.Statuses
	if len(got) != len(want) {
		t.Fatalf("Statuses = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Statuses[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestViewOrders_EmptyPageIsNotAnError(t *testing.T) {
	repo := &fakeRepo{page: order.Page{Orders: []order.Order{}}}
	usecase := order.NewOrderUsecase(repo)

	page, err := usecase.ViewOrders(context.Background(), validQuery())
	if err != nil {
		t.Fatalf("ViewOrders() error = %v, want nil for an empty page", err)
	}
	if len(page.Orders) != 0 {
		t.Errorf("Orders = %+v, want empty", page.Orders)
	}
}

func TestViewOrders_PropagatesRepositoryError(t *testing.T) {
	wantErr := apperror.Internal("boom", nil)
	repo := &fakeRepo{err: wantErr}
	usecase := order.NewOrderUsecase(repo)

	_, err := usecase.ViewOrders(context.Background(), validQuery())
	if !errors.Is(err, wantErr) {
		t.Errorf("ViewOrders() error = %v, want %v", err, wantErr)
	}
}
