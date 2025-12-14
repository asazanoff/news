// internal/handlers/news.go
package handlers

import (
	"service/internal/apperrors"
	"service/internal/models"
	"service/internal/service"
	"strconv"

	"service/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type NewsHandler struct {
	service service.INewsService
	log     *logger.Logger
}

func NewNewsHandler(service service.INewsService, log *logger.Logger) NewsHandler {
	return NewsHandler{
		service: service,
		log:     log,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type SuccessResponse struct {
	Success bool
}

type SuccessResponseCreate struct {
	Success bool
	Id      int64
}

type NewsListsResponse struct {
	Success bool
	News    []models.NewsWithCategories
}

func (h *NewsHandler) CreateNews(c *fiber.Ctx) error {
	var reqForm models.NewsCreateForm
	if err := c.BodyParser(&reqForm); err != nil {
		return apperrors.NewBadRequest("Invalid request body")
	}

	reqForm.Normalize()
	if err := reqForm.Validate(); err != nil {
		return apperrors.NewValidation(err.Error())
	}

	id, err := h.service.CreateNews(reqForm)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponseCreate{
		Success: true,
		Id:      id,
	})
}

func (h *NewsHandler) EditNews(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return apperrors.NewBadRequest("Invalid ID format")
	}

	var editForm models.NewsEditForm
	if err = c.BodyParser(&editForm); err != nil {
		return apperrors.NewBadRequest("Invalid request body")
	}

	editForm.Normalize()
	if err = editForm.Validate(); err != nil {
		return apperrors.NewValidation(err.Error())
	}

	if err = h.service.EditNews(int64(id), editForm); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
	})
}

func (h *NewsHandler) ListNews(c *fiber.Ctx) error {
	limit, err := strconv.ParseInt(c.Query("limit", "10"), 10, 64)
	if err != nil {
		return apperrors.NewBadRequest("limit must be a valid number")
	}

	offset, err := strconv.ParseInt(c.Query("offset", "0"), 10, 64)
	if err != nil {
		return apperrors.NewBadRequest("offset must be a valid number")
	}

	if err = ValidatePaginationParams(limit, offset); err != nil {
		return err
	}

	newsList, err := h.service.ListNews(limit, offset)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(NewsListsResponse{Success: true, News: newsList})
}
