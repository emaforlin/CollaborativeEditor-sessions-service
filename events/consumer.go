package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/emaforlin/ce-sessions-service/repository"
	"github.com/nats-io/nats.go"
)

type EventsConsumer struct {
	nats              *nats.Conn
	sub               *nats.Subscription
	sessionRepository repository.SessionRepository
	wg                sync.WaitGroup
}

const (
	SessionsEventsSubject = "session.>"
	UserJoin              = "user_joined"
	UserLeft              = "user_left"
)

func NewEventConsumer(natsConn *nats.Conn, sessionRepo repository.SessionRepository) *EventsConsumer {
	return &EventsConsumer{
		nats:              natsConn,
		sessionRepository: sessionRepo,
	}
}

func (ec *EventsConsumer) handleEvents(msg *nats.Msg) {
	ec.wg.Add(1)
	defer ec.wg.Done()

	subjectParts := strings.Split(msg.Subject, ".")
	eventType := subjectParts[len(subjectParts)-1]

	var eventPayload PresenceEventPayload
	if err := json.Unmarshal(msg.Data, &eventPayload); err != nil {
		log.Printf("invalid event payload: failed to unmarshal: %v", err)
		return
	}
	if !isPayloadValid(eventPayload) {
		log.Printf("invalid event payload: validation failed for payload: %+v", eventPayload)
		return
	}

	// Use a context with timeout for better resource management
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch eventType {
	case UserJoin:
		log.Printf("Handle user join - User: %s, Document: %s", eventPayload.UserID, eventPayload.DocumentID)
		if err := ec.sessionRepository.AddSessionParticipant(ctx, eventPayload.DocumentID, eventPayload.UserID); err != nil {
			log.Printf("event handling failed: %v", err)
		}
	case UserLeft:
		log.Printf("Handle user left - User: %s, Document: %s", eventPayload.UserID, eventPayload.DocumentID)
		if err := ec.sessionRepository.RemoveSessionParticipant(ctx, eventPayload.DocumentID, eventPayload.UserID); err != nil {
			log.Printf("event handling failed: %v", err)
		}
	default:
		log.Printf("Unknown event type: %s", eventType)
	}
}

func (ec *EventsConsumer) Start() error {
	sub, err := ec.nats.Subscribe(SessionsEventsSubject, ec.handleEvents)
	if err != nil {
		return fmt.Errorf("event subscription failed: %w", err)
	}
	ec.sub = sub
	return nil
}

func (ec *EventsConsumer) Stop() error {
	log.Println("Stopping event consumer...")

	if ec.sub != nil {
		if err := ec.sub.Drain(); err != nil {
			return fmt.Errorf("error draining subscription: %w", err)
		}
	}

	// Wait for all ongoing event processing to complete
	ec.wg.Wait()

	log.Println("Event consumer stopped successfully")
	return nil
}

func isPayloadValid(data PresenceEventPayload) bool {
	return data.DocumentID != "" && data.UserID != ""
}
