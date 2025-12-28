#!/bin/bash
# .claude/kanban/hooks/signal-detector.sh
# PreToolUse hook: Starts task in kanban when agent is invoked
#
# REAL-TIME SIGNAL PROCESSING - Actually updates board.yaml!

set -e

BOARD_FILE=".claude/kanban/board.yaml"
SIGNAL_LOG=".claude/kanban/signals.log"
PROJECT_ROOT="/Users/taipm/GitHub/go-signal-agent"

# Read tool input from stdin (JSON from Claude Code)
INPUT=$(cat)

# Extract tool name and agent info using jq
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)
AGENT_NAME=$(echo "$INPUT" | jq -r '.tool_input.subagent_type // empty' 2>/dev/null)
TASK_DESC=$(echo "$INPUT" | jq -r '.tool_input.description // empty' 2>/dev/null)
TASK_PROMPT=$(echo "$INPUT" | jq -r '.tool_input.prompt // empty' 2>/dev/null | head -c 200)

# Only process Task tool calls with agent names (skip kanban-agent itself to avoid loops)
if [[ "$TOOL_NAME" != "Task" ]] || [[ -z "$AGENT_NAME" ]] || [[ "$AGENT_NAME" == "kanban-agent" ]]; then
    exit 0
fi

# Skip non-work agents (explorers, guides, etc.)
if [[ "$AGENT_NAME" == "Explore" ]] || [[ "$AGENT_NAME" == "claude-code-guide" ]] || [[ "$AGENT_NAME" == "Plan" ]]; then
    exit 0
fi

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
TIMESTAMP_LOCAL=$(date +"%Y-%m-%dT%H:%M:%S%z")
TIMESTAMP_MS=$(date +%s%3N 2>/dev/null || date +%s)

# Create signal log entry
SIGNAL_ENTRY=$(cat <<EOF
{
  "timestamp": "$TIMESTAMP",
  "signal": "agent_activated",
  "agent": "$AGENT_NAME",
  "task_id": "task-$TIMESTAMP_MS",
  "description": "$TASK_DESC",
  "prompt_preview": "$TASK_PROMPT"
}
EOF
)
echo "$SIGNAL_ENTRY" >> "$PROJECT_ROOT/$SIGNAL_LOG"

# ============================================================
# REAL-TIME BOARD UPDATE - Start first backlog task
# ============================================================

BOARD_PATH="$PROJECT_ROOT/$BOARD_FILE"

if [[ ! -f "$BOARD_PATH" ]]; then
    echo "[KANBAN] Agent activated: $AGENT_NAME - $TASK_DESC"
    exit 0
fi

# Use Python for reliable YAML manipulation
python3 << PYTHON_SCRIPT
import yaml
from datetime import datetime

board_path = "$BOARD_PATH"
agent_name = "$AGENT_NAME"
task_desc = "$TASK_DESC"
timestamp = "$TIMESTAMP_LOCAL"

# Read board
with open(board_path, 'r') as f:
    board = yaml.safe_load(f)

# Check WIP limit
in_progress = board.get('columns', {}).get('in_progress', {}).get('tasks', [])
if in_progress is None:
    in_progress = []
wip_limit = board.get('columns', {}).get('in_progress', {}).get('wip_limit', 3)

if len(in_progress) >= wip_limit:
    print(f"[KANBAN] ⚠️ WIP limit reached ({len(in_progress)}/{wip_limit}). Task not auto-started.")
    exit(0)

# Get first task from backlog
backlog = board.get('columns', {}).get('backlog', {}).get('tasks', [])
if not backlog:
    print(f"[KANBAN] No tasks in backlog to start")
    exit(0)

# Move first task to in_progress
task_to_start = backlog.pop(0)
task_id = task_to_start.get('id', 'unknown')

# Update task
task_to_start['started_at'] = timestamp
task_to_start['agent'] = agent_name

# Add to in_progress
in_progress.append(task_to_start)
board['columns']['in_progress']['tasks'] = in_progress

# Update metrics
metrics = board.get('metrics', {})
metrics['current_wip'] = len(in_progress)
board['metrics'] = metrics

# Update last_updated
board['board']['last_updated'] = timestamp

# Update agent active_tasks
if agent_name in board.get('agents', {}):
    active = board['agents'][agent_name].get('active_tasks', [])
    if active is None:
        active = []
    active.append(task_id)
    board['agents'][agent_name]['active_tasks'] = active

# Add history entry
history = board.get('history', [])
history.append({
    'timestamp': timestamp,
    'action': 'task_started',
    'task_id': task_id,
    'by': agent_name
})
board['history'] = history[-50:]  # Keep last 50 entries

# Write back
with open(board_path, 'w') as f:
    yaml.dump(board, f, default_flow_style=False, allow_unicode=True, sort_keys=False)

print(f"[KANBAN] ▶️ Task {task_id} started by {agent_name}")
PYTHON_SCRIPT

exit 0
