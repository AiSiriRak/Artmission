package artist_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/google/uuid"
)

type fakeRepo struct {
	byUserID   map[uuid.UUID]*artist.Profile
	styles     map[uuid.UUID]bool
	getErr     error
	updateErr  error
	countErr   error
	replaceErr error
	replaced   []uuid.UUID
}

func (f *fakeRepo) Create(_ context.Context, p *artist.Profile) error {
	f.byUserID[p.UserID] = cloneProfile(p)
	return nil
}

func (f *fakeRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*artist.Profile, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	p, ok := f.byUserID[userID]
	if !ok {
		return nil, artist.ErrProfileNotFound
	}
	return cloneProfile(p), nil
}

func (f *fakeRepo) UpdateByUserID(_ context.Context, userID uuid.UUID, in artist.ProfileUpdate) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	p, ok := f.byUserID[userID]
	if !ok {
		return artist.ErrProfileNotFound
	}
	p.Description = in.Description
	p.MinPriceSatang = int64Pointer(in.MinPriceSatang)
	p.MaxPriceSatang = int64Pointer(in.MaxPriceSatang)
	p.UpdatedAt = in.UpdatedAt
	return nil
}

func (f *fakeRepo) CountStylesByIDs(_ context.Context, styleIDs []uuid.UUID) (int, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	count := 0
	for _, styleID := range styleIDs {
		if f.styles[styleID] {
			count++
		}
	}
	return count, nil
}

func (f *fakeRepo) ReplaceStyles(_ context.Context, userID uuid.UUID, styleIDs []uuid.UUID) error {
	if f.replaceErr != nil {
		return f.replaceErr
	}
	f.replaced = append([]uuid.UUID(nil), styleIDs...)
	p := f.byUserID[userID]
	p.Styles = make([]artist.Style, len(styleIDs))
	for i, styleID := range styleIDs {
		p.Styles[i] = artist.Style{ID: styleID}
	}
	return nil
}

type fakeTransactioner struct {
	repo  *fakeRepo
	calls int
}

func (t *fakeTransactioner) Transaction(ctx context.Context, fn func(context.Context) error) error {
	t.calls++
	snapshot := make(map[uuid.UUID]*artist.Profile, len(t.repo.byUserID))
	for id, profile := range t.repo.byUserID {
		snapshot[id] = cloneProfile(profile)
	}
	err := fn(ctx)
	if err != nil {
		t.repo.byUserID = snapshot
	}
	return err
}

var (
	_ artist.ProfileRepository = (*fakeRepo)(nil)
	_ artist.Transactioner     = (*fakeTransactioner)(nil)
)

func TestCreateProfile_PersistsDescription(t *testing.T) {
	repo := newFakeRepo()
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})
	userID := uuid.New()
	description := "I paint portraits"

	if err := usecase.CreateProfile(context.Background(), userID, &description); err != nil {
		t.Fatalf("CreateProfile() error = %v, want nil", err)
	}
	got := repo.byUserID[userID]
	if got == nil || got.Description == nil || *got.Description != "I paint portraits" || got.UserID != userID {
		t.Fatalf("CreateProfile() persisted %+v", got)
	}
}

func TestCreateProfile_AllowsNullDescription(t *testing.T) {
	repo := newFakeRepo()
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})
	userID := uuid.New()

	if err := usecase.CreateProfile(context.Background(), userID, nil); err != nil {
		t.Fatalf("CreateProfile() error = %v, want nil", err)
	}
	if got := repo.byUserID[userID]; got == nil || got.Description != nil {
		t.Fatalf("CreateProfile() = %+v, want null description", got)
	}
}

func TestGetProfile_ReturnsRepositoryProfile(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID, ArtistName: "Mali", Description: stringPointer("Portraits")}
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})

	got, err := usecase.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if got.ArtistName != "Mali" || got.Description == nil || *got.Description != "Portraits" {
		t.Fatalf("GetProfile() = %+v", got)
	}
}

func TestUpdateProfile_NormalizesAndPersistsAllEditableFields(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	styleA, styleB := uuid.New(), uuid.New()
	repo.styles[styleA], repo.styles[styleB] = true, true
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old")}
	tx := &fakeTransactioner{repo: repo}
	usecase := artist.NewProfileUsecase(repo, tx)

	got, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{
		Description:    stringPointer("  new description  "),
		StyleIDs:       []uuid.UUID{styleB, styleA, styleB},
		MinPriceSatang: 10_000,
		MaxPriceSatang: 50_000,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if tx.calls != 1 || got.Description == nil || *got.Description != "new description" || *got.MinPriceSatang != 10_000 || *got.MaxPriceSatang != 50_000 {
		t.Fatalf("UpdateProfile() = %+v, transaction calls = %d", got, tx.calls)
	}
	if len(repo.replaced) != 2 || repo.replaced[0].String() > repo.replaced[1].String() {
		t.Fatalf("ReplaceStyles() IDs = %v, want two unique sorted IDs", repo.replaced)
	}
}

func TestUpdateProfile_AllowsClearingStyles(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old"), Styles: []artist.Style{{ID: uuid.New()}}}
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})

	got, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{
		Description: stringPointer("new"), MinPriceSatang: 0, MaxPriceSatang: 0,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got.Styles == nil || len(got.Styles) != 0 {
		t.Fatalf("UpdateProfile() styles = %#v, want non-nil empty slice", got.Styles)
	}
}

func TestUpdateProfile_AllowsNullDescription(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old description")}
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})

	got, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{
		Description: nil, MinPriceSatang: 0, MaxPriceSatang: 0,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got.Description != nil {
		t.Fatalf("UpdateProfile() description = %q, want null", *got.Description)
	}
}

func TestUpdateProfile_RejectsInvalidInputBeforeTransaction(t *testing.T) {
	tests := []struct {
		name string
		in   artist.UpdateProfileInput
		want error
	}{
		{name: "negative minimum", in: artist.UpdateProfileInput{Description: stringPointer("valid"), MinPriceSatang: -1}, want: artist.ErrPriceMustBeNonNegative},
		{name: "negative maximum", in: artist.UpdateProfileInput{Description: stringPointer("valid"), MaxPriceSatang: -1}, want: artist.ErrPriceMustBeNonNegative},
		{name: "reversed range", in: artist.UpdateProfileInput{Description: stringPointer("valid"), MinPriceSatang: 2, MaxPriceSatang: 1}, want: artist.ErrInvalidPriceRange},
		{name: "empty style id", in: artist.UpdateProfileInput{Description: stringPointer("valid"), StyleIDs: []uuid.UUID{uuid.Nil}}, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			tx := &fakeTransactioner{repo: repo}
			usecase := artist.NewProfileUsecase(repo, tx)
			_, err := usecase.UpdateProfile(context.Background(), uuid.New(), tt.in)
			if err == nil {
				t.Fatal("UpdateProfile() error = nil, want validation error")
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("UpdateProfile() error = %v, want %v", err, tt.want)
			}
			if tx.calls != 0 {
				t.Fatalf("transaction calls = %d, want 0", tx.calls)
			}
		})
	}
}

func TestUpdateProfile_RejectsUnknownStyleWithoutChangingProfile(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old")}
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})

	_, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{
		Description: stringPointer("new"), StyleIDs: []uuid.UUID{uuid.New()}, MaxPriceSatang: 10,
	})
	if !errors.Is(err, artist.ErrStyleNotFound) {
		t.Fatalf("UpdateProfile() error = %v, want ErrStyleNotFound", err)
	}
	if got := repo.byUserID[userID].Description; got == nil || *got != "old" {
		t.Fatalf("profile changed after unknown style: %+v", repo.byUserID[userID])
	}
}

func TestUpdateProfile_RollsBackWhenStyleReplacementFails(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	styleID := uuid.New()
	repo.styles[styleID] = true
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old")}
	repo.replaceErr = errors.New("replace failed")
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})

	_, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{
		Description: stringPointer("new"), StyleIDs: []uuid.UUID{styleID}, MaxPriceSatang: 10,
	})
	if err == nil {
		t.Fatal("UpdateProfile() error = nil, want replacement error")
	}
	if got := repo.byUserID[userID]; got.Description == nil || *got.Description != "old" || got.MinPriceSatang != nil || got.MaxPriceSatang != nil {
		t.Fatalf("profile changed after rollback: %+v", got)
	}
}

func TestUpdateProfile_PropagatesRepositoryErrors(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := newFakeRepo()
	repo.countErr = repoErr
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old")}
	usecase := artist.NewProfileUsecase(repo, &fakeTransactioner{repo: repo})

	_, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{Description: stringPointer("new")})
	if !errors.Is(err, repoErr) {
		t.Fatalf("UpdateProfile() error = %v, want %v", err, repoErr)
	}

	repo.countErr = nil
	repo.getErr = repoErr
	_, err = usecase.GetProfile(context.Background(), userID)
	if !errors.Is(err, repoErr) {
		t.Fatalf("GetProfile() error = %v, want %v", err, repoErr)
	}
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byUserID: make(map[uuid.UUID]*artist.Profile), styles: make(map[uuid.UUID]bool)}
}

func cloneProfile(profile *artist.Profile) *artist.Profile {
	cp := *profile
	if profile.Categories != nil {
		cp.Categories = append([]artist.Category{}, profile.Categories...)
	}
	if profile.Styles != nil {
		cp.Styles = append([]artist.Style{}, profile.Styles...)
	}
	if profile.MinPriceSatang != nil {
		cp.MinPriceSatang = int64Pointer(*profile.MinPriceSatang)
	}
	if profile.MaxPriceSatang != nil {
		cp.MaxPriceSatang = int64Pointer(*profile.MaxPriceSatang)
	}
	if profile.Description != nil {
		cp.Description = stringPointer(*profile.Description)
	}
	return &cp
}

func int64Pointer(value int64) *int64    { return &value }
func stringPointer(value string) *string { return &value }
