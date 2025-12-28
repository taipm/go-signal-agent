# Step 05: Session Close

## Objective
Save meeting minutes, display final story, và conclude session.

## Actions

### 1. Generate Meeting Minutes

Compile dialogue_history và final_story thành meeting minutes:

```yaml
meeting_minutes:
  metadata:
    session_date: "{date}"
    session_subject: "{subject}"
    participants:
      - "Solo Developer"
      - "End User"
      - "Observer: {observer_name}"
    duration: "{start_time} - {end_time}"
    total_turns: {turn_count}
    status: "completed" | "partial" | "abandoned"

  summary: "{ai_generated_summary}"

  key_decisions:
    - topic: ""
      decision: ""
      turn: 0

  dialogue_transcript: [...]

  open_questions: []
  deferred_items: []
  next_steps: []

  final_story: {story_object}
```

### 2. Save to File

**Filename format:** `{YYYY-MM-DD}-{subject-slug}.md`

**Subject slug rules:**
- Lowercase
- Replace spaces with hyphens
- Remove special characters
- Max 50 chars

**Example:** `2025-12-28-user-authentication.md`

**Save path:** `.claude/agents/microai/teams/dev-user/logs/`

### 3. Display Completion Message

```
╔══════════════════════════════════════════════════════════════╗
║                  SESSION COMPLETE ✅                          ║
╠══════════════════════════════════════════════════════════════╣
║  Topic:     {subject}                                        ║
║  Date:      {date}                                           ║
║  Turns:     {turn_count}                                     ║
║  Status:    {status}                                         ║
╠══════════════════════════════════════════════════════════════╣
║  📁 Meeting minutes saved to:                                ║
║     ./logs/{filename}                                        ║
╚══════════════════════════════════════════════════════════════╝

───────────────────────────────────────────────────────────────
FINAL USER STORY
───────────────────────────────────────────────────────────────

{formatted_final_story}

───────────────────────────────────────────────────────────────

Key Decisions Made:
{#each key_decisions}
• {topic}: {decision}
{/each}

───────────────────────────────────────────────────────────────

{#if deferred_items}
Deferred for Future:
{#each deferred_items}
• {item}
{/each}
{/if}

───────────────────────────────────────────────────────────────

Next Steps:
1. Review User Story với team
2. Add to product backlog
3. Estimate và plan sprint

───────────────────────────────────────────────────────────────

Thank you for participating in this Dev-User session!

👨‍💻 Solo Developer: "Ready to implement when you are!"
👤 End User: "Looking forward to seeing it in action!"

═══════════════════════════════════════════════════════════════
```

### 4. Partial/Abandoned Session Handling

**If session was *skipped:*
```
╔══════════════════════════════════════════════════════════════╗
║              SESSION CONCLUDED (PARTIAL) ⚠️                   ║
╠══════════════════════════════════════════════════════════════╣
║  Session ended early via *skip command.                      ║
║  Story generated with available information.                 ║
║  Marked as: DRAFT - NEEDS REVIEW                            ║
╚══════════════════════════════════════════════════════════════╝
```

**If session was *exited:*
```
╔══════════════════════════════════════════════════════════════╗
║              SESSION ABANDONED ❌                             ║
╠══════════════════════════════════════════════════════════════╣
║  Session ended via *exit command.                            ║
║  No User Story generated.                                    ║
║  Dialogue transcript saved for reference.                    ║
╚══════════════════════════════════════════════════════════════╝
```

**If max turns reached:*
```
╔══════════════════════════════════════════════════════════════╗
║              SESSION TIMEOUT ⏱️                               ║
╠══════════════════════════════════════════════════════════════╣
║  Maximum turns (20) reached.                                 ║
║  Best-effort story generated from dialogue.                  ║
║  Recommend: Schedule follow-up session.                      ║
╚══════════════════════════════════════════════════════════════╝
```

### 5. Clean Up Session State

```yaml
session_cleanup:
  - clear: dialogue_history (saved to file)
  - clear: session variables
  - retain: output file path for reference
```

## State After Completion
```yaml
stepsCompleted: ["step-01-session-init", "step-02-requirements",
                 "step-03-dialogue-loop", "step-04-story-synthesis",
                 "step-05-session-close"]
phase: "closed"
output_file: "./logs/{filename}"
workflow_complete: true
```

## Session Statistics (Optional)

Display if useful:
```
Session Statistics:
─────────────────────
Total turns: {n}
  - Solo Dev: {n}
  - EndUser: {n}
  - Observer interventions: {n}
Questions asked: {n}
Decisions made: {n}
Items deferred: {n}
Time phases:
  - Requirements: {n} turns
  - Clarification: {n} turns
  - Negotiation: {n} turns
  - Synthesis: {n} turns
```
