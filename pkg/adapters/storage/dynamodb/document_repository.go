package dynamodb

import (
	"context"
	"fmt"
	"os"
	"time"

	"collie-document-manager-backend/pkg/domain"
	"collie-document-manager-backend/pkg/ports"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DocumentItem representa la estructura del ítem de DynamoDB para un Document
type DocumentItem struct {
	ID                string `dynamodbav:"id"`
	FileName          string `dynamodbav:"fileName"`
	S3Key             string `dynamodbav:"s3Key"`
	UploadDate        string `dynamodbav:"uploadDate"` // Almacenar la fecha como string ISO 8601
	Status            string `dynamodbav:"status"`
	OwnerID           string `dynamodbav:"ownerId"`
	RequiresSignature bool   `dynamodbav:"requiresSignature"`
	DocumentType      string `dynamodbav:"documentType"`
	GroupName         string `dynamodbav:"groupName"`
}

type documentRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDocumentRepository crea una nueva instancia de DocumentRepository
func NewDocumentRepository(ctx context.Context, tableName string) (ports.DocumentRepository, error) {
	var cfg aws.Config
	var err error

	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           "http://host.docker.internal:8000",
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
		cfg, err = config.LoadDefaultConfig(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	return &documentRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

// toDocumentItem convierte un domain.Document a DocumentItem
func toDocumentItem(doc *domain.Document) (*DocumentItem, error) {
	return &DocumentItem{
		ID:                doc.ID,
		FileName:          doc.FileName,
		S3Key:             doc.S3Key,
		UploadDate:        doc.UploadDate.Format(time.RFC3339),
		Status:            doc.Status,
		OwnerID:           doc.OwnerID,
		RequiresSignature: doc.RequiresSignature,
		DocumentType:      doc.DocumentType,
		GroupName:         doc.GroupName,
	}, nil
}

// toDomainDocument convierte un DocumentItem a domain.Document
func toDomainDocument(item *DocumentItem) (*domain.Document, error) {
	uploadDate, err := time.Parse(time.RFC3339, item.UploadDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UploadDate: %w", err)
	}
	return &domain.Document{
		ID:                item.ID,
		FileName:          item.FileName,
		S3Key:             item.S3Key,
		UploadDate:        uploadDate,
		Status:            item.Status,
		OwnerID:           item.OwnerID,
		RequiresSignature: item.RequiresSignature,
		DocumentType:      item.DocumentType,
		GroupName:         item.GroupName,
	}, nil
}

// Save implementa ports.DocumentRepository.
func (r *documentRepository) Save(ctx context.Context, doc *domain.Document) error {
	item, err := toDocumentItem(doc)
	if err != nil {
		return fmt.Errorf("failed to convert document to item: %w", err)
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal document item: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}
	return nil
}

// FindByID implementa ports.DocumentRepository.
func (r *documentRepository) FindByID(ctx context.Context, id string) (*domain.Document, error) {
	key, err := attributevalue.MarshalMap(map[string]string{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}
	if result.Item == nil {
		return nil, nil // No encontrado
	}

	var item DocumentItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal document item: %w", err)
	}

	return toDomainDocument(&item)
}

// FindAll implementa ports.DocumentRepository.
func (r *documentRepository) FindAll(ctx context.Context) ([]domain.Document, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB table: %w", err)
	}

	var documentItems []DocumentItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &documentItems)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal document items: %w", err)
	}

	documents := make([]domain.Document, len(documentItems))
	for i, item := range documentItems {
		doc, err := toDomainDocument(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert item to domain document: %w", err)
		}
		documents[i] = *doc
	}
	return documents, nil
}

// Update implementa ports.DocumentRepository.
func (r *documentRepository) Update(ctx context.Context, doc *domain.Document) error {
	item, err := toDocumentItem(doc)
	if err != nil {
		return fmt.Errorf("failed to convert document to item: %w", err)
	}

	update := expression.Set(expression.Name("fileName"), expression.Value(item.FileName))
	update.Set(expression.Name("s3Key"), expression.Value(item.S3Key))
	update.Set(expression.Name("uploadDate"), expression.Value(item.UploadDate))
	update.Set(expression.Name("status"), expression.Value(item.Status))
	update.Set(expression.Name("ownerId"), expression.Value(item.OwnerID))
	update.Set(expression.Name("requiresSignature"), expression.Value(item.RequiresSignature))
	update.Set(expression.Name("documentType"), expression.Value(item.DocumentType))
	update.Set(expression.Name("groupName"), expression.Value(item.GroupName))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("failed to build expression: %w", err)
	}

	// Convertir map[string]*string a map[string]string
	expressionAttributeNames := make(map[string]string)
	for k, v := range expr.Names() {
		expressionAttributeNames[k] = v // CORREGIDO: No desreferenciar v
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: item.ID}},
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})
	if err != nil {
		return fmt.Errorf("failed to update item in DynamoDB: %w", err)
	}
	return nil
}

// Delete implementa ports.DocumentRepository.
func (r *documentRepository) Delete(ctx context.Context, id string) error {
	key, err := attributevalue.MarshalMap(map[string]string{"id": id})
	if err != nil {
		return fmt.Errorf("failed to marshal key for delete: %w", err)
	}

	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete item from DynamoDB: %w", err)
	}
	return nil
}

// Asegurarse de que documentRepository implementa ports.DocumentRepository
var _ ports.DocumentRepository = (*documentRepository)(nil)
