#!/bin/bash
# .claude/kanban/hooks/on-task-complete.sh
# PostToolUse hook: Updates kanban board when Task tool completes
#
# REAL-TIME SIGNAL PROCESSING - Actually updates board.yaml!

set -e

BOARD_FILE=".claude/kanban/board.yaml"
SIGNAL_LOG=".claude/kanban/signals.log"
PROJECT_ROOT="/Users/taipm/GitHub/go-signal-agent"

# Read tool output from stdin (JSON from Claude Code)
INPUT=$(cat)

# Extract tool info using jq
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)

# Only process Task tool completions
if [[ "$TOOL_NAME" != "Task" ]]; then
    exit 0
fi

TOOL_OUTPUT=$(echo "$INPUT" | jq -r '.tool_output // empty' 2>/dev/null | head -c 500)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
TIMESTAMP_LOCAL=$(date +"%Y-%m-%dT%H:%M:%S%z")

# Determine success/failure from output
SUCCESS="true"
if echo "$TOOL_OUTPUT" | grep -qiE "(error|failed|exception|panic)"; then
    SUCCESS="false"
fi

# Log the signal
SIGNAL_ENTRY=$(cat <<EOF
{
  "timestamp": "$TIMESTAMP",
  "signal": "task_completed",
  "success": $SUCCESS,
  "output_preview": "${TOOL_OUTPUT:0:200}"
}
EOF
)
echo "$SIGNAL_ENTRY" >> "$PROJECT_ROOT/$SIGNAL_LOG"

# ============================================================
# REAL-TIME BOARD UPDATE
# ============================================================

BOARD_PATH="$PROJECT_ROOT/$BOARD_FILE"

if [[ ! -f "$BOARD_PATH" ]]; then
    echo "[KANBAN] Board file not found: $BOARD_PATH"
    exit 0
fi

# Check if there are tasks in in_progress column
IN_PROGRESS_TASKS=$(grep -A 100 "^  in_progress:" "$BOARD_PATH" | grep -B 1 "    - id:" | head -20 2>/dev/null || true)

if [[ -z "$IN_PROGRESS_TASKS" ]]; then
    echo "[KANBAN] No tasks in progress to complete"
    exit 0
fi

# Get the first in-progress task ID
TASK_ID=$(grep -A 100 "^  in_progress:" "$BOARD_PATH" | grep "      id:" | head -1 | sed 's/.*id: *"\([^"]*\)".*/\1/' 2>/dev/null || true)

if [[ -z "$TASK_ID" ]]; then
    echo "[KANBAN] Could not find task ID in progress"
    exit 0
fi

# Get task details before moving
TASK_TITLE=$(grep -A 100 "^  in_progress:" "$BOARD_PATH" | grep -A 1 "id: \"$TASK_ID\"" | grep "title:" | head -1 | sed 's/.*title: *"\([^"]*\)".*/\1/' 2>/dev/null || true)

# Use Python for reliable YAML manipulation
python3 << PYTHON_SCRIPT
import yaml
from datetime import datetime

board_path = "$BOARD_PATH"
task_id = "$TASK_ID"
timestamp = "$TIMESTAMP_LOCAL"
success = $SUCCESS

# Read board
with open(board_path, 'r') as f:
    board = yaml.safe_load(f)

# Find task in in_progress
in_progress = board.get('columns', {}).get('in_progress', {}).get('tasks', [])
task_to_move = None
task_index = None

for i, task in enumerate(in_progress):
    if task.get('id') == task_id:
        task_to_move = task
        task_index = i
        break

if task_to_move is None:
    print(f"[KANBAN] Task {task_id} not found in in_progress")
    exit(0)

# Remove from in_progress
in_progress.pop(task_index)

# Add completion timestamp
task_to_move['completed_at'] = timestamp

# Calculate cycle time if started_at exists
if 'started_at' in task_to_move and task_to_move['started_at']:
    try:
        # Simple calculation - just note it was completed
        task_to_move['cycle_time'] = "completed"
    except:
        pass

# Add to done column
done = board.get('columns', {}).get('done', {}).get('tasks', [])
if done is None:
    done = []
done.insert(0, task_to_move)  # Add to top of done
board['columns']['done']['tasks'] = done

# Update metrics
metrics = board.get('metrics', {})
metrics['total_tasks_completed'] = metrics.get('total_tasks_completed', 0) + 1
metrics['current_wip'] = len(in_progress)
board['metrics'] = metrics

# Update last_updated
board['board']['last_updated'] = timestamp

# Update agent stats if agent assigned
agent = task_to_move.get('agent', 'go-dev-agent')
if agent and agent in board.get('agents', {}):
    agent_stats = board['agents'][agent].get('stats', {})
    agent_stats['completed_total'] = agent_stats.get('completed_total', 0) + 1
    agent_stats['completed_today'] = agent_stats.get('completed_today', 0) + 1
    board['agents'][agent]['stats'] = agent_stats
    # Remove from active_tasks
    active = board['agents'][agent].get('active_tasks', [])
    if task_id in active:
        active.remove(task_id)
    board['agents'][agent]['active_tasks'] = active

# Add history entry
history = board.get('history', [])
history.append({
    'timestamp': timestamp,
    'action': 'task_completed',
    'task_id': task_id,
    'by': agent or 'system'
})
board['history'] = history[-50:]  # Keep last 50 entries

# Write back
with open(board_path, 'w') as f:
    yaml.dump(board, f, default_flow_style=False, allow_unicode=True, sort_keys=False)

print(f"[KANBAN] ✅ Task {task_id} moved to DONE")
PYTHON_SCRIPT

exit 0
