package service

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioService struct {
	client     *minio.Client
	bucketName string
	lifeTime   time.Duration
}

func NewMinioService(client *minio.Client, bucketName string, lifeTime time.Duration) *MinioService {
	return &MinioService{
		client:     client,
		bucketName: bucketName,
		lifeTime:   lifeTime,
	}
}

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

func (ms *MinioService) Upload(fileCloudID string, reader io.Reader) error {
	info, err := ms.client.PutObject(context.Background(),
		ms.bucketName, fileCloudID, reader, -1, minio.PutObjectOptions{}) // TODO checkout Content-Length

	if err != nil {
		return err
	}
	log.Println(info)
	return nil
}

// func (ms *MinioService) PresignedPutObject(fileCLoudID string) (string, error) {
// 	url, err := ms.client.PresignedPutObject(context.Background(), ms.bucketName, fileCLoudID, ms.lifeTime)
// 	if err != nil {
// 		return "", err
// 	}
// 	return url.String(), nil
// }

func (ms *MinioService) GetObject(fileCLoudID string) (io.Reader, error) {
	return ms.client.GetObject(context.Background(), ms.bucketName, fileCLoudID, minio.GetObjectOptions{})
}
