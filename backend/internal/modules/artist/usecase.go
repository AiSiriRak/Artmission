package artist

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"
	"github.com/google/uuid"
)

type profileUsecase struct {
	repo    ProfileRepository
	storage ObjectStorage
}

func NewProfileUsecase(repo ProfileRepository, storage ObjectStorage) ProfileUsecase {
	return &profileUsecase{repo: repo, storage: storage}
}

func (u *profileUsecase) CreateProfile(ctx context.Context, userID uuid.UUID, description *string) error {
	now := time.Now()
	return u.repo.Create(ctx, &Profile{
		UserID:      userID,
		Description: normalizeDescription(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (u *profileUsecase) GetProfile(ctx context.Context, userID uuid.UUID, query ProfileQuery) (*Profile, error) {
	normalized, err := normalizeProfileQuery(query)
	if err != nil {
		return nil, err
	}
	profile, err := u.repo.GetByUserID(ctx, userID, normalized)
	if err != nil {
		return nil, err
	}
	u.attachProfileURL(profile)
	return profile, nil
}

func (u *profileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*Profile, error) {
	if in.Description == nil && in.ProfileImage == nil && !in.RemoveProfileImage {
		return nil, ErrNoProfileChanges
	}
	if in.ProfileImage != nil && in.RemoveProfileImage {
		return nil, ErrConflictingProfileImageChange
	}

	current, err := u.repo.GetByUserID(ctx, userID, ProfileQuery{Limit: DefaultReviewLimit})
	if err != nil {
		return nil, err
	}

	update := ProfileUpdate{
		DescriptionSet: in.Description != nil,
		Description:    normalizeDescription(in.Description),
		UpdatedAt:      time.Now(),
	}
	var newKey *string
	if in.ProfileImage != nil {
		content, contentType, extension, err := readProfileImage(in.ProfileImage)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("artist-profiles/%s/%s.%s", userID, uuid.New(), extension)
		if err := u.storage.UploadPublic(ctx, key, bytes.NewReader(content), int64(len(content)), contentType); err != nil {
			return nil, apperror.Internal("failed to upload artist profile image", err)
		}
		newKey = &key
		update.ProfileImageKeySet = true
		update.ProfileImageKey = newKey
	} else if in.RemoveProfileImage {
		update.ProfileImageKeySet = true
	}

	if err := u.repo.UpdateByUserID(ctx, userID, update); err != nil {
		if newKey != nil {
			_ = u.storage.DeletePublic(ctx, *newKey)
		}
		return nil, err
	}

	if update.ProfileImageKeySet && current.ProfileImageKey != nil {
		_ = u.storage.DeletePublic(ctx, *current.ProfileImageKey)
	}

	return u.GetProfile(ctx, userID, ProfileQuery{Limit: DefaultReviewLimit})
}

func (u *profileUsecase) attachProfileURL(profile *Profile) {
	if profile.ProfileImageKey == nil {
		profile.ProfileURL = nil
		return
	}
	url := u.storage.PublicURL(*profile.ProfileImageKey)
	profile.ProfileURL = &url
}

func normalizeProfileQuery(query ProfileQuery) (ProfileQuery, error) {
	if query.Limit == 0 {
		query.Limit = DefaultReviewLimit
	}
	if query.Limit < 1 || query.Limit > MaxReviewLimit {
		return ProfileQuery{}, apperror.InvalidInput(fmt.Sprintf("limit must be between 1 and %d", MaxReviewLimit), nil)
	}
	if query.Offset < 0 {
		return ProfileQuery{}, apperror.InvalidInput("offset must not be negative", nil)
	}
	return query, nil
}

func normalizeDescription(description *string) *string {
	if description == nil {
		return nil
	}
	normalized := strings.TrimSpace(*description)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func readProfileImage(reader io.Reader) ([]byte, string, string, error) {
	content, err := io.ReadAll(io.LimitReader(reader, MaxProfileImageSize+1))
	if err != nil {
		return nil, "", "", apperror.InvalidInput("failed to read profile image", err)
	}
	if len(content) > MaxProfileImageSize {
		return nil, "", "", ErrProfileImageTooLarge
	}
	contentType, extension, ok := detectProfileImageType(content)
	if !ok {
		return nil, "", "", ErrInvalidProfileImage
	}
	return content, contentType, extension, nil
}

func detectProfileImageType(content []byte) (string, string, bool) {
	switch {
	case len(content) >= 3 && bytes.Equal(content[:3], []byte{0xff, 0xd8, 0xff}):
		return "image/jpeg", "jpg", true
	case len(content) >= 8 && bytes.Equal(content[:8], []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}):
		return "image/png", "png", true
	case len(content) >= 12 && string(content[:4]) == "RIFF" && string(content[8:12]) == "WEBP":
		return "image/webp", "webp", true
	default:
		return "", "", false
	}
}
