package plan

import "testing"

func TestParse(t *testing.T) {
	p, err := Parse("Plan Name: Legs\n\nBlock A\nA1. Squat | 3x8 | 100kg | 90s\n- brace\nA2. Split Squat | 3x8+8 | 16kg DBs | 60s\n\nBlock B\nLunge | 10 | BW | rest")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Legs" || len(p.Blocks) != 2 {
		t.Fatalf("got name %q, %d blocks", p.Name, len(p.Blocks))
	}
	sq, split := p.Blocks[0].Exercises[0], p.Blocks[0].Exercises[1]
	if sq.ID != "A1" || sq.Sets != 3 || sq.Reps != "8" || sq.Rest != 90 || len(sq.Cues) != 1 || sq.IsUnilateral {
		t.Errorf("squat parsed wrong: %+v", sq)
	}
	if !split.IsUnilateral {
		t.Errorf("split squat should be unilateral")
	}
	lunge := p.Blocks[1].Exercises[0]
	if lunge.ID != "Ex1" || lunge.Sets != 3 || lunge.Reps != "10" || lunge.Rest != 60 {
		t.Errorf("defaults applied wrong: %+v", lunge)
	}
}

func TestParseErrors(t *testing.T) {
	for _, text := range []string{"", "Plan Name: X", "Block A\nA1. Squat | 3x8 | 100kg"} {
		if _, err := Parse(text); err == nil {
			t.Errorf("expected error for %q", text)
		}
	}
}
