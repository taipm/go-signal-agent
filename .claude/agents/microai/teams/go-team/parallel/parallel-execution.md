# Parallel Execution System

**Version:** 1.0.0

Cho phép nhiều agents hoạt động đồng thời để tăng tốc workflow.

---

## Quick Start

```bash
# Enable parallel execution
*parallel

# Enable with specific concurrency
*parallel:2    # Max 2 agents concurrent
*parallel:3    # Max 3 agents concurrent (default)
*parallel:max  # All available agents

# Disable parallel
*parallel:off
```

---

## Architecture

```
                    ┌─────────────────────────────────────┐
                    │         PARALLEL SCHEDULER          │
                    │  ┌─────────────────────────────────┐│
                    │  │     Task Dependency Graph        ││
                    │  │                                  ││
                    │  │  Spec ──→ Arch ──→ Code ──┬──→ Review
                    │  │                    ↓      │      │
                    │  │                  Test ────┘      │
                    │  │                    ↓             │
                    │  │                Security ─────────┘
                    │  └─────────────────────────────────┘│
                    └─────────────────────────────────────┘
                              ↓         ↓         ↓
                    ┌─────────┐ ┌─────────┐ ┌─────────┐
                    │ Worker  │ │ Worker  │ │ Worker  │
                    │   #1    │ │   #2    │ │   #3    │
                    │ (Coder) │ │ (Test)  │ │(Security│
                    └─────────┘ └─────────┘ └─────────┘
```

---

## Parallelizable Steps

### Dependency Graph

```yaml
step_dependencies:
  step-01-init:
    depends_on: []
    parallel: false  # Must be sequential

  step-02-requirements:
    depends_on: [step-01-init]
    parallel: false  # Needs observer input

  step-03-architecture:
    depends_on: [step-02-requirements]
    parallel: false  # Single architect

  step-04-implementation:
    depends_on: [step-03-architecture]
    parallel: true   # Can split into sub-tasks
    parallelizable_tasks:
      - models
      - interfaces
      - repositories
      - services
      - handlers
      - main_wiring

  step-05-testing:
    depends_on: [step-04-implementation]
    parallel: true   # Can run alongside step-05b

  step-05b-security:
    depends_on: [step-04-implementation]
    parallel: true   # Can run alongside step-05

  step-06-review:
    depends_on: [step-05-testing, step-05b-security]
    parallel: true   # Reviewer + Coder + Test can work together

  step-07-optimization:
    depends_on: [step-06-review]
    parallel: false  # Sequential for stability

  step-08-release:
    depends_on: [step-07-optimization]
    parallel: false  # Sequential

  step-09-synthesis:
    depends_on: [step-08-release]
    parallel: false  # Final step
```

### Visual Timeline

```
Sequential Mode:
├─ Init ─┼─ Req ─┼─ Arch ─┼─ Code ─┼─ Test ─┼─ Sec ─┼─ Review ─┼─ Opt ─┼─ Rel ─┼─ Syn ─┤
Time: ════════════════════════════════════════════════════════════════════════════►

Parallel Mode:
├─ Init ─┼─ Req ─┼─ Arch ─┼─ Code ───────────────┼─ Review ──┼─ Opt ─┼─ Rel ─┼─ Syn ─┤
                          │ ├─ Test ─────────────┤           │
                          │ └─ Security ─────────┘           │
Time: ═══════════════════════════════════════════════════════════════════►
                                   ↑
                            ~30% faster
```

---

## Parallel Execution Patterns

### Pattern 1: Test + Security Parallel

```yaml
parallel_group_1:
  name: "Quality Assurance"
  trigger: step-04-implementation.complete
  tasks:
    - agent: test-agent
      step: step-05-testing
      priority: high

    - agent: security-agent
      step: step-05b-security
      priority: high

  sync_point: step-06-review
  merge_strategy: wait_all
```

**Benefit:** Security scan runs while tests are being written

### Pattern 2: Implementation Sub-tasks

```yaml
parallel_group_2:
  name: "Code Generation"
  trigger: step-03-architecture.complete
  tasks:
    - task: generate_models
      agent: go-coder-agent
      target: internal/model/

    - task: generate_interfaces
      agent: go-coder-agent
      target: internal/service/interfaces.go

    - task: generate_repos
      agent: go-coder-agent
      target: internal/repo/
      depends_on: [generate_interfaces]

  sync_point: all_code_generated
  merge_strategy: sequential_merge
```

**Benefit:** Parallel code generation for independent components

### Pattern 3: Review Loop Parallel

```yaml
parallel_group_3:
  name: "Review Iteration"
  trigger: reviewer_found_issues
  tasks:
    - task: fix_code_issues
      agent: go-coder-agent
      issues: code_issues

    - task: fix_test_issues
      agent: test-agent
      issues: test_issues
      parallel_with: fix_code_issues

  sync_point: all_fixes_complete
  then: re-run-review
```

**Benefit:** Coder và Test Agent fix issues simultaneously

---

## Task Scheduler

### Scheduling Algorithm

```python
def schedule_tasks(task_queue, available_workers):
    ready_tasks = []

    for task in task_queue:
        if all_dependencies_complete(task):
            ready_tasks.append(task)

    # Sort by priority
    ready_tasks.sort(key=lambda t: t.priority, reverse=True)

    # Assign to workers
    assignments = []
    for task in ready_tasks[:len(available_workers)]:
        worker = get_best_worker(task, available_workers)
        assignments.append((worker, task))
        available_workers.remove(worker)

    return assignments
```

### Worker Pool

```json
{
  "worker_pool": {
    "max_workers": 3,
    "active_workers": [
      {
        "id": "worker-1",
        "agent": "go-coder-agent",
        "task": "generate_handlers",
        "started_at": "2025-12-28T23:10:00Z",
        "progress": 60
      },
      {
        "id": "worker-2",
        "agent": "test-agent",
        "task": "write_unit_tests",
        "started_at": "2025-12-28T23:10:30Z",
        "progress": 40
      }
    ],
    "idle_workers": ["worker-3"],
    "queue_length": 2
  }
}
```

---

## Synchronization

### Sync Points

```yaml
sync_points:
  - name: code_complete
    waits_for: [models, interfaces, repos, services, handlers]
    then: trigger_testing

  - name: quality_complete
    waits_for: [testing, security]
    then: trigger_review

  - name: review_complete
    waits_for: [code_fixes, test_fixes]
    then: next_iteration_or_proceed
```

### Merge Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `wait_all` | Wait for all tasks | Quality gates |
| `wait_any` | Continue on first complete | Early detection |
| `sequential_merge` | Merge in order | Code generation |
| `priority_merge` | Merge by priority | Critical fixes first |

---

## Conflict Resolution

### File Conflicts

```yaml
on_file_conflict:
  detection: both_workers_modify_same_file

  strategies:
    - name: lock_file
      behavior: First worker locks, second waits

    - name: merge_changes
      behavior: Attempt git-style merge

    - name: priority_wins
      behavior: Higher priority task wins

  default: lock_file
```

### State Conflicts

```yaml
on_state_conflict:
  detection: inconsistent_session_state

  resolution:
    1. Pause all workers
    2. Identify conflict source
    3. Choose authoritative state
    4. Resume with consistent state
```

---

## Commands

### Enable/Disable

| Command | Description |
|---------|-------------|
| `*parallel` | Enable parallel (default: 3 workers) |
| `*parallel:N` | Enable with N workers |
| `*parallel:max` | Maximum parallelism |
| `*parallel:off` | Disable parallel |

### Monitoring

| Command | Description |
|---------|-------------|
| `*parallel:status` | Show current parallel status |
| `*parallel:workers` | List active workers |
| `*parallel:queue` | Show task queue |
| `*parallel:sync` | Force sync point |

### Configuration

| Command | Description |
|---------|-------------|
| `*parallel:config` | Show configuration |
| `*parallel:set {key} {value}` | Set config value |

---

## State Management

### Parallel State

```json
{
  "parallel": {
    "enabled": true,
    "max_workers": 3,
    "active_tasks": 2,

    "workers": [
      {
        "id": "worker-1",
        "agent": "go-coder-agent",
        "task": "implement_handlers",
        "status": "running",
        "progress": 75,
        "started_at": "2025-12-28T23:15:00Z"
      },
      {
        "id": "worker-2",
        "agent": "test-agent",
        "task": "write_handler_tests",
        "status": "running",
        "progress": 50,
        "started_at": "2025-12-28T23:16:00Z"
      }
    ],

    "task_queue": [
      {
        "id": "task-3",
        "type": "security_scan",
        "agent": "security-agent",
        "depends_on": ["implement_handlers"],
        "status": "waiting"
      }
    ],

    "sync_points": [
      {
        "name": "code_complete",
        "waiting_for": ["implement_handlers"],
        "completed": ["implement_models", "implement_services"]
      }
    ],

    "statistics": {
      "tasks_completed": 5,
      "avg_parallelism": 2.3,
      "time_saved_percent": 28
    }
  }
}
```

---

## Integration with Workflow

### Workflow State Extension

```yaml
go_team_state:
  # ... existing fields ...

  parallel:
    enabled: true
    max_workers: 3
    current_group: "quality_assurance"
    active_tasks: []
    completed_tasks: []
    pending_sync: "quality_complete"
```

### Step Execution

```yaml
on_step_start:
  if step.parallel and parallel.enabled:
    1. Identify parallelizable sub-tasks
    2. Create task entries
    3. Schedule tasks to workers
    4. Monitor progress
    5. Sync at sync_point
    6. Merge results
    7. Continue to next step
```

---

## Performance Metrics

### Speedup Calculation

```
Sequential Time = Sum(step_durations)
Parallel Time = Max(parallel_group_durations) + Sequential_steps

Speedup = Sequential Time / Parallel Time

Example:
- Code: 10 min
- Test: 8 min
- Security: 5 min
- Review: 5 min

Sequential: 10 + 8 + 5 + 5 = 28 min
Parallel: 10 + max(8, 5) + 5 = 23 min
Speedup: 28/23 = 1.22x (~18% faster)
```

### Metrics Tracked

| Metric | Description |
|--------|-------------|
| `avg_parallelism` | Average concurrent tasks |
| `time_saved_percent` | % time saved vs sequential |
| `worker_utilization` | % time workers are busy |
| `sync_wait_time` | Time waiting at sync points |
| `conflict_count` | Number of conflicts resolved |

---

## Safety Mechanisms

### Deadlock Prevention

```yaml
deadlock_prevention:
  - timeout_per_task: 30_minutes
  - cycle_detection: enabled
  - forced_unlock_after: 3_retries
```

### Resource Limits

```yaml
resource_limits:
  max_concurrent_file_writes: 1
  max_concurrent_bash_commands: 2
  max_memory_per_worker: "512MB"
```

### Rollback Support

```yaml
on_parallel_failure:
  1. Pause all workers
  2. Save partial progress
  3. Rollback to last checkpoint
  4. Re-attempt in sequential mode
```

---

## Visualization

### Progress Display

```
═══════════════════════════════════════════════════════════
⚡ PARALLEL EXECUTION - Status
═══════════════════════════════════════════════════════════

Workers: 3 active / 3 max

┌─────────────────────────────────────────────────────────┐
│ Worker 1 [go-coder-agent]                               │
│ Task: implement_handlers                                │
│ Progress: ████████████████████░░░░░░░░░░ 75%           │
│ Time: 3m 45s                                            │
├─────────────────────────────────────────────────────────┤
│ Worker 2 [test-agent]                                   │
│ Task: write_handler_tests                               │
│ Progress: ████████████░░░░░░░░░░░░░░░░░░ 50%           │
│ Time: 2m 30s                                            │
├─────────────────────────────────────────────────────────┤
│ Worker 3 [security-agent]                               │
│ Task: (waiting for dependencies)                        │
│ Queue position: 1                                       │
└─────────────────────────────────────────────────────────┘

Next Sync Point: quality_complete
Waiting for: implement_handlers

Time Saved (est): 18%

═══════════════════════════════════════════════════════════
```

---

## Benefits

1. **Faster Execution:** 20-30% time reduction
2. **Better Resource Utilization:** Multiple agents work simultaneously
3. **Independent Work Streams:** Test and Security don't block each other
4. **Maintained Quality:** Same quality gates apply
5. **Flexible Configuration:** Adjust parallelism as needed

---

## Limitations

1. **Sequential Bottlenecks:** Some steps must be sequential
2. **Conflict Overhead:** Coordination has some cost
3. **Complexity:** More complex state management
4. **Debugging:** Harder to trace parallel issues
