package sonarr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, "test-key")
}

func TestUnmonitorEpisode_UsesBulkMonitorEndpoint(t *testing.T) {
	var gotPath string
	var gotBody struct {
		EpisodeIDs []int `json:"episodeIds"`
		Monitored  bool  `json:"monitored"`
	}

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))

	if err := client.UnmonitorEpisode(Episode{ID: 100, Monitored: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v3/episode/monitor" {
		t.Errorf("expected path /api/v3/episode/monitor, got %q", gotPath)
	}
	if len(gotBody.EpisodeIDs) != 1 || gotBody.EpisodeIDs[0] != 100 {
		t.Errorf("expected episodeIds=[100], got %v", gotBody.EpisodeIDs)
	}
	if gotBody.Monitored {
		t.Error("expected monitored=false")
	}
}

func TestUnmonitorEpisode_NonOKStatus(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))

	if err := client.UnmonitorEpisode(Episode{ID: 100}); err == nil {
		t.Fatal("expected error for non-200/202 response")
	}
}
