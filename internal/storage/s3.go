package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/vincent119/images-filters/internal/config"
)

// S3API defines the interface for S3 client operations to allow mocking
type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3Storage implement Storage interface using AWS S3
type S3Storage struct {
	client S3API
	bucket string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(ctx context.Context, cfg config.S3StorageConfig) (*S3Storage, error) {
	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	if cfg.Endpoint != "" {
		// Custom endpoint resolver for S3 (e.g., MinIO or LocalStack)
		// For AWS SDK v2, this is handled via EndpointResolverWithOptions
		// or BaseEndpoint in newer versions, but let's stick to standard config loading first.
		// Actually, BaseEndpoint is available in LoadOptions in newer SDK versions?
		// Let's use a custom endpoint resolver if needed, but AWS SDK v2 setup can be tricky with endpoints.
		// NOTE: New approach for Endpoint in v2 is usually specifically on the client construction or via config.
		// Let's check if we can set it on client.
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	// Create S3 client options
	clientOpts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Often needed for custom endpoints like MinIO
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOpts...)

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Get retrieves image data from S3
func (s *S3Storage) Get(ctx context.Context, key string) ([]byte, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		// S3 SDK v2 might return different error types for "Not Found" depending on service behavior
		// checking 404 status code helper might be useful but errors.As is standard.
		return nil, fmt.Errorf("failed to get object from s3: %w", err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}

// Put saves image data to S3
func (s *S3Storage) Put(ctx context.Context, key string, data []byte) error {
	contentType := http.DetectContentType(data)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to s3: %w", err)
	}

	return nil
}

// Exists checks if image exists in S3
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		// HeadObject returns 404 as an error
		// We need to check if it's a 404 error
		// Using the standard error unwrap for 404
		// In some SDK versions, it might be types.NotFound or just a generic API error with 404 code
		// Let's try to be robust.
		var responseError interface {
			HTTPStatusCode() int
		}
		if errors.As(err, &responseError) {
			if responseError.HTTPStatusCode() == 404 {
				return false, nil
			}
		}

		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// Delete removes image from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from s3: %w", err)
	}

	return nil
}
