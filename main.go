package main

import (
	"crypto/subtle"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jakubbanas/gym-rabbit/internal/api"
	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
	"github.com/jakubbanas/gym-rabbit/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	// API routes. The REST API is public and read-only: plans are created,
	// edited and deleted only through the token-protected MCP endpoint. The one
	// exception is /performed, so the gym UI can still record finished sessions.
	workoutsRoute := func(w http.ResponseWriter, r *http.Request) {
		trimmed := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/workouts"), "/")
		parts := strings.Split(trimmed, "/")

		switch {
		case trimmed == "":
			// /api/workouts
			if r.Method != http.MethodGet {
				readOnly(w)
				return
			}
			handler.ListWorkouts(w, r)
		case len(parts) == 1:
			// /api/workouts/{id}
			readOnly(w)
		case len(parts) == 2 && parts[1] == "performed":
			// /api/workouts/{id}/performed
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handler.TouchLastPerformed(w, r)
		default:
			http.NotFound(w, r)
		}
	}
	// Register both forms so /api/workouts is served directly instead of 307-redirected
	mux.HandleFunc("/api/workouts", workoutsRoute)
	mux.HandleFunc("/api/workouts/", workoutsRoute)

	// MCP endpoint (Streamable HTTP) for AI assistants
	// FinalSurge integration is enabled when credentials are present
	var fs *finalsurge.Client
	if email, pw := os.Getenv("FINALSURGE_EMAIL"), os.Getenv("FINALSURGE_PASSWORD"); email != "" && pw != "" {
		fs = finalsurge.New(email, pw)
		if u := os.Getenv("FINALSURGE_BASE_URL"); u != "" {
			fs.BaseURL = u
		}
		log.Printf("FinalSurge integration enabled for MCP")
	}
	mcpServer := mcpserver.New(database, fs)
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, nil)
	if token := os.Getenv("MCP_TOKEN"); token != "" {
		mux.Handle("/mcp", bearerAuth(token, mcpHandler))
	} else {
		// MCP can write plans, so it is never served without a token.
		log.Printf("MCP_TOKEN is not set: /mcp is disabled")
		mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "MCP is disabled: set MCP_TOKEN on the server", http.StatusServiceUnavailable)
		})
	}

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

	// Add middleware chain: Logging -> CORS
	handlerWithMiddleware := loggingMiddleware(corsMiddleware(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🐰 Gym Rabbit server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, handlerWithMiddleware); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the underlying writer (needed for MCP streaming flushes).
func (w *statusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		} else if idx := strings.Index(clientIP, ","); idx != -1 {
			clientIP = strings.TrimSpace(clientIP[:idx])
		}

		log.Printf("%s %s %d %s %s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start), clientIP)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// readOnly rejects REST writes, which are only allowed through MCP.
func readOnly(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, `{"error":"the REST API is read-only; create, edit and delete plans through MCP"}`)
}

// bearerAuth requires "Authorization: Bearer <token>".
func bearerAuth(token string, next http.Handler) http.Handler {
	expected := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), expected) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isPathError(err error) bool {
	_, ok := err.(*fs.PathError)
	return ok
}
