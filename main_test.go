package main

import (
	"io/fs"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

func TestRoutes(t *testing.T) {
	dir := t.TempDir()
	database, err := db.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	w := &db.Workout{ID: "w-1", Title: "Lower A", ScheduledDate: "2026-10-21", Blocks: []plan.Block{
		{Name: "Block A", Exercises: []plan.Exercise{{Name: "Squat", Sets: 3, Reps: "5"}}},
	}}
	if err := database.CreateWorkout(w); err != nil {
		t.Fatal(err)
	}

	staticFS := fstest.MapFS{
		"index.html": {Data: []byte("<title>Gym Rabbit</title>")},
		"app.js":     {Data: []byte("console.log(1)")},
	}

	// Ensure embedded assets are bundled cleanly
	embeddedSub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		t.Fatal(err)
	}
	for _, asset := range []string{"index.html", "style.css", "app.js", "logic.js"} {
		if _, err := fs.Stat(embeddedSub, asset); err != nil {
			t.Fatalf("embedded %s missing: %v", asset, err)
		}
	}

	for _, tc := range []struct {
		name, token, method, path, auth string
		status                          int
		body                            string
	}{
		{"list", "secret", "GET", "/api/workouts", "", 200, `"title":"Lower A"`},
		{"create is read-only", "secret", "POST", "/api/workouts", "", 405, ""},
		{"delete is read-only", "secret", "DELETE", "/api/workouts/w-1", "", 405, ""},
		{"performed", "secret", "POST", "/api/workouts/w-1/performed", "", 204, ""},
		{"performed unknown", "secret", "POST", "/api/workouts/w-nope/performed", "", 404, ""},
		{"spa index", "secret", "GET", "/", "", 200, "<title>Gym Rabbit"},
		{"spa fallback", "secret", "GET", "/some/route", "", 200, "<title>Gym Rabbit"},
		{"spa asset", "secret", "GET", "/app.js", "", 200, "console.log"},
		{"mcp without auth", "secret", "POST", "/mcp", "", 401, ""},
		{"mcp wrong token", "secret", "POST", "/mcp", "Bearer nope", 401, ""},
		{"mcp disabled without token", "", "POST", "/mcp", "Bearer anything", 503, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := newMux(database, nil, tc.token, staticFS)
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}"))
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Errorf("status %d, want %d (body %q)", rec.Code, tc.status, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.body) {
				t.Errorf("body %q does not contain %q", rec.Body.String(), tc.body)
			}
			if tc.method == "GET" && strings.HasPrefix(tc.name, "spa ") {
				if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
					t.Errorf("Cache-Control = %q, want no-cache", got)
				}
			}
		})
	}

	if got, _ := database.GetWorkout("w-1"); got.LastPerformed == "" {
		t.Error("performed did not set last_performed")
	}
}

func TestResolveStaticFS(t *testing.T) {
	// From a directory without static/, falls back to embedded assets.
	t.Chdir(t.TempDir())
	fsys := resolveStaticFS()
	if _, err := fs.Stat(fsys, "index.html"); err != nil {
		t.Fatalf("expected embedded index.html fallback, got error: %v", err)
	}
}
