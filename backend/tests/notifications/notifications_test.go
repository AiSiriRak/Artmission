//go:build integration

package notifications

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/objectstorage"
	"github.com/AiSiriRak/Artmission/backend/tests/internal/apptest"
	"github.com/cucumber/godog"
)

var app *apptest.App
var storage *objectstorage.Client

func TestNotificationsFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	s3Config := apptest.StartObjectStorage(ctx, t)
	app = apptest.NewAppWithS3(t, pg.DSN, s3Config)
	var err error
	storage, err = objectstorage.NewS3Client(ctx, s3Config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("../../reports/json", 0755); err != nil {
		t.Fatal(err)
	}
	suite := godog.TestSuite{ScenarioInitializer: InitializeScenario, Options: &godog.Options{
		Format: "pretty,cucumber:../../reports/json/cucumber-notifications.json", Paths: []string{"features"}, TestingT: t,
	}}
	if suite.Run() != 0 {
		t.Fatal("notification feature scenarios failed")
	}
}
