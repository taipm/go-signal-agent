---
stepsCompleted:
  - step-01-init
workflowType: 'bmad-module-creation'
module_name: go-team
date: '2025-12-28'
user_name: '[ĐẠI CA TÀI BÉO]'
inputDocuments:
  - urd.md
---

# Module Plan: go-team

## Module Concept

**Module Name:** Go Team
**Module Code:** `go-team`
**Category:** Development / AI Coding Team
**Type:** Full Pipeline Module (8 agents, 1 orchestration workflow)

**Purpose Statement:**
AI coding team chuyên biệt cho Go development, bao gồm đầy đủ các vai trò từ requirements gathering đến release. Team tự động hóa quy trình phát triển với khả năng sinh code idiomatic Go, viết tests, review, optimize và package.

**Target Audience:**
- Primary: Go developers cần tăng tốc development workflow
- Secondary: Teams đang xây dựng Go microservices hoặc CLI tools

**Scope Definition:**

**In Scope:**
- Requirements gathering và user story creation
- System architecture design với Go patterns
- Code generation (handlers, services, repos, goroutines)
- Unit/integration test generation (table-driven)
- Code review (race conditions, goroutine leaks, style)
- Performance optimization (concurrency, memory)
- CI/CD setup và release packaging

**Out of Scope:**
- Non-Go languages
- Frontend development
- Database schema design (beyond interface)
- Production deployment (beyond CI/CD config)

**Success Criteria:**
- ✅ Build pass
- ✅ Test coverage ≥ 80%
- ✅ Lint sạch (golangci-lint)
- ✅ No race conditions (go test -race)
- ✅ Idiomatic Go code

---

## Agent Architecture

### 8 Agents (URD-faithful)

| # | Agent | File | Role | Tools |
|---|-------|------|------|-------|
| 1 | PM Agent | `agents/pm-agent.md` | Requirements, user stories, acceptance criteria | Read |
| 2 | Architect Agent | `agents/architect-agent.md` | System design, patterns, package structure | Read, Glob |
| 3 | Go Coder Agent | `agents/go-coder-agent.md` | Code generation, idiomatic Go | Read, Write, Edit, Bash |
| 4 | Test Agent | `agents/test-agent.md` | Unit/integration tests, table-driven | Read, Write, Edit, Bash |
| 5 | Reviewer Agent | `agents/reviewer-agent.md` | Code review, race conditions, style | Read, Bash |
| 6 | Optimizer Agent | `agents/optimizer-agent.md` | Performance, concurrency, benchmarks | Read, Bash |
| 7 | DevOps Agent | `agents/devops-agent.md` | Dockerfile, CI/CD, release | Read, Write, Bash |
| 8 | Orchestrator | (embedded in workflow.md) | Điều phối workflow, giao task | Task |

---

## Workflow Design

### Hybrid Orchestration

```
User Request
   ↓
┌─────────────────────────────────────────────────────────────┐
│  PM Agent → Spec                                            │
│     [BREAKPOINT: Observer can review/modify spec]          │
├─────────────────────────────────────────────────────────────┤
│  Architect → Design                                         │
│     [BREAKPOINT: Observer can review/modify design]        │
├─────────────────────────────────────────────────────────────┤
│  Go Coder → Code                                            │
│     ↓                                                       │
│  Test Agent → Tests                                         │
│     ↓                                                       │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  REVIEW LOOP (max 3 iterations)                         ││
│  │  Reviewer → Feedback                                    ││
│  │     ↓                                                   ││
│  │  Coder/Test → Fix                                       ││
│  │     ↓                                                   ││
│  │  [EXIT: All tests pass + lint clean]                    ││
│  └─────────────────────────────────────────────────────────┘│
│     [BREAKPOINT: Observer can review code]                 │
├─────────────────────────────────────────────────────────────┤
│  Optimizer → Improve                                        │
│     ↓                                                       │
│  DevOps → Package                                           │
└─────────────────────────────────────────────────────────────┘
   ↓
Final Output
```

### Observer Controls

| Command | Effect |
|---------|--------|
| `[Enter]` | Continue to next step |
| `*pause` | Pause workflow |
| `*skip-to:<step>` | Jump to specific step |
| `*exit` | End session |
| `@agent: <message>` | Inject message to specific agent |

---

## Step Files

| Step | File | Purpose | Agent |
|------|------|---------|-------|
| 01 | `step-01-init.md` | Initialize session, load context | Orchestrator |
| 02 | `step-02-requirements.md` | Gather requirements, create spec | PM Agent |
| 03 | `step-03-architecture.md` | Design system architecture | Architect Agent |
| 04 | `step-04-implementation.md` | Generate Go code | Go Coder Agent |
| 05 | `step-05-testing.md` | Generate tests | Test Agent |
| 06 | `step-06-review-loop.md` | Review and fix loop | Reviewer + Coder |
| 07 | `step-07-optimization.md` | Performance optimization | Optimizer Agent |
| 08 | `step-08-release.md` | CI/CD and packaging | DevOps Agent |
| 09 | `step-09-synthesis.md` | Final report | Orchestrator |

---

## Templates

| Template | Purpose |
|----------|---------|
| `spec-template.md` | PM output format (user stories, AC) |
| `architecture-template.md` | Architect output (diagrams, structure) |
| `code-review-template.md` | Reviewer report format |
| `release-notes-template.md` | DevOps release notes |

---

## Go Project Structure (Mandated)

```
/cmd/app/main.go
/internal/
   handler/
   service/
   repo/
   model/
   middleware/
/pkg/
/configs/
/tests/
/go.mod
```

**Principles:**
- Interface ở layer trên
- Dependency injection thủ công
- Không circular import
- Context everywhere
- Error wrapping with fmt.Errorf

---

## Tool Integration

### Build & Test
```bash
go build ./...
go test ./... -race -cover
go vet ./...
golangci-lint run
```

### Benchmarking (Optional)
```bash
go test -bench=. -benchmem
```

---

## Session State (YAML)

```yaml
go_team_state:
  topic: ""
  phase: "init"
  current_agent: null
  iteration_count: 0
  max_iterations: 3
  outputs:
    spec: null
    architecture: null
    code_files: []
    test_files: []
    review_comments: []
    optimizations: []
    release_config: null
  metrics:
    build_pass: false
    test_coverage: 0
    lint_clean: false
    race_free: false
```

---

## File Structure

```
.claude/agents/microai/teams/go-team/
├── workflow.md                    # Team orchestration
├── module-plan-go-team.md         # This file
├── urd.md                         # Original requirements
├── agents/
│   ├── pm-agent.md
│   ├── architect-agent.md
│   ├── go-coder-agent.md
│   ├── test-agent.md
│   ├── reviewer-agent.md
│   ├── optimizer-agent.md
│   └── devops-agent.md
├── steps/
│   ├── step-01-init.md
│   ├── step-02-requirements.md
│   ├── step-03-architecture.md
│   ├── step-04-implementation.md
│   ├── step-05-testing.md
│   ├── step-06-review-loop.md
│   ├── step-07-optimization.md
│   ├── step-08-release.md
│   └── step-09-synthesis.md
├── templates/
│   ├── spec-template.md
│   ├── architecture-template.md
│   ├── code-review-template.md
│   └── release-notes-template.md
└── logs/
    └── (session logs saved here)
```

---

## Next Steps

- [x] Define module concept ✅
- [x] Create directory structure ✅
- [ ] Create 7 agent files
- [ ] Create workflow.md
- [ ] Create 9 step files
- [ ] Create 4 template files
- [ ] Create skill command `/go-team`

---

## Legacy Reference

This module follows:
- URD blueprint từ `urd.md`
- MicroAI team structure từ `dev-user` và `mining-team`
- BMAD module patterns

Unique features:
- Hybrid workflow với breakpoints
- Review loop với iteration limits
- Tool integration (go build, test, lint)
- KPI tracking (coverage, lint, race)
