package s3

import (
	"context"
	"fmt"
	"os"
	"time"

	"collie-document-manager-backend/pkg/ports"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3FileStorage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucketName string
}

// NewS3FileStorage crea una nueva instancia de S3FileStorage
func NewS3FileStorage(ctx context.Context, bucketName string) (ports.FileStorage, error) {
	var cfg aws.Config
	var err error

	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		// Configuración para S3 Local (ej. LocalStack)
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           "http://host.docker.internal:4566", // Endpoint de LocalStack para S3
					SigningRegion: "us-east-1",
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
			config.WithEndpointResolverWithOptions(customResolver),
		)
	} else {
		// Configuración para AWS real
		cfg, err = config.LoadDefaultConfig(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	return &s3FileStorage{
		client:    client,
		presigner: presigner,
		bucketName: bucketName,
	}, nil
}

// GeneratePresignedUploadURL implementa ports.FileStorage.
func (s *s3FileStorage) GeneratePresignedUploadURL(fileName string) (string, string, error) {
	// Generar una clave única para el archivo en S3
	// Podrías usar UUID o alguna otra lógica para evitar colisiones
	key := fmt.Sprintf("uploads/%s_%d_%s", fileName, time.Now().Unix(), os.Getenv("AWS_REQUEST_ID")) // Ejemplo de clave

	request, err := s.presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(5*time.Minute)) // URL válida por 5 minutos
	if err != nil {
		return "", "", fmt.Errorf("failed to presign put object: %w", err)
	}

	return request.URL, key, nil
}

// Asegurarse de que s3FileStorage implementa ports.FileStorage
var _ ports.FileStorage = (*s3FileStorage)(nil)
