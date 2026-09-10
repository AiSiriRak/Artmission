// Package objectstorage wires up the S3-compatible client (RustFS locally,
// whichever S3-compatible service backs it in production) for Artmission's
// two buckets: a public one served via direct URLs, and a private one only
// ever reached through presigned URLs.
package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client holds one bound handle per bucket, both sharing the same
// underlying S3-compatible connection and credentials (built once by
// NewS3Client). Public and Private are distinct types so a caller can
// only ever call the operations that make sense for that bucket — e.g.
// GetPresignedURL only exists on Private, PublicURL only on Public.
type Client struct {
	Public  *PublicBucket
	Private *PrivateBucket
}

func NewS3Client(ctx context.Context, cfg config.S3) (*Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load object storage aws config: %w", err)
	}

	raw := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(cfg.Endpoint, "/"))
		o.UsePathStyle = true
	})

	return &Client{
		Public: &PublicBucket{
			raw:     raw,
			name:    cfg.PublicBucketName,
			baseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		},
		Private: &PrivateBucket{
			raw:     raw,
			name:    cfg.PrivateBucketName,
			presign: s3.NewPresignClient(raw),
		},
	}, nil
}

// PingContext confirms both buckets are reachable.
func (c *Client) PingContext(ctx context.Context) error {
	if err := c.Public.headCheck(ctx); err != nil {
		return err
	}
	if err := c.Private.headCheck(ctx); err != nil {
		return err
	}
	return nil
}

// bucket is the operations shared by both buckets.
//
// TODO: add UploadMany/DeleteMany (bounded-concurrency PutObject fan-out;
// chunked DeleteObjects, which supports up to 1000 keys per call) once a
// caller needs to write/remove more than one object at a time.
type bucket struct {
	raw  *s3.Client
	name string
}

func (b *bucket) headCheck(ctx context.Context) error {
	if _, err := b.raw.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(b.name)}); err != nil {
		return fmt.Errorf("head bucket %q: %w", b.name, err)
	}
	return nil
}

// Upload puts body at key with contentType, overwriting any existing
// object at that key.
func (b *bucket) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := b.raw.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(b.name),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload %q to bucket %q: %w", key, b.name, err)
	}
	return nil
}

// Delete removes the object at key. Deleting a key that doesn't exist is
// not an error (S3 DeleteObject semantics).
func (b *bucket) Delete(ctx context.Context, key string) error {
	_, err := b.raw.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %q from bucket %q: %w", key, b.name, err)
	}
	return nil
}

// PublicBucket is the public-read bucket — e.g. artwork images.
type PublicBucket struct {
	bucket
	baseURL string
}

// PublicURL returns key's stable, directly-fetchable URL. Pure string
// construction, no S3 call.
func (b *PublicBucket) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", b.baseURL, path.Clean(key))
}

// KeyFromURL is the inverse of PublicURL. ok is false for URLs this
// bucket did not produce.
func (b *PublicBucket) KeyFromURL(rawURL string) (key string, ok bool) {
	base, err := url.Parse(b.baseURL)
	if err != nil {
		return "", false
	}
	candidate, err := url.Parse(rawURL)
	if err != nil || candidate.Scheme != base.Scheme || candidate.Host != base.Host {
		return "", false
	}
	prefix := strings.TrimRight(base.Path, "/") + "/"
	if !strings.HasPrefix(candidate.Path, prefix) {
		return "", false
	}
	key = strings.TrimPrefix(candidate.Path, prefix)
	if key == "" {
		return "", false
	}
	return key, true
}

// PrivateBucket is the private bucket, only ever reachable through
// presigned URLs — e.g. order deliverables.
type PrivateBucket struct {
	bucket
	presign *s3.PresignClient
}

// GetPresignedURL returns a time-limited GET URL for key. key must never
// be the empty string; callers check for "no object yet" (a nil key)
// before calling this.
func (b *PrivateBucket) GetPresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := b.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
