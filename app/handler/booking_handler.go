package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	authmiddleware "sewasini/app/sewasini/middleware"
	"sewasini/models"
	repositoryroom "sewasini/repository/room"
	servicebooking "sewasini/service/booking"
)

type BookingHandler struct {
	service servicebooking.Service
}

func NewBookingHandler(service servicebooking.Service) *BookingHandler {
	return &BookingHandler{service: service}
}

func (h *BookingHandler) CreateBooking(c echo.Context) error {
	var req models.CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	booking, err := h.service.CreateBooking(c.Request().Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, servicebooking.ErrUserIDRequired),
		errors.Is(err, servicebooking.ErrInvalidBookingDate),
		errors.Is(err, servicebooking.ErrInvalidParticipantCount):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	case errors.Is(err, servicebooking.ErrRoomUnavailable):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, repositoryroom.ErrRoomNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}
