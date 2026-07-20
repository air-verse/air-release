package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// githubReleaseNotes asks GitHub to generate release notes for tag (compared
// against previousTag), then regroups the "What's Changed" entries into
// Features/Bug Fixes/Others based on each PR title's conventional-commit
// prefix. GitHub's own New Contributors and Full Changelog sections, which
// depend on data only GitHub has (PR authors, first-time contributors), are
// kept as-is.
func githubReleaseNotes(tag, previousTag string) (string, error) {
	repo, err := gitHubRepo()
	if err != nil {
		return "", err
	}

	args := []string{"api", fmt.Sprintf("repos/%s/releases/generate-notes", repo),
		"-f", "tag_name=" + tag}
	if previousTag != "" {
		args = append(args, "-f", "previous_tag_name="+previousTag)
	}
	out, err := exec.Command("gh", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh api generate-notes: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}

	var resp struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parsing generate-notes response: %w", err)
	}
	return regroupNotes(resp.Body), nil
}

func gitHubRepo() (string, error) {
	out, err := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh repo view: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

var prEntryRe = regexp.MustCompile(`^\* (.+?) (by @\S+ in #\d+)$`)

// regroupNotes splits GitHub's generated "## What's Changed" entries into
// Features/Bug Fixes/Others sections. Everything outside that section (New
// Contributors, Full Changelog) passes through unchanged.
func regroupNotes(body string) string {
	groups := map[string][]string{}
	var tail []string
	inChanged := false
	for _, line := range strings.Split(body, "\n") {
		switch {
		case strings.HasPrefix(line, "## What's Changed"):
			inChanged = true
			continue
		case strings.HasPrefix(line, "## "):
			if inChanged {
				tail = append(tail, "")
			}
			inChanged = false
		}
		if !inChanged {
			tail = append(tail, line)
			continue
		}
		m := prEntryRe.FindStringSubmatch(line)
		if m == nil {
			continue // blank line or an entry GitHub didn't format as "* title by @author in #NNN"
		}
		title, suffix := m[1], m[2]
		group := entryGroup(title)
		groups[group] = append(groups[group], fmt.Sprintf("- %s %s", title, suffix))
	}

	var b strings.Builder
	b.WriteString("## What's Changed\n")
	for _, g := range []string{"Features", "Bug Fixes", "Others"} {
		if len(groups[g]) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n### %s\n", g)
		for _, e := range groups[g] {
			b.WriteString(e + "\n")
		}
	}
	for _, line := range tail {
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// entryGroup categorizes a PR title using the same conventional-commit
// prefix convention as groupFor, without stripping the prefix: PR titles are
// GitHub-facing text, not commit subjects, so the "fix:"/"feat:" prefix stays
// visible in the release notes.
func entryGroup(title string) string {
	m := conventionalRe.FindStringSubmatch(title)
	if m == nil {
		return "Others"
	}
	switch strings.ToLower(m[1]) {
	case "feat":
		return "Features"
	case "fix":
		return "Bug Fixes"
	default:
		return "Others"
	}
}
