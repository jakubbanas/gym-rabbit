package finalsurge

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeFS emulates the handful of FinalSurge endpoints the client uses.
func fakeFS(t *testing.T, logins *int32, expireFirstToken bool) *httptest.Server {
	var calls int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := func(data any, extra map[string]any) {
			m := map[string]any{"success": true, "data": data}
			for k, v := range extra {
				m[k] = v
			}
			_ = json.NewEncoder(w).Encode(m)
		}
		auth := r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/api/login":
			var b map[string]string
			_ = json.NewDecoder(r.Body).Decode(&b)
			if b["email"] != "a@b.c" || b["password"] != "pw" {
				_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error_description": "Invalid login"})
				return
			}
			n := atomic.AddInt32(logins, 1)
			ok(map[string]any{"token": "tok" + string(rune('0'+n))}, nil)
			return
		}
		if !strings.HasPrefix(auth, "Bearer tok") {
			w.WriteHeader(401)
			return
		}
		if expireFirstToken && auth == "Bearer tok1" && atomic.AddInt32(&calls, 1) > 1 {
			w.WriteHeader(401)
			return
		}
		switch r.URL.Path {
		case "/api/Settings":
			ok(map[string]any{"user_key": "U1"}, nil)
		case "/api/WorkoutList":
			if r.URL.Query().Get("scopekey") != "U1" {
				t.Errorf("scopekey = %q", r.URL.Query().Get("scopekey"))
			}
			ok([]map[string]any{{
				"key": "W1", "workout_date": "2026-10-13T00:00:00", "name": "Easy",
				"workout_status_text": "Planned", "has_structured_workout": true,
				"Activities": []map[string]any{{"activity_type_name": "Run", "planned_duration": 3600.0}},
			}}, nil)
		case "/api/WorkoutSave":
			body, _ := io.ReadAll(r.Body)
			var b map[string]any
			_ = json.Unmarshal(body, &b)
			if b["workout_date"] != "2026-10-20T00:00:00" {
				t.Errorf("workout_date = %v", b["workout_date"])
			}
			act := b["Activity"].(map[string]any)
			if act["planned_duration"].(float64) != 1800 || act["activity_type_key"] != RunActivityTypeKey {
				t.Errorf("activity = %v", act)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "new_workout_key": "NEW"})
		case "/api/WorkoutBuilderSave":
			if r.URL.Query().Get("workout_key") != "NEW" {
				t.Errorf("workout_key = %q", r.URL.Query().Get("workout_key"))
			}
			ok(nil, nil)
		default:
			w.WriteHeader(404)
		}
	}))
}

func TestClientFlow(t *testing.T) {
	var logins int32
	srv := fakeFS(t, &logins, false)
	defer srv.Close()
	c := New("a@b.c", "pw")
	c.BaseURL = srv.URL + "/api/"
	ctx := context.Background()

	ws, err := c.ListWorkouts(ctx, "2026-10-12", "2026-10-18")
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 || ws[0].Key != "W1" || ws[0].Date != "2026-10-13" || ws[0].PlannedSeconds != 3600 || !ws[0].HasStructure {
		t.Fatalf("workouts = %+v", ws)
	}
	key, err := c.CreateRunWorkout(ctx, NewRunWorkout{Date: "2026-10-20", Name: "x", PlannedSeconds: 1800})
	if err != nil || key != "NEW" {
		t.Fatalf("create: %q %v", key, err)
	}
	payload, _ := BuildStructure("x", []Block{{Step: Step{Time: "30:00"}}}, testNow)
	if err := c.SaveStructure(ctx, key, payload); err != nil {
		t.Fatal(err)
	}
	if logins != 1 {
		t.Fatalf("logins = %d, want 1", logins)
	}
}

func TestRelogin(t *testing.T) {
	var logins int32
	srv := fakeFS(t, &logins, true)
	defer srv.Close()
	c := New("a@b.c", "pw")
	c.BaseURL = srv.URL + "/api/"
	ctx := context.Background()
	if _, err := c.ListWorkouts(ctx, "2026-10-12", "2026-10-18"); err != nil {
		t.Fatal(err)
	}
	if logins != 2 {
		t.Fatalf("expected re-login after expired token, logins = %d", logins)
	}
}

func TestBadPassword(t *testing.T) {
	var logins int32
	srv := fakeFS(t, &logins, false)
	defer srv.Close()
	c := New("a@b.c", "wrong")
	c.BaseURL = srv.URL + "/api/"
	if _, err := c.ListWorkouts(context.Background(), "2026-10-12", "2026-10-18"); err == nil {
		t.Fatal("expected login error")
	}
}

var testNow = mustTime("2026-10-12T14:00:00Z")
