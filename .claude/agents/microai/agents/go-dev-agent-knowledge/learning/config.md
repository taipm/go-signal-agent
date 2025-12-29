# Learning System Configuration

**Version:** 1.0.0
**Agent:** go-dev-agent

---

## Settings

### Capture Settings

| Setting | Value | Description |
|---------|-------|-------------|
| auto_capture | true | Automatically capture learnings after tasks |
| capture_threshold | medium | Minimum severity to auto-capture (low/medium/high/critical) |
| max_raw_learnings | 100 | Max items in raw/ before cleanup |
| raw_retention_days | 30 | Days to keep raw learnings |

### Pattern Extraction Settings

| Setting | Value | Description |
|---------|-------|-------------|
| min_occurrences | 3 | Min times a pattern must appear to escalate |
| confidence_threshold | medium | Min confidence to add to pending queue |
| auto_extract_reviews | true | Extract patterns from code reviews |

### Review Settings

| Setting | Value | Description |
|---------|-------|-------------|
| require_human_approval | true | Gate for knowledge updates (DO NOT DISABLE) |
| max_pending_items | 50 | Max items in pending queue |
| pending_retention_days | 90 | Days before pending items auto-archive |

---

## Learning Type Definitions

| Type | Description | Default Severity | mem-search Type |
|------|-------------|------------------|-----------------|
| bugfix | Root cause of a bug and its fix | high | bugfix |
| pattern | Good pattern to follow | medium | discovery |
| anti-pattern | Pattern to avoid | high | discovery |
| decision | Architectural/design decision | medium | decision |
| optimization | Performance improvement | low | change |

---

## Severity Levels

| Level | Description | Auto-Escalate to Pending |
|-------|-------------|--------------------------|
| critical | Production-breaking, security issues | Always |
| high | Significant bugs, performance degradation | Always |
| medium | Code quality, maintainability | If config allows |
| low | Minor improvements, style | Never (stays in raw) |

---

## Commands Reference

| Command | Description | Example |
|---------|-------------|---------|
| `*learn-capture` | Manually capture a learning | `*learn-capture` |
| `*learn-review` | Review pending queue | `*learn-review` |
| `*learn-approve:ID` | Approve learning | `*learn-approve:L001` |
| `*learn-reject:ID` | Reject learning | `*learn-reject:L001` |
| `*learn-modify:ID` | Modify before approve | `*learn-modify:L001` |
| `*learn-defer:ID` | Defer review | `*learn-defer:L001` |
| `*learn-status` | Show system status | `*learn-status` |
| `*learn-search:query` | Search learnings | `*learn-search:goroutine` |
| `*learn-cleanup` | Archive old items | `*learn-cleanup` |

---

## Directory Structure

```
learning/
├── config.md           # This file
├── raw/                # Daily learning captures (temporary)
│   └── YYYY-MM-DD-{session}.md
├── pending/            # Awaiting human review
│   ├── review-queue.md
│   └── pending-{id}.md
└── archive/            # Reviewed items
    └── YYYY-MM/
        └── {date}-{id}.md
```

---

## Integration Points

### With go-dev-agent.md

Learning triggers are embedded in the agent's workflow:
- After code review completion
- After bug fix
- After design decision
- After performance optimization

### With Knowledge Files

Approved learnings update these files:
- `09-learned-patterns.md` - Good patterns
- `10-learned-anti-patterns.md` - Patterns to avoid
- `11-project-decisions.md` - Design decisions

### With mem-search MCP

Learnings are also recorded via mem-search for cross-session search:
```
search(query="topic", obs_type="bugfix,discovery,decision")
```

---

## Changelog

| Date | Version | Changes |
|------|---------|---------|
| 2025-12-29 | 1.0.0 | Initial configuration |
