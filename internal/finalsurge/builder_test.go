package finalsurge

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Captured from the web app on 2026-10-12 (Biegowe #1), accepted by the server.
const capturedEasyRun = `{"target_options":[{"name":"Biegowe #1 – 60' easy","sport":"running","steps":[{"type":"step","id":101,"name":null,"durationType":"TIME","duration":"1:00:00","targetAbsOrPct":"","durationDist":0,"data":[],"distUnit":"km","target":[{"targetType":"pace","zoneBased":false,"targetLow":"5:50","targetHigh":"7:00","targetOption":"min/km","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":true,"zone":0},{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":false,"zone":0}],"intensity":"ACTIVE","comments":null}],"target":"pace","metadata":{"version":"1","editors":[{"source":"webapp","version":"0.0.1","calendar":"2026-10-12 14:33","library":"","plan":"","upload":""}]}}],"target_override":null}`

// Captured dry-run (repeat format) from the web app.
const capturedRepeat = `{"target_options":[{"name":"x","sport":"running","steps":[{"type":"step","id":101,"name":null,"durationType":"TIME","duration":"5:00","targetAbsOrPct":"","durationDist":0,"data":[],"distUnit":"km","target":[{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":false,"zone":0},{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":true,"zone":0}],"intensity":"WARMUP","comments":null},{"type":"repeat","name":null,"id":5001,"data":[{"type":"step","id":102,"name":null,"durationType":"DISTANCE","duration":"00:00","targetAbsOrPct":"","durationDist":0.4,"data":[],"distUnit":"km","target":[{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":false,"zone":0},{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":true,"zone":0}],"intensity":"ACTIVE","comments":null},{"type":"step","id":103,"name":null,"durationType":"TIME","duration":"1:30","targetAbsOrPct":"","durationDist":0,"data":[],"distUnit":"km","target":[{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":false,"zone":0},{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":true,"zone":0}],"intensity":"REST","comments":null}],"repeats":2,"skip_last_rest":false,"intensity":"ACTIVE","durationType":"OPEN","comments":null},{"type":"step","id":104,"name":null,"durationType":"OPEN","duration":"00:00","targetAbsOrPct":"","durationDist":0,"data":[],"distUnit":"km","target":[{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":false,"zone":0},{"targetType":"open","zoneBased":false,"targetLow":"0","targetHigh":"0","targetOption":"","targetTypeOriginal":null,"targetOptionOriginal":null,"targetIsTimeBased":true,"zone":0}],"intensity":"COOLDOWN","comments":null}],"target":"pace","metadata":{"version":"1","editors":[{"source":"webapp","version":"0.0.1","calendar":"2026-10-12 14:35","library":"","plan":"","upload":""}]}}],"target_override":null}`

func mustEqualJSON(t *testing.T, got []byte, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("JSON mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestBuildMatchesCapturedEasyRun(t *testing.T) {
	// The web app sent targetIsTimeBased=false on the second (open) target of a
	// paced step; our builder always uses true there. Normalise before compare.
	got, err := BuildStructure("Biegowe #1 – 60' easy", []Block{
		{Step: Step{Time: "60:00", PaceFast: "5:50", PaceSlow: "7:00"}},
	}, time.Date(2026, 10, 12, 14, 33, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var g map[string]any
	_ = json.Unmarshal(got, &g)
	step := g["target_options"].([]any)[0].(map[string]any)["steps"].([]any)[0].(map[string]any)
	step["target"].([]any)[1].(map[string]any)["targetIsTimeBased"] = false
	got, _ = json.Marshal(g)
	mustEqualJSON(t, got, capturedEasyRun)
}

func TestBuildMatchesCapturedRepeat(t *testing.T) {
	got, err := BuildStructure("x", []Block{
		{Step: Step{Intensity: "warmup", Time: "5:00"}},
		{Repeats: 2, Steps: []Step{{DistanceKm: 0.4}, {Intensity: "rest", Time: "1:30"}}},
		{Step: Step{Intensity: "cooldown", LapButton: true}},
	}, time.Date(2026, 10, 12, 14, 35, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	mustEqualJSON(t, got, capturedRepeat)
}

func TestValidation(t *testing.T) {
	bad := [][]Block{
		{},
		{{Step: Step{}}}, // no duration
		{{Step: Step{Time: "5:00", DistanceKm: 1}}},       // two durations
		{{Step: Step{Time: "abc"}}},                       // bad time
		{{Step: Step{Time: "5:00", PaceFast: "4"}}},       // bad pace
		{{Step: Step{Time: "5:00", Intensity: "sprint"}}}, // bad intensity
		{{Repeats: 3}},                    // repeat w/o steps
		{{Steps: []Step{{Time: "1:00"}}}}, // steps w/o repeats
	}
	for i, b := range bad {
		if _, err := BuildStructure("x", b, time.Now()); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestPaceOrderAndTimeNorm(t *testing.T) {
	got, err := BuildStructure("x", []Block{{Step: Step{Time: "75:00", PaceFast: "5:35", PaceSlow: "5:25"}}}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var g map[string]any
	_ = json.Unmarshal(got, &g)
	st := g["target_options"].([]any)[0].(map[string]any)["steps"].([]any)[0].(map[string]any)
	if st["duration"] != "1:15:00" {
		t.Errorf("duration = %v", st["duration"])
	}
	tg := st["target"].([]any)[0].(map[string]any)
	if tg["targetLow"] != "5:25" || tg["targetHigh"] != "5:35" {
		t.Errorf("pace = %v-%v", tg["targetLow"], tg["targetHigh"])
	}
}

func TestEstimate(t *testing.T) {
	// Biegowe #2 from plan 10.0
	blocks := []Block{
		{Step: Step{Intensity: "warmup", Time: "5:00"}},
		{Step: Step{Time: "15:00"}},
		{Step: Step{Intensity: "rest", Time: "1:30"}},
		{Repeats: 3, Steps: []Step{{Time: "0:15"}, {Intensity: "rest", Time: "1:30"}}},
		{Repeats: 12, Steps: []Step{{DistanceKm: 0.4, PaceFast: "4:30", PaceSlow: "4:40"}, {Intensity: "rest", Time: "1:30"}}},
		{Step: Step{Intensity: "cooldown", Time: "10:00"}},
	}
	got, ok := EstimateSeconds(blocks)
	// 300+900+90+3*105+12*(110+90)+600 = 4605
	if !ok || got != 4605 {
		t.Fatalf("estimate = %d ok=%v, want 4605", got, ok)
	}
}

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
