---
session_date: "{{date}}"
session_subject: "{{subject}}"
participants:
  - Solo Developer
  - End User
  - Observer: {{observer_name}}
total_turns: {{turn_count}}
duration: "{{start_time}} - {{end_time}}"
status: "{{status}}"
output_story: "{{story_status}}"
---

# Meeting Minutes: {{subject}}

## Session Overview

| Field | Value |
|-------|-------|
| **Date** | {{date}} |
| **Topic** | {{subject}} |
| **Duration** | {{start_time}} - {{end_time}} |
| **Total Turns** | {{turn_count}} |
| **Status** | {{status}} |
| **Participants** | Solo Developer, End User, Observer ({{observer_name}}) |

---

## Executive Summary

{{summary}}

---

## Key Decisions

{{#each decisions}}
### Decision {{@index_plus_one}}: {{this.topic}}

| Aspect | Details |
|--------|---------|
| **Decision** | {{this.decision}} |
| **Made at** | Turn {{this.turn}} |
| **Rationale** | {{this.rationale}} |

{{/each}}

---

## Dialogue Transcript

{{#each dialogue_history}}
### Turn {{this.turn}} | {{this.speaker_icon}} {{this.speaker}}

**Phase:** {{this.phase}} | **Time:** {{this.timestamp}}

{{this.message}}

{{#if this.intervention}}
> *Observer Intervention*
{{/if}}

{{#if this.key_points}}
**Key Points:**
{{#each this.key_points}}
- {{this}}
{{/each}}
{{/if}}

---

{{/each}}

---

## Requirements Captured

### Confirmed Requirements
{{#each confirmed_requirements}}
- ✅ {{this}}
{{/each}}

### Open Questions
{{#each open_questions}}
- ❓ {{this}}
{{/each}}

### Assumptions Made
{{#each assumptions}}
- 💭 {{this}}
{{/each}}

---

## Scope Agreement

### In Scope
{{#each in_scope}}
- {{this}}
{{/each}}

### Out of Scope (Deferred)
{{#each out_of_scope}}
- {{this.item}} — *Reason: {{this.reason}}*
{{/each}}

---

## Final Deliverable

{{#if story_finalized}}

### User Story: {{story_title}}

**As a** {{story.as_a}},
**I want** {{story.i_want}},
**So that** {{story.so_that}}.

#### Acceptance Criteria

{{#each story.acceptance_criteria}}
**AC{{@index_plus_one}}: {{this.title}}**
- Given: {{this.given}}
- When: {{this.when}}
- Then: {{this.then}}

{{/each}}

**Status:** ✅ Signed off by End User

{{else}}

### No Finalized Story

Session ended before story could be completed.

**Reason:** {{incomplete_reason}}

**Progress Made:**
{{partial_progress}}

{{/if}}

---

## Action Items

{{#each action_items}}
- [ ] {{this.action}} — *Owner: {{this.owner}}*
{{/each}}

---

## Next Steps

{{#each next_steps}}
{{@index_plus_one}}. {{this}}
{{/each}}

---

## Session Statistics

| Metric | Value |
|--------|-------|
| Total turns | {{turn_count}} |
| Solo Dev turns | {{dev_turns}} |
| EndUser turns | {{user_turns}} |
| Observer interventions | {{observer_interventions}} |
| Questions asked | {{questions_asked}} |
| Decisions made | {{decisions_count}} |
| Items deferred | {{deferred_count}} |

### Phase Breakdown

| Phase | Turns |
|-------|-------|
| Requirements | {{requirements_turns}} |
| Clarification | {{clarification_turns}} |
| Negotiation | {{negotiation_turns}} |
| Synthesis | {{synthesis_turns}} |

---

## Notes

{{additional_notes}}

---

*Meeting minutes generated automatically from dev-user team session*
*Date: {{date}} | File: {{filename}}*
