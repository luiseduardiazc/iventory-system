package domain

import "time"

// Product representa un producto en el catálogo
type Product struct {
	ID          string    `json:"id" db:"id"`
	SKU         string    `json:"sku" db:"sku"`                 // Código único del producto
	Name        string    `json:"name" db:"name"`               // Nombre del producto
	Description string    `json:"description" db:"description"` // Descripción
	Category    string    `json:"category" db:"category"`       // Categoría
	Price       float64   `json:"price" db:"price"`             // Precio
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// Validate verifica que el producto tenga datos válidos
func (p *Product) Validate() error {
	if p.SKU == "" {
		return &ValidationError{Field: "sku", Message: "SKU is required"}
	}
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "Name is required"}
	}
	if p.Price < 0 {
		return &ValidationError{Field: "price", Message: "Price cannot be negative"}
	}
	return nil
}
