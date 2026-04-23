package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Jellyfin    JellyfinConfig
	Sonarr      ArrConfig
	Radarr      RadarrConfig
	MinWatchers int
	DryRun      bool
}

type JellyfinConfig struct {
	URL    string
	APIKey string
	// Users is an explicit allowlist of Jellyfin usernames to consider.
	// Leave empty to use all users discovered via the API.
	Users []string
}

type ArrConfig struct {
	URL string
	// APIKey is found at Settings → General in Sonarr/Radarr.
	APIKey  string
	Enabled bool
	// Unmonitor prevents Sonarr from re-downloading after deletion.
	Unmonitor bool
}

type RadarrConfig struct {
	URL    string
	APIKey string
	Enabled bool
	// Remove deletes the movie from Radarr entirely after the file is deleted.
	Remove bool
}

func FromEnv() (*Config, error) {
	var errs []string

	sonarrEnabled, err := envBool("CULLARR_SONARR_ENABLED")
	if err != nil {
		errs = append(errs, err.Error())
	}
	sonarrUnmonitor, err := envBool("CULLARR_SONARR_UNMONITOR")
	if err != nil {
		errs = append(errs, err.Error())
	}
	radarrEnabled, err := envBool("CULLARR_RADARR_ENABLED")
	if err != nil {
		errs = append(errs, err.Error())
	}
	radarrRemove, err := envBool("CULLARR_RADARR_REMOVE")
	if err != nil {
		errs = append(errs, err.Error())
	}
	minWatchers, err := envInt("CULLARR_MIN_WATCHERS")
	if err != nil {
		errs = append(errs, err.Error())
	}
	dryRun, err := envBool("CULLARR_DRY_RUN")
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid environment variables:\n  %s", strings.Join(errs, "\n  "))
	}

	cfg := &Config{
		Jellyfin: JellyfinConfig{
			URL:    os.Getenv("CULLARR_JELLYFIN_URL"),
			APIKey: os.Getenv("CULLARR_JELLYFIN_APIKEY"),
			Users:  envCSV("CULLARR_JELLYFIN_USERS"),
		},
		Sonarr: ArrConfig{
			URL:       os.Getenv("CULLARR_SONARR_URL"),
			APIKey:    os.Getenv("CULLARR_SONARR_APIKEY"),
			Enabled:   sonarrEnabled,
			Unmonitor: sonarrUnmonitor,
		},
		Radarr: RadarrConfig{
			URL:     os.Getenv("CULLARR_RADARR_URL"),
			APIKey:  os.Getenv("CULLARR_RADARR_APIKEY"),
			Enabled: radarrEnabled,
			Remove:  radarrRemove,
		},
		MinWatchers: minWatchers,
		DryRun:      dryRun,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Jellyfin.URL == "" {
		return fmt.Errorf("CULLARR_JELLYFIN_URL is required")
	}
	if c.Jellyfin.APIKey == "" {
		return fmt.Errorf("CULLARR_JELLYFIN_APIKEY is required")
	}
	if c.Sonarr.Enabled && c.Sonarr.URL == "" {
		return fmt.Errorf("CULLARR_SONARR_URL is required when CULLARR_SONARR_ENABLED=true")
	}
	if c.Radarr.Enabled && c.Radarr.URL == "" {
		return fmt.Errorf("CULLARR_RADARR_URL is required when CULLARR_RADARR_ENABLED=true")
	}
	return nil
}

func envBool(key string) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("%s: invalid boolean %q (use true/false)", key, v)
	}
	return b, nil
}

func envInt(key string) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q", key, v)
	}
	return n, nil
}

func envCSV(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}
