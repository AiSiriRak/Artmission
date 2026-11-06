package rest

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/google/uuid"
)

type walletStub struct {
	payment.WalletUsecase
	userID uuid.UUID
	query  payment.TransactionQuery
	calls  int
}

func (s *walletStub) GetWallet(_ context.Context, id uuid.UUID) (*payment.Wallet, error) {
	s.userID = id
	s.calls++
	return &payment.Wallet{UserID: id}, nil
}
func (s *walletStub) ListWalletTransactions(_ context.Context, id uuid.UUID, query payment.TransactionQuery) (payment.TransactionPage, error) {
	s.userID, s.query = id, query
	s.calls++
	return payment.TransactionPage{Transactions: []payment.WalletTransaction{}, Total: 0}, nil
}

func TestWalletReadAPI(t *testing.T) {
	for _, role := range []user.Role{user.RoleCustomer, user.RoleArtist} {
		t.Run(string(role), func(t *testing.T) {
			stub := &walletStub{}
			api, server := httpserver.New("", "/api/v1", nil, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
			NewWalletHandler(stub, orderAuthStub{role: role}).Register(api)
			rec := serveOrderRequest(server.Handler(), http.MethodGet, "/wallet/me?user_id="+uuid.NewString(), true)
			if rec.Code != 200 {
				t.Fatalf("wallet %d: %s", rec.Code, rec.Body.String())
			}
			var wallet struct {
				UserID    uuid.UUID  `json:"user_id"`
				Currency  string     `json:"currency"`
				Balance   int64      `json:"balance_satang"`
				UpdatedAt *time.Time `json:"updated_at"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &wallet); err != nil {
				t.Fatal(err)
			}
			if wallet.UserID != uuid.MustParse("00000000-0000-0000-0000-000000000001") || wallet.Currency != "THB" || wallet.Balance != 0 || wallet.UpdatedAt != nil {
				t.Fatalf("wallet response %+v", wallet)
			}
			rec = serveOrderRequest(server.Handler(), http.MethodGet, "/transaction/me?limit=3&offset=2&user_id="+uuid.NewString(), true)
			if rec.Code != 200 {
				t.Fatalf("history %d: %s", rec.Code, rec.Body.String())
			}
			var body struct {
				Transactions         []json.RawMessage `json:"transactions"`
				Total, Limit, Offset int
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Transactions == nil || body.Total != 0 || body.Limit != 3 || body.Offset != 2 || stub.query != (payment.TransactionQuery{Limit: 3, Offset: 2}) || stub.userID != wallet.UserID {
				t.Fatalf("history response %s", rec.Body.String())
			}
			for _, path := range []string{"/wallet/me", "/transaction/me"} {
				if rec := serveOrderRequest(server.Handler(), http.MethodGet, path, false); rec.Code != 401 {
					t.Fatalf("unauthenticated %s: %d", path, rec.Code)
				}
			}
			before := stub.calls
			for _, path := range []string{"/transaction/me?limit=0", "/transaction/me?limit=101", "/transaction/me?offset=-1"} {
				if rec := serveOrderRequest(server.Handler(), http.MethodGet, path, true); rec.Code != 422 {
					t.Fatalf("invalid query %s: %d", path, rec.Code)
				}
			}
			if stub.calls != before {
				t.Fatal("invalid request reached usecase")
			}
		})
	}
}
