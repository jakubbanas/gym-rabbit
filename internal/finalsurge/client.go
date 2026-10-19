// Package finalsurge (copied from ../finalsurge-mcp) is a small client for the (unofficial, undocumented)
// FinalSurge web API used by https://beta.finalsurge.com.
//
// Everything here was reverse-engineered from the web app's network traffic
// and may break whenever FinalSurge changes their frontend.
package finalsurge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const DefaultBaseURL = "https://beta.finalsurge.com/api/"

// RunActivityTypeKey is the activity_type_key for "Run".
const RunActivityTypeKey = "00000001-0001-0001-0001-000000000001"

// Client talks to the FinalSurge API. It logs in lazily with email/password
// and re-logs in once when the token stops working.
type Client struct {
	BaseURL  string
	Email    string
	Password string
	HTTP     *http.Client

	mu      sync.Mutex
	token   string
	userKey string
}

func New(email, password string) *Client {
	return &Client{
		BaseURL:  DefaultBaseURL,
		Email:    email,
		Password: password,
		HTTP:     &http.Client{Timeout: 30 * time.Second},
	}
}

// envelope is the common response wrapper of every FinalSurge endpoint.
type envelope struct {
	Success          bool            `json:"success"`
	Data             json.RawMessage `json:"data"`
	ErrorNumber      *int            `json:"error_number"`
	ErrorDescription *string         `json:"error_description"`
	NewWorkoutKey    string          `json:"new_workout_key"`
}

type APIError struct {
	Endpoint    string
	Status      int
	Number      *int
	Description string
}

func (e *APIError) Error() string {
	n := ""
	if e.Number != nil {
		n = fmt.Sprintf(" #%d", *e.Number)
	}
	return fmt.Sprintf("finalsurge %s: HTTP %d, error%s: %s", e.Endpoint, e.Status, n, e.Description)
}

var errUnauthorized = errors.New("unauthorized")

// Login exchanges email/password for a session token and fetches user_key.
func (c *Client) Login(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loginLocked(ctx)
}

func (c *Client) loginLocked(ctx context.Context) error {
	if c.Email == "" || c.Password == "" {
		return errors.New("FINALSURGE_EMAIL and FINALSURGE_PASSWORD must be set")
	}
	body := map[string]string{
		"email":                  c.Email,
		"password":               c.Password,
		"deviceManufacturer":     "",
		"deviceModel":            "finalsurge-mcp",
		"deviceOperatingSystem":  "go",
		"deviceUniqueIdentifier": "",
	}
	env, err := c.do(ctx, http.MethodPost, "login", nil, body, "")
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil || data.Token == "" {
		return errors.New("login failed: no token in response (wrong email/password?)")
	}
	c.token = data.Token

	env, err = c.do(ctx, http.MethodGet, "Settings", nil, nil, c.token)
	if err != nil {
		return fmt.Errorf("reading settings after login: %w", err)
	}
	var settings struct {
		UserKey string `json:"user_key"`
	}
	if err := json.Unmarshal(env.Data, &settings); err != nil || settings.UserKey == "" {
		return errors.New("could not read user_key from /api/Settings")
	}
	c.userKey = settings.UserKey
	return nil
}

// UserKey returns the logged-in user's key (the "scopekey" in most calls).
func (c *Client) UserKey(ctx context.Context) (string, error) {
	if err := c.ensureLogin(ctx); err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.userKey, nil
}

func (c *Client) ensureLogin(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" {
		return nil
	}
	return c.loginLocked(ctx)
}

// call performs an authenticated request, re-logging in once on 401.
func (c *Client) call(ctx context.Context, method, endpoint string, query url.Values, body any) (*envelope, error) {
	if err := c.ensureLogin(ctx); err != nil {
		return nil, err
	}
	c.mu.Lock()
	tok := c.token
	c.mu.Unlock()

	env, err := c.do(ctx, method, endpoint, query, body, tok)
	if errors.Is(err, errUnauthorized) {
		c.mu.Lock()
		c.token = ""
		lerr := c.loginLocked(ctx)
		tok = c.token
		c.mu.Unlock()
		if lerr != nil {
			return nil, lerr
		}
		env, err = c.do(ctx, method, endpoint, query, body, tok)
	}
	return env, err
}

func (c *Client) do(ctx context.Context, method, endpoint string, query url.Values, body any, token string) (*envelope, error) {
	u := strings.TrimRight(c.BaseURL, "/") + "/" + endpoint
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var rdr io.Reader
	if body != nil {
		var buf []byte
		switch b := body.(type) {
		case []byte:
			buf = b
		default:
			var err error
			if buf, err = json.Marshal(body); err != nil {
				return nil, err
			}
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("Authorization", "")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%s: %w", endpoint, errUnauthorized)
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, &APIError{Endpoint: endpoint, Status: resp.StatusCode, Description: truncate(string(raw), 300)}
	}
	if resp.StatusCode >= 300 || !env.Success {
		desc := ""
		if env.ErrorDescription != nil {
			desc = *env.ErrorDescription
		}
		if isAuthError(desc) {
			return nil, fmt.Errorf("%s: %w (%s)", endpoint, errUnauthorized, desc)
		}
		return nil, &APIError{Endpoint: endpoint, Status: resp.StatusCode, Number: env.ErrorNumber, Description: desc}
	}
	return &env, nil
}

func isAuthError(desc string) bool {
	d := strings.ToLower(desc)
	return strings.Contains(d, "token") || strings.Contains(d, "not authorized") || strings.Contains(d, "unauthorized") || strings.Contains(d, "session")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ---------------------------------------------------------------------------
// Workouts

// Workout is a trimmed-down view of a calendar workout.
type Workout struct {
	Key            string  `json:"key"`
	Date           string  `json:"date"`
	Name           string  `json:"name,omitempty"`
	Description    string  `json:"description,omitempty"`
	Status         string  `json:"status,omitempty"`
	Activity       string  `json:"activity,omitempty"`
	PlannedSeconds float64 `json:"planned_seconds,omitempty"`
	PlannedAmount  float64 `json:"planned_amount,omitempty"`
	PlannedUnit    string  `json:"planned_unit,omitempty"`
	ActualSeconds  float64 `json:"actual_seconds,omitempty"`
	ActualAmount   float64 `json:"actual_amount,omitempty"`
	ActualUnit     string  `json:"actual_unit,omitempty"`
	HasStructure   bool    `json:"has_structured_workout"`
}

type rawWorkout struct {
	Key                  string  `json:"key"`
	WorkoutDate          string  `json:"workout_date"`
	Name                 *string `json:"name"`
	Description          *string `json:"description"`
	WorkoutStatusText    string  `json:"workout_status_text"`
	HasStructuredWorkout bool    `json:"has_structured_workout"`
	Activities           []struct {
		ActivityTypeName    string   `json:"activity_type_name"`
		ActivitySubTypeName *string  `json:"activity_sub_type_name"`
		PlannedDuration     *float64 `json:"planned_duration"`
		PlannedAmount       *float64 `json:"planned_amount"`
		PlannedAmountType   *string  `json:"planned_amount_type"`
		Duration            *float64 `json:"duration"`
		Amount              *float64 `json:"amount"`
		AmountType          *string  `json:"amount_type"`
	} `json:"Activities"`
}

func (r rawWorkout) toWorkout() Workout {
	w := Workout{
		Key:          r.Key,
		Date:         strings.SplitN(r.WorkoutDate, "T", 2)[0],
		Status:       r.WorkoutStatusText,
		HasStructure: r.HasStructuredWorkout,
	}
	if r.Name != nil {
		w.Name = *r.Name
	}
	if r.Description != nil {
		w.Description = *r.Description
	}
	if len(r.Activities) > 0 {
		a := r.Activities[0]
		w.Activity = a.ActivityTypeName
		if a.ActivitySubTypeName != nil && *a.ActivitySubTypeName != "" {
			w.Activity += " - " + *a.ActivitySubTypeName
		}
		w.PlannedSeconds = deref(a.PlannedDuration)
		w.PlannedAmount = deref(a.PlannedAmount)
		w.PlannedUnit = derefS(a.PlannedAmountType)
		w.ActualSeconds = deref(a.Duration)
		w.ActualAmount = deref(a.Amount)
		if w.ActualAmount != 0 {
			w.ActualUnit = derefS(a.AmountType)
		}
		if w.PlannedAmount == 0 {
			w.PlannedUnit = ""
		}
	}
	return w
}

func deref(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
func derefS(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ListWorkouts returns planned and completed workouts between start and end
// (inclusive, YYYY-MM-DD).
func (c *Client) ListWorkouts(ctx context.Context, start, end string) ([]Workout, error) {
	uk, err := c.UserKey(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{
		"scope": {"USER"}, "scopekey": {uk},
		"startdate": {start}, "enddate": {end},
		"ishistory": {"false"}, "completedonly": {"false"},
	}
	env, err := c.call(ctx, http.MethodGet, "WorkoutList", q, nil)
	if err != nil {
		return nil, err
	}
	var raws []rawWorkout
	if err := json.Unmarshal(env.Data, &raws); err != nil {
		return nil, fmt.Errorf("decoding WorkoutList: %w", err)
	}
	out := make([]Workout, 0, len(raws))
	for _, r := range raws {
		out = append(out, r.toWorkout())
	}
	return out, nil
}

// NewRunWorkout describes a planned run to create on the calendar.
type NewRunWorkout struct {
	Date            string // YYYY-MM-DD
	Name            string
	Description     string
	PlannedSeconds  int     // 0 = not set
	PlannedDistance float64 // in km; 0 = not set
	Order           int     // position within the day, 1-based; 0 = 1
}

// CreateRunWorkout creates a planned run and returns its workout_key.
func (c *Client) CreateRunWorkout(ctx context.Context, w NewRunWorkout) (string, error) {
	uk, err := c.UserKey(ctx)
	if err != nil {
		return "", err
	}
	if _, err := time.Parse("2006-01-02", w.Date); err != nil {
		return "", fmt.Errorf("date must be YYYY-MM-DD: %w", err)
	}
	order := w.Order
	if order < 1 {
		order = 1
	}
	var dur any
	if w.PlannedSeconds > 0 {
		dur = w.PlannedSeconds
	}
	var amt any
	if w.PlannedDistance > 0 {
		amt = w.PlannedDistance
	}
	body := map[string]any{
		"key":             nil,
		"workout_date":    w.Date + "T00:00:00",
		"order":           order,
		"name":            w.Name,
		"description":     w.Description,
		"is_race":         false,
		"has_routes":      false,
		"has_attachments": false,
		"Activity": map[string]any{
			"elevation_gain_type":    "me",
			"elevation_gain":         nil,
			"elevation_loss_type":    "me",
			"elevation_loss":         nil,
			"activity_type_key":      RunActivityTypeKey,
			"activity_type_name":     "Run",
			"activity_sub_type_key":  "",
			"activity_sub_type_name": "",
			"planned_duration":       dur,
			"planned_amount":         amt,
			"planned_amount_type":    "km",
			"duration":               nil,
			"amount":                 nil,
			"amount_type":            "km",
			"pace":                   nil,
			"pace_type":              "km",
			"hr_avg":                 nil,
			"hr_max":                 nil,
			"power_avg":              nil,
			"power_max":              nil,
			"cadence_avg":            nil,
			"cadence_max":            nil,
			"calories":               nil,
		},
		"felt":                    nil,
		"effort":                  nil,
		"post_workout_notes":      nil,
		"save_to_library":         false,
		"workout_time":            "",
		"race_place_overall":      nil,
		"race_age_group":          nil,
		"location_save_for_later": false,
		"location_name":           nil,
		"location_street":         nil,
		"location_city":           nil,
		"location_state":          nil,
		"location_postalcode":     nil,
		"location_country":        nil,
	}
	env, err := c.call(ctx, http.MethodPost, "WorkoutSave", url.Values{"scope": {"USER"}, "scope_key": {uk}}, body)
	if err != nil {
		return "", err
	}
	if env.NewWorkoutKey == "" {
		return "", errors.New("WorkoutSave succeeded but returned no new_workout_key")
	}
	return env.NewWorkoutKey, nil
}

// SaveStructure stores a Workout Builder structure on an existing workout.
func (c *Client) SaveStructure(ctx context.Context, workoutKey string, payload []byte) error {
	uk, err := c.UserKey(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"scope": {"USER"}, "scopekey": {uk}, "workout_key": {workoutKey}}
	_, err = c.call(ctx, http.MethodPost, "WorkoutBuilderSave", q, payload)
	return err
}

// GetStructure returns the raw Workout Builder JSON of a workout.
func (c *Client) GetStructure(ctx context.Context, workoutKey string) (json.RawMessage, error) {
	uk, err := c.UserKey(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"scope": {"USER"}, "scopekey": {uk}, "workout_key": {workoutKey}, "array": {"true"}, "newobject": {"true"}}
	env, err := c.call(ctx, http.MethodGet, "WorkoutBuilderGet", q, nil)
	if err != nil {
		return nil, err
	}
	return env.Data, nil
}

// DeleteWorkout permanently removes a workout from the calendar.
func (c *Client) DeleteWorkout(ctx context.Context, workoutKey string) error {
	uk, err := c.UserKey(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"scope": {"USER"}, "scopekey": {uk}, "workout_key": {workoutKey}}
	_, err = c.call(ctx, http.MethodGet, "WorkoutDelete", q, nil)
	return err
}
