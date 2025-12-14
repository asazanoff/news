package service

import (
	"errors"
	"service/internal/apperrors"
	"service/internal/models"
	"service/internal/repository/mocks"
	customLog "service/pkg/logger"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testLogger = func() *customLog.Logger {
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	return &customLog.Logger{Logger: log}
}()

func setupRepo(t *testing.T) *mocks.INewsRepository {
	mockRepo := new(mocks.INewsRepository)

	t.Cleanup(func() {
		mockRepo.AssertExpectations(t)
	})

	return mockRepo
}

func TestCreateNews(t *testing.T) {
	createForm := models.NewsCreateForm{
		Title:      "Title",
		Content:    "Content",
		Categories: &[]int64{1, 2, 3},
	}
	var newsId int64 = 1

	t.Run("Success", func(t *testing.T) {
		mockRepo := setupRepo(t)

		mockRepo.On("CreateNews", createForm).Return(newsId, nil)
		service := NewNewsService(mockRepo, testLogger)

		id, err := service.CreateNews(createForm)

		assert.NoError(t, err)
		assert.Equal(t, newsId, id)
	})

	t.Run("Failed", func(t *testing.T) {
		mockRepo := setupRepo(t)
		expectedErr := apperrors.NewInternal("internal error")

		mockRepo.On("CreateNews", createForm).Return(int64(0), expectedErr)
		service := NewNewsService(mockRepo, testLogger)

		_, actualErr := service.CreateNews(createForm)

		assert.Error(t, actualErr)
		assert.EqualError(t, actualErr, expectedErr.Error())
	})
}

func TestListNews(t *testing.T) {
	var limit int64 = 10
	var offset int64 = 0
	newsList := []models.NewsWithCategories{
		{
			News: models.News{
				Title:   "News",
				Content: "All World",
			},
			Categories: []int64{1, 2},
		},
	}

	t.Run("ListNewsSuccess", func(t *testing.T) {
		mockRepo := setupRepo(t)

		mockRepo.On("GetNews", limit, offset).Return(newsList, nil)

		service := NewNewsService(mockRepo, testLogger)

		actualNewsList, actualErr := service.ListNews(limit, offset)

		assert.NoError(t, actualErr)
		assert.Equal(t, actualNewsList, newsList)
	})

	t.Run("ListNewsFailed", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo := setupRepo(t)

		mockRepo.On("GetNews", limit, offset).Return([]models.NewsWithCategories{}, expectedErr)

		service := NewNewsService(mockRepo, testLogger)

		_, actualErr := service.ListNews(limit, offset)

		assert.Error(t, actualErr)
		assert.EqualError(t, actualErr, expectedErr.Error())
	})
}

func TestEditNews(t *testing.T) {
	var newsId int64 = 10
	newTitle := "NewTitle"
	newContent := "NewContent"
	newCategories := []int64{1, 2}
	testDataSuccess := []struct {
		name                       string
		editForm                   models.NewsEditForm
		expectedModifiedFields     map[string]interface{}
		expectedModifiedCategories *[]int64
	}{
		{
			name: "update title only",
			editForm: models.NewsEditForm{
				Title: &newTitle,
			},
			expectedModifiedFields: map[string]interface{}{
				"title": newTitle,
			},
			expectedModifiedCategories: nil,
		},
		{
			name: "update content only",
			editForm: models.NewsEditForm{
				Content: &newContent,
			},
			expectedModifiedFields: map[string]interface{}{
				"content": newContent,
			},
			expectedModifiedCategories: nil,
		},
		{
			name: "update categories only",
			editForm: models.NewsEditForm{
				Categories: &newCategories,
			},
			expectedModifiedFields:     map[string]interface{}{},
			expectedModifiedCategories: &newCategories,
		},
		{
			name: "update all fields",
			editForm: models.NewsEditForm{
				Title:      &newTitle,
				Content:    &newContent,
				Categories: &newCategories,
			},
			expectedModifiedFields: map[string]interface{}{
				"title":   newTitle,
				"content": newContent,
			},
			expectedModifiedCategories: &newCategories,
		},
	}

	for _, tt := range testDataSuccess {
		t.Run("Success_"+tt.name, func(t *testing.T) {
			mockRepo := setupRepo(t)

			mockRepo.On("UpdateNews", newsId, tt.expectedModifiedFields, tt.expectedModifiedCategories).Return(nil)
			service := NewNewsService(mockRepo, testLogger)

			actualErr := service.EditNews(newsId, tt.editForm)

			assert.NoError(t, actualErr)
		})
	}

	t.Run("SuccessNoFieldsToUpdate", func(t *testing.T) {
		editForm := models.NewsEditForm{}
		mockRepo := setupRepo(t)

		service := NewNewsService(mockRepo, testLogger)

		actualErr := service.EditNews(newsId, editForm)

		assert.NoError(t, actualErr)
		mockRepo.AssertNotCalled(t, "UpdateNews")
	})

	t.Run("Failed", func(t *testing.T) {
		editForm := models.NewsEditForm{
			Title: &newTitle,
		}
		expectedErr := apperrors.NewNotFound("News not found")
		mockRepo := setupRepo(t)

		mockRepo.On("UpdateNews", newsId, map[string]interface{}{"title": newTitle}, (*[]int64)(nil)).Return(expectedErr)
		service := NewNewsService(mockRepo, testLogger)

		actualErr := service.EditNews(newsId, editForm)

		assert.Error(t, actualErr)
		assert.EqualError(t, actualErr, expectedErr.Error())
	})
}
