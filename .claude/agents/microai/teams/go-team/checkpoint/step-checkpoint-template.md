---
name: step-checkpoint-template
description: Template for checkpoint integration in step files
version: 1.0.0
---

# Step Checkpoint Integration Template

**Sử dụng:** Copy section này vào cuối mỗi step file để tích hợp checkpoint.

---

## CHECKPOINT INTEGRATION

### Pre-Step Verification

Trước khi bắt đầu step, verify checkpoint trước:

```markdown
## Pre-Step Checkpoint Check

IF step_number > 1:
  1. Verify previous checkpoint exists
     ```
     prev_cp = find_checkpoint(session_id, step_number - 1)
     if not prev_cp:
       warn("Previous checkpoint not found. Proceeding without safety net.")
     else:
       display("✓ Previous checkpoint: {prev_cp.id}")
     ```

  2. Load state from checkpoint
     ```
     if prev_cp and prev_cp.valid:
       verify_state_matches(current_state, prev_cp.state)
     ```

  3. Display checkpoint status
     ```
     Checkpoint Status:
     - Previous: {prev_cp.id} ✓
     - Current step: {step_number}
     - Rollback available: Yes
     ```
```

### Post-Step Checkpoint Creation

Sau khi step hoàn thành và success criteria đạt:

```markdown
## Post-Step Checkpoint

TRIGGER: All success criteria met

ACTIONS:

1. Prepare checkpoint data
   ```yaml
   checkpoint_data:
     step_number: {current_step}
     step_name: "{step_name}"
     state: {current go_team_state}
     outputs: {current outputs}
     files:
       created: {list of new files}
       modified: {list of modified files}
   ```

2. Create checkpoint
   ```
   checkpoint_id = create_checkpoint(checkpoint_data)
   ```

3. Display confirmation
   ```
   ═══════════════════════════════════════════
   ✓ CHECKPOINT CREATED
   ═══════════════════════════════════════════

   Step {step_number}: {step_name} - Complete

   Checkpoint: {checkpoint_id}
   Files: {file_count}
   Git commit: {commit_hash_short}

   This state is now recoverable via:
     *rollback:{step_number}

   ═══════════════════════════════════════════
   ```

4. Proceed to next step
   ```
   Press [Enter] to continue to Step {step_number + 1}
   or use *pause to review before proceeding.
   ```
```

### On Error Handling

Khi có lỗi xảy ra trong step:

```markdown
## On Step Error

IF error occurs during step execution:

1. DO NOT create checkpoint
   ```
   skip_checkpoint = true
   ```

2. Log error
   ```yaml
   error_log:
     step: {step_number}
     error: "{error_message}"
     timestamp: {now}
     state_before_error: {last_known_good_state}
   ```

3. Offer recovery options
   ```
   ⚠️  ERROR IN STEP {step_number}
   ═══════════════════════════════════════════

   Error: {error_message}

   Last checkpoint: {last_checkpoint.id} (Step {last_checkpoint.step})

   Options:
   - *retry        → Retry current step
   - *rollback:{N} → Rollback to step N
   - *pause        → Pause for manual intervention
   - *skip         → Skip and continue (risky)

   ═══════════════════════════════════════════
   ```

4. Wait for observer decision
```

---

## STEP-SPECIFIC CHECKPOINT CONFIGURATIONS

### Step 1: Init

```yaml
checkpoint_config:
  id_format: "cp-01-init"
  includes:
    - session_id
    - topic
    - project_context
  excludes: []
  git_message: "checkpoint: step-01 - session initialized"
```

### Step 2: Requirements

```yaml
checkpoint_config:
  id_format: "cp-02-requirements"
  includes:
    - spec_document
    - user_stories
    - acceptance_criteria
    - api_contracts
  excludes: []
  git_message: "checkpoint: step-02 - requirements gathered"
  special:
    - Save spec to ./docs/go-team/spec.md
```

### Step 3: Architecture

```yaml
checkpoint_config:
  id_format: "cp-03-architecture"
  includes:
    - architecture_document
    - package_structure
    - interface_definitions
    - data_flow
  excludes: []
  git_message: "checkpoint: step-03 - architecture designed"
  special:
    - Save architecture to ./docs/go-team/architecture.md
```

### Step 4: Implementation

```yaml
checkpoint_config:
  id_format: "cp-04-implementation"
  includes:
    - all_code_files
    - folder_structure
    - main.go_with_di
  excludes:
    - test_files (saved in step 5)
  git_message: "checkpoint: step-04 - code implemented"
  special:
    - Verify build passes before checkpoint
```

### Step 5: Testing

```yaml
checkpoint_config:
  id_format: "cp-05-testing"
  includes:
    - all_test_files
    - mock_implementations
    - test_coverage_report
  excludes: []
  git_message: "checkpoint: step-05 - tests created"
  special:
    - Include coverage percentage in checkpoint
```

### Step 6: Review Loop

```yaml
checkpoint_config:
  id_format: "cp-06-review-{iteration}"
  includes:
    - review_comments
    - fixes_applied
    - iteration_metrics
  excludes: []
  git_message: "checkpoint: step-06 - review iteration {N}"
  special:
    - Create checkpoint after EACH iteration
    - Final checkpoint: cp-06-review-final
```

### Step 7: Optimization

```yaml
checkpoint_config:
  id_format: "cp-07-optimization"
  includes:
    - optimization_changes
    - benchmark_results
    - performance_metrics
  excludes: []
  git_message: "checkpoint: step-07 - optimizations applied"
```

### Step 8: Release

```yaml
checkpoint_config:
  id_format: "cp-08-release"
  includes:
    - dockerfile
    - github_actions_workflow
    - makefile
    - release_notes
  excludes: []
  git_message: "checkpoint: step-08 - release configured"
```

### Step 9: Synthesis

```yaml
checkpoint_config:
  id_format: "cp-09-synthesis"
  includes:
    - session_summary
    - final_metrics
    - all_outputs
  excludes: []
  git_message: "checkpoint: step-09 - session complete"
  special:
    - This is the final checkpoint
    - Mark session as complete
```
