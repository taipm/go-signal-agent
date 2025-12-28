# Step 04: Story Synthesis

## Objective
Synthesize dialogue thành formal User Story với Acceptance Criteria.

## Actions

### 1. Extract Key Information from Dialogue

Analyze dialogue_history để extract:

```yaml
extracted_info:
  user_persona: ""           # Who is the user?
  capability: ""             # What do they want?
  business_value: ""         # Why do they want it?
  acceptance_criteria: []    # List of criteria
  in_scope: []               # Agreed scope
  out_of_scope: []           # Deferred items
  technical_notes: ""        # Developer observations
  dependencies: []           # Identified dependencies
  complexity: ""             # simple | medium | complex
```

### 2. Solo Dev Proposes Story

```
[Story Proposal - Solo Dev] 👨‍💻

Dựa trên discussion của chúng ta, đây là User Story tôi đề xuất:

═══════════════════════════════════════════════════════════════

## User Story: {story_title}

**As a** {user_persona},
**I want** {capability},
**So that** {business_value}.

---

### Acceptance Criteria

**AC1: {criterion_1_title}**
- **Given:** {precondition}
- **When:** {action}
- **Then:** {expected_result}

**AC2: {criterion_2_title}**
- **Given:** {precondition}
- **When:** {action}
- **Then:** {expected_result}

{...more criteria as needed...}

---

### Scope Notes

**In Scope:**
- {item_1}
- {item_2}

**Out of Scope (Deferred):**
- {deferred_1}
- {deferred_2}

---

### Technical Notes
{observations_and_considerations}

---

### Complexity: {complexity}

═══════════════════════════════════════════════════════════════

EndUser, bạn review và confirm story này nhé?

───────────────────────────────────────────────────────────────
[Enter] continue | @user: modify | *exit
>
```

### 3. EndUser Review

**If Accepted:**
```
[Story Review - EndUser] 👤

Để tôi review từng criterion...

✓ AC1: {title} - Đúng, đây là điều tôi cần
✓ AC2: {title} - Chính xác
✓ Scope notes: Đồng ý với in/out scope

**ACCEPTED** ✅

Tôi sign off story này. Nó capture chính xác những gì
chúng ta đã thảo luận.

───────────────────────────────────────────────────────────────
```

**If Changes Needed:**
```
[Story Review - EndUser] 👤

Review:
✓ AC1: OK
✗ AC2: Cần điều chỉnh - {reason}
✓ AC3: OK

**CHANGES REQUESTED** 📝

Có thể update AC2 để: {suggested_change}?

───────────────────────────────────────────────────────────────
[Enter] continue | @dev: respond | *skip
>
```

### 4. Iterate if Needed

Solo Dev revises → EndUser reviews → Repeat until accepted

### 5. Finalize Story

Once accepted, format final story:

```yaml
final_story:
  story_id: "STORY-{timestamp}"
  title: "{title}"
  created_date: "{date}"
  session_subject: "{subject}"
  status: "ready-for-implementation"

  definition:
    as_a: "{persona}"
    i_want: "{capability}"
    so_that: "{value}"

  acceptance_criteria:
    - id: "AC1"
      title: "{title}"
      given: "{given}"
      when: "{when}"
      then: "{then}"
    # ... more criteria

  scope:
    in_scope: []
    out_of_scope: []

  technical_notes: ""
  dependencies: []
  complexity: ""

  sign_off:
    enduser: true
    solo_dev: true
```

## Story Quality Checklist

Before finalizing, verify:

```
□ User persona is specific and identifiable
□ Capability is clear and actionable
□ Business value is measurable or observable
□ Each AC has clear Given/When/Then
□ ACs are testable and verifiable
□ Scope boundaries are explicit
□ No ambiguous language ("should", "might", "could")
□ Dependencies identified
□ Complexity estimated
```

## Transition
→ Proceed to Step 05: Session Close

## State After Completion
```yaml
stepsCompleted: ["step-01-session-init", "step-02-requirements",
                 "step-03-dialogue-loop", "step-04-story-synthesis"]
phase: "complete"
story_finalized: true
final_story: {story_object}
```
