---
stepsCompleted:
  - step-01-init
  - step-02-concept
  - step-03-components
  - step-04-structure
  - step-05-config
  - step-06-agents
  - step-07-workflows
  - step-08-installer
  - step-09-documentation
  - step-10-complete
workflowType: 'bmad-module-creation'
module_name: mining-team
date: '2025-12-28'
user_name: '[ĐẠI CA TÀI BÉO]'
inputDocuments: []
status: COMPLETED
---

# Module Plan: mining-team

## Module Concept

**Module Name:** Mining Team
**Module Code:** `mining-team`
**Category:** Technical / Analysis
**Type:** Standard Module (4 agents, 1 orchestration workflow)

**Purpose Statement:**
Team chuyên biệt gồm 4 agents tiếp cận vấn đề từ các góc độ tư duy khác nhau, phối hợp để khám phá những insights sâu, câu hỏi quan trọng, và risks tiềm ẩn mà thường bị bỏ qua trong quá trình phát triển.

**Target Audience:**
- Primary: Developers và Tech Leads cần deep review trước major releases
- Secondary: Teams đang onboard vào codebase mới hoặc audit projects

**Scope Definition:**

**In Scope:**
- Deep questioning với multiple thinking frameworks
- Reverse/contrarian thinking để challenge assumptions
- Codebase exploration và technical analysis
- Production readiness assessment
- Synthesized mining reports với actionable insights

**Out of Scope:**
- Actual code implementation/fixes
- Automated testing execution
- CI/CD pipeline management
- Performance benchmarking

**Success Criteria:**
- Phát hiện ít nhất 1 insight mà sẽ MISSED nếu không có multi-agent approach
- Insights actionable và prioritized
- Team portable across different projects

---

## Memory & Context Architecture

### Approach: Project-aware Context + Session Logs (D + B)

```
┌─────────────────────────────────────────────────────────────┐
│                    MINING SESSION                           │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │   Project   │    │   Shared    │    │   Session   │     │
│  │   Context   │───▶│   Memory    │───▶│    Logs     │     │
│  │   Loader    │    │   (Runtime) │    │  (Persist)  │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│        │                   │                  │             │
│        ▼                   ▼                  ▼             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │     4 MINING AGENTS (Adversarial Collaboration)        ││
│  │  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐              ││
│  │  │ Deep  │◀┼▶│Reverse│◀┼▶│ Code  │◀┼▶│ Prod  │              ││
│  │  │Question│ │Think  │ │Explorer│ │Ready │              ││
│  │  └───────┘ └───────┘ └───────┘ └───────┘              ││
│  │     Each agent can CHALLENGE previous agent's findings ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Context Sources (Auto-loaded)

| Source | Description | Used By |
|--------|-------------|---------|
| `README.md` | Project overview | All agents |
| `docs/` | Documentation folder | All agents |
| `.claude/` | Claude configs, existing agents | Codebase Explorer |
| `go.mod`/`package.json` | Dependencies | Codebase Explorer, Prod Ready |
| Recent git commits | Recent changes | All agents |
| Previous mining logs | Past insights | All agents |

### Session Logs Structure

```
.claude/agents/microai/teams/mining-team/logs/
├── {YYYY-MM-DD}-{topic-slug}.md       # Full session transcript
├── {YYYY-MM-DD}-{topic-slug}-report.md # Synthesized report
└── insights-index.md                   # Accumulated key insights
```

### Insights Accumulation (Party Mode Insight)

**YAML State in Workflow:**
```yaml
mining_state:
  topic: ""
  phase: "init"
  accumulated_insights:
    deep_questions: []
    challenges: []
    technical_findings: []
    production_concerns: []
  cross_references: []  # Links between insights from different agents
```

---

## Component Architecture

### Agents (4 planned) - Standalone Files

| # | Agent | Type | Role | Thinking Frameworks | Tools |
|---|-------|------|------|---------------------|-------|
| 1 | `deep-question-agent` | Specialist | Đào sâu fundamental questions | Socratic, First Principles, 5 Whys | Read, Glob |
| 2 | `reverse-thinking-agent` | Specialist | Challenge assumptions, tư duy ngược | Inversion, Pre-mortem, Devil's Advocate | Read |
| 3 | `codebase-explorer-agent` | Specialist | Khám phá và phân tích mã nguồn | Technical Analysis, Architecture Review | Read, Grep, Glob, Bash, LSP |
| 4 | `production-readiness-agent` | Specialist | Góc nhìn production/end-user | User Journey, Deployment Checklist, Chaos Thinking | Read, Glob |

**Design Decision:** Agents are STANDALONE files (not embedded in steps) for reusability across other workflows.

### Workflow (1 main orchestrator)

**`workflow.md`** - Mining Session Orchestrator
- **Type:** Interactive (turn-based dialogue with observer)
- **Pattern:** Sequential phases with accumulated insights
- **Key feature:** Adversarial collaboration - agents can challenge each other

### Steps (6 files)

| Step | File | Purpose | Agent |
|------|------|---------|-------|
| 01 | `step-01-init.md` | Initialize session, load context | Orchestrator |
| 02 | `step-02-deep-mining.md` | Deep questioning phase | deep-question-agent |
| 03 | `step-03-reverse-mining.md` | Reverse thinking phase | reverse-thinking-agent |
| 04 | `step-04-code-mining.md` | Codebase exploration | codebase-explorer-agent |
| 05 | `step-05-production-check.md` | Production readiness | production-readiness-agent |
| 06 | `step-06-synthesis.md` | Combine & report | Orchestrator |

### Templates (2 files)

| Template | Purpose |
|----------|---------|
| `mining-report-template.md` | Final session report format |
| `insights-index-template.md` | Cross-session insights accumulator |

### Tasks

**Decision:** NO separate task files. All logic embedded in workflow state (YAML) per Wendy's recommendation.

---

## Agent Interaction Model

### Adversarial Collaboration (Party Mode Insight)

```
Deep Question → finds ASSUMPTION X
       ↓
Reverse Thinking → CHALLENGES assumption X with counter-evidence
       ↓
Codebase Explorer → provides CODE EVIDENCE for/against
       ↓
Production Ready → tests against REAL-WORLD usage
       ↓
Synthesis → resolves conflicts, prioritizes findings
```

### Insights Flow

Each step:
1. **READ** accumulated insights from previous phases
2. **CHALLENGE** or **BUILD ON** previous findings
3. **ADD** new insights with cross-references
4. **PASS** forward to next phase

---

## Development Priority

### Phase 1 (MVP)

- [ ] `workflow.md` - Core orchestration with state management
- [ ] 4 agent files với basic personas và thinking frameworks
- [ ] 6 step files với core logic
- [ ] `mining-report-template.md`
- [ ] Skill command `/mine` hoặc `/mining`

### Phase 2 (Enhancement)

- [ ] `insights-index-template.md` - Cross-session learning
- [ ] Advanced context loading (git history analysis)
- [ ] Agent-to-agent direct challenges
- [ ] Integration với existing agents (e.g., go-review-linus)
- [ ] Custom thinking frameworks per project type

---

## File Structure

```
.claude/agents/microai/teams/mining-team/
├── workflow.md                         # Team orchestration
├── module-plan-mining-team.md          # This file
├── agents/
│   ├── deep-question-agent.md          # Standalone, reusable
│   ├── reverse-thinking-agent.md       # Standalone, reusable
│   ├── codebase-explorer-agent.md      # Standalone, reusable
│   └── production-readiness-agent.md   # Standalone, reusable
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
    └── (session logs will be saved here)
```

---

## Party Mode Insights Applied

| Insight | Source | Applied To |
|---------|--------|------------|
| Accumulated insights flow forward | BMad Master | Workflow state YAML |
| Standalone agent files for reusability | Morgan | Agent architecture |
| State in workflow YAML, no separate tasks | Wendy | Removed task files |
| Success = finding missed insights | Dr. Quinn | Updated success criteria |
| Adversarial collaboration | Carson | Agent interaction model |

---

## Next Steps

- [x] Define module concept (step-02) ✅
- [x] Design component details (step-03) ✅
- [x] Create module structure (step-04) ✅
- [x] Create each agent's persona and behaviors ✅
- [x] Create workflow orchestration logic ✅
- [x] Define step files for each mining phase ✅
- [x] Create output templates ✅
- [x] Create skill command for invocation ✅
- [x] Documentation complete ✅
- [x] Module creation COMPLETED ✅

---

## Legacy Reference

This module follows the MicroAI team structure established by `dev-user` team, adapted for multi-perspective "mining" workflow with:
- Enhanced memory/context architecture (D+B approach)
- Adversarial collaboration pattern
- Standalone reusable agents
- YAML-based state management
