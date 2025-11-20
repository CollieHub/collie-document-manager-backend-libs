package ports

import (
	"collie-document-manager-backend/pkg/domain"
	"context"
)

// Primary Port (Interfaces para el servicio de aplicación)
type EmployeeService interface {
	CreateEmployee(ctx context.Context, employee domain.Employee) (*domain.Employee, error)
	GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error)
	GetAllEmployees(ctx context.Context) ([]domain.Employee, error)
	UpdateEmployee(ctx context.Context, id string, employee domain.Employee) (*domain.Employee, error)
	DeleteEmployee(ctx context.Context, id string) error
}

type DocumentService interface {
	CreateDocument(ctx context.Context, doc domain.Document) (*domain.Document, error)
	GetDocumentByID(ctx context.Context, id string) (*domain.Document, error)
	GetAllDocuments(ctx context.Context) ([]domain.Document, error)
	UpdateDocument(ctx context.Context, id string, doc domain.Document) (*domain.Document, error)
	DeleteDocument(ctx context.Context, id string) error
	RequestUploadURL(fileName string) (string, string, error) // Devuelve URL y Key
}

type RoleService interface { // AÑADIDO
	CreateRole(ctx context.Context, role domain.Role) (*domain.Role, error)
	GetRoleByID(ctx context.Context, id string) (*domain.Role, error)
	GetAllRoles(ctx context.Context) ([]domain.Role, error)
	UpdateRole(ctx context.Context, id string, role domain.Role) (*domain.Role, error)
	DeleteRole(ctx context.Context, id string) error
}

// Secondary Port (Interfaces para adaptadores de infraestructura)
type EmployeeRepository interface {
	Save(ctx context.Context, employee *domain.Employee) error
	FindByID(ctx context.Context, id string) (*domain.Employee, error)
	FindAll(ctx context.Context) ([]domain.Employee, error)
	Update(ctx context.Context, employee *domain.Employee) error
	Delete(ctx context.Context, id string) error
}

type DocumentRepository interface {
	Save(ctx context.Context, doc *domain.Document) error
	FindByID(ctx context.Context, id string) (*domain.Document, error)
	FindAll(ctx context.Context) ([]domain.Document, error)
	Update(ctx context.Context, doc *domain.Document) error
	Delete(ctx context.Context, id string) error
}

type RoleRepository interface { // AÑADIDO
	Save(ctx context.Context, role *domain.Role) error
	FindByID(ctx context.Context, id string) (*domain.Role, error)
	FindAll(ctx context.Context) ([]domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id string) error
}

type FileStorage interface {
	GeneratePresignedUploadURL(fileName string) (string, string, error) // Devuelve URL y Key
}
