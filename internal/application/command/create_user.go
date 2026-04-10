package command

import (
	"context"
	"errors"

	"github.com/example/go-rest-api/internal/domain/user"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type EventStore interface {
	Append(ctx context.Context, aggregateID string, aggregateType string, events []user.Event) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, aggregateID string, events []user.Event) error
}

type CreateUserHandler struct {
	store     EventStore
	publisher EventPublisher
}

func NewCreateUserHandler(store EventStore, publisher EventPublisher) *CreateUserHandler {
	return &CreateUserHandler{store: store, publisher: publisher}
}

type CreateUserCommand struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *CreateUserHandler) Handle(ctx context.Context, command CreateUserCommand) (*user.Aggregate, error) {
	exists, err := h.store.ExistsByEmail(ctx, command.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	aggregate, err := user.New(command.Name, command.Email)
	if err != nil {
		return nil, err
	}

	events := aggregate.PendingEvents()
	if err := h.store.Append(ctx, aggregate.ID(), "user", events); err != nil {
		return nil, err
	}
	if err := h.publisher.Publish(ctx, aggregate.ID(), events); err != nil {
		return nil, err
	}

	aggregate.ClearPendingEvents()
	return aggregate, nil
}
