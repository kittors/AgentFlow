package scan

import (
	"html/template"
	"os"
	"path/filepath"

	"github.com/kittors/AgentFlow/internal/kb"
	"github.com/kittors/AgentFlow/internal/projectroot"
)

type DashboardData struct {
	Project      string
	ModuleCount  int
	SourceFiles  int
	SessionCount int
	KBStatus     string
	Modules      []kb.Module
	Sessions     []kb.SessionRecord
}

func GenerateDashboard(root string, sourceDirs []string) (string, error) {
	modules, err := kb.ScanModules(root, sourceDirs)
	if err != nil {
		return "", err
	}
	sessions, err := kb.ListSessions(root)
	if err != nil {
		return "", err
	}

	sourceFiles := 0
	for _, module := range modules {
		sourceFiles += module.FileCount
	}

	kbStatus := "Not initialized"
	if _, err := os.Stat(filepath.Join(root, ".agentflow", "kb")); err == nil {
		kbStatus = "Active"
	}

	data := DashboardData{
		Project:      projectroot.ProjectName(root),
		ModuleCount:  len(modules),
		SourceFiles:  sourceFiles,
		SessionCount: len(sessions),
		KBStatus:     kbStatus,
		Modules:      modules,
		Sessions:     sessions,
	}

	tpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return "", err
	}

	paths := projectroot.NewPaths(root)
	if err := os.MkdirAll(paths.AgentFlow, 0o755); err != nil {
		return "", err
	}
	filename := filepath.Join(paths.AgentFlow, "dashboard.html")
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := tpl.Execute(file, data); err != nil {
		return "", err
	}
	return filename, nil
}

const dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>AgentFlow Dashboard</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 24px; color: #1d1d1f; }
    h1, h2 { margin-bottom: 12px; }
    .stats { display: grid; grid-template-columns: repeat(4, minmax(120px, 1fr)); gap: 12px; margin-bottom: 24px; }
    .card { border: 1px solid #d2d2d7; border-radius: 12px; padding: 16px; background: #fafafc; }
    table { width: 100%; border-collapse: collapse; margin-bottom: 24px; }
    th, td { text-align: left; padding: 10px; border-bottom: 1px solid #e5e5ea; }
  </style>
</head>
<body>
  <h1>AgentFlow Dashboard</h1>
  <div class="stats">
    <div class="card"><strong>Modules</strong><div>{{.ModuleCount}}</div></div>
    <div class="card"><strong>Source Files</strong><div>{{.SourceFiles}}</div></div>
    <div class="card"><strong>Sessions</strong><div>{{.SessionCount}}</div></div>
    <div class="card"><strong>KB Status</strong><div>{{.KBStatus}}</div></div>
  </div>

  <h2>Modules</h2>
  <table>
    <thead><tr><th>Name</th><th>Path</th><th>Files</th></tr></thead>
    <tbody>
      {{range .Modules}}
      <tr><td>{{.Name}}</td><td>{{.Path}}</td><td>{{.FileCount}}</td></tr>
      {{end}}
    </tbody>
  </table>

  <h2>Sessions</h2>
  <table>
    <thead><tr><th>ID</th><th>Path</th><th>Size</th></tr></thead>
    <tbody>
      {{range .Sessions}}
      <tr><td>{{.ID}}</td><td>{{.Path}}</td><td>{{.Size}}</td></tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>
`
