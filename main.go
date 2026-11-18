package main

import (
	"crypto/subtle"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jakubbanas/gym-rabbit/internal/api"
	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
	"github.com/jakubbanas/gym-rabbit/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed static
var embeddedStatic embed.FS

func resolveStaticFS() fs.FS {
	if info, err := os.Stat("static"); err == nil && info.IsDir() {
		return os.DirFS("static")
	}
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		log.Fatalf("Failed to load embedded static assets: %v", err)
	}
	return sub
}

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

	// FinalSurge integration is enabled when credentials are present
	var finalSurge *finalsurge.Client
	if email, pw := os.Getenv("FINALSURGE_EMAIL"), os.Getenv("FINALSURGE_PASSWORD"); email != "" && pw != "" {
		finalSurge = finalsurge.New(email, pw)
		if u := os.Getenv("FINALSURGE_BASE_URL"); u != "" {
			finalSurge.BaseURL = u
		}
		log.Printf("FinalSurge integration enabled for MCP")
	}
	token := os.Getenv("MCP_TOKEN")
	if token == "" {
		log.Printf("MCP_TOKEN is not set: /mcp is disabled")
	}

	mux := newMux(database, finalSurge, token, resolveStaticFS())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🐰 Gym Rabbit server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, loggingMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// newMux wires up the HTTP routes.
func newMux(database *db.DB, finalSurge *finalsurge.Client, mcpToken string, staticFS fs.FS) *http.ServeMux {
	handler := &api.Handler{DB: database}
	mux := http.NewServeMux()

	// The REST API is public and read-only: plans are created, edited and
	// deleted only through the token-protected MCP endpoint. The one exception
	// is /performed, so the gym UI can record started sessions. Other methods
	// get 405 from the mux.
	mux.HandleFunc("GET /api/workouts", handler.ListWorkouts)
	mux.HandleFunc("POST /api/workouts/{id}/performed", handler.TouchLastPerformed)

	// MCP endpoint (Streamable HTTP). MCP can write plans, so without a token
	// it is disabled rather than open. Methods are listed explicitly so the
	// routes don't conflict with the SPA's "GET /".
	var mcpHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "MCP is disabled: set MCP_TOKEN on the server", http.StatusServiceUnavailable)
	})
	if mcpToken != "" {
		mcpServer := mcpserver.New(database, finalSurge)
		mcpHandler = bearerAuth(mcpToken, mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, nil))
	}
	for _, method := range []string{"GET", "POST", "DELETE"} {
		mux.Handle(method+" /mcp", mcpHandler)
	}

	// Serve the embedded static frontend, falling back to index.html for
	// client-side routes.
	fileServer := http.FileServerFS(staticFS)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if rel == "" {
			rel = "index.html"
		}
		if info, err := fs.Stat(staticFS, rel); err != nil || info.IsDir() {
			r.URL.Path = "/"
		}
		w.Header().Set("Cache-Control", "no-cache")
		fileServer.ServeHTTP(w, r)
	})
	return mux
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
