// Package audience decides which users must have watched an item before it
// may be deleted. Items tagged "<prefix><username>" in Sonarr/Radarr are
// limited to the tagged users; untagged items require every user.
package audience

import (
	"strings"
)

const DefaultTagPrefix = "cullarr-"

// NormaliseUser converts a Jellyfin username to the form used in tag labels.
// Sonarr/Radarr only allow lowercase letters, digits and hyphens in tags.
func NormaliseUser(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '.':
			b.WriteRune('-')
		}
	}
	return b.String()
}

// Resolve returns the users who must have watched an item. tagged is true when
// at least one tag matched a known user. unknown lists prefixed tags that
// matched no user. When nothing matches, the audience is all users.
func Resolve(tagLabels, users []string, prefix string) (audience []string, tagged bool, unknown []string) {
	byTag := map[string]string{}
	for _, u := range users {
		byTag[NormaliseUser(u)] = u
	}

	seen := map[string]bool{}
	for _, label := range tagLabels {
		label = strings.ToLower(label)
		if !strings.HasPrefix(label, prefix) {
			continue
		}
		name := strings.TrimPrefix(label, prefix)
		u, ok := byTag[name]
		if !ok {
			unknown = append(unknown, label)
			continue
		}
		if !seen[u] {
			seen[u] = true
			audience = append(audience, u)
		}
	}

	if len(audience) == 0 {
		return users, false, unknown
	}
	return audience, true, unknown
}

// Eligible reports whether an item may be deleted. For tagged items every
// audience user must have watched it. For untagged items every user must have,
// unless minWatchers > 0, which instead requires that many watchers.
func Eligible(audience, watchedBy []string, tagged bool, minWatchers int) bool {
	if !tagged && minWatchers > 0 {
		return len(watchedBy) >= minWatchers
	}
	if len(audience) == 0 {
		return false
	}
	watched := map[string]bool{}
	for _, w := range watchedBy {
		watched[w] = true
	}
	for _, u := range audience {
		if !watched[u] {
			return false
		}
	}
	return true
}
