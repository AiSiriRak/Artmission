package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigReadsSlip2GoFromEnvFile(t *testing.T) {
	envContent := `
		SLIP2GO_API_BASE_URL=https://provider.example
		SLIP2GO_SECRET_KEY=test-secret
		SLIP2GO_PROMPTPAY_CODE=0812345678
		SLIP2GO_PROMPTPAY_TYPE=phone_number
		SLIP2GO_ACCOUNT_NAME=Artmission
		`

	envFile := filepath.Join(t.TempDir(), ".env")

	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		t.Fatalf("failed to create env file: %v", err)
	}

	cfg, err := InitConfig(envFile)
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	got := cfg.Slip2Go()

	if got.APIBaseURL != "https://provider.example" ||
		got.SecretKey != "test-secret" ||
		got.PromptPayCode != "0812345678" ||
		got.PromptPayType != "phone_number" ||
		got.AccountName != "Artmission" {
		t.Fatalf("Slip2Go() = %+v, env file values were not loaded", got)
	}
}
