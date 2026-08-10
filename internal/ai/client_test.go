package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testServer returns an httptest.Server whose handler inspects the request and
// records it (captured in *capture for assertions).
func testServer(t *testing.T, status int, responseBody string, capture *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*capture = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}))
}

func TestHTTPClient_Complete(t *testing.T) {
	var path string
	srv := testServer(t, http.StatusOK, `{"choices":[{"message":{"content":"bonjour -> hello"}}]}`, &path)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Model: "openai"})
	got, err := c.Complete(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if got != "bonjour -> hello" {
		t.Errorf("reply = %q, want %q", got, "bonjour -> hello")
	}
	if path != "/chat/completions" {
		t.Errorf("request path = %q, want /chat/completions", path)
	}
}

func TestHTTPClient_BaseURLTrailingSlash(t *testing.T) {
	var path string
	srv := testServer(t, http.StatusOK, `{"choices":[{"message":{"content":"ok"}}]}`, &path)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL + "/", Model: "m"})
	if _, err := c.Complete(context.Background(), "", ""); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if path != "/chat/completions" {
		t.Errorf("request path = %q, want /chat/completions", path)
	}
}

func TestHTTPClient_BaseURLEndpointIncluded(t *testing.T) {
	var path string
	srv := testServer(t, http.StatusOK, `{"choices":[{"message":{"content":"ok"}}]}`, &path)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL + "/openai", Model: "openai"})
	if _, err := c.Complete(context.Background(), "", ""); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if path != "/openai/chat/completions" {
		t.Errorf("request path = %q, want /openai/chat/completions", path)
	}
}

func TestHTTPClient_SendsAuthHeader(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Model: "m", APIKey: "sk-test"})
	if _, err := c.Complete(context.Background(), "", ""); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if auth != "Bearer sk-test" {
		t.Errorf("auth header = %q, want Bearer sk-test", auth)
	}
}

func TestHTTPClient_SendsJSON(t *testing.T) {
	var body chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Model: "m"})
	if _, err := c.Complete(context.Background(), "sys", "usr"); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if body.Model != "m" {
		t.Errorf("model = %q, want m", body.Model)
	}
	if len(body.Messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(body.Messages))
	}
	if body.Messages[0].Role != "system" || body.Messages[0].Content != "sys" {
		t.Errorf("system message = %+v", body.Messages[0])
	}
	if body.Messages[1].Role != "user" || body.Messages[1].Content != "usr" {
		t.Errorf("user message = %+v", body.Messages[1])
	}
}

func TestHTTPClient_ErrorStatus(t *testing.T) {
	var path string
	srv := testServer(t, http.StatusTooManyRequests, `{"error":{"message":"rate limited"}}`, &path)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Model: "m"})
	_, err := c.Complete(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for 429")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error should contain upstream message, got: %v", err)
	}
}

func TestHTTPClient_EmptyChoices(t *testing.T) {
	var path string
	srv := testServer(t, http.StatusOK, `{"choices":[]}`, &path)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Model: "m"})
	if _, err := c.Complete(context.Background(), "", ""); err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestHTTPClient_UpstreamErrorField(t *testing.T) {
	var path string
	srv := testServer(t, http.StatusOK, `{"error":{"message":"model not found"}}`, &path)
	defer srv.Close()

	c := NewClient(Config{BaseURL: srv.URL, Model: "m"})
	_, err := c.Complete(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for upstream error field")
	}
	if !strings.Contains(err.Error(), "model not found") {
		t.Errorf("error should contain upstream msg, got: %v", err)
	}
}

func TestHTTPClient_ContextCancel(t *testing.T) {
	c := NewClient(Config{BaseURL: "http://127.0.0.1:1", Model: "m"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Complete(ctx, "", ""); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

type fakeClient struct {
	reply string
	err   error
	sys   string
	user  string
}

func (f *fakeClient) Complete(_ context.Context, system, user string) (string, error) {
	f.sys = system
	f.user = user
	return f.reply, f.err
}

func TestClientInterface_Smoke(t *testing.T) {
	var c Client = &fakeClient{reply: "ok"}
	if got, err := c.Complete(context.Background(), "s", "u"); err != nil || got != "ok" {
		t.Fatalf("fake client: got %q err %v", got, err)
	}
}