package radarr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnmonitorMovie_UsesEditorEndpoint(t *testing.T) {
	var gotPath string
	var gotBody struct {
		MovieIDs  []int `json:"movieIds"`
		Monitored bool  `json:"monitored"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	if err := NewClient(srv.URL, "test-key").UnmonitorMovie(7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v3/movie/editor" {
		t.Errorf("expected path /api/v3/movie/editor, got %q", gotPath)
	}
	if len(gotBody.MovieIDs) != 1 || gotBody.MovieIDs[0] != 7 {
		t.Errorf("expected movieIds=[7], got %v", gotBody.MovieIDs)
	}
	if gotBody.Monitored {
		t.Error("expected monitored=false")
	}
}
