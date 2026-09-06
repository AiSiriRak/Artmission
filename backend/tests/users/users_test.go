//go:build integration

package users

import (
	"context"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

var app *apptest.App

func TestUsersFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	app = apptest.NewApp(t, pg.DSN)

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty,cucumber:../../reports/json/cucumber-users.json",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run users feature tests")
	}
}
