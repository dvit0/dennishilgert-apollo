package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioObjectUploadedEvent struct {
	EventName string
	Key       string
}

type Options struct {
	Endpoint        string
	AccessKeyId     string
	SecretAccessKey string
}

type StorageService interface {
	ListContents(ctx context.Context, bucketName string, prefix string) ([]string, error)
	PresignUpload(ctx context.Context, bucketName string, objectName string, expires time.Duration) (*url.URL, error)
	UploadObject(ctx context.Context, bucketName string, objectName string, filePath string) (*minio.UploadInfo, error)
	DownloadObject(ctx context.Context, bucketName string, objectName string, targetPath string) error
}

type storageService struct {
	minioClient minio.Client
}

// NewStorageService creates a new storage service.
func NewStorageService(opts Options) (StorageService, error) {
	minioClient, err := minio.New(opts.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(opts.AccessKeyId, opts.SecretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	return &storageService{
		minioClient: *minioClient,
	}, nil
}

// ListContents returns a list of all objects in the specified bucket.
func (s *storageService) ListContents(ctx context.Context, bucketName string, prefix string) ([]string, error) {
	var contents []string
	objectCh := s.minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix: prefix,
	})
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		contents = append(contents, object.Key)
	}
	return contents, nil
}

// PresignUpload returns a presigned URL for uploading an object to the storage.
func (s *storageService) PresignUpload(ctx context.Context, bucketName string, objectName string, expires time.Duration) (*url.URL, error) {
	presignedUrl, err := s.minioClient.PresignedPutObject(ctx, bucketName, objectName, expires)
	if err != nil {
		return nil, err
	}
	return presignedUrl, nil
}

// UploadObject uploads an object to the storage.
func (s *storageService) UploadObject(ctx context.Context, bucketName string, objectName string, filePath string) (*minio.UploadInfo, error) {
	info, err := s.minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// DownloadObject downloads an object from the storage to the target path.
func (s *storageService) DownloadObject(ctx context.Context, bucketName string, objectName string, targetPath string) error {
	if err := s.minioClient.FGetObject(ctx, bucketName, objectName, targetPath, minio.GetObjectOptions{}); err != nil {
		return err
	}
	return nil
}
