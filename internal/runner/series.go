package runner

import (
	"fmt"

	"github.com/bvdwalt/cullarr/internal/sonarr"
)

// seriesComplete reports whether an ended series has been fully consumed:
// every episode was downloaded at some point, and nothing is left on disk or
// still wanted. deleted holds episode IDs whose files were deleted in this run,
// which are treated as gone even in dry-run mode. Specials are ignored.
func seriesComplete(series sonarr.Series, episodes []sonarr.Episode, imported, deleted map[int]bool) (bool, string) {
	if series.Status != "ended" {
		return false, "series is not ended"
	}
	for _, ep := range episodes {
		if ep.SeasonNumber == 0 {
			continue
		}
		label := fmt.Sprintf("S%02dE%02d", ep.SeasonNumber, ep.EpisodeNumber)
		gone := !ep.HasFile || deleted[ep.ID]
		switch {
		case !gone:
			return false, label + " is still on disk (unwatched)"
		case !deleted[ep.ID] && !imported[ep.ID]:
			return false, label + " was never downloaded"
		case !deleted[ep.ID] && ep.Monitored:
			return false, label + " is still monitored"
		}
	}
	return true, ""
}
