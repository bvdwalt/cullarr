package matcher

import (
	"testing"

	"github.com/bvdwalt/cullarr/internal/jellyfin"
	"github.com/bvdwalt/cullarr/internal/radarr"
	"github.com/bvdwalt/cullarr/internal/sonarr"
)

func TestNormalise(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"The Office (US)", "the office us"},
		{"Better Call Saul", "better call saul"},
		{"It's Always Sunny in Philadelphia", "it s always sunny in philadelphia"},
		{"S.W.A.T.", "s w a t"},
	}
	for _, c := range cases {
		got := normalise(c.in)
		if got != c.want {
			t.Errorf("normalise(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func makeSonarrIndex(series []sonarr.Series, episodes map[int][]sonarr.Episode) (*SonarrEpisodeIndex, error) {
	return BuildSonarrIndex(series, func(id int) ([]sonarr.Episode, error) {
		return episodes[id], nil
	})
}

func TestFindSonarrEpisode_TvdbEpisodeID(t *testing.T) {
	series := []sonarr.Series{{ID: 1, Title: "Breaking Bad", TvdbID: 81189}}
	episodes := map[int][]sonarr.Episode{
		1: {{ID: 100, TvdbID: 9999, SeasonNumber: 1, EpisodeNumber: 1, HasFile: true}},
	}
	idx, err := makeSonarrIndex(series, episodes)
	if err != nil {
		t.Fatal(err)
	}

	item := jellyfin.Item{
		SeriesName:        "Breaking Bad",
		ParentIndexNumber: 1,
		IndexNumber:       1,
		ProviderIds:       map[string]string{"Tvdb": "9999"},
	}
	res := idx.FindSonarrEpisode(item)
	if !res.Found {
		t.Fatal("expected match, got none")
	}
	if res.MatchMethod != "tvdb_episode" {
		t.Errorf("expected tvdb_episode, got %q", res.MatchMethod)
	}
	if res.FuzzyMatch {
		t.Error("expected FuzzyMatch=false")
	}
}

func TestFindSonarrEpisode_SeriesKeyFallback(t *testing.T) {
	// No episode-level TVDB ID on the Jellyfin item → should fall back to
	// (seriesTvdbID, S, E) which is resolved via series title → Sonarr series.
	series := []sonarr.Series{{ID: 1, Title: "Breaking Bad", TvdbID: 81189}}
	episodes := map[int][]sonarr.Episode{
		// Episode TvdbID is set in Sonarr but NOT in Jellyfin ProviderIds.
		1: {{ID: 100, TvdbID: 9999, SeasonNumber: 1, EpisodeNumber: 1, HasFile: true}},
	}
	idx, err := makeSonarrIndex(series, episodes)
	if err != nil {
		t.Fatal(err)
	}

	item := jellyfin.Item{
		SeriesName:        "Breaking Bad",
		ParentIndexNumber: 1,
		IndexNumber:       1,
		ProviderIds:       map[string]string{}, // no episode-level TVDB
	}
	res := idx.FindSonarrEpisode(item)
	if !res.Found {
		t.Fatal("expected match, got none")
	}
	if res.MatchMethod != "tvdb_series_se" {
		t.Errorf("expected tvdb_series_se, got %q", res.MatchMethod)
	}
}

func TestFindSonarrEpisode_TitleFallback(t *testing.T) {
	// Series has no TVDB ID in Sonarr → falls through to title matching.
	series := []sonarr.Series{{ID: 1, Title: "Some Show", TvdbID: 0}}
	episodes := map[int][]sonarr.Episode{
		1: {{ID: 100, SeasonNumber: 2, EpisodeNumber: 3, HasFile: true}},
	}
	idx, err := makeSonarrIndex(series, episodes)
	if err != nil {
		t.Fatal(err)
	}

	item := jellyfin.Item{
		SeriesName:        "Some Show",
		ParentIndexNumber: 2,
		IndexNumber:       3,
		ProviderIds:       map[string]string{},
	}
	res := idx.FindSonarrEpisode(item)
	if !res.Found {
		t.Fatal("expected match, got none")
	}
	if res.MatchMethod != "title_se" {
		t.Errorf("expected title_se, got %q", res.MatchMethod)
	}
	if !res.FuzzyMatch {
		t.Error("expected FuzzyMatch=true for title match")
	}
}

func TestFindSonarrEpisode_NoMatch(t *testing.T) {
	series := []sonarr.Series{{ID: 1, Title: "Breaking Bad", TvdbID: 81189}}
	episodes := map[int][]sonarr.Episode{
		1: {{ID: 100, TvdbID: 9999, SeasonNumber: 1, EpisodeNumber: 1}},
	}
	idx, err := makeSonarrIndex(series, episodes)
	if err != nil {
		t.Fatal(err)
	}

	item := jellyfin.Item{
		SeriesName:        "Unknown Show",
		ParentIndexNumber: 5,
		IndexNumber:       99,
		ProviderIds:       map[string]string{},
	}
	res := idx.FindSonarrEpisode(item)
	if res.Found {
		t.Error("expected no match, got one")
	}
}

func TestFindRadarrMovie_Tmdb(t *testing.T) {
	movies := []radarr.Movie{{ID: 1, Title: "Inception", TmdbID: 27205}}
	idx := BuildRadarrIndex(movies)

	item := jellyfin.Item{Name: "Inception", ProviderIds: map[string]string{"Tmdb": "27205"}}
	res := idx.FindRadarrMovie(item)
	if !res.Found || res.MatchMethod != "tmdb" {
		t.Errorf("expected tmdb match, got found=%v method=%q", res.Found, res.MatchMethod)
	}
}

func TestFindRadarrMovie_ImdbFallback(t *testing.T) {
	movies := []radarr.Movie{{ID: 1, Title: "Inception", TmdbID: 27205, ImdbID: "tt1375666"}}
	idx := BuildRadarrIndex(movies)

	// Jellyfin has IMDB but not TMDB.
	item := jellyfin.Item{Name: "Inception", ProviderIds: map[string]string{"Imdb": "tt1375666"}}
	res := idx.FindRadarrMovie(item)
	if !res.Found || res.MatchMethod != "imdb" {
		t.Errorf("expected imdb match, got found=%v method=%q", res.Found, res.MatchMethod)
	}
}

func TestFindRadarrMovie_TitleFallback(t *testing.T) {
	movies := []radarr.Movie{{ID: 1, Title: "Inception"}}
	idx := BuildRadarrIndex(movies)

	item := jellyfin.Item{Name: "Inception", ProviderIds: map[string]string{}}
	res := idx.FindRadarrMovie(item)
	if !res.Found || res.MatchMethod != "title" || !res.FuzzyMatch {
		t.Errorf("expected fuzzy title match, got found=%v method=%q fuzzy=%v", res.Found, res.MatchMethod, res.FuzzyMatch)
	}
}
