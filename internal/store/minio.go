package store

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
}

// NewMinioClient initializes a new MinIO client.
func NewMinioClient(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinioClient, error) {
	// Initialize a new MinIO client.
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{Client: client}, nil
}
