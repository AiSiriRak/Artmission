package artist_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/google/uuid"
)

type fakeRepo struct {
	byUserID   map[uuid.UUID]*artist.Profile
	getErr     error
	getErrCall int
	getCalls   int
	queries    []artist.ProfileQuery
	updateErr  error
	updates    []artist.ProfileUpdate
}

func (fake *fakeRepo) Create(_ context.Context, profile *artist.Profile) error {
	fake.byUserID[profile.UserID] = cloneProfile(profile)
	return nil
}

func (fake *fakeRepo) GetByUserID(_ context.Context, userID uuid.UUID, query artist.ProfileQuery) (*artist.Profile, error) {
	fake.getCalls++
	fake.queries = append(fake.queries, query)
	if fake.getErr != nil && (fake.getErrCall == 0 || fake.getCalls == fake.getErrCall) {
		return nil, fake.getErr
	}
	profile, ok := fake.byUserID[userID]
	if !ok {
		return nil, artist.ErrProfileNotFound
	}
	return cloneProfile(profile), nil
}

func (fake *fakeRepo) UpdateByUserID(_ context.Context, userID uuid.UUID, update artist.ProfileUpdate) error {
	fake.updates = append(fake.updates, update)
	if fake.updateErr != nil {
		return fake.updateErr
	}
	profile, ok := fake.byUserID[userID]
	if !ok {
		return artist.ErrProfileNotFound
	}
	if update.DescriptionSet {
		profile.Description = cloneString(update.Description)
	}
	if update.ProfileImageKeySet {
		profile.ProfileImageKey = cloneString(update.ProfileImageKey)
	}
	profile.UpdatedAt = update.UpdatedAt
	return nil
}

type uploadCall struct {
	key         string
	content     []byte
	size        int64
	contentType string
}

type fakeStorage struct {
	uploadErr error
	deleteErr error
	uploads   []uploadCall
	deletes   []string
}

func (fake *fakeStorage) UploadPublic(_ context.Context, key string, body io.Reader, size int64, contentType string) error {
	if fake.uploadErr != nil {
		return fake.uploadErr
	}
	content, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	fake.uploads = append(fake.uploads, uploadCall{key: key, content: content, size: size, contentType: contentType})
	return nil
}

func (fake *fakeStorage) DeletePublic(_ context.Context, key string) error {
	fake.deletes = append(fake.deletes, key)
	return fake.deleteErr
}

func (*fakeStorage) PublicURL(key string) string {
	return "https://public.test/" + key
}

var (
	_ artist.ProfileRepository = (*fakeRepo)(nil)
	_ artist.ObjectStorage     = (*fakeStorage)(nil)
)

func TestCreateProfile_NormalizesNullableDescription(t *testing.T) {
	repo := newFakeRepo()
	usecase := artist.NewProfileUsecase(repo, &fakeStorage{})
	userID := uuid.New()
	description := "  I paint portraits  "

	if err := usecase.CreateProfile(context.Background(), userID, &description); err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}
	if got := repo.byUserID[userID]; got.Description == nil || *got.Description != "I paint portraits" {
		t.Fatalf("CreateProfile() description = %#v", got.Description)
	}

	emptyID := uuid.New()
	empty := "   "
	if err := usecase.CreateProfile(context.Background(), emptyID, &empty); err != nil {
		t.Fatalf("CreateProfile(blank) error = %v", err)
	}
	if repo.byUserID[emptyID].Description != nil {
		t.Fatalf("CreateProfile(blank) description = %#v, want nil", repo.byUserID[emptyID].Description)
	}
}

func TestGetProfile_DefaultsPaginationAndResolvesProfileURL(t *testing.T) {
	repo := newFakeRepo()
	storage := &fakeStorage{}
	userID := uuid.New()
	key := "artist-profiles/id/image.png"
	repo.byUserID[userID] = &artist.Profile{UserID: userID, ArtistName: "Mali", ProfileImageKey: &key}
	usecase := artist.NewProfileUsecase(repo, storage)

	got, err := usecase.GetProfile(context.Background(), userID, artist.ProfileQuery{})
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if got.ProfileURL == nil || *got.ProfileURL != "https://public.test/"+key {
		t.Fatalf("GetProfile() profile URL = %#v", got.ProfileURL)
	}
	if len(repo.queries) != 1 || repo.queries[0].Limit != artist.DefaultReviewLimit || repo.queries[0].Offset != 0 {
		t.Fatalf("GetProfile() query = %+v", repo.queries)
	}
}

func TestGetProfile_ValidatesPagination(t *testing.T) {
	tests := []struct {
		name  string
		query artist.ProfileQuery
	}{
		{name: "negative limit", query: artist.ProfileQuery{Limit: -1}},
		{name: "limit above maximum", query: artist.ProfileQuery{Limit: artist.MaxReviewLimit + 1}},
		{name: "negative offset", query: artist.ProfileQuery{Offset: -1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeRepo()
			_, err := artist.NewProfileUsecase(repo, &fakeStorage{}).GetProfile(context.Background(), uuid.New(), test.query)
			if err == nil {
				t.Fatal("GetProfile() error = nil, want validation error")
			}
			if repo.getCalls != 0 {
				t.Fatalf("repository calls = %d, want 0", repo.getCalls)
			}
		})
	}
}

func TestUpdateProfile_AcceptsImageAtExactSizeLimit(t *testing.T) {
	repo := newFakeRepo()
	storage := &fakeStorage{}
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID}
	image := make([]byte, artist.MaxProfileImageSize)
	copy(image, pngImage())

	_, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{ProfileImage: bytes.NewReader(image)})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if len(storage.uploads) != 1 || storage.uploads[0].size != artist.MaxProfileImageSize {
		t.Fatalf("upload = %+v", storage.uploads)
	}
}

func TestUpdateProfile_RejectsEmptyAndConflictingChanges(t *testing.T) {
	image := bytes.NewReader(pngImage())
	tests := []struct {
		name string
		in   artist.UpdateProfileInput
		want error
	}{
		{name: "empty", in: artist.UpdateProfileInput{}, want: artist.ErrNoProfileChanges},
		{name: "upload and remove", in: artist.UpdateProfileInput{ProfileImage: image, RemoveProfileImage: true}, want: artist.ErrConflictingProfileImageChange},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeRepo()
			_, err := artist.NewProfileUsecase(repo, &fakeStorage{}).UpdateProfile(context.Background(), uuid.New(), test.in)
			if !errors.Is(err, test.want) {
				t.Fatalf("UpdateProfile() error = %v, want %v", err, test.want)
			}
			if repo.getCalls != 0 {
				t.Fatalf("repository calls = %d, want 0", repo.getCalls)
			}
		})
	}
}

func TestUpdateProfile_TrimsAndClearsDescription(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("old")}
	usecase := artist.NewProfileUsecase(repo, &fakeStorage{})
	description := "  new description  "

	got, err := usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{Description: &description})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got.Description == nil || *got.Description != "new description" {
		t.Fatalf("UpdateProfile() description = %#v", got.Description)
	}

	blank := "   "
	got, err = usecase.UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{Description: &blank})
	if err != nil {
		t.Fatalf("UpdateProfile(blank) error = %v", err)
	}
	if got.Description != nil {
		t.Fatalf("UpdateProfile(blank) description = %#v, want nil", got.Description)
	}
}

func TestUpdateProfile_UploadsImageAndDeletesPreviousObject(t *testing.T) {
	repo := newFakeRepo()
	storage := &fakeStorage{}
	userID := uuid.New()
	oldKey := "artist-profiles/old.png"
	repo.byUserID[userID] = &artist.Profile{UserID: userID, Description: stringPointer("preserved"), ProfileImageKey: &oldKey}
	image := pngImage()

	got, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{
		ProfileImage: bytes.NewReader(image),
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got.Description == nil || *got.Description != "preserved" {
		t.Fatalf("description was not preserved: %#v", got.Description)
	}
	if len(storage.uploads) != 1 || storage.uploads[0].contentType != "image/png" || storage.uploads[0].size != int64(len(image)) || !bytes.Equal(storage.uploads[0].content, image) {
		t.Fatalf("upload = %+v", storage.uploads)
	}
	if !strings.HasPrefix(storage.uploads[0].key, "artist-profiles/"+userID.String()+"/") || !strings.HasSuffix(storage.uploads[0].key, ".png") {
		t.Fatalf("upload key = %q", storage.uploads[0].key)
	}
	if len(storage.deletes) != 1 || storage.deletes[0] != oldKey {
		t.Fatalf("deletes = %v", storage.deletes)
	}
	if got.ProfileURL == nil || *got.ProfileURL != "https://public.test/"+storage.uploads[0].key {
		t.Fatalf("profile URL = %#v", got.ProfileURL)
	}
}

func TestUpdateProfile_RemovesImageAndIgnoresOldObjectDeleteFailure(t *testing.T) {
	repo := newFakeRepo()
	storage := &fakeStorage{deleteErr: errors.New("storage unavailable")}
	userID := uuid.New()
	oldKey := "artist-profiles/old.webp"
	repo.byUserID[userID] = &artist.Profile{UserID: userID, ProfileImageKey: &oldKey}

	got, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{RemoveProfileImage: true})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got.ProfileImageKey != nil || got.ProfileURL != nil {
		t.Fatalf("image was not cleared: %+v", got)
	}
	if len(storage.deletes) != 1 || storage.deletes[0] != oldKey {
		t.Fatalf("deletes = %v", storage.deletes)
	}
}

func TestUpdateProfile_ValidatesActualImageContentAndSize(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "invalid content", data: []byte("not an image"), want: artist.ErrInvalidProfileImage},
		{name: "too large", data: append(pngImage(), make([]byte, artist.MaxProfileImageSize)...), want: artist.ErrProfileImageTooLarge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeRepo()
			storage := &fakeStorage{}
			userID := uuid.New()
			repo.byUserID[userID] = &artist.Profile{UserID: userID}
			_, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{ProfileImage: bytes.NewReader(test.data)})
			if !errors.Is(err, test.want) {
				t.Fatalf("UpdateProfile() error = %v, want %v", err, test.want)
			}
			if len(storage.uploads) != 0 || len(repo.updates) != 0 {
				t.Fatalf("invalid image changed state: uploads=%d updates=%d", len(storage.uploads), len(repo.updates))
			}
		})
	}
}

func TestUpdateProfile_RecognizesEverySupportedImageType(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		contentType string
		extension   string
	}{
		{name: "jpeg", data: []byte{0xff, 0xd8, 0xff, 0x00}, contentType: "image/jpeg", extension: ".jpg"},
		{name: "png", data: pngImage(), contentType: "image/png", extension: ".png"},
		{name: "webp", data: []byte("RIFF\x00\x00\x00\x00WEBPdata"), contentType: "image/webp", extension: ".webp"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeRepo()
			storage := &fakeStorage{}
			userID := uuid.New()
			repo.byUserID[userID] = &artist.Profile{UserID: userID}
			_, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{ProfileImage: bytes.NewReader(test.data)})
			if err != nil {
				t.Fatalf("UpdateProfile() error = %v", err)
			}
			if len(storage.uploads) != 1 || storage.uploads[0].contentType != test.contentType || !strings.HasSuffix(storage.uploads[0].key, test.extension) {
				t.Fatalf("upload = %+v", storage.uploads)
			}
		})
	}
}

func TestUpdateProfile_CompensatesForRepositoryFailure(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := newFakeRepo()
	repo.updateErr = repoErr
	storage := &fakeStorage{}
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID}

	_, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{ProfileImage: bytes.NewReader(pngImage())})
	if !errors.Is(err, repoErr) {
		t.Fatalf("UpdateProfile() error = %v, want %v", err, repoErr)
	}
	if len(storage.uploads) != 1 || len(storage.deletes) != 1 || storage.deletes[0] != storage.uploads[0].key {
		t.Fatalf("storage compensation uploads=%+v deletes=%v", storage.uploads, storage.deletes)
	}
}

func TestUpdateProfile_PropagatesStorageAndRepositoryReadErrors(t *testing.T) {
	storageErr := errors.New("storage unavailable")
	repo := newFakeRepo()
	storage := &fakeStorage{uploadErr: storageErr}
	userID := uuid.New()
	repo.byUserID[userID] = &artist.Profile{UserID: userID}

	_, err := artist.NewProfileUsecase(repo, storage).UpdateProfile(context.Background(), userID, artist.UpdateProfileInput{ProfileImage: bytes.NewReader(pngImage())})
	if !errors.Is(err, storageErr) {
		t.Fatalf("UpdateProfile() error = %v, want wrapped storage error", err)
	}

	repoErr := errors.New("database unavailable")
	repo.getErr = repoErr
	_, err = artist.NewProfileUsecase(repo, storage).GetProfile(context.Background(), userID, artist.ProfileQuery{})
	if !errors.Is(err, repoErr) {
		t.Fatalf("GetProfile() error = %v, want %v", err, repoErr)
	}
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byUserID: make(map[uuid.UUID]*artist.Profile)}
}

func cloneProfile(profile *artist.Profile) *artist.Profile {
	copy := *profile
	copy.ProfileImageKey = cloneString(profile.ProfileImageKey)
	copy.ProfileURL = cloneString(profile.ProfileURL)
	copy.Description = cloneString(profile.Description)
	copy.Categories = append([]artist.Category{}, profile.Categories...)
	copy.Styles = append([]artist.Style{}, profile.Styles...)
	copy.Reviews = append([]artist.Review{}, profile.Reviews...)
	if profile.MinPriceSatang != nil {
		value := *profile.MinPriceSatang
		copy.MinPriceSatang = &value
	}
	if profile.MaxPriceSatang != nil {
		value := *profile.MaxPriceSatang
		copy.MaxPriceSatang = &value
	}
	if profile.ReviewScore != nil {
		value := *profile.ReviewScore
		copy.ReviewScore = &value
	}
	return &copy
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func pngImage() []byte {
	return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00}
}

func stringPointer(value string) *string { return &value }
