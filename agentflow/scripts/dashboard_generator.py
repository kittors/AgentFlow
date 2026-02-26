"""AgentFlow dashboard generator — create an HTML project status dashboard.

Generates a self-contained HTML file with project metrics, module info,
and recent session history.
"""

from __future__ import annotations

from datetime import datetime, timezone
from pathlib import Path

_HTML_TEMPLATE = """\
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{project_name} — AgentFlow Dashboard</title>
<style>
  :root {{
    --bg: #0d1117; --surface: #161b22; --border: #30363d;
    --text: #e6edf3; --muted: #8b949e; --accent: #58a6ff;
    --green: #3fb950; --yellow: #d29922; --red: #f85149;
  }}
  * {{ margin: 0; padding: 0; box-sizing: border-box; }}
  body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
         background: var(--bg); color: var(--text); padding: 2rem; }}
  .header {{ text-align: center; margin-bottom: 2rem; }}
  .header h1 {{ font-size: 1.8rem; margin-bottom: 0.5rem; }}
  .header .subtitle {{ color: var(--muted); }}
  .grid {{ display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
           gap: 1rem; margin-bottom: 2rem; }}
  .card {{ background: var(--surface); border: 1px solid var(--border);
           border-radius: 8px; padding: 1.2rem; }}
  .card .label {{ color: var(--muted); font-size: 0.85rem; margin-bottom: 0.3rem; }}
  .card .value {{ font-size: 1.6rem; font-weight: 600; }}
  .card .value.green {{ color: var(--green); }}
  .card .value.yellow {{ color: var(--yellow); }}
  .card .value.red {{ color: var(--red); }}
  .section {{ margin-bottom: 2rem; }}
  .section h2 {{ font-size: 1.2rem; margin-bottom: 1rem; color: var(--accent); }}
  table {{ width: 100%; border-collapse: collapse; }}
  th, td {{ padding: 0.6rem 1rem; text-align: left; border-bottom: 1px solid var(--border); }}
  th {{ color: var(--muted); font-weight: 500; font-size: 0.85rem; text-transform: uppercase; }}
  .footer {{ text-align: center; color: var(--muted); font-size: 0.8rem; margin-top: 3rem; }}
</style>
</head>
<body>
<div class="header">
  <h1>📊 {project_name}</h1>
  <p class="subtitle">AgentFlow Dashboard — Generated {generated_at}</p>
</div>

<div class="grid">
  <div class="card">
    <div class="label">Modules</div>
    <div class="value">{module_count}</div>
  </div>
  <div class="card">
    <div class="label">Source Files</div>
    <div class="value">{file_count}</div>
  </div>
  <div class="card">
    <div class="label">Sessions</div>
    <div class="value">{session_count}</div>
  </div>
  <div class="card">
    <div class="label">KB Status</div>
    <div class="value {kb_status_class}">{kb_status}</div>
  </div>
</div>

{modules_section}

{sessions_section}

<div class="footer">
  AgentFlow v1.0.0 — <a href="https://github.com/kittors/AgentFlow" style="color: var(--accent);">GitHub</a>
</div>
</body>
</html>
"""


def _count_source_files(project_root: Path) -> int:
    """Count source files (non-hidden, common extensions)."""
    exts = {".py", ".ts", ".tsx", ".js", ".jsx", ".go", ".rs", ".java"}
    count = 0
    for ext in exts:
        count += len(list(project_root.rglob(f"*{ext}")))
    return count


def generate_dashboard(project_root: Path) -> Path:
    """Generate an HTML dashboard and write it to ``.agentflow/dashboard.html``.

    Returns the path to the generated file.
    """
    kb_root = project_root / ".agentflow" / "kb"
    modules_dir = kb_root / "modules"
    sessions_dir = kb_root / "sessions"

    # Gather metrics
    module_files = list(modules_dir.glob("*.md")) if modules_dir.is_dir() else []
    module_count = max(0, len(module_files) - 1)  # exclude _index.md
    file_count = _count_source_files(project_root)
    session_files = sorted(sessions_dir.glob("*.md"), reverse=True) if sessions_dir.is_dir() else []
    session_count = len(session_files)
    kb_exists = kb_root.is_dir() and (kb_root / "INDEX.md").exists()
    kb_status = "Active" if kb_exists else "Not initialized"
    kb_status_class = "green" if kb_exists else "yellow"

    # Modules table
    modules_rows = ""
    for mf in sorted(module_files):
        if mf.name == "_index.md":
            continue
        modules_rows += f"<tr><td>{mf.stem}</td><td>{mf.stat().st_size}B</td></tr>\n"
    modules_section = ""
    if modules_rows:
        modules_section = f"""
<div class="section">
  <h2>Modules</h2>
  <table><thead><tr><th>Module</th><th>Size</th></tr></thead>
  <tbody>{modules_rows}</tbody></table>
</div>"""

    # Sessions table
    sessions_rows = ""
    for sf in session_files[:10]:
        sessions_rows += f"<tr><td>{sf.stem}</td><td>{sf.stat().st_size}B</td></tr>\n"
    sessions_section = ""
    if sessions_rows:
        sessions_section = f"""
<div class="section">
  <h2>Recent Sessions</h2>
  <table><thead><tr><th>Session</th><th>Size</th></tr></thead>
  <tbody>{sessions_rows}</tbody></table>
</div>"""

    html = _HTML_TEMPLATE.format(
        project_name=project_root.name,
        generated_at=datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC"),
        module_count=module_count,
        file_count=file_count,
        session_count=session_count,
        kb_status=kb_status,
        kb_status_class=kb_status_class,
        modules_section=modules_section,
        sessions_section=sessions_section,
    )

    out = project_root / ".agentflow" / "dashboard.html"
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(html, encoding="utf-8")
    return out


if __name__ == "__main__":
    from .kb_sync import find_project_root

    root = find_project_root()
    if root:
        path = generate_dashboard(root)
        print(f"Dashboard generated: {path}")
    else:
        print("No .agentflow/ directory found.")
