package artist

import (
	"context"
)

type ProfileRepository interface {
	Create(ctx context.Context, p *Profile) error
}
