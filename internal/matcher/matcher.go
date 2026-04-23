// Package matcher resolves watched Jellyfin items to their counterparts in
// Sonarr/Radarr. The two systems share no common primary key; the only stable
// bridge is the set of external provider IDs (TVDB, IMDB, TMDB) that both
// independently store.
//
// Episode matching (priority order):
//  1. Episode-level TVDB ID   — direct 1:1 match.
//  2. (Series TVDB ID, S, E)  — reliable even when the episode-level TVDB ID
//     is absent from Jellyfin's metadata.
//  3. (Normalised title, S, E) — last resort; logged for human review.
//
// Movie matching:
//  1. TMDB ID
//  2. IMDB ID
//  3. Normalised title — last resort, same caveat.
package matcher

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bvdwalt/cullarr/internal/jellyfin"
	"github.com/bvdwalt/cullarr/internal/radarr"
	"github.com/bvdwalt/cullarr/internal/sonarr"
)

type EpisodeKey struct {
	TvdbSeriesID int
	Season       int
	Episode      int
}

type SonarrEpisodeIndex struct {
	byEpisodeTvdb map[int]sonarr.Episode             // episodeTvdbID → Episode
	bySeriesKey   map[EpisodeKey]sonarr.Episode      // (seriesTvdbID, S, E) → Episode
	byTitleKey    map[titleEpisodeKey]sonarr.Episode // (normTitle, S, E) → Episode
	seriesByTitle map[string]sonarr.Series           // normTitle → Series
}

type titleEpisodeKey struct {
	Title   string
	Season  int
	Episode int
}

// BuildSonarrIndex is O(series × episodes) upfront, making each match O(1).
func BuildSonarrIndex(allSeries []sonarr.Series, getEpisodes func(int) ([]sonarr.Episode, error)) (*SonarrEpisodeIndex, error) {
	idx := &SonarrEpisodeIndex{
		byEpisodeTvdb: map[int]sonarr.Episode{},
		bySeriesKey:   map[EpisodeKey]sonarr.Episode{},
		byTitleKey:    map[titleEpisodeKey]sonarr.Episode{},
		seriesByTitle: map[string]sonarr.Series{},
	}

	for _, s := range allSeries {
		idx.seriesByTitle[normalise(s.Title)] = s

		episodes, err := getEpisodes(s.ID)
		if err != nil {
			return nil, fmt.Errorf("building index for series %q (id=%d): %w", s.Title, s.ID, err)
		}

		for _, ep := range episodes {
			if ep.TvdbID != 0 {
				idx.byEpisodeTvdb[ep.TvdbID] = ep
			}
			if s.TvdbID != 0 {
				idx.bySeriesKey[EpisodeKey{TvdbSeriesID: s.TvdbID, Season: ep.SeasonNumber, Episode: ep.EpisodeNumber}] = ep
			}
			idx.byTitleKey[titleEpisodeKey{Title: normalise(s.Title), Season: ep.SeasonNumber, Episode: ep.EpisodeNumber}] = ep
		}
	}

	return idx, nil
}

type MatchResult struct {
	Found       bool
	Episode     sonarr.Episode
	MatchMethod string
	// FuzzyMatch is true for title-based matches, which should be reviewed
	// manually rather than acted on automatically.
	FuzzyMatch bool
}

func (idx *SonarrEpisodeIndex) FindSonarrEpisode(item jellyfin.Item) MatchResult {
	// Tier 1: episode-level TVDB ID
	if tvdbStr, ok := item.ProviderIds["Tvdb"]; ok && tvdbStr != "" {
		if tvdbID, err := strconv.Atoi(tvdbStr); err == nil && tvdbID != 0 {
			if ep, ok := idx.byEpisodeTvdb[tvdbID]; ok {
				return MatchResult{Found: true, Episode: ep, MatchMethod: "tvdb_episode"}
			}
		}
	}

	// Tier 2: (series TVDB ID, season, episode)
	// Jellyfin doesn't embed the series TVDB ID on each episode item, so we
	// cross-reference via the normalised series title.
	seriesNorm := normalise(item.SeriesName)
	if s, ok := idx.seriesByTitle[seriesNorm]; ok && s.TvdbID != 0 {
		key := EpisodeKey{TvdbSeriesID: s.TvdbID, Season: item.ParentIndexNumber, Episode: item.IndexNumber}
		if ep, ok := idx.bySeriesKey[key]; ok {
			return MatchResult{Found: true, Episode: ep, MatchMethod: "tvdb_series_se"}
		}
	}

	// Tier 3: (normalised title, season, episode)
	tKey := titleEpisodeKey{Title: seriesNorm, Season: item.ParentIndexNumber, Episode: item.IndexNumber}
	if ep, ok := idx.byTitleKey[tKey]; ok {
		return MatchResult{Found: true, Episode: ep, MatchMethod: "title_se", FuzzyMatch: true}
	}

	return MatchResult{Found: false}
}

type RadarrMovieIndex struct {
	byTmdb  map[int]radarr.Movie
	byImdb  map[string]radarr.Movie
	byTitle map[string]radarr.Movie
}

func BuildRadarrIndex(movies []radarr.Movie) *RadarrMovieIndex {
	idx := &RadarrMovieIndex{
		byTmdb:  map[int]radarr.Movie{},
		byImdb:  map[string]radarr.Movie{},
		byTitle: map[string]radarr.Movie{},
	}
	for _, m := range movies {
		if m.TmdbID != 0 {
			idx.byTmdb[m.TmdbID] = m
		}
		if m.ImdbID != "" {
			idx.byImdb[m.ImdbID] = m
		}
		idx.byTitle[normalise(m.Title)] = m
	}
	return idx
}

type MovieMatchResult struct {
	Found       bool
	Movie       radarr.Movie
	MatchMethod string
	FuzzyMatch  bool
}

func (idx *RadarrMovieIndex) FindRadarrMovie(item jellyfin.Item) MovieMatchResult {
	// Tier 1: TMDB ID
	if tmdbStr, ok := item.ProviderIds["Tmdb"]; ok && tmdbStr != "" {
		if tmdbID, err := strconv.Atoi(tmdbStr); err == nil && tmdbID != 0 {
			if m, ok := idx.byTmdb[tmdbID]; ok {
				return MovieMatchResult{Found: true, Movie: m, MatchMethod: "tmdb"}
			}
		}
	}

	// Tier 2: IMDB ID
	if imdbID, ok := item.ProviderIds["Imdb"]; ok && imdbID != "" {
		if m, ok := idx.byImdb[imdbID]; ok {
			return MovieMatchResult{Found: true, Movie: m, MatchMethod: "imdb"}
		}
	}

	// Tier 3: normalised title
	if m, ok := idx.byTitle[normalise(item.Name)]; ok {
		return MovieMatchResult{Found: true, Movie: m, MatchMethod: "title", FuzzyMatch: true}
	}

	return MovieMatchResult{Found: false}
}

var nonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

// normalise makes "The Office (US)" == "the office us" for matching purposes.
func normalise(s string) string {
	return strings.TrimSpace(nonAlpha.ReplaceAllString(strings.ToLower(s), " "))
}

func FormatSE(season, episode int) string {
	return fmt.Sprintf("S%02dE%02d", season, episode)
}
