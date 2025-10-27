package handler

import (
	"net/http"
	"strconv"

	"inventory-system/internal/domain"
	"inventory-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler maneja las peticiones HTTP para productos
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler crea un nuevo handler de productos
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct godoc
// @Summary Crear un nuevo producto
// @Tags products
// @Accept json
// @Produce json
// @Param product body domain.Product true "Producto a crear"
// @Success 201 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product domain.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Generar ID con UUID real si no viene
	if product.ID == "" {
		product.ID = uuid.New().String()
	}

	created, err := h.productService.CreateProduct(c.Request.Context(), &product)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetProduct godoc
// @Summary Obtener un producto por ID
// @Tags products
// @Produce json
// @Param id path string true "ID del producto"
// @Success 200 {object} domain.Product
// @Failure 404 {object} ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// ListProducts godoc
// @Summary Listar productos con paginación
// @Tags products
// @Produce json
// @Param limit query int false "Límite de resultados" default(10)
// @Param offset query int false "Offset para paginación" default(0)
// @Param category query string false "Filtrar por categoría"
// @Success 200 {array} domain.Product
// @Router /products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	category := c.Query("category")

	var products []*domain.Product
	var err error

	if category != "" {
		products, err = h.productService.ListProductsByCategory(c.Request.Context(), category, limit, offset)
	} else {
		products, err = h.productService.ListProducts(c.Request.Context(), limit, offset)
	}

	if err != nil {
		handleError(c, err)
		return
	}

	// Get total count
	total, _ := h.productService.CountProducts(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"data":   products,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateProduct godoc
// @Summary Actualizar un producto
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "ID del producto"
// @Param product body domain.Product true "Datos del producto"
// @Success 200 {object} domain.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var product domain.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	product.ID = id

	updated, err := h.productService.UpdateProduct(c.Request.Context(), &product)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteProduct godoc
// @Summary Eliminar un producto
// @Tags products
// @Param id path string true "ID del producto"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := h.productService.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProductBySKU godoc
// @Summary Buscar producto por SKU
// @Tags products
// @Produce json
// @Param sku path string true "SKU del producto"
// @Success 200 {object} domain.Product
// @Failure 404 {object} ErrorResponse
// @Router /products/sku/{sku} [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")

	product, err := h.productService.GetProductBySKU(c.Request.Context(), sku)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// ErrorResponse representa una respuesta de error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// handleError maneja errores de dominio y los convierte en respuestas HTTP
func handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *domain.NotFoundError:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: e.Error(),
		})
	case *domain.ValidationError:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation Error",
			Message: e.Error(),
		})
	case *domain.ConflictError:
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Conflict",
			Message: e.Error(),
		})
	case *domain.InsufficientStockError:
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Insufficient Stock",
			Message: e.Error(),
		})
	case *domain.UnauthorizedError:
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: e.Error(),
		})
	case *domain.ForbiddenError:
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "Forbidden",
			Message: e.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		})
	}
}
