# Cullarr

Deletes watched media files from Sonarr/Radarr based on Jellyfin watch history. Optionally requires multiple users to have watched an item before it is eligible for deletion.

## How it works

1. Fetches watched episodes and movies for each configured Jellyfin user
2. Filters to items watched by at least `CULLARR_MIN_WATCHERS` users
3. Matches each item against Sonarr/Radarr using a three-tier strategy:
   - **Episodes**: TVDB episode ID → (TVDB series ID + S/E numbers) → normalised title
   - **Movies**: TMDB ID → IMDB ID → normalised title
4. Deletes the file and optionally unmonitors the item to prevent re-downloading

Title-based matches (tier 3) are flagged for manual review rather than deleted automatically.

## Quick start

```sh
export CULLARR_JELLYFIN_URL=http://localhost:8096
export CULLARR_JELLYFIN_APIKEY=your-key
export CULLARR_SONARR_URL=http://localhost:8989
export CULLARR_SONARR_APIKEY=your-key
export CULLARR_SONARR_ENABLED=true
export CULLARR_SONARR_UNMONITOR=true
export CULLARR_RADARR_URL=http://localhost:7878
export CULLARR_RADARR_APIKEY=your-key
export CULLARR_RADARR_ENABLED=true
export CULLARR_RADARR_UNMONITOR=true
export CULLARR_DRY_RUN=true

# Preview what would be deleted
cullarr -dry-run

# Delete for real
cullarr
```

## Configuration

All configuration is via environment variables.

| Variable | Required | Description |
|---|---|---|
| `CULLARR_JELLYFIN_URL` | Yes | Jellyfin server URL |
| `CULLARR_JELLYFIN_APIKEY` | Yes | Jellyfin API key (admin key required) |
| `CULLARR_JELLYFIN_USERS` | No | Comma-separated usernames to consider. Omit to use all users. |
| `CULLARR_SONARR_URL` | If Sonarr enabled | Sonarr server URL |
| `CULLARR_SONARR_APIKEY` | If Sonarr enabled | Sonarr API key (Settings → General) |
| `CULLARR_SONARR_ENABLED` | No | Enable Sonarr integration (`true`/`false`, default `false`) |
| `CULLARR_SONARR_UNMONITOR` | No | Unmonitor episodes after deletion (`true`/`false`, default `false`) |
| `CULLARR_RADARR_URL` | If Radarr enabled | Radarr server URL |
| `CULLARR_RADARR_APIKEY` | If Radarr enabled | Radarr API key (Settings → General) |
| `CULLARR_RADARR_ENABLED` | No | Enable Radarr integration (`true`/`false`, default `false`) |
| `CULLARR_RADARR_UNMONITOR` | No | Unmonitor movies after deletion (`true`/`false`, default `false`) |
| `CULLARR_MIN_WATCHERS` | No | Number of users that must have watched before deletion. `0` means all configured users (default `0`) |
| `CULLARR_DRY_RUN` | No | Log what would be deleted without making any changes (`true`/`false`, default `false`) |

Use the internal service URLs for Jellyfin/Sonarr/Radarr (not the public-facing reverse proxy URL), so requests hit the APIs directly without going through SSO.

## Docker

```sh
docker build -t cullarr .

docker run --rm \
  -e CULLARR_JELLYFIN_URL=http://jellyfin:8096 \
  -e CULLARR_JELLYFIN_APIKEY=your-key \
  -e CULLARR_SONARR_URL=http://sonarr:8989 \
  -e CULLARR_SONARR_APIKEY=your-key \
  -e CULLARR_SONARR_ENABLED=true \
  -e CULLARR_SONARR_UNMONITOR=true \
  -e CULLARR_RADARR_URL=http://radarr:7878 \
  -e CULLARR_RADARR_APIKEY=your-key \
  -e CULLARR_RADARR_ENABLED=true \
  -e CULLARR_RADARR_UNMONITOR=true \
  -e CULLARR_DRY_RUN=true \
  cullarr
```

## Docker Compose

Copy `docker-compose.yml`, fill in your values, then run on demand:

```sh
docker compose run --rm cullarr
```

To run on a schedule, invoke this command from a cron job or your server's task scheduler (e.g. Unraid User Scripts, Synology Task Scheduler):

```sh
# Run every night at 3am
0 3 * * * docker compose -f /path/to/docker-compose.yml run --rm cullarr
```

Set `CULLARR_DRY_RUN=true` until you're confident the matches look correct, then switch to `false`.
