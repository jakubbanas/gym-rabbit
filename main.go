package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jakubbanas/gym-rabbit/internal/api"
	"github.com/jakubbanas/gym-rabbit/internal/db"
)

func main() {
	// Determine database path
	dbDir := "data"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	dbPath := filepath.Join(dbDir, "gym-rabbit.db")

	// Initialize database
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	handler := &api.Handler{DB: database}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/workouts/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is /api/workouts or /api/workouts/{id}
		trimmed := strings.TrimPrefix(r.URL.Path, "/api/workouts")
		trimmed = strings.TrimPrefix(trimmed, "/")
		trimmed = strings.TrimSuffix(trimmed, "/")

		if trimmed == "" {
			// /api/workouts
			switch r.Method {
			case http.MethodGet:
				handler.ListWorkouts(w, r)
			case http.MethodPost:
				handler.CreateWorkout(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Has an ID segment
		parts := strings.Split(strings.TrimPrefix(trimmed, "/"), "/")
		if len(parts) == 1 {
			// /api/workouts/{id}
			switch r.Method {
			case http.MethodPut:
				handler.UpdateWorkout(w, r)
			case http.MethodDelete:
				handler.DeleteWorkout(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 2 && parts[1] == "performed" {
			// /api/workouts/{id}/performed
			if r.Method == http.MethodPost {
				handler.TouchLastPerformed(w, r)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.NotFound(w, r)
	})

	// Serve React SPA from frontend/dist if it exists
	distDir := "frontend/dist"
	if _, err := os.Stat(distDir); err == nil {
		frontendFS := http.Dir(distDir)
		fileServer := http.FileServer(frontendFS)

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file directly
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}

			// Check if file exists in dist
			f, err := frontendFS.Open(path)
			if err != nil {
				if os.IsNotExist(err) || isPathError(err) {
					// SPA fallback: serve index.html for client-side routing
					r.URL.Path = "/"
					fileServer.ServeHTTP(w, r)
					return
				}
			}
			if f != nil {
				f.Close()
			}

			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Gym Rabbit API is running. Build the frontend with: cd frontend && npm run build")
		})
	}

	// Add CORS middleware for development
	corsHandler := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🐰 Gym Rabbit server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isPathError(err error) bool {
	_, ok := err.(*fs.PathError)
	return ok
}
