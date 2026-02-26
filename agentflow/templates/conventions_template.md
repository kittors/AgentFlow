# Conventions Template

## extracted.json Structure

```json
{
  "project": "{project_name}",
  "extracted_at": "{ISO_8601_timestamp}",
  "conventions": {
    "naming": {
      "files": "{snake_case|kebab-case|PascalCase}",
      "functions": "{snake_case|camelCase}",
      "classes": "{PascalCase}",
      "constants": "{UPPER_SNAKE_CASE}",
      "variables": "{snake_case|camelCase}"
    },
    "imports": {
      "style": "{absolute|relative}",
      "order": "{stdlib → third_party → local}",
      "grouping": "{blank_line_between_groups}"
    },
    "error_handling": {
      "pattern": "{try_except|result_type|error_codes}",
      "logging": "{structured|print|logger}"
    },
    "typing": {
      "style": "{annotations|docstring|none}",
      "strictness": "{strict|basic|none}"
    },
    "testing": {
      "framework": "{pytest|jest|go_test}",
      "naming": "{test_function_name|describe_it}",
      "coverage_target": "{percentage}"
    },
    "documentation": {
      "docstring_style": "{google|numpy|sphinx|jsdoc}",
      "required_for": "{public_api|all|none}"
    },
    "git": {
      "commit_style": "{conventional|free_form}",
      "branch_naming": "{feature/xxx|feat-xxx}"
    }
  }
}
```

## Convention Validation Rules

```yaml
validation:
  on_develop_stage: true
  severity: warning  # warning | error
  auto_fix: false    # ruff/eslint --fix 建议
  report_format: inline  # inline | summary
```
