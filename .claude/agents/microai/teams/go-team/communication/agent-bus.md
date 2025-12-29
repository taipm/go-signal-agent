# Agent Communication Bus

**Version:** 1.0.0

Hệ thống giao tiếp hai chiều giữa các agents trong Go Team, cho phép:
- Agents query thông tin từ agents khác
- Agents gửi notifications/alerts
- Cross-agent collaboration không qua Orchestrator

---

## Architecture

```
                    ┌─────────────────────────────────────┐
                    │          MESSAGE BUS                │
                    │  ┌─────────────────────────────────┐│
                    │  │     Message Queue               ││
                    │  │  ┌─────┬─────┬─────┬─────┐     ││
                    │  │  │ M1  │ M2  │ M3  │ ... │     ││
                    │  │  └─────┴─────┴─────┴─────┘     ││
                    │  └─────────────────────────────────┘│
                    │  ┌─────────────────────────────────┐│
                    │  │     Topic Subscriptions         ││
                    │  │  architecture → [coder, test]   ││
                    │  │  security → [coder, reviewer]   ││
                    │  │  review → [coder, test]         ││
                    │  └─────────────────────────────────┘│
                    └─────────────────────────────────────┘
                          ↑    ↑    ↑    ↑    ↑    ↑
                          │    │    │    │    │    │
    ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐
    │  PM  │  │ ARCH │  │CODER │  │ TEST │  │REVIEW│  │DEVOPS│
    └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘
```

---

## Message Types

### 1. Query (Request/Response)

Một agent hỏi thông tin từ agent khác.

```json
{
  "id": "msg-001",
  "type": "query",
  "from": "go-coder-agent",
  "to": "architect-agent",
  "topic": "architecture",
  "payload": {
    "question": "What is the interface signature for UserRepository?",
    "context": {
      "file": "internal/repo/user_repo.go",
      "line": 15
    }
  },
  "priority": "normal",
  "timestamp": "2025-12-28T22:00:00Z",
  "timeout_ms": 5000,
  "requires_response": true
}
```

**Response:**

```json
{
  "id": "msg-001-response",
  "type": "response",
  "from": "architect-agent",
  "to": "go-coder-agent",
  "in_reply_to": "msg-001",
  "payload": {
    "answer": "UserRepository interface should have: Create, GetByID, GetByEmail, Update, Delete",
    "reference": "docs/architecture.md:45-60"
  },
  "timestamp": "2025-12-28T22:00:01Z"
}
```

### 2. Notification (Fire-and-Forget)

Agent thông báo sự kiện cho các agents khác.

```json
{
  "id": "msg-002",
  "type": "notification",
  "from": "security-agent",
  "to": "*",
  "topic": "security",
  "payload": {
    "event": "vulnerability_found",
    "severity": "HIGH",
    "details": {
      "type": "SQL Injection",
      "file": "internal/repo/user_repo.go",
      "line": 42,
      "recommendation": "Use parameterized queries"
    }
  },
  "priority": "high",
  "timestamp": "2025-12-28T22:05:00Z",
  "requires_response": false
}
```

### 3. Collaboration Request

Agent yêu cầu agent khác thực hiện một task phụ.

```json
{
  "id": "msg-003",
  "type": "collaboration",
  "from": "reviewer-agent",
  "to": "test-agent",
  "topic": "review",
  "payload": {
    "request": "add_test_case",
    "description": "Missing test for error case when user not found",
    "target": {
      "function": "GetUserByID",
      "file": "internal/service/user_service.go"
    },
    "acceptance_criteria": [
      "Test returns ErrUserNotFound",
      "Mock repo returns nil user"
    ]
  },
  "priority": "normal",
  "timestamp": "2025-12-28T22:10:00Z",
  "requires_response": true
}
```

### 4. Broadcast

Orchestrator hoặc agent gửi thông báo cho tất cả.

```json
{
  "id": "msg-004",
  "type": "broadcast",
  "from": "orchestrator",
  "to": "*",
  "topic": "workflow",
  "payload": {
    "event": "phase_change",
    "from_phase": "implementation",
    "to_phase": "testing",
    "context": {
      "files_created": 8,
      "total_lines": 450
    }
  },
  "priority": "low",
  "timestamp": "2025-12-28T22:15:00Z",
  "requires_response": false
}
```

---

## Topic Subscriptions

Mỗi agent subscribe các topics liên quan:

| Agent | Subscribed Topics |
|-------|-------------------|
| PM Agent | `requirements`, `workflow` |
| Architect | `requirements`, `architecture`, `workflow` |
| Go Coder | `architecture`, `review`, `security`, `workflow` |
| Test Agent | `architecture`, `review`, `testing`, `workflow` |
| Security Agent | `code_change`, `security`, `workflow` |
| Reviewer | `code_change`, `testing`, `security`, `workflow` |
| Optimizer | `review`, `performance`, `workflow` |
| DevOps | `review`, `release`, `workflow` |

---

## Communication Patterns

### Pattern 1: Direct Query

```
Coder ──[query]──> Architect
Coder <──[response]── Architect
```

**Use case:** Coder cần clarify interface signature

**Command:** `@ask:arch "What should UserService return for not found?"`

### Pattern 2: Fan-out Notification

```
Security ──[notification]──> Coder
         └─[notification]──> Reviewer
         └─[notification]──> Test
```

**Use case:** Security phát hiện vulnerability, thông báo cho các agents liên quan

### Pattern 3: Collaboration Chain

```
Reviewer ──[collaboration]──> Coder ──[collaboration]──> Test
         <──[response]──────────────<──[response]────────
```

**Use case:** Reviewer phát hiện bug → Coder fix → Test update test

### Pattern 4: Consensus

```
Orchestrator ──[request_vote]──> All Agents
             <──[vote]─────────────────────
             ──[decision]──> All Agents
```

**Use case:** Quyết định có proceed với high-risk change không

---

## Agent Communication Commands

### In-workflow Commands

| Command | Description | Example |
|---------|-------------|---------|
| `@ask:{agent} "{question}"` | Query specific agent | `@ask:arch "Interface for UserRepo?"` |
| `@notify:{topic} "{message}"` | Send notification to topic | `@notify:security "New file created"` |
| `@request:{agent} "{task}"` | Request collaboration | `@request:test "Add error test"` |
| `@broadcast "{message}"` | Broadcast to all | `@broadcast "Phase change to testing"` |

### Agent-to-Agent Shortcuts

| From | To | Shortcut |
|------|-----|----------|
| Coder | Architect | `?arch` |
| Coder | Test | `?test` |
| Reviewer | Coder | `!fix` |
| Reviewer | Test | `!test` |
| Security | Coder | `!vuln` |
| Any | All | `!all` |

---

## Message Priority

| Level | Description | Timeout | Retry |
|-------|-------------|---------|-------|
| `critical` | Security, breaking changes | 1s | 3x |
| `high` | Review issues, build failures | 3s | 2x |
| `normal` | Standard queries | 5s | 1x |
| `low` | Notifications, info | 10s | 0 |

---

## State Management

### Message Queue State

```json
{
  "queue_id": "go-team-session-001",
  "messages": [
    {"id": "msg-001", "status": "delivered", "delivered_at": "..."},
    {"id": "msg-002", "status": "pending", "retry_count": 1},
    {"id": "msg-003", "status": "timeout", "error": "No response"}
  ],
  "statistics": {
    "total_messages": 45,
    "delivered": 42,
    "pending": 2,
    "failed": 1,
    "avg_response_time_ms": 1250
  }
}
```

### Agent Availability

```json
{
  "agents": {
    "pm-agent": {"status": "idle", "last_active": "..."},
    "architect-agent": {"status": "busy", "current_task": "..."},
    "go-coder-agent": {"status": "active", "current_file": "..."},
    "test-agent": {"status": "idle"},
    "security-agent": {"status": "scanning"},
    "reviewer-agent": {"status": "idle"},
    "optimizer-agent": {"status": "idle"},
    "devops-agent": {"status": "idle"}
  }
}
```

---

## Integration with Workflow

### Hook Points

1. **Pre-Step Hook:** Check pending messages before starting step
2. **Post-Step Hook:** Process and deliver queued messages
3. **Error Hook:** Broadcast errors to relevant agents
4. **Breakpoint Hook:** Allow cross-agent queries during breakpoints

### Example: Coder Queries Architect Mid-Implementation

```markdown
## During Step 4 (Implementation)

Coder is implementing UserService and needs clarification:

1. Coder sends query:
   ```
   @ask:arch "Should GetUser return (User, error) or (*User, error)?"
   ```

2. Message Bus routes to Architect Agent

3. Architect responds:
   ```
   Return (*User, error) - pointer for nil case when not found
   ```

4. Coder receives response and continues

5. Message logged for audit trail
```

---

## Error Handling

### Message Delivery Failed

```
⚠️  MESSAGE DELIVERY FAILED

Message: msg-003
From: reviewer-agent
To: test-agent
Error: Agent not available

Options:
- *retry:{msg-id}    → Retry delivery
- *queue:{msg-id}    → Queue for later
- *skip:{msg-id}     → Skip message
- *escalate          → Escalate to orchestrator
```

### Response Timeout

```
⚠️  RESPONSE TIMEOUT

Query: msg-001
To: architect-agent
Waited: 5000ms

Options:
- *retry-query       → Retry with extended timeout
- *ask-orchestrator  → Ask orchestrator to intervene
- *proceed           → Continue without answer
```

---

## Benefits

1. **Faster Issue Resolution:** Agents can collaborate directly without full workflow cycles
2. **Context Sharing:** Agents share relevant context automatically via topics
3. **Parallel Work:** Multiple agents can communicate simultaneously
4. **Audit Trail:** All messages logged for debugging and analysis
5. **Flexible Workflow:** Agents can dynamically adjust based on peer feedback

---

## Files

- [message-schemas.md](./message-schemas.md) - Detailed message schemas
- [topic-handlers.md](./topic-handlers.md) - Topic subscription handlers
- [agent-protocols.md](./agent-protocols.md) - Per-agent communication protocols
