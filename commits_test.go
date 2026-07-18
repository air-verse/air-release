package main

import (
	"reflect"
	"testing"
)

func record(hash, subject, body string) string {
	return hash + "\x1f" + subject + "\x1f" + body + "\x1e"
}

func TestParseCommits(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want []commit
	}{
		{
			name: "empty output",
			out:  "",
			want: nil,
		},
		{
			name: "conventional commit",
			out:  record("abc1234", "feat: add thing", ""),
			want: []commit{{hash: "abc1234", typ: "feat", subject: "add thing"}},
		},
		{
			name: "scope",
			out:  record("abc1234", "fix(watcher): debounce events", ""),
			want: []commit{{hash: "abc1234", typ: "fix", scope: "watcher", subject: "debounce events"}},
		},
		{
			name: "breaking via bang",
			out:  record("abc1234", "feat!: drop old flag", ""),
			want: []commit{{hash: "abc1234", typ: "feat", breaking: true, subject: "drop old flag"}},
		},
		{
			name: "breaking via body footer",
			out:  record("abc1234", "feat: change defaults", "BREAKING CHANGE: defaults differ"),
			want: []commit{{hash: "abc1234", typ: "feat", breaking: true, subject: "change defaults"}},
		},
		{
			name: "type is lowercased",
			out:  record("abc1234", "Feat: add thing", ""),
			want: []commit{{hash: "abc1234", typ: "feat", subject: "add thing"}},
		},
		{
			name: "non-conventional keeps full subject",
			out:  record("abc1234", "update readme", ""),
			want: []commit{{hash: "abc1234", subject: "update readme"}},
		},
		{
			name: "multiple commits",
			out:  record("aaa", "feat: one", "") + "\n" + record("bbb", "fix: two", ""),
			want: []commit{
				{hash: "aaa", typ: "feat", subject: "one"},
				{hash: "bbb", typ: "fix", subject: "two"},
			},
		},
		{
			name: "malformed record is skipped",
			out:  "just-a-hash\x1e",
			want: nil,
		},
	}
	for _, tt := range tests {
		if got := parseCommits(tt.out); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: parseCommits = %+v, want %+v", tt.name, got, tt.want)
		}
	}
}
