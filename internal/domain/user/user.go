package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const EventTypeUserCreated = "user.created"

var (
	ErrInvalidName  = errors.New("name is required")
	ErrInvalidEmail = errors.New("email is required")
)

type Event interface {
	EventType() string
	OccurredAt() time.Time
}

type UserCreated struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (e UserCreated) EventType() string {
	return EventTypeUserCreated
}

func (e UserCreated) OccurredAt() time.Time {
	return e.CreatedAt
}

type Aggregate struct {
	id            string
	name          string
	email         string
	createdAt     time.Time
	version       int
	pendingEvents []Event
}

func New(name, email string) (*Aggregate, error) {
	trimmedName := strings.TrimSpace(name)
	trimmedEmail := strings.TrimSpace(strings.ToLower(email))

	if trimmedName == "" {
		return nil, ErrInvalidName
	}
	if trimmedEmail == "" {
		return nil, ErrInvalidEmail
	}

	created := UserCreated{
		ID:        uuid.NewString(),
		Name:      trimmedName,
		Email:     trimmedEmail,
		CreatedAt: time.Now().UTC(),
	}

	aggregate := &Aggregate{}
	aggregate.apply(created)
	aggregate.pendingEvents = append(aggregate.pendingEvents, created)
	return aggregate, nil
}

func Rehydrate(history []Event) *Aggregate {
	aggregate := &Aggregate{}
	for _, event := range history {
		aggregate.apply(event)
	}
	aggregate.pendingEvents = nil
	return aggregate
}

func (a *Aggregate) apply(event Event) {
	switch e := event.(type) {
	case UserCreated:
		a.id = e.ID
		a.name = e.Name
		a.email = e.Email
		a.createdAt = e.CreatedAt
		a.version++
	}
}

func (a *Aggregate) ID() string {
	return a.id
}

func (a *Aggregate) PendingEvents() []Event {
	events := make([]Event, len(a.pendingEvents))
	copy(events, a.pendingEvents)
	return events
}

func (a *Aggregate) ClearPendingEvents() {
	a.pendingEvents = nil
}

type ReadModel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
