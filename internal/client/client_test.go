package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}
}

func testClientWithKey(t *testing.T, handler http.Handler, apiKey string) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{
		BaseURL:    server.URL,
		apiKey:     apiKey,
		HTTPClient: server.Client(),
	}
}

func TestHealth_Success(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/group" {
			t.Errorf("expected /group, got %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode([]string{"default"})
	}))

	if err := c.Health(context.Background()); err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

func TestHealth_Failure(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))

	err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error from health check")
	}
}

func TestAuthorizationHeader(t *testing.T) {
	c := testClientWithKey(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected 'Bearer test-key', got %q", auth)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode([]string{})
	}), "test-key")

	c.Health(context.Background())
}

func TestNoAuthorizationHeaderWhenEmpty(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("expected no Authorization header, got %q", auth)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode([]string{})
	}))

	c.Health(context.Background())
}

func TestGetGroups(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/group" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode([]string{"default", "admin"})
	}))

	groups, err := c.GetGroups(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 || groups[0] != "default" || groups[1] != "admin" {
		t.Errorf("unexpected groups: %v", groups)
	}
}

func TestGetGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/group/admin" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		dn := "Администрация"
		json.NewEncoder(w).Encode(Group{
			Name:        "admin",
			DisplayName: &dn,
			Weight:      500,
			Nodes: []Node{
				{Key: "*", Value: true, Type: "permission"},
			},
		})
	}))

	group, err := c.GetGroup(context.Background(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.Name != "admin" {
		t.Errorf("expected admin, got %s", group.Name)
	}
	if group.DisplayName == nil || *group.DisplayName != "Администрация" {
		t.Error("unexpected display name")
	}
	if group.Weight != 500 {
		t.Errorf("expected weight 500, got %d", group.Weight)
	}
}

func TestGetGroup_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not found"))
	}))

	_, err := c.GetGroup(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound, got %v", err)
	}
}

func TestCreateGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var req createGroupRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "newgroup" {
			t.Errorf("expected newgroup, got %s", req.Name)
		}
		json.NewEncoder(w).Encode(Group{Name: "newgroup"})
	}))

	group, err := c.CreateGroup(context.Background(), "newgroup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.Name != "newgroup" {
		t.Errorf("expected newgroup, got %s", group.Name)
	}
}

func TestCreateGroup_Conflict(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(409)
		w.Write([]byte("already exists"))
	}))

	_, err := c.CreateGroup(context.Background(), "default")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsConflict(err) {
		t.Errorf("expected IsConflict, got %v", err)
	}
}

func TestSetGroupNodes_StripsType(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nodes []Node
		json.NewDecoder(r.Body).Decode(&nodes)
		for _, n := range nodes {
			if n.Type != "" {
				t.Errorf("expected type stripped, got %q for key %s", n.Type, n.Key)
			}
		}
		w.WriteHeader(200)
		w.Write([]byte("[]"))
	}))

	nodes := []Node{
		{Key: "perm.a", Value: true, Type: "permission"},
		{Key: "group.admin", Value: true, Type: "inheritance"},
	}

	if err := c.SetGroupNodes(context.Background(), "test", nodes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetry5xx_GET(t *testing.T) {
	attempts := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(500)
			w.Write([]byte("server error"))
			return
		}
		json.NewEncoder(w).Encode([]string{"default"})
	}))

	_, err := c.doRequest(context.Background(), "GET", "/group", nil)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry5xx_Exhausted(t *testing.T) {
	attempts := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(500)
		w.Write([]byte("persistent error"))
	}))

	_, err := c.doRequest(context.Background(), "GET", "/group", nil)
	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
	if attempts != 4 {
		t.Errorf("expected 4 attempts (1 initial + 3 retries), got %d", attempts)
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestNoRetry_POST(t *testing.T) {
	attempts := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}))

	_, err := c.doRequest(context.Background(), "POST", "/group", map[string]string{"name": "test"})
	if err == nil {
		t.Fatal("expected error from POST 500")
	}
	if attempts != 1 {
		t.Errorf("POST should not retry, expected 1 attempt, got %d", attempts)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		err    error
		expect bool
	}{
		{&APIError{StatusCode: 404}, true},
		{&APIError{StatusCode: 200}, false},
		{nil, false},
	}

	for _, tt := range tests {
		if got := IsNotFound(tt.err); got != tt.expect {
			t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.expect)
		}
	}
}

func TestIsConflict(t *testing.T) {
	tests := []struct {
		err    error
		expect bool
	}{
		{&APIError{StatusCode: 409}, true},
		{&APIError{StatusCode: 200}, false},
		{nil, false},
	}

	for _, tt := range tests {
		if got := IsConflict(tt.err); got != tt.expect {
			t.Errorf("IsConflict(%v) = %v, want %v", tt.err, got, tt.expect)
		}
	}
}

func TestCreateTrack(t *testing.T) {
	calls := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch {
		case calls == 1 && r.Method == "POST":
			json.NewEncoder(w).Encode(Track{Name: "admin", Groups: nil})
		case calls == 2 && r.Method == "PATCH":
			w.Write([]byte("ok"))
		case calls == 3 && r.Method == "GET":
			json.NewEncoder(w).Encode(Track{Name: "admin", Groups: []string{"helper", "admin"}})
		}
	}))

	track, err := c.CreateTrack(context.Background(), "admin", []string{"helper", "admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.Name != "admin" || len(track.Groups) != 2 {
		t.Errorf("unexpected track: %+v", track)
	}
}

func TestNew(t *testing.T) {
	c := New("http://localhost:8080", "key", 10*time.Second, false)
	if c.BaseURL != "http://localhost:8080" {
		t.Error("wrong base url")
	}
	if c.apiKey != "key" {
		t.Error("wrong api key")
	}
	if c.HTTPClient.Timeout != 10*time.Second {
		t.Error("wrong timeout")
	}
}

func TestNew_TrimsTrailingSlash(t *testing.T) {
	c := New("http://localhost:8080/", "key", 10*time.Second, false)
	if c.BaseURL != "http://localhost:8080" {
		t.Errorf("expected trailing slash trimmed, got %q", c.BaseURL)
	}
}

func TestDeleteGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/group/testgroup" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(200)
	}))

	if err := c.DeleteGroup(context.Background(), "testgroup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteGroup_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not found"))
	}))

	err := c.DeleteGroup(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound, got %v", err)
	}
}

func TestGetGroupNodes(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/group/admin/nodes" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Node{
			{Key: "perm.a", Value: true},
			{Key: "perm.b", Value: false},
		})
	}))

	nodes, err := c.GetGroupNodes(context.Background(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Key != "perm.a" || !nodes[0].Value {
		t.Errorf("unexpected node 0: %+v", nodes[0])
	}
	if nodes[1].Key != "perm.b" || nodes[1].Value {
		t.Errorf("unexpected node 1: %+v", nodes[1])
	}
}

func TestGetTracks(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/track" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode([]string{"staff", "vip"})
	}))

	tracks, err := c.GetTracks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 2 || tracks[0] != "staff" || tracks[1] != "vip" {
		t.Errorf("unexpected tracks: %v", tracks)
	}
}

func TestGetTrack(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/track/staff" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(Track{Name: "staff", Groups: []string{"helper", "mod", "admin"}})
	}))

	track, err := c.GetTrack(context.Background(), "staff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.Name != "staff" {
		t.Errorf("expected name staff, got %s", track.Name)
	}
	if len(track.Groups) != 3 || track.Groups[0] != "helper" {
		t.Errorf("unexpected groups: %v", track.Groups)
	}
}

func TestGetTrack_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not found"))
	}))

	_, err := c.GetTrack(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound, got %v", err)
	}
}

func TestUpdateTrack(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/track/staff" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req updateTrackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if len(req.Groups) != 2 || req.Groups[0] != "helper" || req.Groups[1] != "mod" {
			t.Errorf("unexpected groups in body: %v", req.Groups)
		}
		w.WriteHeader(200)
	}))

	if err := c.UpdateTrack(context.Background(), "staff", []string{"helper", "mod"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTrack(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/track/staff" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(200)
	}))

	if err := c.DeleteTrack(context.Background(), "staff"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTrack_PatchFails_CleanupSucceeds(t *testing.T) {
	// Use 400 on PATCH to avoid the retry loop (only 5xx triggers retries).
	// Sequence: POST 200 → PATCH 400 → DELETE 200 → error contains "cleaned up".
	var methods []string
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		switch r.Method {
		case "POST":
			json.NewEncoder(w).Encode(Track{Name: "staff"})
		case "PATCH":
			w.WriteHeader(400)
			w.Write([]byte("bad request"))
		case "DELETE":
			w.WriteHeader(200)
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(500)
		}
	}))

	_, err := c.CreateTrack(context.Background(), "staff", []string{"helper"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cleaned up") {
		t.Errorf("expected 'cleaned up' in error, got: %v", err)
	}
	if strings.Contains(err.Error(), "cleanup also failed") {
		t.Errorf("unexpected 'cleanup also failed' in error: %v", err)
	}
}

func TestCreateTrack_PatchFails_CleanupFails(t *testing.T) {
	// Use 400 on both PATCH and DELETE to avoid retry loops.
	// Sequence: POST 200 → PATCH 400 → DELETE 400 → error contains "cleanup also failed".
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			json.NewEncoder(w).Encode(Track{Name: "staff"})
		case "PATCH":
			w.WriteHeader(400)
			w.Write([]byte("patch error"))
		case "DELETE":
			w.WriteHeader(400)
			w.Write([]byte("delete error"))
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(500)
		}
	}))

	_, err := c.CreateTrack(context.Background(), "staff", []string{"helper"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cleanup also failed") {
		t.Errorf("expected 'cleanup also failed' in error, got: %v", err)
	}
}

func TestLockGroup_SerializesSameGroup(t *testing.T) {
	c := &Client{}
	unlock1 := c.LockGroup("admin")

	locked := make(chan bool, 1)
	go func() {
		unlock2 := c.LockGroup("admin")
		locked <- true
		unlock2()
	}()

	// Goroutine should be blocked
	select {
	case <-locked:
		t.Fatal("second lock should be blocked while first is held")
	case <-time.After(50 * time.Millisecond):
		// expected
	}

	unlock1()

	// Now second lock should acquire
	select {
	case <-locked:
		// expected
	case <-time.After(time.Second):
		t.Fatal("second lock should have acquired after first was released")
	}
}

func TestLockGroup_IndependentGroups(t *testing.T) {
	c := &Client{}
	unlock1 := c.LockGroup("admin")

	locked := make(chan bool, 1)
	go func() {
		unlock2 := c.LockGroup("player") // different group
		locked <- true
		unlock2()
	}()

	select {
	case <-locked:
		// expected — different groups don't block each other
	case <-time.After(time.Second):
		t.Fatal("different groups should not block each other")
	}

	unlock1()
}

func TestAPIError_Truncation(t *testing.T) {
	short := &APIError{StatusCode: 400, Body: "short", Method: "GET", Path: "/test"}
	if !strings.Contains(short.Error(), "short") {
		t.Error("short body should appear in full")
	}

	long := &APIError{StatusCode: 500, Body: strings.Repeat("x", 300), Method: "GET", Path: "/test"}
	msg := long.Error()
	if !strings.HasSuffix(msg, "...") {
		t.Error("long body should be truncated with ...")
	}
	// Body part should be ~203 chars (200 + "...")
	if strings.Count(msg, "x") > 200 {
		t.Errorf("body should be truncated to 200 chars, got %d x's", strings.Count(msg, "x"))
	}
}

func TestSetGroupNodes_NilCoercion(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body json.RawMessage
		json.NewDecoder(r.Body).Decode(&body)
		if string(body) != "[]" {
			t.Errorf("expected empty array [], got %s", string(body))
		}
		w.WriteHeader(200)
	}))

	if err := c.SetGroupNodes(context.Background(), "test", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
