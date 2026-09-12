//go:build integration

package artworks

import (
	"context"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

var app *apptest.App

func TestArtworkFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	s3Config := apptest.StartObjectStorage(ctx, t)
	app = apptest.NewAppWithS3(t, pg.DSN, s3Config)

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty,cucumber:../../reports/json/cucumber-artworks.json",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run artwork feature tests")
	}
}
