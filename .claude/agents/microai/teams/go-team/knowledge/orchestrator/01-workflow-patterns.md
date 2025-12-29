# Orchestrator Workflow Patterns

## Workflow Selection Matrix

### Request Type Detection

```
┌─────────────────────────────────────────────────────────────────┐
│                    REQUEST CLASSIFICATION                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Keywords Analysis:                                              │
│  ┌─────────────────┬───────────────────────────────────────┐    │
│  │ "add", "create" │ → new_feature                         │    │
│  │ "fix", "bug"    │ → bugfix                              │    │
│  │ "improve"       │ → enhancement                         │    │
│  │ "refactor"      │ → refactor                            │    │
│  │ "optimize"      │ → performance                         │    │
│  │ "secure", "vuln"│ → security_fix                        │    │
│  │ "deploy", "ci"  │ → devops                              │    │
│  └─────────────────┴───────────────────────────────────────┘    │
│                                                                  │
│  Complexity Assessment:                                          │
│  - LOW: Single file change, clear scope                         │
│  - MEDIUM: Multiple files, defined boundaries                   │
│  - HIGH: System-wide, architectural decisions                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Workflow Templates

### 1. Full Pipeline (new_feature - HIGH complexity)

```
Step 01: Init
    ↓
Step 01b: Codebase Analysis (if existing code)
    ↓
Step 02: Requirements → [BREAKPOINT]
    ↓
Step 03: Architecture → [BREAKPOINT]
    ↓
Step 04: Implementation
    ↓
Step 05: Testing
    ↓
Step 05b: Security → [SECURITY GATE]
    ↓
Step 06: Review Loop (max 3 iterations) → [BREAKPOINT]
    ↓
Step 07: Optimization
    ↓
Step 08: Release
    ↓
Step 09: Synthesis
```

**Duration:** 30-60 minutes
**Agents:** All 8 agents

---

### 2. Standard Pipeline (new_feature - MEDIUM complexity)

```
Step 01: Init
    ↓
Step 02: Requirements (condensed)
    ↓
Step 03: Architecture (condensed)
    ↓
Step 04: Implementation
    ↓
Step 05: Testing
    ↓
Step 06: Review Loop (max 2 iterations)
    ↓
Step 09: Synthesis
```

**Duration:** 15-30 minutes
**Agents:** PM, Architect, Coder, Test, Reviewer

---

### 3. Quick Fix Route (bugfix)

```
Step 01: Init (quick)
    ↓
Coder: Fix bug
    ↓
Test: Update/add tests
    ↓
Reviewer: Quick review
    ↓
Step 09: Synthesis (condensed)
```

**Duration:** 5-15 minutes
**Agents:** Coder, Test, Reviewer

---

### 4. Security Fix Route (security_fix)

```
Step 01: Init (quick)
    ↓
Security: Deep analysis
    ↓
Coder: Apply fix
    ↓
Security: Verify fix
    ↓
Reviewer: Review
    ↓
Step 09: Synthesis
```

**Duration:** 10-20 minutes
**Agents:** Security, Coder, Reviewer

---

### 5. Refactor Route (refactor)

```
Step 01: Init
    ↓
Architect: Review current structure, propose changes
    ↓
Coder: Apply refactoring
    ↓
Test: Ensure tests still pass
    ↓
Reviewer: Review changes
    ↓
Step 09: Synthesis
```

**Duration:** 15-30 minutes
**Agents:** Architect, Coder, Test, Reviewer

---

### 6. Performance Route (performance)

```
Step 01: Init
    ↓
Optimizer: Profile and identify bottlenecks
    ↓
Coder: Apply optimizations
    ↓
Optimizer: Benchmark before/after
    ↓
Reviewer: Review
    ↓
Step 09: Synthesis
```

**Duration:** 15-25 minutes
**Agents:** Optimizer, Coder, Reviewer

---

### 7. DevOps Only Route (devops)

```
Step 01: Init (quick)
    ↓
DevOps: Create/update CI, Docker, Makefile
    ↓
Step 09: Synthesis (condensed)
```

**Duration:** 5-10 minutes
**Agents:** DevOps only

---

## Agent Coordination Patterns

### Sequential Pattern (Default)

```
Agent A → Output → Agent B → Output → Agent C
```

Use when: Output from one agent is input to next.

### Parallel Pattern

```
        ┌→ Agent A ─┐
Input ──┤           ├→ Aggregator → Output
        └→ Agent B ─┘
```

Use when: Independent tasks that can run simultaneously.

**Example:** Testing + Security scanning

### Loop Pattern

```
        ┌──────────────┐
        ↓              │
Agent A → Agent B → Decision
                       │ (if fail)
                       └──────────┘
```

Use when: Iterative improvement needed.

**Example:** Review Loop (Reviewer → Coder → Reviewer)

### Conditional Pattern

```
Input → Decision
          │
    ┌─────┴─────┐
    ↓           ↓
  Route A    Route B
```

Use when: Different paths based on conditions.

**Example:** Security severity determines next step.

---

## Handoff Protocol

### Pre-Handoff Checklist

```yaml
before_handoff:
  - [ ] Previous step outputs validated
  - [ ] Required context prepared
  - [ ] Next agent's prerequisites met
  - [ ] Checkpoint created (if enabled)
  - [ ] Kanban updated (if enabled)
```

### Handoff Message Structure

```markdown
## HANDOFF: {from_agent} → {to_agent}

### Context
{relevant_context}

### Inputs Received
{list_of_inputs}

### Expected Outputs
{list_of_expected_outputs}

### Constraints
{any_constraints}

### Notes from Previous Agent
{notes}
```

### Post-Handoff Validation

```yaml
after_handoff:
  - [ ] Agent acknowledged receipt
  - [ ] Inputs processed correctly
  - [ ] No blocking issues
  - [ ] Progress tracking updated
```

---

## Decision Trees

### When to Pause for User

```
Is user input required?
├── YES → Pause and ask
└── NO
    └── Is it a breakpoint step?
        ├── YES → Pause for review
        └── NO
            └── Is confidence < 80%?
                ├── YES → Pause and confirm
                └── NO → Continue
```

### When to Rollback

```
Did step fail?
├── YES
│   └── Is it retry-able?
│       ├── YES
│       │   └── Retries < max?
│       │       ├── YES → Retry
│       │       └── NO → Rollback
│       └── NO → Rollback
└── NO → Continue
```

### When to Escalate

```
Issue severity?
├── LOW → Auto-handle
├── MEDIUM → Log and continue
├── HIGH → Pause and notify user
└── CRITICAL → Stop and escalate immediately
```

---

## Communication Patterns

### Broadcast

```yaml
type: broadcast
from: orchestrator
to: all_agents
message: "Session configuration updated"
```

Use for: Global announcements, config changes.

### Direct Message

```yaml
type: direct
from: orchestrator
to: security-agent
message: "Urgent: Scan this file for vulnerabilities"
```

Use for: Targeted requests, urgent issues.

### Request-Response

```yaml
# Request
type: request
from: orchestrator
to: architect-agent
message: "Is this design compatible with existing patterns?"
expect_response: true
timeout: 30s

# Response
type: response
from: architect-agent
to: orchestrator
in_reply_to: {request_id}
message: "Yes, compatible with Clean Architecture pattern"
```

Use for: Queries that need answers before proceeding.

### Notification

```yaml
type: notification
from: test-agent
to: orchestrator
message: "Tests complete: 85% coverage"
priority: normal
```

Use for: Status updates, non-blocking info.

---

## Error Recovery Patterns

### Retry with Backoff

```
Attempt 1 → Fail → Wait 1s
Attempt 2 → Fail → Wait 2s
Attempt 3 → Fail → Wait 4s
Attempt 4 → Fail → Escalate
```

### Fallback to Alternative

```
Primary Agent → Fail
    ↓
Check for alternative agent
    ↓
Alternative Agent → Process
    ↓
Continue workflow
```

### Graceful Degradation

```
Full feature → Fail
    ↓
Reduce scope
    ↓
Core feature only → Success
    ↓
Document limitations
```

---

## Metrics Collection

### Per-Step Metrics

```yaml
step_metrics:
  step_id: "step-04"
  agent: "go-coder-agent"
  started_at: "2025-12-29T00:30:00Z"
  completed_at: "2025-12-29T00:38:00Z"
  duration_seconds: 480
  tokens:
    input: 5000
    output: 3000
    cached: 2000
  files:
    created: 8
    modified: 2
  status: "success"
```

### Session Aggregates

```yaml
session_metrics:
  total_duration: 1800  # 30 minutes
  total_tokens: 45000
  estimated_cost: 0.55
  steps_completed: 9
  steps_skipped: 0
  iterations: 2
  agents_used:
    - pm-agent
    - architect-agent
    - go-coder-agent
    - test-agent
    - security-agent
    - reviewer-agent
    - devops-agent
  quality:
    coverage: 85
    lint_clean: true
    race_free: true
    security_passed: true
```

---

## Best Practices

### 1. Context Minimization

- Only pass relevant context to each agent
- Summarize large outputs before handoff
- Use references instead of full content when possible

### 2. Fail Fast

- Validate inputs early
- Check prerequisites before starting
- Detect blocking issues ASAP

### 3. Progress Visibility

- Update status frequently
- Show meaningful progress indicators
- Provide estimates when possible

### 4. Graceful Handling

- Always have a fallback plan
- Save state before risky operations
- Enable recovery from any point
