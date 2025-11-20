package domain

import "time"

type Employee struct {
	ID       string
	Name     string
	Email    string
	Status   string
	LinkDate time.Time
	RoleID   string // AÑADIDO: ID del rol asignado al empleado
}
