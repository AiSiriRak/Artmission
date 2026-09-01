//go:build integration

// Package auth contains the BDD suite for the auth domain
// (register/login/refresh/logout).
package auth

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

func TestAuthFeatures(t *testing.T) {
	// Generous: covers a cold Docker image pull, not just container start —
	// on a warm cache this returns in seconds regardless.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	app = apptest.NewApp(t, pg.DSN)

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty,cucumber:../../reports/cucumber-auth.json",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run auth feature tests")
	}
}
