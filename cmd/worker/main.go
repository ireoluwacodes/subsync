package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/jobs"
	"github.com/ireoluwacodes/subsync/internal/logger"
	"github.com/ireoluwacodes/subsync/internal/observability"
	"github.com/ireoluwacodes/subsync/internal/queue"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustInit(cfg.AppEnv, cfg.LogLevel)
	defer logger.Sync(log)

	flushSentry, err := observability.InitSentry(cfg)
	if err != nil {
		log.Error("failed to initialize sentry", zap.Error(err))
		os.Exit(1)
	}
	defer flushSentry()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Error("failed to connect to postgres", zap.Error(err))
		os.Exit(1)
	}
	defer database.Close()

	q, err := queue.Connect(cfg.RedisURL)
	if err != nil {
		log.Error("failed to connect to redis", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.Error("failed to close redis", zap.Error(err))
		}
	}()

	handlers, err := jobs.NewHandlers(ctx, cfg, log, database, q)
	if err != nil {
		log.Error("failed to wire worker handlers", zap.Error(err))
		os.Exit(1)
	}

	server, err := q.NewServer(cfg.RedisURL, 10)
	if err != nil {
		log.Error("failed to create asynq server", zap.Error(err))
		os.Exit(1)
	}

	registry := jobs.NewRegistry(handlers)
	registry.RegisterAll()

	scheduler, err := q.NewScheduler(cfg.RedisURL)
	if err != nil {
		log.Error("failed to create asynq scheduler", zap.Error(err))
		os.Exit(1)
	}

	if err := jobs.RegisterPeriodicTasks(scheduler); err != nil {
		log.Error("failed to register periodic tasks", zap.Error(err))
		os.Exit(1)
	}

	go func() {
		log.Info("worker server starting")
		if err := server.Run(registry.Mux()); err != nil {
			log.Error("asynq server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	go func() {
		log.Info("scheduler starting")
		if err := scheduler.Run(); err != nil {
			log.Error("scheduler error", zap.Error(err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down worker")

	server.Shutdown()
	scheduler.Shutdown()
}
