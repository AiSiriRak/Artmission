package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/auth"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/modules/user"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/httpserver"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/logger"
	"github.com/google/uuid"
)

type paymentHandlerAuth struct {
	auth.AuthUsecase
	claims *auth.TokenClaims
}

func (f paymentHandlerAuth) Authenticate(_ context.Context, token string) (*auth.TokenClaims, error) {
	if token != "valid-token" {
		return nil, errors.New("invalid token")
	}
	return f.claims, nil
}

type paymentHandlerUsecase struct {
	qrResult       *payment.QRPayment
	qrErr          error
	verifyResult   *payment.VerifiedPayment
	verifyErr      error
	qrCalls        int
	verifyCalls    int
	generateCalls  int
	gotCustomer    uuid.UUID
	gotOrder       uuid.UUID
	gotImage       []byte
	gotContentType string
}

func (f *paymentHandlerUsecase) GetQR(
	_ context.Context,
	customerID, orderID uuid.UUID,
) (*payment.QRPayment, error) {
	f.qrCalls++
	f.gotCustomer, f.gotOrder = customerID, orderID
	return f.qrResult, f.qrErr
}

func (f *paymentHandlerUsecase) GenerateQR(
	_ context.Context,
	customerID, orderID uuid.UUID,
) (*payment.QRPayment, error) {
	f.generateCalls++
	f.gotCustomer, f.gotOrder = customerID, orderID
	return f.qrResult, f.qrErr
}

func (f *paymentHandlerUsecase) ExpirePayments(context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *paymentHandlerUsecase) VerifySlip(
	_ context.Context,
	customerID, orderID uuid.UUID,
	image []byte,
	contentType string,
) (*payment.VerifiedPayment, error) {
	f.verifyCalls++
	f.gotCustomer, f.gotOrder = customerID, orderID
	f.gotImage = image
	f.gotContentType = contentType
	return f.verifyResult, f.verifyErr
}

func TestPaymentHandlerGetQRReturnsExistingPayload(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	expires := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	usecase := &paymentHandlerUsecase{qrResult: &payment.QRPayment{
		PaymentID: uuid.New(), OrderID: orderID, Payload: "stored-payload", ExpiresAt: expires,
	}}
	handler := newPaymentTestHandlerForUser(t, user.RoleCustomer, customerID, usecase)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/orders/"+orderID.String()+"/payment/qr", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body struct {
		PaymentID string    `json:"payment_id"`
		OrderID   string    `json:"order_id"`
		QRPayload string    `json:"qr_payload"`
		QRExpire  time.Time `json:"qr_expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.PaymentID == "" || body.OrderID != orderID.String() ||
		body.QRPayload != "stored-payload" || !body.QRExpire.Equal(expires) {
		t.Errorf("response body = %+v, unexpected QR response", body)
	}
	if usecase.qrCalls != 1 || usecase.gotCustomer != customerID || usecase.gotOrder != orderID ||
		usecase.generateCalls != 0 {
		t.Errorf("GET calls = %d, generate calls = %d, customer/order = %s/%s",
			usecase.qrCalls, usecase.generateCalls, usecase.gotCustomer, usecase.gotOrder)
	}
}

func TestPaymentHandlerGetQRReturnsNotGeneratedStatus(t *testing.T) {
	usecase := &paymentHandlerUsecase{qrErr: payment.ErrQRNotReady}
	handler := newPaymentTestHandler(t, user.RoleCustomer, usecase)
	rec := servePaymentRequest(handler, http.MethodGet, "/orders/"+uuid.NewString()+"/payment/qr", "Bearer valid-token", nil, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if usecase.generateCalls != 0 {
		t.Fatal("GET triggered QR generation")
	}
}

func TestPaymentHandlerRequiresAuthentication(t *testing.T) {
	usecase := &paymentHandlerUsecase{}
	handler := newPaymentTestHandler(t, user.RoleCustomer, usecase)
	rec := servePaymentRequest(handler, http.MethodGet, "/orders/"+uuid.NewString()+"/payment/qr", "", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if usecase.qrCalls != 0 {
		t.Fatal("unauthenticated request reached payment usecase")
	}
}

func TestPaymentHandlerVerifiesUploadedSlip(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	usecase := &paymentHandlerUsecase{verifyResult: &payment.VerifiedPayment{
		PaymentID: uuid.New(), OrderID: orderID, Status: payment.StatusHeld,
	}}
	handler := newPaymentTestHandlerForUser(t, user.RoleCustomer, customerID, usecase)
	body, contentType := slipMultipartBody(t)
	rec := servePaymentRequest(
		handler,
		http.MethodPost,
		"/orders/"+orderID.String()+"/payment/verify",
		"Bearer valid-token",
		body,
		contentType,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if usecase.verifyCalls != 1 || usecase.gotCustomer != customerID ||
		usecase.gotOrder != orderID || !bytes.Equal(usecase.gotImage, []byte("\x89PNG\r\n\x1a\n")) ||
		usecase.gotContentType != "image/png" {
		t.Errorf("verify request = calls:%d customer:%s order:%s contentType:%s image:%x",
			usecase.verifyCalls, usecase.gotCustomer, usecase.gotOrder, usecase.gotContentType, usecase.gotImage)
	}
}

func TestPaymentHandlerReportsSlipRejectionAsSuccessFalse(t *testing.T) {
	orderID := uuid.New()
	usecase := &paymentHandlerUsecase{verifyErr: &payment.SlipRejection{
		Reason: "AMOUNT_MISMATCH", Message: "The slip amount does not match the order amount.",
	}}
	handler := newPaymentTestHandler(t, user.RoleCustomer, usecase)
	body, contentType := slipMultipartBody(t)
	rec := servePaymentRequest(handler, http.MethodPost,
		"/orders/"+orderID.String()+"/payment/verify", "Bearer valid-token", body, contentType)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var got struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Success || got.Reason != "AMOUNT_MISMATCH" || got.Message == "" {
		t.Errorf("response = %+v, want success=false with reason and message", got)
	}
}

func newPaymentTestHandler(
	t *testing.T,
	role user.Role,
	usecase payment.Usecase,
) http.Handler {
	return newPaymentTestHandlerForUser(t, role, uuid.New(), usecase)
}

func newPaymentTestHandlerForUser(
	t *testing.T,
	role user.Role,
	userID uuid.UUID,
	usecase payment.Usecase,
) http.Handler {
	t.Helper()
	authStub := paymentHandlerAuth{claims: &auth.TokenClaims{
		UserID: userID, Role: role, SessionID: uuid.New(),
	}}
	handler := NewPaymentHandler(usecase, authStub)
	api, server := httpserver.New(":0", "/api/v1", nil, logger.NewLogger(false), nil)
	handler.Register(api)
	return server.Handler()
}

func servePaymentRequest(
	handler http.Handler,
	method, path, authorization string,
	body io.Reader,
	contentType string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/v1"+path, body)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func slipMultipartBody(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="slip.png"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("\x89PNG\r\n\x1a\n")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return &body, writer.FormDataContentType()
}

func TestPaymentHandlerPostQRGeneratesForOwner(t *testing.T) {
	orderID := uuid.New()
	usecase := &paymentHandlerUsecase{qrResult: &payment.QRPayment{
		PaymentID: uuid.New(), OrderID: orderID, Payload: "new-payload", ExpiresAt: time.Now().Add(time.Hour),
	}}
	handler := newPaymentTestHandler(t, user.RoleCustomer, usecase)
	rec := servePaymentRequest(handler, http.MethodPost, "/orders/"+orderID.String()+"/payment/qr", "Bearer valid-token", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if usecase.generateCalls != 1 || usecase.qrCalls != 0 || usecase.gotOrder != orderID {
		t.Errorf("generate/get calls = %d/%d, order = %s", usecase.generateCalls, usecase.qrCalls, usecase.gotOrder)
	}
	var body struct {
		Success          bool   `json:"success"`
		Message          string `json:"message"`
		AlreadyGenerated bool   `json:"already_generated"`
		QRPayload        string `json:"qr_payload"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.AlreadyGenerated || body.QRPayload != "new-payload" || body.Message == "" {
		t.Errorf("body = %+v, want success with new payload", body)
	}
}

func TestPaymentHandlerPostQRReportsAlreadyGenerated(t *testing.T) {
	orderID := uuid.New()
	usecase := &paymentHandlerUsecase{qrResult: &payment.QRPayment{
		PaymentID: uuid.New(), OrderID: orderID, Payload: "old-payload",
		ExpiresAt: time.Now().Add(time.Hour), AlreadyExisted: true,
	}}
	handler := newPaymentTestHandler(t, user.RoleCustomer, usecase)
	rec := servePaymentRequest(handler, http.MethodPost, "/orders/"+orderID.String()+"/payment/qr", "Bearer valid-token", nil, "")
	var body struct {
		Success          bool   `json:"success"`
		AlreadyGenerated bool   `json:"already_generated"`
		QRPayload        string `json:"qr_payload"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || !body.Success || !body.AlreadyGenerated || body.QRPayload != "old-payload" {
		t.Errorf("status=%d body=%+v, want success, already_generated, existing payload", rec.Code, body)
	}
}
