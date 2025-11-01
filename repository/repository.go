package repository

import "context"

type SessionRepository interface {
	AddSession(ctx context.Context, docID, userID string) error
	RemoveSession(ctx context.Context, docID, userID string) error
	GetActiveSessions(ctx context.Context, docId string) ([]string, error)
}
