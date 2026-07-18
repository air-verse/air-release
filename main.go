// Command air-release computes the next semantic version from conventional
// commits since the latest tag and generates a changelog section.
//
// Run it from the root of the repository you want to release:
//
//	air-release           # preview next version and changelog
//	air-release -write    # also prepend the section to CHANGELOG.md
//	air-release -tag      # also create the git tag (implies -write)
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var conventionalRe = regexp.MustCompile(`^([a-zA-Z]+)(\(([^)]+)\))?(!)?: *(.+)$`)

type commit struct {
	hash     string
	typ      string // feat, fix, ... ; empty if not conventional
	scope    string
	breaking bool
	subject  string
}

func main() {
	write := flag.Bool("write", false, "prepend the new section to CHANGELOG.md")
	tag := flag.Bool("tag", false, "create the git tag (implies -write)")
	flag.Parse()

	latest, err := gitOut("describe", "--tags", "--abbrev=0")
	if err != nil {
		if !strings.Contains(err.Error(), "cannot describe anything") {
			fatal("finding latest tag: %v", err)
		}
		// New project with no tags yet: first release, counted from v0.0.0.
		latest = ""
	}
	var major, minor, patch int
	if latest != "" {
		major, minor, patch, err = parseVersion(latest)
		if err != nil {
			fatal("%v", err)
		}
	}

	commits, err := commitsSince(latest)
	if err != nil {
		fatal("reading commits: %v", err)
	}
	if len(commits) == 0 {
		if latest == "" {
			fmt.Println("no commits yet, nothing to release")
		} else {
			fmt.Printf("no commits since %s, nothing to release\n", latest)
		}
		return
	}

	bump := bumpLevel(commits)
	switch bump {
	case "major":
		major, minor, patch = major+1, 0, 0
	case "minor":
		minor, patch = minor+1, 0
	default:
		patch++
	}
	next := fmt.Sprintf("v%d.%d.%d", major, minor, patch)

	section := changelogSection(next, commits)
	if latest == "" {
		fmt.Println("latest tag:   (none, first release)")
	} else {
		fmt.Printf("latest tag:   %s\n", latest)
	}
	fmt.Printf("next version: %s (%s bump, %d commits)\n\n", next, bump, len(commits))
	fmt.Println(section)

	if *write || *tag {
		if err := prependChangelog(section); err != nil {
			fatal("updating CHANGELOG.md: %v", err)
		}
		fmt.Println("CHANGELOG.md updated")
	}
	if *tag {
		if _, err := gitOut("tag", "-a", next, "-m", "Release "+next); err != nil {
			fatal("creating tag: %v", err)
		}
		fmt.Printf("tag %s created; push it with: git push origin %s\n", next, next)
	}
}

func gitOut(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func parseVersion(tag string) (major, minor, patch int, err error) {
	m := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`).FindStringSubmatch(tag)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("latest tag %q is not a vX.Y.Z version", tag)
	}
	major, _ = strconv.Atoi(m[1])
	minor, _ = strconv.Atoi(m[2])
	patch, _ = strconv.Atoi(m[3])
	return major, minor, patch, nil
}

func commitsSince(tag string) ([]commit, error) {
	// %x1f separates fields, %x1e separates commits.
	rangeSpec := "HEAD"
	if tag != "" {
		rangeSpec = tag + "..HEAD"
	}
	out, err := gitOut("log", rangeSpec, "--no-merges", "--pretty=format:%h%x1f%s%x1f%b%x1e")
	if err != nil {
		if tag == "" && strings.Contains(err.Error(), "unknown revision") {
			return nil, nil // repository has no commits yet
		}
		return nil, err
	}
	var commits []commit
	for _, rec := range strings.Split(out, "\x1e") {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		fields := strings.SplitN(rec, "\x1f", 3)
		if len(fields) < 2 {
			continue
		}
		c := commit{hash: fields[0], subject: fields[1]}
		body := ""
		if len(fields) == 3 {
			body = fields[2]
		}
		if m := conventionalRe.FindStringSubmatch(c.subject); m != nil {
			c.typ = strings.ToLower(m[1])
			c.scope = m[3]
			c.breaking = m[4] == "!"
			c.subject = m[5]
		}
		if strings.Contains(body, "BREAKING CHANGE") {
			c.breaking = true
		}
		commits = append(commits, c)
	}
	return commits, nil
}

func bumpLevel(commits []commit) string {
	level := "patch"
	for _, c := range commits {
		if c.breaking {
			return "major"
		}
		if c.typ == "feat" {
			level = "minor"
		}
	}
	return level
}

func changelogSection(version string, commits []commit) string {
	groups := map[string][]commit{}
	for _, c := range commits {
		groups[groupFor(c)] = append(groups[groupFor(c)], c)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## %s (%s)\n", version, time.Now().Format("2006-01-02"))
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

func prependChangelog(section string) error {
	const path = "CHANGELOG.md"
	const header = "# Changelog\n"
	old, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	rest := strings.TrimPrefix(string(old), header)
	content := header + "\n" + section + rest
	return os.WriteFile(path, []byte(content), 0o644)
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "air-release: "+format+"\n", args...)
	os.Exit(1)
}
