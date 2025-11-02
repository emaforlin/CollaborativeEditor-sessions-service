package repository

import "context"

type SessionRepository interface {
	AddSessionParticipant(ctx context.Context, docID, userID string) error
	RemoveSessionParticipant(ctx context.Context, docID, userID string) error
	GetSessionParticipants(ctx context.Context, docID string) ([]string, error)
}
