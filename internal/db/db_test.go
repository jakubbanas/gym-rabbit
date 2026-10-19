package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

func TestOldTextFormatIsDropped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`CREATE TABLE workouts (id TEXT PRIMARY KEY, title TEXT NOT NULL, text TEXT NOT NULL,
		last_performed TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
		INSERT INTO workouts (id, title, text, created_at, updated_at) VALUES ('w-1', 'Old', 'Block A', 'c', 'u')`); err != nil {
		t.Fatal(err)
	}
	conn.Close()

	// Opening twice checks the drop happens once and the new table then persists.
	for i := 0; i < 2; i++ {
		d, err := New(path)
		if err != nil {
			t.Fatalf("open %d: %v", i, err)
		}
		ws, err := d.ListWorkouts()
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			if len(ws) != 0 {
				t.Errorf("old workouts should be gone, got %+v", ws)
			}
			w := &Workout{ID: NewID(), Title: "New", ScheduledDate: "2026-10-21", Blocks: []plan.Block{{Name: "Block A", Exercises: []plan.Exercise{{Name: "Row", Sets: 3, Reps: "8"}}}}}
			if err := d.CreateWorkout(w); err != nil {
				t.Fatalf("create: %v", err)
			}
		} else if len(ws) != 1 || ws[0].Blocks[0].Exercises[0].Name != "Row" {
			t.Errorf("new workout should survive reopening: %+v", ws)
		}
		d.Close()
	}
}
