# Go Team Signal Emitter

## Overview

Signal emitter module cho go-team workflow. Emit events to Kanban board for real-time tracking.

---

## Signal Emission Protocol

### Signal Structure

```yaml
signal:
  id: "sig-{uuid}"                    # Unique signal ID
  type: "step_started"                # Signal type
  timestamp: "2025-12-29T01:30:00+07:00"
  session_id: "go-team-{session-uuid}"

  # Payload (varies by type)
  payload:
    step: "step-02"
    agent: "pm-agent"
    # ... type-specific data
```

### Signal Types

| Type | Emitted When | Payload |
|------|--------------|---------|
| `session_started` | Session begins | topic, workflow_type |
| `session_completed` | Session ends | result, duration, metrics |
| `step_started` | Step begins | step, agent |
| `step_completed` | Step ends | step, agent, duration, outputs |
| `step_blocked` | Step cannot proceed | step, reason, severity |
| `agent_activated` | Agent starts task | agent, task_id |
| `agent_completed` | Agent finishes | agent, task_id, result |
| `security_gate` | Security scan done | result, severity, issues |
| `review_iteration` | Review loop cycles | iteration, max, issues |
| `breakpoint_hit` | Waiting for approval | step, type |
| `config_changed` | Config modified | key, old_value, new_value |
| `checkpoint_created` | Checkpoint saved | checkpoint_id, step |
| `wip_warning` | Near WIP limit | column, current, limit |
| `wip_exceeded` | WIP limit hit | column, action |
| `error` | Error occurred | step, agent, error_type, message |

---

## Emitter Functions

### emit_signal(type, payload)

```python
def emit_signal(signal_type: str, payload: dict):
    """
    Emit a signal to the Kanban board.

    Args:
        signal_type: Type of signal (step_started, etc.)
        payload: Signal-specific data

    Returns:
        signal_id: Unique identifier for the emitted signal
    """
    # Check if signals enabled
    if not get_config("kanban.enabled"):
        return None

    if not get_config(f"kanban.signals.emit_{signal_type}", True):
        return None

    # Build signal
    signal = {
        "id": f"sig-{generate_uuid()}",
        "type": signal_type,
        "timestamp": now_iso(),
        "session_id": go_team_state.session_id,
        "payload": payload
    }

    # Add to queue
    signal_queue.append(signal)

    # Process immediately if sync mode
    if get_config("kanban.signals.sync_mode", True):
        process_signal(signal)

    # Update last signal
    go_team_state.signals.last_signal = signal

    return signal["id"]
```

### process_signal(signal)

```python
def process_signal(signal: dict):
    """Process a signal and update Kanban board."""

    signal_type = signal["type"]
    payload = signal["payload"]

    # Route to handler
    handlers = {
        "session_started": handle_session_started,
        "session_completed": handle_session_completed,
        "step_started": handle_step_started,
        "step_completed": handle_step_completed,
        "step_blocked": handle_step_blocked,
        "agent_activated": handle_agent_activated,
        "security_gate": handle_security_gate,
        "review_iteration": handle_review_iteration,
        "breakpoint_hit": handle_breakpoint_hit,
    }

    handler = handlers.get(signal_type)
    if handler:
        handler(payload)

    # Update processed count
    go_team_state.signals.processed_count += 1
```

---

## Signal Handlers

### handle_session_started

```python
def handle_session_started(payload):
    """Initialize board for new session."""

    # Update board
    board.session_id = payload["session_id"]
    board.workflow.started_at = payload["timestamp"]
    board.workflow.current_step = "step-01"

    # Create task in backlog
    task = create_task(
        id=f"task-{payload['session_id']}",
        title=payload["topic"],
        type=payload["workflow_type"],
        status="active"
    )
    board.columns.backlog.tasks.append(task)

    # Update metrics
    board.metrics.total_sessions += 1
```

### handle_step_started

```python
def handle_step_started(payload):
    """Move task to appropriate column."""

    step = payload["step"]
    agent = payload["agent"]

    # Get column for step
    column = get_column_for_step(step)
    if not column:
        return

    # Check WIP limit
    if len(column.tasks) >= column.wip_limit:
        emit_signal("wip_exceeded", {
            "column": column.name,
            "current": len(column.tasks),
            "limit": column.wip_limit
        })
        if get_config("kanban.enforce_wip"):
            return  # Block if enforced

    # Create step task
    task = create_task(
        id=f"task-{step}-{go_team_state.session_id}",
        title=f"{step}: {get_step_name(step)}",
        agent=agent,
        started_at=payload["timestamp"],
        status="in_progress"
    )

    # Move to column
    move_task(task, column)

    # Update workflow state
    board.workflow.current_step = step
    board.workflow.current_agent = agent

    # Update agent state
    board.agents[agent].active_tasks.append(task.id)
```

### handle_step_completed

```python
def handle_step_completed(payload):
    """Mark step as done and move task."""

    step = payload["step"]
    agent = payload["agent"]
    duration = payload["duration_seconds"]
    outputs = payload.get("outputs", [])

    # Find task
    task = find_task_by_step(step)
    if not task:
        return

    # Update task
    task.status = "completed"
    task.completed_at = payload["timestamp"]
    task.duration = duration
    task.outputs = outputs

    # Move to next column or done
    current_column = get_column_for_step(step)
    next_step = get_next_step(step)

    if next_step:
        # Keep in current column until next step starts
        pass
    else:
        # Final step - move to done
        move_task(task, board.columns.done)

    # Update agent stats
    board.agents[agent].stats.tasks_completed += 1
    update_agent_avg_time(agent, duration)

    # Remove from active
    board.agents[agent].active_tasks.remove(task.id)

    # Update metrics
    board.metrics.total_tasks_processed += 1
```

### handle_security_gate

```python
def handle_security_gate(payload):
    """Handle security scan results."""

    result = payload["result"]  # PASSED, PASSED_WITH_WARNINGS, BLOCKED
    severity = payload.get("severity")
    issues_count = payload.get("issues_count", 0)

    # Update task
    task = find_task_by_step("step-05b")
    if task:
        task.security_result = result
        task.security_issues = issues_count

    # Handle blocking
    if result == "BLOCKED":
        board.metrics.security_blocks += 1

        # Move task to blocked column
        move_task(task, board.columns.blocked)

        emit_signal("step_blocked", {
            "step": "step-05b",
            "reason": f"Security gate blocked: {severity} severity issues",
            "severity": "high"
        })
```

### handle_review_iteration

```python
def handle_review_iteration(payload):
    """Track review loop progress."""

    iteration = payload["iteration"]
    max_iterations = payload["max_iterations"]
    issues_found = payload["issues_found"]

    # Update workflow state
    board.workflow.iteration = iteration
    board.workflow.max_iterations = max_iterations

    # Update task
    task = find_task_by_step("step-06")
    if task:
        task.iteration = iteration
        task.issues_this_iteration = issues_found

    # Check if nearing max
    if iteration >= max_iterations - 1:
        emit_signal("wip_warning", {
            "type": "iteration_limit",
            "current": iteration,
            "limit": max_iterations,
            "message": f"Review loop at iteration {iteration}/{max_iterations}"
        })
```

### handle_breakpoint_hit

```python
def handle_breakpoint_hit(payload):
    """Record breakpoint for observer approval."""

    step = payload["step"]
    breakpoint_type = payload["type"]

    # Update workflow
    board.workflow.breakpoints_hit.append({
        "step": step,
        "type": breakpoint_type,
        "timestamp": payload["timestamp"]
    })

    # Update current task
    task = find_task_by_step(step)
    if task:
        task.status = "awaiting_approval"
        task.breakpoint_type = breakpoint_type
```

---

## Emission Points in Workflow

### Step 01: Init

```yaml
step_01_init:
  on_start:
    - emit_signal("session_started", {
        "topic": user_topic,
        "workflow_type": determined_workflow,
        "timestamp": now()
      })

  on_complete:
    - emit_signal("step_completed", {
        "step": "step-01",
        "agent": "orchestrator-agent",
        "duration_seconds": elapsed,
        "outputs": ["session_state.json"]
      })
```

### Step 02: Requirements

```yaml
step_02_requirements:
  on_start:
    - emit_signal("step_started", {
        "step": "step-02",
        "step_name": "Requirements",
        "agent": "pm-agent"
      })
    - emit_signal("agent_activated", {
        "agent": "pm-agent",
        "task_id": task_id
      })

  on_breakpoint:
    - emit_signal("breakpoint_hit", {
        "step": "step-02",
        "type": "approval_required"
      })

  on_complete:
    - emit_signal("step_completed", {
        "step": "step-02",
        "agent": "pm-agent",
        "duration_seconds": elapsed,
        "outputs": ["spec.md"]
      })
```

### Step 05b: Security

```yaml
step_05b_security:
  on_start:
    - emit_signal("step_started", {
        "step": "step-05b",
        "step_name": "Security Gate",
        "agent": "security-agent"
      })

  on_scan_complete:
    - emit_signal("security_gate", {
        "result": scan_result,
        "severity": highest_severity,
        "issues_count": len(issues),
        "issues_by_severity": severity_counts
      })

  on_blocked:
    - emit_signal("step_blocked", {
        "step": "step-05b",
        "reason": "Critical/High vulnerabilities found",
        "severity": "critical"
      })
```

### Step 06: Review Loop

```yaml
step_06_review:
  on_start:
    - emit_signal("step_started", {
        "step": "step-06",
        "step_name": "Review Loop",
        "agent": "reviewer-agent"
      })

  on_each_iteration:
    - emit_signal("review_iteration", {
        "iteration": current_iteration,
        "max_iterations": max_iterations,
        "issues_found": len(issues),
        "issues_by_type": issue_types
      })

  on_fixer_activated:
    - emit_signal("agent_activated", {
        "agent": "fixer-agent",
        "task_id": fix_task_id,
        "fix_type": "simple"
      })

  on_coder_activated:
    - emit_signal("agent_activated", {
        "agent": "go-coder-agent",
        "task_id": fix_task_id,
        "fix_type": "complex"
      })

  on_complete:
    - emit_signal("step_completed", {
        "step": "step-06",
        "agent": "reviewer-agent",
        "duration_seconds": elapsed,
        "total_iterations": iterations_used,
        "issues_resolved": total_fixed
      })
```

### Step 09: Synthesis

```yaml
step_09_synthesis:
  on_complete:
    - emit_signal("session_completed", {
        "result": session_result,
        "duration_seconds": total_elapsed,
        "metrics": {
          "coverage": final_coverage,
          "security_passed": security_passed,
          "files_created": len(created_files),
          "files_modified": len(modified_files),
          "review_iterations": total_iterations,
          "agents_used": agents_activated
        }
      })
```

---

## Signal Queue Management

### Queue Structure

```yaml
signal_queue:
  pending: []
  processing: null
  processed: []
  failed: []

  stats:
    total_emitted: 0
    total_processed: 0
    total_failed: 0
    avg_process_time_ms: 0
```

### Queue Operations

```python
def enqueue_signal(signal):
    """Add signal to pending queue."""
    signal_queue.pending.append(signal)
    signal_queue.stats.total_emitted += 1

def process_queue():
    """Process all pending signals."""
    while signal_queue.pending:
        signal = signal_queue.pending.pop(0)
        signal_queue.processing = signal

        try:
            process_signal(signal)
            signal_queue.processed.append(signal)
            signal_queue.stats.total_processed += 1
        except Exception as e:
            signal["error"] = str(e)
            signal_queue.failed.append(signal)
            signal_queue.stats.total_failed += 1

        signal_queue.processing = None

def get_queue_status():
    """Get current queue status."""
    return {
        "pending": len(signal_queue.pending),
        "processing": signal_queue.processing is not None,
        "processed": signal_queue.stats.total_processed,
        "failed": signal_queue.stats.total_failed
    }
```

---

## Board Update Functions

### move_task(task, column)

```python
def move_task(task, target_column):
    """Move task to target column."""
    # Remove from current column
    for col in board.columns.values():
        if task in col.tasks:
            col.tasks.remove(task)
            break

    # Add to target
    target_column.tasks.append(task)

    # Update board timestamp
    board.last_updated = now()
```

### update_board_display()

```python
def update_board_display():
    """Refresh board visual display."""
    display = []

    display.append("=" * 70)
    display.append(f"  GO TEAM KANBAN - {board.session_id}")
    display.append(f"  Started: {board.workflow.started_at}")
    display.append("=" * 70)

    # Column headers
    headers = " | ".join([col.name[:10] for col in board.columns.values()])
    display.append(headers)
    display.append("-" * 70)

    # Tasks per column
    max_tasks = max(len(col.tasks) for col in board.columns.values())
    for i in range(max_tasks):
        row = []
        for col in board.columns.values():
            if i < len(col.tasks):
                task = col.tasks[i]
                status = "→" if task.status == "in_progress" else "✓" if task.status == "completed" else "○"
                row.append(f"{status} {task.title[:8]}")
            else:
                row.append(" " * 10)
        display.append(" | ".join(row))

    display.append("=" * 70)

    return "\n".join(display)
```

---

## Usage in Orchestrator

### Integration Example

```yaml
# In orchestrator-agent.md
workflow_execution:
  before_step:
    - load_config()
    - emit_signal("step_started", step_info)

  after_step:
    - emit_signal("step_completed", step_result)
    - if has_breakpoint:
        emit_signal("breakpoint_hit", breakpoint_info)

  on_error:
    - emit_signal("error", error_info)
    - handle_error()
```

### Config-Driven Emission

```python
# Only emit if configured
if get_config("kanban.signals.emit_step_start"):
    emit_signal("step_started", payload)

# Check specific signal types
if get_config("kanban.signals.emit_agent_activate"):
    emit_signal("agent_activated", payload)
```

---

## Testing Signal Emission

### Test Cases

```yaml
test_signal_emission:
  - name: "Session signals"
    steps:
      - start_session
      - verify: signal_emitted("session_started")
      - complete_session
      - verify: signal_emitted("session_completed")

  - name: "Step signals"
    steps:
      - start_step("step-02")
      - verify: signal_emitted("step_started", step="step-02")
      - complete_step("step-02")
      - verify: signal_emitted("step_completed", step="step-02")

  - name: "Security gate signals"
    steps:
      - run_security_scan(issues=[CRITICAL])
      - verify: signal_emitted("security_gate", result="BLOCKED")
      - verify: signal_emitted("step_blocked")

  - name: "Review iteration signals"
    steps:
      - start_review_loop
      - complete_iteration(1)
      - verify: signal_emitted("review_iteration", iteration=1)
      - complete_iteration(2)
      - verify: signal_emitted("review_iteration", iteration=2)
```
