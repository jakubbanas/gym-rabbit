package api

import (
	"encoding/json"
	"net/http"

	"github.com/jakubbanas/gym-rabbit/internal/db"
)

// Handler holds the database reference and exposes HTTP handler methods.
type Handler struct {
	DB *db.DB
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// ListWorkouts handles GET /api/workouts
func (h *Handler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.DB.ListWorkouts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workouts")
		return
	}
	writeJSON(w, http.StatusOK, workouts)
}

// TouchLastPerformed handles POST /api/workouts/{id}/performed
func (h *Handler) TouchLastPerformed(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.DB.TouchLastPerformed(id); err != nil {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
