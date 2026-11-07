package slip2go

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/payment"
)

const (
	generateQRPath  = "/qr-payment/generate-qr-code"
	verifySlipPath  = "/verify-slip/qr-image/info"
	maxResponseSize = 1 << 20
)

type Config struct {
	APIBaseURL    string
	SecretKey     string
	PromptPayCode string
	PromptPayType string
	AccountName   string
}

type client struct {
	config     Config
	httpClient *http.Client
}

var _ payment.QRProvider = (*client)(nil)

func NewClient(cfg Config) payment.QRProvider {
	return newClient(cfg, &http.Client{
		Timeout: 10 * time.Second,
		// Never follow redirects: the bearer token and slip image must not be forwarded to another URL.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})
}

func newClient(cfg Config, httpClient *http.Client) *client {
	return &client{config: cfg, httpClient: httpClient}
}

type generateQRRequest struct {
	PromptPayCode string `json:"promptPayCode"`
	PromptPayType string `json:"promptPayType"`
	AccountName   string `json:"accountName"`
	Amount        string `json:"amount"`
}

type generateQRResponse struct {
	Code string `json:"code"`
	Data struct {
		QRCode      string `json:"qrCode"`
		AccountName string `json:"accountName"`
		Amount      string `json:"amount"`
	} `json:"data"`
}

type slipVerificationConditions struct {
	CheckDuplicate bool                `json:"checkDuplicate"`
	CheckReceiver  []receiverCondition `json:"checkReceiver"`
	CheckAmount    amountCondition     `json:"checkAmount"`
	CheckDate      dateCondition       `json:"checkDate"`
}

type dateCondition struct {
	Type string `json:"type"`
	Date string `json:"date"`
}

type receiverCondition struct {
	AccountType   string `json:"accountType"`
	AccountNumber string `json:"accountNumber"`
}

type amountCondition struct {
	Type   string `json:"type"`
	Amount string `json:"amount"`
}

type verifySlipResponse struct {
	Code string `json:"code"`
	Data struct {
		TransRef string          `json:"transRef"`
		Amount   json.RawMessage `json:"amount"`
		DateTime string          `json:"dateTime"`
	} `json:"data"`
}

func (c *client) GeneratePromptPayQR(
	ctx context.Context,
	amountSatang int64,
) (payment.QRImage, error) {
	if err := c.validateConfig(); err != nil {
		return payment.QRImage{}, err
	}
	if amountSatang < 0 {
		return payment.QRImage{}, fmt.Errorf("payment amount must not be negative")
	}

	requestBody, err := json.Marshal(generateQRRequest{
		PromptPayCode: c.config.PromptPayCode,
		PromptPayType: c.config.PromptPayType,
		AccountName:   c.config.AccountName,
		Amount:        formatSatang(amountSatang),
	})
	if err != nil {
		return payment.QRImage{}, fmt.Errorf("encode Slip2Go QR request: %w", err)
	}

	resp, err := c.doJSON(ctx, generateQRPath, requestBody)
	if err != nil {
		return payment.QRImage{}, err
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp.Body)
	if err != nil {
		return payment.QRImage{}, err
	}
	var result generateQRResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return payment.QRImage{}, fmt.Errorf("decode Slip2Go QR response: %w", err)
	}
	if result.Code != "200" {
		return payment.QRImage{}, fmt.Errorf("Slip2Go rejected QR request with code %q", result.Code)
	}
	if strings.TrimSpace(result.Data.QRCode) == "" ||
		result.Data.AccountName != c.config.AccountName ||
		result.Data.Amount != formatSatang(amountSatang) {
		return payment.QRImage{}, fmt.Errorf("Slip2Go returned mismatched QR data")
	}

	return payment.QRImage{
		Payload:      result.Data.QRCode,
		AccountName:  result.Data.AccountName,
		AmountSatang: amountSatang,
	}, nil
}

func (c *client) VerifySlipImage(
	ctx context.Context,
	image []byte,
	contentType string,
	amountSatang int64,
	notBefore time.Time,
) (payment.SlipVerification, error) {
	if err := c.validateConfig(); err != nil {
		return payment.SlipVerification{}, err
	}
	if len(image) == 0 || !isImageContentType(contentType) || amountSatang < 0 {
		return payment.SlipVerification{}, fmt.Errorf("invalid Slip2Go verification input")
	}

	accountType, ok := verificationAccountType(c.config.PromptPayType)
	if !ok {
		return payment.SlipVerification{}, fmt.Errorf("Slip2Go PromptPay type cannot be verified")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="slip%s"`, imageExtension(contentType)))
	partHeader.Set("Content-Type", contentType)
	file, err := writer.CreatePart(partHeader)
	if err != nil {
		return payment.SlipVerification{}, fmt.Errorf("create Slip2Go image field: %w", err)
	}
	if _, err := file.Write(image); err != nil {
		return payment.SlipVerification{}, fmt.Errorf("write Slip2Go image field: %w", err)
	}

	conditions, err := json.Marshal(slipVerificationConditions{
		CheckDuplicate: true,
		CheckReceiver: []receiverCondition{{
			AccountType:   accountType,
			AccountNumber: c.config.PromptPayCode,
		}},
		CheckAmount: amountCondition{
			Type:   "eq",
			Amount: formatSatang(amountSatang),
		},
		// Reject slips made before the payment existed.
		CheckDate: dateCondition{
			Type: "gte",
			Date: notBefore.UTC().Format("2006-01-02T15:04:05.000Z"),
		},
	})
	if err != nil {
		return payment.SlipVerification{}, fmt.Errorf("encode Slip2Go verification conditions: %w", err)
	}
	if err := writer.WriteField("payload", string(conditions)); err != nil {
		return payment.SlipVerification{}, fmt.Errorf("write Slip2Go verification conditions: %w", err)
	}
	if err := writer.Close(); err != nil {
		return payment.SlipVerification{}, fmt.Errorf("finish Slip2Go verification request: %w", err)
	}

	endpoint, err := c.endpoint(verifySlipPath)
	if err != nil {
		return payment.SlipVerification{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body.Bytes()))
	if err != nil {
		return payment.SlipVerification{}, fmt.Errorf("create Slip2Go verification request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.SecretKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "Slip2Go verification request failed", "error", err)
		return payment.SlipVerification{}, fmt.Errorf("send Slip2Go verification request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := readResponseBody(resp.Body)
	if err != nil {
		return payment.SlipVerification{}, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var failure struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(responseBody, &failure)
		slog.WarnContext(ctx, "Slip2Go verification returned an error status",
			"status", resp.StatusCode, "provider_code", failure.Code)
		if (resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound) &&
			failure.Code != "" {
			return payment.SlipVerification{ResultCode: failure.Code}, nil
		}
		return payment.SlipVerification{}, fmt.Errorf(
			"Slip2Go returned HTTP status %d (code %q: %s)",
			resp.StatusCode, failure.Code, failure.Message,
		)
	}

	var result verifySlipResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return payment.SlipVerification{}, fmt.Errorf("decode Slip2Go verification response: %w", err)
	}
	verification := payment.SlipVerification{ResultCode: result.Code}
	if result.Code != "200000" && result.Code != "200200" {
		return verification, nil
	}
	amount, err := parseAmountSatang(result.Data.Amount)
	if err != nil {
		return payment.SlipVerification{}, fmt.Errorf("Slip2Go verification response has invalid amount: %w", err)
	}
	if strings.TrimSpace(result.Data.TransRef) == "" {
		return payment.SlipVerification{}, fmt.Errorf("Slip2Go verification response is missing transaction reference")
	}
	dateTime, err := parseSlipDateTime(result.Data.DateTime)
	if err != nil {
		return payment.SlipVerification{}, fmt.Errorf("Slip2Go verification response has invalid dateTime: %w", err)
	}

	verification.Valid = amount == amountSatang
	verification.TransRef = result.Data.TransRef
	verification.AmountSatang = amount
	verification.DateTime = dateTime
	return verification, nil
}

func (c *client) doJSON(ctx context.Context, path string, body []byte) (*http.Response, error) {
	endpoint, err := c.endpoint(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create Slip2Go request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send Slip2Go request: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var failure struct {
			Code string `json:"code"`
		}
		if responseBody, err := readResponseBody(resp.Body); err == nil {
			_ = json.Unmarshal(responseBody, &failure)
		}
		resp.Body.Close()
		slog.WarnContext(ctx, "Slip2Go request returned an error status",
			"path", path, "status", resp.StatusCode, "provider_code", failure.Code)
		return nil, fmt.Errorf("Slip2Go returned HTTP status %d", resp.StatusCode)
	}
	return resp, nil
}

func (c *client) endpoint(path string) (string, error) {
	if err := c.validateConfig(); err != nil {
		return "", err
	}
	return strings.TrimRight(c.config.APIBaseURL, "/") + path, nil
}

func readResponseBody(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Slip2Go response: %w", err)
	}
	if len(data) > maxResponseSize {
		return nil, fmt.Errorf("Slip2Go response exceeded the size limit")
	}
	return data, nil
}

func verificationAccountType(promptPayType string) (string, bool) {
	switch promptPayType {
	case "phone_number":
		return "02001", true
	case "citizen_id":
		return "02003", true
	case "e_wallet":
		return "02004", true
	default:
		return "", false
	}
}

func isImageContentType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func imageExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func parseAmountSatang(value json.RawMessage) (int64, error) {
	if len(value) == 0 || string(value) == "null" {
		return 0, fmt.Errorf("amount is missing")
	}
	var amount string
	if value[0] == '"' {
		if err := json.Unmarshal(value, &amount); err != nil {
			return 0, err
		}
	} else {
		amount = string(value)
	}

	parts := strings.Split(amount, ".")
	if len(parts) == 1 {
		parts = append(parts, "0")
	}
	if len(parts) != 2 || parts[0] == "" || len(parts[1]) < 1 || len(parts[1]) > 2 {
		return 0, fmt.Errorf("amount must be a decimal baht value")
	}
	baht, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || baht < 0 {
		return 0, fmt.Errorf("amount has an invalid baht component")
	}
	satangPart := parts[1]
	if len(satangPart) == 1 {
		satangPart += "0"
	}
	satang, err := strconv.ParseInt(satangPart, 10, 64)
	if err != nil || satang < 0 {
		return 0, fmt.Errorf("amount has an invalid satang component")
	}
	return baht*100 + satang, nil
}

func parseSlipDateTime(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, fmt.Errorf("dateTime is missing")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func formatSatang(amount int64) string {
	return fmt.Sprintf("%d.%02d", amount/100, amount%100)
}

func (c *client) validateConfig() error {
	parsed, err := url.Parse(c.config.APIBaseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("Slip2Go API base URL must be an absolute HTTPS URL")
	}
	if c.config.SecretKey == "" ||
		c.config.PromptPayCode == "" ||
		c.config.AccountName == "" {
		return fmt.Errorf("Slip2Go credentials and PromptPay recipient configuration are required")
	}
	switch c.config.PromptPayType {
	case "phone_number", "citizen_id", "e_wallet":
		return nil
	default:
		return fmt.Errorf("Slip2Go PromptPay type is invalid")
	}
}
