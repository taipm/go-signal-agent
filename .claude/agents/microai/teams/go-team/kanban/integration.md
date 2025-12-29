# Go Team - Kanban Integration

## Overview

Tích hợp kanban-agent với go-team workflow để:
- Track tiến độ real-time
- Enforce WIP limits
- Collect metrics per agent
- Generate reports

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GO TEAM SESSION                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐     signals     ┌──────────────┐          │
│  │ Orchestrator │ ───────────────→│ Kanban Agent │          │
│  │    Agent     │←─────────────── │  (Observer)  │          │
│  └──────┬───────┘   wip_status    └──────┬───────┘          │
│         │                                 │                  │
│         │ activates                       │ updates          │
│         ↓                                 ↓                  │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                   AGENTS                              │   │
│  │  PM → Arch → Coder → Test → Security → Review →...   │   │
│  └──────────────────────────────────────────────────────┘   │
│         │                                 ↑                  │
│         │ step_complete                   │ board_update     │
│         └─────────────────────────────────┘                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Signal Types

### From Orchestrator to Kanban

```yaml
# Session started
signal:
  type: session_started
  session_id: "go-team-{uuid}"
  topic: "JWT Authentication"
  workflow: "full_pipeline"
  timestamp: "2025-12-29T00:30:00+07:00"

# Step started
signal:
  type: step_started
  session_id: "go-team-{uuid}"
  step: "step-02"
  step_name: "Requirements"
  agent: "pm-agent"
  timestamp: "2025-12-29T00:31:00+07:00"

# Step completed
signal:
  type: step_completed
  session_id: "go-team-{uuid}"
  step: "step-02"
  agent: "pm-agent"
  duration_seconds: 180
  outputs:
    - "spec.md"
  timestamp: "2025-12-29T00:34:00+07:00"

# Agent activated
signal:
  type: agent_activated
  session_id: "go-team-{uuid}"
  agent: "architect-agent"
  task_id: "task-{session}-arch"
  timestamp: "2025-12-29T00:34:00+07:00"

# Security gate result
signal:
  type: security_gate
  session_id: "go-team-{uuid}"
  result: "BLOCKED"  # or "PASSED" or "PASSED_WITH_WARNINGS"
  severity: "HIGH"
  issues_count: 2
  timestamp: "2025-12-29T00:45:00+07:00"

# Review iteration
signal:
  type: review_iteration
  session_id: "go-team-{uuid}"
  iteration: 2
  max_iterations: 3
  issues_found: 3
  timestamp: "2025-12-29T00:50:00+07:00"

# Session completed
signal:
  type: session_completed
  session_id: "go-team-{uuid}"
  result: "SUCCESS"  # or "PARTIAL" or "ABORTED"
  duration_seconds: 1800
  metrics:
    coverage: 85
    security_passed: true
    files_created: 12
  timestamp: "2025-12-29T01:00:00+07:00"
```

### From Kanban to Orchestrator

```yaml
# WIP limit warning
signal:
  type: wip_limit_warning
  agent: "go-coder-agent"
  current_wip: 3
  max_wip: 3
  message: "Agent at WIP limit"

# WIP limit exceeded (block new tasks)
signal:
  type: wip_limit_exceeded
  agent: "go-coder-agent"
  action: "block_new_tasks"

# Metrics update
signal:
  type: metrics_update
  session_id: "go-team-{uuid}"
  throughput: 5
  avg_cycle_time: 15
  blocked_count: 0
```

---

## Workflow Hooks

### Step Transition Hooks

```yaml
# Add to each step file's execution
on_step_start:
  - action: kanban_update
    signal:
      type: step_started
      step: "${stepNumber}"
      agent: "${agent}"

on_step_complete:
  - action: kanban_update
    signal:
      type: step_completed
      step: "${stepNumber}"
      agent: "${agent}"
      duration: "${elapsed_time}"

on_step_blocked:
  - action: kanban_update
    signal:
      type: task_blocked
      step: "${stepNumber}"
      reason: "${block_reason}"
```

---

## Board Columns Mapping

| Workflow Step | Kanban Column | Agent |
|---------------|---------------|-------|
| Step 01: Init | (no column) | Orchestrator |
| Step 01b: Codebase | (no column) | Orchestrator |
| Step 02: Requirements | Requirements | PM Agent |
| Step 03: Architecture | Architecture | Architect Agent |
| Step 04: Implementation | Development | Go Coder Agent |
| Step 05: Testing | Testing | Test Agent |
| Step 05b: Security | Security Gate | Security Agent |
| Step 06: Review Loop | Review | Reviewer Agent |
| Step 07: Optimization | Optimization | Optimizer Agent |
| Step 08: Release | Release | DevOps Agent |
| Step 09: Synthesis | Done | Orchestrator |

---

## Visual Board Display

```
═══════════════════════════════════════════════════════════════════════════════
                    GO TEAM KANBAN - Session: JWT Auth
                    Started: 00:30 | Elapsed: 15m
═══════════════════════════════════════════════════════════════════════════════
│ REQUIREMENTS │ ARCHITECTURE │ DEVELOPMENT │ TESTING │ SECURITY │ REVIEW    │
│ (PM Agent)   │ (Architect)  │ (Coder)     │ (Test)  │ (Gate)   │ (Reviewer)│
│──────────────│──────────────│─────────────│─────────│──────────│───────────│
│ ✓ Spec done  │ → Designing  │ ○ Pending   │ ○       │ ○        │ ○         │
│   3m         │   @architect │             │         │          │           │
│              │   ⏱ 5m       │             │         │          │           │
│              │              │             │         │          │           │
═══════════════════════════════════════════════════════════════════════════════
Pipeline: ██████░░░░░░░░░░░░░░ 30% | Coverage: N/A | Security: Pending
Controls: *status | *metrics | *agents
═══════════════════════════════════════════════════════════════════════════════
```

---

## WIP Limit Enforcement

### Per-Column Limits

| Column | WIP Limit | Reason |
|--------|-----------|--------|
| Requirements | 1 | Focus on one spec at a time |
| Architecture | 1 | Single design focus |
| Development | 3 | Parallel file creation allowed |
| Testing | 2 | Test + update cycles |
| Security Gate | 1 | Blocking gate |
| Review | 2 | Review + fix cycles |
| Optimization | 1 | Single focus |
| Release | 1 | Single release |

### Enforcement Logic

```python
def can_start_task(agent, column):
    current_wip = count_tasks_in_column(column)
    max_wip = column.wip_limit

    if current_wip >= max_wip:
        send_signal({
            "type": "wip_limit_exceeded",
            "agent": agent,
            "column": column.name,
            "current": current_wip,
            "limit": max_wip
        })
        return False
    return True
```

---

## Metrics Collection

### Per-Session Metrics

```yaml
session_metrics:
  session_id: "go-team-{uuid}"
  started_at: "2025-12-29T00:30:00+07:00"
  completed_at: "2025-12-29T01:00:00+07:00"
  duration_minutes: 30

  steps:
    requirements:
      duration: 180
      agent: pm-agent
      outputs: ["spec.md"]
    architecture:
      duration: 300
      agent: architect-agent
      outputs: ["architecture.md"]
    # ... etc

  agents:
    pm-agent:
      tasks_completed: 1
      time_spent: 180
    architect-agent:
      tasks_completed: 1
      time_spent: 300
    # ... etc

  quality:
    coverage: 85
    lint_clean: true
    race_free: true
    security_passed: true
    review_iterations: 2

  files:
    created: 12
    modified: 3
    lines_added: 1500
    lines_removed: 50
```

### Aggregate Metrics

```yaml
aggregate_metrics:
  total_sessions: 25
  success_rate: 92%
  avg_duration: 28 minutes

  by_workflow:
    full_pipeline:
      count: 15
      avg_duration: 35 minutes
      success_rate: 90%
    quick_fix:
      count: 8
      avg_duration: 12 minutes
      success_rate: 100%

  by_agent:
    pm-agent:
      avg_time: 5 minutes
      tasks: 25
    go-coder-agent:
      avg_time: 12 minutes
      files_created: 180
      lines_written: 15000
    security-agent:
      scans: 20
      vulns_found: 8
      vulns_critical: 1
```

---

## Commands Extension

### New Commands for Go-Team

| Command | Description |
|---------|-------------|
| `*board` | Show go-team kanban board |
| `*board:full` | Show full board with all columns |
| `*board:session` | Show current session board |
| `*metrics:session` | Show current session metrics |
| `*metrics:agents` | Show all agents performance |
| `*metrics:history` | Show historical metrics |

### Usage Examples

```
User: *board

╔═══════════════════════════════════════════════════════════════╗
║              GO TEAM - Current Session                         ║
║              Topic: JWT Authentication                         ║
╠═══════════════════════════════════════════════════════════════╣
║ REQ  │ ARCH │ DEV  │ TEST │ SEC  │ REVIEW │ OPT  │ REL  │ DONE║
║ ✓    │ →    │ ○    │ ○    │ ○    │ ○      │ ○    │ ○    │     ║
╠═══════════════════════════════════════════════════════════════╣
║ Current: Architect Agent designing system                      ║
║ Progress: 22% | Time: 8m | Est. remaining: 25m                ║
╚═══════════════════════════════════════════════════════════════╝
```

---

## Implementation Checklist

### Phase 1: Basic Integration
- [x] Create go-team-board.yaml
- [x] Define column mapping
- [x] Create integration.md

### Phase 2: Signal Handlers
- [x] Add signal emission to orchestrator
- [x] Add signal handling to kanban-agent
- [x] Implement WIP limit checks
- [x] Create signal-emitter.md

### Phase 3: Workflow Hooks
- [x] Add on_step_start hooks
- [x] Add on_step_complete hooks
- [x] Add on_step_blocked hooks
- [x] Document emission points

### Phase 4: Config System
- [x] Create config.yaml
- [x] Create config-loader.md
- [x] Add config commands (*config, *iterations, *coverage)
- [x] Integrate config in orchestrator

### Phase 5: Testing
- [x] Create test-config-signals.sh
- [x] All 53 tests passing

### Phase 6: Metrics & Reporting (TODO)
- [ ] Implement metrics collection hooks
- [ ] Add *board command handler
- [ ] Add *metrics commands handler
- [ ] Create session reports
