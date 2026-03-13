package skill

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUnzipTotalBytes = 250 << 20
	maxUnzipFileBytes  = 50 << 20
)

func UnzipRoot(zipPath, destDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var totalDeclared uint64
	for _, file := range reader.File {
		cleanName := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanName, "..") || strings.Contains(cleanName, ":") || filepath.IsAbs(cleanName) {
			return "", fmt.Errorf("unsafe zip path: %s", file.Name)
		}
		if file.UncompressedSize64 > maxUnzipFileBytes {
			return "", fmt.Errorf("zip entry too large: %s", file.Name)
		}
		totalDeclared += file.UncompressedSize64
	}
	if totalDeclared > maxUnzipTotalBytes {
		return "", fmt.Errorf("zip content too large")
	}

	var totalWritten int64
	for _, file := range reader.File {
		written, err := unzipFile(file, destDir)
		if err != nil {
			return "", err
		}
		totalWritten += written
		if totalWritten > maxUnzipTotalBytes {
			return "", fmt.Errorf("zip extraction exceeded size limit")
		}
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(destDir, entry.Name()), nil
		}
	}
	return "", errors.New("zip did not contain a root directory")
}

func unzipFile(file *zip.File, destDir string) (int64, error) {
	targetPath := filepath.Join(destDir, filepath.FromSlash(file.Name))
	if file.FileInfo().IsDir() {
		return 0, os.MkdirAll(targetPath, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return 0, err
	}

	input, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer input.Close()

	mode := os.FileMode(0o644)
	if file.Mode()&0o111 != 0 {
		mode = 0o755
	}
	output, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return 0, err
	}
	limited := io.LimitReader(input, maxUnzipFileBytes+1)
	written, err := io.Copy(output, limited)
	if err != nil {
		output.Close()
		return 0, err
	}
	if written > maxUnzipFileBytes {
		output.Close()
		return 0, fmt.Errorf("zip entry exceeded size limit: %s", file.Name)
	}
	if err := output.Close(); err != nil {
		return 0, err
	}
	if err := os.Chmod(targetPath, mode); err != nil {
		return 0, err
	}
	return written, nil
}
