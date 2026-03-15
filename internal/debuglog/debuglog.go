// Package debuglog provides lightweight debug logging with timing support.
// Logs are written to ~/.agentflow/debug.log when enabled via AGENTFLOW_DEBUG=1.
// It is safe to call any function without calling Init first; they will be no-ops.
package debuglog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	logger  *log.Logger
	enabled bool
)

// Init initialises the debug logger. It creates the log file under ~/.agentflow/
// and enables logging. If AGENTFLOW_DEBUG is not "1" this is a silent no-op.
func Init() {
	if os.Getenv("AGENTFLOW_DEBUG") != "1" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	dir := filepath.Join(home, ".agentflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	path := filepath.Join(dir, "debug.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}

	mu.Lock()
	defer mu.Unlock()
	logger = log.New(f, "", 0)
	enabled = true
	logger.Printf("=== AgentFlow debug session started at %s ===", time.Now().Format(time.RFC3339))
}

// Log writes a formatted debug line with a timestamp prefix.
func Log(format string, args ...any) {
	if !enabled {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	logger.Printf("%s  %s", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
}

// Timed returns a function that, when called, logs the elapsed time since
// the Timed call was made. Typical usage:
//
//	done := debuglog.Timed("statusPanel")
//	defer done()
func Timed(label string) func() {
	if !enabled {
		return func() {}
	}
	start := time.Now()
	Log("[START] %s", label)
	return func() {
		Log("[END]   %s  (%s)", label, time.Since(start).Round(time.Millisecond))
	}
}
