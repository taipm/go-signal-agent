# Autonomous Mode

**Version:** 1.0.0

Cho phép Go Team chạy tự động mà không cần observer can thiệp tại breakpoints.

---

## Quick Start

```bash
# Enable autonomous mode
*auto

# Enable with specific level
*auto:cautious    # Pause on warnings
*auto:balanced    # Pause on errors only (default)
*auto:aggressive  # Never pause, log all

# Disable autonomous mode
*auto:off
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AUTONOMOUS CONTROLLER                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Decision Engine │  │ Quality Gates   │  │ Auto-Approve│ │
│  │                 │  │                 │  │             │ │
│  │ • Rule matching │  │ • Build pass    │  │ • Specs     │ │
│  │ • Risk scoring  │  │ • Coverage 80%+ │  │ • Designs   │ │
│  │ • Context eval  │  │ • Lint clean    │  │ • Reviews   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     MONITORING LAYER                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Progress Track  │  │ Anomaly Detect  │  │ Rollback    │ │
│  │                 │  │                 │  │ Trigger     │ │
│  │ • Step timing   │  │ • Stuck detect  │  │ • On error  │ │
│  │ • Metric trend  │  │ • Loop detect   │  │ • On anomaly│ │
│  │ • Resource use  │  │ • Quality drop  │  │ • Manual    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Autonomy Levels

### Level 1: Cautious (Conservative)

```yaml
level: cautious
behavior:
  breakpoints: pause_on_warning
  quality_threshold: strict
  auto_approve:
    specs: false        # Always ask
    architecture: false # Always ask
    code_changes: true  # Auto if build pass
    security: false     # Always ask on any finding
  max_iterations: 2
  rollback_on:
    - build_fail
    - test_coverage_drop
    - any_security_issue
```

**Use case:** First time using Go Team, critical projects

### Level 2: Balanced (Default)

```yaml
level: balanced
behavior:
  breakpoints: pause_on_error
  quality_threshold: standard
  auto_approve:
    specs: true         # Auto if complete
    architecture: true  # Auto if valid
    code_changes: true  # Auto if build + tests pass
    security: partial   # Auto on LOW/MEDIUM, pause on HIGH/CRITICAL
  max_iterations: 3
  rollback_on:
    - build_fail_twice
    - coverage_below_70
    - high_security_issue
```

**Use case:** Standard development, familiar projects

### Level 3: Aggressive (Maximum Speed)

```yaml
level: aggressive
behavior:
  breakpoints: never_pause
  quality_threshold: minimal
  auto_approve:
    specs: true
    architecture: true
    code_changes: true
    security: true      # Log and continue
  max_iterations: 5
  rollback_on:
    - build_fail_three_times
    - coverage_below_50
    - critical_security_only
```

**Use case:** Prototyping, non-production code, demos

---

## Decision Rules

### Auto-Approve Specs (Step 2)

```yaml
auto_approve_spec:
  conditions:
    - user_stories >= 1
    - acceptance_criteria_per_story >= 2
    - api_contract_defined: optional
    - scope_documented: true

  reject_if:
    - empty_requirements
    - ambiguous_scope
    - missing_constraints
```

### Auto-Approve Architecture (Step 3)

```yaml
auto_approve_architecture:
  conditions:
    - package_structure_defined: true
    - interfaces_defined: true
    - follows_go_conventions: true
    - dependency_direction_valid: true

  reject_if:
    - circular_dependencies
    - missing_error_handling_strategy
    - no_interfaces_defined
```

### Auto-Approve Code Changes (Step 4-6)

```yaml
auto_approve_code:
  conditions:
    - build_pass: true
    - tests_pass: true
    - coverage >= threshold
    - lint_clean: true
    - no_race_conditions: true

  reject_if:
    - build_fail
    - test_fail
    - coverage_drop > 10%
    - new_lint_errors
```

### Auto-Approve Security (Step 5b)

```yaml
auto_approve_security:
  by_level:
    cautious:
      approve: []
      pause: [LOW, MEDIUM, HIGH, CRITICAL]

    balanced:
      approve: [LOW, MEDIUM]
      pause: [HIGH, CRITICAL]

    aggressive:
      approve: [LOW, MEDIUM, HIGH]
      pause: [CRITICAL]
```

---

## Anomaly Detection

### Stuck Detection

```yaml
stuck_detection:
  triggers:
    - same_step_duration > 10_minutes
    - same_error_repeated > 3_times
    - no_progress_in_iterations > 2

  actions:
    - log_anomaly
    - notify_observer (if available)
    - auto_rollback (if enabled)
    - pause_for_intervention
```

### Loop Detection

```yaml
loop_detection:
  triggers:
    - review_iterations > max_iterations
    - fix_attempt_same_issue > 3
    - oscillating_metrics (coverage up/down repeatedly)

  actions:
    - break_loop
    - document_blocker
    - escalate_or_skip
```

### Quality Drop Detection

```yaml
quality_drop_detection:
  triggers:
    - coverage_decreased > 5%
    - new_lint_errors_introduced
    - build_time_increased > 50%

  actions:
    - flag_regression
    - compare_with_checkpoint
    - suggest_rollback
```

---

## Commands

### Enable/Disable

| Command | Description |
|---------|-------------|
| `*auto` | Enable autonomous mode (balanced) |
| `*auto:cautious` | Enable cautious mode |
| `*auto:balanced` | Enable balanced mode |
| `*auto:aggressive` | Enable aggressive mode |
| `*auto:off` | Disable autonomous mode |

### Monitoring

| Command | Description |
|---------|-------------|
| `*auto:status` | Show current autonomy status |
| `*auto:log` | Show autonomous decisions log |
| `*auto:override` | Manually override current decision |
| `*auto:pause` | Temporarily pause autonomy |
| `*auto:resume` | Resume autonomous mode |

### Configuration

| Command | Description |
|---------|-------------|
| `*auto:config` | Show current configuration |
| `*auto:set {key} {value}` | Set configuration value |
| `*auto:threshold {metric} {value}` | Set quality threshold |

---

## State Management

### Autonomous State

```json
{
  "autonomous": {
    "enabled": true,
    "level": "balanced",
    "started_at": "2025-12-28T23:00:00Z",
    "decisions_made": 5,
    "manual_overrides": 0,

    "current_decision": {
      "step": 3,
      "decision": "approve",
      "reason": "All architecture criteria met",
      "confidence": 0.92
    },

    "quality_gates": {
      "build_pass": true,
      "test_coverage": 85,
      "lint_clean": true,
      "security_clear": true
    },

    "anomalies": [],

    "decision_log": [
      {
        "timestamp": "2025-12-28T23:05:00Z",
        "step": 2,
        "decision": "approve",
        "reason": "Spec complete with 3 user stories"
      }
    ]
  }
}
```

---

## Integration with Workflow

### Step Hook

```yaml
on_breakpoint:
  if autonomous.enabled:
    decision = evaluate_decision_rules(current_step, outputs, metrics)

    if decision.approve:
      log_decision(decision)
      continue_to_next_step()

    else if decision.pause:
      notify_observer(decision.reason)
      wait_for_input()

    else if decision.rollback:
      trigger_rollback(decision.target_checkpoint)
```

### Quality Gate Integration

```yaml
before_auto_approve:
  1. Check quality gates
  2. Evaluate decision rules
  3. Check for anomalies
  4. Calculate confidence score
  5. Make decision:
     - confidence >= 0.8 → approve
     - confidence 0.5-0.8 → depends on level
     - confidence < 0.5 → pause
```

---

## Observer Notifications

Even in autonomous mode, observer receives updates:

```
═══════════════════════════════════════════════════════════
🤖 AUTONOMOUS MODE - Status Update
═══════════════════════════════════════════════════════════

Session: go-team-session-001
Mode: balanced
Running: 15 minutes

Progress:
✓ Step 1: Init (auto)
✓ Step 2: Requirements (auto-approved)
✓ Step 3: Architecture (auto-approved)
→ Step 4: Implementation (in progress)

Decisions Made: 3
Quality Score: 92%

Next checkpoint: After implementation complete

Press [Enter] to intervene or wait for completion...
═══════════════════════════════════════════════════════════
```

---

## Safety Mechanisms

### Hard Stops (Cannot be overridden)

1. **Critical Security Vulnerability** - Always pauses
2. **Build Fail 3+ Times** - Always pauses
3. **Infinite Loop Detected** - Always breaks
4. **Resource Exhaustion** - Always pauses
5. **Observer Interrupt (Ctrl+C)** - Always respects

### Soft Stops (Configurable)

1. Coverage threshold not met
2. High security findings
3. Lint warnings
4. Long-running step
5. Multiple review iterations

---

## Benefits

1. **Faster Development:** No waiting at breakpoints
2. **Consistent Quality:** Automated quality gates
3. **Reduced Cognitive Load:** Observer can focus on results
4. **24/7 Operation:** Can run unattended
5. **Audit Trail:** All decisions logged

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Wrong auto-decision | Checkpoint/rollback available |
| Stuck in loop | Loop detection + auto-break |
| Quality degradation | Quality gates + threshold checks |
| Security issues missed | Hard stop on critical |
| Observer disconnect | Auto-save + resume capability |
