package ginhttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/example/go-rest-api/internal/application/command"
	"github.com/example/go-rest-api/internal/application/query"
	"github.com/example/go-rest-api/internal/domain/user"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	CreateUser *command.CreateUserHandler
	GetUser    *query.GetUserHandler
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/users", func(ctx *gin.Context) {
		var request command.CreateUserCommand
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		aggregate, err := deps.CreateUser.Handle(ctx.Request.Context(), request)
		if err != nil {
			switch {
			case errors.Is(err, command.ErrEmailAlreadyExists):
				ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			case errors.Is(err, user.ErrInvalidName), errors.Is(err, user.ErrInvalidEmail):
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			}
			return
		}

		ctx.JSON(http.StatusAccepted, gin.H{"id": aggregate.ID(), "status": "accepted"})
	})

	router.GET("/users/:id", func(ctx *gin.Context) {
		model, err := deps.GetUser.ByID(ctx.Request.Context(), ctx.Param("id"))
		if err != nil {
			handleQueryError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, model)
	})

	router.GET("/users", func(ctx *gin.Context) {
		email := strings.TrimSpace(ctx.Query("email"))
		if email == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter is required"})
			return
		}

		model, err := deps.GetUser.ByEmail(ctx.Request.Context(), email)
		if err != nil {
			handleQueryError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, model)
	})

	return router
}

func handleQueryError(ctx *gin.Context, err error) {
	if errors.Is(err, query.ErrUserNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
}
