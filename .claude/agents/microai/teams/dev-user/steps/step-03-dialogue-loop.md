# Step 03: Dialogue Loop

## Objective
Orchestrate turn-based dialogue giữa Solo Dev và EndUser cho đến khi đạt được agreement về User Story.

## Main Loop

```
WHILE (turn_count < max_turns) AND (NOT story_finalized):
    execute_turn()
    check_observer_intervention()
    determine_next_speaker()
    check_completion_signals()
```

## Turn Execution

### Solo Dev Turn Template
```
[Turn {n} - Solo Dev] 👨‍💻

{response_to_previous}

{main_content}:
- Questions to clarify, OR
- Options to propose, OR
- Summary to confirm

{handoff}:
- "[Chờ câu trả lời của bạn...]"
- "[Bạn chọn option nào?]"
- "[Xác nhận để tiếp tục?]"

───────────────────────────────────────────────────────────────
Turn {n} | Phase: {phase} | Speaker: Solo Dev
[Enter] continue | @dev/@user/@guide: inject | *skip/*exit
>
```

### EndUser Turn Template
```
[Turn {n} - EndUser] 👤

{direct_answer_or_statement}

{context_or_examples}

{handoff}:
- "[Bạn nghĩ sao?]"
- "[Còn câu hỏi gì không?]"
- "[Tiếp tục đi!]"

───────────────────────────────────────────────────────────────
Turn {n} | Phase: {phase} | Speaker: EndUser
[Enter] continue | @dev/@user/@guide: inject | *skip/*exit
>
```

## Observer Intervention Handling

### Parse Observer Input
```yaml
input_patterns:
  - pattern: "^$|^\\s*$"              # Empty/Enter
    action: "continue"
  - pattern: "^@dev:\\s*(.+)$"
    action: "inject_as_dev"
    capture: "message"
  - pattern: "^@user:\\s*(.+)$"
    action: "inject_as_enduser"
    capture: "message"
  - pattern: "^@guide:\\s*(.+)$"
    action: "facilitator_note"
    capture: "message"
  - pattern: "^\\*skip$"
    action: "skip_to_synthesis"
  - pattern: "^\\*exit$"
    action: "end_session"
  - pattern: "^\\*pause$"
    action: "pause_dialogue"
  - pattern: "^\\*restart$"
    action: "restart_session"
```

### Intervention Response Templates

**@dev injection:**
```
[Turn {n} - Solo Dev] 👨‍💻 (Observer Override)

{injected_message}

───────────────────────────────────────────────────────────────
```

**@user injection:**
```
[Turn {n} - EndUser] 👤 (Observer Override)

{injected_message}

───────────────────────────────────────────────────────────────
```

**@guide note:**
```
[Facilitator Note] 📋

{guide_message}

Agents, please address this point in your next turn.

───────────────────────────────────────────────────────────────
```

## Speaker Determination Logic

```yaml
next_speaker_rules:
  - condition: "question_asked"
    next: "other_agent"

  - condition: "options_proposed"
    next: "other_agent"

  - condition: "summary_presented"
    next: "other_agent"

  - condition: "explicit_handoff"
    next: "named_agent"

  - condition: "facilitator_intervention"
    next: "addressed_agent"

  - condition: "story_proposed"
    next: "enduser"  # For review
```

## Phase Transition Logic

```yaml
phase_transitions:
  requirements_to_clarification:
    trigger: "Solo Dev asks first clarifying question"
    action: "phase = 'clarification'"

  clarification_to_negotiation:
    trigger: "All major questions answered, discussing scope/trade-offs"
    action: "phase = 'negotiation'"

  negotiation_to_synthesis:
    trigger: "Agreement reached, Solo Dev says 'Let me summarize...'"
    action: "phase = 'synthesis'"

  synthesis_to_complete:
    trigger: "EndUser says 'I agree' or 'Sign off'"
    action: "story_finalized = true"
```

## Completion Signals

### From Solo Dev
- "Tôi đề xuất User Story sau..."
- "Đây là AC tôi đề xuất..."
- "Tóm tắt để sign off..."

### From EndUser
- "Tôi đồng ý với story này"
- "Sign off"
- "Looks good, proceed"
- "Accepted"

## State Tracking Per Turn

```yaml
turn_record:
  turn: {number}
  speaker: "solo-dev" | "enduser" | "observer"
  speaker_icon: "👨‍💻" | "👤" | "👁️"
  message: "{content}"
  timestamp: "{ISO_timestamp}"
  phase: "{current_phase}"
  intervention: false | true
  key_points_extracted: []
  decisions_made: []
  questions_raised: []
  questions_answered: []
```

## Max Turns Warning

At turn 15:
```
[Facilitator Warning] ⚠️

Chúng ta đã ở turn 15/20. Nếu cần thêm thời gian,
hãy focus vào finalizing scope và acceptance criteria.

Gợi ý: Solo Dev có thể summarize current understanding
và propose User Story với những gì đã thống nhất.

───────────────────────────────────────────────────────────────
```

At turn 20 (max):
```
[Facilitator] ⏱️

Đã đạt giới hạn 20 turns. Session sẽ chuyển sang
Story Synthesis với progress hiện tại.

Nếu chưa có agreement đầy đủ, story sẽ được mark
là "Draft - Needs Review".

───────────────────────────────────────────────────────────────
```

## Transition
→ When story_finalized OR turn_count >= 20:
   Proceed to Step 04: Story Synthesis

## State After Completion
```yaml
stepsCompleted: ["step-01-session-init", "step-02-requirements", "step-03-dialogue-loop"]
phase: "synthesis"
story_finalized: true | false
turn_count: {final_count}
dialogue_history: [{...turns...}]
```
