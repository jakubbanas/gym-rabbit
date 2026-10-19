package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

// Workout is a saved strength session.
type Workout struct {
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	Notes         string       `json:"notes,omitempty"`
	ScheduledDate string       `json:"scheduled_date"`
	Blocks        []plan.Block `json:"blocks"`
	LastPerformed string       `json:"last_performed,omitempty"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
}

// DB wraps the sql.DB connection.
type DB struct {
	conn *sql.DB
}

// New opens or creates the SQLite database at the given path.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

const workoutsSchema = `
	CREATE TABLE IF NOT EXISTS workouts (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		notes TEXT NOT NULL DEFAULT '',
		scheduled_date TEXT NOT NULL,
		blocks TEXT NOT NULL DEFAULT '[]',
		last_performed TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);`

func migrate(conn *sql.DB) error {
	if _, err := conn.Exec(workoutsSchema); err != nil {
		return err
	}
	// Workouts from before the structured format (Magic Paste text) are
	// discarded rather than converted: the old table is dropped and recreated.
	legacy, err := hasColumn(conn, "text")
	if err != nil || !legacy {
		return err
	}
	log.Printf("dropping workouts saved in the old Magic Paste text format")
	if _, err := conn.Exec(`DROP TABLE workouts`); err != nil {
		return err
	}
	_, err = conn.Exec(workoutsSchema)
	return err
}

func hasColumn(conn *sql.DB, name string) (bool, error) {
	var n int
	err := conn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('workouts') WHERE name = ?`, name).Scan(&n)
	return n > 0, err
}

var lastID atomic.Int64

// NewID creates a unique workout ID based on timestamp. IDs are strictly
// increasing, so calls in a tight loop (e.g. a bulk plan upload) never collide.
func NewID() string {
	for {
		last, n := lastID.Load(), time.Now().UnixNano()
		if n <= last {
			n = last + 1
		}
		if lastID.CompareAndSwap(last, n) {
			return fmt.Sprintf("w-%d", n)
		}
	}
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

const workoutColumns = `id, title, notes, scheduled_date, blocks, last_performed, created_at, updated_at`

type scanner interface{ Scan(dest ...any) error }

func scanWorkout(row scanner) (*Workout, error) {
	var w Workout
	var blocks string
	if err := row.Scan(&w.ID, &w.Title, &w.Notes, &w.ScheduledDate, &blocks, &w.LastPerformed, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(blocks), &w.Blocks); err != nil {
		return nil, fmt.Errorf("workout %s has corrupt blocks: %w", w.ID, err)
	}
	return &w, nil
}

// ListWorkouts returns all saved workouts ordered by scheduled date.
func (d *DB) ListWorkouts() ([]Workout, error) {
	return d.queryWorkouts(`SELECT ` + workoutColumns + ` FROM workouts ORDER BY scheduled_date, created_at`)
}

// ListWorkoutsBetween returns workouts scheduled between start and end
// (inclusive, YYYY-MM-DD), ordered by date.
func (d *DB) ListWorkoutsBetween(start, end string) ([]Workout, error) {
	return d.queryWorkouts(`SELECT `+workoutColumns+` FROM workouts WHERE scheduled_date BETWEEN ? AND ? ORDER BY scheduled_date, created_at`, start, end)
}

func (d *DB) queryWorkouts(query string, args ...any) ([]Workout, error) {
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workouts := []Workout{}
	for rows.Next() {
		w, err := scanWorkout(rows)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, *w)
	}
	return workouts, rows.Err()
}

// GetWorkout returns a single workout by ID, or nil if it doesn't exist.
func (d *DB) GetWorkout(id string) (*Workout, error) {
	w, err := scanWorkout(d.conn.QueryRow(`SELECT `+workoutColumns+` FROM workouts WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return w, err
}

// CreateWorkout inserts a new workout.
func (d *DB) CreateWorkout(w *Workout) error {
	blocks, err := json.Marshal(w.Blocks)
	if err != nil {
		return err
	}
	ts := now()
	w.CreatedAt = ts
	w.UpdatedAt = ts
	_, err = d.conn.Exec(
		`INSERT INTO workouts (`+workoutColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		w.ID, w.Title, w.Notes, w.ScheduledDate, string(blocks), w.LastPerformed, w.CreatedAt, w.UpdatedAt,
	)
	return err
}

// UpdateWorkout replaces a workout's title, notes, date and blocks. It leaves
// last_performed alone.
func (d *DB) UpdateWorkout(w *Workout) error {
	blocks, err := json.Marshal(w.Blocks)
	if err != nil {
		return err
	}
	w.UpdatedAt = now()
	result, err := d.conn.Exec(
		`UPDATE workouts SET title = ?, notes = ?, scheduled_date = ?, blocks = ?, updated_at = ? WHERE id = ?`,
		w.Title, w.Notes, w.ScheduledDate, string(blocks), w.UpdatedAt, w.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workout not found: %s", w.ID)
	}
	return nil
}

// DeleteWorkout removes a workout by ID.
func (d *DB) DeleteWorkout(id string) error {
	result, err := d.conn.Exec(`DELETE FROM workouts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workout not found: %s", id)
	}
	return nil
}

// TouchLastPerformed updates the last_performed timestamp for a workout.
func (d *DB) TouchLastPerformed(id string) error {
	ts := now()
	result, err := d.conn.Exec(`UPDATE workouts SET last_performed = ?, updated_at = ? WHERE id = ?`, ts, ts, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("workout not found: %s", id)
	}
	return nil
}
