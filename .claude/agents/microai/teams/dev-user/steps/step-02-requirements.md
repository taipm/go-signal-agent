# Step 02: Requirements Gathering

## Objective
EndUser presents initial requirements về topic đã chọn.

## Actions

### 1. EndUser Presents (In-Character)

EndUser agent nên:
- Bắt đầu với greeting ngắn gọn
- Describe the problem hoặc opportunity
- Provide high-level requirements
- End với invitation for questions

**Template:**
```
[Turn 1 - EndUser] 👤

Chào! Cảm ơn đã gặp tôi hôm nay.

Về {subject}, đây là điều tôi cần:

{high_level_description}

Cụ thể, tôi muốn:
- {requirement_1}
- {requirement_2}
- {requirement_3}

{optional_constraints_or_context}

Bạn có câu hỏi gì cho tôi không?

───────────────────────────────────────────────────────────────
[Enter] continue | @dev/@user/@guide: inject | *skip/*exit
>
```

### 2. Record in Dialogue History
```yaml
dialogue_history:
  - turn: 1
    speaker: "enduser"
    speaker_icon: "👤"
    message: "{full_message}"
    timestamp: "{timestamp}"
    phase: "requirements"
```

### 3. Update State
```yaml
turn_count: 1
current_speaker: "solo-dev"
phase: "clarification"  # Dev will now ask questions
```

## Guidance for EndUser Content

### Good Initial Requirements
- Problem-focused, not solution-focused
- Business value clear
- Scope roughly defined
- Open to developer questions

### Example
```
Về User Authentication, tôi cần:

Chúng tôi đang build web app cho internal team (~50 users).
Hiện tại mọi người dùng shared credentials - rất không an toàn.

Tôi muốn:
- Mỗi user có account riêng
- Có thể login/logout
- Dashboard chỉ accessible sau khi đăng nhập

Timeline: Cần đi production trong 2 tuần.

Bạn có câu hỏi gì cho tôi?
```

## Transition
→ Proceed to Step 03: Dialogue Loop

## State After Completion
```yaml
stepsCompleted: ["step-01-session-init", "step-02-requirements"]
phase: "clarification"
current_speaker: "solo-dev"
turn_count: 1
```
