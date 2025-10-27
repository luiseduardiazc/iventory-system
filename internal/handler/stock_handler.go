package handler

import (
	"net/http"
	"strconv"

	"inventory-system/internal/service"

	"github.com/gin-gonic/gin"
)

// StockHandler maneja las peticiones HTTP para stock
type StockHandler struct {
	stockService *service.StockService
}

// NewStockHandler crea un nuevo handler de stock
func NewStockHandler(stockService *service.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

// GetStockByProductAndStore godoc
// @Summary Obtener stock de un producto en una tienda
// @Tags stock
// @Produce json
// @Param productId path string true "ID del producto"
// @Param storeId path string true "ID de la tienda"
// @Success 200 {object} domain.Stock
// @Failure 404 {object} ErrorResponse
// @Router /stock/{productId}/{storeId} [get]
func (h *StockHandler) GetStockByProductAndStore(c *gin.Context) {
	productID := c.Param("productId")
	storeID := c.Param("storeId")

	stock, err := h.stockService.GetStockByProductAndStore(c.Request.Context(), productID, storeID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stock)
}

// GetAllStockByProduct godoc
// @Summary Obtener stock de un producto en TODAS las tiendas
// @Tags stock
// @Produce json
// @Param productId path string true "ID del producto"
// @Success 200 {array} domain.Stock
// @Failure 404 {object} ErrorResponse
// @Router /stock/product/{productId} [get]
func (h *StockHandler) GetAllStockByProduct(c *gin.Context) {
	productID := c.Param("productId")

	stocks, err := h.stockService.GetAllStockByProduct(c.Request.Context(), productID)
	if err != nil {
		handleError(c, err)
		return
	}

	// Calcular disponibilidad total
	totalQuantity := 0
	totalReserved := 0
	for _, stock := range stocks {
		totalQuantity += stock.Quantity
		totalReserved += stock.Reserved
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id":      productID,
		"stores":          stocks,
		"total_quantity":  totalQuantity,
		"total_reserved":  totalReserved,
		"total_available": totalQuantity - totalReserved,
	})
}

// GetAllStockByStore godoc
// @Summary Obtener todo el stock de una tienda
// @Tags stock
// @Produce json
// @Param storeId path string true "ID de la tienda"
// @Success 200 {array} domain.Stock
// @Router /stock/store/{storeId} [get]
func (h *StockHandler) GetAllStockByStore(c *gin.Context) {
	storeID := c.Param("storeId")

	stocks, err := h.stockService.GetAllStockByStore(c.Request.Context(), storeID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"store_id": storeID,
		"items":    stocks,
		"count":    len(stocks),
	})
}

// UpdateStockRequest representa la petición para actualizar stock
type UpdateStockRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

// UpdateStock godoc
// @Summary Actualizar cantidad de stock
// @Tags stock
// @Accept json
// @Produce json
// @Param productId path string true "ID del producto"
// @Param storeId path string true "ID de la tienda"
// @Param request body UpdateStockRequest true "Nueva cantidad"
// @Success 200 {object} domain.Stock
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "Optimistic lock failure"
// @Router /stock/{productId}/{storeId} [put]
func (h *StockHandler) UpdateStock(c *gin.Context) {
	productID := c.Param("productId")
	storeID := c.Param("storeId")

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	stock, err := h.stockService.UpdateStock(c.Request.Context(), productID, storeID, req.Quantity)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stock)
}

// AdjustStockRequest representa la petición para ajustar stock
type AdjustStockRequest struct {
	Adjustment int `json:"adjustment" binding:"required"`
}

// AdjustStock godoc
// @Summary Ajustar stock (incrementar o decrementar)
// @Tags stock
// @Accept json
// @Produce json
// @Param productId path string true "ID del producto"
// @Param storeId path string true "ID de la tienda"
// @Param request body AdjustStockRequest true "Ajuste (positivo o negativo)"
// @Success 200 {object} domain.Stock
// @Failure 400 {object} ErrorResponse
// @Router /stock/{productId}/{storeId}/adjust [post]
func (h *StockHandler) AdjustStock(c *gin.Context) {
	productID := c.Param("productId")
	storeID := c.Param("storeId")

	var req AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	stock, err := h.stockService.AdjustStock(c.Request.Context(), productID, storeID, req.Adjustment)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stock)
}

// TransferStockRequest representa la petición para transferir stock
type TransferStockRequest struct {
	ProductID   string `json:"product_id" binding:"required"`
	FromStoreID string `json:"from_store_id" binding:"required"`
	ToStoreID   string `json:"to_store_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
}

// TransferStock godoc
// @Summary Transferir stock entre tiendas
// @Tags stock
// @Accept json
// @Produce json
// @Param request body TransferStockRequest true "Datos de transferencia"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "Stock insuficiente"
// @Router /stock/transfer [post]
func (h *StockHandler) TransferStock(c *gin.Context) {
	var req TransferStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	err := h.stockService.TransferStock(c.Request.Context(), req.ProductID, req.FromStoreID, req.ToStoreID, req.Quantity)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Stock transferred successfully",
		"product_id":    req.ProductID,
		"from_store_id": req.FromStoreID,
		"to_store_id":   req.ToStoreID,
		"quantity":      req.Quantity,
	})
}

// GetLowStockItems godoc
// @Summary Obtener productos con stock bajo
// @Tags stock
// @Produce json
// @Param threshold query int false "Umbral de stock bajo" default(10)
// @Success 200 {array} domain.Stock
// @Router /stock/low-stock [get]
func (h *StockHandler) GetLowStockItems(c *gin.Context) {
	threshold, _ := strconv.Atoi(c.DefaultQuery("threshold", "10"))

	stocks, err := h.stockService.GetLowStockItems(c.Request.Context(), threshold)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"threshold": threshold,
		"items":     stocks,
		"count":     len(stocks),
	})
}

// InitializeStockRequest representa la petición para inicializar stock
type InitializeStockRequest struct {
	ProductID       string `json:"product_id" binding:"required"`
	StoreID         string `json:"store_id" binding:"required"`
	InitialQuantity int    `json:"initial_quantity" binding:"required,min=0"`
}

// InitializeStock godoc
// @Summary Inicializar stock de un producto en una tienda
// @Tags stock
// @Accept json
// @Produce json
// @Param request body InitializeStockRequest true "Datos de inicialización"
// @Success 201 {object} domain.Stock
// @Failure 400 {object} ErrorResponse
// @Router /stock [post]
func (h *StockHandler) InitializeStock(c *gin.Context) {
	var req InitializeStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	stock, err := h.stockService.InitializeStock(c.Request.Context(), req.ProductID, req.StoreID, req.InitialQuantity)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, stock)
}

// CheckAvailability godoc
// @Summary Verificar disponibilidad de stock
// @Tags stock
// @Produce json
// @Param productId path string true "ID del producto"
// @Param storeId path string true "ID de la tienda"
// @Param quantity query int true "Cantidad requerida"
// @Success 200 {object} map[string]interface{}
// @Router /stock/{productId}/{storeId}/availability [get]
func (h *StockHandler) CheckAvailability(c *gin.Context) {
	productID := c.Param("productId")
	storeID := c.Param("storeId")
	quantity, _ := strconv.Atoi(c.Query("quantity"))

	if quantity <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid quantity",
			Message: "Quantity must be positive",
		})
		return
	}

	available, err := h.stockService.CheckAvailability(c.Request.Context(), productID, storeID, quantity)
	if err != nil {
		handleError(c, err)
		return
	}

	actualAvailable, _ := h.stockService.GetAvailableStock(c.Request.Context(), productID, storeID)

	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"store_id":   storeID,
		"requested":  quantity,
		"available":  actualAvailable,
		"sufficient": available,
	})
}
