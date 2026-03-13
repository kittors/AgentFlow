package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveSkillDir(rootDir, skillName string) (string, string, error) {
	rootDir = filepath.Clean(rootDir)
	if strings.TrimSpace(skillName) != "" {
		candidates := []string{
			filepath.Join(rootDir, "skills", skillName),
			filepath.Join(rootDir, skillName),
		}
		for _, candidate := range candidates {
			if ok, _ := fileExists(filepath.Join(candidate, "SKILL.md")); ok {
				return candidate, skillName, nil
			}
		}
		return "", "", fmt.Errorf("skill not found in repo: %s", skillName)
	}

	dirs, err := FindSkillDirs(rootDir)
	if err != nil {
		return "", "", err
	}
	if len(dirs) == 0 {
		return "", "", errors.New("no skills found in repository")
	}
	if len(dirs) > 1 {
		return "", "", fmt.Errorf("multiple skills found; specify --skill (found %d)", len(dirs))
	}
	inferred := filepath.Base(dirs[0])
	return dirs[0], inferred, nil
}

func FindSkillDirs(rootDir string) ([]string, error) {
	rootDir = filepath.Clean(rootDir)

	var matches []string
	err := filepath.WalkDir(rootDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "node_modules" || name == ".github" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "SKILL.md" {
			return nil
		}
		matches = append(matches, filepath.Dir(path))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}
