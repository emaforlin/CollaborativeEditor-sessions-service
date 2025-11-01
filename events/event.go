package events

type PresenceEventPayload struct {
	DocumentID string `json:"document_id"`
	UserID     string `json:"user_id"`
}
