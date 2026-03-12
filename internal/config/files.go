package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AgentFlowMarker = "AGENTFLOW_ROUTER:"
	MarkerScanBytes = 1024
)

func HasMarker(data []byte) bool {
	return strings.Contains(string(data), AgentFlowMarker)
}

func IsAgentFlowFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, MarkerScanBytes)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}
	return HasMarker(buffer[:n])
}

func BackupPath(path string, now time.Time) string {
	source := filepath.Clean(path)
	return filepath.Join(
		filepath.Dir(source),
		fmt.Sprintf("%s_%s_bak%s", strings.TrimSuffix(filepath.Base(source), filepath.Ext(source)), now.Format("20060102150405"), filepath.Ext(source)),
	)
}

func RenameAsidePath(path string, now time.Time) string {
	return fmt.Sprintf("%s._agentflow_old_%s", path, now.Format("20060102150405"))
}

func BackupUserFile(path string) (string, error) {
	source := filepath.Clean(path)
	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}

	backup := BackupPath(source, time.Now())

	input, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(backup, input, info.Mode()); err != nil {
		return "", err
	}
	return backup, nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func SafeWrite(path string, data []byte, mode fs.FileMode) error {
	if mode == 0 {
		mode = 0o644
	}
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".agentflow_*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	cleanup := func() {
		tmp.Close()
		_ = os.Remove(tmpName)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	if err := os.Rename(tmpName, path); err == nil {
		return nil
	} else if os.IsPermission(err) {
		_ = os.Remove(tmpName)
		return os.WriteFile(path, data, mode)
	} else {
		_ = os.Remove(tmpName)
		return err
	}
}

func SafeRemove(path string) error {
	err := os.RemoveAll(path)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if !os.IsPermission(err) {
		return err
	}

	renameTarget := RenameAsidePath(path, time.Now())
	return os.Rename(path, renameTarget)
}
