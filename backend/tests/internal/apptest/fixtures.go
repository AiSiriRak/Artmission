//go:build integration

package apptest

import (
	"strconv"
	"sync/atomic"
)

var suffixCounter atomic.Uint64

// UniqueSuffix returns a short, deterministically unique token for building
// scenario-scoped usernames/emails.
//
// Isolation strategy for this suite is unique data, not truncate-between-scenarios:
// every scenario creates its own user(s), so scenarios never observe each
// other's rows and can run against one shared, migrated-once container for the
// whole test binary.
func UniqueSuffix() string {
	return strconv.FormatUint(suffixCounter.Add(1), 10)
}
