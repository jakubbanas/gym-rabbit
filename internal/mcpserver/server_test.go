package mcpserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
)

// call runs a tool and decodes its structured output into out. It returns the
// error text for tool errors.
func call(t *testing.T, cs *mcp.ClientSession, tool string, args map[string]any, out any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		return res.Content[0].(*mcp.TextContent).Text
	}
	if out != nil {
		b, _ := json.Marshal(res.StructuredContent)
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatal(err)
		}
	}
	return ""
}

func TestListTrainingPlanMergesRunsAndStrength(t *testing.T) {
	fake := &fakeFinalSurge{}
	srv := fake.server()
	defer srv.Close()
	fs := finalsurge.New("a@b.c", "pw")
	fs.BaseURL = srv.URL + "/api/"
	cs, _ := connect(t, fs)

	if _, out := upload(t, cs, map[string]any{"sessions": hybridPlan}); out.Created != 4 {
		t.Fatalf("upload: %+v", out)
	}

	var plan planOutput
	if msg := call(t, cs, "list_training_plan", map[string]any{"start_date": "2026-10-20", "end_date": "2026-10-23"}, &plan); msg != "" {
		t.Fatal(msg)
	}
	var got []string
	for _, s := range plan.Sessions {
		got = append(got, s.Date+" "+s.Type+" "+s.Name)
	}
	want := "2026-10-20 run Easy 40|2026-10-21 strength Lower A|2026-10-22 run 12x400|2026-10-23 strength Upper B"
	if strings.Join(got, "|") != want {
		t.Errorf("sessions:\n got %s\nwant %s", strings.Join(got, "|"), want)
	}

	// Strength workouts outside the range are left out.
	call(t, cs, "list_training_plan", map[string]any{"start_date": "2026-10-22", "end_date": "2026-10-22"}, &plan)
	for _, s := range plan.Sessions {
		if s.Type == "strength" {
			t.Errorf("unexpected strength session in range: %+v", s)
		}
	}

	if msg := call(t, cs, "list_training_plan", map[string]any{"start_date": "2026-10-23", "end_date": "2026-10-20"}, nil); !strings.Contains(msg, "before") {
		t.Errorf("reversed range: %q", msg)
	}
}

func TestUpdateAndDeleteStrengthWorkout(t *testing.T) {
	cs, database := connect(t, nil)
	_, out := upload(t, cs, map[string]any{"sessions": hybridPlan[1:2]})
	id := out.Sessions[0].WorkoutID

	// Full replacement: the omitted note is cleared, the date moves.
	newBlocks := []map[string]any{{"exercises": []map[string]any{{"name": "Front Squat", "sets": 4, "reps": "4", "rest_seconds": 150}}}}
	var w struct {
		Title, Notes, ScheduledDate string
		Blocks                      []struct{ Exercises []struct{ Name string } }
	}
	msg := call(t, cs, "update_strength_workout", map[string]any{"id": id, "title": "Lower A v2", "scheduled_date": "2026-10-22", "blocks": newBlocks}, &w)
	if msg != "" {
		t.Fatal(msg)
	}
	saved, _ := database.GetWorkout(id)
	if saved.Title != "Lower A v2" || saved.ScheduledDate != "2026-10-22" || saved.Notes != "" || saved.Blocks[0].Exercises[0].Name != "Front Squat" {
		t.Errorf("after update: %+v", saved)
	}

	msg = call(t, cs, "update_strength_workout", map[string]any{"id": id, "title": " ", "scheduled_date": "22.10.2026",
		"blocks": []map[string]any{{"exercises": []map[string]any{{"name": "Squat", "sets": 0, "reps": "", "rest_seconds": 90}}}}}, nil)
	for _, want := range []string{"title", "scheduled_date", "blocks[0].exercises[0].sets", "blocks[0].exercises[0].reps"} {
		if !strings.Contains(msg, want) {
			t.Errorf("update error should mention %q:\n%s", want, msg)
		}
	}
	if msg := call(t, cs, "update_strength_workout", map[string]any{"id": "w-missing", "title": "X", "scheduled_date": "2026-10-22", "blocks": newBlocks}, nil); !strings.Contains(msg, "not found") {
		t.Errorf("update of missing workout: %q", msg)
	}

	if msg := call(t, cs, "delete_strength_workout", map[string]any{"id": id}, nil); msg != "" {
		t.Fatal(msg)
	}
	if ws, _ := database.ListWorkouts(); len(ws) != 0 {
		t.Errorf("workout not deleted")
	}
}

func TestToolSet(t *testing.T) {
	names := func(cs *mcp.ClientSession) string {
		res, err := cs.ListTools(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		var n []string
		for _, tool := range res.Tools {
			n = append(n, tool.Name)
		}
		return strings.Join(n, " ")
	}
	cs, _ := connect(t, nil)
	if got, want := names(cs), "delete_strength_workout get_strength_workout list_training_plan update_strength_workout upload_training_plan"; got != want {
		t.Errorf("tools without FinalSurge:\n got %s\nwant %s", got, want)
	}
	cs, _ = connect(t, finalsurge.New("a@b.c", "pw"))
	if got, want := names(cs), "delete_run delete_strength_workout get_run_structure get_strength_workout list_training_plan update_run_structure update_strength_workout upload_training_plan"; got != want {
		t.Errorf("tools with FinalSurge:\n got %s\nwant %s", got, want)
	}
}
