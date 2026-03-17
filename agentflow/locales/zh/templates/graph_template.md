# Knowledge Graph Template

## nodes.json Structure

```json
{
  "version": 1,
  "nodes": [
    {
      "id": "{unique_id}",
      "type": "module|file|function|class|concept|dependency",
      "name": "{display_name}",
      "path": "{file_path_or_null}",
      "metadata": {
        "description": "{brief_description}",
        "created_at": "{ISO_8601}",
        "updated_at": "{ISO_8601}",
        "tags": ["{tag1}", "{tag2}"]
      }
    }
  ]
}
```

## edges.json Structure

```json
{
  "version": 1,
  "edges": [
    {
      "source": "{node_id}",
      "target": "{node_id}",
      "type": "imports|extends|implements|depends_on|calls|contains|references",
      "weight": 1.0,
      "metadata": {
        "description": "{relationship_detail}",
        "created_at": "{ISO_8601}"
      }
    }
  ]
}
```

## Edge Types

| Type | Description | Example |
|------|-------------|---------|
| `imports` | Module import relationship | `auth.py` imports `utils.py` |
| `extends` | Class inheritance | `AdminUser` extends `User` |
| `implements` | Interface implementation | `SQLRepo` implements `Repository` |
| `depends_on` | Package dependency | `project` depends on `fastapi` |
| `calls` | Function call relationship | `handler()` calls `validate()` |
| `contains` | Parent-child containment | `module/` contains `routes.py` |
| `references` | General reference | `README` references `API spec` |

## Query Patterns

```yaml
# Find all dependencies of a module
query: neighbors(node_id, edge_type="depends_on", direction="outbound")

# Find all files that import a given module
query: neighbors(node_id, edge_type="imports", direction="inbound")

# Find circular dependencies
query: cycles(edge_type="imports")

# Find orphan nodes (no connections)
query: isolates()
```
