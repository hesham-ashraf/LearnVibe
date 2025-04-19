package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sony/gobreaker"
)

// StorageService handles object storage operations
type StorageService struct {
	client     *minio.Client
	bucketName string
}

// NewStorageService creates a new storage service
func NewStorageService(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*StorageService, error) {
	// Initialize MinIO client with retry logic
	var minioClient *minio.Client
	var err error

	operation := func() error {
		// Initialize MinIO client
		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			return fmt.Errorf("failed to create MinIO client: %v", err)
		}

		// Check if the bucket exists
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err != nil {
			return fmt.Errorf("failed to check if bucket exists: %v", err)
		}

		// Create bucket if it doesn't exist
		if !exists {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				return fmt.Errorf("failed to create bucket: %v", err)
			}
			log.Printf("Created bucket: %s\n", bucketName)
		}

		return nil
	}

	// Retry with exponential backoff for transient connection issues
	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	// Circuit Breaker setup for MinIO
	settings := gobreaker.Settings{
		Name:    "MinIOService",
		Timeout: 10 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	// Final check with circuit breaker
	_, err = cb.Execute(func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return minioClient.BucketExists(ctx, bucketName)
	})
	if err != nil {
		return nil, fmt.Errorf("MinIO connection failed: %v", err)
	}

	return &StorageService{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to the storage service
func (s *StorageService) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	// Circuit breaker for upload operation
	settings := gobreaker.Settings{
		Name:    "MinIOUploadService",
		Timeout: 30 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	_, err := cb.Execute(func() (interface{}, error) {
		// Upload with exponential backoff retry
		operation := func() error {
			_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
				ContentType: contentType,
			})
			return err
		}

		// Use a shorter backoff for uploads
		backOff := backoff.NewExponentialBackOff()
		backOff.MaxElapsedTime = 2 * time.Minute

		return nil, backoff.Retry(operation, backOff)
	})

	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for downloading a file
func (s *StorageService) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	// Circuit breaker for presigned URL generation
	settings := gobreaker.Settings{
		Name:    "MinIOPresignedURLService",
		Timeout: 5 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	result, err := cb.Execute(func() (interface{}, error) {
		// Get presigned URL with retry
		var presignedURL string
		operation := func() error {
			// Generate presigned URL
			reqParams := url.Values{} // Use url.Values instead of map[string]string

			url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, expiry, reqParams)
			if err != nil {
				return err
			}

			presignedURL = url.String()
			return nil
		}

		// Retry with exponential backoff
		err := backoff.Retry(operation, backoff.NewExponentialBackOff())
		if err != nil {
			return "", err
		}

		return presignedURL, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return result.(string), nil
}

// DeleteFile deletes a file from the storage
func (s *StorageService) DeleteFile(ctx context.Context, objectName string) error {
	// Circuit breaker for delete operation
	settings := gobreaker.Settings{
		Name:    "MinIODeleteService",
		Timeout: 10 * time.Second,
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	_, err := cb.Execute(func() (interface{}, error) {
		// Delete with retry
		operation := func() error {
			return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
		}

		return nil, backoff.Retry(operation, backoff.NewExponentialBackOff())
	})

	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}