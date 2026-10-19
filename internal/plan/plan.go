// Package plan defines the structure of a strength session. It is the format
// AI agents write through MCP and the web app reads, so the jsonschema tags
// double as the documentation the agent sees.
package plan

import (
	"fmt"
	"strings"
)

const (
	maxBlocks    = 26 // labelled A–Z
	maxExercises = 20
	maxSets      = 50
	maxRest      = 3600
)

// Exercise is one exercise with its sets.
type Exercise struct {
	Name        string   `json:"name" jsonschema:"exercise name, e.g. Back Squat"`
	Sets        int      `json:"sets" jsonschema:"number of working sets"`
	Reps        string   `json:"reps" jsonschema:"reps per set, e.g. 8 or 8-10 or AMRAP; for timed holds use seconds like 45s"`
	PerSide     bool     `json:"per_side,omitempty" jsonschema:"true if reps are per side (single-arm/single-leg); the app then logs left and right separately"`
	WeightKg    float64  `json:"weight_kg,omitempty" jsonschema:"load in kg per side/implement as you'd read it on the bar or dumbbell; omit for bodyweight"`
	LoadNote    string   `json:"load_note,omitempty" jsonschema:"optional load detail shown next to the weight, e.g. per dumbbell, RPE 8, red band"`
	RestSeconds int      `json:"rest_seconds" jsonschema:"rest after each set in seconds; 0 to go straight to the next exercise in a superset"`
	Cues        []string `json:"cues,omitempty" jsonschema:"optional short technique cues shown on the exercise card"`
}

// Block is a group of exercises done together (a superset or circuit). For
// straight sets, use one exercise per block.
type Block struct {
	Name      string     `json:"name,omitempty" jsonschema:"optional label; defaults to Block A, Block B, …"`
	Exercises []Exercise `json:"exercises" jsonschema:"exercises in this block, done in order"`
}

// Normalize fills defaults (block names) and trims text fields in place.
func Normalize(blocks []Block) {
	for i := range blocks {
		b := &blocks[i]
		b.Name = strings.TrimSpace(b.Name)
		if b.Name == "" && i < maxBlocks {
			b.Name = "Block " + string(rune('A'+i))
		}
		for j := range b.Exercises {
			e := &b.Exercises[j]
			e.Name = strings.TrimSpace(e.Name)
			e.Reps = strings.TrimSpace(e.Reps)
			e.LoadNote = strings.TrimSpace(e.LoadNote)
			cues := e.Cues[:0]
			for _, c := range e.Cues {
				if c = strings.TrimSpace(c); c != "" {
					cues = append(cues, c)
				}
			}
			e.Cues = cues
		}
	}
}

// Validate checks blocks and reports every problem with its path
// (e.g. blocks[1].exercises[0].sets), so an agent can fix them all at once.
func Validate(blocks []Block) error {
	var problems []string
	add := func(path, msg string, args ...any) {
		problems = append(problems, path+": "+fmt.Sprintf(msg, args...))
	}

	if len(blocks) == 0 {
		add("blocks", "at least one block is required")
	}
	if len(blocks) > maxBlocks {
		add("blocks", "at most %d blocks", maxBlocks)
	}
	for i, b := range blocks {
		bp := fmt.Sprintf("blocks[%d]", i)
		if len(b.Exercises) == 0 {
			add(bp+".exercises", "at least one exercise is required")
		}
		if len(b.Exercises) > maxExercises {
			add(bp+".exercises", "at most %d exercises per block", maxExercises)
		}
		for j, e := range b.Exercises {
			ep := fmt.Sprintf("%s.exercises[%d]", bp, j)
			if strings.TrimSpace(e.Name) == "" {
				add(ep+".name", "is required")
			}
			if e.Sets < 1 || e.Sets > maxSets {
				add(ep+".sets", "must be between 1 and %d, got %d", maxSets, e.Sets)
			}
			if strings.TrimSpace(e.Reps) == "" {
				add(ep+".reps", "is required")
			}
			if e.WeightKg < 0 {
				add(ep+".weight_kg", "must not be negative")
			}
			if e.RestSeconds < 0 || e.RestSeconds > maxRest {
				add(ep+".rest_seconds", "must be between 0 and %d, got %d", maxRest, e.RestSeconds)
			}
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid workout:\n- %s", strings.Join(problems, "\n- "))
	}
	return nil
}
