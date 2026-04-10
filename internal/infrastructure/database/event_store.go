package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/go-rest-api/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventStore struct {
	pool *pgxpool.Pool
}

func NewEventStore(pool *pgxpool.Pool) *EventStore {
	return &EventStore{pool: pool}
}

func (s *EventStore) Append(ctx context.Context, aggregateID string, aggregateType string, events []user.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentVersion int
	err = tx.QueryRow(ctx, "SELECT COALESCE(MAX(event_version), 0) FROM events WHERE aggregate_id = $1", aggregateID).Scan(&currentVersion)
	if err != nil {
		return err
	}

	for index, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO events (id, aggregate_id, aggregate_type, event_type, event_version, payload, occurred_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, uuid.NewString(), aggregateID, aggregateType, event.EventType(), currentVersion+index+1, payload, event.OccurredAt())
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *EventStore) LoadUserEvents(ctx context.Context, aggregateID string) ([]user.Event, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT event_type, payload
		FROM events
		WHERE aggregate_id = $1
		ORDER BY event_version ASC
	`, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []user.Event
	for rows.Next() {
		var eventType string
		var payload []byte
		if err := rows.Scan(&eventType, &payload); err != nil {
			return nil, err
		}

		event, err := parseUserEvent(eventType, payload)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (s *EventStore) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	trimmedEmail := strings.TrimSpace(strings.ToLower(email))
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM events
			WHERE aggregate_type = 'user'
			  AND event_type = $1
			  AND LOWER(payload->>'email') = $2
		)
	`, user.EventTypeUserCreated, trimmedEmail).Scan(&exists)
	return exists, err
}

func parseUserEvent(eventType string, payload []byte) (user.Event, error) {
	switch eventType {
	case user.EventTypeUserCreated:
		var event user.UserCreated
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, err
		}
		return event, nil
	default:
		return nil, fmt.Errorf("unsupported user event type: %s", eventType)
	}
}
