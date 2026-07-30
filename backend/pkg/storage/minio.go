package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage provides an interface to interact with MinIO (S3 Compatible Storage)
type MinioStorage struct {
	client *minio.Client
	bucket string
}

// NewMinioStorage initializes the MinIO client and creates the bucket if it doesn't exist
func NewMinioStorage(endpoint, accessKey, secretKey, bucket, allowedIPs string, useSSL bool) (*MinioStorage, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	ctx := context.Background()
	// Check if bucket exists, create if not
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check minio bucket: %w", err)
	}
	if !exists {
		log.Printf("MinIO bucket '%s' not found. Creating it...", bucket)
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Always ensure bucket policy allows public read access for viewing uploaded images
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, bucket)
	if errPolicy := minioClient.SetBucketPolicy(ctx, bucket, policy); errPolicy != nil {
		log.Printf("Warning: failed to set public read bucket policy: %v", errPolicy)
	}

	return &MinioStorage{
		client: minioClient,
		bucket: bucket,
	}, nil
}

func (m *MinioStorage) UploadStream(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	if m == nil || m.client == nil {
		return fmt.Sprintf("/uploads/%s", objectName), nil
	}
	info, err := m.client.PutObject(ctx, m.bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio upload error: %w", err)
	}

	// Assuming HTTP scheme for public URL without SSL, or HTTPS if useSSL is true
	scheme := "http"
	if m.client.EndpointURL().Scheme == "https" {
		scheme = "https"
	}

	publicURL := fmt.Sprintf("%s://%s/%s/%s", scheme, m.client.EndpointURL().Host, m.bucket, info.Key)
	return publicURL, nil
}

// Delete removes an object from the bucket
func (m *MinioStorage) Delete(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
}

// UpdateIPWhitelistPolicy dynamically updates the bucket policy's IP whitelist.
// Called when an Admin changes MINIO_ALLOWED_IPS through the dashboard.
func (m *MinioStorage) UpdateIPWhitelistPolicy(ctx context.Context, allowedIPs string) error {
	ipList := []string{}
	for _, ip := range strings.Split(allowedIPs, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			ipList = append(ipList, ip)
		}
	}
	ipJSON, _ := json.Marshal(ipList)

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": ["s3:GetObject"],
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Resource": ["arn:aws:s3:::%s/*"],
				"Condition": {
					"IpAddress": {
						"aws:SourceIp": %s
					}
				}
			}
		]
	}`, m.bucket, string(ipJSON))

	return m.client.SetBucketPolicy(ctx, m.bucket, policy)
}
