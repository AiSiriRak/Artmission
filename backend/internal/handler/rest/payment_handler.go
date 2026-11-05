package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

const maxSlipImageSize = 5 << 20

type PaymentHandler struct {
	usecase     payment.Usecase
	authUsecase auth.AuthUsecase
}

func NewPaymentHandler(usecase payment.Usecase, authUsecase auth.AuthUsecase) *PaymentHandler {
	return &PaymentHandler{usecase: usecase, authUsecase: authUsecase}
}

func (h *PaymentHandler) Register(api huma.API) {
	huma.Get(api, "/orders/{order_id}/payment/qr", h.getQR,
		huma.OperationTags("payments"),
		func(o *huma.Operation) {
			o.OperationID = "get-payment-qr"
			o.Summary = "GetPaymentQR"
			o.Description = "Retrieve the authenticated customer's existing, unexpired payment QR payload."
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireRole(api, user.RoleCustomer),
			)
		},
	)

	huma.Post(api, "/orders/{order_id}/payment/qr", h.generateQR,
		huma.OperationTags("payments"),
		func(o *huma.Operation) {
			o.OperationID = "generate-payment-qr"
			o.Summary = "GeneratePaymentQR"
			o.Description = "Generate the payment QR once for the authenticated customer's NOT_PAID order. Repeated calls return the stored QR without calling Slip2Go again."
			o.DefaultStatus = http.StatusOK
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireRole(api, user.RoleCustomer),
			)
		},
	)

	huma.Post(api, "/orders/{order_id}/payment/verify", h.verifySlip,
		huma.OperationTags("payments"),
		func(o *huma.Operation) {
			o.OperationID = "verify-payment-slip"
			o.Summary = "VerifyPaymentSlip"
			o.Description = "Verify a slip image for the authenticated customer's active order payment."
			o.DefaultStatus = http.StatusOK
			o.Middlewares = append(
				o.Middlewares,
				requireAuth(api, h.authUsecase),
				requireRole(api, user.RoleCustomer),
			)
		},
	)
}

type GetPaymentQRInput struct {
	OrderID uuid.UUID `path:"order_id"`
}

type GetPaymentQROutput struct {
	Body paymentQRView
}

type paymentQRView struct {
	Success   bool      `json:"success" doc:"true when a usable QR payload is returned"`
	Message   string    `json:"message"`
	PaymentID uuid.UUID `json:"payment_id"`
	OrderID   uuid.UUID `json:"order_id"`
	QRPayload string    `json:"qr_payload"`
	QRExpire  time.Time `json:"qr_expire"`
}

type GeneratePaymentQROutput struct {
	Body generatedPaymentQRView
}

type generatedPaymentQRView struct {
	Success          bool      `json:"success" doc:"true when a usable QR payload is returned"`
	Message          string    `json:"message"`
	AlreadyGenerated bool      `json:"already_generated" doc:"true when an existing QR was returned and Slip2Go was not called"`
	PaymentID        uuid.UUID `json:"payment_id"`
	OrderID          uuid.UUID `json:"order_id"`
	QRPayload        string    `json:"qr_payload"`
	QRExpire         time.Time `json:"qr_expire"`
}

func (h *PaymentHandler) getQR(
	ctx context.Context,
	input *GetPaymentQRInput,
) (*GetPaymentQROutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	result, err := h.usecase.GetQR(ctx, info.UserID, input.OrderID)
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return &GetPaymentQROutput{Body: paymentQRView{
		Success:   true,
		Message:   "Payment QR retrieved.",
		PaymentID: result.PaymentID,
		OrderID:   result.OrderID,
		QRPayload: result.Payload,
		QRExpire:  result.ExpiresAt,
	}}, nil
}

func (h *PaymentHandler) generateQR(
	ctx context.Context,
	input *GetPaymentQRInput,
) (*GeneratePaymentQROutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}
	result, err := h.usecase.GenerateQR(ctx, info.UserID, input.OrderID)
	if err != nil {
		return nil, mapPaymentError(err)
	}
	message := "Payment QR generated successfully."
	if result.AlreadyExisted {
		message = "Payment QR was already generated; returning the existing QR."
	}
	return &GeneratePaymentQROutput{Body: generatedPaymentQRView{
		Success:          true,
		Message:          message,
		AlreadyGenerated: result.AlreadyExisted,
		PaymentID:        result.PaymentID,
		OrderID:          result.OrderID,
		QRPayload:        result.Payload,
		QRExpire:         result.ExpiresAt,
	}}, nil
}

type verifySlipForm struct {
	File huma.FormFile `form:"file" contentType:"image/jpeg,image/png,image/webp" required:"true"`
}

type VerifyPaymentSlipInput struct {
	OrderID uuid.UUID `path:"order_id"`
	RawBody huma.MultipartFormFiles[verifySlipForm]
}

type VerifyPaymentSlipOutput struct {
	Body verifiedPaymentView
}

type verifiedPaymentView struct {
	Success   bool       `json:"success" doc:"true when the slip passed and the payment was recorded"`
	Message   string     `json:"message" doc:"Human-readable result; explains why the slip was not accepted"`
	Reason    string     `json:"reason,omitempty" doc:"Machine-readable rejection reason when success is false"`
	OrderID   uuid.UUID  `json:"order_id"`
	PaymentID *uuid.UUID `json:"payment_id,omitempty"`
	Status    string     `json:"status,omitempty"`
}

func (h *PaymentHandler) verifySlip(
	ctx context.Context,
	input *VerifyPaymentSlipInput,
) (*VerifyPaymentSlipOutput, error) {
	info, ok := authInfoFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing authentication")
	}

	image, contentType, err := readSlipImage(input.RawBody.Data().File)
	if err != nil {
		return nil, err
	}
	result, err := h.usecase.VerifySlip(ctx, info.UserID, input.OrderID, image, contentType)
	var rejection *payment.SlipRejection
	if errors.As(err, &rejection) {
		return &VerifyPaymentSlipOutput{Body: verifiedPaymentView{
			Success: false,
			Message: rejection.Message,
			Reason:  rejection.Reason,
			OrderID: input.OrderID,
		}}, nil
	}
	if err != nil {
		return nil, mapPaymentError(err)
	}
	return &VerifyPaymentSlipOutput{Body: verifiedPaymentView{
		Success:   true,
		Message:   "Payment verified successfully.",
		OrderID:   result.OrderID,
		PaymentID: &result.PaymentID,
		Status:    string(result.Status),
	}}, nil
}

func readSlipImage(file huma.FormFile) ([]byte, string, error) {
	if file.File == nil {
		return nil, "", huma.Error422UnprocessableEntity("slip image is required")
	}
	if file.Size > maxSlipImageSize {
		if err := file.File.Close(); err != nil {
			return nil, "", huma.Error500InternalServerError("could not close uploaded slip image", err)
		}
		return nil, "", huma.NewError(http.StatusRequestEntityTooLarge, "slip image exceeds the 5 MiB limit")
	}
	data, readErr := io.ReadAll(io.LimitReader(file.File, maxSlipImageSize+1))
	closeErr := file.File.Close()
	if readErr != nil || closeErr != nil {
		return nil, "", huma.Error500InternalServerError("could not read uploaded slip image", errors.Join(readErr, closeErr))
	}
	if len(data) > maxSlipImageSize {
		return nil, "", huma.NewError(http.StatusRequestEntityTooLarge, "slip image exceeds the 5 MiB limit")
	}
	if len(data) == 0 {
		return nil, "", huma.Error422UnprocessableEntity("slip image is empty")
	}

	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return data, contentType, nil
	default:
		return nil, "", huma.Error422UnprocessableEntity("slip image must be JPEG, PNG, or WebP")
	}
}

func mapPaymentError(err error) error {
	if errors.Is(err, payment.ErrQRNotReady) {
		return huma.Error404NotFound("payment QR has not been generated; call POST on the same path to create it")
	}
	if errors.Is(err, payment.ErrQRExpired) {
		return huma.NewError(http.StatusGone, "payment QR has expired")
	}
	return mapAppError(err)
}
