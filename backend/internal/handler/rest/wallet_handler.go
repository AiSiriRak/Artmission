package rest

import (
	"context"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type WalletHandler struct {
	wallet payment.WalletUsecase
	auth   auth.AuthUsecase
}

func NewWalletHandler(wallet payment.WalletUsecase, authUsecase auth.AuthUsecase) *WalletHandler {
	return &WalletHandler{wallet: wallet, auth: authUsecase}
}

func (h *WalletHandler) Register(api huma.API) {
	operation := func(id, summary, description string) func(*huma.Operation) {
		return func(o *huma.Operation) {
			o.OperationID, o.Summary, o.Description = id, summary, description
			o.Middlewares = append(o.Middlewares, requireAuth(api, h.auth), requireAnyRole(api, user.RoleCustomer, user.RoleArtist))
		}
	}
	huma.Get(api, "/wallet/me", h.getWallet, huma.OperationTags("wallet"), operation("get-wallet", "GetWallet", "Get the authenticated user's internal wallet balance in satang."))
	huma.Get(api, "/transaction/me", h.getTransactions, huma.OperationTags("wallet"), operation("get-wallet-transactions", "GetWalletTransactions", "List the authenticated user's credited order payouts and refunds, newest first."))
}

type getWalletInput struct{}
type walletView struct {
	UserID        uuid.UUID  `json:"user_id"`
	Currency      string     `json:"currency"`
	BalanceSatang int64      `json:"balance_satang"`
	UpdatedAt     *time.Time `json:"updated_at"`
}
type getWalletOutput struct{ Body walletView }

func (h *WalletHandler) getWallet(ctx context.Context, _ *getWalletInput) (*getWalletOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	wallet, err := h.wallet.GetWallet(ctx, info.UserID)
	if err != nil {
		return nil, mapAppError(err)
	}
	return &getWalletOutput{Body: walletView{UserID: wallet.UserID, Currency: "THB", BalanceSatang: wallet.BalanceSatang, UpdatedAt: wallet.UpdatedAt}}, nil
}

type getWalletTransactionsInput struct {
	Limit  int `query:"limit" default:"20" minimum:"1" maximum:"100"`
	Offset int `query:"offset" default:"0" minimum:"0"`
}
type walletTransactionView struct {
	ID                uuid.UUID                  `json:"id"`
	OrderID           uuid.UUID                  `json:"order_id"`
	OrderName         string                     `json:"order_name"`
	Type              payment.TransactionType    `json:"type"`
	AmountSatang      int64                      `json:"amount_satang"`
	SettlementOutcome *payment.SettlementOutcome `json:"settlement_outcome"`
	CreatedAt         time.Time                  `json:"created_at"`
}
type getWalletTransactionsOutput struct {
	Body struct {
		Transactions []walletTransactionView `json:"transactions"`
		Total        int                     `json:"total"`
		Limit        int                     `json:"limit"`
		Offset       int                     `json:"offset"`
	}
}

func (h *WalletHandler) getTransactions(ctx context.Context, in *getWalletTransactionsInput) (*getWalletTransactionsOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	query := payment.TransactionQuery{Limit: in.Limit, Offset: in.Offset}
	if query.Limit == 0 {
		query.Limit = payment.DefaultTransactionLimit
	}
	page, err := h.wallet.ListWalletTransactions(ctx, info.UserID, query)
	if err != nil {
		return nil, mapAppError(err)
	}
	out := &getWalletTransactionsOutput{}
	out.Body.Total, out.Body.Limit, out.Body.Offset = page.Total, query.Limit, query.Offset
	out.Body.Transactions = make([]walletTransactionView, len(page.Transactions))
	for i, t := range page.Transactions {
		out.Body.Transactions[i] = walletTransactionView{ID: t.ID, OrderID: t.OrderID, OrderName: t.OrderName, Type: t.Type, AmountSatang: t.AmountSatang, SettlementOutcome: t.SettlementOutcome, CreatedAt: t.CreatedAt}
	}
	return out, nil
}
