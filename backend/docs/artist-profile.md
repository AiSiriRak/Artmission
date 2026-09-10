# Artist Profiles

Artist profiles are a 1:1 extension of artist users. The public name always comes from `users.username`; it is not copied into `artist_profiles`.

## API

`GET /api/v1/artists/{artist_id}` is public. It returns the artist ID and name, description, distinct categories derived from the artist's portfolio samples, independently selected styles, nullable price range in satang, and nullable review score.

`PUT /api/v1/artists/me` is restricted to an authenticated artist. It replaces the description, complete style selection, and price range. Both prices must be nonnegative and the minimum must not exceed the maximum. An empty `style_ids` array clears the selection.

Newly registered artists start with no selected styles and null prices. Registration still requires only the artist description; prices become non-null after the first profile update.

## Persistence

- `artist_profiles` stores description, price range, review score, and timestamps.
- `artist_styles` stores the artist's editable style selection.
- Categories are queried from `artist_samples.category_id` and are never written by the profile endpoint.
- Profile fields and style rows are updated in one transaction.
