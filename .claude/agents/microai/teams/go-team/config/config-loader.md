# Go Team Config Loader

## Overview

Config loader module cho go-team workflow. Load, validate, và manage configuration at runtime.

---

## Config Loading Protocol

### 1. Initialization

```yaml
on_session_start:
  1. Load config from: .claude/agents/microai/teams/go-team/config/config.yaml
  2. Apply any session overrides from state
  3. Validate all values are within allowed ranges
  4. Store in go_team_state.config
```

### 2. Config State Structure

```yaml
go_team_state:
  config:
    loaded_at: "2025-12-29T01:30:00+07:00"
    source: "config.yaml"
    overrides: {}

    # Active configuration (after overrides applied)
    active:
      iterations:
        max: 3
        current: 0
      coverage:
        threshold: 80
      security:
        block_critical: true
        block_high: true
      kanban:
        enabled: true
        emit_signals: true
```

---

## Command Handlers

### *config - Show Configuration

```yaml
command: "*config"
handler:
  action: display_config
  output: |
    ═══════════════════════════════════════════════════
      GO TEAM CONFIGURATION
    ═══════════════════════════════════════════════════

    Iterations:     ${config.iterations.max} (default: 3)
    Coverage:       ${config.coverage.threshold}% (default: 80%)
    Security Gate:  ${config.security.block_high ? "Strict" : "Lenient"}
    Kanban:         ${config.kanban.enabled ? "Enabled" : "Disabled"}
    Parallel:       ${config.parallel.enabled ? "Enabled" : "Disabled"}
    Autonomous:     ${config.autonomous.enabled ? "Enabled" : "Disabled"}

    Use *config:key for specific value
    Use *config:key=value to modify
```

### *config:key - Show Specific Value

```yaml
command: "*config:{key}"
handler:
  action: get_config_value
  params:
    key: "${key}"
  output: "${key}: ${value}"
```

### *config:key=value - Set Value

```yaml
command: "*config:{key}={value}"
handler:
  action: set_config_value
  params:
    key: "${key}"
    value: "${value}"
  validation:
    - check_key_exists
    - check_value_in_range
    - check_type_match
  on_success:
    - update_state
    - emit_signal: config_changed
    - display: "Config updated: ${key} = ${value}"
  on_error:
    - display_error
```

### *iterations - Iteration Commands

```yaml
# Show current
command: "*iterations"
handler:
  output: "Max iterations: ${config.iterations.max} (current: ${state.review.iteration})"

# Set value
command: "*iterations:{N}"
handler:
  validation:
    - N >= 1 AND N <= 10
  action:
    - set config.iterations.max = N
    - emit_signal: iterations_changed
  output: "Iterations set to ${N}"

# Add more
command: "*iterations:+{N}"
handler:
  validation:
    - current + N <= 10
  action:
    - config.iterations.max += N
  output: "Added ${N} iterations. New max: ${config.iterations.max}"

# Reset
command: "*iterations:reset"
handler:
  action:
    - config.iterations.max = 3
  output: "Iterations reset to default (3)"
```

### *coverage - Coverage Commands

```yaml
# Show current
command: "*coverage"
handler:
  output: "Coverage threshold: ${config.coverage.threshold}%"

# Set value
command: "*coverage:{N}"
handler:
  validation:
    - N >= 50 AND N <= 100
  action:
    - set config.coverage.threshold = N
  output: "Coverage threshold set to ${N}%"

# Reset
command: "*coverage:reset"
handler:
  action:
    - config.coverage.threshold = 80
  output: "Coverage reset to default (80%)"
```

---

## Validation Rules

### Range Validators

```yaml
validators:
  iterations.max:
    type: integer
    min: 1
    max: 10
    default: 3

  coverage.threshold:
    type: integer
    min: 50
    max: 100
    default: 80

  agents.timeouts.default:
    type: integer
    min: 60
    max: 1800
    default: 300

  fixer.max_lines:
    type: integer
    min: 5
    max: 50
    default: 20

  metrics.tokens.budget:
    type: integer
    min: 10000
    max: 500000
    default: 100000
```

### Boolean Validators

```yaml
validators:
  kanban.enabled:
    type: boolean
    default: true

  security.gate.block_on_critical:
    type: boolean
    default: true
    immutable: true  # Cannot be changed

  checkpoint.enabled:
    type: boolean
    default: true
```

---

## Config Persistence

### Save Current Config

```yaml
command: "*config:save"
handler:
  action:
    - write_config_to_file
    - backup_previous_config
  output: "Configuration saved to config.yaml"
```

### Load Config

```yaml
on_load:
  1. Check file exists
  2. Parse YAML
  3. Validate schema
  4. Apply defaults for missing values
  5. Merge with session overrides
  6. Return active config
```

### Config Override Priority

```
1. Runtime commands (*config:key=value)  - Highest
2. Session state overrides
3. config.yaml file
4. Hardcoded defaults                     - Lowest
```

---

## Integration with Workflow

### Access Config in Steps

```yaml
# In any step file
step_execution:
  before:
    - load_config
    - max_iterations = get_config("iterations.max")
    - coverage_threshold = get_config("coverage.threshold")

  during:
    - if get_config("kanban.signals.emit_step_start"):
        emit_signal("step_started", step_info)
```

### Config-Driven Behavior

```yaml
# Example: Review Loop uses config
step_06_review_loop:
  max_iterations: ${get_config("iterations.max")}

  exit_conditions:
    coverage: ">= ${get_config('coverage.threshold')}"
    lint: "${get_config('quality.lint.required') ? 'clean' : 'any'}"
    race: "${get_config('quality.race.required') ? 'none' : 'any'}"
```

---

## Error Handling

### Invalid Config Value

```yaml
on_invalid_value:
  message: |
    Invalid config value for '${key}'
    Expected: ${validator.description}
    Got: ${value}
    Allowed range: ${validator.min} - ${validator.max}
  action: reject_and_keep_current
```

### Missing Config File

```yaml
on_missing_config:
  action:
    - create_default_config
    - notify_user
  message: "Config file not found. Created default config.yaml"
```

### Corrupted Config

```yaml
on_parse_error:
  action:
    - restore_from_backup
    - fallback_to_defaults
  message: "Config file corrupted. Restored from backup."
```

---

## Config Functions (Pseudo-code)

```python
def get_config(key: str, default=None):
    """Get config value by dot-notation key."""
    parts = key.split('.')
    value = go_team_state.config.active

    for part in parts:
        if part in value:
            value = value[part]
        else:
            return default

    return value

def set_config(key: str, value):
    """Set config value with validation."""
    # Validate
    validator = validators.get(key)
    if validator and not validator.validate(value):
        raise ConfigValidationError(f"Invalid value for {key}")

    # Apply
    parts = key.split('.')
    target = go_team_state.config.active
    for part in parts[:-1]:
        target = target[part]
    target[parts[-1]] = value

    # Record override
    go_team_state.config.overrides[key] = value

    # Emit signal
    if get_config("kanban.enabled"):
        emit_signal("config_changed", {"key": key, "value": value})

def load_config():
    """Load config from file and apply overrides."""
    # Load file
    config = parse_yaml("config/config.yaml")

    # Validate schema
    validate_config_schema(config)

    # Apply session overrides
    for key, value in go_team_state.config.overrides.items():
        set_nested(config, key, value)

    # Store active config
    go_team_state.config.active = config
    go_team_state.config.loaded_at = now()

    return config

def reset_config(key: str = None):
    """Reset config to defaults."""
    if key:
        # Reset specific key
        default = get_default(key)
        set_config(key, default)
        del go_team_state.config.overrides[key]
    else:
        # Reset all
        go_team_state.config.overrides = {}
        load_config()
```

---

## Usage Examples

### Example 1: Increase Iterations Mid-Session

```
User: *iterations:5

System:
  ✓ Iterations set to 5
  Current progress: Iteration 2 of 5

  Use *iterations to check, or *iterations:reset for default
```

### Example 2: Lower Coverage for Prototype

```
User: *coverage:60

System:
  ⚠ Coverage threshold lowered to 60%
  Note: Production code typically requires 80%+

  Use *coverage:reset to restore default
```

### Example 3: View Full Config

```
User: *config

System:
  ═══════════════════════════════════════════════════
    GO TEAM CONFIGURATION
  ═══════════════════════════════════════════════════

  Workflow:
    Type: full_pipeline
    Steps: 9 enabled

  Quality:
    Iterations: 5 (modified from default 3)
    Coverage: 60% (modified from default 80%)
    Lint: Required
    Race Check: Required

  Security:
    Block Critical: Yes (locked)
    Block High: Yes
    Tools: gosec, govulncheck, staticcheck

  Kanban:
    Enabled: Yes
    Signal Emission: Yes
    WIP Enforcement: Yes

  Metrics:
    Token Tracking: Yes
    Budget: 100,000 tokens

  Modified values marked with *
  Use *config:reset to restore all defaults
```
