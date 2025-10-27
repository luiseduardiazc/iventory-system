package handler

import (
	"fmt"
	"net/http"

	"inventory-system/internal/domain"
	"inventory-system/internal/service"

	"github.com/gin-gonic/gin"
)

// ReservationHandler maneja las peticiones HTTP para reservas
type ReservationHandler struct {
	reservationService *service.ReservationService
}

// NewReservationHandler crea un nuevo handler de reservas
func NewReservationHandler(reservationService *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{
		reservationService: reservationService,
	}
}

// CreateReservationRequest representa la petición para crear una reserva
type CreateReservationRequest struct {
	ProductID  string `json:"product_id" binding:"required"`
	StoreID    string `json:"store_id" binding:"required"`
	CustomerID string `json:"customer_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	TTLMinutes int    `json:"ttl_minutes" binding:"required,min=1,max=1440"` // Max 24 horas
}

// CreateReservation godoc
// @Summary Crear una nueva reserva de stock
// @Tags reservations
// @Accept json
// @Produce json
// @Param request body CreateReservationRequest true "Datos de la reserva"
// @Success 201 {object} domain.Reservation
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse "Stock insuficiente"
// @Router /reservations [post]
func (h *ReservationHandler) CreateReservation(c *gin.Context) {
	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// DEBUG: Log request
	fmt.Printf("DEBUG CreateReservation: ProductID=%s, StoreID=%s, CustomerID=%s, Quantity=%d, TTL=%d\n",
		req.ProductID, req.StoreID, req.CustomerID, req.Quantity, req.TTLMinutes)

	reservation, err := h.reservationService.CreateReservation(
		c.Request.Context(),
		req.ProductID,
		req.StoreID,
		req.CustomerID,
		req.Quantity,
		req.TTLMinutes,
	)

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// GetReservation godoc
// @Summary Obtener una reserva por ID
// @Tags reservations
// @Produce json
// @Param id path string true "ID de la reserva"
// @Success 200 {object} domain.Reservation
// @Failure 404 {object} ErrorResponse
// @Router /reservations/{id} [get]
func (h *ReservationHandler) GetReservation(c *gin.Context) {
	id := c.Param("id")

	reservation, err := h.reservationService.GetReservation(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// ConfirmReservation godoc
// @Summary Confirmar una reserva (procesa la venta)
// @Tags reservations
// @Produce json
// @Param id path string true "ID de la reserva"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /reservations/{id}/confirm [post]
func (h *ReservationHandler) ConfirmReservation(c *gin.Context) {
	id := c.Param("id")

	err := h.reservationService.ConfirmReservation(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Reservation confirmed successfully",
		"reservation_id": id,
		"status":         "CONFIRMED",
	})
}

// CancelReservation godoc
// @Summary Cancelar una reserva
// @Tags reservations
// @Produce json
// @Param id path string true "ID de la reserva"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /reservations/{id}/cancel [post]
func (h *ReservationHandler) CancelReservation(c *gin.Context) {
	id := c.Param("id")

	err := h.reservationService.CancelReservation(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Reservation cancelled successfully",
		"reservation_id": id,
		"status":         "CANCELLED",
	})
}

// GetPendingByStore godoc
// @Summary Obtener reservas pendientes de una tienda
// @Tags reservations
// @Produce json
// @Param storeId path string true "ID de la tienda"
// @Success 200 {array} domain.Reservation
// @Router /reservations/store/{storeId}/pending [get]
func (h *ReservationHandler) GetPendingByStore(c *gin.Context) {
	storeID := c.Param("storeId")

	reservations, err := h.reservationService.GetPendingByStore(c.Request.Context(), storeID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"store_id":     storeID,
		"reservations": reservations,
		"count":        len(reservations),
	})
}

// GetReservationsByProduct godoc
// @Summary Obtener reservas de un producto en una tienda
// @Tags reservations
// @Produce json
// @Param productId path string true "ID del producto"
// @Param storeId path string true "ID de la tienda"
// @Param status query string false "Filtrar por estado (PENDING, CONFIRMED, CANCELLED, EXPIRED)"
// @Success 200 {array} domain.Reservation
// @Router /reservations/product/{productId}/store/{storeId} [get]
func (h *ReservationHandler) GetReservationsByProduct(c *gin.Context) {
	productID := c.Param("productId")
	storeID := c.Param("storeId")
	statusStr := c.Query("status")

	var status *domain.ReservationStatus
	if statusStr != "" {
		s := domain.ReservationStatus(statusStr)
		status = &s
	}

	reservations, err := h.reservationService.GetReservationsByProduct(c.Request.Context(), productID, storeID, status)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id":   productID,
		"store_id":     storeID,
		"status":       statusStr,
		"reservations": reservations,
		"count":        len(reservations),
	})
}

// GetReservationStats godoc
// @Summary Obtener estadísticas de reservas
// @Tags reservations
// @Produce json
// @Success 200 {object} map[string]int
// @Router /reservations/stats [get]
func (h *ReservationHandler) GetReservationStats(c *gin.Context) {
	stats, err := h.reservationService.GetReservationStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
