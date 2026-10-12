// Package mcpserver exposes Gym Rabbit's workout plans over the Model Context
// Protocol, so AI assistants can list, read, write and validate plans.
package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

const formatURI = "gymrabbit://format"

// formatSpec documents the Magic Paste format for the model.
const formatSpec = `# Gym Rabbit "Magic Paste" plan format

Plan Name: [Descriptive Plan Title]

Block [A-Z]
[Exercise ID]. [Exercise Name] | [Sets]x[Reps] | [Weight] | [Rest]
- [Optional technique cue or note line]

Rules:
- Every exercise line has exactly 4 fields separated by pipes (|).
- At least one "Block ..." line is required.
- Reps containing "+" (e.g. 8+8) mark the exercise as unilateral (left/right logging).
- Rest is in seconds; the first number in the field is used (e.g. 90s -> 90).
- Weight is free text (100kg, 20kg DBs, Bodyweight).
- Lines starting with "-" directly below an exercise are form cues for that exercise.

Example:

Plan Name: Lower Body Strength

Block A
A1. Back Squat | 4x6 | 100kg | 120s
- Brace before descent
A2. Bulgarian Split Squat | 3x8+8 | 16kg DBs | 90s

Block B
B1. Romanian Deadlift | 3x10 | 80kg | 90s
`

type workoutSummary struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	LastPerformed string `json:"lastPerformed" jsonschema:"RFC3339 timestamp of the last completed session, empty if never performed"`
	UpdatedAt     string `json:"updatedAt"`
}

type listOutput struct {
	Workouts []workoutSummary `json:"workouts"`
}

type idInput struct {
	ID string `json:"id" jsonschema:"the workout ID, e.g. w-1735689600000000000"`
}

type workoutOutput struct {
	db.Workout
	Plan *plan.WorkoutPlan `json:"plan,omitempty" jsonschema:"the parsed plan structure, omitted if the stored text does not parse"`
}

type writeInput struct {
	Title string `json:"title" jsonschema:"display title for the saved plan"`
	Text  string `json:"text" jsonschema:"the plan in Magic Paste format; read the gymrabbit://format resource for the spec"`
}

type updateInput struct {
	ID    string `json:"id" jsonschema:"the workout ID to update"`
	Title string `json:"title" jsonschema:"display title for the saved plan"`
	Text  string `json:"text" jsonschema:"the full replacement plan text in Magic Paste format"`
}

type validateInput struct {
	Text string `json:"text" jsonschema:"plan text in Magic Paste format"`
}

type okOutput struct {
	OK bool `json:"ok"`
}

// New builds the MCP server backed by the given database.
func New(database *db.DB) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "gym-rabbit", Title: "Gym Rabbit", Version: "1.0.0"}, &mcp.ServerOptions{
		Instructions: "Gym Rabbit stores workout plans written in a pipe-delimited plain-text format. " +
			"Read the " + formatURI + " resource before writing plan text, and use validate_plan to check drafts.",
	})

	s.AddResource(&mcp.Resource{
		URI:         formatURI,
		Name:        "plan-format",
		Title:       "Magic Paste plan format",
		Description: "Specification and example of the plain-text workout plan format",
		MIMEType:    "text/markdown",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{
			{URI: formatURI, MIMEType: "text/markdown", Text: formatSpec},
		}}, nil
	})

	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_workouts",
		Description: "List saved workout plans, most recently updated first.",
		Annotations: readOnly,
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listOutput, error) {
		workouts, err := database.ListWorkouts()
		if err != nil {
			return nil, listOutput{}, fmt.Errorf("failed to list workouts: %w", err)
		}
		out := listOutput{Workouts: make([]workoutSummary, 0, len(workouts))}
		for _, w := range workouts {
			out.Workouts = append(out.Workouts, workoutSummary{ID: w.ID, Title: w.Title, LastPerformed: w.LastPerformed, UpdatedAt: w.UpdatedAt})
		}
		return nil, out, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_workout",
		Description: "Get a saved workout plan's raw text and its parsed structure.",
		Annotations: readOnly,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, workoutOutput, error) {
		w, err := database.GetWorkout(in.ID)
		if err != nil {
			return nil, workoutOutput{}, fmt.Errorf("failed to get workout: %w", err)
		}
		if w == nil {
			return nil, workoutOutput{}, fmt.Errorf("workout not found: %s", in.ID)
		}
		parsed, _ := plan.Parse(w.Text)
		return nil, workoutOutput{Workout: *w, Plan: parsed}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "validate_plan",
		Description: "Parse plan text without saving it. Returns the parsed structure, or an error describing the first bad line.",
		Annotations: readOnly,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in validateInput) (*mcp.CallToolResult, *plan.WorkoutPlan, error) {
		parsed, err := plan.Parse(in.Text)
		if err != nil {
			return nil, nil, err
		}
		return nil, parsed, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_workout",
		Description: "Save a new workout plan. The text is validated against the Magic Paste format first.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in writeInput) (*mcp.CallToolResult, db.Workout, error) {
		if err := validateWrite(in.Title, in.Text); err != nil {
			return nil, db.Workout{}, err
		}
		w := db.Workout{ID: db.NewID(), Title: in.Title, Text: in.Text}
		if err := database.CreateWorkout(&w); err != nil {
			return nil, db.Workout{}, fmt.Errorf("failed to create workout: %w", err)
		}
		return nil, w, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_workout",
		Description: "Replace the title and text of an existing workout plan. The text is validated first.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updateInput) (*mcp.CallToolResult, db.Workout, error) {
		if err := validateWrite(in.Title, in.Text); err != nil {
			return nil, db.Workout{}, err
		}
		if err := database.UpdateWorkout(&db.Workout{ID: in.ID, Title: in.Title, Text: in.Text}); err != nil {
			return nil, db.Workout{}, err
		}
		w, err := database.GetWorkout(in.ID)
		if err != nil || w == nil {
			return nil, db.Workout{}, fmt.Errorf("failed to reload workout %s", in.ID)
		}
		return nil, *w, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_workout",
		Description: "Permanently delete a saved workout plan.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, okOutput, error) {
		if err := database.DeleteWorkout(in.ID); err != nil {
			return nil, okOutput{}, err
		}
		return nil, okOutput{OK: true}, nil
	})

	return s
}

func validateWrite(title, text string) error {
	if title == "" || text == "" {
		return fmt.Errorf("title and text are required")
	}
	if _, err := plan.Parse(text); err != nil {
		return fmt.Errorf("invalid plan text (see %s): %w", formatURI, err)
	}
	return nil
}
