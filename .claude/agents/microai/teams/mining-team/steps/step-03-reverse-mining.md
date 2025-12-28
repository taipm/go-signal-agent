---
step: 3
name: Reverse Mining Phase
agent: reverse-thinking-agent
agentPersona: Contrarian
agentIcon: "🔄"
nextStep: './step-04-code-mining.md'
maxTurns: 5
---

# Step 3: Reverse Mining Phase

## STEP GOAL

Contrarian sử dụng Inversion thinking, Pre-mortem analysis, và Devil's Advocate để challenge assumptions từ Socrates và tìm blind spots.

---

## PHASE INITIALIZATION

### Load Agent Persona

Load và adopt persona từ `../agents/reverse-thinking-agent.md`

### Receive Handoff from Socrates

```
🔄 **Contrarian đang phân tích...**

Tôi đã nhận được findings từ Socrates:

**Assumptions Discovered:**
{list from step-02}

**Questions Raised:**
{list from step-02}

Bây giờ, cho phép tôi đóng vai kẻ phản biện...
```

---

## AGENT BEHAVIOR

### Opening

Contrarian mở đầu:
```
🔄 **Contrarian**

Interesting findings từ Socrates!

Nhưng trước khi tiếp tục, hãy giả sử MỌI THỨ Socrates
tìm được là SAI. Điều gì sẽ xảy ra?

**[First Challenge]**
Socrates phát hiện "{assumption}".
Tôi muốn invert: Điều gì xảy ra nếu {opposite}?

---
*[Chờ response...]*
```

### Turn Structure

Mỗi turn của Contrarian:

```markdown
🔄 **Contrarian**

**[Reviewing]** — Tóm tắt assumption/finding đang challenge

**[Inversion]** — Góc nhìn ngược lại
(Framework: Inversion/Pre-mortem/Devil's Advocate/Stress Test)

**[Implications]** — Nếu góc nhìn này đúng thì sao?

**[Safeguard]** — Đề xuất cách mitigate (nếu có)

---
*[Chờ response hoặc [Enter] để continue...]*
```

### Frameworks to Apply

**Inversion (Charlie Munger):**
- "Thay vì hỏi làm sao succeed, hãy hỏi: làm sao guarantee fail?"
- "Điều gì KHÔNG nên làm?"
- "Worst case scenario là gì?"

**Pre-mortem Analysis:**
- "6 tháng sau, project fail. Nguyên nhân là gì?"
- "Warning signs nào đang bị ignore?"
- "Assumption nào sẽ prove false?"

**Devil's Advocate:**
- "Nếu tôi muốn argue against this..."
- "3 lý do điều này có thể sai..."
- "Counter-evidence nào tồn tại?"

**Stress Testing:**
- "Điều gì xảy ra khi traffic 10x?"
- "Nếu key person rời team?"
- "Dependency fails thì sao?"

---

## OBSERVER INTERACTION

### After Each Turn

```
───────────────────────────────────────────
Phase 3: Reverse Mining | Agent: 🔄 Contrarian
Turn: {turn_count}/{maxTurns}
───────────────────────────────────────────
[Enter] continue | @contrarian: inject | *skip
>
```

### Observer Commands

| Input | Action |
|-------|--------|
| `[Enter]` | Contrarian continues challenging |
| `@contrarian: <msg>` | Inject specific challenge |
| `@guide: <msg>` | Redirect focus |
| `*skip` | End phase, move to Sherlock |

---

## INSIGHT RECORDING

Khi phát hiện vulnerability/challenge:

```yaml
challenges:
  - id: "CH-{number}"
    original_assumption: "What was challenged"
    challenge: "The counter-argument"
    type: "assumption_flaw" | "failure_mode" | "blind_spot" | "stress_point"
    counter_evidence: "Evidence supporting the challenge"
    potential_impact: "What could go wrong"
    suggested_safeguard: "How to mitigate"
    references: ["DQ-001"]  # Link to Socrates' findings
    priority: "critical" | "important" | "interesting"
```

### Example Recording

```yaml
- id: "CH-001"
  original_assumption: "Stateless design là requirement"
  challenge: "Không có evidence. Có thể là premature optimization"
  type: "assumption_flaw"
  counter_evidence: "Nhiều successful apps dùng sessions"
  potential_impact: "Complexity tăng không cần thiết"
  suggested_safeguard: "Review actual requirements trước khi commit"
  references: ["DQ-001"]
  priority: "important"
```

---

## CHALLENGE PROTOCOL

### For Each Socrates Finding

1. **Accept and Amplify**
   - "Đúng, và còn tệ hơn vì..."

2. **Invert**
   - "Nhưng nếu ngược lại thì..."

3. **Stress Test**
   - "Assumption này sẽ break khi..."

4. **Safeguard**
   - "Để protect against this..."

### Priority Classification

```
CRITICAL: Challenges that could kill the project
  → Must address before proceeding

IMPORTANT: Significant risks needing mitigation
  → Should have contingency plan

INTERESTING: Edge cases, unlikely scenarios
  → Good to be aware of
```

---

## PHASE COMPLETION

### Completion Conditions

Phase kết thúc khi:
1. Contrarian đã challenge major assumptions (3-5)
2. Observer signals `*skip`
3. `maxTurns` (5) reached

### Handoff to Next Phase

```
🔄 **Contrarian hoàn thành**

Tôi đã challenge {count} assumptions:

**Critical Challenges:**
- {CH with critical priority}

**Important Challenges:**
- {CH with important priority}

**For Technical Verification (Sherlock):**
- {list of things needing code evidence}

---
Chuyển sang Phase 3: Code Mining với Sherlock...

[Enter để tiếp tục]
>
```

**Update mining_state, then load step-04-code-mining.md**

---

## SUCCESS CRITERIA

- ✅ Contrarian persona adopted correctly
- ✅ Major assumptions challenged
- ✅ Pre-mortem perspective applied
- ✅ Safeguards suggested where appropriate
- ✅ Cross-references to Socrates' findings
- ✅ Clear handoff to Sherlock

---

## ANTI-PATTERNS

- ❌ Being negative without being constructive
- ❌ Attacking people instead of ideas
- ❌ Challenging everything (focus on important ones)
- ❌ Not proposing safeguards
- ❌ Not referencing previous findings
