package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/example/go-rest-api/internal/domain/user"
	redislib "github.com/redis/go-redis/v9"
)

type UserReadRepository struct {
	client *redislib.Client
}

func NewUserReadRepository(client *redislib.Client) *UserReadRepository {
	return &UserReadRepository{client: client}
}

func (r *UserReadRepository) Save(ctx context.Context, model user.ReadModel) error {
	key := userKey(model.ID)
	values := map[string]any{
		"id":         model.ID,
		"name":       model.Name,
		"email":      model.Email,
		"created_at": model.CreatedAt.UTC().Format(time.RFC3339Nano),
	}

	pipe := r.client.TxPipeline()
	pipe.HSet(ctx, key, values)
	pipe.HSet(ctx, "users:email_index", strings.ToLower(model.Email), model.ID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *UserReadRepository) GetByID(ctx context.Context, id string) (*user.ReadModel, error) {
	result, err := r.client.HGetAll(ctx, userKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return parseReadModel(result)
}

func (r *UserReadRepository) GetByEmail(ctx context.Context, email string) (*user.ReadModel, error) {
	id, err := r.client.HGet(ctx, "users:email_index", strings.ToLower(strings.TrimSpace(email))).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *UserReadRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.client.HExists(ctx, "users:email_index", strings.ToLower(strings.TrimSpace(email))).Result()
	return count, err
}

func parseReadModel(values map[string]string) (*user.ReadModel, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, values["created_at"])
	if err != nil {
		return nil, err
	}
	return &user.ReadModel{
		ID:        values["id"],
		Name:      values["name"],
		Email:     values["email"],
		CreatedAt: createdAt,
	}, nil
}

func userKey(id string) string {
	return "user:" + id
}
