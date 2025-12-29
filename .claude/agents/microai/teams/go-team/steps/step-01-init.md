---
stepNumber: 1
nextStep: './step-02-requirements.md'
agent: orchestrator
hasBreakpoint: false
checkpoint:
  enabled: true
  id_format: "cp-01-init"
  auto_create: true
session:
  auto_save: true
  auto_save_interval: 30
  resumable: true
---

# Step 01: Session Initialization

## STEP GOAL

Initialize the Go Team session by gathering project context, loading relevant files, and preparing the workflow state. **Check for interrupted sessions and offer resume option.**

## EXECUTION SEQUENCE

### 0. Check for Interrupted Sessions (NEW)

```markdown
BEFORE starting new session:

1. Load session registry
   ```
   registry = load_json("./checkpoints/sessions.json")
   ```

2. Find interrupted/resumable sessions
   ```
   resumable = registry.sessions.filter(s =>
     s.status in ["interrupted", "paused"] && s.resumable
   )
   ```

3. IF resumable sessions found:
   ```
   ═══════════════════════════════════════════════════════════
   ⚠️  INTERRUPTED SESSIONS FOUND
   ═══════════════════════════════════════════════════════════

   Found {count} session(s) that can be resumed:

   ┌────┬──────────────────┬─────────────────────────┬────────┬────────────┐
   │ #  │ Session ID       │ Topic                   │ Step   │ Status     │
   ├────┼──────────────────┼─────────────────────────┼────────┼────────────┤
   │ 1  │ session-abc123   │ user-auth-service       │ 4/9    │ interrupted│
   │ 2  │ session-def456   │ payment-gateway         │ 6/9    │ paused     │
   └────┴──────────────────┴─────────────────────────┴────────┴────────────┘

   Commands:
   - *resume         → Resume most recent session
   - *resume:1       → Resume session #1
   - *resume:{id}    → Resume by session ID
   - *new            → Start new session (ignore above)
   - *abandon:{id}   → Abandon session permanently

   ═══════════════════════════════════════════════════════════
   ```

4. IF *resume selected:
   - Load session recovery protocol
   - Skip to "Resume Session" section below

5. IF *new or no interrupted sessions:
   - Continue to step 1 (Welcome Observer)
```

### 1. Welcome Observer

```
=== GO TEAM - AI Coding Team cho Go ===

Chào mừng! Tôi là Orchestrator của Go Team.

Team của chúng tôi gồm:
- PM Agent: Thu thập requirements
- Architect Agent: Thiết kế hệ thống
- Go Coder Agent: Viết code
- Test Agent: Viết tests
- Reviewer Agent: Review code
- Optimizer Agent: Tối ưu performance
- DevOps Agent: CI/CD và release

Workflow: Requirements → Design → Code → Test → Review → Optimize → Release
```

### 2. Gather Topic

Hỏi observer:
- Dự án/feature gì cần phát triển?
- Có repository/codebase sẵn không?
- Mục tiêu chính là gì?

### 3. Load Project Context

Tìm và đọc các files quan trọng:
- `README.md` - Project overview
- `go.mod` - Dependencies
- Existing code structure (nếu có)

### 4. Initialize Session State

```yaml
go_team_state:
  topic: "{user_provided_topic}"
  date: "{current_date}"
  phase: "init"
  current_agent: "orchestrator"
  current_step: 1
```

### 5. Display Workflow Roadmap

```
Workflow cho session này:

1. [✓] Init - Đang thực hiện
2. [ ] Requirements - PM Agent thu thập specs
3. [ ] Architecture - Architect thiết kế hệ thống
4. [ ] Implementation - Coder viết code
5. [ ] Testing - Test Agent viết tests
6. [ ] Review Loop - Review và fix issues
7. [ ] Optimization - Tối ưu performance
8. [ ] Release - CI/CD setup
9. [ ] Synthesis - Tổng kết

Observer Controls:
- [Enter] → Tiếp tục
- *pause → Tạm dừng
- *skip-to:N → Nhảy đến step N
- *exit → Kết thúc session
```

### 6. Prepare for Requirements Phase

- Confirm project context loaded
- Summarize what we know
- Transition message: "Chuyển sang Requirements phase. PM Agent sẽ thu thập specs..."

## SUCCESS CRITERIA

- [ ] Topic collected from observer
- [ ] Project context loaded (if exists)
- [ ] Session state initialized
- [ ] Workflow roadmap displayed
- [ ] Ready to proceed to step 02

---

## CHECKPOINT INTEGRATION

### Initialize Checkpoint System

```markdown
At session start:

1. Generate session ID
   ```
   session_id = generate_uuid()
   ```

2. Create checkpoint directory
   ```
   checkpoint_dir = ./.claude/agents/microai/teams/go-team/checkpoints/{session_id}
   mkdir -p {checkpoint_dir}
   ```

3. Initialize manifest
   ```json
   {
     "session_id": "{session_id}",
     "topic": "{topic}",
     "created_at": "{timestamp}",
     "checkpoints": [],
     "rollback_history": []
   }
   ```

4. Create git branch (if git enabled)
   ```bash
   git checkout -b go-team/{session_id}
   ```

5. Display checkpoint status
   ```
   ✓ Checkpoint system initialized
     Session: {session_id}
     Storage: {checkpoint_dir}
     Git branch: go-team/{session_id}
   ```
```

### Post-Step Checkpoint

```markdown
After init complete:

1. Capture state
   ```yaml
   checkpoint_data:
     step: 1
     step_name: "init"
     state:
       topic: "{topic}"
       date: "{date}"
       phase: "init"
       current_step: 1
     outputs: {}
     files:
       created: []
   ```

2. Create checkpoint
   ```
   checkpoint_id = "cp-01-init-{timestamp}"
   save_checkpoint(checkpoint_data)
   ```

3. Git commit
   ```bash
   git add -A
   git commit -m "checkpoint: step-01 - session initialized

   Session: {session_id}
   Topic: {topic}
   "
   ```

4. Display confirmation
   ```
   ═══════════════════════════════════════════
   ✓ CHECKPOINT CREATED: cp-01-init
   ═══════════════════════════════════════════

   Session initialized and checkpoint saved.

   Rollback command: *rollback:1

   Press [Enter] to continue to Requirements phase...
   ═══════════════════════════════════════════
   ```
```

---

## SESSION MANAGEMENT INTEGRATION

### Initialize Session Registry

```markdown
On workflow start:

1. Ensure sessions.json exists
   ```
   if not exists("./checkpoints/sessions.json"):
     create_file("./checkpoints/sessions.json", {
       "version": "1.0.0",
       "sessions": [],
       "last_active_session": null
     })
   ```

2. Register new session
   ```json
   {
     "id": "{session_id}",
     "topic": "{topic}",
     "status": "active",
     "created_at": "{now}",
     "updated_at": "{now}",
     "current_step": 1,
     "current_checkpoint": null,
     "progress": {
       "steps_completed": 0,
       "total_steps": 9,
       "percentage": 0
     },
     "git_branch": "go-team/{session_id}",
     "resumable": true
   }
   ```

3. Start auto-save timer
   ```
   auto_save_timer = start_timer(interval=30s, callback=save_live_state)
   ```

4. Display session info
   ```
   ═══════════════════════════════════════════
   SESSION REGISTERED
   ═══════════════════════════════════════════

   Session ID: {session_id}
   Auto-save: Enabled (every 30s)

   If interrupted, resume with:
     *resume or *resume:{session_id}
   ═══════════════════════════════════════════
   ```
```

### Resume Session Protocol

```markdown
When *resume or *resume:{id} is triggered:

1. Load session data
   ```
   session = load_session(id)
   manifest = load_manifest(session.id)
   recovery = load_recovery_metadata(session.id)
   ```

2. Display recovery options
   ```
   ═══════════════════════════════════════════════════════════
   SESSION RECOVERY
   ═══════════════════════════════════════════════════════════

   Session: {session.id}
   Topic: {session.topic}
   Interrupted: {recovery.interrupted_at}
   Last step: {session.current_step} - {step_name}

   Recovery Options:

   [1] Resume from checkpoint (RECOMMENDED)
       └─ Checkpoint: {recovery.last_checkpoint}
       └─ Data loss: ~{actions} actions since checkpoint
       └─ Safe and reliable

   [2] Resume from auto-save (EXPERIMENTAL)
       └─ Last save: {state.saved_at}
       └─ Data loss: ~30 seconds
       └─ May have inconsistencies

   [3] Start fresh from step 1
       └─ Discard all progress
       └─ Clean slate

   Enter choice [1/2/3]:
   ═══════════════════════════════════════════════════════════
   ```

3. Execute recovery

   **Option 1: Checkpoint Recovery**
   ```
   checkpoint = load_checkpoint(session.id, recovery.last_checkpoint)

   # Restore git
   git checkout {checkpoint.git.branch}
   git reset --hard {checkpoint.git.commit_hash}

   # Restore state
   go_team_state = checkpoint.state
   outputs = checkpoint.outputs

   # Update registry
   session.status = "active"
   save_registry()
   ```

   **Option 2: Auto-Save Recovery**
   ```
   live_state = load_live_state(session.id)
   go_team_state = live_state.state
   outputs = live_state.outputs

   # Attempt git restore
   try:
     git checkout {session.git_branch}
   catch:
     warn("Git state may be inconsistent")
   ```

4. Confirm and continue
   ```
   ═══════════════════════════════════════════
   ✓ SESSION RESUMED SUCCESSFULLY
   ═══════════════════════════════════════════

   Session: {session.id}
   Current step: {current_step} - {step_name}
   Recovery type: {checkpoint|auto-save}

   State restored. Ready to continue.

   Press [Enter] to proceed to step {current_step + 1}...
   ═══════════════════════════════════════════
   ```
```

### Auto-Save Protocol

```markdown
Every 30 seconds (or on important actions):

1. Capture live state
   ```json
   {
     "session_id": "{session_id}",
     "saved_at": "{now}",
     "save_trigger": "auto|action",
     "state": {go_team_state},
     "outputs": {current_outputs},
     "context": {
       "last_action": "{description}",
       "current_file": "{file}",
       "pending_tasks": [...]
     },
     "recovery_point": {
       "checkpoint": "{last_checkpoint}",
       "actions_since_checkpoint": {count}
     }
   }
   ```

2. Write atomically
   ```
   write_atomic("./checkpoints/{session_id}/state.json", live_state)
   ```

3. Update registry
   ```
   registry.sessions[session_id].updated_at = now()
   save_registry()
   ```
```

### Interrupt Detection

```markdown
On session end without normal exit:

1. Detect interrupt
   - Ctrl+C captured → status = "paused"
   - Timeout → status = "interrupted"
   - Error → status = "interrupted"

2. Save final state
   ```
   save_live_state(force=true)
   ```

3. Create recovery metadata
   ```json
   {
     "session_id": "{id}",
     "interrupted_at": "{now}",
     "interrupt_reason": "{reason}",
     "last_known_state": {
       "step": {step},
       "agent": "{agent}",
       "action": "{action}"
     },
     "recovery_options": [...]
   }
   ```

4. Update registry
   ```
   session.status = "interrupted"
   session.resumable = true
   save_registry()
   ```
```

---

## NEXT STEP

Load and execute `./step-02-requirements.md`
