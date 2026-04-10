package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/example/go-rest-api/internal/application/command"
	"github.com/example/go-rest-api/internal/application/projection"
	"github.com/example/go-rest-api/internal/application/query"
	redisrepo "github.com/example/go-rest-api/internal/infrastructure/cache/redis"
	"github.com/example/go-rest-api/internal/infrastructure/config"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/example/go-rest-api/internal/infrastructure/database"
	kafkabus "github.com/example/go-rest-api/internal/infrastructure/messaging/kafka"
	ginhttp "github.com/example/go-rest-api/internal/interfaces/http/gin"
	redislib "github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		zlog.Fatal().Msg("DATABASE_URL is not set")
	}

	levelEnv := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch levelEnv {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zlog.Info().Msg("starting application")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := database.Connect()
	defer db.Close()

	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer redisClient.Close()

	readRepo := redisrepo.NewUserReadRepository(redisClient)
	store := database.NewEventStore(db)
	publisher := kafkabus.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()

	createUser := command.NewCreateUserHandler(store, publisher)
	getUser := query.NewGetUserHandler(readRepo)
	router := ginhttp.NewRouter(ginhttp.RouterDependencies{
		CreateUser: createUser,
		GetUser:    getUser,
	})

	if cfg.ProjectionEnabled {
		projector := projection.NewUserProjector(cfg.KafkaBrokers, cfg.KafkaTopic, readRepo, cfg.KafkaRetryDelay)
		go projector.Start(ctx)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			zlog.Error().Err(err).Msg("failed to shutdown http server")
		}
	}()

	zlog.Info().Str("port", cfg.Port).Msg("server running")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		zlog.Fatal().Err(err).Msg(fmt.Sprintf("server failed on :%s", cfg.Port))
	}
}
