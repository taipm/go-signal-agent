---
description: Invoke Mining Team để đào sâu insights từ nhiều góc độ (project)
---

# Mining Team Session

Load và execute workflow từ `.claude/agents/microai/teams/mining-team/workflow.md`

## Quick Start

Bạn sẽ điều phối một mining session với 4 specialized agents:

- 🔮 **Socrates** - Deep Questions (First Principles, Socratic Method)
- 🔄 **Contrarian** - Reverse Thinking (Inversion, Pre-mortem)
- 🔍 **Sherlock** - Codebase Explorer (Technical Analysis)
- 🚀 **Ops** - Production Readiness (Real-world Scenarios)

## Instructions

1. Read workflow file at `.claude/agents/microai/teams/mining-team/workflow.md`
2. Follow the workflow architecture exactly
3. Load step files sequentially from `./steps/`
4. Adopt each agent's persona from `./agents/` when their phase is active
5. Allow observer (user) to intervene at any point
6. Generate mining report at the end

## Session Topic

If user provided a topic after `/mine`, use that as the session topic.
Otherwise, ask user for the topic to mine.

## Observer Controls

During the session, user can:
- `[Enter]` - Continue with current agent
- `@socrates:`, `@contrarian:`, `@sherlock:`, `@ops:` - Inject message to specific agent
- `@guide:` - Redirect conversation
- `*skip` - Skip current phase
- `*synthesize` - Jump to synthesis
- `*exit` - End session

## Output

Session logs and mining report will be saved to:
`.claude/agents/microai/teams/mining-team/logs/`

---

Begin by loading the workflow and executing Step 1: Session Initialization.
