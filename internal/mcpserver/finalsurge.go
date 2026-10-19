package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
)

// Ported from ../finalsurge-mcp/main.go. Tools are prefixed with finalsurge_
// so they don't clash with the Gym Rabbit workout tools.

const structureHelp = `Structure is a list of blocks. A block is either a single step or a repeat.
Step fields: intensity (warmup|active|rest|cooldown), exactly one of time ("m:ss"/"h:mm:ss") / distance_km / lap_button,
optional pace_fast + pace_slow in min/km (e.g. "5:25","5:35"), optional name.
Repeat block: {"repeats": 12, "steps": [ {...}, {...} ]}.
Example (12x400m): [{"intensity":"warmup","time":"10:00","pace_fast":"5:40","pace_slow":"5:50"},
{"repeats":12,"steps":[{"distance_km":0.4,"pace_fast":"4:30","pace_slow":"4:40"},{"intensity":"rest","time":"1:30"}]},
{"intensity":"cooldown","time":"10:00"}]`

type fsListIn struct {
	StartDate string `json:"start_date" jsonschema:"first day, YYYY-MM-DD"`
	EndDate   string `json:"end_date" jsonschema:"last day (inclusive), YYYY-MM-DD"`
}

type runIn struct {
	Date            string             `json:"date" jsonschema:"workout date, YYYY-MM-DD"`
	Name            string             `json:"name" jsonschema:"workout name (also shown on the watch)"`
	Description     string             `json:"description,omitempty" jsonschema:"free-text description, e.g. the coach's original notes"`
	PlannedMinutes  int                `json:"planned_minutes,omitempty" jsonschema:"planned duration in minutes; if omitted and structure is given, it is estimated from the structure"`
	PlannedDistance float64            `json:"planned_distance_km,omitempty" jsonschema:"planned distance in km"`
	Structure       []finalsurge.Block `json:"structure,omitempty" jsonschema:"optional Workout Builder structure (see tool description)"`
}

type fsStructureIn struct {
	WorkoutKey string             `json:"workout_key" jsonschema:"key of an existing workout (from finalsurge_list_workouts or finalsurge_create_run_workout)"`
	Name       string             `json:"name" jsonschema:"name shown on the device"`
	Structure  []finalsurge.Block `json:"structure" jsonschema:"Workout Builder structure (see tool description)"`
}

type fsKeyIn struct {
	WorkoutKey string `json:"workout_key" jsonschema:"FinalSurge workout key"`
}

type runResult struct {
	WorkoutKey              string `json:"workout_key"`
	Date                    string `json:"date"`
	PlannedSeconds          int    `json:"planned_seconds"`
	PlannedSecondsEstimated bool   `json:"planned_seconds_estimated"`
	StructureSaved          bool   `json:"structure_saved,omitempty"`
	StructureError          string `json:"structure_error,omitempty"`
	Note                    string `json:"note,omitempty"`
}

// preparedRun is a validated run, ready to be sent to FinalSurge.
type preparedRun struct {
	workout   finalsurge.NewRunWorkout
	payload   []byte // Workout Builder JSON, nil if no structure
	estimated bool
}

// prepareRun validates a run and builds its payloads without calling FinalSurge.
func prepareRun(in runIn) (*preparedRun, error) {
	if _, err := time.Parse("2006-01-02", in.Date); err != nil {
		return nil, fmt.Errorf("date must be YYYY-MM-DD, got %q", in.Date)
	}
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	p := &preparedRun{workout: finalsurge.NewRunWorkout{
		Date: in.Date, Name: in.Name, Description: in.Description,
		PlannedSeconds: in.PlannedMinutes * 60, PlannedDistance: in.PlannedDistance,
	}}
	if len(in.Structure) > 0 {
		var err error
		if p.payload, err = finalsurge.BuildStructure(in.Name, in.Structure, time.Now()); err != nil {
			return nil, fmt.Errorf("invalid structure: %w", err)
		}
		if p.workout.PlannedSeconds == 0 {
			if est, ok := finalsurge.EstimateSeconds(in.Structure); ok && est > 0 {
				p.workout.PlannedSeconds, p.estimated = est, true
			}
		}
	}
	return p, nil
}

// create sends the run to FinalSurge. If the structure fails after the workout
// was created, it still returns the result (with StructureError), because the workout exists.
func (p *preparedRun) create(ctx context.Context, fs *finalsurge.Client) (runResult, error) {
	key, err := fs.CreateRunWorkout(ctx, p.workout)
	if err != nil {
		return runResult{}, err
	}
	res := runResult{WorkoutKey: key, Date: p.workout.Date, PlannedSeconds: p.workout.PlannedSeconds, PlannedSecondsEstimated: p.estimated}
	if p.payload != nil {
		if err := fs.SaveStructure(ctx, key, p.payload); err != nil {
			res.StructureError = err.Error()
			res.Note = "workout was created but the structure was not saved; retry with finalsurge_set_workout_structure"
		} else {
			res.StructureSaved = true
		}
	}
	return res, nil
}

func registerFinalSurge(s *mcp.Server, fs *finalsurge.Client) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finalsurge_list_workouts",
		Description: "List planned and completed workouts in the FinalSurge calendar between two dates (inclusive).",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in fsListIn) (*mcp.CallToolResult, any, error) {
		ws, err := fs.ListWorkouts(ctx, in.StartDate, in.EndDate)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"workouts": ws}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "finalsurge_create_run_workout",
		Description: "Create a planned run in FinalSurge on a given date, optionally with a structured Workout Builder plan " +
			"(steps, repeats, pace targets) that syncs to the watch. Returns the new workout_key. " +
			"To schedule several runs and strength sessions at once, use upload_training_plan.\n\n" + structureHelp,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in runIn) (*mcp.CallToolResult, runResult, error) {
		p, err := prepareRun(in)
		if err != nil {
			return nil, runResult{}, err
		}
		res, err := p.create(ctx, fs)
		return nil, res, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finalsurge_set_workout_structure",
		Description: "Replace the Workout Builder structure of an existing FinalSurge workout.\n\n" + structureHelp,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in fsStructureIn) (*mcp.CallToolResult, any, error) {
		payload, err := finalsurge.BuildStructure(in.Name, in.Structure, time.Now())
		if err != nil {
			return nil, nil, fmt.Errorf("invalid structure: %w", err)
		}
		if err := fs.SaveStructure(ctx, in.WorkoutKey, payload); err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"workout_key": in.WorkoutKey, "structure_saved": true}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finalsurge_get_workout_structure",
		Description: "Read the raw Workout Builder structure of a FinalSurge workout (as stored by FinalSurge).",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in fsKeyIn) (*mcp.CallToolResult, any, error) {
		raw, err := fs.GetStructure(ctx, in.WorkoutKey)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "finalsurge_delete_workout",
		Description: "Permanently delete a workout from the FinalSurge calendar. Cannot be undone; confirm with the user first.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in fsKeyIn) (*mcp.CallToolResult, any, error) {
		if err := fs.DeleteWorkout(ctx, in.WorkoutKey); err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"workout_key": in.WorkoutKey, "deleted": true}, nil
	})
}
