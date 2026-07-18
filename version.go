package main

import (
	"fmt"
	"regexp"
	"strconv"
)

var versionRe = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

func parseVersion(tag string) (major, minor, patch int, err error) {
	m := versionRe.FindStringSubmatch(tag)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("latest tag %q is not a vX.Y.Z version", tag)
	}
	major, _ = strconv.Atoi(m[1])
	minor, _ = strconv.Atoi(m[2])
	patch, _ = strconv.Atoi(m[3])
	return major, minor, patch, nil
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

// nextVersion computes the next version from the latest tag ("" for a repo
// with no tags yet, counted from v0.0.0) and the commits since it.
func nextVersion(latest string, commits []commit) (next, bump string, err error) {
	var major, minor, patch int
	if latest != "" {
		major, minor, patch, err = parseVersion(latest)
		if err != nil {
			return "", "", err
		}
	}

	bump = bumpLevel(commits)
	if major == 0 {
		// Pre-1.0 the API is not considered stable (SemVer item 4), so bumps
		// shift one level down: breaking changes bump minor, features bump
		// patch. v1.0.0 is only ever released by tagging it explicitly.
		switch bump {
		case "major":
			bump = "minor"
		case "minor":
			bump = "patch"
		}
	}
	major, minor, patch = applyBump(major, minor, patch, bump)
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), bump, nil
}

// forcedVersion computes the next version from an explicitly chosen bump
// level, bypassing commit analysis and the pre-1.0 downshift.
func forcedVersion(latest, bump string) (string, error) {
	var major, minor, patch int
	if latest != "" {
		var err error
		major, minor, patch, err = parseVersion(latest)
		if err != nil {
			return "", err
		}
	}
	major, minor, patch = applyBump(major, minor, patch, bump)
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}

func applyBump(major, minor, patch int, bump string) (int, int, int) {
	switch bump {
	case "major":
		return major + 1, 0, 0
	case "minor":
		return major, minor + 1, 0
	default:
		return major, minor, patch + 1
	}
}
