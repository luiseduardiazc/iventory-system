package domain

import "fmt"

// DomainError es la interfaz base para todos los errores de dominio
type DomainError interface {
	error
	Code() string
}

// ValidationError representa un error de validación
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Code() string {
	return "VALIDATION_ERROR"
}

// NotFoundError representa un recurso no encontrado
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

func (e *NotFoundError) Code() string {
	return "NOT_FOUND"
}

// ConflictError representa un conflicto (ej. optimistic lock)
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

func (e *ConflictError) Code() string {
	return "CONFLICT"
}

// InsufficientStockError representa stock insuficiente
type InsufficientStockError struct {
	ProductID string
	StoreID   string
	Available int
	Requested int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for product %s in store %s: available=%d, requested=%d",
		e.ProductID, e.StoreID, e.Available, e.Requested)
}

func (e *InsufficientStockError) Code() string {
	return "INSUFFICIENT_STOCK"
}

// ReservationExpiredError representa una reserva expirada
type ReservationExpiredError struct {
	ReservationID string
	ExpiresAt     string
}

func (e *ReservationExpiredError) Error() string {
	return fmt.Sprintf("reservation %s expired at %s", e.ReservationID, e.ExpiresAt)
}

func (e *ReservationExpiredError) Code() string {
	return "RESERVATION_EXPIRED"
}

// InvalidStateError representa una transición de estado inválida
type InvalidStateError struct {
	CurrentState    string
	AttemptedAction string
}

func (e *InvalidStateError) Error() string {
	return fmt.Sprintf("cannot %s from state %s", e.AttemptedAction, e.CurrentState)
}

func (e *InvalidStateError) Code() string {
	return "INVALID_STATE"
}

// UnauthorizedError representa un error de autenticación
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

func (e *UnauthorizedError) Code() string {
	return "UNAUTHORIZED"
}

// ForbiddenError representa un error de autorización (permisos)
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

func (e *ForbiddenError) Code() string {
	return "FORBIDDEN"
}
