package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGroupFor(t *testing.T) {
	tests := []struct {
		c    commit
		want string
	}{
		{commit{breaking: true, typ: "fix"}, "Breaking Changes"},
		{commit{typ: "feat"}, "Features"},
		{commit{typ: "fix"}, "Bug Fixes"},
		{commit{typ: "perf"}, "Performance"},
		{commit{typ: "chore"}, "Others"},
		{commit{}, "Others"},
	}
	for _, tt := range tests {
		if got := groupFor(tt.c); got != tt.want {
			t.Errorf("groupFor(%+v) = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestChangelogSection(t *testing.T) {
	date := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	commits := []commit{
		{hash: "aaa", typ: "fix", scope: "watcher", subject: "debounce events"},
		{hash: "bbb", typ: "feat", subject: "add color output"},
		{hash: "ccc", typ: "feat", breaking: true, subject: "drop old flag"},
		{hash: "ddd", subject: "update readme"},
	}
	got := changelogSection("v1.2.3", date, commits)
	want := `## v1.2.3 (2026-07-18)

### Breaking Changes

- drop old flag (ccc)

### Features

- add color output (bbb)

### Bug Fixes

- **watcher:** debounce events (aaa)

### Others

- update readme (ddd)
`
	if got != want {
		t.Errorf("changelogSection mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestChangelogSectionOmitsEmptyGroups(t *testing.T) {
	date := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	got := changelogSection("v0.0.2", date, []commit{{hash: "aaa", typ: "fix", subject: "small fix"}})
	want := `## v0.0.2 (2026-07-18)

### Bug Fixes

- small fix (aaa)
`
	if got != want {
		t.Errorf("changelogSection mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrependChangelogNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	if err := prependChangelog(path, "## v0.0.1 (2026-07-18)\n"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Changelog\n\n## v0.0.1 (2026-07-18)\n"
	if string(got) != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestPrependChangelogExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	old := "# Changelog\n\n## v0.0.1 (2026-07-01)\n\n### Bug Fixes\n\n- old fix (aaa)\n"
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := prependChangelog(path, "## v0.0.2 (2026-07-18)\n"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Changelog\n\n## v0.0.2 (2026-07-18)\n\n## v0.0.1 (2026-07-01)\n\n### Bug Fixes\n\n- old fix (aaa)\n"
	if string(got) != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
