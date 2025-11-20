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
	"github.com/google/uuid"
)

// EmployeeItem representa la estructura del ítem de DynamoDB para un Employee
type EmployeeItem struct {
	ID       string `dynamodbav:"id"`
	Name     string `dynamodbav:"name"`
	Email    string `dynamodbav:"email"`
	Status   string `dynamodbav:"status"`
	LinkDate string `dynamodbav:"linkDate"`         // Almacenar la fecha como string ISO 8601
	RoleID   string `dynamodbav:"roleId,omitempty"` // AÑADIDO: ID del rol, omitempty para no guardar si está vacío
}

type employeeRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewEmployeeRepository crea una nueva instancia de EmployeeRepository
func NewEmployeeRepository(ctx context.Context, tableName string) (ports.EmployeeRepository, error) {
	var cfg aws.Config
	var err error

	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		// Configuración para DynamoDB Local si AWS_SAM_LOCAL está presente
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           "http://host.docker.internal:8000", // Asegúrate de que Docker esté ejecutándose y DynamoDB Local esté accesible aquí
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

	return &employeeRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

// toEmployeeItem convierte un domain.Employee a EmployeeItem
func toEmployeeItem(emp *domain.Employee) (*EmployeeItem, error) {
	return &EmployeeItem{
		ID:       emp.ID,
		Name:     emp.Name,
		Email:    emp.Email,
		Status:   emp.Status,
		LinkDate: emp.LinkDate.Format(time.RFC3339),
		RoleID:   emp.RoleID, // AÑADIDO
	}, nil
}

// toDomainEmployee convierte un EmployeeItem a domain.Employee
func toDomainEmployee(item *EmployeeItem) (*domain.Employee, error) {
	linkDate, err := time.Parse(time.RFC3339, item.LinkDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LinkDate: %w", err)
	}
	return &domain.Employee{
		ID:       item.ID,
		Name:     item.Name,
		Email:    item.Email,
		Status:   item.Status,
		LinkDate: linkDate,
		RoleID:   item.RoleID, // AÑADIDO
	}, nil
}

// Save implementa ports.EmployeeRepository.
func (r *employeeRepository) Save(ctx context.Context, employee *domain.Employee) error {
	if employee.ID == "" {
		employee.ID = uuid.New().String()
	}
	item, err := toEmployeeItem(employee)
	if err != nil {
		return fmt.Errorf("failed to convert employee to item: %w", err)
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal employee item: %w", err)
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

// FindByID implementa ports.EmployeeRepository.
func (r *employeeRepository) FindByID(ctx context.Context, id string) (*domain.Employee, error) {
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

	var item EmployeeItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal employee item: %w", err)
	}

	return toDomainEmployee(&item)
}

// FindAll implementa ports.EmployeeRepository.
func (r *employeeRepository) FindAll(ctx context.Context) ([]domain.Employee, error) {
	// Usar Scan por simplicidad, pero para tablas grandes, considerar paginación o Query con un índice.
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan DynamoDB table: %w", err)
	}

	var employeeItems []EmployeeItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &employeeItems)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal employee items: %w", err)
	}

	employees := make([]domain.Employee, len(employeeItems))
	for i, item := range employeeItems {
		emp, err := toDomainEmployee(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert item to domain employee: %w", err)
		}
		employees[i] = *emp
	}
	return employees, nil
}

// Update implementa ports.EmployeeRepository.
func (r *employeeRepository) Update(ctx context.Context, employee *domain.Employee) error {
	item, err := toEmployeeItem(employee)
	if err != nil {
		return fmt.Errorf("failed to convert employee to item: %w", err)
	}

	// Construir expresión de actualización
	update := expression.Set(expression.Name("name"), expression.Value(item.Name))
	update.Set(expression.Name("email"), expression.Value(item.Email))
	update.Set(expression.Name("status"), expression.Value(item.Status))
	update.Set(expression.Name("linkDate"), expression.Value(item.LinkDate))
	update.Set(expression.Name("roleId"), expression.Value(item.RoleID)) // AÑADIDO: Actualizar RoleID

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("failed to build expression: %w", err)
	}

	// Convertir map[string]*string a map[string]string
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

// Delete implementa ports.EmployeeRepository.
func (r *employeeRepository) Delete(ctx context.Context, id string) error {
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

// Asegurarse de que employeeRepository implementa ports.EmployeeRepository
var _ ports.EmployeeRepository = (*employeeRepository)(nil)
