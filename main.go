// Command air-release computes the next semantic version from conventional
// commits since the latest tag and generates a changelog section.
//
// Run it from the root of the repository you want to release:
//
//	air-release                # preview next version and changelog
//	air-release -write         # also prepend the section to CHANGELOG.md
//	air-release -tag           # also create the git tag
//	air-release -tag -release  # also push the tag and create a GitHub release via gh
//
// Flags combine explicitly and nothing is implied: CHANGELOG.md is only
// written with -write, and -release requires -tag.
//
// While the major version is 0, bumps shift one level down (breaking
// changes bump minor, features bump patch); v1.0.0 must be tagged manually.
//
// The bump level is inferred from the commits, but -bump major|minor|patch
// forces a specific level, bypassing inference and the pre-1.0 downshift.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	write := flag.Bool("write", false, "prepend the new section to CHANGELOG.md")
	tag := flag.Bool("tag", false, "create the git tag")
	release := flag.Bool("release", false, "push the tag and create a GitHub release via gh (requires -tag)")
	forced := flag.String("bump", "", "force the bump level (major, minor, or patch) instead of inferring it from commits")
	flag.Parse()
	if *release && !*tag {
		fatal("-release requires -tag: run air-release -tag -release")
	}
	switch *forced {
	case "", "major", "minor", "patch":
	default:
		fatal("-bump must be major, minor, or patch (got %q)", *forced)
	}

	latest, err := gitOut("describe", "--tags", "--abbrev=0")
	if err != nil {
		if !strings.Contains(err.Error(), "cannot describe anything") {
			fatal("finding latest tag: %v", err)
		}
		// New project with no tags yet: first release, counted from v0.0.0.
		latest = ""
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

	var next, bump string
	if *forced != "" {
		bump = *forced
		next, err = forcedVersion(latest, bump)
	} else {
		next, bump, err = nextVersion(latest, commits)
	}
	if err != nil {
		fatal("%v", err)
	}

	section := changelogSection(next, time.Now(), commits)
	if latest == "" {
		fmt.Println("latest tag:   (none, first release)")
	} else {
		fmt.Printf("latest tag:   %s\n", latest)
	}
	if *forced != "" {
		fmt.Printf("next version: %s (%s bump, forced, %d commits)\n\n", next, bump, len(commits))
	} else {
		fmt.Printf("next version: %s (%s bump, %d commits)\n\n", next, bump, len(commits))
	}
	fmt.Println(section)

	if *write {
		if err := prependChangelog("CHANGELOG.md", section); err != nil {
			fatal("updating CHANGELOG.md: %v", err)
		}
		fmt.Println("CHANGELOG.md updated")
	}
	if *tag {
		if _, err := gitOut("tag", "-a", next, "-m", "Release "+next); err != nil {
			fatal("creating tag: %v", err)
		}
		if *release {
			fmt.Printf("tag %s created\n", next)
		} else {
			fmt.Printf("tag %s created; push it with: git push origin %s\n", next, next)
		}
	}
	if *release {
		if err := createGitHubRelease(next, section); err != nil {
			fatal("creating GitHub release: %v", err)
		}
		fmt.Printf("GitHub release %s created\n", next)
	}
}

// createGitHubRelease pushes the tag and creates a release with the gh CLI,
// using the generated changelog section as the release notes.
func createGitHubRelease(version, notes string) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found in PATH; install it from https://cli.github.com/")
	}
	if _, err := gitOut("push", "origin", version); err != nil {
		return fmt.Errorf("pushing tag: %w", err)
	}
	cmd := exec.Command("gh", "release", "create", version,
		"--title", version, "--notes", notes)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "air-release: "+format+"\n", args...)
	os.Exit(1)
}
