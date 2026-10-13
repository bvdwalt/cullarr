package runner

import (
	"strings"
	"testing"

	"github.com/bvdwalt/cullarr/internal/sonarr"
)

func ep(id, season, num int, hasFile, monitored bool) sonarr.Episode {
	return sonarr.Episode{ID: id, SeasonNumber: season, EpisodeNumber: num, HasFile: hasFile, Monitored: monitored}
}

func ids(xs ...int) map[int]bool {
	m := map[int]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func TestSeriesComplete(t *testing.T) {
	ended := sonarr.Series{Status: "ended"}

	tests := []struct {
		name     string
		series   sonarr.Series
		episodes []sonarr.Episode
		imported map[int]bool
		deleted  map[int]bool
		want     bool
		reason   string
	}{
		{
			name:     "ended, all imported earlier and cleaned up",
			series:   ended,
			episodes: []sonarr.Episode{ep(1, 1, 1, false, false), ep(2, 1, 2, false, false)},
			imported: ids(1, 2),
			want:     true,
		},
		{
			name:     "ended, last episodes deleted this run",
			series:   ended,
			episodes: []sonarr.Episode{ep(1, 1, 1, false, false), ep(2, 1, 2, true, true)},
			imported: ids(1, 2),
			deleted:  ids(2),
			want:     true,
		},
		{
			name:     "continuing series is never complete",
			series:   sonarr.Series{Status: "continuing"},
			episodes: []sonarr.Episode{ep(1, 1, 1, false, false)},
			imported: ids(1),
			want:     false,
			reason:   "not ended",
		},
		{
			name:     "season never downloaded blocks completion",
			series:   ended,
			episodes: []sonarr.Episode{ep(1, 1, 1, false, false), ep(2, 2, 1, false, false)},
			imported: ids(1),
			want:     false,
			reason:   "never downloaded",
		},
		{
			name:     "unwatched file on disk blocks completion",
			series:   ended,
			episodes: []sonarr.Episode{ep(1, 1, 1, false, false), ep(2, 1, 2, true, true)},
			imported: ids(1, 2),
			want:     false,
			reason:   "still on disk",
		},
		{
			name:     "monitored episode without file blocks completion",
			series:   ended,
			episodes: []sonarr.Episode{ep(1, 1, 1, false, true)},
			imported: ids(1),
			want:     false,
			reason:   "still monitored",
		},
		{
			name:     "specials are ignored",
			series:   ended,
			episodes: []sonarr.Episode{ep(1, 1, 1, false, false), ep(9, 0, 1, false, false)},
			imported: ids(1),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := seriesComplete(tt.series, tt.episodes, tt.imported, tt.deleted)
			if got != tt.want {
				t.Fatalf("got %v (%s), want %v", got, reason, tt.want)
			}
			if !tt.want && !strings.Contains(reason, tt.reason) {
				t.Errorf("reason %q does not contain %q", reason, tt.reason)
			}
		})
	}
}
