package cmd

import (
	"fmt"

	"github.com/bvdwalt/cullarr/internal/config"
	"github.com/bvdwalt/cullarr/internal/jellyfin"
	"github.com/bvdwalt/cullarr/internal/logger"
	"github.com/bvdwalt/cullarr/internal/matcher"
	"github.com/bvdwalt/cullarr/internal/radarr"
	"github.com/bvdwalt/cullarr/internal/sonarr"
)

func Run(cfg *config.Config) error {
	log := logger.New(cfg.DryRun)

	jf := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.APIKey)

	log.Section("Jellyfin: resolving users")
	users, err := resolveUsers(jf, cfg.Jellyfin.Users)
	if err != nil {
		return err
	}
	log.Info("Using %d user(s): %v", len(users), userNames(users))

	if len(cfg.Jellyfin.Users) > 0 {
		found := map[string]bool{}
		for _, u := range users {
			found[u.Name] = true
		}
		for _, name := range cfg.Jellyfin.Users {
			if !found[name] {
				log.Warn("configured user %q not found in Jellyfin — skipping", name)
			}
		}
	}

	minWatchers := cfg.MinWatchers
	if minWatchers == 0 {
		minWatchers = len(users)
	}
	log.Info("Min watchers required for deletion: %d", minWatchers)

	log.Section("Jellyfin: collecting watch data")

	// Keyed by Jellyfin item ID to avoid double-counting the same episode
	// watched twice by the same user.
	episodeWatchers := map[string][]string{}
	episodeItems := map[string]jellyfin.Item{}
	movieWatchers := map[string][]string{}
	movieItems := map[string]jellyfin.Item{}

	for _, u := range users {
		episodes, err := jf.GetWatchedEpisodes(u.ID)
		if err != nil {
			return fmt.Errorf("fetching watched episodes for user %q: %w", u.Name, err)
		}
		for _, ep := range episodes {
			episodeWatchers[ep.ID] = append(episodeWatchers[ep.ID], u.Name)
			episodeItems[ep.ID] = ep
		}
		log.Info("%-20s  episodes watched: %d", u.Name, len(episodes))

		movies, err := jf.GetWatchedMovies(u.ID)
		if err != nil {
			return fmt.Errorf("fetching watched movies for user %q: %w", u.Name, err)
		}
		for _, m := range movies {
			movieWatchers[m.ID] = append(movieWatchers[m.ID], u.Name)
			movieItems[m.ID] = m
		}
		log.Info("%-20s  movies watched:   %d", u.Name, len(movies))
	}

	var eligibleEpisodes []eligibleItem
	for id, watchers := range episodeWatchers {
		if len(watchers) >= minWatchers {
			eligibleEpisodes = append(eligibleEpisodes, eligibleItem{item: episodeItems[id], watchedBy: watchers})
		}
	}

	var eligibleMovies []eligibleItem
	for id, watchers := range movieWatchers {
		if len(watchers) >= minWatchers {
			eligibleMovies = append(eligibleMovies, eligibleItem{item: movieItems[id], watchedBy: watchers})
		}
	}

	log.Info("Eligible episodes (watched by >=%d users): %d", minWatchers, len(eligibleEpisodes))
	log.Info("Eligible movies   (watched by >=%d users): %d", minWatchers, len(eligibleMovies))

	if cfg.Sonarr.Enabled && len(eligibleEpisodes) > 0 {
		if err := processSonarr(cfg, log, eligibleEpisodes); err != nil {
			return err
		}
	}

	if cfg.Radarr.Enabled && len(eligibleMovies) > 0 {
		if err := processRadarr(cfg, log, eligibleMovies); err != nil {
			return err
		}
	}

	log.Summary()
	return nil
}

func processSonarr(cfg *config.Config, log *logger.Logger, eligible []eligibleItem) error {
	log.Section("Sonarr: building episode index")

	sc := sonarr.NewClient(cfg.Sonarr.URL, cfg.Sonarr.APIKey)

	allSeries, err := sc.GetAllSeries()
	if err != nil {
		return fmt.Errorf("sonarr: %w", err)
	}
	log.Info("Series in Sonarr: %d", len(allSeries))

	idx, err := matcher.BuildSonarrIndex(allSeries, sc.GetEpisodes)
	if err != nil {
		return err
	}

	log.Section("Sonarr: processing watched episodes")

	for _, ei := range eligible {
		result := idx.FindSonarrEpisode(ei.item)
		detail := matcher.FormatSE(ei.item.ParentIndexNumber, ei.item.IndexNumber)
		title := ei.item.SeriesName

		if !result.Found {
			log.Unmatched("episode", title, detail,
				"no match in Sonarr via tvdb_episode / tvdb_series_se / title_se",
				ei.watchedBy)
			continue
		}

		ep := result.Episode

		// Fuzzy title matches are quarantined for manual review — a
		// normalisation mismatch could cause the wrong episode to be deleted.
		if result.FuzzyMatch {
			log.Unmatched("episode", title, detail,
				fmt.Sprintf("matched via %q (fuzzy) — verify manually before deleting", result.MatchMethod),
				ei.watchedBy)
			continue
		}

		if !ep.HasFile {
			log.Skipped("episode", title, detail, "no file on disk")
			continue
		}

		log.Deleted("episode", title, detail, ei.watchedBy, result.MatchMethod)
		if !cfg.DryRun {
			if err := sc.DeleteEpisodeFile(ep.EpisodeFileID); err != nil {
				return fmt.Errorf("deleting episode file: %w", err)
			}
		}

		if cfg.Sonarr.Unmonitor {
			log.Unmonitored("episode", title, detail)
			if !cfg.DryRun {
				if err := sc.UnmonitorEpisode(ep); err != nil {
					return fmt.Errorf("unmonitoring episode: %w", err)
				}
			}
		}
	}

	return nil
}

func processRadarr(cfg *config.Config, log *logger.Logger, eligible []eligibleItem) error {
	log.Section("Radarr: building movie index")

	rc := radarr.NewClient(cfg.Radarr.URL, cfg.Radarr.APIKey)

	allMovies, err := rc.GetAllMovies()
	if err != nil {
		return fmt.Errorf("radarr: %w", err)
	}
	log.Info("Movies in Radarr: %d", len(allMovies))

	idx := matcher.BuildRadarrIndex(allMovies)

	log.Section("Radarr: processing watched movies")

	for _, ei := range eligible {
		result := idx.FindRadarrMovie(ei.item)

		if !result.Found {
			log.Unmatched("movie", ei.item.Name, "",
				"no match in Radarr via tmdb / imdb / title",
				ei.watchedBy)
			continue
		}

		m := result.Movie

		if result.FuzzyMatch {
			log.Unmatched("movie", ei.item.Name, "",
				fmt.Sprintf("matched via %q (fuzzy) — verify manually", result.MatchMethod),
				ei.watchedBy)
			continue
		}

		if !m.HasFile {
			log.Skipped("movie", ei.item.Name, "", "no file on disk")
			continue
		}

		log.Deleted("movie", ei.item.Name, "", ei.watchedBy, result.MatchMethod)
		if !cfg.DryRun {
			if err := rc.DeleteMovieFile(m.MovieFileID); err != nil {
				return fmt.Errorf("deleting movie file: %w", err)
			}
		}

		if cfg.Radarr.Unmonitor {
			log.Unmonitored("movie", ei.item.Name, "")
			if !cfg.DryRun {
				if err := rc.UnmonitorMovie(m); err != nil {
					return fmt.Errorf("unmonitoring movie: %w", err)
				}
			}
		}
	}

	return nil
}

type eligibleItem struct {
	item      jellyfin.Item
	watchedBy []string
}

func resolveUsers(jf *jellyfin.Client, allowlist []string) ([]jellyfin.User, error) {
	all, err := jf.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("fetching Jellyfin users: %w", err)
	}

	if len(allowlist) == 0 {
		return all, nil
	}

	allowed := map[string]bool{}
	for _, name := range allowlist {
		allowed[name] = true
	}

	var filtered []jellyfin.User
	for _, u := range all {
		if allowed[u.Name] {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func userNames(users []jellyfin.User) []string {
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Name
	}
	return names
}
