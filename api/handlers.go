package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/emaforlin/ce-sessions-service/repository"
)

const defaultRequestTimeout = 200 * time.Millisecond

type SessionsHandler struct {
	repository repository.SessionRepository
}

type SessionsResponse struct {
	Participants []string `json:"participants"`
}

func (h *SessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	documentID := r.PathValue("document_id")

	if documentID == "" {
		http.Error(w, "Invalid URL: missing document_id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultRequestTimeout)
	defer cancel()

	participants, err := h.repository.GetSessionParticipants(ctx, documentID)
	if err != nil {
		http.Error(w, "Failed to get session participants", http.StatusInternalServerError)
		return
	}

	response := SessionsResponse{
		Participants: participants,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

func NewSessionsHandler(repo repository.SessionRepository) *SessionsHandler {
	return &SessionsHandler{repository: repo}
}
