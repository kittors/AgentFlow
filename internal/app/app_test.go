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
