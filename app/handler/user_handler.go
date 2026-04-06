package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"sewasini/models"
	repositoryuser "sewasini/repository/user"
	serviceuser "sewasini/service/user"
)

type UserHandler struct {
	service serviceuser.Service
}

func NewUserHandler(service serviceuser.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterRoutes(group *echo.Group) {
	group.POST("/users", h.CreateUser)
	group.POST("/users/send-otp", h.SendOTP)
	group.POST("/users/verify-otp", h.VerifyOTP)
	group.GET("/users", h.ListUsers)
	group.GET("/users/:id", h.GetUserByID)
	group.PUT("/users/:id", h.UpdateUser)
	group.DELETE("/users/:id", h.DeleteUser)
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	user, err := h.service.RegisterUser(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	users, err := h.service.ListUsers(c.Request().Context())
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) SendOTP(c echo.Context) error {
	var req models.OTPSendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	if err := h.service.SendOTP(c.Request().Context(), req); err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "otp sent"})
}

func (h *UserHandler) VerifyOTP(c echo.Context) error {
	var req models.OTPVerifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	user, err := h.service.VerifyOTP(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserByID(c echo.Context) error {
	user, err := h.service.GetUserByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	user, err := h.service.UpdateUser(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	if err := h.service.DeleteUser(c.Request().Context(), c.Param("id")); err != nil {
		return h.handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repositoryuser.ErrUserNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	case errors.Is(err, serviceuser.ErrEmailAlreadyUsed):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, serviceuser.ErrInvalidOTP):
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": err.Error()})
	case errors.Is(err, serviceuser.ErrOTPExpiredOrNotFound):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	case errors.Is(err, serviceuser.ErrPhoneNumberRequired):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}
