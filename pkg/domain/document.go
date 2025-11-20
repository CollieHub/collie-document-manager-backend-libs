package domain

import "time"

type Document struct {
	ID                string    `json:"id"`
	FileName          string    `json:"fileName"`
	S3Key             string    `json:"s3Key"`
	UploadDate        time.Time `json:"uploadDate"`
	Status            string    `json:"status"`  // Ej: "PENDING_UPLOAD", "UPLOADED", "PROCESSED"
	OwnerID           string    `json:"ownerId"` // ID del empleado o empresa
	RequiresSignature bool      `json:"requiresSignature"`
	DocumentType      string    `json:"documentType"`
	GroupName         string    `json:"groupName"`
	Recipient         string    `json:"recipient"`
}
