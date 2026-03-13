package skill

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kittors/AgentFlow/internal/config"
	"github.com/kittors/AgentFlow/internal/targets"
)

const defaultHTTPTimeout = 2 * time.Minute

type InstallOptions struct {
	Skill string
	Ref   string
	Force bool
}

type Manager struct {
	Client    *http.Client
	HomeDir   string
	UserAgent string
}

func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return &Manager{
		Client:    &http.Client{Timeout: defaultHTTPTimeout},
		HomeDir:   homeDir,
		UserAgent: "agentflow-go",
	}
}

func (m *Manager) List(target targets.Target) ([]string, error) {
	dir, err := m.skillsDir(target)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (m *Manager) Uninstall(target targets.Target, name string) error {
	dir, err := m.skillsDir(target)
	if err != nil {
		return err
	}
	if err := ValidateSkillName(name); err != nil {
		return err
	}
	return config.SafeRemove(filepath.Join(dir, name))
}

func (m *Manager) Install(target targets.Target, source string, options InstallOptions) (string, error) {
	skillsDir, err := m.skillsDir(target)
	if err != nil {
		return "", err
	}

	repo := strings.TrimSpace(source)
	skillName := strings.TrimSpace(options.Skill)
	if repo == "" {
		return "", errors.New("missing source")
	}

	if skillName != "" {
		if err := ValidateSkillName(skillName); err != nil {
			return "", err
		}
	}

	if isSkillsDotShURL(repo) {
		resolvedRepo, resolvedSkill, err := ResolveSkillsDotSh(m.Client, repo, m.UserAgent)
		if err != nil {
			return "", err
		}
		repo = resolvedRepo
		if skillName == "" && strings.TrimSpace(resolvedSkill) != "" {
			if err := ValidateSkillName(resolvedSkill); err != nil {
				return "", err
			}
			skillName = resolvedSkill
		}
	}

	owner, name, err := ParseGitHubRepo(repo)
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(options.Ref)
	if ref == "" {
		ref, err = ResolveDefaultBranch(m.Client, owner, name, m.UserAgent)
		if err != nil {
			return "", err
		}
	}

	cacheDir := filepath.Join(m.HomeDir, ".cache", "agentflow", "skills")
	if err := config.EnsureDir(cacheDir); err != nil {
		return "", err
	}

	zipPath, err := DownloadGitHubZip(m.Client, cacheDir, owner, name, ref, m.UserAgent)
	if err != nil {
		return "", err
	}
	defer os.Remove(zipPath)

	extractDir, err := os.MkdirTemp("", "agentflow-skill-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(extractDir)

	rootDir, err := UnzipRoot(zipPath, extractDir)
	if err != nil {
		return "", err
	}

	sourceDir, inferredSkill, err := ResolveSkillDir(rootDir, skillName)
	if err != nil {
		return "", err
	}
	if skillName == "" {
		skillName = inferredSkill
	}
	if skillName == "" {
		skillName = filepath.Base(sourceDir)
	}

	declaredName, _ := ParseSkillName(filepath.Join(sourceDir, "SKILL.md"))
	if strings.TrimSpace(declaredName) != "" {
		if err := ValidateSkillName(declaredName); err != nil {
			return "", err
		}
		skillName = declaredName
	}

	if err := ValidateSkillName(skillName); err != nil {
		return "", err
	}

	destDir := filepath.Join(skillsDir, skillName)
	if _, err := os.Stat(destDir); err == nil {
		if !options.Force {
			return "", fmt.Errorf("skill already installed: %s", skillName)
		}
		if err := config.SafeRemove(destDir); err != nil {
			return "", err
		}
	}
	if err := CopyDir(sourceDir, destDir); err != nil {
		return "", err
	}
	return skillName, nil
}

func (m *Manager) skillsDir(target targets.Target) (string, error) {
	if target.Name != "codex" {
		return "", fmt.Errorf("target does not support skills yet: %s", target.Name)
	}
	return filepath.Join(m.HomeDir, target.Dir, "skills"), nil
}
