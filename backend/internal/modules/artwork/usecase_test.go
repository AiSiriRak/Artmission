package artwork

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	artworks       []Artwork
	err            error
	categoryLabels []string
	styleLabels    [][]string
	created        *Artwork
	deleted        struct {
		artworkID uuid.UUID
		artistID  uuid.UUID
	}
}

func (repo *fakeRepository) ListByArtistID(context.Context, uuid.UUID) ([]Artwork, error) {
	return repo.artworks, repo.err
}

func (repo *fakeRepository) FindOrCreateCategory(_ context.Context, label string) (uuid.UUID, error) {
	repo.categoryLabels = append(repo.categoryLabels, label)
	return uuid.New(), repo.err
}

func (repo *fakeRepository) FindOrCreateStyles(_ context.Context, labels []string) ([]uuid.UUID, error) {
	repo.styleLabels = append(repo.styleLabels, append([]string{}, labels...))
	ids := make([]uuid.UUID, len(labels))
	for index := range ids {
		ids[index] = uuid.New()
	}
	return ids, repo.err
}

func (repo *fakeRepository) Create(_ context.Context, item *Artwork, _ uuid.UUID, _ []uuid.UUID) error {
	if repo.err != nil {
		return repo.err
	}
	copy := *item
	copy.Styles = append([]string{}, item.Styles...)
	copy.Samples = append([]Sample{}, item.Samples...)
	repo.created = &copy
	return nil
}

func (repo *fakeRepository) DeleteOwnedBy(_ context.Context, artworkID, artistID uuid.UUID) error {
	repo.deleted.artworkID = artworkID
	repo.deleted.artistID = artistID
	return repo.err
}

type fakeTransaction struct {
	calls int
	err   error
}

func (tx *fakeTransaction) Transaction(ctx context.Context, fn func(context.Context) error) error {
	tx.calls++
	if tx.err != nil {
		return tx.err
	}
	return fn(ctx)
}

func TestListByArtistIDPreservesSampleURLsAndNormalizesSlices(t *testing.T) {
	repo := &fakeRepository{artworks: []Artwork{
		{ID: uuid.New(), Samples: []Sample{{ImageURL: "https://example.com/one.webp"}}},
		{ID: uuid.New()},
	}}
	usecase := NewUsecase(repo, &fakeTransaction{})

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
	got, err := NewUsecase(&fakeRepository{}, &fakeTransaction{}).ListByArtistID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListByArtistID() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByArtistID() = %#v, want empty non-nil slice", got)
	}
}

func TestListByArtistIDReturnsRepositoryError(t *testing.T) {
	want := errors.New("repository failed")
	got, err := NewUsecase(&fakeRepository{err: want}, &fakeTransaction{}).ListByArtistID(context.Background(), uuid.New())
	if !errors.Is(err, want) || !reflect.DeepEqual(got, []Artwork(nil)) {
		t.Errorf("ListByArtistID() = (%#v, %v), want (nil, %v)", got, err, want)
	}
}

func TestCreateNormalizesAndPersistsArtistOwnedArtwork(t *testing.T) {
	repo := &fakeRepository{}
	tx := &fakeTransaction{}
	artistID := uuid.New()
	created, err := NewUsecase(repo, tx).Create(context.Background(), CreateInput{
		ArtistID:    artistID,
		Name:        "  Book Cover  ",
		Category:    "  Book  ",
		Styles:      []string{" Pixel Art ", "Cartoon", "Pixel Art"},
		Description: "  A colorful book-cover commission  ",
		Samples: []Sample{
			{ImageURL: " https://example.com/one.png "},
			{ImageURL: "https://example.com/two.png"},
		},
		MinimumDeadlineDays: 7,
		PriceSatang:         250000,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if tx.calls != 1 || repo.created == nil {
		t.Fatalf("transaction calls=%d created=%+v", tx.calls, repo.created)
	}
	if !reflect.DeepEqual(repo.categoryLabels, []string{"Book"}) || !reflect.DeepEqual(repo.styleLabels, [][]string{{"Pixel Art", "Cartoon"}}) {
		t.Errorf("category labels=%#v style labels=%#v", repo.categoryLabels, repo.styleLabels)
	}
	if created.ID == uuid.Nil || created.ArtistID != artistID || created.Name != "Book Cover" || created.Description != "A colorful book-cover commission" {
		t.Errorf("created artwork = %+v", created)
	}
	if len(created.Samples) != 2 || created.Samples[0].ImageURL != "https://example.com/one.png" || created.Samples[0].SortOrder != 0 || created.Samples[1].SortOrder != 1 {
		t.Errorf("samples = %+v", created.Samples)
	}
}

func TestCreateRejectsInvalidInputBeforeTransaction(t *testing.T) {
	tests := []CreateInput{
		validCreateInput(),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.Name = "  " }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.Category = "  " }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.Description = "  " }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.Styles = []string{"  "} }),
		withCreateInput(validCreateInput(), func(input *CreateInput) {
			input.ArtistID = uuid.New()
			input.Samples = []Sample{{ImageURL: "not a URL"}}
		}),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.PriceSatang = -1 }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.MinimumDeadlineDays = 0 }),
	}
	for index, input := range tests {
		repo := &fakeRepository{}
		tx := &fakeTransaction{}
		if _, err := NewUsecase(repo, tx).Create(context.Background(), input); err == nil {
			t.Errorf("case %d: Create() error = nil", index)
		}
		if tx.calls != 0 || repo.created != nil {
			t.Errorf("case %d changed state: calls=%d created=%+v", index, tx.calls, repo.created)
		}
	}
}

func TestDeleteScopesDeletionToArtist(t *testing.T) {
	repo := &fakeRepository{}
	artistID, artworkID := uuid.New(), uuid.New()
	if err := NewUsecase(repo, &fakeTransaction{}).Delete(context.Background(), artistID, artworkID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if repo.deleted.artistID != artistID || repo.deleted.artworkID != artworkID {
		t.Errorf("delete scope = artist=%s artwork=%s", repo.deleted.artistID, repo.deleted.artworkID)
	}
}

func validCreateInput() CreateInput {
	return CreateInput{
		Name:                "Book Cover",
		Category:            "Book",
		Description:         "A colorful book-cover commission",
		MinimumDeadlineDays: 7,
		PriceSatang:         250000,
	}
}

func withCreateInput(input CreateInput, update func(*CreateInput)) CreateInput {
	update(&input)
	return input
}
