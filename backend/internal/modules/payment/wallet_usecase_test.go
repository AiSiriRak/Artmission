package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type walletRepoStub struct {
	WalletRepository
	query  TransactionQuery
	userID uuid.UUID
}

func (s *walletRepoStub) ListWalletTransactions(_ context.Context, id uuid.UUID, q TransactionQuery) (TransactionPage, error) {
	s.query, s.userID = q, id
	return TransactionPage{}, nil
}

func TestWalletHistoryValidation(t *testing.T) {
	repo := &walletRepoStub{}
	u := NewWalletUsecase(repo)
	id := uuid.New()
	if _, err := u.ListWalletTransactions(context.Background(), id, TransactionQuery{}); err != nil {
		t.Fatal(err)
	}
	if repo.query.Limit != 20 || repo.userID != id {
		t.Fatal("missing default or authenticated scope")
	}
	for _, q := range []TransactionQuery{{Limit: -1}, {Limit: 101}, {Offset: -1}} {
		if _, err := u.ListWalletTransactions(context.Background(), id, q); err == nil {
			t.Fatalf("accepted %+v", q)
		}
	}
	if _, err := u.GetWallet(context.Background(), uuid.Nil); err == nil {
		t.Fatal("accepted missing user")
	}
	if _, err := u.ListWalletTransactions(context.Background(), uuid.Nil, TransactionQuery{}); err == nil {
		t.Fatal("accepted missing history owner")
	}
}
