# Go-Team Knowledge Base

**Version:** 1.0.0

Kho kiến thức cho các agents trong go-team. Mỗi agent có knowledge folder riêng, plus shared knowledge chung.

---

## Cấu Trúc

```
knowledge/
├── shared/                      # Kiến thức dùng chung cho TẤT CẢ agents
│   ├── 01-go-fundamentals.md    # Go basics: context, errors, interfaces
│   ├── 02-error-patterns.md     # Error handling patterns
│   ├── 03-logging-standards.md  # Structured logging với slog/zerolog
│   └── 04-testing-patterns.md   # Table-driven tests, testify
│
├── pm/                          # PM Agent knowledge
│   └── 01-user-stories.md       # User story format, INVEST, BDD
│
├── architect/                   # Architect Agent knowledge
│   └── 01-architecture-patterns.md  # Clean, Hexagonal, Layered
│
├── coder/                       # Go Coder Agent knowledge
│   └── 01-implementation-order.md   # Layer-by-layer implementation
│
├── test/                        # Test Agent knowledge
│   └── 01-test-strategies.md    # Unit, Integration, Coverage
│
├── security/                    # Security Agent knowledge
│   └── 01-owasp-top10.md        # OWASP vulnerabilities
│
├── reviewer/                    # Reviewer Agent knowledge
│   └── 01-review-checklist.md   # Code review checklist
│
├── optimizer/                   # Optimizer Agent knowledge
│   └── 01-performance-patterns.md   # Memory, concurrency optimization
│
└── devops/                      # DevOps Agent knowledge
    └── 01-docker-cicd.md        # Docker, GitHub Actions, Makefile
```

---

## Cách Sử Dụng

### Trong Agent Definition

```yaml
---
name: go-coder-agent
model: sonnet
knowledge:
  shared:
    - ../knowledge/shared/01-go-fundamentals.md
    - ../knowledge/shared/02-error-patterns.md
  specific:
    - ../knowledge/coder/01-implementation-order.md
---
```

### Reference Format

Mỗi knowledge file có:
1. **TL;DR** - Summary nhanh
2. **Code Examples** - ✅ CORRECT vs ❌ WRONG patterns
3. **Quick Reference** - Bảng tra cứu nhanh
4. **Related Knowledge** - Links đến files liên quan

---

## Knowledge Matrix

| Agent | Shared | Specific |
|-------|--------|----------|
| PM | fundamentals | user-stories |
| Architect | fundamentals, errors | architecture-patterns |
| Coder | ALL shared | implementation-order |
| Test | fundamentals, testing | test-strategies |
| Security | fundamentals, errors | owasp-top10 |
| Reviewer | ALL shared | review-checklist |
| Optimizer | fundamentals | performance-patterns |
| DevOps | fundamentals | docker-cicd |

---

## Adding New Knowledge

1. Tạo file trong folder tương ứng
2. Follow format: `##-topic-name.md`
3. Include sections: TL;DR, Examples, Quick Reference
4. Update agent definition để reference

### Template

```markdown
# Topic Name - Agent Knowledge

**Version:** 1.0.0
**Agent:** Agent Name

---

## TL;DR

- Key point 1
- Key point 2
- Key point 3

---

## 1. Section One

### Pattern/Concept

```go
// ❌ WRONG
bad code example

// ✅ CORRECT
good code example
```

---

## Quick Reference

| Item | Description |
|------|-------------|
| Key1 | Value1 |

---

## Related Knowledge

- [link1](./path.md) - Description
```

---

## Maintenance

- Review quarterly
- Update khi Go version mới
- Add patterns từ code reviews
- Remove deprecated patterns
