package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/go-rest-api/internal/domain/user"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
	topic  string
}

type message struct {
	AggregateID string          `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Payload     json.RawMessage `json:"payload"`
}

func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
		topic: topic,
	}
}

func (p *Publisher) Publish(ctx context.Context, aggregateID string, events []user.Event) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(events))
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal kafka event payload: %w", err)
		}

		body, err := json.Marshal(message{
			AggregateID: aggregateID,
			EventType:   event.EventType(),
			OccurredAt:  event.OccurredAt(),
			Payload:     payload,
		})
		if err != nil {
			return fmt.Errorf("marshal kafka message: %w", err)
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(aggregateID),
			Value: body,
			Time:  event.OccurredAt(),
		})
	}

	return p.writer.WriteMessages(ctx, messages...)
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
