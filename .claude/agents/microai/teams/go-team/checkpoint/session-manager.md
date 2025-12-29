---
name: session-manager
description: Session persistence and auto-resume for Go Team workflow
version: 1.0.0
---

# Session Manager

**Mục đích:** Quản lý session persistence và auto-resume, cho phép khôi phục workflow sau khi session bị gián đoạn (crash, timeout, disconnect).

---

## ARCHITECTURE

```
┌─────────────────────────────────────────────────────────────────┐
│                      SESSION MANAGER                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐       │
│  │   Session     │  │   Auto-Save   │  │   Recovery    │       │
│  │   Registry    │  │   Engine      │  │   Engine      │       │
│  └───────────────┘  └───────────────┘  └───────────────┘       │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            ▼                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                   SESSION STORAGE                        │   │
│  │  ./checkpoints/                                         │   │
│  │    ├── sessions.json          # Session registry        │   │
│  │    ├── {session-id}/                                    │   │
│  │    │   ├── manifest.json      # Session manifest        │   │
│  │    │   ├── state.json         # Live state (auto-saved) │   │
│  │    │   ├── cp-*.json          # Checkpoints             │   │
│  │    │   └── recovery.json      # Recovery metadata       │   │
│  │    └── ...                                              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## SESSION REGISTRY

### Registry Schema

```json
{
  "version": "1.0.0",
  "sessions": [
    {
      "id": "session-abc123",
      "topic": "user-authentication-service",
      "status": "active|completed|interrupted|failed",
      "created_at": "2025-12-28T21:00:00Z",
      "updated_at": "2025-12-28T21:45:00Z",
      "current_step": 4,
      "current_checkpoint": "cp-04-implementation",
      "progress": {
        "steps_completed": 4,
        "total_steps": 9,
        "percentage": 44
      },
      "git_branch": "go-team/session-abc123",
      "resumable": true
    }
  ],
  "last_active_session": "session-abc123"
}
```

### Session Status

| Status | Description |
|--------|-------------|
| `active` | Session đang chạy |
| `completed` | Session hoàn thành (step 9 done) |
| `interrupted` | Session bị gián đoạn (crash, timeout) |
| `failed` | Session thất bại (unrecoverable error) |
| `paused` | Session được pause bởi user |

---

## AUTO-SAVE ENGINE

### Live State File

Tự động save state mỗi 30 giây hoặc sau mỗi action quan trọng:

```json
{
  "session_id": "session-abc123",
  "saved_at": "2025-12-28T21:45:30Z",
  "save_trigger": "auto|action|checkpoint",

  "state": {
    "topic": "user-authentication-service",
    "phase": "implementation",
    "current_step": 4,
    "current_agent": "go-coder-agent",
    "iteration_count": 0,
    "breakpoint_active": false,
    "metrics": {
      "build_pass": true,
      "test_coverage": 0,
      "lint_clean": false,
      "race_free": false
    }
  },

  "context": {
    "last_action": "Implementing service layer",
    "pending_tasks": ["Complete repository", "Wire DI"],
    "current_file": "internal/service/auth_service.go",
    "current_line": 45
  },

  "outputs": {
    "spec": "...",
    "architecture": "...",
    "code_files": ["cmd/main.go", "internal/handler/..."],
    "partial_output": "..."
  },

  "recovery_point": {
    "checkpoint": "cp-03-architecture",
    "actions_since_checkpoint": 15,
    "can_resume_from_action": true
  }
}
```

### Auto-Save Triggers

| Trigger | Interval | Description |
|---------|----------|-------------|
| `auto` | 30 seconds | Periodic background save |
| `action` | On action | After file write, bash command |
| `checkpoint` | On checkpoint | Full state save |
| `breakpoint` | On breakpoint | Before waiting for observer |
| `agent_switch` | On handoff | Before switching agents |

---

## RECOVERY ENGINE

### Recovery Metadata

```json
{
  "session_id": "session-abc123",
  "interrupted_at": "2025-12-28T21:45:00Z",
  "interrupt_reason": "timeout|crash|disconnect|error",

  "last_known_state": {
    "step": 4,
    "agent": "go-coder-agent",
    "action": "Writing file internal/service/auth_service.go"
  },

  "recovery_options": [
    {
      "type": "checkpoint",
      "target": "cp-03-architecture",
      "description": "Resume from last checkpoint (safe)",
      "data_loss": "15 actions since checkpoint"
    },
    {
      "type": "live_state",
      "target": "state.json",
      "description": "Resume from last auto-save (experimental)",
      "data_loss": "~30 seconds of work"
    }
  ],

  "recommended": "checkpoint",
  "auto_resume_enabled": true
}
```

---

## OBSERVER COMMANDS

### Session Management

| Command | Description |
|---------|-------------|
| `*sessions` | List all sessions (active, interrupted, completed) |
| `*resume` | Resume last interrupted session |
| `*resume:{id}` | Resume specific session |
| `*session-info` | Show current session details |
| `*session-info:{id}` | Show specific session details |
| `*abandon:{id}` | Mark session as abandoned (no resume) |
| `*cleanup` | Clean up old/completed sessions |

---

## OPERATIONS

### 1. SESSION INITIALIZATION

```markdown
## Initialize New Session

TRIGGER: Go Team workflow starts

ACTIONS:

1. Generate session ID
   ```
   session_id = "session-" + uuid_short()
   ```

2. Check for interrupted sessions
   ```
   interrupted = find_interrupted_sessions()
   if interrupted.length > 0:
     prompt_resume_or_new(interrupted)
   ```

3. Create session entry
   ```json
   {
     "id": "{session_id}",
     "topic": "{topic}",
     "status": "active",
     "created_at": "{now}",
     "current_step": 1,
     "resumable": true
   }
   ```

4. Initialize auto-save
   ```
   start_auto_save_timer(interval=30s)
   ```

5. Display session info
   ```
   ═══════════════════════════════════════════
   GO TEAM SESSION STARTED
   ═══════════════════════════════════════════

   Session ID: {session_id}
   Topic: {topic}

   Auto-save: Enabled (every 30s)
   Recovery: Available via *resume

   If session is interrupted, you can resume
   with: *resume or *resume:{session_id}
   ═══════════════════════════════════════════
   ```
```

### 2. CHECK FOR RESUMABLE SESSIONS

```markdown
## Check Resumable Sessions

TRIGGER: Go Team workflow starts OR *sessions command

ACTIONS:

1. Load session registry
   ```
   registry = load_json("./checkpoints/sessions.json")
   ```

2. Filter resumable sessions
   ```
   resumable = registry.sessions.filter(s =>
     s.status in ["interrupted", "paused", "active"] &&
     s.resumable == true
   )
   ```

3. Display if found
   ```
   ═══════════════════════════════════════════
   RESUMABLE SESSIONS FOUND
   ═══════════════════════════════════════════

   Found {count} session(s) that can be resumed:

   ┌────┬──────────────────┬─────────────────────┬────────┬──────────┐
   │ #  │ Session ID       │ Topic               │ Step   │ Status   │
   ├────┼──────────────────┼─────────────────────┼────────┼──────────┤
   │ 1  │ session-abc123   │ user-auth-service   │ 4/9    │ interrupted│
   │ 2  │ session-def456   │ payment-gateway     │ 6/9    │ paused   │
   └────┴──────────────────┴─────────────────────┴────────┴──────────┘

   Commands:
   - *resume         → Resume most recent (session-abc123)
   - *resume:1       → Resume session #1
   - *resume:def456  → Resume by ID
   - *new            → Start new session (ignore above)

   ═══════════════════════════════════════════
   ```
```

### 3. RESUME SESSION

```markdown
## Resume Interrupted Session

TRIGGER: *resume or *resume:{id}

VALIDATION:

1. Find session
   ```
   session = find_session(id)
   if not session:
     error("Session not found: {id}")
   ```

2. Check resumable
   ```
   if session.status == "completed":
     error("Session already completed")
   if session.status == "failed":
     error("Session failed and cannot be resumed")
   ```

3. Load recovery metadata
   ```
   recovery = load_json("./checkpoints/{id}/recovery.json")
   ```

EXECUTION:

1. Display recovery options
   ```
   ═══════════════════════════════════════════
   SESSION RECOVERY
   ═══════════════════════════════════════════

   Session: {session_id}
   Topic: {topic}
   Interrupted: {interrupted_at}
   Last step: {step} - {step_name}

   Recovery Options:

   [1] Resume from checkpoint (RECOMMENDED)
       └─ Checkpoint: cp-{N}-{name}
       └─ Data loss: ~{actions} actions
       └─ Safe and reliable

   [2] Resume from auto-save (EXPERIMENTAL)
       └─ Saved: {saved_at}
       └─ Data loss: ~30 seconds
       └─ May have inconsistencies

   [3] Start fresh from step 1
       └─ Discard all progress
       └─ Clean slate

   Enter choice [1/2/3]:
   ═══════════════════════════════════════════
   ```

2. Execute recovery based on choice

   **Option 1: Checkpoint Recovery**
   ```
   checkpoint = load_checkpoint(session_id, recovery.checkpoint)

   # Restore git state
   git checkout {checkpoint.git.branch}
   git reset --hard {checkpoint.git.commit_hash}

   # Restore session state
   state = checkpoint.state
   outputs = checkpoint.outputs

   # Update registry
   session.status = "active"
   session.current_step = checkpoint.step_number
   save_registry()

   # Restart auto-save
   start_auto_save_timer()
   ```

   **Option 2: Auto-Save Recovery**
   ```
   live_state = load_json("./checkpoints/{id}/state.json")

   # Restore from live state
   state = live_state.state
   outputs = live_state.outputs
   context = live_state.context

   # Attempt to restore git state
   try:
     git checkout {session.git_branch}
   catch:
     warn("Git state may be inconsistent")

   # Mark as experimental
   session.recovery_type = "experimental"
   ```

3. Confirm recovery
   ```
   ═══════════════════════════════════════════
   ✓ SESSION RESUMED
   ═══════════════════════════════════════════

   Session: {session_id}
   Recovered from: {recovery_type}
   Current step: {step} - {step_name}

   State restored:
   - Topic: {topic}
   - Outputs: spec ✓, architecture ✓, code ✓
   - Metrics: build={build}, coverage={coverage}%

   Git branch: {branch}

   Ready to continue. Press [Enter] to proceed...
   ═══════════════════════════════════════════
   ```
```

### 4. AUTO-SAVE OPERATION

```markdown
## Auto-Save Protocol

TRIGGER: Timer (30s) OR Action completion

ACTIONS:

1. Capture current state
   ```
   current_state = {
     session_id: session_id,
     saved_at: now(),
     save_trigger: trigger_type,
     state: go_team_state,
     context: {
       last_action: last_action_description,
       pending_tasks: pending_tasks,
       current_file: current_file,
     },
     outputs: current_outputs,
     recovery_point: {
       checkpoint: last_checkpoint_id,
       actions_since_checkpoint: action_count,
     }
   }
   ```

2. Write to state file (atomic)
   ```
   temp_file = "./checkpoints/{id}/state.json.tmp"
   write_json(temp_file, current_state)
   rename(temp_file, "./checkpoints/{id}/state.json")
   ```

3. Update registry
   ```
   registry.sessions[session_id].updated_at = now()
   registry.sessions[session_id].current_step = current_step
   save_registry()
   ```

4. Log (silent, no user notification)
   ```
   log.debug("Auto-save completed", {session_id, step})
   ```
```

### 5. INTERRUPT DETECTION

```markdown
## Detect Session Interruption

TRIGGER: Session ends unexpectedly (no normal exit)

DETECTION METHODS:

1. **Heartbeat timeout**
   - Auto-save updates registry every 30s
   - If no update for 60s, mark as potentially interrupted
   - Background process checks registry

2. **Crash detection**
   - On workflow start, check last_active_session
   - If status is "active" but no response, mark interrupted

3. **Explicit interrupt**
   - Ctrl+C captured
   - Save final state before exit
   - Mark as "paused" (intentional) vs "interrupted" (crash)

ACTIONS:

1. Save current state
   ```
   save_live_state(force=true)
   ```

2. Create recovery metadata
   ```json
   {
     "session_id": "{id}",
     "interrupted_at": "{now}",
     "interrupt_reason": "{reason}",
     "last_known_state": {
       "step": 4,
       "agent": "go-coder-agent",
       "action": "Writing file..."
     }
   }
   ```

3. Update registry
   ```
   session.status = "interrupted"
   session.resumable = true
   save_registry()
   ```

4. Clean up resources
   - Release file locks
   - Save pending git changes
```

### 6. SESSION COMPLETION

```markdown
## Complete Session

TRIGGER: Step 9 (Synthesis) completes successfully

ACTIONS:

1. Save final checkpoint
   ```
   create_checkpoint("cp-09-synthesis-final")
   ```

2. Update registry
   ```
   session.status = "completed"
   session.completed_at = now()
   session.resumable = false
   save_registry()
   ```

3. Stop auto-save
   ```
   stop_auto_save_timer()
   ```

4. Display completion
   ```
   ═══════════════════════════════════════════
   ✓ SESSION COMPLETED
   ═══════════════════════════════════════════

   Session: {session_id}
   Topic: {topic}
   Duration: {duration}

   Final metrics:
   - Build: PASS
   - Coverage: {coverage}%
   - Lint: CLEAN
   - Security: PASSED

   Session data retained for reference.
   Use *session-info:{id} to view details.
   ═══════════════════════════════════════════
   ```
```

---

## STARTUP FLOW

```
┌─────────────────────────────────────────────────────────────┐
│                    GO TEAM STARTUP                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Load session registry                                   │
│     ↓                                                       │
│  2. Check for interrupted sessions                          │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Interrupted sessions found?                          │   │
│  │                                                      │   │
│  │  YES → Display resume prompt                         │   │
│  │        ├─ *resume → Go to step 4                     │   │
│  │        └─ *new → Go to step 3                        │   │
│  │                                                      │   │
│  │  NO → Go to step 3                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│     ↓                                                       │
│  3. Create new session                                      │
│     ↓                                                       │
│  4. Start/Resume workflow                                   │
│     ↓                                                       │
│  5. Enable auto-save                                        │
│     ↓                                                       │
│  6. Begin step execution                                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## STORAGE STRUCTURE

```
.claude/agents/microai/teams/go-team/checkpoints/
├── sessions.json                    # Session registry
├── session-abc123/
│   ├── manifest.json               # Session manifest
│   ├── state.json                  # Live state (auto-saved)
│   ├── recovery.json               # Recovery metadata
│   ├── cp-01-init.json
│   ├── cp-02-requirements.json
│   ├── cp-03-architecture.json
│   ├── cp-04-implementation.json   # Last checkpoint
│   └── rollback-history.json
├── session-def456/
│   └── ...
└── session-old789/
    └── ... (completed, can be cleaned up)
```

---

## INTEGRATION WITH WORKFLOW

### Workflow State Extension

```yaml
go_team_state:
  # ... existing fields ...

  # Session management extension
  session:
    id: "{session_id}"
    status: "active"
    created_at: "ISO-8601"
    auto_save:
      enabled: true
      interval: 30
      last_save: "ISO-8601"
    recovery:
      enabled: true
      last_checkpoint: "cp-03-architecture"
      actions_since_checkpoint: 15
```

### Step Integration

```markdown
## Session Integration in Steps

### On Step Start
- Record step start in live state
- Update session registry

### On Step Complete
- Create checkpoint
- Update session progress
- Trigger auto-save

### On Error
- Save error state
- Update recovery metadata
- Do NOT mark as failed (allow resume)
```
