package handlers

import (
	"net/http"
	"strconv"
	"strings"

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

	sub, err := h.service.Create(&req)
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
	sub, err := h.service.GetByID(id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, sub)
}

// GetAll godoc
// @Summary List subscriptions
// @Description Returns subscription records with pagination.
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Page size (default 20, max 100)"
// @Param offset query int false "Rows to skip (default 0)"
// @Success 200 {array} models.Subscription
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) GetAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	subs, err := h.service.GetAll(limit, offset)
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

	sub, err := h.service.Update(id, &req)
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
	if err := h.service.Delete(id); err != nil {
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

	total, err := h.service.GetSumByPeriod(filter)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.SumResponse{Total: total})
}

func writeError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := err.Error()

	if strings.Contains(message, "invalid") {
		status = http.StatusBadRequest
	} else if strings.Contains(message, "not found") {
		status = http.StatusNotFound
	}

	c.JSON(status, models.ErrorResponse{Error: message})
}
