package projection

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/example/go-rest-api/internal/domain/user"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type ReadModelRepository interface {
	Save(ctx context.Context, model user.ReadModel) error
}

type UserProjector struct {
	brokers    []string
	topic      string
	repo       ReadModelRepository
	retryDelay time.Duration
}

type kafkaMessage struct {
	AggregateID string          `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Payload     json.RawMessage `json:"payload"`
}

func NewUserProjector(brokers []string, topic string, repo ReadModelRepository, retryDelay time.Duration) *UserProjector {
	return &UserProjector{brokers: brokers, topic: topic, repo: repo, retryDelay: retryDelay}
}

func (p *UserProjector) Start(ctx context.Context) {
	for {
		if err := p.consume(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error().Err(err).Msg("projection consumer stopped; retrying")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(p.retryDelay):
		}
	}
}

func (p *UserProjector) consume(ctx context.Context) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     p.brokers,
		Topic:       p.topic,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	defer reader.Close()

	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		if err := p.project(ctx, message.Value); err != nil {
			log.Error().Err(err).Bytes("message", message.Value).Msg("failed to project message")
		}
	}
}

func (p *UserProjector) project(ctx context.Context, raw []byte) error {
	var envelope kafkaMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}

	switch envelope.EventType {
	case user.EventTypeUserCreated:
		var event user.UserCreated
		if err := json.Unmarshal(envelope.Payload, &event); err != nil {
			return err
		}
		return p.repo.Save(ctx, user.ReadModel{
			ID:        event.ID,
			Name:      event.Name,
			Email:     strings.ToLower(event.Email),
			CreatedAt: event.CreatedAt,
		})
	default:
		return nil
	}
}
