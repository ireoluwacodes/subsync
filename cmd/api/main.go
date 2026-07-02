package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ireoluwacodes/subsync/internal/auth"
	"github.com/ireoluwacodes/subsync/internal/config"
	"github.com/ireoluwacodes/subsync/internal/crypto"
	"github.com/ireoluwacodes/subsync/internal/db"
	"github.com/ireoluwacodes/subsync/internal/logger"
	"github.com/ireoluwacodes/subsync/internal/nomba"
	"github.com/ireoluwacodes/subsync/internal/queue"
	"github.com/ireoluwacodes/subsync/internal/router"
	"github.com/ireoluwacodes/subsync/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustInit(cfg.AppEnv, cfg.LogLevel)
	defer logger.Sync(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Connect(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Error("failed to connect to postgres", zap.Error(err))
		os.Exit(1)
	}
	defer database.Close()

	if err := db.Migrate(ctx, database); err != nil {
		log.Error("failed to run database migrations", zap.Error(err))
		os.Exit(1)
	}
	log.Info("database migrations applied")

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

	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	}

	key, err := crypto.ParseKey(cfg.DevEncryptionKey())
	if err != nil {
		log.Error("invalid NOMBA_CREDENTIALS_ENCRYPTION_KEY", zap.Error(err))
		os.Exit(1)
	}
	enc, err := crypto.NewCredentialEncryptor(key)
	if err != nil {
		log.Error("failed to init credential encryptor", zap.Error(err))
		os.Exit(1)
	}

	repos := db.NewRepos(database, enc)
	nombaClient := nomba.NewClient(log, nil)
	jwtSvc := auth.NewJWTService(cfg)
	svcs := service.NewServices(repos, cfg, nombaClient, jwtSvc, q)

	engine := router.Setup(cfg, database, q, repos, svcs)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("api server starting",
			zap.String("port", cfg.HTTPPort),
			zap.String("env", cfg.AppEnv),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down api server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
}
