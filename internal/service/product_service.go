package service

import (
	"context"
	"fmt"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

// ProductService maneja la lógica de negocio para productos
type ProductService struct {
	productRepo *repository.ProductRepository
	eventRepo   *repository.EventRepository
}

// NewProductService crea una nueva instancia del servicio
func NewProductService(
	productRepo *repository.ProductRepository,
	eventRepo *repository.EventRepository,
) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		eventRepo:   eventRepo,
	}
}

// CreateProduct crea un nuevo producto
func (s *ProductService) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	// Validar producto
	if err := product.Validate(); err != nil {
		return nil, err
	}

	// Verificar que no exista un producto con el mismo SKU
	existing, err := s.productRepo.GetBySKU(ctx, product.SKU)
	if err == nil && existing != nil {
		return nil, &domain.ConflictError{
			Message: fmt.Sprintf("product with SKU %s already exists", product.SKU),
		}
	}

	// Crear producto
	err = s.productRepo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProduct obtiene un producto por ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

// GetProductBySKU obtiene un producto por SKU
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return s.productRepo.GetBySKU(ctx, sku)
}

// ListProducts lista todos los productos con paginación
func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.productRepo.List(ctx, limit, offset)
}

// ListProductsByCategory lista productos de una categoría
func (s *ProductService) ListProductsByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.productRepo.ListByCategory(ctx, category, limit, offset)
}

// UpdateProduct actualiza un producto
func (s *ProductService) UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	// Validar producto
	if err := product.Validate(); err != nil {
		return nil, err
	}

	// Verificar que el producto existe
	existing, err := s.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	// Si cambia el SKU, verificar que no exista otro producto con ese SKU
	if existing.SKU != product.SKU {
		other, err := s.productRepo.GetBySKU(ctx, product.SKU)
		if err == nil && other != nil && other.ID != product.ID {
			return nil, &domain.ConflictError{
				Message: fmt.Sprintf("another product with SKU %s already exists", product.SKU),
			}
		}
	}

	// Actualizar
	err = s.productRepo.Update(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return s.productRepo.GetByID(ctx, product.ID)
}

// DeleteProduct elimina un producto
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	// TODO: Validar que no tenga stock en ninguna tienda antes de eliminar
	// Por ahora solo eliminamos
	return s.productRepo.Delete(ctx, id)
}

// CountProducts cuenta el total de productos
func (s *ProductService) CountProducts(ctx context.Context) (int, error) {
	return s.productRepo.Count(ctx)
}

// SearchProducts busca productos por nombre o descripción (simple)
func (s *ProductService) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error) {
	// Por ahora retorna todos - TODO: implementar búsqueda full-text
	return s.ListProducts(ctx, limit, offset)
}
