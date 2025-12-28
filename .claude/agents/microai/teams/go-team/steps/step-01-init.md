---
stepNumber: 1
nextStep: './step-02-requirements.md'
agent: orchestrator
hasBreakpoint: false
checkpoint:
  enabled: true
  id_format: "cp-01-init"
  auto_create: true
---

# Step 01: Session Initialization

## STEP GOAL

Initialize the Go Team session by gathering project context, loading relevant files, and preparing the workflow state.

## EXECUTION SEQUENCE

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

## NEXT STEP

Load and execute `./step-02-requirements.md`
