---
stepNumber: 2
nextStep: './step-03-architecture.md'
agent: pm-agent
hasBreakpoint: true
---

# Step 02: Requirements Gathering

## STEP GOAL

PM Agent thu thập requirements từ observer, tạo user stories với acceptance criteria, và định nghĩa scope rõ ràng.

## AGENT ACTIVATION

Load persona từ `../agents/pm-agent.md`

Activate với context:
- Topic từ step 01
- Project context đã load
- Observer preferences

## EXECUTION SEQUENCE

### 1. PM Agent Introduction

```
[PM Agent]

Chào! Tôi là PM Agent, chịu trách nhiệm thu thập và clarify requirements.

Dựa trên topic "{topic}", tôi có một số câu hỏi:
```

### 2. Clarifying Questions

PM Agent hỏi 3-5 câu hỏi chính:
- Ai là user chính của feature này?
- Business value mong muốn là gì?
- Có constraints về technology/timeline không?
- Integration với systems khác?
- Edge cases cần xử lý?

### 3. Create User Stories

Với mỗi feature/capability, tạo user story:

```markdown
**US-01: {title}**

**As a** {user persona},
**I want** {capability},
**So that** {business value}.

### Acceptance Criteria

**AC1:** Given {context}, When {action}, Then {expected}
**AC2:** Given {context}, When {action}, Then {expected}
```

### 4. Define Scope

Tóm tắt:
- **In Scope:** Những gì sẽ làm
- **Out of Scope:** Những gì không làm (deferred)
- **Assumptions:** Các giả định
- **Dependencies:** Phụ thuộc external

### 5. API Contract (nếu applicable)

```
### API Endpoints

POST /api/v1/{resource}
- Request: {schema}
- Response: {schema}
- Errors: 400, 401, 404, 500
```

### 6. Present Spec to Observer

```
[PM Agent]

Đây là spec draft cho "{topic}":

{formatted spec}

---
═══════════════ BREAKPOINT ═══════════════

Observer, vui lòng review spec này.

Options:
- [Enter] → Approve và tiếp tục đến Architecture
- @pm: <feedback> → Yêu cầu PM Agent điều chỉnh
- *pause → Tạm dừng để review offline
```

## BREAKPOINT HANDLING

Đây là breakpoint đầu tiên. Observer có quyền:
1. Approve spec → Proceed to Architecture
2. Request changes → PM Agent revises
3. Pause → Save state và wait

## OUTPUT

Spec document saved to session state:
```yaml
outputs:
  spec:
    user_stories: [...]
    scope:
      in_scope: [...]
      out_scope: [...]
    api_contract: {...}
    approved: true/false
```

## SUCCESS CRITERIA

- [ ] At least 1 user story created
- [ ] Acceptance criteria defined
- [ ] Scope clearly documented
- [ ] Observer approved spec
- [ ] Ready for Architecture phase

## NEXT STEP

After breakpoint approval, load `./step-03-architecture.md`
