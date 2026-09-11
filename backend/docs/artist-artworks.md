# Artist Artworks

`GET /api/v1/artists/{artist_id}/artworks` is public and returns every artwork created by the artist. Results are ordered by `created_at` descending and then `artwork_id` descending.

```json
{
  "artworks": [
    {
      "artwork_id": "00000000-0000-0000-0000-000000000001",
      "name": "Book Cover",
      "category": "Illustration",
      "styles": ["Watercolor"],
      "description": "A colorful book-cover commission",
      "artwork_samples": [
        {
          "image_url": "https://storage.example.com/artworks/example/preview.webp"
        }
      ],
      "minimum_deadline_days": 7,
      "price_satang": 50000
    }
  ]
}
```

An existing artist without artworks receives `200` with `{"artworks": []}`. A missing artist, or an artist whose user account was deleted, receives `404`.
