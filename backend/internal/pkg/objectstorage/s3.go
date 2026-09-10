package objectstorage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	raw               *s3.Client
	presign           *s3.PresignClient
	publicBucketName  string
	privateBucketName string
	publicBaseURL     string
}

func (c *Client) UploadPublic(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := c.raw.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.publicBucketName),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	return err
}

func (c *Client) DeletePublic(ctx context.Context, key string) error {
	_, err := c.raw.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.publicBucketName),
		Key:    aws.String(key),
	})
	return err
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
