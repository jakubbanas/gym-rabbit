package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jakubbanas/gym-rabbit/internal/db"
	"github.com/jakubbanas/gym-rabbit/internal/finalsurge"
	"github.com/jakubbanas/gym-rabbit/internal/plan"
)

const maxPlanSessions = 200

type planSession struct {
	Type  string `json:"type" jsonschema:"run (goes to the FinalSurge calendar) or strength (saved in Gym Rabbit)"`
	Date  string `json:"date" jsonschema:"day of the session, YYYY-MM-DD"`
	Name  string `json:"name" jsonschema:"session name, e.g. Easy 40' or Lower A (shown on the watch for runs)"`
	Notes string `json:"notes,omitempty" jsonschema:"optional note for the athlete (for runs it becomes the FinalSurge description)"`

	// run fields
	PlannedMinutes  int                `json:"planned_minutes,omitempty" jsonschema:"run only: planned duration in minutes; estimated from structure if omitted"`
	PlannedDistance float64            `json:"planned_distance_km,omitempty" jsonschema:"run only: planned distance in km"`
	Structure       []finalsurge.Block `json:"structure,omitempty" jsonschema:"run only: Workout Builder structure (see tool description)"`

	// strength fields
	Blocks []plan.Block `json:"blocks,omitempty" jsonschema:"strength only: the exercises, grouped into blocks (see tool description)"`
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
		Description: "Add sessions to the training plan: a whole running + strength block in one call, or a single " +
			"session. Each session has a type and a date: " +
			"runs are created in the FinalSurge calendar (with Workout Builder structure that syncs to the watch), " +
			"strength sessions are saved as Gym Rabbit workouts scheduled for that date.\n\n" +
			"The whole plan is validated first; if any session is invalid, nothing is written and every problem is reported. " +
			"Use dry_run to check a plan without writing. Writes are not transactional across the two systems: if a write fails " +
			"midway, the result lists exactly which sessions were created and which failed, so only the failed ones need retrying." +
			runNote + "\n\nRun " + structureHelp + "\n\nStrength: " + strengthHelp,
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
			// Indent nested field errors under their session.
			msg := strings.ReplaceAll(err.Error(), "\n- ", "\n    - ")
			problems = append(problems, fmt.Sprintf("session %d (%s %s %q): %s", i, s.Type, s.Date, s.Name, msg))
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
	if _, err := parseDate("date", s.Date); err != nil {
		return nil, err
	}
	p := &preparedSession{result: sessionResult{Index: i, Type: s.Type, Date: s.Date, Name: s.Name}}

	switch strings.ToLower(strings.TrimSpace(s.Type)) {
	case "run":
		if !runsEnabled {
			return nil, errors.New("FinalSurge is not configured on this server")
		}
		if len(s.Blocks) > 0 {
			return nil, errors.New("blocks are for strength sessions; runs use structure")
		}
		run, err := prepareRun(s)
		if err != nil {
			return nil, err
		}
		p.result.Type, p.run = "run", run

	case "strength":
		if len(s.Structure) > 0 {
			return nil, errors.New("structure is for runs; strength sessions use blocks")
		}
		w, err := newWorkout(s.Name, s.Notes, s.Date, s.Blocks)
		if err != nil {
			return nil, err
		}
		p.result.Type, p.result.Name, p.strength = "strength", w.Title, w

	default:
		return nil, fmt.Errorf("type must be run or strength, got %q", s.Type)
	}
	return p, nil
}
