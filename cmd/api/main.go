package main

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/example/go-rest-api/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// initialize structured logger
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

	database.Connect()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	zlog.Info().Msg("Server running on :8080")
	r.Run(":8080")
}
