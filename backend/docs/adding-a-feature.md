# Adding a Feature

> Living doc: describes current practice, not a mandate. Code and doc disagree? Trust the code, then fix the doc (or don't — nobody's blocked on it). Ask instead of guessing if something's unclear.

A concrete, step-by-step walkthrough for shipping one new module end to end, following the patterns in [Project & Coding Structure](project-structure.md) and [Conventions](conventions.md). Worked example: adding an **artist profile** module (`GET/PUT /artists/me`) — pick a real module the same shape when you actually build one.

Order matters: each step only depends on the ones before it, and you can run `go build ./...` after every step to confirm it still compiles.

## 1. Entity — `internal/modules/artist/entity.go`

The plain domain struct. No `bun`/`json` tags — this type is shared by every layer.

```go
package artist

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID      uuid.UUID
	Bio         string
	Category    string
	Style       string
	AvgRating   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

## 2. Port — `internal/modules/artist/port.go`

The interfaces the module *offers* (`ProfileUsecase`) and *needs* (`ProfileRepository`), plus any command types those ports take. Nothing here knows Postgres exists.

```go
package artist

import (
	"context"

	"github.com/google/uuid"
)

type ProfileUsecase interface {
	GetMyProfile(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateMyProfile(ctx context.Context, userID uuid.UUID, bio, category, style string) (*Profile, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, p *Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpdateByUserID(ctx context.Context, userID uuid.UUID, p *Profile) error
}
```

## 3. Errors — `internal/modules/artist/errors.go`

Sentinel `apperror` values for anything a caller might branch on.

```go
package artist

import "github.com/AiSiriRak/Artmission/backend/internal/pkg/apperror"

var ErrProfileNotFound = apperror.NotFound("artist profile not found")
```

## 4. Usecase + test — `internal/modules/artist/usecase.go`

The implementation of the driving port, and the business rule (here: a profile can only be created for a user with `Role == user.RoleArtist`).

```go
package artist

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type profileUsecase struct {
	repo ProfileRepository
}

func NewProfileUsecase(repo ProfileRepository) ProfileUsecase {
	return &profileUsecase{repo: repo}
}

// ... GetMyProfile / UpdateMyProfile implementations
```

Write `usecase_test.go` alongside it *before* wiring anything else up — a hand-written `fakeRepo` implementing `ProfileRepository` over a `map[uuid.UUID]*Profile`, same shape as `internal/modules/user/usecase_test.go`. If the business rule is awkward to test here, that's a sign it's in the wrong layer.

## 5. Postgres adapter — `internal/adapters/postgres/artist_repository.go`

Implements `artist.ProfileRepository`. New file, same `postgres` package as every other adapter (adapters are grouped by technology, not by domain).

```go
package postgres

import (
	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/uptrace/bun"
)

type artistProfileModel struct {
	bun.BaseModel `bun:"table:artist_profiles,alias:ap"`

	UserID    uuid.UUID `bun:"user_id,pk"`
	Bio       string    `bun:"bio"`
	Category  string    `bun:"category"`
	Style     string    `bun:"style"`
	AvgRating float64   `bun:"avg_rating"`
	CreatedAt time.Time `bun:"created_at,nullzero"`
	UpdatedAt time.Time `bun:"updated_at,nullzero"`
}

// newArtistProfileModel / toDomain conversions — see user_repository.go for the shape.

type artistRepository struct{ exec baserepo.Executor }

var _ artist.ProfileRepository = (*artistRepository)(nil)

func NewArtistRepository(db *bun.DB) artist.ProfileRepository {
	return &artistRepository{exec: baserepo.NewExecutor(db)}
}

// Create / GetByUserID / UpdateByUserID — table-specific queries via b.exec.Run(...).
```

Add the table's migration alongside it: `task migrate-create -- create_artist_profiles`, then fill in the generated file under `internal/pkg/migrations/`.

## 6. REST handler — `internal/handler/rest/artist_handler.go`

Owns request/response DTOs and huma operation registration. Follows `auth_handler.go`'s shape: a `*ArtistHandler` struct, a `Register(api huma.API)` method, one method per operation, registered with `huma.Get`/`huma.Post`/etc.

```go
package rest

import (
	"context"

	"github.com/AiSiriRak/Artmission/backend/internal/modules/artist"
	"github.com/danielgtaylor/huma/v2"
)

type ArtistHandler struct {
	profileUsecase artist.ProfileUsecase
}

func NewArtistHandler(profileUsecase artist.ProfileUsecase) *ArtistHandler {
	return &ArtistHandler{profileUsecase: profileUsecase}
}

func (h *ArtistHandler) Register(api huma.API) {
	huma.Get(api, "/artists/me", h.getMyProfile,
		huma.OperationTags("artists"),
        func(o *huma.Operation) {
            o.OperationID = "get-my-artist-profile"
            o.Summary = "GetMyArtistProfile"
            o.Description = "Get the authenticated artist's own profile"
            o.Middlewares = append(o.Middlewares, requireAuth(api, h.authUsecase), requireRole(api, user.RoleArtist))
        },
    )

	// PUT /artists/me follows the same shape.
}

// getMyProfile: decode nothing (userID comes from AuthInfo via requestctx.go),
// call h.profileUsecase.GetMyProfile, map *artist.Profile -> a DTO, map errors via mapAppError.
```

Auth-guarded routes attach `requireAuth`/`requireRole` per-operation (not globally) — see `internal/handler/rest/middleware.go`.

## 7. Wire it in `internal/wiring/wiring.go`

Extend the existing `Wire` function's "adapters → modules → handlers → routes" block, in the same dependency order as everything else, then register the handler's routes right after constructing it:

```go
artistRepo := postgres.NewArtistRepository(cfg.DB)
artistUsecase := artist.NewProfileUsecase(artistRepo, userUsecase)
artistHandler := rest.NewArtistHandler(artistUsecase)
// ...
authHandler.Register(api)
orderHandler.Register(api)
artistHandler.Register(api)
```

Wiring it here — instead of directly in `cmd/cmd_serve.go` — means `tests/internal/apptest` picks it up automatically too; there's no second place to remember.

## 8. (Optional) Note the new module in `project-structure.md`

If you added a new top-level directory under `internal/modules/` or `internal/adapters/`, a one-line addition to [Project & Coding Structure](project-structure.md)'s tree keeps it useful for the next reader — but this is a courtesy, not a gate. Don't hold up a PR for it, and don't feel obligated to hunt down every doc that might now be one module short; the tree already says it's a snapshot, not a live listing.

## 9. Verify

```bash
gofmt -l . && go vet ./... && go build ./... && go test ./...
```

Then a real smoke test: `docker compose up -d`, `go run . --env-file .env serve`, and `curl`/check `/api/v1/docs` for the new operations before calling it done.

## When *not* to follow this exact shape

Not every addition is a new module. A field added to an existing entity, a new query on an existing repository, a new route on an existing handler — extend the existing files in place; don't create a parallel module for one method. This walkthrough is for something genuinely new: a distinct bounded concern with its own entity and business rules.
