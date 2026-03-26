package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("expected 'Bearer test-key', got %q", auth)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode([]string{})
	}))
	c.APIKey = "test-key"

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
	calls := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			// POST create
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var req createGroupRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "newgroup" {
				t.Errorf("expected newgroup, got %s", req.Name)
			}
			json.NewEncoder(w).Encode(Group{Name: "newgroup"})
		} else {
			// GET read-back
			json.NewEncoder(w).Encode(Group{Name: "newgroup"})
		}
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
	if c.APIKey != "key" {
		t.Error("wrong api key")
	}
	if c.HTTPClient.Timeout != 10*time.Second {
		t.Error("wrong timeout")
	}
}
