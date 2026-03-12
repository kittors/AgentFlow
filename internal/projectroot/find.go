package projectroot

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrRootNotFound = errors.New("agentflow project root not found")

func Find(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if info, err := os.Stat(filepath.Join(current, AgentFlowDir)); err == nil && info.IsDir() {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", ErrRootNotFound
		}
		current = parent
	}
}
