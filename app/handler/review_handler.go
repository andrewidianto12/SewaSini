package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	authmiddleware "sewasini/app/sewasini/middleware"
	"sewasini/models"
	repositoryreview "sewasini/repository/review"
	servicereview "sewasini/service/review"
)

type ReviewHandler struct {
	service servicereview.Service
}

func NewReviewHandler(service servicereview.Service) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) CreateReview(c echo.Context) error {
	var req models.CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	userRole, _ := c.Get(authmiddleware.ContextUserRoleKey).(string)
	review, err := h.service.CreateReview(c.Request().Context(), userID, userRole, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, review)
}

func (h *ReviewHandler) ListReviews(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	reviews, err := h.service.ListReviews(c.Request().Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, reviews)
}

func (h *ReviewHandler) GetReviewByID(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	review, err := h.service.GetReviewByID(c.Request().Context(), userID, c.Param("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, review)
}

func (h *ReviewHandler) ListReviewsByRoomID(c echo.Context) error {
	roomID := c.Param("ruangan_id")
	if roomID == "" {
		roomID = c.Param("id")
	}

	reviews, err := h.service.ListReviewsByRoomID(c.Request().Context(), roomID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, reviews)
}

func (h *ReviewHandler) UpdateReview(c echo.Context) error {
	var req models.UpdateReviewRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	review, err := h.service.UpdateReview(c.Request().Context(), userID, c.Param("id"), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, review)
}

func (h *ReviewHandler) DeleteReview(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	if err := h.service.DeleteReview(c.Request().Context(), userID, c.Param("id")); err != nil {
		return h.handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ReviewHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repositoryreview.ErrReviewNotFound), errors.Is(err, repositoryreview.ErrBookingNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	case errors.Is(err, repositoryreview.ErrReviewAlreadyExists):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, servicereview.ErrUserIDRequired), errors.Is(err, servicereview.ErrRoomIDRequired), errors.Is(err, servicereview.ErrReviewUpdateEmpty), errors.Is(err, servicereview.ErrBookingMismatch):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	case errors.Is(err, servicereview.ErrInsufficientRole), errors.Is(err, servicereview.ErrForbiddenReviewAccess):
		return c.JSON(http.StatusForbidden, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}
