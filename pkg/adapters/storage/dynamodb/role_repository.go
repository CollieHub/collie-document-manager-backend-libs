package dynamodb

import (
	"context"
	"fmt"
	"os"

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

// RoleItem representa la estructura del ítem de DynamoDB para un Role
type RoleItem struct {
	ID          string `dynamodbav:"id"`
	Name        string `dynamodbav:"name"`
	Description string `dynamodbav:"description"`
}

type roleRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewRoleRepository crea una nueva instancia de RoleRepository
func NewRoleRepository(ctx context.Context, tableName string) (ports.RoleRepository, error) {
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

	return &roleRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

// toRoleItem convierte un domain.Role a RoleItem
func toRoleItem(role *domain.Role) (*RoleItem, error) {
	return &RoleItem{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}, nil
}

// toDomainRole convierte un RoleItem a domain.Role
func toDomainRole(item *RoleItem) (*domain.Role, error) {
	return &domain.Role{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
	}, nil
}

// Save implementa ports.RoleRepository.
func (r *roleRepository) Save(ctx context.Context, role *domain.Role) error {
	item, err := toRoleItem(role)
	if err != nil {
		return fmt.Errorf("failed to convert role to item: %w", err)
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal role item: %w", err)
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

// FindByID implementa ports.RoleRepository.
func (r *roleRepository) FindByID(ctx context.Context, id string) (*domain.Role, error) {
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

	var item RoleItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal role item: %w", err)
	}

	return toDomainRole(&item)
}

// FindAll implementa ports.RoleRepository.
func (r *roleRepository) FindAll(ctx context.Context) ([]domain.Role, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB table: %w", err)
	}

	var roleItems []RoleItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &roleItems)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal role items: %w", err)
	}

	roles := make([]domain.Role, len(roleItems))
	for i, item := range roleItems {
		role, err := toDomainRole(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert item to domain role: %w", err)
		}
		roles[i] = *role
	}
	return roles, nil
}

// Update implementa ports.RoleRepository.
func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	item, err := toRoleItem(role)
	if err != nil {
		return fmt.Errorf("failed to convert role to item: %w", err)
	}

	update := expression.Set(expression.Name("name"), expression.Value(item.Name))
	update.Set(expression.Name("description"), expression.Value(item.Description))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("failed to build expression: %w", err)
	}

	expressionAttributeNames := make(map[string]string)
	for k, v := range expr.Names() {
		expressionAttributeNames[k] = v
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

// Delete implementa ports.RoleRepository.
func (r *roleRepository) Delete(ctx context.Context, id string) error {
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

// Asegurarse de que roleRepository implementa ports.RoleRepository
var _ ports.RoleRepository = (*roleRepository)(nil)
