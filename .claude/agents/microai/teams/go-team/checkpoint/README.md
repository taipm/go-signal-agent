# Go Team Checkpoint & Session System

**Version:** 1.1.0

Checkpoint/Rollback và Session Auto-Resume mechanism cho Go Team workflow, cho phép:
- Khôi phục về trạng thái trước đó (rollback)
- Tự động resume session sau khi bị gián đoạn (crash, timeout, disconnect)

---

## Quick Start

### Checkpoint Commands

| Command | Mô tả |
|---------|-------|
| `*checkpoints` | Liệt kê tất cả checkpoints |
| `*cp-show:{N}` | Xem chi tiết checkpoint tại step N |
| `*cp-diff:{N}` | So sánh từ step N đến hiện tại |
| `*rollback:{N}` | Rollback về step N |
| `*cp-validate` | Kiểm tra tính toàn vẹn của checkpoints |
| `*cp-export` | Export checkpoints thành archive |

### Session Management Commands

| Command | Mô tả |
|---------|-------|
| `*sessions` | Liệt kê tất cả sessions |
| `*resume` | Resume session gần nhất bị gián đoạn |
| `*resume:{id}` | Resume session cụ thể |
| `*session-info` | Xem thông tin session hiện tại |
| `*session-info:{id}` | Xem thông tin session cụ thể |
| `*abandon:{id}` | Đánh dấu session không thể resume |
| `*cleanup` | Dọn dẹp sessions cũ/đã hoàn thành |

### Ví Dụ

```bash
# Xem danh sách checkpoints
*checkpoints

# Rollback về sau khi hoàn thành Architecture
*rollback:3

# Xem thay đổi từ Requirements đến hiện tại
*cp-diff:2

# Rollback về iteration 1 của Review loop
*rollback:cp-06-review-1
```

---

## Architecture

```
checkpoint/
├── README.md                    # File này
├── checkpoint-manager.md        # Core architecture & data structures
├── checkpoint-operations.md     # Operation handlers (create, rollback, diff)
└── step-checkpoint-template.md  # Template cho step integration

checkpoints/{session-id}/        # Runtime checkpoint storage
├── manifest.json                # Session manifest
├── cp-01-init.json
├── cp-02-requirements.json
├── cp-03-architecture.json
├── cp-04-implementation.json
├── cp-05-testing.json
├── cp-06-review-1.json
├── cp-06-review-2.json
├── cp-06-review-final.json
├── cp-07-optimization.json
├── cp-08-release.json
├── cp-09-synthesis.json
└── rollback-history.json
```

---

## Checkpoint Points

| Step | Checkpoint ID | Trigger |
|------|---------------|---------|
| 1 | `cp-01-init` | After session initialization |
| 2 | `cp-02-requirements` | After spec approved at breakpoint |
| 3 | `cp-03-architecture` | After architecture approved at breakpoint |
| 4 | `cp-04-implementation` | After all code generated |
| 5 | `cp-05-testing` | After tests written |
| 6 | `cp-06-review-{N}` | After each review iteration |
| 6 | `cp-06-review-final` | After review loop complete |
| 7 | `cp-07-optimization` | After optimizations applied |
| 8 | `cp-08-release` | After release config created |
| 9 | `cp-09-synthesis` | Final checkpoint |

---

## Data Structure

### Checkpoint File

```json
{
  "id": "cp-03-architecture-20251228-213000",
  "session_id": "abc-123-def",
  "step_number": 3,
  "step_name": "architecture",
  "created_at": "2025-12-28T21:30:00Z",
  "state": {
    "topic": "user-service",
    "phase": "architecture",
    "current_step": 3,
    "iteration_count": 0,
    "metrics": {
      "build_pass": false,
      "test_coverage": 0,
      "lint_clean": false,
      "race_free": false
    }
  },
  "outputs": {
    "spec": "...",
    "architecture": "..."
  },
  "files": {
    "created": ["docs/architecture.md"],
    "modified": []
  },
  "git": {
    "enabled": true,
    "branch": "go-team/abc-123-def",
    "commit_hash": "a1b2c3d4e5f6"
  },
  "checksum": "sha256:..."
}
```

### Manifest File

```json
{
  "session_id": "abc-123-def",
  "topic": "user-service",
  "created_at": "2025-12-28T21:00:00Z",
  "updated_at": "2025-12-28T21:30:00Z",
  "checkpoints": [
    {"id": "cp-01-init", "step": 1, "valid": true},
    {"id": "cp-02-requirements", "step": 2, "valid": true},
    {"id": "cp-03-architecture", "step": 3, "valid": true}
  ],
  "current_checkpoint": "cp-03-architecture",
  "rollback_history": []
}
```

---

## Git Integration

Checkpoint system tích hợp với Git:

1. **Branch per session:** `go-team/{session-id}`
2. **Commit per checkpoint:** Tagged với checkpoint ID
3. **Rollback = Git reset:** `git reset --hard {checkpoint_commit}`

### Commit Message Format

```
checkpoint: step-{N} - {step_name}

Session: {session_id}
Checkpoint: {checkpoint_id}

Files changed:
- internal/handler/handler.go
- internal/service/service.go
```

---

## Error Handling

### Checkpoint Creation Failed

```
⚠️  CHECKPOINT CREATION FAILED

Options:
- *retry-cp    → Retry
- *skip-cp     → Continue without (risky!)
- *pause       → Manual intervention
```

### Rollback Failed

```
⚠️  ROLLBACK FAILED

Options:
- *rollback-force → Force (destructive)
- *manual-restore → Manual recovery
- *continue       → Keep current state
```

### Checkpoint Corrupted

```
⚠️  CHECKPOINT CORRUPTED: cp-{N}

Options:
- *rollback:{N-1}   → Last valid
- *cp-rebuild:{N}   → Rebuild from git
- *continue         → Proceed (unavailable)
```

---

## Benefits

1. **Safety Net:** Rollback bất kỳ lúc nào về trạng thái đã biết
2. **Audit Trail:** Lịch sử đầy đủ của session
3. **Error Recovery:** Khôi phục nhanh khi có lỗi
4. **Experimentation:** Thử nghiệm an toàn, rollback nếu không hiệu quả
5. **Git Integration:** Backup tự động qua git commits

---

## Files

- [checkpoint-manager.md](./checkpoint-manager.md) - Core architecture
- [checkpoint-operations.md](./checkpoint-operations.md) - Operation handlers
- [step-checkpoint-template.md](./step-checkpoint-template.md) - Integration template
