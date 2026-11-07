package payment

import (
	"math"
	"testing"
)

func TestCalculateSettlement(t *testing.T) {
	for _, outcome := range []SettlementOutcome{Completed, DeadlineExpired, ArtistCancelled, CustomerCancelled} {
		for _, amount := range []int64{0, 1, 3, 10000, 10001, math.MaxInt64} {
			refund, artist, err := CalculateSettlement(outcome, amount)
			if err != nil || refund < 0 || artist < 0 || refund+artist != amount {
				t.Fatalf("%s %d: refund=%d artist=%d err=%v", outcome, amount, refund, artist, err)
			}
			var expected int64
			switch outcome {
			case DeadlineExpired, ArtistCancelled:
				expected = amount
			case CustomerCancelled:
				expected = amount / 2
			}
			if refund != expected {
				t.Fatalf("%s %d: refund=%d want %d", outcome, amount, refund, expected)
			}
		}
		if _, _, err := CalculateSettlement(outcome, -1); err == nil {
			t.Fatal("accepted negative amount")
		}
	}
	if _, _, err := CalculateSettlement("UNKNOWN", 100); err == nil {
		t.Fatal("accepted unknown outcome")
	}
}
