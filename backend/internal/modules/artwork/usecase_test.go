package artwork

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	artworks       []Artwork
	err            error
	categoryLabels []string
	styleLabels    [][]string
	created        *Artwork
	updated        *Artwork
	replacedURLs   []string
	deletedURLs    []string
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

func (repo *fakeRepository) UpdateOwnedBy(_ context.Context, item *Artwork, _ uuid.UUID, _ []uuid.UUID) ([]string, error) {
	if repo.err != nil {
		return nil, repo.err
	}
	copy := *item
	copy.Styles = append([]string{}, item.Styles...)
	copy.Samples = append([]Sample{}, item.Samples...)
	repo.updated = &copy
	return repo.replacedURLs, nil
}

func (repo *fakeRepository) DeleteOwnedBy(_ context.Context, artworkID, artistID uuid.UUID) ([]string, error) {
	repo.deleted.artworkID = artworkID
	repo.deleted.artistID = artistID
	return repo.deletedURLs, repo.err
}

type fakeStorage struct {
	uploads []string
	deletes []string
	err     error
}

func (storage *fakeStorage) UploadPublic(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	if storage.err != nil {
		return storage.err
	}
	if _, err := io.ReadAll(body); err != nil {
		return err
	}
	storage.uploads = append(storage.uploads, key)
	return nil
}

func (storage *fakeStorage) DeletePublicURL(_ context.Context, imageURL string) error {
	storage.deletes = append(storage.deletes, imageURL)
	return nil
}

func (*fakeStorage) PublicURL(key string) string { return "https://public.test/" + key }

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
	usecase := NewUsecase(repo, &fakeTransaction{}, &fakeStorage{})

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
	got, err := NewUsecase(&fakeRepository{}, &fakeTransaction{}, &fakeStorage{}).ListByArtistID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListByArtistID() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByArtistID() = %#v, want empty non-nil slice", got)
	}
}

func TestListByArtistIDReturnsRepositoryError(t *testing.T) {
	want := errors.New("repository failed")
	got, err := NewUsecase(&fakeRepository{err: want}, &fakeTransaction{}, &fakeStorage{}).ListByArtistID(context.Background(), uuid.New())
	if !errors.Is(err, want) || !reflect.DeepEqual(got, []Artwork(nil)) {
		t.Errorf("ListByArtistID() = (%#v, %v), want (nil, %v)", got, err, want)
	}
}

func TestCreateNormalizesAndPersistsArtistOwnedArtwork(t *testing.T) {
	repo := &fakeRepository{}
	tx := &fakeTransaction{}
	storage := &fakeStorage{}
	artistID := uuid.New()
	created, err := NewUsecase(repo, tx, storage).Create(context.Background(), CreateInput{
		ArtistID:    artistID,
		Name:        "  Book Cover  ",
		Category:    "  Book  ",
		Styles:      []string{" Pixel Art ", "Cartoon", "Pixel Art"},
		Description: "  A colorful book-cover commission  ",
		SampleFiles: []io.Reader{
			bytes.NewReader(testPNG()),
			bytes.NewReader(testPNG()),
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
	if len(created.Samples) != 2 || !strings.HasPrefix(created.Samples[0].ImageURL, "https://public.test/artworks/") || created.Samples[0].SortOrder != 0 || created.Samples[1].SortOrder != 1 || len(storage.uploads) != 2 {
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
			input.SampleFiles = []io.Reader{strings.NewReader("not an image")}
		}),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.PriceSatang = -1 }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.MinimumDeadlineDays = 0 }),
	}
	for index, input := range tests {
		repo := &fakeRepository{}
		tx := &fakeTransaction{}
		if _, err := NewUsecase(repo, tx, &fakeStorage{}).Create(context.Background(), input); err == nil {
			t.Errorf("case %d: Create() error = nil", index)
		}
		if tx.calls != 0 || repo.created != nil {
			t.Errorf("case %d changed state: calls=%d created=%+v", index, tx.calls, repo.created)
		}
	}
}

func TestDeleteScopesDeletionToArtist(t *testing.T) {
	repo := &fakeRepository{deletedURLs: []string{"https://public.test/artworks/deleted.png"}}
	storage := &fakeStorage{}
	artistID, artworkID := uuid.New(), uuid.New()
	if err := NewUsecase(repo, &fakeTransaction{}, storage).Delete(context.Background(), artistID, artworkID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if repo.deleted.artistID != artistID || repo.deleted.artworkID != artworkID {
		t.Errorf("delete scope = artist=%s artwork=%s", repo.deleted.artistID, repo.deleted.artworkID)
	}
	if !reflect.DeepEqual(storage.deletes, repo.deletedURLs) {
		t.Errorf("storage deletes = %v, want %v", storage.deletes, repo.deletedURLs)
	}
}

func TestUpdateNormalizesAndPersistsArtistOwnedArtwork(t *testing.T) {
	repo := &fakeRepository{replacedURLs: []string{"https://public.test/artworks/replaced.png"}}
	tx := &fakeTransaction{}
	storage := &fakeStorage{}
	artistID, artworkID := uuid.New(), uuid.New()
	updated, err := NewUsecase(repo, tx, storage).Update(context.Background(), UpdateInput{
		ArtworkID: artworkID,
		CreateInput: CreateInput{
			ArtistID:            artistID,
			Name:                "  Updated Cover  ",
			Category:            "  Book  ",
			Styles:              []string{" Cartoon ", "Cartoon"},
			Description:         "  Updated description  ",
			SampleFiles:         []io.Reader{bytes.NewReader(testPNG())},
			MinimumDeadlineDays: 10,
			PriceSatang:         300000,
		},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if tx.calls != 1 || repo.updated == nil {
		t.Fatalf("transaction calls=%d updated=%+v", tx.calls, repo.updated)
	}
	if updated.ID != artworkID || updated.ArtistID != artistID || updated.Name != "Updated Cover" || updated.Description != "Updated description" {
		t.Errorf("updated artwork = %+v", updated)
	}
	if !reflect.DeepEqual(updated.Styles, []string{"Cartoon"}) || len(updated.Samples) != 1 || !strings.HasPrefix(updated.Samples[0].ImageURL, "https://public.test/artworks/") || updated.Samples[0].SortOrder != 0 {
		t.Errorf("styles=%#v samples=%+v", updated.Styles, updated.Samples)
	}
	if !reflect.DeepEqual(storage.deletes, repo.replacedURLs) {
		t.Errorf("storage deletes = %v, want %v", storage.deletes, repo.replacedURLs)
	}
}

func TestCreateDeletesUploadedSamplesWhenPersistenceFails(t *testing.T) {
	repo := &fakeRepository{err: errors.New("database unavailable")}
	storage := &fakeStorage{}
	input := validCreateInput()
	input.ArtistID = uuid.New()
	input.SampleFiles = []io.Reader{bytes.NewReader(testPNG())}

	if _, err := NewUsecase(repo, &fakeTransaction{}, storage).Create(context.Background(), input); err == nil {
		t.Fatal("Create() error = nil")
	}
	if len(storage.uploads) != 1 || len(storage.deletes) != 1 {
		t.Fatalf("uploads=%v deletes=%v, want one compensated upload", storage.uploads, storage.deletes)
	}
}

func TestUpdateRejectsInvalidInputBeforeTransaction(t *testing.T) {
	valid := validCreateInput()
	valid.ArtistID = uuid.New()
	tests := []UpdateInput{
		{CreateInput: valid},
		{ArtworkID: uuid.New(), CreateInput: validCreateInput()},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.Name = "  " })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.Category = "  " })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.Description = "  " })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.Styles = []string{"  "} })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.SampleFiles = []io.Reader{strings.NewReader("not an image")} })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.PriceSatang = -1 })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.MinimumDeadlineDays = 0 })},
	}
	for index, input := range tests {
		repo := &fakeRepository{}
		tx := &fakeTransaction{}
		if _, err := NewUsecase(repo, tx, &fakeStorage{}).Update(context.Background(), input); err == nil {
			t.Errorf("case %d: Update() error = nil", index)
		}
		if tx.calls != 0 || repo.updated != nil {
			t.Errorf("case %d changed state: calls=%d updated=%+v", index, tx.calls, repo.updated)
		}
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

func testPNG() []byte {
	return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00}
}
