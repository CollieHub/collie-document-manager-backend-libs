package services

import (
	"context"
	"fmt"

	"collie-document-manager-backend/pkg/domain"
	"collie-document-manager-backend/pkg/ports"

	"github.com/google/uuid"
)

type roleService struct {
	repo ports.RoleRepository
}

// NewRoleService crea una nueva instancia de RoleService
func NewRoleService(repo ports.RoleRepository) ports.RoleService {
	return &roleService{
		repo: repo,
	}
}

// CreateRole implementa ports.RoleService.
func (s *roleService) CreateRole(ctx context.Context, role domain.Role) (*domain.Role, error) {
	if role.ID == "" {
		role.ID = uuid.New().String()
	}
	// Aquí podrías añadir validaciones adicionales para el rol

	err := s.repo.Save(ctx, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to save role: %w", err)
	}
	return &role, nil
}

// GetRoleByID implementa ports.RoleService.
func (s *roleService) GetRoleByID(ctx context.Context, id string) (*domain.Role, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find role by ID: %w", err)
	}
	if role == nil {
		return nil, nil // Rol no encontrado
	}
	return role, nil
}

// GetAllRoles implementa ports.RoleService.
func (s *roleService) GetAllRoles(ctx context.Context) ([]domain.Role, error) {
	roles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	return roles, nil
}

// UpdateRole implementa ports.RoleService.
func (s *roleService) UpdateRole(ctx context.Context, id string, role domain.Role) (*domain.Role, error) {
	existingRole, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing role for update: %w", err)
	}
	if existingRole == nil {
		return nil, fmt.Errorf("role with ID %s not found", id)
	}

	// Actualizar solo los campos proporcionados
	if role.Name != "" {
		existingRole.Name = role.Name
	}
	if role.Description != "" {
		existingRole.Description = role.Description
	}

	err = s.repo.Update(ctx, existingRole)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return existingRole, nil
}

// DeleteRole implementa ports.RoleService.
func (s *roleService) DeleteRole(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// Asegurarse de que roleService implementa ports.RoleService
var _ ports.RoleService = (*roleService)(nil)
