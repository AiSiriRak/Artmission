//go:build integration

// Package orders contains the BDD suite for the orders domain
// (view hiring history).
package orders

import (
	"context"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

// app is shared across every scenario in this package — one container +
// one app per test binary run, not per scenario.
var app *apptest.App

func TestOrdersFeatures(t *testing.T) {
	// Generous: covers a cold Docker image pull, not just container start —
	// on a warm cache this returns in seconds regardless.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	app = apptest.NewApp(t, pg.DSN)

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run orders feature tests")
	}
}
