package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
)

// fakeFinalSurge records created workouts and saved structures.
type fakeFinalSurge struct {
	mu         sync.Mutex
	created    []map[string]any
	structures []string
}

func (f *fakeFinalSurge) server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		reply := func(m map[string]any) { m["success"] = true; _ = json.NewEncoder(w).Encode(m) }
		switch r.URL.Path {
		case "/api/login":
			reply(map[string]any{"data": map[string]any{"token": "tok"}})
		case "/api/Settings":
			reply(map[string]any{"data": map[string]any{"user_key": "u1"}})
		case "/api/WorkoutSave":
			var b map[string]any
			_ = json.NewDecoder(r.Body).Decode(&b)
			f.created = append(f.created, b)
			reply(map[string]any{"new_workout_key": "fs-" + b["name"].(string)})
		case "/api/WorkoutList":
			list := []map[string]any{}
			for _, c := range f.created {
				list = append(list, map[string]any{
					"key": "fs-" + c["name"].(string), "workout_date": c["workout_date"], "name": c["name"],
					"workout_status_text": "Planned", "Activities": []map[string]any{{"activity_type_name": "Run"}},
				})
			}
			reply(map[string]any{"data": list})
		case "/api/WorkoutBuilderSave":
			f.structures = append(f.structures, r.URL.Query().Get("workout_key"))
			reply(map[string]any{})
		default:
			http.NotFound(w, r)
		}
	}))
}

func connect(t *testing.T, fs *finalsurge.Client) (*mcp.ClientSession, *db.DB) {
	t.Helper()
	database, err := db.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := New(database, fs).Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, database
}

func upload(t *testing.T, cs *mcp.ClientSession, args map[string]any) (*mcp.CallToolResult, uploadPlanOut) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "upload_training_plan", Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	var out uploadPlanOut
	if !res.IsError {
		b, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(b, &out)
	}
	return res, out
}

var hybridPlan = []map[string]any{
	{"type": "run", "date": "2026-10-20", "name": "Easy 40", "planned_minutes": 40},
	{"type": "strength", "date": "2026-10-21", "name": "Lower A", "blocks": []map[string]any{
		{"exercises": []map[string]any{{"name": "Squat", "sets": 3, "reps": "5", "weight_kg": 100, "rest_seconds": 120}}},
	}},
	{"type": "run", "date": "2026-10-22", "name": "12x400", "structure": []map[string]any{
		{"intensity": "warmup", "time": "10:00"},
		{"repeats": 12, "steps": []map[string]any{{"distance_km": 0.4, "pace_fast": "4:30", "pace_slow": "4:40"}, {"intensity": "rest", "time": "1:30"}}},
		{"intensity": "cooldown", "time": "10:00"},
	}},
	{"type": "strength", "date": "2026-10-23", "name": "Upper B", "notes": "Deload", "blocks": []map[string]any{
		{"name": "Superset", "exercises": []map[string]any{
			{"name": "Bench", "sets": 3, "reps": "8", "weight_kg": 70, "rest_seconds": 0},
			{"name": "Row", "sets": 3, "reps": "8", "per_side": true, "weight_kg": 24, "load_note": "per dumbbell", "rest_seconds": 60},
		}},
	}},
}

func TestUploadTrainingPlan(t *testing.T) {
	fake := &fakeFinalSurge{}
	srv := fake.server()
	defer srv.Close()
	fs := finalsurge.New("a@b.c", "pw")
	fs.BaseURL = srv.URL + "/api/"
	cs, database := connect(t, fs)

	res, out := upload(t, cs, map[string]any{"sessions": hybridPlan})
	if res.IsError {
		t.Fatalf("upload failed: %v", res.Content[0].(*mcp.TextContent).Text)
	}
	if out.Created != 4 || out.Failed != 0 {
		t.Fatalf("created %d, failed %d: %+v", out.Created, out.Failed, out.Sessions)
	}
	if len(fake.created) != 2 || len(fake.structures) != 1 || fake.structures[0] != "fs-12x400" {
		t.Errorf("finalsurge calls: created %d, structures %v", len(fake.created), fake.structures)
	}
	if run := out.Sessions[2].Run; run == nil || !run.StructureSaved || !run.PlannedSecondsEstimated {
		t.Errorf("structured run result: %+v", run)
	}

	ws, _ := database.ListWorkouts()
	byDate := map[string]db.Workout{}
	for _, w := range ws {
		byDate[w.ScheduledDate] = w
	}
	lower, upper := byDate["2026-10-21"], byDate["2026-10-23"]
	if len(ws) != 2 || lower.Title != "Lower A" || upper.Title != "Upper B" || upper.Notes != "Deload" {
		t.Fatalf("strength workouts: %+v", ws)
	}
	if lower.Blocks[0].Name != "Block A" {
		t.Errorf("default block name: %q", lower.Blocks[0].Name)
	}
	if row := upper.Blocks[0].Exercises[1]; !row.PerSide || row.WeightKg != 24 || row.LoadNote != "per dumbbell" {
		t.Errorf("row: %+v", row)
	}
}

func TestUploadTrainingPlanRejectsInvalidWithoutWriting(t *testing.T) {
	fake := &fakeFinalSurge{}
	srv := fake.server()
	defer srv.Close()
	fs := finalsurge.New("a@b.c", "pw")
	fs.BaseURL = srv.URL + "/api/"
	cs, database := connect(t, fs)

	bad := append([]map[string]any{}, hybridPlan...)
	bad = append(bad,
		map[string]any{"type": "strength", "date": "2026-10-24", "name": "Bad", "blocks": []map[string]any{
			{"exercises": []map[string]any{{"name": "Squat", "sets": 0, "reps": "5", "rest_seconds": 90}}},
		}},
		map[string]any{"type": "run", "date": "24.10.2026", "name": "Long"},
		map[string]any{"type": "swim", "date": "2026-10-25", "name": "Swim"},
	)
	res, _ := upload(t, cs, map[string]any{"sessions": bad})
	if !res.IsError {
		t.Fatal("expected an error")
	}
	msg := res.Content[0].(*mcp.TextContent).Text
	for _, want := range []string{"session 4", "blocks[0].exercises[0].sets", "session 5", "session 6", "nothing was written"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error should mention %q:\n%s", want, msg)
		}
	}
	ws, _ := database.ListWorkouts()
	if len(fake.created) != 0 || len(ws) != 0 {
		t.Errorf("wrote despite invalid plan: %d runs, %d strength", len(fake.created), len(ws))
	}
}

func TestUploadTrainingPlanDryRunAndNoFinalSurge(t *testing.T) {
	cs, database := connect(t, nil)

	res, _ := upload(t, cs, map[string]any{"sessions": hybridPlan, "dry_run": true})
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "FinalSurge is not configured") {
		t.Errorf("runs should be rejected without FinalSurge: %+v", res.Content)
	}

	res, out := upload(t, cs, map[string]any{"sessions": hybridPlan[1:2], "dry_run": true})
	if res.IsError || out.Created != 0 || len(out.Sessions) != 1 || out.Sessions[0].Status != "valid" {
		t.Errorf("dry run: %+v", out)
	}
	if ws, _ := database.ListWorkouts(); len(ws) != 0 {
		t.Errorf("dry run wrote %d workouts", len(ws))
	}

}
