package query

import (
	"context"
	"errors"

	"github.com/example/go-rest-api/internal/domain/user"
)

var ErrUserNotFound = errors.New("user not found")

type ReadRepository interface {
	GetByID(ctx context.Context, id string) (*user.ReadModel, error)
	GetByEmail(ctx context.Context, email string) (*user.ReadModel, error)
}

type GetUserHandler struct {
	repo ReadRepository
}

func NewGetUserHandler(repo ReadRepository) *GetUserHandler {
	return &GetUserHandler{repo: repo}
}

func (h *GetUserHandler) ByID(ctx context.Context, id string) (*user.ReadModel, error) {
	model, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, ErrUserNotFound
	}
	return model, nil
}

func (h *GetUserHandler) ByEmail(ctx context.Context, email string) (*user.ReadModel, error) {
	model, err := h.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, ErrUserNotFound
	}
	return model, nil
}
