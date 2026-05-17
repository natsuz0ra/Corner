package version

import "testing"

func TestInfoFillsDefaults(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	defer func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	}()
	Version, Commit, Date = "", "", ""

	info := Info()
	if info.Version != "dev" || info.Commit != "unknown" || info.Date != "unknown" {
		t.Fatalf("unexpected defaults: %+v", info)
	}
}

func TestInfoReturnsInjectedValues(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	defer func() {
		Version, Commit, Date = oldVersion, oldCommit, oldDate
	}()
	Version, Commit, Date = "v1.26.2", "abc123", "2026-05-17T01:02:03Z"

	info := Info()
	if info.Version != "v1.26.2" || info.Commit != "abc123" || info.Date != "2026-05-17T01:02:03Z" {
		t.Fatalf("unexpected info: %+v", info)
	}
}
