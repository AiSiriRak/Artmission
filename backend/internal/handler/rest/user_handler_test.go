package rest

import "testing"

func TestMaskAccountNumber(t *testing.T) {
	tests := []struct {
		name          string
		accountNumber string
		want          string
	}{
		{name: "shows only the final four characters", accountNumber: "1234567890", want: "••••7890"},
		{name: "does not reveal a four character account number", accountNumber: "1234", want: "••••"},
		{name: "does not reveal a shorter account number", accountNumber: "123", want: "••••"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskAccountNumber(tt.accountNumber); got != tt.want {
				t.Errorf("maskAccountNumber(%q) = %q, want %q", tt.accountNumber, got, tt.want)
			}
		})
	}
}
