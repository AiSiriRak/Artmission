package payment

import (
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type SettlementOutcome string

const (
	Completed         SettlementOutcome = "COMPLETED"
	DeadlineExpired   SettlementOutcome = "DEADLINE_EXPIRED"
	ArtistCancelled   SettlementOutcome = "ARTIST_CANCELLED"
	CustomerCancelled SettlementOutcome = "CUSTOMER_CANCELLED"
)

type TransactionType string

const (
	Charge           TransactionType = "CHARGE"
	ReleaseToArtist  TransactionType = "RELEASE_TO_ARTIST"
	RefundToCustomer TransactionType = "REFUND_TO_CUSTOMER"
)

type TransferStatus string

const (
	// TransferPending is an allocation awaiting execution, including incomplete historical settlements.
	TransferPending TransferStatus = "PENDING"
	// TransferSuccess means the wallet credit committed; it does not imply a bank transfer.
	TransferSuccess TransferStatus = "SUCCESS"
	// TransferFailed records a failed execution attempt, not a failed order payment.
	TransferFailed TransferStatus = "FAILED"
)

type Transaction struct {
	ID              uuid.UUID
	PaymentID       uuid.UUID
	Type            TransactionType
	AmountSatang    int64
	TransferStatus  *TransferStatus
	TransferRef     *string
	TransferredAt   *time.Time
	FailureReason   *string
	CreatedAt       time.Time
	RecipientUserID *uuid.UUID
}

type Summary struct {
	RefundTransactionID        *uuid.UUID
	ArtistPayoutTransactionID  *uuid.UUID
	Status                     string
	Source                     string
	AmountSatang               int64
	SettlementOutcome          *SettlementOutcome
	SettledAt                  *time.Time
	RefundSatang               int64
	ArtistPayoutSatang         int64
	RefundTransferStatus       *TransferStatus
	ArtistPayoutTransferStatus *TransferStatus
}

type Wallet struct {
	UserID        uuid.UUID
	BalanceSatang int64
	UpdatedAt     *time.Time
}

type WalletTransaction struct {
	ID                uuid.UUID
	OrderID           uuid.UUID
	OrderName         string
	Type              TransactionType
	AmountSatang      int64
	SettlementOutcome *SettlementOutcome
	CreatedAt         time.Time
}

type TransactionQuery struct {
	Limit  int
	Offset int
}

type TransactionPage struct {
	Transactions []WalletTransaction
	Total        int
}

func CalculateSettlement(outcome SettlementOutcome, amountSatang int64) (refundSatang, artistSatang int64, err error) {
	if amountSatang < 0 {
		return 0, 0, apperror.InvalidInput("payment amount must not be negative", nil)
	}
	switch outcome {
	case Completed:
		// Platform fee is currently zero; future fee policy belongs here.
		return 0, amountSatang, nil
	case DeadlineExpired, ArtistCancelled:
		return amountSatang, 0, nil
	case CustomerCancelled:
		return amountSatang / 2, amountSatang - amountSatang/2, nil
	default:
		return 0, 0, apperror.InvalidInput("unsupported settlement outcome", nil)
	}
}
