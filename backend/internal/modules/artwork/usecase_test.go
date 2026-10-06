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
	artworks        []Artwork
	searchPage      SearchPage
	err             error
	searchQuery     SearchQuery
	categoryID      uuid.UUID
	styleIDs        []uuid.UUID
	created         *Artwork
	updated         *Artwork
	retainedSamples []Sample
	updateDeletes   []string
	deleted         struct {
		artworkID uuid.UUID
		artistID  uuid.UUID
	}
}

func (repo *fakeRepository) Search(_ context.Context, query SearchQuery) (SearchPage, error) {
	repo.searchQuery = query
	return repo.searchPage, repo.err
}

func (repo *fakeRepository) ListByArtistID(context.Context, uuid.UUID) ([]Artwork, error) {
	return repo.artworks, repo.err
}

func (repo *fakeRepository) ListAllCategories(context.Context) ([]Category, error) {
	return nil, repo.err
}

func (repo *fakeRepository) ListAllStyles(context.Context) ([]Style, error) {
	return nil, repo.err
}

func (repo *fakeRepository) Create(_ context.Context, item *Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID) error {
	repo.categoryID = categoryID
	repo.styleIDs = append([]uuid.UUID{}, styleIDs...)
	if repo.err != nil {
		return repo.err
	}
	copy := *item
	copy.Styles = append([]string{}, item.Styles...)
	copy.Samples = append([]Sample{}, item.Samples...)
	repo.created = &copy
	return nil
}

func (repo *fakeRepository) UpdateOwnedBy(_ context.Context, item *Artwork, categoryID uuid.UUID, styleIDs []uuid.UUID, deletedSampleURLs []string) ([]string, error) {
	repo.categoryID = categoryID
	repo.styleIDs = append([]uuid.UUID{}, styleIDs...)
	if repo.err != nil {
		return nil, repo.err
	}
	repo.updateDeletes = append([]string{}, deletedSampleURLs...)
	uploadedSamples := append([]Sample{}, item.Samples...)
	item.Samples = append(append([]Sample{}, repo.retainedSamples...), uploadedSamples...)
	for index := range item.Samples {
		item.Samples[index].SortOrder = index
	}
	copy := *item
	copy.Styles = append([]string{}, item.Styles...)
	copy.Samples = append([]Sample{}, item.Samples...)
	repo.updated = &copy
	return repo.updateDeletes, nil
}

func (repo *fakeRepository) DeleteOwnedBy(_ context.Context, artworkID, artistID uuid.UUID) error {
	repo.deleted.artworkID = artworkID
	repo.deleted.artistID = artistID
	return repo.err
}

type fakeStorage struct {
	uploads []string
	deletes []string
	err     error
}

func (storage *fakeStorage) Upload(_ context.Context, key string, body io.Reader, _ string) error {
	if storage.err != nil {
		return storage.err
	}
	if _, err := io.ReadAll(body); err != nil {
		return err
	}
	storage.uploads = append(storage.uploads, key)
	return nil
}

func (storage *fakeStorage) Delete(_ context.Context, key string) error {
	storage.deletes = append(storage.deletes, key)
	return nil
}

func (*fakeStorage) PublicURL(key string) string { return "https://public.test/" + key }

func (*fakeStorage) KeyFromURL(rawURL string) (string, bool) {
	const prefix = "https://public.test/"
	if !strings.HasPrefix(rawURL, prefix) {
		return "", false
	}
	key := strings.TrimPrefix(rawURL, prefix)
	if key == "" {
		return "", false
	}
	return key, true
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

func TestSearchDefaultsPageSizeAndEmptySlice(t *testing.T) {
	repo := &fakeRepository{}
	got, err := NewUsecase(repo, &fakeTransaction{}, &fakeStorage{}).Search(context.Background(), SearchQuery{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if got.Items == nil || len(got.Items) != 0 || got.Page != 1 {
		t.Errorf("Search() = %#v, want empty page 1", got)
	}
	if repo.searchQuery.Sort != DefaultSearchSort || repo.searchQuery.Limit != SearchPageSize || repo.searchQuery.Offset != 0 || repo.searchQuery.Page != 1 {
		t.Errorf("normalized query = %+v", repo.searchQuery)
	}
}

func TestSearchComputesOffsetAndResolvesProfileURL(t *testing.T) {
	key := "artist-profiles/one.webp"
	score := 4.5
	repo := &fakeRepository{searchPage: SearchPage{Items: []SearchItem{{
		Artwork: Artwork{Name: "Cover"},
		Artist:  ArtistSummary{Name: "Ada", ProfileImageKey: &key, ReviewScore: &score},
	}}, Total: 21}}
	got, err := NewUsecase(repo, &fakeTransaction{}, &fakeStorage{}).Search(context.Background(), SearchQuery{
		ArtistName: "  Ada  ",
		Category:   " Book ",
		Styles:     []string{" Pixel Art "},
		Sort:       SearchSortPriceAsc,
		Page:       2,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if repo.searchQuery.ArtistName != "Ada" || repo.searchQuery.Offset != SearchPageSize || repo.searchQuery.Limit != SearchPageSize || repo.searchQuery.Sort != SearchSortPriceAsc {
		t.Errorf("normalized query = %+v", repo.searchQuery)
	}
	if repo.searchQuery.Category != "Book" || !reflect.DeepEqual(repo.searchQuery.Styles, []string{"Pixel Art"}) {
		t.Errorf("filters = category=%v styles=%v", repo.searchQuery.Category, repo.searchQuery.Styles)
	}
	if got.Page != 2 || got.Total != 21 || got.Items[0].Artist.ProfileURL == nil || *got.Items[0].Artist.ProfileURL != "https://public.test/"+key {
		t.Errorf("Search() = %#v", got)
	}
}

func TestSearchAcceptsZeroMinReviewScore(t *testing.T) {
	zero := 0.0
	repo := &fakeRepository{}
	if _, err := NewUsecase(repo, &fakeTransaction{}, &fakeStorage{}).Search(context.Background(), SearchQuery{MinReviewScore: &zero}); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if repo.searchQuery.MinReviewScore == nil || *repo.searchQuery.MinReviewScore != 0 {
		t.Errorf("min review score = %v", repo.searchQuery.MinReviewScore)
	}
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	negative := int64(-1)
	invertedMin, invertedMax := int64(200), int64(100)
	badScore := 6.0
	negativeScore := -0.1
	tests := []SearchQuery{
		{Category: "  "},
		{Styles: []string{"  "}},
		{MinPriceSatang: &negative},
		{MaxPriceSatang: &negative},
		{MinPriceSatang: &invertedMin, MaxPriceSatang: &invertedMax},
		{MinReviewScore: &badScore},
		{MinReviewScore: &negativeScore},
		{Sort: SearchSort("popularity")},
		{Page: -1},
	}
	for index, query := range tests {
		repo := &fakeRepository{}
		if _, err := NewUsecase(repo, &fakeTransaction{}, &fakeStorage{}).Search(context.Background(), query); err == nil {
			t.Errorf("case %d: Search() error = nil", index)
		}
		if repo.searchQuery.Limit != 0 {
			t.Errorf("case %d called repository: %+v", index, repo.searchQuery)
		}
	}
}

func TestSearchReturnsRepositoryError(t *testing.T) {
	want := errors.New("repository failed")
	got, err := NewUsecase(&fakeRepository{err: want}, &fakeTransaction{}, &fakeStorage{}).Search(context.Background(), SearchQuery{})
	if !errors.Is(err, want) || !reflect.DeepEqual(got, SearchPage{}) {
		t.Errorf("Search() = (%#v, %v), want (zero, %v)", got, err, want)
	}
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
	categoryID := uuid.New()
	pixelStyleID := uuid.New()
	cartoonStyleID := uuid.New()
	created, err := NewUsecase(repo, tx, storage).Create(context.Background(), CreateInput{
		ArtistID:    artistID,
		Name:        "  Book Cover  ",
		CategoryID:  categoryID,
		StyleIDs:    []uuid.UUID{pixelStyleID, cartoonStyleID, pixelStyleID},
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
	if repo.categoryID != categoryID || !reflect.DeepEqual(repo.styleIDs, []uuid.UUID{pixelStyleID, cartoonStyleID}) {
		t.Errorf("category id=%s style ids=%v", repo.categoryID, repo.styleIDs)
	}
	if created.ID == uuid.Nil || created.ArtistID != artistID || created.Name != "Book Cover" || created.Description != "A colorful book-cover commission" {
		t.Errorf("created artwork = %+v", created)
	}
	if len(created.Samples) != 2 || !strings.HasPrefix(created.Samples[0].ImageURL, "https://public.test/artists/") || created.Samples[0].SortOrder != 0 || created.Samples[1].SortOrder != 1 || len(storage.uploads) != 2 {
		t.Errorf("samples = %+v", created.Samples)
	}
}

func TestCreateRejectsInvalidInputBeforeTransaction(t *testing.T) {
	tests := []CreateInput{
		validCreateInput(),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.Name = "  " }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.CategoryID = uuid.Nil }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.Description = "  " }),
		withCreateInput(validCreateInput(), func(input *CreateInput) { input.ArtistID = uuid.New(); input.StyleIDs = []uuid.UUID{uuid.Nil} }),
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

func TestReadSampleImageDetectsSupportedMIMETypes(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		contentType string
		extension   string
	}{
		{name: "jpeg", content: []byte{0xff, 0xd8, 0xff, 0x00}, contentType: "image/jpeg", extension: "jpg"},
		{name: "png", content: testPNG(), contentType: "image/png", extension: "png"},
		{name: "webp", content: []byte("RIFF\x00\x00\x00\x00WEBPdata"), contentType: "image/webp", extension: "webp"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content, contentType, extension, err := readSampleImage(bytes.NewReader(test.content))
			if err != nil {
				t.Fatalf("readSampleImage() error = %v", err)
			}
			if !bytes.Equal(content, test.content) || contentType != test.contentType || extension != test.extension {
				t.Fatalf("readSampleImage() = (%x, %q, %q), want (%x, %q, %q)", content, contentType, extension, test.content, test.contentType, test.extension)
			}
		})
	}
}

func TestReadSampleImageDerivesSizeLimitErrorFromMaximum(t *testing.T) {
	content := append(testPNG(), make([]byte, MaxSampleImageSize)...)
	_, _, _, err := readSampleImage(bytes.NewReader(content))
	if !errors.Is(err, ErrSampleImageTooLarge) || !strings.Contains(err.Error(), "5 MiB") {
		t.Fatalf("readSampleImage() error = %v", err)
	}
}

func TestDeleteScopesSoftDeletionToArtistAndKeepsSamples(t *testing.T) {
	repo := &fakeRepository{}
	storage := &fakeStorage{}
	artistID, artworkID := uuid.New(), uuid.New()
	if err := NewUsecase(repo, &fakeTransaction{}, storage).Delete(context.Background(), artistID, artworkID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if repo.deleted.artistID != artistID || repo.deleted.artworkID != artworkID {
		t.Errorf("delete scope = artist=%s artwork=%s", repo.deleted.artistID, repo.deleted.artworkID)
	}
	if len(storage.deletes) != 0 {
		t.Errorf("storage deletes = %v", storage.deletes)
	}
}

func TestUpdateNormalizesAndPersistsArtistOwnedArtwork(t *testing.T) {
	deletedURL := "https://public.test/artists/artist/artworks/artwork/deleted.png"
	retainedURL := "https://public.test/artists/artist/artworks/artwork/retained.png"
	repo := &fakeRepository{retainedSamples: []Sample{{ImageURL: retainedURL}}}
	tx := &fakeTransaction{}
	storage := &fakeStorage{}
	artistID, artworkID := uuid.New(), uuid.New()
	updated, err := NewUsecase(repo, tx, storage).Update(context.Background(), UpdateInput{
		ArtworkID:         artworkID,
		DeletedSampleURLs: []string{" " + deletedURL + " ", deletedURL},
		CreateInput: CreateInput{
			ArtistID:            artistID,
			Name:                "  Updated Cover  ",
			CategoryID:          uuid.New(),
			StyleIDs:            []uuid.UUID{uuid.MustParse("00000000-0000-0000-0000-0000000000c2"), uuid.MustParse("00000000-0000-0000-0000-0000000000c2")},
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
	if !reflect.DeepEqual(repo.styleIDs, []uuid.UUID{uuid.MustParse("00000000-0000-0000-0000-0000000000c2")}) || len(updated.Samples) != 2 || updated.Samples[0].ImageURL != retainedURL || !strings.HasPrefix(updated.Samples[1].ImageURL, "https://public.test/artists/") || updated.Samples[1].SortOrder != 1 {
		t.Errorf("style ids=%v samples=%+v", repo.styleIDs, updated.Samples)
	}
	if !reflect.DeepEqual(repo.updateDeletes, []string{deletedURL}) || !reflect.DeepEqual(storage.deletes, []string{"artists/artist/artworks/artwork/deleted.png"}) {
		t.Errorf("repository deletes=%v storage deletes=%v", repo.updateDeletes, storage.deletes)
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
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.CategoryID = uuid.Nil })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.Description = "  " })},
		{ArtworkID: uuid.New(), CreateInput: withCreateInput(valid, func(input *CreateInput) { input.StyleIDs = []uuid.UUID{uuid.Nil} })},
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
		CategoryID:          uuid.MustParse("00000000-0000-0000-0000-0000000000c1"),
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
