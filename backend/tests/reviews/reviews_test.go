//go:build integration

package reviews

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

var app *apptest.App

func TestReviewsFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	// These endpoints do not upload files or check storage readiness.
	app = apptest.NewApp(t, pg.DSN)

	if err := os.MkdirAll("../../reports/json", 0755); err != nil {
		t.Fatal(err)
	}
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty,cucumber:../../reports/json/cucumber-reviews.json",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("review feature tests failed")
	}
}
