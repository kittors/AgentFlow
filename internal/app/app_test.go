package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelpWritesUsage(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"help"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "install [target|--all]") {
		t.Fatalf("expected usage output, got %q", stdout.String())
	}
}

func TestRunUnknownCommandReturnsOne(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	application := New(stdout, stderr)

	if code := application.Run([]string{"nope"}); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Unknown command") && !strings.Contains(stderr.String(), "未知命令") {
		t.Fatalf("expected unknown command output, got %q", stderr.String())
	}
}

func TestErrorPanelSplitsMultilineErrors(t *testing.T) {
	panel := errorPanel("CLI install failed", fakeError("line one\nline two"))
	if len(panel.Lines) != 2 {
		t.Fatalf("expected multiline error to be split, got %#v", panel.Lines)
	}
	if panel.Lines[0] != "line one" || panel.Lines[1] != "line two" {
		t.Fatalf("unexpected panel lines: %#v", panel.Lines)
	}
}

type fakeError string

func (e fakeError) Error() string {
	return string(e)
}
