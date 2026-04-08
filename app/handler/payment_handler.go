package handler

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"

	authmiddleware "sewasini/app/sewasini/middleware"
	"sewasini/models"
	repositorybooking "sewasini/repository/booking"
	repositorytransaction "sewasini/repository/transaction"
	repositoryuser "sewasini/repository/user"
	servicetransaction "sewasini/service/transaction"
)

type PaymentHandler struct {
	service servicetransaction.Service
}

func NewPaymentHandler(service servicetransaction.Service) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) CreatePayment(c echo.Context) error {
	var req models.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	userID, _ := c.Get(authmiddleware.ContextUserIDKey).(string)
	resp, err := h.service.CreatePayment(c.Request().Context(), userID, req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *PaymentHandler) PaymentCallback(c echo.Context) error {
	expectedToken := strings.TrimSpace(os.Getenv("XENDIT_CALLBACK_TOKEN"))
	if expectedToken != "" && c.Request().Header.Get("x-callback-token") != expectedToken {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid callback token"})
	}

	var req models.XenditCallbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	req.WebhookID = strings.TrimSpace(c.Request().Header.Get("webhook-id"))

	err := h.service.HandleCallback(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, servicetransaction.ErrCallbackIgnored) {
			return c.JSON(http.StatusOK, map[string]string{"message": "callback ignored"})
		}
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "payment callback processed"})
}

func (h *PaymentHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, servicetransaction.ErrUserIDRequired),
		errors.Is(err, servicetransaction.ErrPaymentMethodRequired):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	case errors.Is(err, servicetransaction.ErrBookingOwnership):
		return c.JSON(http.StatusForbidden, map[string]string{"message": err.Error()})
	case errors.Is(err, servicetransaction.ErrBookingAlreadyPaid),
		errors.Is(err, servicetransaction.ErrBookingNotPayable):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, repositorybooking.ErrBookingNotFound),
		errors.Is(err, repositorytransaction.ErrTransactionNotFound),
		errors.Is(err, repositoryuser.ErrUserNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}
