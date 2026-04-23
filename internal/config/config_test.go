package config

import (
	"strings"
	"testing"
)

func setEnv(t *testing.T, pairs map[string]string) {
	t.Helper()
	for k, v := range pairs {
		t.Setenv(k, v)
	}
}

func baseEnv() map[string]string {
	return map[string]string{
		"CULLARR_JELLYFIN_URL":    "http://jellyfin:8096",
		"CULLARR_JELLYFIN_APIKEY": "jf-key",
	}
}

func TestFromEnv_Minimal(t *testing.T) {
	setEnv(t, baseEnv())

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jellyfin.URL != "http://jellyfin:8096" {
		t.Errorf("unexpected Jellyfin URL: %q", cfg.Jellyfin.URL)
	}
	if cfg.Sonarr.Enabled || cfg.Radarr.Enabled {
		t.Error("expected Sonarr and Radarr disabled by default")
	}
	if cfg.DryRun {
		t.Error("expected DryRun=false by default")
	}
	if cfg.MinWatchers != 0 {
		t.Errorf("expected MinWatchers=0 by default, got %d", cfg.MinWatchers)
	}
}

func TestFromEnv_AllFields(t *testing.T) {
	setEnv(t, map[string]string{
		"CULLARR_JELLYFIN_URL":      "http://jf:8096",
		"CULLARR_JELLYFIN_APIKEY":   "jf-key",
		"CULLARR_JELLYFIN_USERS":    "alice,bob",
		"CULLARR_SONARR_URL":        "http://sonarr:8989",
		"CULLARR_SONARR_APIKEY":     "sonarr-key",
		"CULLARR_SONARR_ENABLED":    "true",
		"CULLARR_SONARR_UNMONITOR":  "true",
		"CULLARR_RADARR_URL":        "http://radarr:7878",
		"CULLARR_RADARR_APIKEY":     "radarr-key",
		"CULLARR_RADARR_ENABLED":    "true",
		"CULLARR_RADARR_UNMONITOR":  "true",
		"CULLARR_MIN_WATCHERS":      "2",
		"CULLARR_DRY_RUN":           "true",
	})

	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Jellyfin.Users[0] != "alice" || cfg.Jellyfin.Users[1] != "bob" {
		t.Errorf("unexpected users: %v", cfg.Jellyfin.Users)
	}
	if !cfg.Sonarr.Enabled || !cfg.Sonarr.Unmonitor {
		t.Error("expected Sonarr enabled and unmonitor=true")
	}
	if !cfg.Radarr.Enabled || !cfg.Radarr.Unmonitor {
		t.Error("expected Radarr enabled and unmonitor=true")
	}
	if cfg.MinWatchers != 2 {
		t.Errorf("expected MinWatchers=2, got %d", cfg.MinWatchers)
	}
	if !cfg.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestFromEnv_MissingJellyfinURL(t *testing.T) {
	setEnv(t, map[string]string{
		"CULLARR_JELLYFIN_APIKEY": "jf-key",
	})

	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "CULLARR_JELLYFIN_URL") {
		t.Errorf("expected error about CULLARR_JELLYFIN_URL, got %v", err)
	}
}

func TestFromEnv_MissingJellyfinAPIKey(t *testing.T) {
	setEnv(t, map[string]string{
		"CULLARR_JELLYFIN_URL": "http://jf:8096",
	})

	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "CULLARR_JELLYFIN_APIKEY") {
		t.Errorf("expected error about CULLARR_JELLYFIN_APIKEY, got %v", err)
	}
}

func TestFromEnv_SonarrEnabledWithoutURL(t *testing.T) {
	env := baseEnv()
	env["CULLARR_SONARR_ENABLED"] = "true"
	setEnv(t, env)

	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "CULLARR_SONARR_URL") {
		t.Errorf("expected error about CULLARR_SONARR_URL, got %v", err)
	}
}

func TestFromEnv_RadarrEnabledWithoutURL(t *testing.T) {
	env := baseEnv()
	env["CULLARR_RADARR_ENABLED"] = "true"
	setEnv(t, env)

	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "CULLARR_RADARR_URL") {
		t.Errorf("expected error about CULLARR_RADARR_URL, got %v", err)
	}
}

func TestFromEnv_InvalidBool(t *testing.T) {
	env := baseEnv()
	env["CULLARR_DRY_RUN"] = "yes-please"
	setEnv(t, env)

	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "CULLARR_DRY_RUN") {
		t.Errorf("expected error about CULLARR_DRY_RUN, got %v", err)
	}
}

func TestFromEnv_InvalidInt(t *testing.T) {
	env := baseEnv()
	env["CULLARR_MIN_WATCHERS"] = "two"
	setEnv(t, env)

	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "CULLARR_MIN_WATCHERS") {
		t.Errorf("expected error about CULLARR_MIN_WATCHERS, got %v", err)
	}
}
