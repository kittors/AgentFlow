# ~init Command

> Initialize project knowledge base.

```yaml
Trigger: User inputs ~init
Flow:
  1. Create .agentflow/ directory (if not exists)
  2. Create .agentflow/kb/ directory
  3. Scan project directory structure
  4. Identify tech stack (package.json, pyproject.toml, Cargo.toml etc.)
  5. Generate .agentflow/kb/INDEX.md (project overview)
  6. Generate .agentflow/kb/context.md (technical context)
  7. Scan modules and generate .agentflow/kb/modules/_index.md + individual module docs
  8. GRAPH_MODE=1: Build initial knowledge graph (.agentflow/kb/graph/nodes.json, edges.json)
  9. CONVENTION_CHECK=1: Extract coding conventions to .agentflow/kb/conventions/extracted.json
  10. Output initialization summary

Directory creation order:
  .agentflow/
  .agentflow/kb/
  .agentflow/kb/modules/
  .agentflow/kb/plan/
  .agentflow/kb/graph/         # When GRAPH_MODE=1
  .agentflow/kb/conventions/   # When CONVENTION_CHECK=1
  .agentflow/kb/archive/
  .agentflow/sessions/         # Session records (not under kb/)

Complex-level large projects:
  - [RLM: native sub-agent] Parallel module scanning (assigned by directory)
```

<init_execution_protocol>

## Initialization Execution Protocol (CRITICAL — must complete all steps)

> **Core principle**: The purpose of ~init is to transform the .agentflow/ directory from empty to containing a knowledge base with actual content.
> Empty directories = initialization failure. Every subdirectory must contain actual files.

### Phase 1: Directory Structure Creation

```bash
# Must execute first — create all directories
mkdir -p .agentflow/kb/modules .agentflow/kb/plan \
         .agentflow/kb/graph .agentflow/kb/conventions .agentflow/kb/archive \
         .agentflow/sessions
```

### Phase 2: Go Command Invocation (MUST — these commands generate actual content)

> The following commands use AgentFlow's built-in Go CLI to generate knowledge base content.
> If `agentflow ...` is unavailable, manually execute equivalent operations.

```yaml
Step A: Template initialization (create base knowledge base structure)
  Command: agentflow init --quiet
  Output: Template reference files under .agentflow/kb/, all subdirectories
  Failure degradation: Manually create INDEX.md and context.md

Step B: Module scanning and KB sync (populate modules/ directory)
  Command: agentflow kb sync --quiet
  Output: .agentflow/kb/modules/_index.md + individual {module}.md for each module
  Failure degradation: Manually scan project source directories, generate module list

Step C: Coding convention extraction (when CONVENTION_CHECK=1, populate conventions/ directory)
  Command: agentflow conventions --quiet
  Output: .agentflow/kb/conventions/extracted.json
  Failure degradation: Manually analyze code style, generate JSON-format convention document

Step D: Knowledge graph initialization (when GRAPH_MODE=1, populate graph/ directory)
  Command: agentflow graph --quiet
  Output: .agentflow/kb/graph/nodes.json, edges.json
  Failure degradation: Manually create initial node and edge data
```

### Phase 3: Manual Content Supplementation (parts not covered by helper scripts)

```yaml
Step E: Generate INDEX.md (project overview)
  - If Phase 2 Step A did not generate it, create manually
  - Content: project name, tech stack, architecture overview, entry points

Step F: Generate context.md (technical context)
  - Content: dependency list, build tools, runtime environment, configuration files

Step G: Create initial session record
  Command: agentflow session save --quiet --stage=INIT
  Failure degradation: Manually create .agentflow/sessions/{timestamp}.md
```

### Phase 4: Verification Checkpoint (MUST — must pass before initialization completes)

```yaml
Verification command: ls -la .agentflow/kb/modules/ .agentflow/kb/conventions/ .agentflow/sessions/

Verification criteria:
  ✅ .agentflow/kb/INDEX.md exists and is non-empty
  ✅ .agentflow/kb/context.md exists and is non-empty
  ✅ .agentflow/kb/modules/_index.md exists and is non-empty
  ✅ .agentflow/kb/modules/ has at least one module document
  ✅ When CONVENTION_CHECK=1: .agentflow/kb/conventions/extracted.json exists
  ✅ When GRAPH_MODE=1: .agentflow/kb/graph/nodes.json exists

Verification failure:
  - Check which step failed
  - Use degradation plan to manually create missing files
  - Re-verify

Prohibited behavior:
  - Reporting initialization complete without executing Phase 2 helper scripts
  - Reporting success after only creating empty directories
  - Skipping Phase 4 verification
```

</init_execution_protocol>
