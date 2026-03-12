package kb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kittors/AgentFlow/internal/projectroot"
)

type SessionInput struct {
	Tasks         []string
	Decisions     []string
	Issues        []string
	NextSteps     []string
	FilesModified []string
	Stage         string
	SessionID     string
}

type SessionRecord struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func CreateSession(root string, input SessionInput) (string, error) {
	paths := projectroot.NewPaths(root)
	if err := os.MkdirAll(paths.Sessions, 0o755); err != nil {
		return "", err
	}

	now := time.Now()
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = now.Format("20060102_150405")
	}

	lines := []string{
		fmt.Sprintf("# Session: %s", sessionID),
		fmt.Sprintf("Date: %s", now.Format("2006-01-02 15:04:05")),
	}
	if input.Stage != "" {
		lines = append(lines, fmt.Sprintf("Stage: %s", input.Stage))
	}
	lines = append(lines, "")

	appendSection := func(title string, items []string) {
		lines = append(lines, fmt.Sprintf("## %s", title))
		if len(items) == 0 {
			lines = append(lines, "- (none)")
		} else {
			for _, item := range items {
				lines = append(lines, fmt.Sprintf("- %s", item))
			}
		}
		lines = append(lines, "")
	}

	appendSection("Tasks", input.Tasks)
	appendSection("Decisions", input.Decisions)
	appendSection("Issues", input.Issues)
	appendSection("Files Modified", input.FilesModified)
	appendSection("Next Steps", input.NextSteps)

	filename := filepath.Join(paths.Sessions, sessionID+".md")
	return filename, os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0o644)
}

func SaveStageSnapshot(root, stage, progress, context string) (string, error) {
	paths := projectroot.NewPaths(root)
	if err := os.MkdirAll(paths.Sessions, 0o755); err != nil {
		return "", err
	}

	now := time.Now()
	snapshotID := now.Format("20060102_150405") + "_snap"
	lines := []string{
		fmt.Sprintf("# Stage Snapshot: %s", stage),
		fmt.Sprintf("Date: %s", now.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Stage: %s", stage),
		fmt.Sprintf("Progress: %s", progress),
		"",
	}
	if context != "" {
		lines = append(lines, "## Context", context, "")
	}

	filename := filepath.Join(paths.Sessions, snapshotID+".md")
	return filename, os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0o644)
}

func ListSessions(root string) ([]SessionRecord, error) {
	paths := projectroot.NewPaths(root)
	entries, err := os.ReadDir(paths.Sessions)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]SessionRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		fullPath := filepath.Join(paths.Sessions, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		records = append(records, SessionRecord{
			ID:   strings.TrimSuffix(entry.Name(), ".md"),
			Path: fullPath,
			Size: info.Size(),
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].ID > records[j].ID
	})
	return records, nil
}

func LoadLatestSession(root string) (map[string]string, error) {
	records, err := ListSessions(root)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	content, err := os.ReadFile(records[0].Path)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"id":      records[0].ID,
		"path":    records[0].Path,
		"content": string(content),
	}, nil
}

func PruneSessions(root string, keep int) (int, error) {
	records, err := ListSessions(root)
	if err != nil {
		return 0, err
	}
	if keep < 0 {
		keep = 0
	}
	removed := 0
	for _, record := range records[keep:] {
		if err := os.Remove(record.Path); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func ExportSessions(root, format string) (string, error) {
	records, err := ListSessions(root)
	if err != nil {
		return "", err
	}
	if format == "json" {
		payload, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return "", err
		}
		return string(payload), nil
	}

	lines := make([]string, 0, len(records))
	for _, record := range records {
		lines = append(lines, fmt.Sprintf("[%s] %s (%dB)", record.ID, record.Path, record.Size))
	}
	return strings.Join(lines, "\n"), nil
}
