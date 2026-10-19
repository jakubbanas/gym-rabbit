package finalsurge

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Step is a single Workout Builder step. Exactly one of TimeSeconds,
// DistanceKm or LapButton should be set.
type Step struct {
	// Intensity: warmup, active (default), rest, cooldown.
	Intensity string `json:"intensity,omitempty" jsonschema:"one of: warmup, active, rest, cooldown (default active)"`
	// Duration as m:ss or h:mm:ss, e.g. "15:00", "0:15", "1:00:00".
	Time string `json:"time,omitempty" jsonschema:"step duration as m:ss or h:mm:ss, e.g. 15:00 or 0:15"`
	// Distance in kilometres, e.g. 0.4 for 400 m.
	DistanceKm float64 `json:"distance_km,omitempty" jsonschema:"step distance in km, e.g. 0.4 for 400 m"`
	// LapButton ends the step when the athlete presses lap.
	LapButton bool `json:"lap_button,omitempty" jsonschema:"step ends on lap button press instead of time/distance"`
	// Pace target in min/km, e.g. "5:25" .. "5:35". Leave empty for no target.
	PaceFast string `json:"pace_fast,omitempty" jsonschema:"faster end of the pace target in min/km, e.g. 5:25"`
	PaceSlow string `json:"pace_slow,omitempty" jsonschema:"slower end of the pace target in min/km, e.g. 5:35"`
	// Optional step title / comment shown on the device.
	Name string `json:"name,omitempty" jsonschema:"optional step title"`
}

// Block is either a single step or a repeat of several steps.
type Block struct {
	Step
	// Repeats > 0 turns the block into a repeat of Steps.
	Repeats int    `json:"repeats,omitempty" jsonschema:"if > 0 this block is a repeat of steps this many times"`
	Steps   []Step `json:"steps,omitempty" jsonschema:"steps inside the repeat"`
}

var (
	reTime = regexp.MustCompile(`^(\d{1,2}:)?\d{1,2}:\d{2}$`)
	rePace = regexp.MustCompile(`^\d{1,2}:\d{2}$`)
)

func intensityOf(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "active", "run", "work":
		return "ACTIVE", nil
	case "warmup", "warm_up", "warm up", "rozgrzewka":
		return "WARMUP", nil
	case "rest", "recovery", "marsz", "przerwa":
		return "REST", nil
	case "cooldown", "cool_down", "cool down", "schlodzenie", "schłodzenie":
		return "COOLDOWN", nil
	}
	return "", fmt.Errorf("unknown intensity %q (use warmup, active, rest, cooldown)", s)
}

// normTime turns "15:00" / "0:15" / "1:00:00" / "75:00" into the format the
// web app sends ("15:00", "0:15", "1:15:00").
func normTime(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !reTime.MatchString(s) {
		return "", fmt.Errorf("time %q must be m:ss or h:mm:ss", s)
	}
	parts := strings.Split(s, ":")
	total := 0
	for _, p := range parts {
		n, _ := strconv.Atoi(p)
		total = total*60 + n
	}
	if total <= 0 {
		return "", fmt.Errorf("time %q must be > 0", s)
	}
	h, m, sec := total/3600, (total%3600)/60, total%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec), nil
	}
	return fmt.Sprintf("%d:%02d", m, sec), nil
}

func paceSeconds(p string) int {
	parts := strings.Split(p, ":")
	m, _ := strconv.Atoi(parts[0])
	s, _ := strconv.Atoi(parts[1])
	return m*60 + s
}

// Wire format of /api/WorkoutBuilderSave. Field names, null vs "" and the
// two-element target array mirror the captured web-app payloads in
// builder_test.go; nil pointers and nil `any` fields encode as null.

type savePayload struct {
	TargetOptions  []targetOption `json:"target_options"`
	TargetOverride any            `json:"target_override"`
}

type targetOption struct {
	Name     string   `json:"name"`
	Sport    string   `json:"sport"`
	Steps    []any    `json:"steps"` // wireStep or wireRepeat
	Target   string   `json:"target"`
	Metadata metadata `json:"metadata"`
}

type metadata struct {
	Version string   `json:"version"`
	Editors []editor `json:"editors"`
}

type editor struct {
	Source   string `json:"source"`
	Version  string `json:"version"`
	Calendar string `json:"calendar"`
	Library  string `json:"library"`
	Plan     string `json:"plan"`
	Upload   string `json:"upload"`
}

type wireStep struct {
	Type           string    `json:"type"`
	ID             int       `json:"id"`
	Name           *string   `json:"name"`
	DurationType   string    `json:"durationType"`
	Duration       string    `json:"duration"`
	TargetAbsOrPct string    `json:"targetAbsOrPct"`
	DurationDist   float64   `json:"durationDist"`
	Data           []any     `json:"data"` // always empty for a step
	DistUnit       string    `json:"distUnit"`
	Target         [2]target `json:"target"`
	Intensity      string    `json:"intensity"`
	Comments       *string   `json:"comments"`
}

type wireRepeat struct {
	Type         string     `json:"type"`
	Name         *string    `json:"name"`
	ID           int        `json:"id"`
	Data         []wireStep `json:"data"`
	Repeats      int        `json:"repeats"`
	SkipLastRest bool       `json:"skip_last_rest"`
	Intensity    string     `json:"intensity"`
	DurationType string     `json:"durationType"`
	Comments     *string    `json:"comments"`
}

type target struct {
	TargetType           string  `json:"targetType"`
	ZoneBased            bool    `json:"zoneBased"`
	TargetLow            string  `json:"targetLow"`
	TargetHigh           string  `json:"targetHigh"`
	TargetOption         string  `json:"targetOption"`
	TargetTypeOriginal   *string `json:"targetTypeOriginal"`
	TargetOptionOriginal *string `json:"targetOptionOriginal"`
	TargetIsTimeBased    bool    `json:"targetIsTimeBased"`
	Zone                 int     `json:"zone"`
}

func openTarget(timeBased bool) target {
	return target{TargetType: "open", TargetLow: "0", TargetHigh: "0", TargetIsTimeBased: timeBased}
}

func paceTarget(fast, slow string) target {
	return target{TargetType: "pace", TargetLow: fast, TargetHigh: slow, TargetOption: "min/km", TargetIsTimeBased: true}
}

// nullIfEmpty maps "" to JSON null, as the web app does for unset names.
func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type idGen struct{ step, repeat int }

func (g *idGen) buildStep(s Step) (wireStep, error) {
	intensity, err := intensityOf(s.Intensity)
	if err != nil {
		return wireStep{}, err
	}
	set := 0
	if s.Time != "" {
		set++
	}
	if s.DistanceKm > 0 {
		set++
	}
	if s.LapButton {
		set++
	}
	if set != 1 {
		return wireStep{}, fmt.Errorf("each step needs exactly one of time, distance_km or lap_button (got %+v)", s)
	}

	durationType, duration, dist := "OPEN", "00:00", 0.0
	switch {
	case s.Time != "":
		durationType = "TIME"
		if duration, err = normTime(s.Time); err != nil {
			return wireStep{}, err
		}
	case s.DistanceKm > 0:
		durationType, dist = "DISTANCE", s.DistanceKm
	}

	first := openTarget(false)
	if s.PaceFast != "" || s.PaceSlow != "" {
		fast, slow := strings.TrimSpace(s.PaceFast), strings.TrimSpace(s.PaceSlow)
		if fast == "" {
			fast = slow
		}
		if slow == "" {
			slow = fast
		}
		if !rePace.MatchString(fast) || !rePace.MatchString(slow) {
			return wireStep{}, fmt.Errorf("pace must be m:ss per km, got %q-%q", fast, slow)
		}
		if paceSeconds(fast) > paceSeconds(slow) {
			fast, slow = slow, fast
		}
		first = paceTarget(fast, slow)
	}

	g.step++
	return wireStep{
		Type:         "step",
		ID:           g.step,
		Name:         nullIfEmpty(s.Name),
		DurationType: durationType,
		Duration:     duration,
		DurationDist: dist,
		Data:         []any{},
		DistUnit:     "km",
		Target:       [2]target{first, openTarget(true)},
		Intensity:    intensity,
	}, nil
}

// BuildStructure turns blocks into the JSON body of /api/WorkoutBuilderSave.
func BuildStructure(name string, blocks []Block, now time.Time) ([]byte, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("structure needs at least one block")
	}
	g := &idGen{step: 100, repeat: 5000}
	steps := make([]any, 0, len(blocks))
	for i, b := range blocks {
		if b.Repeats > 0 {
			if len(b.Steps) == 0 {
				return nil, fmt.Errorf("block %d: repeat needs steps", i+1)
			}
			inner := make([]wireStep, 0, len(b.Steps))
			for j, s := range b.Steps {
				st, err := g.buildStep(s)
				if err != nil {
					return nil, fmt.Errorf("block %d step %d: %w", i+1, j+1, err)
				}
				inner = append(inner, st)
			}
			g.repeat++
			steps = append(steps, wireRepeat{
				Type:         "repeat",
				Name:         nullIfEmpty(b.Name),
				ID:           g.repeat,
				Data:         inner,
				Repeats:      b.Repeats,
				Intensity:    "ACTIVE",
				DurationType: "OPEN",
			})
			continue
		}
		if len(b.Steps) > 0 {
			return nil, fmt.Errorf("block %d: has steps but repeats is 0", i+1)
		}
		st, err := g.buildStep(b.Step)
		if err != nil {
			return nil, fmt.Errorf("block %d: %w", i+1, err)
		}
		steps = append(steps, st)
	}

	return json.Marshal(savePayload{
		TargetOptions: []targetOption{{
			Name:   name,
			Sport:  "running",
			Steps:  steps,
			Target: "pace",
			Metadata: metadata{
				Version: "1",
				Editors: []editor{{
					Source:   "webapp",
					Version:  "0.0.1",
					Calendar: now.Format("2006-01-02 15:04"),
				}},
			},
		}},
	})
}

// EstimateSeconds returns the total time of all time-based steps plus, for
// distance steps with a pace target, distance × mid-pace. ok is false if some
// step could not be estimated.
func EstimateSeconds(blocks []Block) (total int, ok bool) {
	ok = true
	one := func(s Step) int {
		if s.Time != "" {
			t, err := normTime(s.Time)
			if err != nil {
				ok = false
				return 0
			}
			n := 0
			for _, p := range strings.Split(t, ":") {
				v, _ := strconv.Atoi(p)
				n = n*60 + v
			}
			return n
		}
		if s.DistanceKm > 0 && (s.PaceFast != "" || s.PaceSlow != "") {
			f, sl := s.PaceFast, s.PaceSlow
			if f == "" {
				f = sl
			}
			if sl == "" {
				sl = f
			}
			if rePace.MatchString(f) && rePace.MatchString(sl) {
				return int(s.DistanceKm * float64(paceSeconds(f)+paceSeconds(sl)) / 2)
			}
		}
		ok = false
		return 0
	}
	for _, b := range blocks {
		if b.Repeats > 0 {
			sub := 0
			for _, s := range b.Steps {
				sub += one(s)
			}
			total += sub * b.Repeats
		} else {
			total += one(b.Step)
		}
	}
	return total, ok
}
