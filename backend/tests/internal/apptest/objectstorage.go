//go:build integration

package apptest

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/AiSiriRak/Artmission/backend/internal/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const rustFSImage = "rustfs/rustfs:latest"

func StartObjectStorage(ctx context.Context, tb testing.TB) config.S3 {
	tb.Helper()

	const (
		accessKey     = "artmission-test"
		secretKey     = "artmission-test-secret"
		publicBucket  = "artmission-public-test"
		privateBucket = "artmission-private-test"
		region        = "us-east-1"
	)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        rustFSImage,
			ExposedPorts: []string{"9000/tcp"},
			Env: map[string]string{
				"RUSTFS_ACCESS_KEY": accessKey,
				"RUSTFS_SECRET_KEY": secretKey,
			},
			WaitingFor: wait.ForHTTP("/health").WithPort("9000/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		tb.Fatalf("apptest: start RustFS container: %v", err)
	}
	tb.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			tb.Logf("apptest: terminate RustFS container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		tb.Fatalf("apptest: read RustFS host: %v", err)
	}
	port, err := container.MappedPort(ctx, "9000/tcp")
	if err != nil {
		tb.Fatalf("apptest: read RustFS port: %v", err)
	}
	endpoint := "http://" + net.JoinHostPort(host, port.Port())

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		tb.Fatalf("apptest: configure RustFS client: %v", err)
	}
	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})
	for _, bucket := range []string{publicBucket, privateBucket} {
		if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
			tb.Fatalf("apptest: create bucket %q: %v", bucket, err)
		}
	}
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, publicBucket)
	if _, err := client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{Bucket: aws.String(publicBucket), Policy: aws.String(policy)}); err != nil {
		tb.Fatalf("apptest: make public bucket readable: %v", err)
	}

	return config.S3{
		PublicBucketName:  publicBucket,
		PrivateBucketName: privateBucket,
		PublicBaseURL:     endpoint + "/" + publicBucket,
		Endpoint:          endpoint,
		Region:            region,
		AccessKeyID:       accessKey,
		SecretAccessKey:   secretKey,
	}
}
