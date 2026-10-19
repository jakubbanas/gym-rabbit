package plan

import (
	"strings"
	"testing"
)

func TestNormalizeNamesBlocks(t *testing.T) {
	blocks := []Block{
		{Exercises: []Exercise{{Name: " Squat ", Sets: 3, Reps: "5", Cues: []string{" brace ", " "}}}},
		{Name: "Finisher", Exercises: []Exercise{{Name: "Plank", Sets: 2, Reps: "45s"}}},
	}
	Normalize(blocks)
	if blocks[0].Name != "Block A" || blocks[1].Name != "Finisher" {
		t.Errorf("block names: %q, %q", blocks[0].Name, blocks[1].Name)
	}
	if e := blocks[0].Exercises[0]; e.Name != "Squat" || len(e.Cues) != 1 || e.Cues[0] != "brace" {
		t.Errorf("trimmed exercise: %+v", e)
	}
}

func TestValidateReportsEveryProblemWithPath(t *testing.T) {
	err := Validate([]Block{
		{Exercises: []Exercise{{Name: "Squat", Sets: 0, Reps: "5", RestSeconds: 90}}},
		{Exercises: []Exercise{}},
		{Exercises: []Exercise{{Sets: 3, Reps: "", WeightKg: -5, RestSeconds: -1}}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{
		"blocks[0].exercises[0].sets",
		"blocks[1].exercises: at least one",
		"blocks[2].exercises[0].name",
		"blocks[2].exercises[0].reps",
		"blocks[2].exercises[0].weight_kg",
		"blocks[2].exercises[0].rest_seconds",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in:\n%v", want, err)
		}
	}
	if Validate(nil) == nil {
		t.Error("empty plan should be invalid")
	}
}
