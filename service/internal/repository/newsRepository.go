package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"service/internal/apperrors"
	"service/internal/models"

	"service/pkg/logger"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
)

var (
	//go:embed sql/select_news_by_limit_and_offset.sql
	SqlSelectNewsByLimitAndOffset string
	//go:embed sql/delete_news_categories.sql
	SqlDeleteNewsCategories string
	//go:embed sql/insert_news_categories.sql
	SqlInsertNewsCategories string
)

//go:generate mockery --name=INewsRepository --output=mocks --outpkg=mocks --case=snake --with-expecter
type INewsRepository interface {
	GetNews(limit, offset int64) ([]models.NewsWithCategories, error)
	CreateNews(createForm models.NewsCreateForm) (int64, error)
	UpdateNews(newsId int64, updateFields map[string]interface{}, categories *[]int64) error
}

type NewsRepository struct {
	db  *reform.DB
	log *logger.Logger
	ctx context.Context
}

func NewNewsRepository(db *reform.DB, log *logger.Logger, ctx context.Context) INewsRepository {
	return &NewsRepository{
		db:  db,
		log: log,
		ctx: ctx,
	}
}

func (r *NewsRepository) GetNews(limit, offset int64) ([]models.NewsWithCategories, error) {
	const op = "repository.news.GetNews"

	rows, err := r.db.QueryContext(r.ctx, SqlSelectNewsByLimitAndOffset, limit, offset)
	if err != nil {
		r.log.WithError(err).WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		}).Error("Failed to select news")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	newsList := make([]models.NewsWithCategories, 0)
	for rows.Next() {
		var n models.NewsWithCategories
		var categories []int64

		if err = rows.Scan(&n.ID, &n.Title, &n.Content, pq.Array(&categories)); err != nil {
			r.log.WithError(err).Error("Failed to scan news row")
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		if categories == nil {
			categories = []int64{}
		}
		n.Categories = categories

		newsList = append(newsList, n)
	}

	if err = rows.Err(); err != nil {
		r.log.WithError(err).Error("Error iterating news rows")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newsList, nil
}

func (r *NewsRepository) CreateNews(createForm models.NewsCreateForm) (int64, error) {
	const op = "repository.news.CreateNews"

	tx, err := r.db.Begin()
	if err != nil {
		r.log.WithError(err).Error("Failed to begin transaction")
		return 0, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer r.rollbackOnError(tx, op)

	news := &models.News{
		Title:   createForm.Title,
		Content: createForm.Content,
	}

	if err = tx.Save(news); err != nil {
		r.log.WithError(err).WithField("title", createForm.Title).Error("Failed to insert news")
		return 0, fmt.Errorf("%s: failed to insert news: %w", op, err)
	}

	newsID := news.ID

	if createForm.Categories != nil && len(*createForm.Categories) > 0 {
		if err = r.insertCategories(tx, newsID, *createForm.Categories); err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	if err = tx.Commit(); err != nil {
		r.log.WithError(err).Error("Failed to commit transaction")
		return 0, fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	r.log.WithField("news_id", newsID).Info("News created successfully")
	return newsID, nil
}

func (r *NewsRepository) UpdateNews(newsId int64, updateFields map[string]interface{}, categories *[]int64) error {
	const op = "repository.news.UpdateNews"

	tx, err := r.db.Begin()
	if err != nil {
		r.log.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer r.rollbackOnError(tx, op)

	news, err := r.findNewsByID(tx, newsId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(updateFields) > 0 {
		if title, ok := updateFields["title"]; ok {
			news.Title = title.(string)
		}

		if content, ok := updateFields["content"]; ok {
			news.Content = content.(string)
		}

		if err = tx.Update(news); err != nil {
			r.log.WithError(err).WithField("news_id", newsId).Error("Failed to update news")
			return fmt.Errorf("%s: failed to update: %w", op, err)
		}
	}

	if categories != nil {
		if err = r.updateCategories(tx, newsId, *categories); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err = tx.Commit(); err != nil {
		r.log.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("%s: failed to commit: %w", op, err)
	}

	r.log.WithField("news_id", newsId).Info("News updated successfully")
	return nil
}

func (r *NewsRepository) findNewsByID(tx *reform.TX, newsId int64) (*models.News, error) {
	record, err := tx.FindByPrimaryKeyFrom(models.NewsTable, newsId)
	if err != nil {
		if errors.Is(err, reform.ErrNoRows) {
			r.log.WithField("news_id", newsId).Warn("News not found")
			return nil, apperrors.NewNotFound("News not found")
		}
		r.log.WithError(err).WithField("news_id", newsId).Error("Failed to find news")
		return nil, fmt.Errorf("failed to find news: %w", err)
	}

	return record.(*models.News), nil
}

func (r *NewsRepository) rollbackOnError(tx *reform.TX, op string) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		r.log.WithError(err).WithField("operation", op).Error("Failed to rollback transaction")
	}
}

func (r *NewsRepository) insertCategories(tx *reform.TX, newsId int64, categoryIDs []int64) error {
	for _, categoryID := range categoryIDs {
		newsCategory := &models.NewsCategory{
			NewsId:     newsId,
			CategoryId: categoryID,
		}

		if err := tx.Save(newsCategory); err != nil {
			r.log.WithError(err).WithFields(logrus.Fields{
				"news_id":     newsId,
				"category_id": categoryID,
			}).Error("Failed to insert news category")
			return fmt.Errorf("failed to insert category %d: %w", categoryID, err)
		}
	}

	r.log.WithFields(logrus.Fields{
		"news_id":    newsId,
		"categories": len(categoryIDs),
	}).Debug("Categories inserted")

	return nil
}

func (r *NewsRepository) updateCategories(tx *reform.TX, newsId int64, categoryIDs []int64) error {
	if _, err := tx.ExecContext(r.ctx, SqlDeleteNewsCategories, newsId); err != nil {
		r.log.WithError(err).WithField("news_id", newsId).Error("Failed to delete old categories")
		return fmt.Errorf("failed to delete old categories: %w", err)
	}

	if len(categoryIDs) > 0 {
		for _, categoryID := range categoryIDs {
			if _, err := tx.ExecContext(r.ctx, SqlInsertNewsCategories, newsId, categoryID); err != nil {
				r.log.WithError(err).WithFields(logrus.Fields{
					"news_id":     newsId,
					"category_id": categoryID,
				}).Error("Failed to insert category")
				return fmt.Errorf("failed to insert category %d: %w", categoryID, err)
			}
		}

		r.log.WithFields(logrus.Fields{
			"news_id":    newsId,
			"categories": len(categoryIDs),
		}).Debug("Categories updated")
	}

	return nil
}
