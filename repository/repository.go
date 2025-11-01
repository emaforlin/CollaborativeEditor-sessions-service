package repository

type SessionRepository interface {
	AddSession(docID, userID string) error
	RemoveSession(docID, userID string) error
	GetActiveSessions(docId string) ([]string, error)
}
