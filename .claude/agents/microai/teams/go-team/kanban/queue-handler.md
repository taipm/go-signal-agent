# Signal Queue Handler

## Overview

Semi-automatic signal handling cho Go Team Kanban integration. Signals được emit tự động khi agents hoạt động, board được cập nhật khi user request.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     GO TEAM SESSION                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐                      ┌──────────────────────┐ │
│  │ Orchestrator │──emit_signal()──────→│   signal-queue.json  │ │
│  └──────┬───────┘                      │   (persistent store) │ │
│         │                              └──────────┬───────────┘ │
│         │ activates                               │             │
│         ↓                                         │             │
│  ┌──────────────┐                                 │             │
│  │   Agents     │                                 │             │
│  │ PM→Arch→...  │                                 │             │
│  └──────────────┘                                 │             │
│                                                   │             │
│  ┌──────────────────────────────────────────────┐ │             │
│  │ User Commands                                 │ │             │
│  │ *board  ──────────────────────────────────────┼─┘             │
│  │ *status ──────── read & display ──────────────┘              │
│  └──────────────────────────────────────────────┘               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Signal Queue Operations

### 1. emit_signal() - Ghi signal vào queue

```python
def emit_signal(signal_type: str, payload: dict):
    """
    Emit signal to queue (auto-called by orchestrator).

    Args:
        signal_type: Type of signal
        payload: Signal data
    """
    # Load queue
    queue = load_json("kanban/signal-queue.json")

    # Create signal
    signal = {
        "id": f"sig-{generate_id()}",
        "type": signal_type,
        "timestamp": now_iso(),
        "payload": payload,
        "processed": False
    }

    # Add to pending
    queue["queue"]["pending"].append(signal)
    queue["metrics"]["signals_emitted"] += 1
    queue["last_updated"] = now_iso()

    # Update board state based on signal type
    update_board_state(queue, signal)

    # Save
    save_json("kanban/signal-queue.json", queue)

    return signal["id"]
```

### 2. update_board_state() - Cập nhật board từ signal

```python
def update_board_state(queue: dict, signal: dict):
    """Update board state based on signal type."""

    signal_type = signal["type"]
    payload = signal["payload"]
    board = queue["board_state"]

    if signal_type == "session_started":
        queue["session"]["id"] = payload.get("session_id")
        queue["session"]["topic"] = payload.get("topic")
        queue["session"]["started_at"] = signal["timestamp"]
        # Add task to backlog
        board["columns"]["backlog"]["tasks"].append({
            "id": f"task-{queue['session']['id']}",
            "title": payload.get("topic"),
            "status": "active"
        })
        board["columns"]["backlog"]["wip"] = 1

    elif signal_type == "step_started":
        step = payload.get("step")
        agent = payload.get("agent")
        queue["session"]["current_step"] = step
        queue["session"]["current_agent"] = agent

        # Map step to column
        column = step_to_column(step)
        if column:
            # Move task to column
            move_task_to_column(board, column)

    elif signal_type == "step_completed":
        step = payload.get("step")
        duration = payload.get("duration_seconds", 0)

        queue["metrics"]["steps_completed"] += 1
        queue["metrics"]["durations"][step] = duration

        # Update progress
        board["progress_percent"] = int(
            (queue["metrics"]["steps_completed"] / queue["metrics"]["total_steps"]) * 100
        )

        # Mark column task as done
        column = step_to_column(step)
        if column:
            mark_column_done(board, column)

    elif signal_type == "agent_activated":
        agent = payload.get("agent")
        if agent not in queue["metrics"]["agents_activated"]:
            queue["metrics"]["agents_activated"].append(agent)

    elif signal_type == "security_gate":
        result = payload.get("result")
        if result == "BLOCKED":
            # Move to blocked column
            board["columns"]["blocked"]["tasks"].append({
                "id": "security-block",
                "title": f"Security: {payload.get('severity')}",
                "status": "blocked"
            })
            board["columns"]["blocked"]["wip"] += 1

    elif signal_type == "session_completed":
        result = payload.get("result")
        # Move all to done
        clear_all_columns(board)
        board["columns"]["done"]["tasks"].append({
            "id": queue["session"]["id"],
            "title": queue["session"]["topic"],
            "status": result
        })
        board["progress_percent"] = 100

    # Update total WIP
    board["total_wip"] = sum(
        col["wip"] for col in board["columns"].values()
    )
```

### 3. step_to_column() - Map step to kanban column

```python
def step_to_column(step: str) -> str:
    """Map workflow step to kanban column."""
    mapping = {
        "step-01": None,           # Init - no column
        "step-01b": None,          # Codebase analysis - no column
        "step-02": "requirements",
        "step-03": "architecture",
        "step-04": "development",
        "step-05": "testing",
        "step-05b": "security",
        "step-06": "review",
        "step-07": "optimization",
        "step-08": "release",
        "step-09": None            # Synthesis - moves to done
    }
    return mapping.get(step)
```

---

## Board Display Commands

### *board - Display current board

```python
def cmd_board():
    """Display kanban board from queue state."""
    queue = load_json("kanban/signal-queue.json")
    session = queue["session"]
    board = queue["board_state"]
    metrics = queue["metrics"]

    if not session["id"]:
        return "No active session. Start with /go-team"

    output = []
    output.append("═" * 70)
    output.append(f"  GO TEAM KANBAN - {session['topic']}")
    output.append(f"  Session: {session['id']}")
    output.append(f"  Started: {session['started_at']}")
    output.append("═" * 70)

    # Column headers
    cols = ["REQ", "ARCH", "DEV", "TEST", "SEC", "REVIEW", "OPT", "REL", "DONE"]
    output.append(" │ ".join(f"{c:^8}" for c in cols))
    output.append("─" * 70)

    # Column status
    status_map = {
        "requirements": get_col_status(board, "requirements"),
        "architecture": get_col_status(board, "architecture"),
        "development": get_col_status(board, "development"),
        "testing": get_col_status(board, "testing"),
        "security": get_col_status(board, "security"),
        "review": get_col_status(board, "review"),
        "optimization": get_col_status(board, "optimization"),
        "release": get_col_status(board, "release"),
        "done": get_col_status(board, "done")
    }

    statuses = [status_map[k] for k in [
        "requirements", "architecture", "development", "testing",
        "security", "review", "optimization", "release", "done"
    ]]
    output.append(" │ ".join(f"{s:^8}" for s in statuses))

    output.append("═" * 70)

    # Progress bar
    pct = board["progress_percent"]
    filled = int(pct / 5)
    bar = "█" * filled + "░" * (20 - filled)
    output.append(f"Progress: {bar} {pct}%")

    # Current status
    if session["current_agent"]:
        output.append(f"Current: {session['current_agent']} @ {session['current_step']}")

    # Metrics
    output.append(f"Steps: {metrics['steps_completed']}/{metrics['total_steps']}")
    output.append(f"Agents: {', '.join(metrics['agents_activated']) or 'None'}")

    output.append("═" * 70)
    output.append("Commands: *board:full | *status | *metrics")

    return "\n".join(output)

def get_col_status(board: dict, column: str) -> str:
    """Get display status for column."""
    col = board["columns"].get(column, {})
    tasks = col.get("tasks", [])

    if not tasks:
        return "○"

    task = tasks[-1]  # Most recent task
    status = task.get("status", "")

    if status == "active" or status == "in_progress":
        return "→"
    elif status == "completed" or status == "done":
        return "✓"
    elif status == "blocked":
        return "✗"
    else:
        return "○"
```

### *status - Quick status

```python
def cmd_status():
    """Quick status display."""
    queue = load_json("kanban/signal-queue.json")
    session = queue["session"]
    metrics = queue["metrics"]

    if not session["id"]:
        return "No active session."

    return f"""
───────────────────────────────────────
GO TEAM STATUS
───────────────────────────────────────
Topic:    {session['topic']}
Step:     {session['current_step']}
Agent:    {session['current_agent']}
Progress: {queue['board_state']['progress_percent']}%
───────────────────────────────────────
"""
```

### *metrics - Session metrics

```python
def cmd_metrics():
    """Display session metrics."""
    queue = load_json("kanban/signal-queue.json")
    metrics = queue["metrics"]

    output = []
    output.append("═══════════════════════════════════════")
    output.append("  SESSION METRICS")
    output.append("═══════════════════════════════════════")
    output.append(f"Signals Emitted:   {metrics['signals_emitted']}")
    output.append(f"Signals Processed: {metrics['signals_processed']}")
    output.append(f"Steps Completed:   {metrics['steps_completed']}/{metrics['total_steps']}")
    output.append("")
    output.append("Agents Activated:")
    for agent in metrics['agents_activated']:
        duration = metrics['durations'].get(agent, "N/A")
        output.append(f"  • {agent}: {duration}")
    output.append("")
    output.append("Step Durations:")
    for step, duration in metrics['durations'].items():
        output.append(f"  • {step}: {duration}s")
    output.append("═══════════════════════════════════════")

    return "\n".join(output)
```

---

## Auto-Emit Integration

### Orchestrator Hooks

Thêm vào orchestrator workflow execution:

```yaml
# In workflow execution
workflow_hooks:
  on_session_start:
    - action: emit_signal
      type: "session_started"
      payload:
        session_id: "${session_id}"
        topic: "${topic}"
        workflow: "${workflow_type}"

  before_step:
    - action: emit_signal
      type: "step_started"
      payload:
        step: "${step_id}"
        step_name: "${step_name}"
        agent: "${agent_name}"

  after_step:
    - action: emit_signal
      type: "step_completed"
      payload:
        step: "${step_id}"
        agent: "${agent_name}"
        duration_seconds: "${elapsed}"

  on_agent_activate:
    - action: emit_signal
      type: "agent_activated"
      payload:
        agent: "${agent_name}"

  on_session_end:
    - action: emit_signal
      type: "session_completed"
      payload:
        result: "${result}"
        metrics: "${quality_metrics}"
```

### Step File Integration

Mỗi step file gọi emit ở đầu và cuối:

```markdown
## Step Execution

### On Start
\`\`\`
emit_signal("step_started", {
  step: "step-02",
  step_name: "Requirements",
  agent: "pm-agent"
})
\`\`\`

### On Complete
\`\`\`
emit_signal("step_completed", {
  step: "step-02",
  agent: "pm-agent",
  duration_seconds: elapsed_time,
  outputs: ["spec.md"]
})
\`\`\`
```

---

## Reset Queue

```python
def reset_queue():
    """Reset queue to initial state."""
    initial = {
        "queue": {"pending": [], "processed": []},
        "session": {
            "id": None, "topic": None,
            "started_at": None, "current_step": None, "current_agent": None
        },
        "board_state": {
            "columns": {
                col: {"tasks": [], "wip": 0}
                for col in ["backlog", "requirements", "architecture",
                           "development", "testing", "security", "review",
                           "optimization", "release", "blocked", "done"]
            },
            "progress_percent": 0,
            "total_wip": 0
        },
        "metrics": {
            "signals_emitted": 0, "signals_processed": 0,
            "steps_completed": 0, "total_steps": 9,
            "agents_activated": [], "durations": {}
        },
        "last_updated": None
    }
    save_json("kanban/signal-queue.json", initial)
    return "Queue reset."
```

---

## Usage Example

```
# 1. User starts session
User: /go-team
User: Add JWT authentication

# 2. Orchestrator auto-emits
emit_signal("session_started", {topic: "JWT authentication", ...})

# 3. PM Agent starts
emit_signal("step_started", {step: "step-02", agent: "pm-agent"})
emit_signal("agent_activated", {agent: "pm-agent"})

# 4. User checks board anytime
User: *board

# Output:
═══════════════════════════════════════════════════════════════════════
  GO TEAM KANBAN - JWT authentication
  Session: go-team-abc123
  Started: 2025-12-29T02:00:00
═══════════════════════════════════════════════════════════════════════
    REQ   │  ARCH  │  DEV   │  TEST  │  SEC   │ REVIEW │  OPT   │  REL   │  DONE
──────────────────────────────────────────────────────────────────────
    →     │   ○    │   ○    │   ○    │   ○    │   ○    │   ○    │   ○    │
═══════════════════════════════════════════════════════════════════════
Progress: ██░░░░░░░░░░░░░░░░░░ 11%
Current: pm-agent @ step-02
Steps: 1/9
Agents: pm-agent
═══════════════════════════════════════════════════════════════════════

# 5. After PM completes
emit_signal("step_completed", {step: "step-02", duration: 180})

# 6. Architect starts
emit_signal("step_started", {step: "step-03", agent: "architect-agent"})

# 7. User checks again
User: *board

# Board now shows:
    REQ   │  ARCH  │  DEV   │  ...
    ✓     │   →    │   ○    │  ...
```

---

## Files

| File | Purpose |
|------|---------|
| `signal-queue.json` | Persistent signal storage |
| `queue-handler.md` | Handler documentation |
| `signal-emitter.md` | Emission protocol |
| `integration.md` | Overall integration guide |
