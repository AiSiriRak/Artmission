package payment

import (
	"fmt"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
)

var (
	ErrOrderNotFound        = apperror.NotFound("order not found")
	ErrOrderNotPayable      = apperror.Conflict("order is not eligible for payment")
	ErrPaymentNotVerifiable = apperror.Conflict("payment is not eligible for verification")
	ErrQRNotReady           = apperror.Conflict("payment QR has not been generated")
	ErrQRInProgress         = apperror.Conflict("payment QR generation is already in progress")
	ErrQRExpired            = apperror.Conflict("payment QR has expired")
	ErrSlipRejected         = apperror.Conflict("slip could not be verified")
	ErrDuplicateSlip        = apperror.Conflict("slip has already been used")
	ErrMissingCustomerID    = apperror.InvalidInput("missing customer id", nil)
	ErrMissingOrderID       = apperror.InvalidInput("missing order id", nil)
	ErrInvalidSlipImage     = apperror.InvalidInput("slip image is invalid", nil)
)

// SlipRejection is a business outcome of slip verification (not a system failure).
// It is reported to the client as success=false with a readable message.
type SlipRejection struct {
	Reason  string
	Message string
}

func (e *SlipRejection) Error() string { return e.Message }

func (e *SlipRejection) Is(target error) bool {
	return target == ErrSlipRejected
}

func newSlipRejection(reason, message string) *SlipRejection {
	return &SlipRejection{Reason: reason, Message: message}
}

func rejectionFromVerification(v SlipVerification, orderAmountSatang int64) *SlipRejection {
	switch v.ResultCode {
	case "200501":
		return newSlipRejection("SLIP_ALREADY_USED", "This slip has already been used.")
	case "200401":
		return newSlipRejection("WRONG_RECEIVER", "The slip was not transferred to the shop's payment account.")
	case "200402":
		return newSlipRejection("AMOUNT_MISMATCH", "The slip amount does not match the order amount.")
	case "200404":
		return newSlipRejection("SLIP_NOT_FOUND", "No matching transfer was found for this slip. Use the original slip image.")
	case "200500":
		return newSlipRejection("SLIP_FRAUD", "The slip looks invalid or tampered with.")
	case "400002":
		return newSlipRejection("INVALID_IMAGE", "The file is not a readable slip image. Upload the original slip with its QR code visible.")
	case "200000", "200200":
		if v.AmountSatang != orderAmountSatang {
			return newSlipRejection("AMOUNT_MISMATCH", fmt.Sprintf(
				"The slip amount (%s THB) does not match the order amount (%s THB).",
				formatBaht(v.AmountSatang), formatBaht(orderAmountSatang)))
		}
	}
	return newSlipRejection("SLIP_REJECTED", "The slip could not be verified (provider code "+v.ResultCode+").")
}

func formatBaht(satang int64) string {
	return fmt.Sprintf("%d.%02d", satang/100, satang%100)
}
