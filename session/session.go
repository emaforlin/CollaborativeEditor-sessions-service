package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type SessionManager struct {
	natsConn    *nats.Conn
	redisClient *redis.Client
	ctx         context.Context
	wg          sync.WaitGroup
}

type UserConnectionEvent struct {
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
}

func NewSessionManager(natsConn *nats.Conn, redisClient *redis.Client) *SessionManager {
	return &SessionManager{
		natsConn:    natsConn,
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

func (sm *SessionManager) Start() error {
	_, err := sm.natsConn.Subscribe("document.*.user", sm.handleConnectionEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to document users events: %w", err)
	}

	log.Println("Session manager started, listening for connection events...")
	return nil
}

func (sm *SessionManager) handleConnectionEvent(msg *nats.Msg) {
	sm.wg.Add(1)
	var event UserConnectionEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal connection event: %v", err)
		return
	}

	documentID := strings.Split(msg.Subject, ".")[1]

	log.Printf("Processing connection event: %+v", event)

	switch event.Action {
	case "join":
		if err := sm.addUserToDocument(documentID, event.UserID); err != nil {
			log.Printf("Failed to add user %s to document %s: %v", event.UserID, documentID, err)
		}
	case "left":
		if err := sm.removeUserFromDocument(documentID, event.UserID); err != nil {
			log.Printf("Failed to remove user %s from document %s: %v", event.UserID, documentID, err)
		}
	default:
		log.Printf("Unknown action: %s", event.Action)
	}
}

// addUserToDocument adds a user to a document's active session
func (sm *SessionManager) addUserToDocument(documentID, userID string) error {
	key := fmt.Sprintf("session:%s", documentID)

	// Add user to the set
	err := sm.redisClient.SAdd(sm.ctx, key, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to add user to session: %w", err)
	}

	// Set expiration for the key (24 hours)
	err = sm.redisClient.Expire(sm.ctx, key, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for session key: %w", err)
	}

	log.Printf("User %s added to document %s session", userID, documentID)
	return nil
}

// removeUserFromDocument removes a user from a document's active session
func (sm *SessionManager) removeUserFromDocument(documentID, userID string) error {
	key := fmt.Sprintf("session:%s", documentID)

	// Remove user from the set
	err := sm.redisClient.SRem(sm.ctx, key, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove user from session: %w", err)
	}

	// Check if the set is empty and remove the key if so
	count, err := sm.redisClient.SCard(sm.ctx, key).Result()
	if err != nil {
		log.Printf("Failed to check session count for document %s: %v", documentID, err)
	} else if count == 0 {
		err = sm.redisClient.Del(sm.ctx, key).Err()
		if err != nil {
			log.Printf("Failed to delete empty session key for document %s: %v", documentID, err)
		}
	}

	log.Printf("User %s removed from document %s session", userID, documentID)
	return nil
}

// GetActiveUsers returns the list of active users for a document
func (sm *SessionManager) GetActiveUsers(documentID string) ([]string, error) {
	key := fmt.Sprintf("session:%s", documentID)

	users, err := sm.redisClient.SMembers(sm.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active users for document %s: %w", documentID, err)
	}

	return users, nil
}

// GetActiveDocuments returns all documents that have active sessions
func (sm *SessionManager) GetActiveDocuments() ([]string, error) {
	keys, err := sm.redisClient.Keys(sm.ctx, "session:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active document keys: %w", err)
	}

	var documentIDs []string
	for _, key := range keys {
		// Extract document ID from key (remove "session:" prefix)
		if len(key) > 8 {
			documentIDs = append(documentIDs, key[8:])
		}
	}

	return documentIDs, nil
}

// GetSessionCount returns the number of active users in a document session
func (sm *SessionManager) GetSessionCount(documentID string) (int64, error) {
	key := fmt.Sprintf("session:%s", documentID)

	count, err := sm.redisClient.SCard(sm.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get session count for document %s: %w", documentID, err)
	}

	return count, nil
}

// Stop gracefully shuts down the session manager
func (sm *SessionManager) Stop() error {
	if sm.natsConn != nil {
		sm.natsConn.Close()
	}

	if sm.redisClient != nil {
		return sm.redisClient.Close()
	}

	log.Println("Session manager stopped")
	return nil
}
