package grok

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestEncodedProjectsAreFoundBeforeScanning(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		dir  string
		slug string
	}{
		{"unreserved", "plain-_.~", "plain-_.~"},
		{"space", "space here", "space%20here"},
		{"plus", "plus+sign", "plus%2Bsign"},
		{"percent", "percent%value", "percent%25value"},
		{"literal escape", "literal%2F", "literal%252F"},
		{"exclamation", "bang!mark", "bang%21mark"},
		{"quote", "single'quote", "single%27quote"},
		{"parentheses", "parentheses()", "parentheses%28%29"},
		{"at sign", "at@sign", "at%40sign"},
		{"hash", "hash#tag", "hash%23tag"},
		{"dollar", "dollar$sign", "dollar%24sign"},
		{"ampersand", "amp&sign", "amp%26sign"},
		{"equals", "equal=value", "equal%3Dvalue"},
		{"comma", "comma,value", "comma%2Cvalue"},
		{"semicolon", "semi;colon", "semi%3Bcolon"},
		{"brackets", "square[brackets]", "square%5Bbrackets%5D"},
		{"braces", "curly{braces}", "curly%7Bbraces%7D"},
		{"caret", "caret^sign", "caret%5Esign"},
		{"backtick", "back`tick", "back%60tick"},
		{"unicode", "\u4e2d\u6587", "%E4%B8%AD%E6%96%87"},
		{"emoji", "\U0001f600", "%F0%9F%98%80"},
		{"accent", "caf\u00e9", "caf%C3%A9"},
		{"mixed", " !'()+%#\u4e2d", "%20%21%27%28%29%2B%25%23%E4%B8%AD"},
		{"nested", `nested\child`, "nested%5Cchild"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertFiledSummary(t, `C:\work\`+test.dir, "C%3A%5Cwork%5C"+test.slug)
		})
	}
}

func TestANativeProjectIsFoundBeforeScanning(t *testing.T) {
	t.Parallel()

	slug := "%2Fwork%2Fdemo"
	if filepath.Separator == '\\' {
		slug = "%5Cwork%5Cdemo"
	}

	assertFiledSummary(t, filepath.FromSlash("/work/demo"), slug)
}

func TestLongEncodedProjectsAreFoundBeforeScanning(t *testing.T) {
	t.Parallel()

	for _, length := range []int{254, 255} {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			t.Parallel()

			const prefix = "C%3A%5Cwork%5C"

			name := strings.Repeat("a", length-len(prefix))
			assertFiledSummary(t, `C:\work\`+name, prefix+name)
		})
	}
}

func TestAHashedProjectIsFoundByScanning(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	path := filepath.Join(home, "sessions", "long-0123456789abcdef", session, summaryName)
	writeSummaryFile(t, path, "This pane")

	dir := `C:\work\` + strings.Repeat("\u4e2d", 40)
	if got := NewReader(home).Title(session, dir); got != "This pane" {
		t.Errorf("title = %q, want the hashed project's summary", got)
	}
}

func assertFiledSummary(t *testing.T, dir, slug string) {
	t.Helper()

	home := t.TempDir()
	// Neither fixture uses slugOf, and scanning alone must find the decoy.
	writeSummaryFile(
		t,
		filepath.Join(home, "sessions", "000-decoy", session, summaryName),
		"Other pane",
	)
	writeSummaryFile(t, filepath.Join(home, "sessions", slug, session, summaryName), "This pane")

	if got := NewReader(home).Title(session, dir); got != "This pane" {
		t.Errorf("title = %q, want the pane's own project before the scan", got)
	}
}
