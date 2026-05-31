package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jakubbanas/gym-rabbit/internal/db"
)

// Handler holds the database reference and exposes HTTP handler methods.
type Handler struct {
	DB *db.DB
}

type createWorkoutRequest struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type updateWorkoutRequest struct {
	Title string `json:"title"`
	Text  string `json:"text"`
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

// extractID extracts the last path segment from a URL like /api/workouts/{id}
func extractID(r *http.Request) string {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// generateID creates a simple unique ID based on timestamp.
func generateID() string {
	return fmt.Sprintf("w-%d", time.Now().UnixNano())
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

// CreateWorkout handles POST /api/workouts
func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var req createWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title == "" || req.Text == "" {
		writeError(w, http.StatusBadRequest, "title and text are required")
		return
	}

	workout := &db.Workout{
		ID:    generateID(),
		Title: req.Title,
		Text:  req.Text,
	}

	if err := h.DB.CreateWorkout(workout); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workout")
		return
	}

	writeJSON(w, http.StatusCreated, workout)
}

// UpdateWorkout handles PUT /api/workouts/{id}
func (h *Handler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	id := extractID(r)
	if id == "" || id == "workouts" {
		writeError(w, http.StatusBadRequest, "workout ID is required")
		return
	}

	var req updateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Title == "" || req.Text == "" {
		writeError(w, http.StatusBadRequest, "title and text are required")
		return
	}

	workout := &db.Workout{
		ID:    id,
		Title: req.Title,
		Text:  req.Text,
	}

	if err := h.DB.UpdateWorkout(workout); err != nil {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}

	// Fetch the updated workout to return full data
	updated, err := h.DB.GetWorkout(id)
	if err != nil || updated == nil {
		writeJSON(w, http.StatusOK, workout)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeleteWorkout handles DELETE /api/workouts/{id}
func (h *Handler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	id := extractID(r)
	if id == "" || id == "workouts" {
		writeError(w, http.StatusBadRequest, "workout ID is required")
		return
	}

	if err := h.DB.DeleteWorkout(id); err != nil {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TouchLastPerformed handles POST /api/workouts/{id}/performed
func (h *Handler) TouchLastPerformed(w http.ResponseWriter, r *http.Request) {
	// URL is /api/workouts/{id}/performed — extract the id (second-to-last segment)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "workout ID is required")
		return
	}
	id := parts[len(parts)-2]

	if err := h.DB.TouchLastPerformed(id); err != nil {
		writeError(w, http.StatusNotFound, "workout not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
