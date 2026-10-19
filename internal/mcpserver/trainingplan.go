package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

const maxPlanSessions = 200

type planSession struct {
	Type string `json:"type" jsonschema:"run (goes to the FinalSurge calendar) or strength (saved in Gym Rabbit)"`
	Date string `json:"date" jsonschema:"day of the session, YYYY-MM-DD"`
	Name string `json:"name,omitempty" jsonschema:"session name; required for runs, defaults to the plan's 'Plan Name:' for strength"`

	// run fields
	Description     string             `json:"description,omitempty" jsonschema:"run only: free-text description"`
	PlannedMinutes  int                `json:"planned_minutes,omitempty" jsonschema:"run only: planned duration in minutes; estimated from structure if omitted"`
	PlannedDistance float64            `json:"planned_distance_km,omitempty" jsonschema:"run only: planned distance in km"`
	Structure       []finalsurge.Block `json:"structure,omitempty" jsonschema:"run only: Workout Builder structure (see tool description)"`

	// strength fields
	Text string `json:"text,omitempty" jsonschema:"strength only: the session in Gym Rabbit Magic Paste format (see gymrabbit://format)"`
}

type uploadPlanIn struct {
	Sessions []planSession `json:"sessions" jsonschema:"all sessions of the plan, in any order"`
	DryRun   bool          `json:"dry_run,omitempty" jsonschema:"validate everything and report what would be created, without writing anything"`
}

type sessionResult struct {
	Index int    `json:"index" jsonschema:"position of the session in the input list"`
	Type  string `json:"type"`
	Date  string `json:"date"`
	Name  string `json:"name"`
	// Status is "valid" (dry run), "created", or "failed".
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`

	WorkoutID string     `json:"workout_id,omitempty" jsonschema:"Gym Rabbit workout ID (strength)"`
	Run       *runResult `json:"run,omitempty" jsonschema:"FinalSurge result (run)"`
}

type uploadPlanOut struct {
	DryRun   bool            `json:"dry_run"`
	Created  int             `json:"created"`
	Failed   int             `json:"failed"`
	Sessions []sessionResult `json:"sessions"`
}

// preparedSession is one validated session, ready to write.
type preparedSession struct {
	result   sessionResult
	run      *preparedRun // run
	strength *db.Workout  // strength
}

func registerTrainingPlan(s *mcp.Server, database *db.DB, fs *finalsurge.Client) {
	runNote := ""
	if fs == nil {
		runNote = " NOTE: FinalSurge is not configured on this server (FINALSURGE_EMAIL/FINALSURGE_PASSWORD), so run sessions are rejected."
	}

	mcp.AddTool(s, &mcp.Tool{
		Name: "upload_training_plan",
		Description: "Upload a whole hybrid running + strength plan in one call. Each session has a type and a date: " +
			"runs are created in the FinalSurge calendar (with Workout Builder structure that syncs to the watch), " +
			"strength sessions are saved as Gym Rabbit workouts scheduled for that date.\n\n" +
			"The whole plan is validated first; if any session is invalid, nothing is written and every problem is reported. " +
			"Use dry_run to check a plan without writing. Writes are not transactional across the two systems: if a write fails " +
			"midway, the result lists exactly which sessions were created and which failed, so only the failed ones need retrying." +
			runNote + "\n\nRun " + structureHelp + "\n\nStrength text format: read the " + formatURI + " resource.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in uploadPlanIn) (*mcp.CallToolResult, uploadPlanOut, error) {
		prepared, err := prepareSessions(in.Sessions, fs != nil)
		if err != nil {
			return nil, uploadPlanOut{}, err
		}

		out := uploadPlanOut{DryRun: in.DryRun, Sessions: make([]sessionResult, 0, len(prepared))}
		runsPerDay := map[string]int{}
		for _, p := range prepared {
			r := p.result
			if in.DryRun {
				r.Status = "valid"
				out.Sessions = append(out.Sessions, r)
				continue
			}
			var werr error
			switch {
			case p.run != nil:
				runsPerDay[r.Date]++
				p.run.workout.Order = runsPerDay[r.Date]
				var rr runResult
				if rr, werr = p.run.create(ctx, fs); werr == nil {
					r.Run = &rr
				}
			case p.strength != nil:
				if werr = database.CreateWorkout(p.strength); werr == nil {
					r.WorkoutID = p.strength.ID
				}
			}
			if werr != nil {
				r.Status, r.Error = "failed", werr.Error()
				out.Failed++
			} else {
				r.Status = "created"
				out.Created++
			}
			out.Sessions = append(out.Sessions, r)
		}
		return nil, out, nil
	})
}

// prepareSessions validates every session and returns them in input
// order. It reports all problems at once so the agent can fix them in one go.
func prepareSessions(sessions []planSession, runsEnabled bool) ([]preparedSession, error) {
	if len(sessions) == 0 {
		return nil, errors.New("sessions is empty")
	}
	if len(sessions) > maxPlanSessions {
		return nil, fmt.Errorf("too many sessions (%d); upload at most %d per call", len(sessions), maxPlanSessions)
	}

	var problems []string
	out := make([]preparedSession, 0, len(sessions))
	for i, s := range sessions {
		p, err := prepareSession(i, s, runsEnabled)
		if err != nil {
			problems = append(problems, fmt.Sprintf("session %d (%s %s %q): %v", i, s.Type, s.Date, s.Name, err))
			continue
		}
		out = append(out, *p)
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("plan not uploaded, nothing was written. Fix these sessions:\n- %s", strings.Join(problems, "\n- "))
	}
	return out, nil
}

func prepareSession(i int, s planSession, runsEnabled bool) (*preparedSession, error) {
	if _, err := time.Parse("2006-01-02", s.Date); err != nil {
		return nil, fmt.Errorf("date must be YYYY-MM-DD")
	}
	p := &preparedSession{result: sessionResult{Index: i, Type: s.Type, Date: s.Date, Name: s.Name}}

	switch strings.ToLower(strings.TrimSpace(s.Type)) {
	case "run":
		if !runsEnabled {
			return nil, errors.New("FinalSurge is not configured on this server")
		}
		if s.Text != "" {
			return nil, errors.New("text is for strength sessions; runs use structure")
		}
		run, err := prepareRun(runIn{
			Date: s.Date, Name: s.Name, Description: s.Description,
			PlannedMinutes: s.PlannedMinutes, PlannedDistance: s.PlannedDistance, Structure: s.Structure,
		})
		if err != nil {
			return nil, err
		}
		p.result.Type, p.run = "run", run

	case "strength":
		if len(s.Structure) > 0 {
			return nil, errors.New("structure is for runs; strength sessions use text")
		}
		if s.Text == "" {
			return nil, errors.New("text is required for strength sessions")
		}
		parsed, err := plan.Parse(s.Text)
		if err != nil {
			return nil, fmt.Errorf("invalid strength text: %w", err)
		}
		title := s.Name
		if title == "" {
			title = parsed.Name
		}
		p.result.Type, p.result.Name = "strength", title
		p.strength = &db.Workout{ID: db.NewID(), Title: title, Text: s.Text, ScheduledDate: s.Date}

	default:
		return nil, fmt.Errorf("type must be run or strength, got %q", s.Type)
	}
	return p, nil
}
