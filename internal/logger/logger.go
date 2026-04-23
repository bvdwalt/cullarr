package logger

import (
	"fmt"
	"strings"
	"time"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	grey   = "\033[90m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	blue   = "\033[34m"
)

type ActionKind string

const (
	ActionDeleted     ActionKind = "DELETED"
	ActionDryRun      ActionKind = "DRY-RUN"
	ActionSkipped     ActionKind = "SKIPPED"   // watched but no file on disk
	ActionUnmatched   ActionKind = "UNMATCHED" // could not correlate Jellyfin→Arr
	ActionUnmonitored ActionKind = "UNMONITORED"
)

type Entry struct {
	Kind        ActionKind
	MediaType   string // "episode" or "movie"
	Title       string
	Detail      string // e.g. "S01E03"
	WatchedBy   []string
	MatchMethod string
}

type Logger struct {
	dryRun  bool
	entries []Entry
	start   time.Time
}

func New(dryRun bool) *Logger {
	l := &Logger{dryRun: dryRun, start: time.Now()}
	if dryRun {
		fmt.Printf("\n%s%s🔍 DRY-RUN MODE — no files will be deleted%s\n\n", bold, yellow, reset)
	}
	return l
}

func (l *Logger) Section(title string) {
	fmt.Printf("%s%s── %s %s%s\n", bold, blue, title, strings.Repeat("─", max(0, 50-len(title))), reset)
}

func (l *Logger) Info(format string, args ...any) {
	fmt.Printf("%s  %s\n", grey, fmt.Sprintf(format, args...)+reset)
}

func (l *Logger) Warn(format string, args ...any) {
	fmt.Printf("  %s%s[WARN]%s %s\n", bold, yellow, reset, fmt.Sprintf(format, args...))
}

func (l *Logger) Deleted(mediaType, title, detail string, watchedBy []string, matchMethod string) {
	e := Entry{
		Kind:        ActionDeleted,
		MediaType:   mediaType,
		Title:       title,
		Detail:      detail,
		WatchedBy:   watchedBy,
		MatchMethod: matchMethod,
	}
	if l.dryRun {
		e.Kind = ActionDryRun
		fmt.Printf("  %s%s[DRY-RUN]%s %s %s%s%s  — watched by: %s  (match: %s)\n",
			bold, yellow, reset,
			mediaType, bold, label(title, detail), reset,
			strings.Join(watchedBy, ", "), matchMethod)
	} else {
		fmt.Printf("  %s%s[DELETED]%s %s %s%s%s  — watched by: %s\n",
			bold, green, reset,
			mediaType, bold, label(title, detail), reset,
			strings.Join(watchedBy, ", "))
	}
	l.entries = append(l.entries, e)
}

func (l *Logger) Unmonitored(mediaType, title, detail string) {
	e := Entry{Kind: ActionUnmonitored, MediaType: mediaType, Title: title, Detail: detail}
	prefix := "[UNMONITORED]"
	if l.dryRun {
		prefix = "[DRY-RUN:UNMONITOR]"
	}
	fmt.Printf("  %s%s%s%s %s %s%s%s\n",
		bold, cyan, prefix, reset,
		mediaType, bold, label(title, detail), reset)
	l.entries = append(l.entries, e)
}

func (l *Logger) Skipped(mediaType, title, detail, reason string) {
	e := Entry{Kind: ActionSkipped, MediaType: mediaType, Title: title, Detail: detail}
	fmt.Printf("  %s[SKIPPED]%s %s %s%s%s — %s\n",
		grey, reset, mediaType, bold, label(title, detail), reset, reason)
	l.entries = append(l.entries, e)
}

func (l *Logger) Unmatched(mediaType, title, detail, reason string, watchedBy []string) {
	e := Entry{
		Kind:      ActionUnmatched,
		MediaType: mediaType,
		Title:     title,
		Detail:    detail,
		WatchedBy: watchedBy,
	}
	fmt.Printf("  %s%s[UNMATCHED]%s %s %s%s%s — %s\n",
		bold, red, reset, mediaType, bold, label(title, detail), reset, reason)
	l.entries = append(l.entries, e)
}

func (l *Logger) Summary() {
	elapsed := time.Since(l.start).Round(time.Millisecond)

	counts := map[ActionKind]int{}
	for _, e := range l.entries {
		counts[e.Kind]++
	}

	fmt.Printf("\n%s%s── Summary %s%s\n", bold, blue, strings.Repeat("─", 41), reset)
	fmt.Printf("  Elapsed:     %s\n", elapsed)

	if l.dryRun {
		fmt.Printf("  Would delete: %s%d%s\n", bold, counts[ActionDryRun], reset)
	} else {
		fmt.Printf("  Deleted:      %s%s%d%s\n", bold, green, counts[ActionDeleted], reset)
	}
	fmt.Printf("  Unmonitored:  %d\n", counts[ActionUnmonitored])
	fmt.Printf("  Skipped:      %d\n", counts[ActionSkipped])

	if n := counts[ActionUnmatched]; n > 0 {
		fmt.Printf("  %sUnmatched:    %d  ← review these manually%s\n", red, n, reset)
		fmt.Printf("\n%s%s── Unmatched items (need manual review) %s%s\n", bold, red, strings.Repeat("─", 12), reset)
		for _, e := range l.entries {
			if e.Kind == ActionUnmatched {
				fmt.Printf("  %s %s  (watched by: %s)\n",
					e.MediaType, label(e.Title, e.Detail), strings.Join(e.WatchedBy, ", "))
			}
		}
	}
	fmt.Println()
}

func label(title, detail string) string {
	if detail == "" {
		return title
	}
	return fmt.Sprintf("%s — %s", title, detail)
}
