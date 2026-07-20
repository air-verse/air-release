package main

import "testing"

func TestRegroupNotes(t *testing.T) {
	body := `## What's Changed
* Extend stale workflow timeout by @xiantang in #903
* Increase stale workflow operation limit by @xiantang in #904
* Add review guidelines for coding agents by @xiantang in #905
* Add configurable color output mode by @xiantang in #907
* fix: rewatch files after atomic saves by @xiantang in #908
* follow-up: fix watcher recovery after atomic saves by @xiantang in #909
* Accept .config/air.toml by @bersace in #716
* fix: keep built binary after app shutdown by @mariusvniekerk in #911

## New Contributors
* @bersace made their first contribution in #716

**Full Changelog**: https://github.com/air-verse/air/compare/v1.65.2...v1.65.3`

	got := regroupNotes(body)
	want := `## What's Changed

### Bug Fixes
- fix: rewatch files after atomic saves by @xiantang in #908
- fix: keep built binary after app shutdown by @mariusvniekerk in #911

### Others
- Extend stale workflow timeout by @xiantang in #903
- Increase stale workflow operation limit by @xiantang in #904
- Add review guidelines for coding agents by @xiantang in #905
- Add configurable color output mode by @xiantang in #907
- follow-up: fix watcher recovery after atomic saves by @xiantang in #909
- Accept .config/air.toml by @bersace in #716

## New Contributors
* @bersace made their first contribution in #716

**Full Changelog**: https://github.com/air-verse/air/compare/v1.65.2...v1.65.3
`
	if got != want {
		t.Errorf("regroupNotes() mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestEntryGroup(t *testing.T) {
	cases := map[string]string{
		"feat: add configurable color output mode": "Features",
		"fix: rewatch files after atomic saves":    "Bug Fixes",
		"Add review guidelines for coding agents":  "Others",
		"follow-up: fix watcher recovery":          "Others",
	}
	for title, want := range cases {
		if got := entryGroup(title); got != want {
			t.Errorf("entryGroup(%q) = %q, want %q", title, got, want)
		}
	}
}
