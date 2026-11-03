package payment_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type fakeRepository struct {
	mu             sync.Mutex
	order          *payment.OrderForPayment
	qr             *payment.Payment
	candidates     []payment.QRGenerationCandidate
	verification   *payment.PaymentForVerification
	expiredIDs     []uuid.UUID
	getErr         error
	qrErr          error
	listErr        error
	candidateErr   error
	lockErr        error
	saveErr        error
	expireErr      error
	verifyErr      error
	confirmErr     error
	savedPayload   string
	confirmedRef   string
	locks          map[uuid.UUID]bool
	expiredOrderID uuid.UUID
	orderCancelled bool
}

func (f *fakeRepository) GetOrderForPayment(context.Context, uuid.UUID, uuid.UUID) (*payment.OrderForPayment, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.order == nil {
		return nil, payment.ErrOrderNotFound
	}
	return f.order, nil
}

func (f *fakeRepository) GetQRForOrder(context.Context, uuid.UUID) (*payment.Payment, error) {
	if f.qrErr != nil {
		return nil, f.qrErr
	}
	if f.qr == nil {
		return nil, payment.ErrQRNotReady
	}
	return f.qr, nil
}

func (f *fakeRepository) ListQRGenerationCandidates(
	_ context.Context,
	_ int,
) ([]payment.QRGenerationCandidate, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.qr != nil && f.qr.QRPayload != nil {
		return nil, nil
	}
	return f.candidates, nil
}

func (f *fakeRepository) GetQRGenerationCandidate(
	_ context.Context,
	paymentID uuid.UUID,
) (*payment.QRGenerationCandidate, error) {
	if f.candidateErr != nil {
		return nil, f.candidateErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.qr != nil && f.qr.QRPayload != nil {
		return nil, nil
	}
	for _, candidate := range f.candidates {
		if candidate.PaymentID == paymentID {
			copy := candidate
			return &copy, nil
		}
	}
	return nil, nil
}

func (f *fakeRepository) TryLockQRGeneration(
	_ context.Context,
	paymentID uuid.UUID,
) (payment.QRGenerationLock, bool, error) {
	if f.lockErr != nil {
		return nil, false, f.lockErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.locks == nil {
		f.locks = make(map[uuid.UUID]bool)
	}
	if f.locks[paymentID] {
		return nil, false, nil
	}
	f.locks[paymentID] = true
	return fakeQRLock{release: func() error {
		f.mu.Lock()
		delete(f.locks, paymentID)
		f.mu.Unlock()
		return nil
	}}, true, nil
}

func (f *fakeRepository) SaveQRPayload(
	_ context.Context,
	_ uuid.UUID,
	payload string,
) (bool, error) {
	if f.saveErr != nil {
		return false, f.saveErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.qr != nil && f.qr.QRPayload != nil {
		return false, nil
	}
	f.savedPayload = payload
	if f.qr != nil {
		f.qr.QRPayload = &payload
	}
	return true, nil
}

func (f *fakeRepository) ListExpiredPaymentIDs(context.Context, int) ([]uuid.UUID, error) {
	return f.expiredIDs, nil
}

func (f *fakeRepository) ExpirePayment(
	_ context.Context,
	paymentID uuid.UUID,
) (uuid.UUID, bool, bool, error) {
	if f.expireErr != nil {
		return uuid.Nil, false, false, f.expireErr
	}
	for _, id := range f.expiredIDs {
		if id == paymentID {
			return f.expiredOrderID, f.orderCancelled, true, nil
		}
	}
	return uuid.Nil, false, false, nil
}

func (f *fakeRepository) GetPaymentForVerification(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (*payment.PaymentForVerification, error) {
	if f.verifyErr != nil {
		return nil, f.verifyErr
	}
	if f.verification == nil {
		return nil, payment.ErrOrderNotFound
	}
	return f.verification, nil
}

func (f *fakeRepository) ConfirmSlipPayment(
	_ context.Context,
	_, _ uuid.UUID,
	transRef string,
) (*payment.VerifiedPayment, error) {
	if f.confirmErr != nil {
		return nil, f.confirmErr
	}
	f.confirmedRef = transRef
	return &payment.VerifiedPayment{
		PaymentID: f.verification.PaymentID,
		OrderID:   f.verification.OrderID,
		Status:    payment.StatusHeld,
	}, nil
}

type fakeQRLock struct {
	release func() error
}

func (l fakeQRLock) Release() error { return l.release() }

type fakeProvider struct {
	mu           sync.Mutex
	qr           payment.QRImage
	qrErr        error
	verification payment.SlipVerification
	verifyErr    error
	qrCalls      int
	verifyCalls  int
	amount       int64
}

func (f *fakeProvider) GeneratePromptPayQR(_ context.Context, amount int64) (payment.QRImage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.qrCalls++
	f.amount = amount
	return f.qr, f.qrErr
}

func (f *fakeProvider) VerifySlipImage(
	context.Context,
	[]byte,
	string,
	int64,
	time.Time,
) (payment.SlipVerification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.verifyCalls++
	return f.verification, f.verifyErr
}

func TestGetQRReturnsStoredPayloadWithoutGenerating(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	expires := time.Now().Add(time.Hour)
	payload := "stored-qr-payload"
	repo := &fakeRepository{
		order: &payment.OrderForPayment{ID: orderID, Status: payment.EligibleOrderStatus},
		qr: &payment.Payment{
			ID: uuid.New(), OrderID: orderID, Status: payment.StatusPending,
			QRPayload: &payload, QRExpire: &expires,
		},
	}
	provider := &fakeProvider{}

	result, err := payment.NewPaymentUsecase(repo, provider).GetQR(context.Background(), customerID, orderID)
	if err != nil {
		t.Fatalf("GetQR() error = %v", err)
	}
	if result.Payload != payload || result.ExpiresAt != expires || result.OrderID != orderID {
		t.Errorf("GetQR() = %+v, want existing payload and expiry", result)
	}
	if provider.qrCalls != 0 || provider.verifyCalls != 0 {
		t.Errorf("GET invoked provider %d times", provider.qrCalls+provider.verifyCalls)
	}
}

func TestGetQRDoesNotReturnIneligibleOrExpiredPayload(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	payload := "stored-payload"
	for _, test := range []struct {
		name        string
		orderStatus string
		expires     time.Time
		wantErr     error
	}{
		{name: "cancelled order", orderStatus: "CANCEL", expires: time.Now().Add(time.Hour), wantErr: payment.ErrOrderNotPayable},
		{name: "expired qr", orderStatus: payment.EligibleOrderStatus, expires: time.Now().Add(-time.Second), wantErr: payment.ErrQRExpired},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepository{
				order: &payment.OrderForPayment{ID: orderID, Status: test.orderStatus},
				qr: &payment.Payment{
					ID: uuid.New(), OrderID: orderID, Status: payment.StatusPending,
					QRPayload: &payload, QRExpire: &test.expires,
				},
			}
			_, err := payment.NewPaymentUsecase(repo, &fakeProvider{}).GetQR(context.Background(), customerID, orderID)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("GetQR() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestGetQRReturnsNotReadyWhenPayloadIsMissing(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	expires := time.Now().Add(time.Hour)
	repo := &fakeRepository{
		order: &payment.OrderForPayment{ID: orderID, Status: payment.EligibleOrderStatus},
		qr: &payment.Payment{
			ID: uuid.New(), OrderID: orderID, Status: payment.StatusPending, QRExpire: &expires,
		},
	}
	_, err := payment.NewPaymentUsecase(repo, &fakeProvider{}).GetQR(context.Background(), customerID, orderID)
	if !errors.Is(err, payment.ErrQRNotReady) {
		t.Fatalf("GetQR() error = %v, want QR not ready", err)
	}
}

func newGenerateFixture(amount int64, status string, expires time.Time) (*fakeRepository, uuid.UUID, uuid.UUID) {
	paymentID, orderID := uuid.New(), uuid.New()
	repo := &fakeRepository{
		order: &payment.OrderForPayment{ID: orderID, Status: status},
		candidates: []payment.QRGenerationCandidate{{
			PaymentID: paymentID, OrderID: orderID, AmountSatang: amount,
		}},
		qr: &payment.Payment{
			ID: paymentID, OrderID: orderID, Status: payment.StatusPending, QRExpire: &expires,
		},
	}
	return repo, paymentID, orderID
}

func TestGenerateQRPersistsPayloadOnlyOnce(t *testing.T) {
	repo, _, orderID := newGenerateFixture(12345, payment.EligibleOrderStatus, time.Now().Add(time.Hour))
	provider := &fakeProvider{qr: payment.QRImage{Payload: "new-qr-payload", AmountSatang: 12345}}
	uc := payment.NewPaymentUsecase(repo, provider)
	customerID := uuid.New()

	first, err := uc.GenerateQR(context.Background(), customerID, orderID)
	if err != nil || first.Payload != "new-qr-payload" || first.AlreadyExisted {
		t.Fatalf("first GenerateQR() = %+v, %v; want newly generated", first, err)
	}
	second, err := uc.GenerateQR(context.Background(), customerID, orderID)
	if err != nil || second.Payload != "new-qr-payload" || !second.AlreadyExisted {
		t.Fatalf("second GenerateQR() = %+v, %v; want existing QR", second, err)
	}
	if provider.qrCalls != 1 || provider.amount != 12345 || repo.savedPayload != "new-qr-payload" {
		t.Errorf("provider calls/amount or persisted payload = %d/%d/%q", provider.qrCalls, provider.amount, repo.savedPayload)
	}
}

func TestGenerateQRDoesNotPersistWhenProviderFails(t *testing.T) {
	repo, _, orderID := newGenerateFixture(500, payment.EligibleOrderStatus, time.Now().Add(time.Hour))
	provider := &fakeProvider{qrErr: errors.New("provider timeout")}
	_, err := payment.NewPaymentUsecase(repo, provider).GenerateQR(context.Background(), uuid.New(), orderID)
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInternal {
		t.Fatalf("GenerateQR() error = %v, want internal provider error", err)
	}
	if repo.savedPayload != "" || repo.qr.QRPayload != nil {
		t.Fatal("provider failure left a usable QR payload")
	}
	if provider.qrCalls != 1 {
		t.Errorf("provider calls = %d, want 1 (no retry loop)", provider.qrCalls)
	}
}

func TestGenerateQRRejectsIneligibleOrExpiredPayment(t *testing.T) {
	for _, test := range []struct {
		name    string
		status  string
		expires time.Time
		wantErr error
	}{
		{"cancelled order", "CANCEL", time.Now().Add(time.Hour), payment.ErrOrderNotPayable},
		{"expired qr", payment.EligibleOrderStatus, time.Now().Add(-time.Second), payment.ErrQRExpired},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, _, orderID := newGenerateFixture(500, test.status, test.expires)
			provider := &fakeProvider{qr: payment.QRImage{Payload: "x", AmountSatang: 500}}
			_, err := payment.NewPaymentUsecase(repo, provider).GenerateQR(context.Background(), uuid.New(), orderID)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("GenerateQR() error = %v, want %v", err, test.wantErr)
			}
			if provider.qrCalls != 0 {
				t.Error("provider was called for an ineligible payment")
			}
		})
	}
}

func TestGenerateQRSerializesConcurrentRequests(t *testing.T) {
	repo, _, orderID := newGenerateFixture(500, payment.EligibleOrderStatus, time.Now().Add(time.Hour))
	provider := &fakeProvider{qr: payment.QRImage{Payload: "single-payload", AmountSatang: 500}}
	uc := payment.NewPaymentUsecase(repo, provider)
	customerID := uuid.New()

	start := make(chan struct{})
	var wait sync.WaitGroup
	for range 4 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, err := uc.GenerateQR(context.Background(), customerID, orderID)
			if err != nil && !errors.Is(err, payment.ErrQRInProgress) {
				t.Errorf("GenerateQR() error = %v", err)
			}
		}()
	}
	close(start)
	wait.Wait()

	if provider.qrCalls != 1 {
		t.Errorf("provider calls = %d, want 1", provider.qrCalls)
	}
}
func TestExpirePaymentsReturnsCancelledOrderIDs(t *testing.T) {
	paymentID, orderID := uuid.New(), uuid.New()
	repo := &fakeRepository{
		expiredIDs:     []uuid.UUID{paymentID},
		expiredOrderID: orderID,
		orderCancelled: true,
	}
	cancelled, err := payment.NewPaymentUsecase(repo, &fakeProvider{}).ExpirePayments(context.Background())
	if err != nil {
		t.Fatalf("ExpirePayments() error = %v", err)
	}
	if len(cancelled) != 1 || cancelled[0] != orderID {
		t.Fatalf("cancelled orders = %v, want [%s]", cancelled, orderID)
	}
}

func TestVerifySlipRejectsExpiredOrCancelledOrdersBeforeProvider(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	payload := "qr"
	for _, test := range []struct {
		name        string
		orderStatus string
		status      payment.Status
		expires     time.Time
		wantErr     error
	}{
		{name: "cancelled order", orderStatus: "CANCEL", status: payment.StatusPending, expires: time.Now().Add(time.Hour), wantErr: payment.ErrOrderNotPayable},
		{name: "expired qr", orderStatus: payment.EligibleOrderStatus, status: payment.StatusPending, expires: time.Now().Add(-time.Second), wantErr: payment.ErrQRExpired},
		{name: "already verified", orderStatus: "IN_PROCESS", status: payment.StatusHeld, expires: time.Now().Add(time.Hour), wantErr: payment.ErrOrderNotPayable},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepository{verification: &payment.PaymentForVerification{
				PaymentID: uuid.New(), OrderID: orderID, AmountSatang: 1234,
				Status: test.status, OrderStatus: test.orderStatus,
				QRPayload: &payload, QRExpire: &test.expires,
			}}
			provider := &fakeProvider{verification: payment.SlipVerification{Valid: true, TransRef: "bank-ref", AmountSatang: 1234}}
			_, err := payment.NewPaymentUsecase(repo, provider).VerifySlip(
				context.Background(), customerID, orderID, []byte("image"), "image/png",
			)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("VerifySlip() error = %v, want %v", err, test.wantErr)
			}
			if provider.verifyCalls != 0 {
				t.Fatal("provider was called for an ineligible payment")
			}
		})
	}
}

func TestVerifySlipConfirmsOnlyValidUniqueReference(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	payload := "qr"
	expires := time.Now().Add(time.Hour)
	repo := &fakeRepository{verification: &payment.PaymentForVerification{
		PaymentID: uuid.New(), OrderID: orderID, AmountSatang: 1234,
		Status: payment.StatusPending, OrderStatus: payment.EligibleOrderStatus,
		QRPayload: &payload, QRExpire: &expires,
	}}
	provider := &fakeProvider{verification: payment.SlipVerification{
		Valid: true, ResultCode: "200000", TransRef: "bank-ref", AmountSatang: 1234,
	}}

	result, err := payment.NewPaymentUsecase(repo, provider).VerifySlip(
		context.Background(), customerID, orderID, []byte("image"), "image/png",
	)
	if err != nil {
		t.Fatalf("VerifySlip() error = %v", err)
	}
	if result.Status != payment.StatusHeld || repo.confirmedRef != "bank-ref" || provider.verifyCalls != 1 {
		t.Errorf("verification result/repository/provider = %+v/%q/%d", result, repo.confirmedRef, provider.verifyCalls)
	}
}

func TestVerifySlipReportsReadableRejectionReasons(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	payload := "qr"
	expires := time.Now().Add(time.Hour)
	for _, test := range []struct {
		name         string
		verification payment.SlipVerification
		wantReason   string
	}{
		{"duplicate", payment.SlipVerification{ResultCode: "200501"}, "SLIP_ALREADY_USED"},
		{"slip older than the payment", payment.SlipVerification{ResultCode: "200403"}, "SLIP_TOO_OLD"},
		{"wrong receiver", payment.SlipVerification{ResultCode: "200401"}, "WRONG_RECEIVER"},
		{"provider amount mismatch", payment.SlipVerification{ResultCode: "200402"}, "AMOUNT_MISMATCH"},
		{"slip amount differs", payment.SlipVerification{ResultCode: "200200", TransRef: "r", AmountSatang: 9999}, "AMOUNT_MISMATCH"},
		{"invalid image", payment.SlipVerification{ResultCode: "400002"}, "INVALID_IMAGE"},
		{"unknown code", payment.SlipVerification{ResultCode: "999999"}, "SLIP_REJECTED"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepository{verification: &payment.PaymentForVerification{
				PaymentID: uuid.New(), OrderID: orderID, AmountSatang: 1234,
				Status: payment.StatusPending, OrderStatus: payment.EligibleOrderStatus,
				QRPayload: &payload, QRExpire: &expires,
			}}
			_, err := payment.NewPaymentUsecase(repo, &fakeProvider{verification: test.verification}).VerifySlip(
				context.Background(), customerID, orderID, []byte("image"), "image/png",
			)
			var rejection *payment.SlipRejection
			if !errors.As(err, &rejection) || rejection.Reason != test.wantReason || rejection.Message == "" {
				t.Fatalf("VerifySlip() error = %v, want rejection %s", err, test.wantReason)
			}
			if repo.confirmedRef != "" {
				t.Fatal("rejected slip was persisted")
			}
		})
	}
}

func TestVerifySlipRejectsProviderFailureAndAmountMismatch(t *testing.T) {
	customerID, orderID := uuid.New(), uuid.New()
	payload := "qr"
	expires := time.Now().Add(time.Hour)
	for _, test := range []struct {
		name         string
		verification payment.SlipVerification
		providerErr  error
		wantErrCode  apperror.Code
	}{
		{name: "provider mismatch", verification: payment.SlipVerification{Valid: true, TransRef: "ref", AmountSatang: 9999}, wantErrCode: apperror.CodeInternal},
		{name: "provider times out", providerErr: errors.New("timeout"), wantErrCode: apperror.CodeInternal},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepository{verification: &payment.PaymentForVerification{
				PaymentID: uuid.New(), OrderID: orderID, AmountSatang: 1234,
				Status: payment.StatusPending, OrderStatus: payment.EligibleOrderStatus,
				QRPayload: &payload, QRExpire: &expires,
			}}
			provider := &fakeProvider{verification: test.verification, verifyErr: test.providerErr}
			_, err := payment.NewPaymentUsecase(repo, provider).VerifySlip(
				context.Background(), customerID, orderID, []byte("image"), "image/png",
			)
			var appErr *apperror.Error
			if !errors.As(err, &appErr) || appErr.Code != test.wantErrCode {
				t.Fatalf("VerifySlip() error = %v, want code %s", err, test.wantErrCode)
			}
			if repo.confirmedRef != "" {
				t.Fatal("provider failure or mismatch was persisted")
			}
		})
	}
}
