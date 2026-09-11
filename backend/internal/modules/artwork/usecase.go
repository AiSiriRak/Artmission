package artwork

import (
	"context"

	"github.com/google/uuid"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) ListByArtistID(ctx context.Context, artistID uuid.UUID) ([]Artwork, error) {
	artworks, err := u.repo.ListByArtistID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	if artworks == nil {
		artworks = make([]Artwork, 0)
	}
	for artworkIndex := range artworks {
		if artworks[artworkIndex].Styles == nil {
			artworks[artworkIndex].Styles = make([]string, 0)
		}
		if artworks[artworkIndex].Samples == nil {
			artworks[artworkIndex].Samples = make([]Sample, 0)
		}
	}
	return artworks, nil
}
