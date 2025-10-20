package handlers

import (
	"errors"
	"net/http"

	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/request"
	"emplacc-api/api/v1/middleware"
	"emplacc-api/api/v1/presenter"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"

	"github.com/labstack/echo/v4"
)

var (
	_ request.CreateSubscription
)

type SubscriptionHandler struct {
	service ports.SubscriptionService
}

func NewSubscriptionHandler(service ports.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// @Summary Register Subscription Routes
// @Description Register routes for subscription management
// @Tags Subscriptions
func RegisterSubscriptionRoutes(group *echo.Group, service ports.SubscriptionService) {
	handler := NewSubscriptionHandler(service)

	group.GET("/subscriptions", handler.ListSubscriptions)
	group.GET("/subscriptions/:id", handler.GetSubscription)
	group.GET("/users/:user_id/subscriptions", handler.ListSubscriptionsByUser)
	group.GET("/subscriptions/targets/:subscription_id", handler.ListSubscriptionsByTarget)
	group.POST("/subscriptions", handler.CreateSubscription)
	group.DELETE("/subscriptions/:id", handler.DeleteSubscription)
}

// @Summary List Subscriptions
// @Description Retrieve a paginated list of subscriptions
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.SubscriptionsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c echo.Context) error {
	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	result, err := h.service.ListSubscriptions(c.Request().Context(), params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("subscriptions_fetch_failed", "failed to list subscriptions"))
	}

	pagination, payload := presenter.MapSubscriptionsPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Get Subscription
// @Description Retrieve a subscription by its ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid subscription identifier"))
	}

	sub, err := h.service.GetSubscription(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("subscription_not_found", "subscription not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("subscriptions_fetch_failed", "failed to get subscription"))
	}

	return respondSuccess(c, http.StatusOK, presenter.ToSubscriptionDTO(sub))
}

// @Summary List Subscriptions By User
// @Description Retrieve a paginated list of subscriptions for a specific user
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.SubscriptionsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{user_id}/subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptionsByUser(c echo.Context) error {
	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid user identifier"))
	}

	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	result, err := h.service.ListSubscriptionsByUser(c.Request().Context(), userID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("subscriptions_fetch_failed", "failed to list subscriptions for user"))
	}

	pagination, payload := presenter.MapSubscriptionsPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary List Subscriptions By Target
// @Description Retrieve a paginated list of subscriptions for a specific target
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription_id path string true "Subscription Target ID"
// @Param type_id query int8 false "Type ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.SubscriptionsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /subscriptions/targets/{subscription_id} [get]
func (h *SubscriptionHandler) ListSubscriptionsByTarget(c echo.Context) error {
	targetID, err := parseUUID(c.Param("subscription_id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid subscription target identifier"))
	}

	typeParam := c.QueryParam("type_id")
	typeID, err := parseOptionalInt8(typeParam)
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_query", "type_id must be a valid 8-bit integer"))
	}

	pageNum, size := resolvePagination(c.QueryParam("page"), c.QueryParam("page_size"), 1, 20)
	params := ports.PaginationParams{Page: pageNum, PageSize: size}

	result, err := h.service.ListSubscriptionsByTarget(c.Request().Context(), targetID, typeID, params)
	if err != nil {
		return respondError(c, http.StatusInternalServerError, dto.NewError("subscriptions_fetch_failed", "failed to list subscriptions by target"))
	}

	pagination, payload := presenter.MapSubscriptionsPage(result)
	return respondPaginated(c, http.StatusOK, payload, pagination)
}

// @Summary Create Subscription
// @Description Create a new subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription body request.CreateSubscription true "Subscription creation payload"
// @Success 201 {object} dto.SubscriptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c echo.Context) error {
	req, err := middleware.BindAndValidate(c, middleware.ValidateCreateSubscriptionPayload)
	if err != nil {
		return middleware.RespondValidationError(c, err)
	}

	input := ports.CreateSubscriptionInput{
		UserID:         req.UserUUID,
		SubscriptionID: req.SubscriptionUUID,
		TypeID:         req.TypeID,
	}

	sub, err := h.service.CreateSubscription(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			return respondError(c, http.StatusBadRequest, dto.NewError("invalid_payload", err.Error()))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("subscription_create_failed", "failed to create subscription"))
	}

	return respondSuccess(c, http.StatusCreated, presenter.ToSubscriptionDTO(sub))
}

// @Summary Delete Subscription
// @Description Delete a subscription by its ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c echo.Context) error {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		return respondError(c, http.StatusBadRequest, dto.NewError("invalid_id", "invalid subscription identifier"))
	}

	if err := h.service.DeleteSubscription(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return respondError(c, http.StatusNotFound, dto.NewError("subscription_not_found", "subscription not found"))
		}
		return respondError(c, http.StatusInternalServerError, dto.NewError("subscription_delete_failed", "failed to delete subscription"))
	}

	return c.NoContent(http.StatusNoContent)
}
