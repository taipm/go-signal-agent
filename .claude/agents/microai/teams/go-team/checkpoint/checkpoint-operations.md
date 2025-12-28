---
name: checkpoint-operations
description: Checkpoint operation handlers for Go Team workflow
version: 1.0.0
---

# Checkpoint Operations

**Mục đích:** Xử lý các lệnh checkpoint từ observer và thực thi các operations.

---

## OPERATION: CREATE CHECKPOINT

### Trigger

Tự động sau mỗi step hoàn thành hoặc manual qua command.

### Protocol

```markdown
## Create Checkpoint

INPUT:
- step_number: Current step (1-9)
- step_name: Name of the step
- session_state: Current go_team_state
- outputs: Current outputs object

EXECUTION:

1. Generate checkpoint ID
   ```
   cp_id = "cp-{step_number:02d}-{step_name}-{timestamp}"
   timestamp = format(now(), "YYYYMMDD-HHmmss")
   ```

2. Capture state snapshot
   ```yaml
   snapshot:
     state: {copy of go_team_state}
     outputs: {copy of outputs}
     files:
       created: [list of new files]
       modified: [list of modified files]
   ```

3. Git checkpoint (if enabled)
   ```bash
   # Stage all changes
   git add -A

   # Check if there are changes to commit
   if git diff --staged --quiet; then
     echo "No changes to commit"
   else
     git commit -m "checkpoint: step-{N} - {step_name}

   Session: {session_id}
   Checkpoint: {cp_id}
   "
   fi

   # Store commit hash
   commit_hash=$(git rev-parse HEAD)
   ```

4. Write checkpoint file
   ```
   path = {storage_path}/{session_id}/{cp_id}.json
   content = {
     id: cp_id,
     session_id: session_id,
     step_number: step_number,
     step_name: step_name,
     created_at: now(),
     state: snapshot.state,
     outputs: snapshot.outputs,
     files: snapshot.files,
     git: {
       enabled: true,
       commit_hash: commit_hash,
       branch: git_branch
     },
     checksum: sha256(content)
   }
   write_json(path, content)
   ```

5. Update manifest
   ```
   manifest_path = {storage_path}/{session_id}/manifest.json
   manifest.checkpoints.push({
     id: cp_id,
     step: step_number,
     timestamp: now(),
     valid: true
   })
   manifest.current_checkpoint = cp_id
   write_json(manifest_path, manifest)
   ```

6. Update session state
   ```yaml
   go_team_state.checkpoint:
     current_checkpoint: {cp_id}
     checkpoint_count: {count + 1}
     last_checkpoint_at: {now}
   ```

7. Notify observer
   ```
   ✓ Checkpoint created: {cp_id}
     Step: {step_number} - {step_name}
     Files: {file_count}
     Git: {commit_hash_short}
   ```

OUTPUT:
- Checkpoint file created
- Manifest updated
- Session state updated
- Observer notified
```

---

## OPERATION: LIST CHECKPOINTS

### Trigger

Observer enters `*checkpoints` or `*cp-list`

### Protocol

```markdown
## List Checkpoints

INPUT:
- session_id: Current session ID

EXECUTION:

1. Load manifest
   ```
   manifest_path = {storage_path}/{session_id}/manifest.json
   manifest = read_json(manifest_path)
   ```

2. Format output
   ```
   =================================================
   CHECKPOINTS - Session: {session_id}
   =================================================
   Topic: {manifest.topic}
   Created: {manifest.created_at}

   Available checkpoints:
   ┌────┬────────────────────────┬─────────────────────┬────────┐
   │ #  │ Checkpoint ID          │ Timestamp           │ Status │
   ├────┼────────────────────────┼─────────────────────┼────────┤
   │ 1  │ cp-01-init             │ 2025-12-28 21:00:00 │ ✓      │
   │ 2  │ cp-02-requirements     │ 2025-12-28 21:15:00 │ ✓      │
   │ 3  │ cp-03-architecture     │ 2025-12-28 21:30:00 │ ✓      │
   │ 4  │ cp-04-implementation   │ 2025-12-28 21:45:00 │ ← curr │
   └────┴────────────────────────┴─────────────────────┴────────┘

   Commands:
   - *rollback:{N}    → Rollback to step N
   - *cp-show:{N}     → Show checkpoint details
   - *cp-diff:{N}     → Diff from step N to current
   =================================================
   ```

OUTPUT:
- Formatted checkpoint list displayed to observer
```

---

## OPERATION: SHOW CHECKPOINT

### Trigger

Observer enters `*cp-show:{N}`

### Protocol

```markdown
## Show Checkpoint Details

INPUT:
- step_number: The step number to show (N)
- session_id: Current session ID

EXECUTION:

1. Find checkpoint
   ```
   checkpoint = find_checkpoint_by_step(session_id, step_number)
   if not checkpoint:
     error("Checkpoint not found for step {step_number}")
   ```

2. Load checkpoint data
   ```
   cp_path = {storage_path}/{session_id}/{checkpoint.id}.json
   cp_data = read_json(cp_path)
   ```

3. Format output
   ```
   =================================================
   CHECKPOINT DETAILS: {checkpoint.id}
   =================================================

   Step: {cp_data.step_number} - {cp_data.step_name}
   Created: {cp_data.created_at}

   STATE:
   - Phase: {cp_data.state.phase}
   - Iteration: {cp_data.state.iteration_count}
   - Metrics:
     - Build: {cp_data.state.metrics.build_pass}
     - Coverage: {cp_data.state.metrics.test_coverage}%
     - Lint: {cp_data.state.metrics.lint_clean}
     - Race-free: {cp_data.state.metrics.race_free}

   OUTPUTS:
   - Spec: {cp_data.outputs.spec ? "Yes" : "No"}
   - Architecture: {cp_data.outputs.architecture ? "Yes" : "No"}
   - Code files: {cp_data.outputs.code_files.length}
   - Test files: {cp_data.outputs.test_files.length}

   FILES:
   Created: {cp_data.files.created.length}
   Modified: {cp_data.files.modified.length}

   GIT:
   - Branch: {cp_data.git.branch}
   - Commit: {cp_data.git.commit_hash}

   VALIDATION:
   - Checksum: {cp_data.checksum}
   - Valid: {cp_data.valid ? "✓" : "✗"}
   =================================================
   ```

OUTPUT:
- Detailed checkpoint information displayed
```

---

## OPERATION: DIFF CHECKPOINTS

### Trigger

Observer enters `*cp-diff:{N}` or `*cp-diff:{A}:{B}`

### Protocol

```markdown
## Checkpoint Diff

INPUT:
- from_step: Starting checkpoint step (default: specified N)
- to_step: Ending checkpoint step (default: current)
- session_id: Current session ID

EXECUTION:

1. Load checkpoints
   ```
   from_cp = load_checkpoint(session_id, from_step)
   to_cp = load_checkpoint(session_id, to_step)
   ```

2. Calculate state diff
   ```
   state_diff = {
     phase: [from_cp.state.phase, to_cp.state.phase],
     current_step: [from_cp.state.current_step, to_cp.state.current_step],
     iteration_count: [from_cp.state.iteration_count, to_cp.state.iteration_count],
     metrics: diff_objects(from_cp.state.metrics, to_cp.state.metrics)
   }
   ```

3. Calculate output diff
   ```
   output_diff = {
     spec: from_cp.outputs.spec != to_cp.outputs.spec,
     architecture: from_cp.outputs.architecture != to_cp.outputs.architecture,
     code_files: diff_arrays(from_cp.outputs.code_files, to_cp.outputs.code_files),
     test_files: diff_arrays(from_cp.outputs.test_files, to_cp.outputs.test_files)
   }
   ```

4. Calculate file diff
   ```
   file_diff = {
     added: files_in_to_not_in_from,
     removed: files_in_from_not_in_to,
     modified: files_in_both_with_changes
   }
   ```

5. Git diff (if enabled)
   ```bash
   git log --oneline {from_cp.git.commit_hash}..{to_cp.git.commit_hash}
   git diff --stat {from_cp.git.commit_hash}..{to_cp.git.commit_hash}
   ```

6. Format output
   ```
   =================================================
   CHECKPOINT DIFF
   =================================================
   From: {from_cp.id} (Step {from_step})
   To:   {to_cp.id} (Step {to_step})

   STATE CHANGES:
   - phase: {from_cp.state.phase} → {to_cp.state.phase}
   - current_step: {from_step} → {to_step}
   - iteration_count: {from_cp.state.iteration_count} → {to_cp.state.iteration_count}

   METRICS CHANGES:
   - build_pass: {from} → {to}
   - test_coverage: {from}% → {to}%
   - lint_clean: {from} → {to}

   OUTPUT CHANGES:
   + spec: {added_lines} lines added
   + architecture: {added_lines} lines added
   + code_files: {count} files added
   + test_files: {count} files added

   FILE CHANGES:
   + {new_file_1} (new)
   + {new_file_2} (new)
   ~ {modified_file} (modified)
   - {deleted_file} (deleted)

   GIT COMMITS:
   {commit_count} commits between checkpoints
   - {sha1}: {message1}
   - {sha2}: {message2}
   =================================================
   ```

OUTPUT:
- Comprehensive diff displayed
```

---

## OPERATION: ROLLBACK

### Trigger

Observer enters `*rollback:{N}` or `*rollback:cp-{id}`

### Protocol

```markdown
## Rollback to Checkpoint

INPUT:
- target: Step number or checkpoint ID
- session_id: Current session ID
- current_state: Current go_team_state

EXECUTION:

1. Parse target
   ```
   if target starts with "cp-":
     checkpoint = find_checkpoint_by_id(session_id, target)
   else:
     checkpoint = find_checkpoint_by_step(session_id, int(target))

   if not checkpoint:
     error("Checkpoint not found: {target}")
   ```

2. Validate checkpoint
   ```
   cp_data = load_checkpoint(session_id, checkpoint.id)

   if not validate_checksum(cp_data):
     error("Checkpoint corrupted: {checkpoint.id}")

   if cp_data.step_number >= current_state.current_step:
     error("Cannot rollback forward. Target: {cp_data.step_number}, Current: {current_state.current_step}")
   ```

3. Request confirmation
   ```
   ⚠️  ROLLBACK CONFIRMATION
   =================================================

   Rolling back to: {checkpoint.id}
   Target step: {cp_data.step_number} - {cp_data.step_name}
   Current step: {current_state.current_step}

   This will:
   ✗ Discard steps {cp_data.step_number + 1} to {current_state.current_step}
   ✗ Revert {files_to_revert} files
   ✗ Discard {commits_to_discard} git commits

   Git changes:
   - Current: {current_commit_short}
   - Target:  {cp_data.git.commit_hash_short}

   Type 'yes' to confirm or 'no' to cancel:
   =================================================
   ```

4. Wait for confirmation
   ```
   response = await_observer_input()
   if response.lower() != 'yes':
     display("Rollback cancelled.")
     return
   ```

5. Execute git rollback (if enabled)
   ```bash
   # Ensure we're on the correct branch
   git checkout {cp_data.git.branch}

   # Hard reset to checkpoint commit
   git reset --hard {cp_data.git.commit_hash}

   # Verify reset
   current_hash = $(git rev-parse HEAD)
   if current_hash != cp_data.git.commit_hash:
     error("Git reset failed")
   ```

6. Restore session state
   ```
   go_team_state = deep_copy(cp_data.state)
   go_team_state.checkpoint.current_checkpoint = checkpoint.id
   ```

7. Restore outputs
   ```
   outputs = deep_copy(cp_data.outputs)
   ```

8. Record rollback in history
   ```
   rollback_record = {
     id: generate_uuid(),
     timestamp: now(),
     from_checkpoint: current_checkpoint,
     from_step: current_state.current_step,
     to_checkpoint: checkpoint.id,
     to_step: cp_data.step_number,
     reason: "observer_requested"
   }
   manifest.rollback_history.push(rollback_record)
   go_team_state.checkpoint.rollback_history.push(rollback_record)
   save_manifest(manifest)
   ```

9. Invalidate future checkpoints
   ```
   for cp in manifest.checkpoints:
     if cp.step > cp_data.step_number:
       cp.valid = false
       cp.invalidated_by = rollback_record.id
   save_manifest(manifest)
   ```

10. Notify observer
    ```
    ✓ ROLLBACK COMPLETE
    =================================================

    Restored to: {checkpoint.id}
    Current step: {cp_data.step_number} - {cp_data.step_name}

    State restored:
    - Phase: {cp_data.state.phase}
    - Outputs: {output_summary}
    - Git: {cp_data.git.commit_hash_short}

    {invalidated_count} future checkpoint(s) invalidated.

    Ready to continue from step {cp_data.step_number + 1}.

    Commands:
    - [Enter] → Continue to next step
    - *pause  → Review current state
    - *checkpoints → View updated checkpoint list
    =================================================
    ```

OUTPUT:
- Session state restored
- Outputs restored
- Git reset completed
- Rollback history recorded
- Future checkpoints invalidated
- Observer notified
```

---

## OPERATION: VALIDATE CHECKPOINTS

### Trigger

Observer enters `*cp-validate`

### Protocol

```markdown
## Validate All Checkpoints

INPUT:
- session_id: Current session ID

EXECUTION:

1. Load manifest
   ```
   manifest = load_manifest(session_id)
   ```

2. Validate each checkpoint
   ```
   results = []
   for cp in manifest.checkpoints:
     cp_data = load_checkpoint(session_id, cp.id)

     validation = {
       id: cp.id,
       step: cp.step,
       file_exists: file_exists(cp_path),
       checksum_valid: validate_checksum(cp_data),
       git_valid: validate_git_commit(cp_data.git.commit_hash),
       state_valid: validate_state_schema(cp_data.state)
     }

     validation.overall = all([
       validation.file_exists,
       validation.checksum_valid,
       validation.git_valid,
       validation.state_valid
     ])

     results.push(validation)
   ```

3. Format output
   ```
   =================================================
   CHECKPOINT VALIDATION
   =================================================
   Session: {session_id}
   Checkpoints: {manifest.checkpoints.length}

   Results:
   ┌────┬────────────────────┬──────┬──────┬─────┬───────┬─────────┐
   │ #  │ Checkpoint         │ File │ Hash │ Git │ State │ Overall │
   ├────┼────────────────────┼──────┼──────┼─────┼───────┼─────────┤
   │ 1  │ cp-01-init         │  ✓   │  ✓   │  ✓  │   ✓   │   ✓     │
   │ 2  │ cp-02-requirements │  ✓   │  ✓   │  ✓  │   ✓   │   ✓     │
   │ 3  │ cp-03-architecture │  ✓   │  ✗   │  ✓  │   ✓   │   ✗     │
   └────┴────────────────────┴──────┴──────┴─────┴───────┴─────────┘

   Summary:
   - Valid: {valid_count}
   - Invalid: {invalid_count}

   {if invalid_count > 0}
   ⚠️  Invalid checkpoints detected!

   Options:
   - *cp-rebuild:{N} → Attempt to rebuild from git
   - *rollback:{last_valid} → Rollback to last valid
   {endif}
   =================================================
   ```

OUTPUT:
- Validation results displayed
- Issues highlighted
- Recovery options provided
```

---

## OPERATION: EXPORT CHECKPOINTS

### Trigger

Observer enters `*cp-export`

### Protocol

```markdown
## Export Checkpoints

INPUT:
- session_id: Current session ID

EXECUTION:

1. Create export archive
   ```
   export_dir = {storage_path}/{session_id}/exports
   export_file = {export_dir}/checkpoint-export-{timestamp}.tar.gz

   mkdir -p {export_dir}
   ```

2. Collect all checkpoint files
   ```
   files_to_export = [
     manifest.json,
     cp-*.json,
     rollback-history.json
   ]
   ```

3. Create archive
   ```bash
   tar -czvf {export_file} -C {storage_path}/{session_id} .
   ```

4. Calculate archive hash
   ```
   archive_hash = sha256(export_file)
   ```

5. Notify observer
   ```
   ✓ CHECKPOINTS EXPORTED
   =================================================

   Archive: {export_file}
   Size: {file_size}
   Hash: {archive_hash}

   Contains:
   - Manifest
   - {checkpoint_count} checkpoints
   - Rollback history

   To restore:
   - Extract archive to checkpoint directory
   - Use *cp-validate to verify integrity
   =================================================
   ```

OUTPUT:
- Archive file created
- Export details displayed
```

---

## ERROR HANDLING

### Checkpoint Creation Error

```markdown
IF checkpoint creation fails:

1. Log error
   ```
   error_log = {
     timestamp: now(),
     operation: "create_checkpoint",
     step: step_number,
     error: error_message,
     stack: error_stack
   }
   append_to_log(error_log)
   ```

2. Notify observer
   ```
   ⚠️  CHECKPOINT CREATION FAILED
   =================================================

   Step: {step_number} - {step_name}
   Error: {error_message}

   Options:
   - *retry-cp    → Retry checkpoint creation
   - *skip-cp     → Continue without checkpoint (risky!)
   - *pause       → Pause for manual intervention

   ⚠️  Proceeding without checkpoint may cause data loss!
   =================================================
   ```

3. Wait for observer decision
```

### Rollback Error

```markdown
IF rollback fails:

1. Attempt recovery
   ```
   try:
     # Restore from backup if exists
     restore_from_backup()
   catch:
     # Log and continue to error reporting
   ```

2. Check git status
   ```bash
   git status
   git log -1 --oneline
   ```

3. Notify observer
   ```
   ⚠️  ROLLBACK FAILED
   =================================================

   Target: {checkpoint.id}
   Error: {error_message}

   Current state:
   - Git branch: {current_branch}
   - Git commit: {current_commit}
   - Session step: {current_step}

   Recovery options:
   - *rollback-force → Force rollback (⚠️ destructive)
   - *manual-restore → Enter manual recovery mode
   - *continue       → Continue from current state
   - *abort          → Abort session

   ⚠️  Manual intervention may be required!
   =================================================
   ```
```

### Checkpoint Corruption

```markdown
IF checkpoint validation fails:

1. Mark as corrupted
   ```
   checkpoint.valid = false
   checkpoint.corrupted_at = now()
   save_manifest()
   ```

2. Find last valid checkpoint
   ```
   last_valid = manifest.checkpoints
     .filter(cp => cp.valid)
     .sort_by(cp => cp.step)
     .last()
   ```

3. Notify observer
   ```
   ⚠️  CHECKPOINT CORRUPTED
   =================================================

   Corrupted: {checkpoint.id}
   Reason: {validation_error}

   Last valid checkpoint: {last_valid.id} (Step {last_valid.step})

   Options:
   - *rollback:{last_valid.step} → Rollback to last valid
   - *cp-rebuild:{checkpoint.step} → Attempt rebuild from git
   - *continue → Continue (checkpoint unavailable)
   =================================================
   ```
```
