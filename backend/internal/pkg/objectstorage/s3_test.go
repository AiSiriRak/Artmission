package objectstorage

import "testing"

func TestKeyFromURL(t *testing.T) {
	bucket := &PublicBucket{baseURL: "https://cdn.example.com/public"}

	t.Run("round-trips PublicURL", func(t *testing.T) {
		key := "artists/a1/artworks/w1/sample.png"
		got, ok := bucket.KeyFromURL(bucket.PublicURL(key))
		if !ok || got != key {
			t.Fatalf("KeyFromURL(PublicURL(%q)) = (%q, %v)", key, got, ok)
		}
	})

	t.Run("rejects foreign URLs", func(t *testing.T) {
		if _, ok := bucket.KeyFromURL("https://example.com/book-cover.png"); ok {
			t.Fatal("expected foreign URL to be rejected")
		}
	})
}
