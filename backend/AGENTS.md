# Agent notes

Read these before changing code:

- [Architecture](docs/architecture.md)
- [Project & Coding Structure](docs/project-structure.md)
- [Conventions](docs/conventions.md)
- [Adding a Feature](docs/adding-a-feature.md)

Trust the code if it disagrees with a doc, then fix the doc (or don't — nobody's blocked on it).

## Comments

Comment a **rule the next line cannot show**. Do not restate the type, the architecture, or knowledge the caller already has.

Hexagonal vocabulary (`driving port`, `driven port`, `usecase other layers call into`) belongs in `docs/architecture.md`, not on every interface.

**Package comments** are one sentence of ownership:

```go
// Package user owns account identity and credentials.
```

Drop the rest: later increments, “this slice only…”, how another package depends on this one unless that dependency is a cycle you are preventing.

**Keep** (a rule you cannot see from the signature or the next line):

```go
// Artist is required when Role is artist, and must be nil for a customer.
Artist *ArtistProfileInput

// RoleAdmin is never accepted (admins are seeded/ops-managed, not self-registered).
Register(ctx context.Context, in RegisterInput) (*User, error)

// Authenticate returns ErrInvalidCredential for both "no such user"
// and "wrong password" so a caller cannot distinguish account existence.

// ArtistRegistrar is implemented by the artist module and injected at wiring
// time so this package never imports artist (artist will later import user).
```

**Delete**:

```go
// UserUsecase is the driving port other layers call into.
// UserRepository is the driven port for user persistence.
// GetByID looks up a user by id, used by modules/auth to rehydrate the user.
// ProfileUsecase is the driving port for artist-profile writes.
// Soft-deleted users look like they do not exist (the repository never returns them).
```

The last one is a repository filter (`deleted_at` / bun `soft_delete`), not a usecase rule.

When in doubt, omit the comment.
