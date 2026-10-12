// Package plan is a Go port of the frontend "Magic Paste" parser
// (frontend/src/lib/parser.ts). It lets the server validate plan text
// before saving it, so plans created outside the web UI still load in gym mode.
// Keep the two parsers in sync.
package plan

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Exercise is a single exercise line within a block.
type Exercise struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Sets         int      `json:"sets"`
	Reps         string   `json:"reps"`
	Weight       string   `json:"weight"`
	IsUnilateral bool     `json:"isUnilateral"`
	Rest         int      `json:"rest"`
	Cues         []string `json:"cues"`
}

// Block groups exercises, e.g. "Block A".
type Block struct {
	Name      string     `json:"name"`
	Exercises []Exercise `json:"exercises"`
}

// WorkoutPlan is the parsed structure of a plan's text.
type WorkoutPlan struct {
	Name   string  `json:"name"`
	Blocks []Block `json:"blocks"`
}

var (
	idNameRe   = regexp.MustCompile(`^([A-Za-z0-9]+)\.\s+(.*)$`)
	setsRepsRe = regexp.MustCompile(`(\d+)x(.*)`)
	restRe     = regexp.MustCompile(`(\d+)`)
)

// Parse converts Magic Paste text into a WorkoutPlan, mirroring parseWorkout in parser.ts.
func Parse(text string) (*WorkoutPlan, error) {
	p := &WorkoutPlan{Name: "Gym Session", Blocks: []Block{}}
	var current *Block

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "Plan Name:"):
			p.Name = strings.TrimSpace(strings.Replace(line, "Plan Name:", "", 1))
		case strings.HasPrefix(line, "Block"):
			p.Blocks = append(p.Blocks, Block{Name: line, Exercises: []Exercise{}})
			current = &p.Blocks[len(p.Blocks)-1]
		case strings.HasPrefix(line, "-"):
			if current != nil && len(current.Exercises) > 0 {
				last := &current.Exercises[len(current.Exercises)-1]
				last.Cues = append(last.Cues, strings.TrimSpace(line[1:]))
			}
		default:
			parts := strings.Split(line, "|")
			if len(parts) != 4 {
				return nil, fmt.Errorf("formatting error on line: %q: must be exactly 4 parts separated by pipes (|)", line)
			}
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			if current == nil {
				p.Blocks = append(p.Blocks, Block{Name: "Block A", Exercises: []Exercise{}})
				current = &p.Blocks[len(p.Blocks)-1]
			}

			idName, setsReps, weight, rest := parts[0], parts[1], parts[2], parts[3]

			id := fmt.Sprintf("Ex%d", len(current.Exercises)+1)
			name := idName
			if m := idNameRe.FindStringSubmatch(idName); m != nil {
				id, name = m[1], m[2]
			}

			sets, reps := 3, setsReps
			if m := setsRepsRe.FindStringSubmatch(setsReps); m != nil {
				sets, _ = strconv.Atoi(m[1])
				reps = m[2]
			}

			restSeconds := 60
			if m := restRe.FindStringSubmatch(rest); m != nil {
				restSeconds, _ = strconv.Atoi(m[1])
			}

			current.Exercises = append(current.Exercises, Exercise{
				ID:           id,
				Name:         name,
				Sets:         sets,
				Reps:         reps,
				Weight:       weight,
				IsUnilateral: strings.Contains(reps, "+"),
				Rest:         restSeconds,
				Cues:         []string{},
			})
		}
	}

	if len(p.Blocks) == 0 {
		return nil, errors.New("no workout blocks found: make sure to use 'Block A' to define sections")
	}
	return p, nil
}
