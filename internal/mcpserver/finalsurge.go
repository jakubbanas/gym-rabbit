package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
)

// Run tools, ported from ../finalsurge-mcp/main.go.

const structureHelp = `Structure is a list of blocks. A block is either a single step or a repeat.
Step fields: intensity (warmup|active|rest|cooldown), exactly one of time ("m:ss"/"h:mm:ss") / distance_km / lap_button,
optional pace_fast + pace_slow in min/km (e.g. "5:25","5:35"), optional name.
Repeat block: {"repeats": 12, "steps": [ {...}, {...} ]}.
Example (12x400m): [{"intensity":"warmup","time":"10:00","pace_fast":"5:40","pace_slow":"5:50"},
{"repeats":12,"steps":[{"distance_km":0.4,"pace_fast":"4:30","pace_slow":"4:40"},{"intensity":"rest","time":"1:30"}]},
{"intensity":"cooldown","time":"10:00"}]`

type fsStructureIn struct {
	WorkoutKey string             `json:"workout_key" jsonschema:"key of an existing workout (from list_training_plan or upload_training_plan)"`
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
func prepareRun(in planSession) (*preparedRun, error) {
	if _, err := parseDate("date", in.Date); err != nil {
		return nil, err
	}
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	p := &preparedRun{workout: finalsurge.NewRunWorkout{
		Date: in.Date, Name: in.Name, Description: in.Notes,
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
			res.Note = "workout was created but the structure was not saved; retry with update_run_structure"
		} else {
			res.StructureSaved = true
		}
	}
	return res, nil
}

// registerFinalSurge adds the run-specific tools. Runs are created with
// upload_training_plan and listed with list_training_plan.
func registerFinalSurge(s *mcp.Server, fs *finalsurge.Client) {
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_run_structure",
		Description: "Read the raw Workout Builder structure of a FinalSurge run (as stored by FinalSurge).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in fsKeyIn) (*mcp.CallToolResult, any, error) {
		raw, err := fs.GetStructure(ctx, in.WorkoutKey)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_run_structure",
		Description: "Replace the Workout Builder structure (steps, repeats, pace targets) of an existing FinalSurge run.\n\n" + structureHelp,
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
		Name:        "delete_run",
		Description: "Permanently delete a workout from the FinalSurge calendar. Cannot be undone; confirm with the user first.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in fsKeyIn) (*mcp.CallToolResult, any, error) {
		if err := fs.DeleteWorkout(ctx, in.WorkoutKey); err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"workout_key": in.WorkoutKey, "deleted": true}, nil
	})
}
