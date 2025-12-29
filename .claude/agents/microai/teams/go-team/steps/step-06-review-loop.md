---
stepNumber: 6
nextStep: './step-07-optimization.md'
agent: reviewer-agent
hasBreakpoint: true
iteration:
  default: 3
  min: 1
  max: 10
  configurable: true
  extend_on_request: true
checkpoint:
  enabled: true
  id_format: "cp-06-review-{iteration}"
  per_iteration: true
  final_id: "cp-06-review-final"
---

# Step 06: Review Loop

## STEP GOAL

Reviewer Agent reviews code quality, runs static analysis, and iterates with Coder Agent until all issues are resolved.

## AGENT ACTIVATION

Load persona từ `../agents/reviewer-agent.md`

Secondary agents (for fixes):
- `../agents/fixer-agent.md` - For simple fixes (lint, style, <20 lines)
- `../agents/go-coder-agent.md` - For complex fixes (architecture, logic)
- `../agents/test-agent.md` - For test updates

## LOOP PROTOCOL

```
iteration = 0
max_iterations = get_config("iteration.default") || 3  # Configurable!

WHILE (iteration < max_iterations) AND (NOT all_checks_pass):

  1. Reviewer Agent reviews code
     - Run static analysis
     - Classify issues: SIMPLE vs COMPLEX

  2. IF issues_found > 0:
       - FOR EACH issue:
           IF is_simple_fix(issue):
             → Route to Fixer Agent
           ELSE:
             → Route to Coder Agent
       - iteration++

  3. ELSE:
       - EXIT loop (success)

IF iteration >= max_iterations AND NOT all_checks_pass:
  - Present status to observer
  - Ask: continue, accept, or abort
```

## FIX ROUTING DECISION

```
┌─────────────────────────────────────────────────────────────────┐
│                     ISSUE CLASSIFICATION                         │
└───────────────────────────┬─────────────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              │  Analyze issue scope       │
              └─────────────┬─────────────┘
                            │
         ┌──────────────────┴──────────────────┐
         │                                     │
         ▼                                     ▼
┌─────────────────────┐              ┌─────────────────────┐
│   SIMPLE FIX        │              │   COMPLEX FIX       │
│   → Fixer Agent     │              │   → Coder Agent     │
├─────────────────────┤              ├─────────────────────┤
│ • Lint/style issues │              │ • Algorithm changes │
│ • < 20 lines change │              │ • Interface changes │
│ • Error wrapping    │              │ • Architecture mods │
│ • Add mutex/sync    │              │ • New feature logic │
│ • Input validation  │              │ • Critical security │
│ • Comment/docs      │              │ • > 20 lines change │
└─────────────────────┘              └─────────────────────┘
```

### Routing Matrix

| Issue Category | Criteria | Route To | Model |
|----------------|----------|----------|-------|
| Lint/Style | Any size | Fixer | sonnet |
| Naming | < 10 lines | Fixer | sonnet |
| Error wrapping | < 10 lines | Fixer | sonnet |
| Add mutex | < 15 lines | Fixer | sonnet |
| Input validation | < 20 lines | Fixer | sonnet |
| Algorithm | Any size | Coder | opus |
| Interface change | Any size | Coder | opus |
| New logic | Any size | Coder | opus |
| Critical security | Any size | Coder | opus |
| > 20 lines | Any type | Coder | opus |

## ITERATION CONFIGURATION

### Default Settings

```yaml
iteration:
  default: 3          # Default max iterations
  min: 1              # Minimum allowed
  max: 10             # Maximum allowed
  configurable: true  # Can be changed at runtime
  extend_on_request: true  # Allow extending when max reached
```

### Configuration Commands

| Command | Description | Example |
|---------|-------------|---------|
| `*iterations` | Show current iteration config | - |
| `*iterations:N` | Set max iterations to N | `*iterations:5` |
| `*iterations:+N` | Add N more iterations | `*iterations:+2` |
| `*iterations:reset` | Reset to default (3) | - |

### Runtime Examples

```bash
# Before starting review loop
*iterations:5       # Allow up to 5 iterations

# During review loop (when max reached)
*iterations:+2      # Add 2 more iterations and continue

# Reset to default
*iterations:reset   # Back to 3
```

### Configuration via Workflow

In `workflow.md`, set iteration limit:

```yaml
go_team_state:
  max_iterations: 5  # Override default
```

Or via command:

```bash
# At session start
*config:max_iterations=5
```

## EXECUTION SEQUENCE

### 1. Reviewer Agent Introduction

```
[Reviewer Agent]

Bắt đầu review code cho "{topic}"...

Running checks:
- go vet
- golangci-lint
- go test -race
```

### 2. Run Static Analysis

```bash
# Go vet
go vet ./...

# Linting (if golangci-lint available)
golangci-lint run

# Race detection
go test -race ./...
```

### 3. Manual Code Review

Reviewer checks:
- [ ] Race conditions (shared state)
- [ ] Goroutine leaks
- [ ] Error handling (all errors checked)
- [ ] Context usage (passed everywhere)
- [ ] Resource management (defer close)
- [ ] Security (no hardcoded secrets)
- [ ] Style (naming, formatting)

### 4. Generate Review Report

```markdown
## Code Review Report - Iteration {N}

### Tool Results

**go vet:**
{output or "PASS"}

**golangci-lint:**
{output or "PASS"}

**go test -race:**
{output or "PASS"}

### Manual Review

#### Critical Issues (Must Fix)

**[C1] Race condition in cache**
- File: internal/repo/memory_repo.go:45
- Issue: Map accessed without lock
- Fix: Add sync.RWMutex

**[C2] Unchecked error**
- File: internal/handler/handler.go:67
- Issue: json.Encode error ignored
- Fix: Check and log error

#### Warnings

**[W1] Missing context check**
- File: internal/service/service.go:34
- Suggestion: Check ctx.Done() in loop

### Summary
- Critical: 2
- Warnings: 1
- Status: NEEDS_FIX
```

### 5. Fix Issues (if any)

IF issues_found > 0:

#### Route to Fixer Agent (Simple Fixes)
```
[Fixer Agent]

Fixing {N} simple issues...

Fix 1: Lint - missing comment (simple)
{show code change}

Fix 2: Error wrapping (simple)
{show code change}

Verification: go build ./... ✓

Passing to Test Agent for re-run...
```

#### Route to Coder Agent (Complex Fixes)
```
[Go Coder Agent]

Fixing {N} complex issues...

Fix 1: Race condition (complex - needs sync.RWMutex pattern)
{show code change}

Fix 2: Algorithm rewrite (complex - performance issue)
{show code change}

Re-running tests...
```

#### Escalation from Fixer to Coder
```
[Fixer Agent]

⚠️ ESCALATION: Issue too complex

Issue: "Refactor authentication flow"
Estimated changes: 45 lines
Reason: Exceeds 20-line threshold, requires interface change

Returning to Orchestrator → Route to Coder Agent
```

### 6. Re-Review

Loop back to step 2 với iteration++

### 7. Final Status

IF all_checks_pass:

```
[Reviewer Agent]

✅ Code Review PASSED

All checks:
- go vet: PASS
- golangci-lint: PASS
- go test -race: PASS
- Manual review: PASS

Metrics:
- Build: PASS
- Coverage: {X}%
- Lint: CLEAN
- Race-free: YES

---
═══════════════ BREAKPOINT ═══════════════

Observer, code đã pass review.

Options:
- [Enter] → Tiếp tục đến Optimization
- @reviewer: <feedback> → Request additional review
- *pause → Review code manually
```

ELSE IF max_iterations_reached:

```
[Reviewer Agent]

⚠️ Max iterations reached ({max_iterations})

Remaining issues:
{list of unresolved issues}

Options:
- *continue → Add more iterations
- *accept → Accept current state
- *abort → Abort session
```

## OUTPUT

```yaml
outputs:
  review_comments:
    - iteration: 1
      critical: [{...}]
      warnings: [{...}]
      status: "needs_fix"
    - iteration: 2
      critical: []
      warnings: [{...}]
      status: "pass"
  metrics:
    build_pass: true
    test_coverage: 85
    lint_clean: true
    race_free: true
```

## SUCCESS CRITERIA

- [ ] All static analysis tools pass
- [ ] No critical issues remaining
- [ ] Coverage meets target (80%+)
- [ ] Race detection passes
- [ ] Observer approves at breakpoint
- [ ] Ready for Optimization phase

---

## CHECKPOINT INTEGRATION

### Pre-Step Checkpoint Verification

```markdown
Before starting review loop:

1. Verify previous checkpoint
   ```
   prev_cp = find_checkpoint(session_id, 5)  # step-05-testing
   if not prev_cp:
     warn("No checkpoint from Testing phase")
   else:
     display("✓ Previous checkpoint: {prev_cp.id}")
   ```

2. Display checkpoint status
   ```
   Checkpoint Status:
   - Last checkpoint: cp-05-testing ✓
   - Rollback available: *rollback:5
   ```
```

### Per-Iteration Checkpoint

```markdown
After EACH review iteration:

1. Capture iteration state
   ```yaml
   checkpoint_data:
     step: 6
     step_name: "review"
     iteration: {current_iteration}
     state:
       phase: "review"
       iteration_count: {iteration}
       metrics:
         build_pass: {status}
         test_coverage: {%}
         lint_clean: {status}
         race_free: {status}
     outputs:
       review_comments: [{...}]
       fixes_applied: [{...}]
     files:
       modified: [list of fixed files]
   ```

2. Create iteration checkpoint
   ```
   checkpoint_id = "cp-06-review-{iteration}-{timestamp}"
   save_checkpoint(checkpoint_data)
   ```

3. Git commit for iteration
   ```bash
   git add -A
   git commit -m "checkpoint: step-06 - review iteration {iteration}

   Session: {session_id}
   Iteration: {iteration}/{max_iterations}
   Issues fixed: {count}
   "
   ```

4. Display iteration checkpoint
   ```
   ───────────────────────────────────────────
   ✓ Iteration {iteration} checkpoint saved

   Checkpoint: cp-06-review-{iteration}
   Issues fixed: {fixed_count}
   Remaining: {remaining_count}

   Rollback to this iteration: *rollback:cp-06-review-{iteration}
   ───────────────────────────────────────────
   ```
```

### Final Review Checkpoint

```markdown
When review loop completes (all checks pass):

1. Capture final state
   ```yaml
   checkpoint_data:
     step: 6
     step_name: "review-final"
     iteration: "final"
     state:
       phase: "review-complete"
       iteration_count: {total_iterations}
       metrics:
         build_pass: true
         test_coverage: {%}
         lint_clean: true
         race_free: true
     outputs:
       review_comments: [all iterations]
       final_status: "PASSED"
   ```

2. Create final checkpoint
   ```
   checkpoint_id = "cp-06-review-final-{timestamp}"
   save_checkpoint(checkpoint_data)
   ```

3. Git commit
   ```bash
   git add -A
   git commit -m "checkpoint: step-06 - review PASSED

   Session: {session_id}
   Total iterations: {count}
   Coverage: {%}
   All checks: PASS
   "
   ```

4. Display at breakpoint
   ```
   ═══════════════════════════════════════════
   ✓ REVIEW COMPLETE - CHECKPOINT SAVED
   ═══════════════════════════════════════════

   Final checkpoint: cp-06-review-final

   Summary:
   - Iterations: {count}
   - Issues fixed: {total}
   - Coverage: {%}
   - All checks: PASS

   Checkpoints created:
   - cp-06-review-1
   - cp-06-review-2
   - cp-06-review-final ← current

   Rollback options:
   - *rollback:cp-06-review-1 → Before first fixes
   - *rollback:cp-06-review-2 → Before second fixes
   - *rollback:5 → Back to Testing phase

   ═══════════════════════════════════════════
   ═══════════════ BREAKPOINT ═══════════════
   ═══════════════════════════════════════════

   Press [Enter] to continue to Optimization phase...
   ```
```

### On Max Iterations Reached

```markdown
If max_iterations reached without passing:

1. Create checkpoint with current state
   ```
   checkpoint_id = "cp-06-review-max-reached"
   checkpoint_data.state.status = "MAX_ITERATIONS_REACHED"
   save_checkpoint(checkpoint_data)
   ```

2. Offer recovery options
   ```
   ⚠️  MAX ITERATIONS REACHED
   ═══════════════════════════════════════════

   Checkpoint saved: cp-06-review-max-reached

   Remaining issues:
   {list of unresolved issues}

   Options:
   - *continue → Add more iterations (extend limit)
   - *accept → Accept current state and continue
   - *rollback:4 → Back to Implementation (try different approach)
   - *rollback:3 → Back to Architecture (redesign)
   - *abort → Abort session

   ═══════════════════════════════════════════
   ```
```

---

## NEXT STEP

After breakpoint approval, load `./step-07-optimization.md`
