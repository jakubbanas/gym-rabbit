package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Workout represents a saved workout plan.
type Workout struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Text          string `json:"text"`
	LastPerformed string `json:"lastPerformed"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
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

func migrate(conn *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS workouts (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		text TEXT NOT NULL,
		last_performed TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);`
	_, err := conn.Exec(schema)
	return err
}

// NewID creates a simple unique workout ID based on timestamp.
func NewID() string {
	return fmt.Sprintf("w-%d", time.Now().UnixNano())
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ListWorkouts returns all saved workouts ordered by most recently updated.
func (d *DB) ListWorkouts() ([]Workout, error) {
	rows, err := d.conn.Query(`SELECT id, title, text, last_performed, created_at, updated_at FROM workouts ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		if err := rows.Scan(&w.ID, &w.Title, &w.Text, &w.LastPerformed, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workouts = append(workouts, w)
	}

	if workouts == nil {
		workouts = []Workout{}
	}
	return workouts, rows.Err()
}

// GetWorkout returns a single workout by ID.
func (d *DB) GetWorkout(id string) (*Workout, error) {
	var w Workout
	err := d.conn.QueryRow(`SELECT id, title, text, last_performed, created_at, updated_at FROM workouts WHERE id = ?`, id).
		Scan(&w.ID, &w.Title, &w.Text, &w.LastPerformed, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// CreateWorkout inserts a new workout.
func (d *DB) CreateWorkout(w *Workout) error {
	ts := now()
	w.CreatedAt = ts
	w.UpdatedAt = ts
	_, err := d.conn.Exec(
		`INSERT INTO workouts (id, title, text, last_performed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		w.ID, w.Title, w.Text, w.LastPerformed, w.CreatedAt, w.UpdatedAt,
	)
	return err
}

// UpdateWorkout updates an existing workout's title and text.
func (d *DB) UpdateWorkout(w *Workout) error {
	w.UpdatedAt = now()
	result, err := d.conn.Exec(
		`UPDATE workouts SET title = ?, text = ?, updated_at = ? WHERE id = ?`,
		w.Title, w.Text, w.UpdatedAt, w.ID,
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
