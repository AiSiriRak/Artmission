package artwork

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	artworks []Artwork
	err      error
}

func (repo *fakeRepository) ListByArtistID(context.Context, uuid.UUID) ([]Artwork, error) {
	return repo.artworks, repo.err
}

func TestListByArtistIDPreservesSampleURLsAndNormalizesSlices(t *testing.T) {
	repo := &fakeRepository{artworks: []Artwork{
		{ID: uuid.New(), Samples: []Sample{{ImageURL: "https://example.com/one.webp"}}},
		{ID: uuid.New()},
	}}
	usecase := NewUsecase(repo)

	got, err := usecase.ListByArtistID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListByArtistID() error = %v", err)
	}
	if got[0].Samples[0].ImageURL != "https://example.com/one.webp" {
		t.Errorf("ImageURL = %q", got[0].Samples[0].ImageURL)
	}
	if got[0].Styles == nil || got[1].Styles == nil || got[1].Samples == nil {
		t.Errorf("expected non-nil response slices: %+v", got)
	}
}

func TestListByArtistIDReturnsEmptySlice(t *testing.T) {
	got, err := NewUsecase(&fakeRepository{}).ListByArtistID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListByArtistID() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByArtistID() = %#v, want empty non-nil slice", got)
	}
}

func TestListByArtistIDReturnsRepositoryError(t *testing.T) {
	want := errors.New("repository failed")
	got, err := NewUsecase(&fakeRepository{err: want}).ListByArtistID(context.Background(), uuid.New())
	if !errors.Is(err, want) || !reflect.DeepEqual(got, []Artwork(nil)) {
		t.Errorf("ListByArtistID() = (%#v, %v), want (nil, %v)", got, err, want)
	}
}
