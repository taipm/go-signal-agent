# Mining Team

Team chuyên biệt gồm 4 agents tiếp cận vấn đề từ các góc độ tư duy khác nhau, phối hợp để khám phá những insights sâu, câu hỏi quan trọng, và risks tiềm ẩn mà thường bị bỏ qua trong quá trình phát triển.

## Overview

Mining Team cung cấp:
- **Deep Questioning** - Khám phá assumptions ẩn với Socratic method
- **Reverse Thinking** - Challenge mọi giả định với Inversion và Pre-mortem
- **Codebase Exploration** - Phân tích kỹ thuật với evidence từ code
- **Production Readiness** - Đánh giá từ góc nhìn SRE/DevOps
- **Synthesized Reports** - Tổng hợp insights với priority và recommendations

## Installation

**Portable Installation (Recommended):**

Copy toàn bộ folder vào project của bạn:
```bash
cp -r .claude/agents/microai/teams/mining-team /path/to/your/project/.claude/agents/microai/teams/
```

Sau đó copy skill command:
```bash
cp .claude/commands/mine.md /path/to/your/project/.claude/commands/
```

## Quick Start

```bash
# Invoke mining session với topic
/mine authentication system

# Hoặc không có topic (sẽ được hỏi)
/mine
```

## Team Members

| Agent | Persona | Role | Frameworks |
|-------|---------|------|------------|
| 🔮 `deep-question-agent` | Socrates | Deep Questions | Socratic, First Principles, 5 Whys |
| 🔄 `reverse-thinking-agent` | Contrarian | Challenge Assumptions | Inversion, Pre-mortem, Devil's Advocate |
| 🔍 `codebase-explorer-agent` | Sherlock | Code Analysis | Technical Analysis, Architecture Review |
| 🚀 `production-readiness-agent` | Ops | Production Ready | User Journey, Chaos Thinking, SLA |

## Workflow

```
Step 01: Initialize session, load project context
    ↓
Step 02: 🔮 Socrates - fundamental questions
    ↓    ← Observer can intervene
Step 03: 🔄 Contrarian - challenge assumptions
    ↓    ← Observer can intervene
Step 04: 🔍 Sherlock - technical deep-dive
    ↓    ← Observer can intervene
Step 05: 🚀 Ops - production concerns
    ↓    ← Observer can intervene
Step 06: Synthesize insights, generate report
```

## Observer Controls

Trong session, bạn có thể:

| Command | Effect |
|---------|--------|
| `[Enter]` | Continue với agent hiện tại |
| `@socrates: <msg>` | Inject message cho Socrates |
| `@contrarian: <msg>` | Inject message cho Contrarian |
| `@sherlock: <msg>` | Yêu cầu Sherlock investigate |
| `@ops: <msg>` | Hỏi Ops về production |
| `@guide: <msg>` | Redirect conversation |
| `*skip` | Bỏ qua phase hiện tại |
| `*synthesize` | Nhảy đến tổng hợp |
| `*exit` | Kết thúc session |

## Module Structure

```
mining-team/
├── workflow.md                    # Team orchestration
├── README.md                      # This file
├── module-plan-mining-team.md     # Planning document
├── agents/
│   ├── deep-question-agent.md     # 🔮 Socrates
│   ├── reverse-thinking-agent.md  # 🔄 Contrarian
│   ├── codebase-explorer-agent.md # 🔍 Sherlock
│   └── production-readiness-agent.md # 🚀 Ops
├── steps/
│   ├── step-01-init.md
│   ├── step-02-deep-mining.md
│   ├── step-03-reverse-mining.md
│   ├── step-04-code-mining.md
│   ├── step-05-production-check.md
│   └── step-06-synthesis.md
├── templates/
│   ├── mining-report-template.md
│   └── insights-index-template.md
└── logs/
    └── (session logs saved here)
```

## Features

### Project-aware Context
Auto-loads at session start:
- `README.md` - Project overview
- `docs/` - Documentation folder
- `go.mod`/`package.json` - Dependencies
- Recent git commits - Recent changes
- Previous mining logs - Past insights

### Session Logs
Mỗi session generates:
- `{date}-{topic}.md` - Full session transcript
- `{date}-{topic}-report.md` - Synthesized mining report
- `insights-index.md` - Accumulated key insights

### Adversarial Collaboration
Agents can challenge each other's findings:
```
Socrates → finds assumption
    ↓
Contrarian → challenges assumption
    ↓
Sherlock → provides code evidence
    ↓
Ops → tests against production reality
```

## Examples

### Example 1: Review Authentication System

```bash
/mine authentication system
```

Session will:
1. Load auth-related files
2. Socrates asks: "Why JWT instead of sessions?"
3. Contrarian challenges: "What if token refresh fails?"
4. Sherlock finds: Hardcoded secret in code
5. Ops assesses: Production risk of secret exposure

### Example 2: Pre-release Review

```bash
/mine production readiness for v2.0
```

Session focuses on deployment concerns, scalability, and real-world scenarios.

## Output

### Mining Report Highlights

```markdown
# Mining Report: authentication system

## Critical Findings
1. JWT secret hardcoded - DEPLOYMENT BLOCKER
2. No rate limiting on login endpoint - HIGH

## Recommendations
- Immediate: Move secret to env var
- Short-term: Add rate limiting
- Long-term: Implement secret rotation
```

## Development Status

- [x] Structure created
- [x] Agents implemented with personas
- [x] Workflow orchestration complete
- [x] Step files created
- [x] Report templates ready
- [x] Skill command created
- [ ] Testing in real projects
- [ ] Cross-session learning optimization

## Requirements

- Claude Code CLI
- Project with `.claude/` directory structure
- No external dependencies

## Author

Created by [ĐẠI CA TÀI BÉO] on 2025-12-28

---

**Module Code:** mining-team
**Category:** Technical / Analysis
**Type:** Standard Module (4 agents, 6 steps)
**Version:** 1.0.0

*Mine deep, discover more!* ⛏️
