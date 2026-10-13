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

func TestUnmonitorSeries_UsesEditorEndpoint(t *testing.T) {
	var gotPath string
	var gotBody struct {
		SeriesIDs []int `json:"seriesIds"`
		Monitored bool  `json:"monitored"`
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

	if err := client.UnmonitorSeries(46); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v3/series/editor" {
		t.Errorf("expected path /api/v3/series/editor, got %q", gotPath)
	}
	if len(gotBody.SeriesIDs) != 1 || gotBody.SeriesIDs[0] != 46 {
		t.Errorf("expected seriesIds=[46], got %v", gotBody.SeriesIDs)
	}
	if gotBody.Monitored {
		t.Error("expected monitored=false")
	}
}

func TestDeleteSeries_KeepsFilesAndExclusionOff(t *testing.T) {
	var gotPath, gotQuery string

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))

	if err := client.DeleteSeries(46); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v3/series/46" {
		t.Errorf("expected path /api/v3/series/46, got %q", gotPath)
	}
	if gotQuery != "addImportListExclusion=false&deleteFiles=false" {
		t.Errorf("unexpected query %q", gotQuery)
	}
}

func TestGetImportedEpisodeIDs_FiltersToImportEvents(t *testing.T) {
	var gotPath, gotSeriesID string

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotSeriesID = r.URL.Query().Get("seriesId")
		_, _ = w.Write([]byte(`[
			{"episodeId": 1, "eventType": "grabbed"},
			{"episodeId": 1, "eventType": "downloadFolderImported"},
			{"episodeId": 2, "eventType": "downloadFailed"},
			{"episodeId": 3, "eventType": "downloadFolderImported"}
		]`))
	}))

	got, err := client.GetImportedEpisodeIDs(46)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v3/history/series" || gotSeriesID != "46" {
		t.Errorf("unexpected request: path=%q seriesId=%q", gotPath, gotSeriesID)
	}
	if !got[1] || !got[3] || got[2] || len(got) != 2 {
		t.Errorf("expected only episodes 1 and 3, got %v", got)
	}
}
