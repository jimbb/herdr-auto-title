package grok

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	session = "01a143aa-66ab-7082-a367-6c8c9566b580"
	started = "/work/dashboard"
)

func TestAGeneratedTitleNamesTheSession(t *testing.T) {
	t.Parallel()

	home := writeSummary(t, started, "Fix the login redirect")

	got := NewReader(home).Title(session, started)
	if got != "Fix the login redirect" {
		t.Errorf("title = %q", got)
	}
}

func TestThePaneDirectoryIsTriedBeforeAScan(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	// The decoy sorts before the pane's project, so a scan alone finds it first.
	writeSummaryIn(t, home, "/work/aaa-other", "The other one")
	writeSummaryIn(t, home, started, "This pane")

	got := NewReader(home).Title(session, started)
	if got != "This pane" {
		t.Errorf("title = %q, want the pane's own project", got)
	}
}

func TestAMovedSessionIsFoundByScanning(t *testing.T) {
	t.Parallel()

	home := writeSummary(t, "/work/elsewhere", "Still this session")

	got := NewReader(home).Title(session, started)
	if got != "Still this session" {
		t.Errorf("title = %q", got)
	}
}

func TestAMissingSummaryIsNotTheSession(t *testing.T) {
	t.Parallel()

	if got := NewReader(t.TempDir()).Title(session, started); got != "" {
		t.Errorf("title = %q, want nothing", got)
	}
}

func TestARewriteReplacesTheTitle(t *testing.T) {
	t.Parallel()

	home := writeSummary(t, started, "First cut")
	path := summaryPath(home, started)
	stamp(t, path, 0)

	reader := NewReader(home)

	if got := reader.Title(session, started); got != "First cut" {
		t.Fatalf("title = %q", got)
	}

	writeSummaryIn(t, home, started, "Renamed")
	stamp(t, path, time.Second)

	if got := reader.Title(session, started); got != "Renamed" {
		t.Errorf("title = %q, want the rewrite", got)
	}
}

func TestAnUnchangedSummaryIsReadOnce(t *testing.T) {
	t.Parallel()

	home := writeSummary(t, started, "Fix the login redirect")
	path := summaryPath(home, started)
	stamp(t, path, 0)

	reader := NewReader(home)

	reader.Title(session, started)

	// A rewrite the mtime does not show is one only a re-read could see.
	writeSummaryIn(t, home, started, "Renamed")
	stamp(t, path, 0)

	if got := reader.Title(session, started); got != "Fix the login redirect" {
		t.Errorf("title = %q, want the title kept without re-reading", got)
	}
}

func TestNoHomeAnswersNothing(t *testing.T) {
	t.Parallel()

	if got := NewReader("").Title(session, started); got != "" {
		t.Errorf("title = %q, want nothing", got)
	}
}

func TestRetainForgetsAClosedSession(t *testing.T) {
	t.Parallel()

	home := writeSummary(t, started, "Fix the login redirect")
	reader := NewReader(home)
	reader.Title(session, started)

	reader.Retain(nil)

	if err := os.Remove(summaryPath(home, started)); err != nil {
		t.Fatal(err)
	}

	if got := reader.Title(session, started); got != "" {
		t.Errorf("title = %q, want the forgotten session re-found as missing", got)
	}
}

func writeSummary(t *testing.T, dir, title string) string {
	t.Helper()

	home := t.TempDir()
	writeSummaryIn(t, home, dir, title)

	return home
}

func writeSummaryIn(t *testing.T, home, dir, title string) {
	t.Helper()

	path := summaryPath(home, dir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	body := `{"generated_title":"` + title + `","cwd":"` + dir + `"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// stamp sets a summary's mtime to a fixed point plus offset, so a test does not
// depend on how finely the filesystem keeps time.
func stamp(t *testing.T, path string, offset time.Duration) {
	t.Helper()

	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(offset)
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
}

func summaryPath(home, dir string) string {
	return filepath.Join(home, "sessions", slugOf(dir), session, summaryName)
}
