package seed

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"math/rand/v2"

	"github.com/fogleman/gg"
)

// originalImageSize and previewImageSize are the pixel dimensions
// generated for each original/preview pair.
const (
	originalImageSize = 1024
	previewImageSize  = 400
)

type objectBucket interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
}

// uploadPlaceholderPair generates and uploads an original/preview image
// pair to bucket at originalKey/previewKey. Both are rendered from
// seedKey, so the same seedKey always uploads pixel-identical bytes.
func uploadPlaceholderPair(ctx context.Context, bucket objectBucket, seedKey, originalKey, previewKey string) error {
	original, err := generatePlaceholderPNG(seedKey, originalImageSize)
	if err != nil {
		return fmt.Errorf("generate placeholder image for %q: %w", seedKey, err)
	}
	if err := bucket.Upload(ctx, originalKey, bytes.NewReader(original), "image/png"); err != nil {
		return err
	}

	preview, err := generatePlaceholderPNG(seedKey, previewImageSize)
	if err != nil {
		return fmt.Errorf("generate placeholder preview image for %q: %w", seedKey, err)
	}
	return bucket.Upload(ctx, previewKey, bytes.NewReader(preview), "image/png")
}

func deletePlaceholderPair(ctx context.Context, bucket objectBucket, originalKey, previewKey string) error {
	if err := bucket.Delete(ctx, originalKey); err != nil {
		return err
	}
	return bucket.Delete(ctx, previewKey)
}

// generatePlaceholderPNG renders a deterministic abstract-art placeholder:
// a two-color gradient background with a scatter of translucent blobs,
// seeded by key.
func generatePlaceholderPNG(key string, size int) ([]byte, error) {
	seed1, seed2 := hashSeed(key)
	rng := rand.New(rand.NewPCG(seed1, seed2))

	dc := gg.NewContext(size, size)

	grad := gg.NewLinearGradient(0, 0, float64(size), float64(size))
	grad.AddColorStop(0, randomColor(rng, 255))
	grad.AddColorStop(1, randomColor(rng, 255))
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(size), float64(size))
	dc.Fill()

	const blobCount = 6
	for range blobCount {
		x := rng.Float64() * float64(size)
		y := rng.Float64() * float64(size)
		r := (0.15 + rng.Float64()*0.25) * float64(size)
		dc.SetColor(randomColor(rng, 130))
		dc.DrawCircle(x, y, r)
		dc.Fill()
	}

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, fmt.Errorf("encode placeholder png: %w", err)
	}
	return buf.Bytes(), nil
}

// hashSeed turns key into two deterministic uint64 seeds for
// rand.NewPCG.
func hashSeed(key string) (uint64, uint64) {
	hash := sha256.Sum256([]byte(key))
	seed1 := binary.BigEndian.Uint64(hash[0:8])
	seed2 := binary.BigEndian.Uint64(hash[8:16])
	return seed1, seed2
}

// randomColor draws a random, reasonably saturated color with the given
// alpha from rng.
func randomColor(rng *rand.Rand, alpha uint8) color.NRGBA {
	return color.NRGBA{
		R: uint8(30 + rng.IntN(200)),
		G: uint8(30 + rng.IntN(200)),
		B: uint8(30 + rng.IntN(200)),
		A: alpha,
	}
}
