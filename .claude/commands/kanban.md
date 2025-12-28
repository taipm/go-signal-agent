---
name: kanban
description: Invoke Kanban board manager for task tracking and visualization
argument-hint: "[show|add|start|done|block|metrics|agent|archive] [args]"
---

# Kanban Board Manager

Manage your project's Kanban board with visual task tracking.

## Usage

```
/kanban              # Show board (default)
/kanban show         # Display current board
/kanban add <title>  # Add task to backlog
/kanban start <id>   # Start working on task
/kanban done <id>    # Mark task complete
/kanban block <id>   # Mark task blocked
/kanban metrics      # Show productivity stats
/kanban agent <name> # Show agent's tasks
/kanban archive      # Archive completed tasks
```

## Arguments

$ARGUMENTS

## Execution

Invoke the kanban-agent to execute the requested command:

1. **Load Board State**
   - Read `.claude/kanban/board.yaml`
   - Parse current columns and tasks

2. **Execute Command**
   - Based on arguments, perform the requested action
   - Update board state accordingly

3. **Display Results**
   - Show visual board or confirmation
   - Suggest next actions if appropriate

## Command Details

### show (default)
Display the full Kanban board with all columns and tasks.
Shows WIP limits and cycle times for in-progress items.

### add <title>
Add a new task to the backlog.
- Generates unique task ID
- Sets created_at timestamp
- Default priority: medium

### start <task-id> [agent]
Move a task from backlog to in-progress.
- Checks WIP limit before starting
- Assigns to specified agent (default: go-dev-agent)
- Sets started_at timestamp

### done <task-id>
Move a task from in-progress to done.
- Sets completed_at timestamp
- Calculates cycle time
- Updates agent statistics

### block <task-id> [reason]
Mark a task as blocked.
- Moves to blocked column
- Records blocking reason

### metrics
Show productivity metrics:
- Total tasks by column
- Completion rate
- Average cycle time
- Agent performance

### agent <agent-name>
Show tasks for a specific agent:
- Active tasks and WIP status
- Completed today
- Average cycle time

### archive
Archive completed tasks:
- Snapshot board to archive folder
- Clear done column
- Reset daily counters
