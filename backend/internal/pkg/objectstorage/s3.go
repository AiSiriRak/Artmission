package objectstorage

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TODO: add Upload/Delete once the artist deliverable-submission and
// artwork-image endpoints exist; GetPresignedURL and PublicURL are the
// only operations ViewOrders needs today.
type Client struct {
	raw               *s3.Client
	presign           *s3.PresignClient
	publicBucketName  string
	privateBucketName string
	publicBaseURL     string
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

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(strings.TrimRight(cfg.Endpoint, "/"))
		o.UsePathStyle = true
	})

	return &Client{
		raw:               client,
		presign:           s3.NewPresignClient(client),
		publicBucketName:  cfg.PublicBucketName,
		privateBucketName: cfg.PrivateBucketName,
		publicBaseURL:     strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (c *Client) GetPresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.privateBucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (c *Client) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", c.publicBaseURL, path.Clean(key))
}

func (c *Client) PingContext(ctx context.Context) error {
	if _, err := c.raw.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.privateBucketName)}); err != nil {
		return fmt.Errorf("head private bucket %q: %w", c.privateBucketName, err)
	}
	if _, err := c.raw.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.publicBucketName)}); err != nil {
		return fmt.Errorf("head public bucket %q: %w", c.publicBucketName, err)
	}
	return nil
}
