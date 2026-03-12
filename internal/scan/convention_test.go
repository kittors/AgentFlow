package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanPythonConventionsAndSave(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, filepath.Base(root))
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}
	content := "" +
		"class ExampleClass:\n" +
		"    pass\n\n" +
		"CONSTANT_VALUE = 1\n\n" +
		"def sample_function():\n" +
		"    \"\"\"Args:\n" +
		"    x: test\n" +
		"    Returns:\n" +
		"    value\n" +
		"    \"\"\"\n" +
		"    return CONSTANT_VALUE\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "sample.py"), []byte(content), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	report, err := ScanPythonConventions(root, nil)
	if err != nil {
		t.Fatalf("ScanPythonConventions returned error: %v", err)
	}
	if report.Language != "python" {
		t.Fatalf("expected python language, got %s", report.Language)
	}

	filename, err := SaveConventions(root, report)
	if err != nil {
		t.Fatalf("SaveConventions returned error: %v", err)
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("expected output file, got %v", err)
	}
}
