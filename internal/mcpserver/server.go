// Package mcpserver exposes the training plan over the Model Context Protocol:
// strength workouts stored in Gym Rabbit and, when configured, runs in FinalSurge.
package mcpserver

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

// strengthHelp is appended to tool descriptions that take strength blocks.
const strengthHelp = `A strength workout is a list of blocks. A block is a group of exercises done together
(superset or circuit); for straight sets use one exercise per block. The app labels exercises A1, A2, B1, …
Exercise fields: name, sets, reps (string: "8", "8-10", "AMRAP", "45s"), rest_seconds (0 = go straight to the next
exercise in the block), optional per_side (reps are per side), weight_kg (omit for bodyweight), load_note, cues.
Example blocks:
[{"exercises":[{"name":"Back Squat","sets":4,"reps":"6","weight_kg":100,"rest_seconds":150,"cues":["Brace before descent"]}]},
 {"name":"Superset","exercises":[
   {"name":"Bulgarian Split Squat","sets":3,"reps":"8","per_side":true,"weight_kg":16,"load_note":"per dumbbell","rest_seconds":0},
   {"name":"Plank","sets":3,"reps":"45s","rest_seconds":90}]}]`

type rangeInput struct {
	StartDate string `json:"start_date" jsonschema:"first day, YYYY-MM-DD"`
	EndDate   string `json:"end_date" jsonschema:"last day (inclusive), YYYY-MM-DD"`
}

// planEntry is one session in list_training_plan, run or strength.
type planEntry struct {
	Date string `json:"date"`
	Type string `json:"type" jsonschema:"strength (Gym Rabbit) or run (any FinalSurge calendar entry)"`
	Name string `json:"name"`
	ID   string `json:"id" jsonschema:"strength: workout ID for get/update/delete_strength_workout; run: FinalSurge workout_key"`

	// strength
	Exercises     int    `json:"exercises,omitempty"`
	LastPerformed string `json:"last_performed,omitempty" jsonschema:"strength: when the session was last started in the app"`

	// run
	Activity       string  `json:"activity,omitempty" jsonschema:"run: FinalSurge activity type"`
	Status         string  `json:"status,omitempty" jsonschema:"run: planned or completed status from FinalSurge"`
	PlannedSeconds float64 `json:"planned_seconds,omitempty"`
	PlannedKm      float64 `json:"planned_km,omitempty"`
	HasStructure   bool    `json:"has_structure,omitempty" jsonschema:"run: has a Workout Builder structure"`
}

type planOutput struct {
	Sessions []planEntry `json:"sessions"`
	Note     string      `json:"note,omitempty"`
}

type idInput struct {
	ID string `json:"id" jsonschema:"the strength workout ID (from list_training_plan or upload_training_plan)"`
}

type updateInput struct {
	ID            string       `json:"id" jsonschema:"the strength workout ID to update"`
	Title         string       `json:"title" jsonschema:"workout title"`
	ScheduledDate string       `json:"scheduled_date" jsonschema:"planned day, YYYY-MM-DD"`
	Notes         string       `json:"notes,omitempty" jsonschema:"note for the athlete; omitting it clears the current note"`
	Blocks        []plan.Block `json:"blocks" jsonschema:"the full replacement list of blocks"`
}

type okOutput struct {
	OK bool `json:"ok"`
}

// New builds the MCP server backed by the given database. If fs is non-nil,
// runs are enabled (in upload_training_plan, list_training_plan and the run tools).
func New(database *db.DB, fs *finalsurge.Client) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "gym-rabbit", Title: "Gym Rabbit", Version: "2.0.0"}, &mcp.ServerOptions{
		Instructions: "Gym Rabbit manages a dated training plan: strength workouts (structured blocks of exercises the " +
			"athlete runs in the gym app) and, if FinalSurge is configured, runs in the FinalSurge calendar. " +
			"Add sessions with upload_training_plan, see the plan with list_training_plan. Every write is validated " +
			"and errors name the exact field to fix.",
	})

	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_training_plan",
		Description: "List every session between two dates (inclusive), runs and strength together, ordered by date.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in rangeInput) (*mcp.CallToolResult, planOutput, error) {
		start, err := parseDate("start_date", in.StartDate)
		if err != nil {
			return nil, planOutput{}, err
		}
		end, err := parseDate("end_date", in.EndDate)
		if err != nil {
			return nil, planOutput{}, err
		}
		if end.Before(start) {
			return nil, planOutput{}, fmt.Errorf("end_date is before start_date")
		}

		workouts, err := database.ListWorkoutsBetween(in.StartDate, in.EndDate)
		if err != nil {
			return nil, planOutput{}, fmt.Errorf("failed to list strength workouts: %w", err)
		}
		out := planOutput{Sessions: []planEntry{}}
		for _, w := range workouts {
			n := 0
			for _, b := range w.Blocks {
				n += len(b.Exercises)
			}
			out.Sessions = append(out.Sessions, planEntry{
				Date: w.ScheduledDate, Type: "strength", Name: w.Title, ID: w.ID,
				Exercises: n, LastPerformed: w.LastPerformed,
			})
		}

		if fs == nil {
			out.Note = "FinalSurge is not configured on this server, so runs are not listed."
		} else {
			runs, err := fs.ListWorkouts(ctx, in.StartDate, in.EndDate)
			if err != nil {
				return nil, planOutput{}, fmt.Errorf("failed to list FinalSurge workouts: %w", err)
			}
			for _, r := range runs {
				out.Sessions = append(out.Sessions, planEntry{
					Date: r.Date, Type: "run", Name: r.Name, ID: r.Key,
					Activity: r.Activity, Status: r.Status, PlannedSeconds: r.PlannedSeconds,
					PlannedKm: r.PlannedAmount, HasStructure: r.HasStructure,
				})
			}
		}
		sort.SliceStable(out.Sessions, func(i, j int) bool { return out.Sessions[i].Date < out.Sessions[j].Date })
		return nil, out, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_strength_workout",
		Description: "Get a strength workout with all its blocks and exercises.",
		Annotations: readOnly,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, db.Workout, error) {
		w, err := database.GetWorkout(in.ID)
		if err != nil {
			return nil, db.Workout{}, fmt.Errorf("failed to get workout: %w", err)
		}
		if w == nil {
			return nil, db.Workout{}, fmt.Errorf("workout not found: %s", in.ID)
		}
		return nil, *w, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "update_strength_workout",
		Description: "Replace a strength workout: title, date, notes and blocks are all overwritten. Read it with " +
			"get_strength_workout first and send back the full, edited workout.\n\n" + strengthHelp,
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updateInput) (*mcp.CallToolResult, db.Workout, error) {
		w, err := newWorkout(in.Title, in.Notes, in.ScheduledDate, in.Blocks)
		if err != nil {
			return nil, db.Workout{}, err
		}
		w.ID = in.ID
		if err := database.UpdateWorkout(w); err != nil {
			return nil, db.Workout{}, err
		}
		updated, err := database.GetWorkout(in.ID)
		if err != nil || updated == nil {
			return nil, db.Workout{}, fmt.Errorf("failed to reload workout %s", in.ID)
		}
		return nil, *updated, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_strength_workout",
		Description: "Permanently delete a strength workout.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, okOutput, error) {
		if err := database.DeleteWorkout(in.ID); err != nil {
			return nil, okOutput{}, err
		}
		return nil, okOutput{OK: true}, nil
	})

	registerTrainingPlan(s, database, fs)
	if fs != nil {
		registerFinalSurge(s, fs)
	}

	return s
}

// newWorkout normalizes and validates a strength workout, reporting every
// problem at once. It assigns a new ID.
func newWorkout(title, notes, scheduledDate string, blocks []plan.Block) (*db.Workout, error) {
	title, notes = strings.TrimSpace(title), strings.TrimSpace(notes)
	plan.Normalize(blocks)

	var problems []string
	if title == "" {
		problems = append(problems, "title: is required")
	}
	if _, err := parseDate("scheduled_date", scheduledDate); err != nil {
		problems = append(problems, err.Error())
	}
	if err := plan.Validate(blocks); err != nil {
		problems = append(problems, strings.TrimPrefix(err.Error(), "invalid workout:\n- "))
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid workout:\n- %s", strings.Join(problems, "\n- "))
	}
	return &db.Workout{ID: db.NewID(), Title: title, Notes: notes, ScheduledDate: scheduledDate, Blocks: blocks}, nil
}

func parseDate(field, date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return t, fmt.Errorf("%s: must be YYYY-MM-DD, got %q", field, date)
	}
	return t, nil
}
