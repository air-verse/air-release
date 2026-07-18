package main

import (
	"regexp"
	"strings"
)

var conventionalRe = regexp.MustCompile(`^([a-zA-Z]+)(\(([^)]+)\))?(!)?: *(.+)$`)

type commit struct {
	hash     string
	typ      string // feat, fix, ... ; empty if not conventional
	scope    string
	breaking bool
	subject  string
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
	return parseCommits(out), nil
}

// parseCommits parses git log output where fields are separated by \x1f and
// commits by \x1e (hash, subject, body).
func parseCommits(out string) []commit {
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
	return commits
}
