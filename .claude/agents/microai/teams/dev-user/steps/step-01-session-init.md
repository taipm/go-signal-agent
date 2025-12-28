# Step 01: Session Initialization

## Objective
Khởi tạo session, load agents, và prepare cho dialogue.

## Actions

### 1. Load Agent Personas
```yaml
load:
  - path: "../agents/solo-dev.md"
    as: solo_dev_persona
  - path: "../agents/enduser.md"
    as: enduser_persona
```

### 2. Initialize Session State
```yaml
session:
  date: "{system_date}"
  subject: "{user_provided_topic}"
  turn_count: 0
  phase: "requirements"
  current_speaker: "enduser"
  dialogue_history: []
```

### 3. Display Welcome Message

```
╔══════════════════════════════════════════════════════════════╗
║               DEV-USER TEAM SESSION                          ║
╠══════════════════════════════════════════════════════════════╣
║  Topic: {subject}                                            ║
║  Date:  {date}                                               ║
╠══════════════════════════════════════════════════════════════╣
║  Participants:                                               ║
║  👨‍💻 Solo Developer - Technical Implementation Partner       ║
║  👤 End User - Business Requirements Partner                 ║
║  👁️ Observer (You) - Can intervene at any turn              ║
╚══════════════════════════════════════════════════════════════╝

OBSERVER CONTROLS:
┌────────────────┬────────────────────────────────────────────┐
│ [Enter]        │ Continue dialogue                          │
│ @dev: <msg>    │ Speak as Solo Developer                    │
│ @user: <msg>   │ Speak as End User                          │
│ @guide: <msg>  │ Redirect the conversation                  │
│ *skip          │ Jump to story synthesis                    │
│ *exit          │ End session                                │
└────────────────┴────────────────────────────────────────────┘

Session Goal: Collaboratively create a User Story with clear
Acceptance Criteria through structured dialogue.

═══════════════════════════════════════════════════════════════

Facilitator: Bắt đầu session! EndUser, hãy present requirements
của bạn về "{subject}".

───────────────────────────────────────────────────────────────
```

## Transition
→ Proceed to Step 02: Requirements Gathering

## State After Completion
```yaml
stepsCompleted: ["step-01-session-init"]
phase: "requirements"
current_speaker: "enduser"
```
