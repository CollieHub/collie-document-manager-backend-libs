package services

import (
	"context"
	"fmt"
	"time"

	"collie-document-manager-backend/pkg/domain"
	"collie-document-manager-backend/pkg/ports"

	"github.com/google/uuid"
)

type documentService struct {
	fileStorage ports.FileStorage
	repo        ports.DocumentRepository // AÑADIDO
}

// NewDocumentService crea una nueva instancia de DocumentService
func NewDocumentService(fs ports.FileStorage, repo ports.DocumentRepository) ports.DocumentService { // ACTUALIZADO
	return &documentService{
		fileStorage: fs,
		repo:        repo,
	}
}

// CreateDocument implementa ports.DocumentService.
func (s *documentService) CreateDocument(ctx context.Context, doc domain.Document) (*domain.Document, error) {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.UploadDate.IsZero() {
		doc.UploadDate = time.Now()
	}
	if doc.Status == "" {
		doc.Status = "PENDING_UPLOAD" // Estado por defecto
	}

	err := s.repo.Save(ctx, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}
	return &doc, nil
}

// GetDocumentByID implementa ports.DocumentService.
func (s *documentService) GetDocumentByID(ctx context.Context, id string) (*domain.Document, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find document by ID: %w", err)
	}
	if doc == nil {
		return nil, nil // Documento no encontrado
	}
	return doc, nil
}

// GetAllDocuments implementa ports.DocumentService.
func (s *documentService) GetAllDocuments(ctx context.Context) ([]domain.Document, error) {
	docs, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all documents: %w", err)
	}
	return docs, nil
}

// UpdateDocument implementa ports.DocumentService.
func (s *documentService) UpdateDocument(ctx context.Context, id string, doc domain.Document) (*domain.Document, error) {
	existingDoc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing document for update: %w", err)
	}
	if existingDoc == nil {
		return nil, fmt.Errorf("document with ID %s not found", id)
	}

	// Actualizar solo los campos proporcionados
	if doc.FileName != "" {
		existingDoc.FileName = doc.FileName
	}
	if doc.S3Key != "" {
		existingDoc.S3Key = doc.S3Key
	}
	if doc.Status != "" {
		existingDoc.Status = doc.Status
	}
	if doc.OwnerID != "" {
		existingDoc.OwnerID = doc.OwnerID
	}
	// RequiresSignature, DocumentType, GroupName también podrían actualizarse si es necesario
	existingDoc.RequiresSignature = doc.RequiresSignature
	if doc.DocumentType != "" {
		existingDoc.DocumentType = doc.DocumentType
	}
	if doc.GroupName != "" {
		existingDoc.GroupName = doc.GroupName
	}

	err = s.repo.Update(ctx, existingDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}
	return existingDoc, nil
}

// DeleteDocument implementa ports.DocumentService.
func (s *documentService) DeleteDocument(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

func (s *documentService) RequestUploadURL(fileName string) (string, string, error) {
	return s.fileStorage.GeneratePresignedUploadURL(fileName)
}

// Asegurarse de que documentService implementa ports.DocumentService
var _ ports.DocumentService = (*documentService)(nil)
