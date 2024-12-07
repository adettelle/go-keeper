// Package service provides functionality for interacting with MinIO, including bucket management
// and object operations like upload and retrieval.
package service

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
)

// MinioService encapsulates operations for interacting with a MinIO server.
// It provides methods for bucket creation, file uploads, and file retrieval.
type MinioService struct {
	client     *minio.Client // MinIO client used for API interactions.
	bucketName string        // Name of the bucket where objects are stored.
	lifeTime   time.Duration // The duration for which objects in the bucket are valid.
}

func NewMinioService(client *minio.Client, bucketName string, lifeTime time.Duration) *MinioService {
	return &MinioService{
		client:     client,
		bucketName: bucketName,
		lifeTime:   lifeTime,
	}
}

// CreateBucket ensures that the bucket specified in the MinioService configuration exists.
// If the bucket does not exist, it is created.
func (ms *MinioService) CreateBucket() error {
	bucketExists, err := ms.client.BucketExists(context.Background(), ms.bucketName)
	if err != nil {
		return err
	}
	if bucketExists {
		return nil
	}

	err = ms.client.MakeBucket(context.Background(), ms.bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Upload uploads a file to the configured MinIO bucket.
func (ms *MinioService) Upload(fileCloudID string, reader io.Reader) error {
	info, err := ms.client.PutObject(context.Background(),
		ms.bucketName, fileCloudID, reader, -1, minio.PutObjectOptions{}) // TODO checkout Content-Length

	if err != nil {
		return err
	}
	log.Println(info)
	return nil
}

// GetObject retrieves a file from the configured MinIO bucket.
func (ms *MinioService) GetObject(fileCLoudID string) (io.Reader, error) {
	return ms.client.GetObject(context.Background(), ms.bucketName, fileCLoudID, minio.GetObjectOptions{})
}
