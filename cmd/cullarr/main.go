package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bvdwalt/cullarr/internal/config"
	"github.com/bvdwalt/cullarr/internal/runner"
)

const usage = `cullarr — delete watched media files from Sonarr/Radarr based on Jellyfin watch status

Usage:
  cullarr [-dry-run]

Flags:
  -dry-run   Force dry-run mode (no mutations), overrides CULLARR_DRY_RUN

Environment variables:
  CULLARR_JELLYFIN_URL        Jellyfin server URL
  CULLARR_JELLYFIN_APIKEY     Jellyfin API key (admin key required)
  CULLARR_JELLYFIN_USERS      Comma-separated usernames to consider (default: all)

  CULLARR_SONARR_URL          Sonarr server URL
  CULLARR_SONARR_APIKEY       Sonarr API key
  CULLARR_SONARR_ENABLED      Enable Sonarr integration (true/false)
  CULLARR_SONARR_UNMONITOR    Unmonitor episodes after deletion (true/false)

  CULLARR_RADARR_URL          Radarr server URL
  CULLARR_RADARR_APIKEY       Radarr API key
  CULLARR_RADARR_ENABLED      Enable Radarr integration (true/false)
  CULLARR_RADARR_UNMONITOR    Unmonitor movies after deletion (true/false)

  CULLARR_MIN_WATCHERS        Users that must have watched before deletion (0 = all)
  CULLARR_DRY_RUN             Dry-run mode — log what would be deleted (true/false)
`

func main() {
	dryRun := flag.Bool("dry-run", false, "force dry-run mode regardless of CULLARR_DRY_RUN")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	cfg, err := config.FromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		cfg.DryRun = true
	}

	if err := runner.Run(cfg); err != nil {
		os.Exit(1)
	}
}
