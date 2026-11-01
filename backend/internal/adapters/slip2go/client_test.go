package slip2go

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
)

func TestGeneratePromptPayQR(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		code    string
		account string
		amount  string
		payload string
		wantErr bool
	}{
		{name: "valid payload", status: http.StatusOK, code: "200", account: "Artmission", amount: "12.34", payload: "000201-test"},
		{name: "provider business error", status: http.StatusOK, code: "400", account: "Artmission", amount: "12.34", payload: "000201-test", wantErr: true},
		{name: "http error", status: http.StatusBadGateway, code: "200", account: "Artmission", amount: "12.34", payload: "000201-test", wantErr: true},
		{name: "wrong account", status: http.StatusOK, code: "200", account: "Other", amount: "12.34", payload: "000201-test", wantErr: true},
		{name: "wrong amount", status: http.StatusOK, code: "200", account: "Artmission", amount: "12.35", payload: "000201-test", wantErr: true},
		{name: "missing payload", status: http.StatusOK, code: "200", account: "Artmission", amount: "12.34", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != generateQRPath {
					t.Errorf("request path = %q, want %q", r.URL.Path, generateQRPath)
				}
				if r.Header.Get("Authorization") != "Bearer test-secret" {
					t.Errorf("Authorization header = %q, want Bearer-prefixed configured secret", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type header = %q, want application/json", r.Header.Get("Content-Type"))
				}
				var request generateQRRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if request.Amount != "12.34" || request.PromptPayCode != "0812345678" {
					t.Errorf("request = %+v, want configured destination and exact amount", request)
				}

				w.WriteHeader(tt.status)
				_ = json.NewEncoder(w).Encode(generateQRResponse{
					Code: tt.code,
					Data: struct {
						QRCode      string `json:"qrCode"`
						AccountName string `json:"accountName"`
						Amount      string `json:"amount"`
					}{QRCode: tt.payload, AccountName: tt.account, Amount: tt.amount},
				})
			}))
			defer server.Close()

			got, err := newClient(testConfig(server.URL), server.Client()).GeneratePromptPayQR(context.Background(), 1234)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GeneratePromptPayQR() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && (got.Payload != tt.payload || got.AmountSatang != 1234) {
				t.Errorf("GeneratePromptPayQR() = %+v, unexpected QR data", got)
			}
		})
	}
}

func TestVerifySlipImageChecksReceiverAmountAndDuplicate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != verifySlipPath {
			t.Errorf("request path = %q, want %q", r.URL.Path, verifySlipPath)
		}
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			t.Errorf("Authorization header = %q, want Bearer-prefixed configured secret", r.Header.Get("Authorization"))
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Errorf("Content-Type = %q, want multipart form data", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart request: %v", err)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("get file field: %v", err)
		} else {
			defer file.Close()
			content, err := io.ReadAll(file)
			if err != nil || string(content) != "image-bytes" || header.Filename != "slip.png" {
				t.Errorf("uploaded file = %q, %q, %v", content, header.Filename, err)
			}
		}

		var conditions slipVerificationConditions
		if err := json.Unmarshal([]byte(r.FormValue("payload")), &conditions); err != nil {
			t.Errorf("decode verification conditions: %v", err)
		}
		if !conditions.CheckDuplicate || len(conditions.CheckReceiver) != 1 ||
			conditions.CheckReceiver[0].AccountType != "02001" ||
			conditions.CheckReceiver[0].AccountNumber != "0812345678" ||
			conditions.CheckAmount.Type != "eq" || conditions.CheckAmount.Amount != "12.34" {
			t.Errorf("verification conditions = %+v, unexpected receiver/amount/duplicate checks", conditions)
		}

		_, _ = io.WriteString(w, `{"code":"200000","data":{"transRef":"bank-ref-1","amount":"12.34"}}`)
	}))
	defer server.Close()

	got, err := newClient(testConfig(server.URL), server.Client()).VerifySlipImage(
		context.Background(),
		[]byte("image-bytes"),
		"image/png",
		1234,
	)
	if err != nil {
		t.Fatalf("VerifySlipImage() error = %v", err)
	}
	if !got.Valid || got.TransRef != "bank-ref-1" || got.AmountSatang != 1234 || got.ResultCode != "200000" {
		t.Errorf("VerifySlipImage() = %+v, unexpected verification result", got)
	}
}

func TestVerifySlipImageRejectsDuplicateAndInvalidData(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantErr   bool
		wantValid bool
	}{
		{name: "duplicate result", body: `{"code":"200501"}`},
		{name: "amount mismatch", body: `{"code":"200000","data":{"transRef":"bank-ref-1","amount":"12.35"}}`},
		{name: "missing transaction reference", body: `{"code":"200200","data":{"amount":"12.34"}}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, tt.body)
			}))
			defer server.Close()

			got, err := newClient(testConfig(server.URL), server.Client()).VerifySlipImage(
				context.Background(),
				[]byte("image-bytes"),
				"image/png",
				1234,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifySlipImage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.Valid != tt.wantValid {
				t.Errorf("VerifySlipImage() = %+v, Valid = %t, want %t", got, got.Valid, tt.wantValid)
			}
		})
	}
}

func TestProviderErrorDoesNotExposeResponseBody(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "provider response contains confidential data", http.StatusBadRequest)
	}))
	defer server.Close()

	_, err := newClient(testConfig(server.URL), server.Client()).GeneratePromptPayQR(context.Background(), 1234)
	if err == nil || strings.Contains(err.Error(), "confidential") || strings.Contains(err.Error(), "test-secret") {
		t.Fatalf("error = %v, want sanitized provider error", err)
	}
}

func TestGeneratePromptPayQRRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "http URL", cfg: Config{APIBaseURL: "http://example.test", SecretKey: "secret", PromptPayCode: "code", PromptPayType: "phone_number", AccountName: "name"}},
		{name: "missing secret", cfg: Config{APIBaseURL: "https://example.test", PromptPayCode: "code", PromptPayType: "phone_number", AccountName: "name"}},
		{name: "unsupported recipient type", cfg: Config{APIBaseURL: "https://example.test", SecretKey: "secret", PromptPayCode: "code", PromptPayType: "account", AccountName: "name"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newClient(tt.cfg, http.DefaultClient).GeneratePromptPayQR(context.Background(), 1234)
			if err == nil {
				t.Fatal("GeneratePromptPayQR() error = nil, want configuration error")
			}
		})
	}
}

func TestGeneratePromptPayQRHonorsContextTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := newClient(testConfig(server.URL), server.Client()).GeneratePromptPayQR(ctx, 1234)
	if err == nil {
		t.Fatal("GeneratePromptPayQR() error = nil, want context timeout error")
	}
}

func TestParseAmountSatang(t *testing.T) {
	for _, test := range []struct {
		input string
		want  int64
	}{
		{input: `"12.34"`, want: 1234},
		{input: `12.3`, want: 1230},
		{input: `100`, want: 10000},
		{input: `"100"`, want: 10000},
	} {
		got, err := parseAmountSatang(json.RawMessage(test.input))
		if err != nil || got != test.want {
			t.Errorf("parseAmountSatang(%s) = %d, %v; want %d", test.input, got, err, test.want)
		}
	}
	if _, err := parseAmountSatang(json.RawMessage(`12.345`)); err == nil {
		t.Fatal("parseAmountSatang() accepted more than two decimal places")
	}
}

func testConfig(baseURL string) Config {
	return Config{
		APIBaseURL:    baseURL,
		SecretKey:     "test-secret",
		PromptPayCode: "0812345678",
		PromptPayType: "phone_number",
		AccountName:   "Artmission",
	}
}

var _ payment.QRProvider = (*client)(nil)
