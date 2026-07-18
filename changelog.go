package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func changelogSection(version string, date time.Time, commits []commit) string {
	groups := map[string][]commit{}
	for _, c := range commits {
		groups[groupFor(c)] = append(groups[groupFor(c)], c)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## %s (%s)\n", version, date.Format("2006-01-02"))
	for _, g := range []string{"Breaking Changes", "Features", "Bug Fixes", "Performance", "Others"} {
		cs := groups[g]
		if len(cs) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n### %s\n\n", g)
		for _, c := range cs {
			if c.scope != "" {
				fmt.Fprintf(&b, "- **%s:** %s (%s)\n", c.scope, c.subject, c.hash)
			} else {
				fmt.Fprintf(&b, "- %s (%s)\n", c.subject, c.hash)
			}
		}
	}
	return b.String()
}

func groupFor(c commit) string {
	switch {
	case c.breaking:
		return "Breaking Changes"
	case c.typ == "feat":
		return "Features"
	case c.typ == "fix":
		return "Bug Fixes"
	case c.typ == "perf":
		return "Performance"
	default:
		return "Others"
	}
}

func prependChangelog(path, section string) error {
	const header = "# Changelog\n"
	old, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	rest := strings.TrimPrefix(string(old), header)
	content := header + "\n" + section + rest
	return os.WriteFile(path, []byte(content), 0o644)
}
