package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"service/internal/apperrors"
	"service/internal/handlers/errors"
	"strings"

	"service/internal/models"
	"service/internal/service/mocks"
	customLog "service/pkg/logger"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testLogger = func() *customLog.Logger {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	return &customLog.Logger{Logger: log}
}()

func setupService(t *testing.T) *mocks.INewsService {
	mockService := new(mocks.INewsService)

	t.Cleanup(func() {
		mockService.AssertExpectations(t)
	})

	return mockService
}

func TestCreateNews(t *testing.T) {
	var createdNewsId int64 = 1
	title := "Amazing news"
	content := "This is really amazing news"
	categories := []int64{1, 2, 3}

	validRequestData := []struct {
		name       string
		createForm models.NewsCreateForm
	}{
		{
			name: "without categories",
			createForm: models.NewsCreateForm{
				Title:   title,
				Content: content,
			},
		},
		{
			name: "with categories",
			createForm: models.NewsCreateForm{
				Title:      title,
				Content:    content,
				Categories: &categories,
			},
		},
	}

	for _, rd := range validRequestData {
		t.Run(fmt.Sprintf("Success_%s", rd.name), func(t *testing.T) {
			requestBody, err := json.Marshal(rd.createForm)
			if err != nil {
				t.Fatal(err)
			}

			mockService := setupService(t)
			mockService.On("CreateNews", rd.createForm).Return(createdNewsId, nil)

			handler := NewNewsHandler(mockService, testLogger)
			app := fiber.New()
			app.Post("/create", handler.CreateNews)

			req := httptest.NewRequest("POST", "/create", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			var response SuccessResponseCreate
			json.Unmarshal(body, &response)

			assert.True(t, response.Success)
			assert.Equal(t, createdNewsId, response.Id)
		})
	}

	t.Run("FailedEmptyBody", func(t *testing.T) {
		mockService := setupService(t)
		handler := NewNewsHandler(mockService, testLogger)

		app := fiber.New(fiber.Config{
			ErrorHandler: errors.ErrorHandler(testLogger),
		})
		app.Post("/create", handler.CreateNews)

		req := httptest.NewRequest("POST", "/create", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertNotCalled(t, "CreateNews")
	})

	invalidRequestData := []struct {
		name     string
		body     string
		errorMsg string
	}{
		{
			name:     "title is empty",
			body:     `{"title":"","content":"Some content"}`,
			errorMsg: "title length must be between 1 and 255",
		},
		{
			name:     "title too long",
			body:     fmt.Sprintf(`{"title":"%s","content":"Content"}`, strings.Repeat("a", 256)),
			errorMsg: "title length must be between 1 and 255",
		},
		{
			name:     "content is empty",
			body:     `{"title":"Title","content":""}`,
			errorMsg: "content length must be greater 1",
		},
	}

	for _, rd := range invalidRequestData {
		t.Run(fmt.Sprintf("FailedInvalidData_%s", rd.name), func(t *testing.T) {
			mockService := setupService(t)
			handler := NewNewsHandler(mockService, testLogger)

			app := fiber.New(fiber.Config{
				ErrorHandler: errors.ErrorHandler(testLogger),
			})
			app.Post("/create", handler.CreateNews)

			req := httptest.NewRequest("POST", "/create", bytes.NewReader([]byte(rd.body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), rd.errorMsg)
			mockService.AssertNotCalled(t, "CreateNews")
		})
	}

	t.Run("FailedInternalError", func(t *testing.T) {
		createForm := models.NewsCreateForm{
			Title:   title,
			Content: content,
		}
		requestBody, _ := json.Marshal(createForm)

		mockService := setupService(t)
		mockService.On("CreateNews", createForm).Return(int64(0), apperrors.NewInternal("database error"))

		handler := NewNewsHandler(mockService, testLogger)

		app := fiber.New(fiber.Config{
			ErrorHandler: errors.ErrorHandler(testLogger),
		})
		app.Post("/create", handler.CreateNews)

		req := httptest.NewRequest("POST", "/create", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "database error")
	})
}

func TestListNews(t *testing.T) {
	newsList := []models.NewsWithCategories{
		{
			News: models.News{
				ID:      1,
				Title:   "News 1",
				Content: "Content 1",
			},
			Categories: []int64{1, 2},
		},
		{
			News: models.News{
				ID:      2,
				Title:   "News 2",
				Content: "Content 2",
			},
			Categories: []int64{},
		},
	}

	t.Run("Success", func(t *testing.T) {
		mockService := setupService(t)
		mockService.On("ListNews", int64(10), int64(0)).Return(newsList, nil)

		handler := NewNewsHandler(mockService, testLogger)
		app := fiber.New()
		app.Get("/list", handler.ListNews)

		req := httptest.NewRequest("GET", "/list?limit=10&offset=0", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response NewsListsResponse
		json.Unmarshal(body, &response)

		assert.True(t, response.Success)
		assert.Equal(t, newsList, response.News)
	})

	t.Run("SuccessWithoutLimitAndOffset", func(t *testing.T) {
		mockService := setupService(t)
		mockService.On("ListNews", int64(10), int64(0)).Return(newsList, nil)

		handler := NewNewsHandler(mockService, testLogger)
		app := fiber.New()
		app.Get("/list", handler.ListNews)

		req := httptest.NewRequest("GET", "/list", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var response NewsListsResponse
		json.Unmarshal(body, &response)

		assert.True(t, response.Success)
		assert.Equal(t, newsList, response.News)
	})

	invalidPaginationData := []struct {
		name     string
		url      string
		errorMsg string
	}{
		{
			name:     "limit is negative",
			url:      "/list?limit=-1&offset=0",
			errorMsg: "limit must be greater or equal 1",
		},
		{
			name:     "limit is too large",
			url:      "/list?limit=101&offset=0",
			errorMsg: "limit must be less or equal to 100",
		},
		{
			name:     "offset is negative",
			url:      "/list?limit=10&offset=-1",
			errorMsg: "offset cannot be negative",
		},
		{
			name:     "limit is not a number",
			url:      "/list?limit=abc&offset=0",
			errorMsg: "limit must be a valid number",
		},
		{
			name:     "offset is not a number",
			url:      "/list?limit=10&offset=abc",
			errorMsg: "offset must be a valid number",
		},
	}

	for _, rd := range invalidPaginationData {
		t.Run(fmt.Sprintf("FailedInvalidPagination_%s", rd.name), func(t *testing.T) {
			mockService := setupService(t)
			handler := NewNewsHandler(mockService, testLogger)

			app := fiber.New(fiber.Config{
				ErrorHandler: errors.ErrorHandler(testLogger),
			})
			app.Get("/list", handler.ListNews)

			req := httptest.NewRequest("GET", rd.url, nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), rd.errorMsg)
			mockService.AssertNotCalled(t, "CreateNews")

		})
	}
}

func TestEditNews(t *testing.T) {
	var newsId int64 = 10
	newTitle := "Updated Title"
	newContent := "Updated Content"
	newCategories := []int64{1, 2}

	validEditData := []struct {
		name     string
		editForm models.NewsEditForm
	}{
		{
			name: "update title only",
			editForm: models.NewsEditForm{
				Title: &newTitle,
			},
		},
		{
			name: "update content only",
			editForm: models.NewsEditForm{
				Content: &newContent,
			},
		},
		{
			name: "update categories only",
			editForm: models.NewsEditForm{
				Categories: &newCategories,
			},
		},
		{
			name: "update all fields",
			editForm: models.NewsEditForm{
				Title:      &newTitle,
				Content:    &newContent,
				Categories: &newCategories,
			},
		},
	}

	for _, rd := range validEditData {
		t.Run(fmt.Sprintf("Success_%s", rd.name), func(t *testing.T) {
			requestBody, err := json.Marshal(rd.editForm)
			if err != nil {
				t.Fatal(err)
			}

			mockService := setupService(t)
			mockService.On("EditNews", newsId, rd.editForm).Return(nil)

			handler := NewNewsHandler(mockService, testLogger)
			app := fiber.New()
			app.Post("/edit/:id", handler.EditNews)

			req := httptest.NewRequest("POST", fmt.Sprintf("/edit/%d", newsId), bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			var response SuccessResponse
			json.Unmarshal(body, &response)

			assert.True(t, response.Success)
		})
	}

	t.Run("SuccessNoFieldsToUpdate", func(t *testing.T) {
		editForm := models.NewsEditForm{}
		requestBody, _ := json.Marshal(editForm)

		mockService := setupService(t)
		handler := NewNewsHandler(mockService, testLogger)

		app := fiber.New(fiber.Config{
			ErrorHandler: errors.ErrorHandler(testLogger),
		})
		app.Post("/edit/:id", handler.EditNews)

		req := httptest.NewRequest("POST", fmt.Sprintf("/edit/%d", newsId), bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "body have to contains news param")
		mockService.AssertNotCalled(t, "EditNews")
	})

	validNewsDataWithSpace := []struct {
		name            string
		title           string
		content         string
		expectedTitle   string
		expectedContent string
	}{
		{
			name:            "Title with space",
			title:           "   Title with spaces ",
			content:         "Content",
			expectedTitle:   "Title with spaces",
			expectedContent: "Content",
		},
		{
			name:            "Content with space",
			title:           "Title",
			content:         "   Content with spaces ",
			expectedTitle:   "Title",
			expectedContent: "Content with spaces",
		},
		{
			name:            "Title and content with spaces",
			title:           "   Title with spaces ",
			content:         "   Content with spaces ",
			expectedTitle:   "Title with spaces",
			expectedContent: "Content with spaces",
		},
	}

	for _, vn := range validNewsDataWithSpace {
		t.Run(fmt.Sprintf("Success%s", vn.name), func(t *testing.T) {
			editForm := models.NewsEditForm{
				Title:   &vn.title,
				Content: &vn.content,
			}
			editFormExpected := models.NewsEditForm{
				Title:   &vn.expectedTitle,
				Content: &vn.expectedContent,
			}
			requestBody, _ := json.Marshal(editForm)

			mockService := setupService(t)
			mockService.On("EditNews", newsId, editFormExpected).Return(nil)

			handler := NewNewsHandler(mockService, testLogger)
			app := fiber.New()
			app.Post("/edit/:id", handler.EditNews)

			req := httptest.NewRequest("POST", fmt.Sprintf("/edit/%d", newsId), bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			mockService.AssertNotCalled(t, "EditNews")
		})
	}

	t.Run("FailedNewsNotFound", func(t *testing.T) {
		editForm := models.NewsEditForm{
			Title: &newTitle,
		}
		requestBody, _ := json.Marshal(editForm)

		mockService := setupService(t)
		mockService.On("EditNews", newsId, editForm).Return(apperrors.NewNotFound("News not found"))

		handler := NewNewsHandler(mockService, testLogger)

		app := fiber.New(fiber.Config{
			ErrorHandler: errors.ErrorHandler(testLogger),
		})
		app.Post("/edit/:id", handler.EditNews)

		req := httptest.NewRequest("POST", fmt.Sprintf("/edit/%d", newsId), bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "News not found")
	})

	invalidIdData := []struct {
		name string
		id   string
	}{
		{
			name: "id is not a number",
			id:   "abc",
		},
		{
			name: "id is negative",
			id:   "-1",
		},
	}

	for _, rd := range invalidIdData {
		t.Run(fmt.Sprintf("FailedInvalidId_%s", rd.name), func(t *testing.T) {
			editForm := models.NewsEditForm{
				Title: &newTitle,
			}
			requestBody, _ := json.Marshal(editForm)

			mockService := setupService(t)
			handler := NewNewsHandler(mockService, testLogger)

			app := fiber.New(fiber.Config{
				ErrorHandler: errors.ErrorHandler(testLogger),
			})
			app.Post("/edit/:id", handler.EditNews)

			req := httptest.NewRequest("POST", fmt.Sprintf("/edit/%s", rd.id), bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), "Invalid ID format")
			mockService.AssertNotCalled(t, "EditNews")
		})
	}

	invalidNewsData := []struct {
		name     string
		body     string
		errorMsg string
	}{
		{
			name:     "title only spaces",
			body:     `{"title":"     "}`,
			errorMsg: "title length must be between 1 and 255",
		},
		{
			name:     "content only spaces",
			body:     `{"content":"     "}`,
			errorMsg: "content length must be greater 1",
		},
		{
			name:     "title more than 255 chars",
			body:     fmt.Sprintf(`{"title":"%s"}`, strings.Repeat("a", 256)),
			errorMsg: "title length must be between 1 and 255",
		},
		//{
		//	name:     "both title and content only spaces",
		//	body:     `{"title":"   ","content":"   "}`,
		//	errorMsg: "title length must be between 1 and 255",
		//},
	}

	for _, in := range invalidNewsData {
		t.Run(fmt.Sprintf("Failed%s", in.name), func(t *testing.T) {
			mockService := setupService(t)
			handler := NewNewsHandler(mockService, testLogger)

			app := fiber.New(fiber.Config{
				ErrorHandler: errors.ErrorHandler(testLogger),
			})
			app.Post("/edit/:id", handler.EditNews)

			req := httptest.NewRequest("POST", fmt.Sprintf("/edit/%d", newsId), bytes.NewReader([]byte(in.body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), in.errorMsg)
			mockService.AssertNotCalled(t, "EditNews")
		})
	}

}
