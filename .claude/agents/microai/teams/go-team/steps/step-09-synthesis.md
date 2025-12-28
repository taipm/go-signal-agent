---
stepNumber: 9
nextStep: null
agent: orchestrator
hasBreakpoint: false
---

# Step 09: Final Synthesis

## STEP GOAL

Generate final session report, summarize all work done, and provide handoff documentation.

## EXECUTION SEQUENCE

### 1. Gather All Outputs

Collect from session state:
- Spec từ PM Agent
- Architecture từ Architect Agent
- Code files từ Coder Agent
- Test files từ Test Agent
- Review results từ Reviewer Agent
- Optimizations từ Optimizer Agent
- Release config từ DevOps Agent

### 2. Generate Session Summary

```markdown
# Go Team Session Report

## Project: {topic}
**Date:** {date}
**Duration:** {session_duration}

---

## Summary

{1-2 sentence description of what was built}

---

## Deliverables

### Specification
- User Stories: {count}
- Acceptance Criteria: {count}
- API Endpoints: {count}

### Architecture
- Pattern: {pattern_name}
- Packages: {count}
- Interfaces: {count}

### Implementation
| File | Purpose |
|------|---------|
| cmd/app/main.go | Entry point |
| internal/model/*.go | Domain models |
| internal/repo/*.go | Data access |
| internal/service/*.go | Business logic |
| internal/handler/*.go | HTTP handlers |

### Tests
- Test files: {count}
- Coverage: {percentage}%
- All passing: YES

### Quality Metrics
| Metric | Status |
|--------|--------|
| Build | ✅ PASS |
| Tests | ✅ PASS |
| Lint | ✅ CLEAN |
| Race Detection | ✅ FREE |
| Coverage | {X}% |

### Release Artifacts
- Dockerfile (scratch base, ~10MB)
- GitHub Actions CI/CD
- Makefile

---

## Quick Start

```bash
# Build
make build

# Run
make run

# Test
make test

# Docker
make docker
docker run -p 8080:8080 app
```

---

## Files Created

{list all files with paths}

---

## Next Steps (Recommendations)

1. {recommendation 1}
2. {recommendation 2}
3. {recommendation 3}

---

## Session Log

Full session log saved to:
`./docs/go-team/{date}-{topic}.md`
```

### 3. Save Session Log

Save full session history to:
`./docs/go-team/{date}-{topic-slug}.md`

Include:
- All agent interactions
- Observer interventions
- Decisions made
- Issues found and fixed

### 4. Display Final Output

```
═══════════════════════════════════════════════════════════
                    SESSION COMPLETE
═══════════════════════════════════════════════════════════

Project: {topic}

✅ Specification: Complete
✅ Architecture: Designed
✅ Implementation: Done
✅ Tests: Passing ({coverage}%)
✅ Review: Passed
✅ Optimization: Applied
✅ Release Config: Ready

Files Created: {count}
Session Log: ./docs/go-team/{date}-{topic}.md

───────────────────────────────────────────────────────────

Quick Start:
  make build && make run

Docker:
  docker build -t app . && docker run -p 8080:8080 app

───────────────────────────────────────────────────────────

Thank you for using Go Team! 🚀

═══════════════════════════════════════════════════════════
```

### 5. Cleanup

- Update workflow state to completed
- Mark all steps as done
- Save final metrics

## OUTPUT

```yaml
session_complete: true
final_report:
  topic: "{topic}"
  date: "{date}"
  files_created: {count}
  coverage: {percentage}
  all_checks_pass: true
  log_file: "./docs/go-team/{date}-{topic}.md"
```

## SUCCESS CRITERIA

- [ ] All outputs collected
- [ ] Session report generated
- [ ] Log file saved
- [ ] Final summary displayed
- [ ] Session marked complete

## WORKFLOW COMPLETE

This is the final step. Workflow ends here.
