package main

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		tag                 string
		major, minor, patch int
		wantErr             bool
	}{
		{tag: "v1.2.3", major: 1, minor: 2, patch: 3},
		{tag: "v0.0.0", major: 0, minor: 0, patch: 0},
		{tag: "v10.20.30", major: 10, minor: 20, patch: 30},
		{tag: "1.2.3", wantErr: true},
		{tag: "v1.2", wantErr: true},
		{tag: "v1.2.3-rc1", wantErr: true},
		{tag: "release-1", wantErr: true},
		{tag: "", wantErr: true},
	}
	for _, tt := range tests {
		major, minor, patch, err := parseVersion(tt.tag)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseVersion(%q): expected error, got %d.%d.%d", tt.tag, major, minor, patch)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseVersion(%q): unexpected error: %v", tt.tag, err)
			continue
		}
		if major != tt.major || minor != tt.minor || patch != tt.patch {
			t.Errorf("parseVersion(%q) = %d.%d.%d, want %d.%d.%d",
				tt.tag, major, minor, patch, tt.major, tt.minor, tt.patch)
		}
	}
}

func TestBumpLevel(t *testing.T) {
	tests := []struct {
		name    string
		commits []commit
		want    string
	}{
		{"no conventional commits", []commit{{subject: "whatever"}}, "patch"},
		{"fix only", []commit{{typ: "fix"}}, "patch"},
		{"feat", []commit{{typ: "fix"}, {typ: "feat"}}, "minor"},
		{"breaking wins", []commit{{typ: "feat"}, {typ: "fix", breaking: true}}, "major"},
		{"empty", nil, "patch"},
	}
	for _, tt := range tests {
		if got := bumpLevel(tt.commits); got != tt.want {
			t.Errorf("%s: bumpLevel = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNextVersion(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		commits  []commit
		wantNext string
		wantBump string
	}{
		{"post-1.0 breaking bumps major", "v1.2.3", []commit{{breaking: true}}, "v2.0.0", "major"},
		{"post-1.0 feat bumps minor", "v1.2.3", []commit{{typ: "feat"}}, "v1.3.0", "minor"},
		{"post-1.0 fix bumps patch", "v1.2.3", []commit{{typ: "fix"}}, "v1.2.4", "patch"},
		{"pre-1.0 breaking shifts to minor", "v0.5.2", []commit{{breaking: true}}, "v0.6.0", "minor"},
		{"pre-1.0 feat shifts to patch", "v0.5.2", []commit{{typ: "feat"}}, "v0.5.3", "patch"},
		{"pre-1.0 fix stays patch", "v0.5.2", []commit{{typ: "fix"}}, "v0.5.3", "patch"},
		{"first release counts from v0.0.0", "", []commit{{typ: "feat"}}, "v0.0.1", "patch"},
		{"first release with breaking", "", []commit{{breaking: true}}, "v0.1.0", "minor"},
	}
	for _, tt := range tests {
		next, bump, err := nextVersion(tt.latest, tt.commits)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
			continue
		}
		if next != tt.wantNext || bump != tt.wantBump {
			t.Errorf("%s: nextVersion = (%q, %q), want (%q, %q)",
				tt.name, next, bump, tt.wantNext, tt.wantBump)
		}
	}
}

func TestNextVersionBadTag(t *testing.T) {
	if _, _, err := nextVersion("not-a-version", []commit{{typ: "fix"}}); err == nil {
		t.Error("expected error for non-semver latest tag")
	}
}
