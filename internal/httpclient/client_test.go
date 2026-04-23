package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(srv.URL, "X-Api-Key", "test-key"), srv
}

func TestGet_Success(t *testing.T) {
	type payload struct{ Name string }

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing or wrong auth header")
		}
		_ = json.NewEncoder(w).Encode(payload{Name: "cullarr"})
	}))

	var out payload
	if err := client.Get("/test", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Name != "cullarr" {
		t.Errorf("unexpected response: %+v", out)
	}
}

func TestGet_WithQueryParams(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("seriesId") != "42" {
			t.Errorf("expected seriesId=42, got %q", r.URL.Query().Get("seriesId"))
		}
		_, _ = w.Write([]byte("{}"))
	}))

	var out map[string]any
	params := make(map[string][]string)
	params["seriesId"] = []string{"42"}
	if err := client.Get("/test", params, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_NonOKStatus(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))

	var out any
	err := client.Get("/test", nil, &out)
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestDelete_Success(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.Delete("/test/1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_NonOKStatus(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))

	if err := client.Delete("/test/1"); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestPut_Success(t *testing.T) {
	type payload struct{ Monitored bool }

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json")
		}
		var body payload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if body.Monitored {
			t.Error("expected Monitored=false")
		}
		w.WriteHeader(http.StatusAccepted)
	}))

	body, _ := json.Marshal(map[string]bool{"Monitored": false})
	if err := client.Put("/test/1", body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPut_NonOKStatus(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))

	if err := client.Put("/test/1", []byte("{}")); err == nil {
		t.Fatal("expected error for non-200/202 response")
	}
}
