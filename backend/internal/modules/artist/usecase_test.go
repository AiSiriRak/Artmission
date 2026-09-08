package artist_test

import (
	"context"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/google/uuid"
)

type fakeRepo struct {
	byUserID map[uuid.UUID]*artist.Profile
}

func (f *fakeRepo) Create(_ context.Context, p *artist.Profile) error {
	cp := *p
	f.byUserID[p.UserID] = &cp
	return nil
}

var _ artist.ProfileRepository = (*fakeRepo)(nil)

func TestCreateProfile_PersistsDescription(t *testing.T) {
	repo := &fakeRepo{byUserID: map[uuid.UUID]*artist.Profile{}}
	usecase := artist.NewProfileUsecase(repo)
	userID := uuid.New()

	if err := usecase.CreateProfile(context.Background(), userID, "I paint portraits"); err != nil {
		t.Fatalf("CreateProfile() error = %v, want nil", err)
	}
	got, ok := repo.byUserID[userID]
	if !ok {
		t.Fatal("CreateProfile() did not persist a profile")
	}
	if got.Description != "I paint portraits" {
		t.Errorf("CreateProfile() description = %q, want %q", got.Description, "I paint portraits")
	}
	if got.UserID != userID {
		t.Errorf("CreateProfile() userID = %v, want %v", got.UserID, userID)
	}
}
