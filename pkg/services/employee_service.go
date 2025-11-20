package services

import (
	"context"
	"fmt"
	"time"

	"collie-document-manager-backend/pkg/domain"
	"collie-document-manager-backend/pkg/ports"

	"github.com/google/uuid"
)

type employeeService struct {
	repo ports.EmployeeRepository
}

// NewEmployeeService crea una nueva instancia de EmployeeService
func NewEmployeeService(repo ports.EmployeeRepository) ports.EmployeeService {
	return &employeeService{
		repo: repo,
	}
}

// CreateEmployee implementa ports.EmployeeService.
func (s *employeeService) CreateEmployee(ctx context.Context, employee domain.Employee) (*domain.Employee, error) {
	if employee.ID == "" {
		employee.ID = uuid.New().String()
	}
	if employee.LinkDate.IsZero() {
		employee.LinkDate = time.Now()
	}
	if employee.Status == "" {
		employee.Status = "Activo" // Estado por defecto
	}
	// RoleID puede ser vacío al crear, se asignará después

	err := s.repo.Save(ctx, &employee)
	if err != nil {
		return nil, fmt.Errorf("failed to save employee: %w", err)
	}
	return &employee, nil
}

// GetEmployeeByID implementa ports.EmployeeService.
func (s *employeeService) GetEmployeeByID(ctx context.Context, id string) (*domain.Employee, error) {
	employee, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find employee by ID: %w", err)
	}
	if employee == nil {
		return nil, nil // Empleado no encontrado
	}
	return employee, nil
}

// GetAllEmployees implementa ports.EmployeeService.
func (s *employeeService) GetAllEmployees(ctx context.Context) ([]domain.Employee, error) {
	employees, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all employees: %w", err)
	}
	return employees, nil
}

// UpdateEmployee implementa ports.EmployeeService.
func (s *employeeService) UpdateEmployee(ctx context.Context, id string, employee domain.Employee) (*domain.Employee, error) {
	existingEmployee, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing employee for update: %w", err)
	}
	if existingEmployee == nil {
		return nil, fmt.Errorf("employee with ID %s not found", id)
	}

	// Actualizar solo los campos proporcionados
	if employee.Name != "" {
		existingEmployee.Name = employee.Name
	}
	if employee.Email != "" {
		existingEmployee.Email = employee.Email
	}
	if employee.Status != "" {
		existingEmployee.Status = employee.Status
	}
	// Permitir actualizar RoleID
	existingEmployee.RoleID = employee.RoleID // AÑADIDO: Actualizar RoleID

	err = s.repo.Update(ctx, existingEmployee)
	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}
	return existingEmployee, nil
}

// DeleteEmployee implementa ports.EmployeeService.
func (s *employeeService) DeleteEmployee(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}

// Asegurarse de que employeeService implementa ports.EmployeeService
var _ ports.EmployeeService = (*employeeService)(nil)
