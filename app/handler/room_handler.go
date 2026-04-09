package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"sewasini/models"
	repositoryroom "sewasini/repository/room"
	serviceroom "sewasini/service/room"
)

type RoomHandler struct {
	service serviceroom.Service
}

func NewRoomHandler(service serviceroom.Service) *RoomHandler {
	return &RoomHandler{service: service}
}

func (h *RoomHandler) ListRooms(c echo.Context) error {
	filter, err := buildRoomFilter(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	rooms, err := h.service.List(c.Request().Context(), filter)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, rooms)
}

func (h *RoomHandler) GetRoomByID(c echo.Context) error {
	room, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) handleError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repositoryroom.ErrRoomNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
	}
}

func buildRoomFilter(c echo.Context) (models.RuanganFilter, error) {
	filter := models.RuanganFilter{
		Search:     strings.TrimSpace(c.QueryParam("search")),
		Kategori:   strings.TrimSpace(c.QueryParam("kategori")),
		KategoriID: strings.TrimSpace(c.QueryParam("kategori_id")),
		Kota:       strings.TrimSpace(c.QueryParam("kota")),
		Page:       1,
		Limit:      10,
	}
	if filter.Search == "" {
		filter.Search = strings.TrimSpace(c.QueryParam("q"))
	}
	if filter.Search == "" {
		filter.Search = strings.TrimSpace(c.QueryParam("nama"))
	}
	if filter.Kota == "" {
		filter.Kota = strings.TrimSpace(c.QueryParam("location"))
	}
	if filter.Kota == "" {
		filter.Kota = strings.TrimSpace(c.QueryParam("lokasi"))
	}

	if minHarga := strings.TrimSpace(c.QueryParam("min_harga")); minHarga != "" {
		value, err := strconv.ParseInt(minHarga, 10, 64)
		if err != nil {
			return filter, errors.New("min_harga must be a valid integer")
		}
		filter.MinHarga = value
	}

	if maxHarga := strings.TrimSpace(c.QueryParam("max_harga")); maxHarga != "" {
		value, err := strconv.ParseInt(maxHarga, 10, 64)
		if err != nil {
			return filter, errors.New("max_harga must be a valid integer")
		}
		filter.MaxHarga = value
	}

	if kapasitas := strings.TrimSpace(c.QueryParam("kapasitas")); kapasitas != "" {
		value, err := strconv.Atoi(kapasitas)
		if err != nil {
			return filter, errors.New("kapasitas must be a valid integer")
		}
		filter.Kapasitas = value
	}

	if page := strings.TrimSpace(c.QueryParam("page")); page != "" {
		value, err := strconv.Atoi(page)
		if err != nil {
			return filter, errors.New("page must be a valid integer")
		}
		filter.Page = value
	}

	if limit := strings.TrimSpace(c.QueryParam("limit")); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil {
			return filter, errors.New("limit must be a valid integer")
		}
		filter.Limit = value
	}

	if tanggal := strings.TrimSpace(c.QueryParam("tanggal_ketersediaan")); tanggal != "" {
		value, err := time.Parse("2006-01-02", tanggal)
		if err != nil {
			return filter, errors.New("tanggal_ketersediaan must use format YYYY-MM-DD")
		}
		filter.TanggalKetersediaan = &value
	}

	if filter.MinHarga < 0 || filter.MaxHarga < 0 {
		return filter, errors.New("harga filter cannot be negative")
	}
	if filter.Kapasitas < 0 {
		return filter, errors.New("kapasitas cannot be negative")
	}
	if filter.Page <= 0 {
		return filter, errors.New("page must be greater than 0")
	}
	if filter.Limit <= 0 {
		return filter, errors.New("limit must be greater than 0")
	}
	if filter.Limit > 100 {
		return filter, errors.New("limit cannot be greater than 100")
	}
	if filter.MaxHarga > 0 && filter.MinHarga > filter.MaxHarga {
		return filter, errors.New("min_harga cannot be greater than max_harga")
	}

	return filter, nil
}
