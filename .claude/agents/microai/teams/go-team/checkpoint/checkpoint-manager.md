---
name: checkpoint-manager
description: Checkpoint/Rollback mechanism for Go Team workflow
version: 1.0.0
---

# Checkpoint Manager

**Mục đích:** Quản lý checkpoints và rollback cho Go Team workflow, cho phép khôi phục về trạng thái trước đó khi có lỗi hoặc cần thay đổi hướng đi.

---

## ARCHITECTURE

```
┌─────────────────────────────────────────────────────────────┐
│                    CHECKPOINT MANAGER                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Checkpoint  │  │  Rollback   │  │   History   │         │
│  │   Creator   │  │   Engine    │  │   Tracker   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         │                │                │                 │
│         └────────────────┼────────────────┘                 │
│                          ▼                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              CHECKPOINT STORAGE                      │   │
│  │  ./checkpoints/{session-id}/                        │   │
│  │    ├── cp-01-init.json                              │   │
│  │    ├── cp-02-requirements.json                      │   │
│  │    ├── cp-03-architecture.json                      │   │
│  │    ├── cp-04-implementation.json                    │   │
│  │    ├── cp-05-testing.json                           │   │
│  │    ├── cp-06-review-{iteration}.json                │   │
│  │    ├── cp-07-optimization.json                      │   │
│  │    ├── cp-08-release.json                           │   │
│  │    └── manifest.json                                │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## CHECKPOINT DATA STRUCTURE

### Checkpoint Schema

```yaml
checkpoint:
  id: "cp-{step}-{timestamp}"
  session_id: "{uuid}"
  step_number: 1-9
  step_name: "init|requirements|architecture|..."
  created_at: "ISO-8601 timestamp"

  # State snapshot
  state:
    topic: ""
    phase: ""
    current_step: 1
    iteration_count: 0
    metrics:
      build_pass: false
      test_coverage: 0
      lint_clean: false
      race_free: false

  # Outputs at this point
  outputs:
    spec: null | "{content}"
    architecture: null | "{content}"
    code_files: []
    test_files: []
    review_comments: []

  # Git state (if applicable)
  git:
    enabled: true
    branch: "go-team/{session-id}"
    commit_hash: "{sha}"
    staged_files: []

  # File system state
  files:
    created: []
    modified: []
    deleted: []

  # Validation
  checksum: "sha256:{hash}"
  valid: true
```

### Manifest Schema

```yaml
manifest:
  session_id: "{uuid}"
  topic: "{project name}"
  created_at: "ISO-8601"
  updated_at: "ISO-8601"

  checkpoints:
    - id: "cp-01-init"
      step: 1
      timestamp: "..."
      valid: true
    - id: "cp-02-requirements"
      step: 2
      timestamp: "..."
      valid: true
    # ...

  current_checkpoint: "cp-03-architecture"
  rollback_history: []
```

---

## CHECKPOINT OPERATIONS

### 1. CREATE CHECKPOINT

Tự động tạo checkpoint sau mỗi step hoàn thành:

```markdown
## Create Checkpoint Protocol

TRIGGER: After each step completion (before proceeding to next step)

ACTIONS:
1. Capture current state
   - Read session state from workflow
   - Collect all outputs generated
   - Snapshot file system changes

2. Create git checkpoint (if enabled)
   - Stage all changes: `git add -A`
   - Commit: `git commit -m "checkpoint: step-{N} - {step_name}"`
   - Store commit hash

3. Write checkpoint file
   - Generate checkpoint ID: cp-{step}-{timestamp}
   - Serialize state to JSON
   - Calculate checksum
   - Save to ./checkpoints/{session-id}/

4. Update manifest
   - Add checkpoint to list
   - Update current_checkpoint
   - Save manifest.json

5. Confirm to observer
   ```
   ✓ Checkpoint created: cp-{step}-{step_name}
     Files: {count} | Git: {commit_short}
   ```
```

### 2. LIST CHECKPOINTS

```markdown
## List Checkpoints Command

TRIGGER: Observer enters `*checkpoints` or `*cp-list`

OUTPUT:
```
=== CHECKPOINTS ===

Session: {session-id}
Topic: {topic}

Available checkpoints:
┌────┬─────────────────────┬─────────────────────┬────────┐
│ #  │ Checkpoint          │ Timestamp           │ Status │
├────┼─────────────────────┼─────────────────────┼────────┤
│ 1  │ cp-01-init          │ 2025-12-28 21:00:00│ ✓      │
│ 2  │ cp-02-requirements  │ 2025-12-28 21:15:00│ ✓      │
│ 3  │ cp-03-architecture  │ 2025-12-28 21:30:00│ ✓      │
│ 4  │ cp-04-implementation│ 2025-12-28 21:45:00│ ← curr │
└────┴─────────────────────┴─────────────────────┴────────┘

Commands:
- *rollback:N     → Rollback to checkpoint N
- *cp-diff:N      → Show diff from checkpoint N to current
- *cp-show:N      → Show checkpoint N details
```
```

### 3. ROLLBACK

```markdown
## Rollback Protocol

TRIGGER: Observer enters `*rollback:{step}` or `*rollback:cp-{id}`

VALIDATION:
1. Check checkpoint exists and is valid
2. Confirm with observer:
   ```
   ⚠️  ROLLBACK CONFIRMATION

   Rolling back to: cp-{step}-{name}

   This will:
   - Restore state to step {N}
   - Revert {X} files
   - Discard steps {N+1} to {current}

   Git changes:
   - Reset to commit: {sha_short}
   - Discard {Y} commits

   Continue? [y/N]
   ```

EXECUTION:
1. Git rollback (if enabled)
   ```bash
   git checkout {branch}
   git reset --hard {checkpoint_commit}
   ```

2. Restore state
   - Load checkpoint state
   - Update session state
   - Set current_step to checkpoint step

3. Restore outputs
   - Clear outputs after checkpoint
   - Load checkpoint outputs

4. Update manifest
   - Record rollback in history
   - Update current_checkpoint

5. Confirm completion
   ```
   ✓ Rollback complete

   Restored to: cp-{step}-{name}
   Current step: {step}

   Ready to continue from step {step+1}.
   Press [Enter] to proceed or *pause to review.
   ```

ROLLBACK HISTORY:
- Record each rollback for audit
- Include: from_checkpoint, to_checkpoint, timestamp, reason
```

### 4. DIFF BETWEEN CHECKPOINTS

```markdown
## Checkpoint Diff Command

TRIGGER: *cp-diff:{from}:{to} or *cp-diff:{N} (diff from N to current)

OUTPUT:
```
=== CHECKPOINT DIFF ===

From: cp-02-requirements
To:   cp-04-implementation (current)

STATE CHANGES:
- phase: requirements → implementation
- current_step: 2 → 4
- iteration_count: 0 → 0

OUTPUT CHANGES:
+ spec: (added - 45 lines)
+ architecture: (added - 120 lines)
+ code_files: 8 files added

FILE CHANGES:
+ internal/handler/handler.go (new)
+ internal/service/service.go (new)
+ internal/repo/repo.go (new)
~ go.mod (modified)

GIT DIFF:
3 commits between checkpoints
- abc1234: checkpoint: step-02
- def5678: checkpoint: step-03
- ghi9012: checkpoint: step-04
```
```

---

## AUTOMATIC CHECKPOINT TRIGGERS

| Step | Trigger Point | Checkpoint ID |
|------|---------------|---------------|
| 1 | After init complete | cp-01-init |
| 2 | After spec approved at breakpoint | cp-02-requirements |
| 3 | After architecture approved at breakpoint | cp-03-architecture |
| 4 | After all code files generated | cp-04-implementation |
| 5 | After tests written and passing | cp-05-testing |
| 6 | After each review iteration | cp-06-review-{N} |
| 6 | After review loop complete | cp-06-review-final |
| 7 | After optimization complete | cp-07-optimization |
| 8 | After release config created | cp-08-release |
| 9 | Final checkpoint | cp-09-synthesis |

---

## OBSERVER COMMANDS

| Command | Description |
|---------|-------------|
| `*checkpoints` | List all checkpoints |
| `*cp-list` | Alias for *checkpoints |
| `*cp-show:{N}` | Show checkpoint details |
| `*cp-diff:{N}` | Diff from checkpoint N to current |
| `*cp-diff:{A}:{B}` | Diff between two checkpoints |
| `*rollback:{N}` | Rollback to step N |
| `*rollback:cp-{id}` | Rollback to specific checkpoint |
| `*cp-export` | Export checkpoints to archive |
| `*cp-validate` | Validate all checkpoints |

---

## ERROR RECOVERY

### Checkpoint Creation Failure

```markdown
IF checkpoint creation fails:
1. Log error with details
2. Notify observer:
   ```
   ⚠️  Checkpoint creation failed

   Error: {error_message}

   Options:
   - *retry-cp → Retry checkpoint creation
   - *skip-cp → Continue without checkpoint (risky)
   - *pause → Pause for manual intervention
   ```
3. Do NOT proceed until resolved
```

### Rollback Failure

```markdown
IF rollback fails:
1. Attempt to restore to safe state
2. If git rollback failed:
   - Try manual file restoration
   - Report git status

3. Notify observer:
   ```
   ⚠️  Rollback failed

   Error: {error_message}

   Current state:
   - Git: {status}
   - Files: {status}

   Options:
   - *rollback-force → Force rollback (destructive)
   - *manual-restore → Enter manual recovery mode
   - *abort → Abort session
   ```
```

### Checkpoint Corruption

```markdown
IF checkpoint validation fails:
1. Mark checkpoint as invalid
2. Check adjacent checkpoints
3. Notify observer:
   ```
   ⚠️  Checkpoint corrupted: cp-{N}

   Last valid: cp-{N-1}

   Options:
   - *rollback:{N-1} → Rollback to last valid
   - *cp-rebuild:{N} → Attempt rebuild from git
   - *continue → Continue (checkpoint unavailable)
   ```
```

---

## GIT INTEGRATION

### Branch Strategy

```bash
# Create session branch at init
git checkout -b go-team/{session-id}

# Each checkpoint creates a commit
git add -A
git commit -m "checkpoint: step-{N} - {step_name}"

# Tag important checkpoints
git tag "checkpoint/step-{N}" -m "{description}"

# Rollback uses git reset
git reset --hard {checkpoint_commit}
```

### Commit Message Format

```
checkpoint: step-{N} - {step_name}

Session: {session-id}
Topic: {topic}
Checkpoint: cp-{step}-{timestamp}

Files changed:
- {list of files}

Metrics:
- Build: {pass/fail}
- Coverage: {%}
```

---

## STORAGE STRUCTURE

```
.claude/agents/microai/teams/go-team/
└── checkpoints/
    └── {session-id}/
        ├── manifest.json           # Session manifest
        ├── cp-01-init.json         # Checkpoint data
        ├── cp-02-requirements.json
        ├── cp-03-architecture.json
        ├── cp-04-implementation.json
        ├── cp-05-testing.json
        ├── cp-06-review-1.json     # Review iteration 1
        ├── cp-06-review-2.json     # Review iteration 2
        ├── cp-06-review-final.json # Review complete
        ├── cp-07-optimization.json
        ├── cp-08-release.json
        ├── cp-09-synthesis.json
        └── rollback-history.json   # Rollback audit log
```

---

## INTEGRATION WITH WORKFLOW

### Workflow State Extension

```yaml
go_team_state:
  # ... existing fields ...

  # Checkpoint extension
  checkpoint:
    enabled: true
    session_id: "{uuid}"
    current_checkpoint: "cp-03-architecture"
    checkpoint_count: 3
    last_checkpoint_at: "ISO-8601"
    git_integration: true
    git_branch: "go-team/{session-id}"
```

### Step Integration

Each step file will include:

```markdown
## CHECKPOINT INTEGRATION

### Pre-Step
- Verify previous checkpoint exists (if not step 1)
- Load state from current checkpoint

### Post-Step
- Create checkpoint after success criteria met
- Include all outputs in checkpoint
- Confirm checkpoint creation before proceeding

### On Error
- Do NOT create checkpoint
- Offer rollback to previous checkpoint
```
