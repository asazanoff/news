package internal

import (
	"context"
	"database/sql"
	"fmt"
	"service/internal/configs"
	"service/internal/handlers"
	"service/internal/handlers/errors"
	"service/internal/handlers/middleware"
	handler "service/internal/handlers/news"
	"service/internal/repository"
	"service/internal/service"
	"service/pkg/db"
	"service/pkg/logger"
	"time"

	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	log    *logger.Logger
	config configs.Config
	app    *fiber.App
	db     *sql.DB
}

func NewServer(ctx context.Context, log *logger.Logger) (*Server, error) {
	cnf, err := configs.NewParsedConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	database, reform, err := db.InitReformDB(cnf.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init reform db: %w", err)
	}

	repo := repository.NewNewsRepository(reform, log, ctx)
	newsService := service.NewNewsService(repo, log)
	newsHandler := handler.NewNewsHandler(newsService, log)
	app := fiber.New(fiber.Config{
		ErrorHandler: errors.ErrorHandler(log),
		ReadTimeout:  time.Duration(cnf.Service.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cnf.Service.WriteTimeout) * time.Second,
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.WithFields(logrus.Fields{
				"panic":  e,
				"method": c.Method(),
				"path":   c.Path(),
			}).Error("Panic recovered")
		},
	}))

	//в чем разница если устанавливать в сетап роутс мидлваре
	app.Use(middleware.HTTPLogger(log))

	handlers.SetupRoutes(app, newsHandler, middleware.AuthMiddleware(cnf.AuthToken, log))

	return &Server{
		config: cnf,
		app:    app,
		db:     database,
		log:    log,
	}, nil
}

func (s *Server) Start() error {
	s.log.Infof("Start server on port %s", s.config.Port)

	if err := s.app.Listen(":" + s.config.Port); err != nil {
		return fmt.Errorf("error start server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Start shutdown service")
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.app.ShutdownWithContext(ctx); err != nil {
			s.log.Errorf("Error shutdown server: %v", err)
			return fmt.Errorf("error shutdown server: %w", err)
		}
		s.log.Info("Server shutdown successfully")
		return nil
	})

	g.Go(func() error {
		if err := s.db.Close(); err != nil {
			s.log.Errorf("Error close database: %v", err)
			return fmt.Errorf("error close database: %w", err)
		}
		s.log.Info("database close successfully")
		return nil
	})

	return g.Wait()
}
