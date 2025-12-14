package service

import (
	"service/internal/models"
	"service/internal/repository"
	"service/pkg/logger"

	"github.com/sirupsen/logrus"
)

//go:generate mockery --name=INewsService --output=mocks --outpkg=mocks --case=snake --with-expecter
type INewsService interface {
	CreateNews(createForm models.NewsCreateForm) (int64, error)
	EditNews(newsId int64, editForm models.NewsEditForm) error
	ListNews(limit, offset int64) ([]models.NewsWithCategories, error)
}
type NewsService struct {
	repo repository.INewsRepository
	log  *logger.Logger
}

func NewNewsService(repo repository.INewsRepository, log *logger.Logger) INewsService {
	return &NewsService{
		repo: repo,
		log:  log,
	}
}

func (s *NewsService) CreateNews(createForm models.NewsCreateForm) (int64, error) {
	return s.repo.CreateNews(createForm)
}

func (s *NewsService) EditNews(newsId int64, editForm models.NewsEditForm) error {
	updateFields := make(map[string]interface{})
	if editForm.Title != nil {
		updateFields["title"] = *editForm.Title
	}
	if editForm.Content != nil {
		updateFields["content"] = *editForm.Content
	}

	if len(updateFields) > 0 || editForm.Categories != nil {
		if err := s.repo.UpdateNews(newsId, updateFields, editForm.Categories); err != nil {
			logrus.Error(err)
			return err
		}
	}
	//в другом случае ошибка no fields to updates?

	return nil
}

func (s *NewsService) ListNews(limit, offset int64) ([]models.NewsWithCategories, error) {
	newsList, err := s.repo.GetNews(limit, offset)
	if err != nil {
		return []models.NewsWithCategories{}, err
	}

	return newsList, nil
}
