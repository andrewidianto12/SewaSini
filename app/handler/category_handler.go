package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"sewasini/models"
	repositorycategory "sewasini/repository/category"
	servicecategory "sewasini/service/category"
)

type CategoryHandler struct {
	service servicecategory.Service
}

func NewCategoryHandler(service servicecategory.Service) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	var req models.CreateKategoriRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	category, err := h.service.CreateCategory(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) ListCategories(c echo.Context) error {
	categories, err := h.service.ListCategories(c.Request().Context())
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) GetCategoryByID(c echo.Context) error {
	category, err := h.service.GetCategoryByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	var req models.UpdateKategoriRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	category, err := h.service.UpdateCategory(c.Request().Context(), c.Param("id"), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	if err := h.service.DeleteCategory(c.Request().Context(), c.Param("id")); err != nil {
		return h.handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *CategoryHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repositorycategory.ErrCategoryNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	case errors.Is(err, repositorycategory.ErrCategoryAlreadyExists):
		return c.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
	case errors.Is(err, servicecategory.ErrCategoryNameRequired), errors.Is(err, servicecategory.ErrCategoryUpdateEmpty):
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}
