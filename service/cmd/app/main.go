package main

import (
	"context"
	"time"

	//_ "orderService/docs"
	"os"
	"os/signal"
	"service/internal"
	"service/pkg/logger"
	"syscall"
)

// @title Order Service
// @version 1.0
// @description Order service API Docs

// @host 	localhost:8080
// @BasePath /api
func main() {
	log := logger.Init()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	server, err := internal.NewServer(context.Background(), log)
	if err != nil {
		log.Fatal(err)
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("Service starting...")
		serverErrors <- server.Start()
	}()

	select {
	case err = <-serverErrors:
		log.Fatal(err)
	case sig := <-shutdown:
		log.Infof("Getting shutdown signal : %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err = server.Stop(ctx); err != nil {
			log.Errorf("Error with graceful shutdown: %v", err)
			log.Fatal("Принудительное завершение работы")
		}

		log.Info("The server has been stopped successfully")
	}
}
