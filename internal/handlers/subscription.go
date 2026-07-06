package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/t0fox/subscription-aggregator-api/internal/models"
	"github.com/t0fox/subscription-aggregator-api/internal/service"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(service *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// Create godoc
// @Summary Create subscription
// @Description Creates a new online subscription record.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.SubscriptionCreateRequest true "Subscription payload"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req models.SubscriptionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Returns one subscription record by UUID.
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription UUID"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, sub)
}

// GetAll godoc
// @Summary List subscriptions
// @Description Returns a paginated list of subscription records.
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Page size (default 50)"
// @Param offset query int false "Rows to skip (default 0)"
// @Success 200 {array} models.Subscription
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) GetAll(c *gin.Context) {
	limit := parseIntQuery(c, "limit", 50)
	offset := parseIntQuery(c, "offset", 0)

	subs, err := h.service.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, subs)
}

// Update godoc
// @Summary Update subscription
// @Description Updates subscription fields by UUID.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription UUID"
// @Param subscription body models.SubscriptionUpdateRequest true "Subscription update payload"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.SubscriptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, sub)
}

// Delete godoc
// @Summary Delete subscription
// @Description Deletes a subscription record by UUID.
// @Tags subscriptions
// @Param id path string true "Subscription UUID"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetSum godoc
// @Summary Calculate subscriptions sum
// @Description Calculates total subscription cost for a selected period with optional user and service filters.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param filter body models.SubscriptionSumRequest true "Sum filter"
// @Success 200 {object} models.SumResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/sum [post]
func (h *SubscriptionHandler) GetSum(c *gin.Context) {
	var req models.SubscriptionSumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate := req.StartDate
	endDate := req.EndDate
	filter := &models.SubscriptionFilter{
		UserID:      req.UserID,
		ServiceName: req.ServiceName,
		StartDate:   &startDate,
		EndDate:     &endDate,
	}

	total, err := h.service.GetSumByPeriod(c.Request.Context(), filter)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.SumResponse{Total: total})
}

// parseIntQuery reads a non-negative int query param, falling back to def.
func parseIntQuery(c *gin.Context, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return def
	}
	return v
}

// writeError maps domain errors to HTTP status codes via errors.Is.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, models.ErrValidation):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
	case errors.Is(err, models.ErrNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
	}
}
