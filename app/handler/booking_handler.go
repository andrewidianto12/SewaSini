package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	authmiddleware "sewasini/app/sewasini/middleware"
	"sewasini/models"
	repositorybooking "sewasini/repository/booking"
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

func (h *BookingHandler) ListBookings(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	bookings, err := h.service.ListUserBookings(c.Request().Context(), userID)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) GetBookingByID(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	booking, err := h.service.GetUserBookingByID(c.Request().Context(), userID, c.Param("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, booking)
}

func (h *BookingHandler) UpdateBooking(c echo.Context) error {
	var req models.UpdateBookingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}

	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	booking, err := h.service.UpdateBooking(c.Request().Context(), userID, c.Param("id"), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, booking)
}

func (h *BookingHandler) CancelBooking(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	if err := h.service.CancelBooking(c.Request().Context(), userID, c.Param("id")); err != nil {
		return h.handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *BookingHandler) GetBookingStatus(c echo.Context) error {
	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	booking, err := h.service.GetUserBookingByID(c.Request().Context(), userID, c.Param("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":             booking.ID,
		"status":         booking.Status,
		"payment_status": booking.PaymentStatus,
	})
}

func (h *BookingHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, servicebooking.ErrUserIDRequired),
		errors.Is(err, servicebooking.ErrInvalidBookingDate),
		errors.Is(err, servicebooking.ErrInvalidParticipantCount):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	case errors.Is(err, servicebooking.ErrBookingOwnership):
		return c.JSON(http.StatusForbidden, map[string]string{"message": err.Error()})
	case errors.Is(err, servicebooking.ErrRoomUnavailable):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, servicebooking.ErrBookingNotEditable):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, repositorybooking.ErrBookingNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	case errors.Is(err, repositoryroom.ErrRoomNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}
