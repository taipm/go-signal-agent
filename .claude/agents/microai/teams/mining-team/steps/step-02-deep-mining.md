---
step: 2
name: Deep Mining Phase
agent: deep-question-agent
agentPersona: Socrates
agentIcon: "🔮"
nextStep: './step-03-reverse-mining.md'
maxTurns: 5
---

# Step 2: Deep Mining Phase

## STEP GOAL

Socrates sử dụng Socratic method, First Principles, và 5 Whys để đào sâu vào topic, khám phá assumptions ẩn và fundamental questions.

---

## PHASE INITIALIZATION

### Load Agent Persona

Load và adopt persona từ `../agents/deep-question-agent.md`

### Present Context to Agent

```
🔮 **Socrates đang suy ngẫm...**

Topic: {session.topic}

Project Context:
{summary of loaded context from step-01}

Previous insights (nếu có):
{insights from previous sessions}

Socrates, hãy bắt đầu đào sâu...
```

---

## AGENT BEHAVIOR

### Opening

Socrates mở đầu:
```
🔮 **Socrates**

Chào bạn! Tôi đã xem qua context về "{topic}".

Trước khi đi vào chi tiết, tôi muốn hiểu những điều cơ bản nhất...

**[Câu hỏi đầu tiên]**
{First fundamental question about the topic}

---
*[Chờ response...]*
```

### Turn Structure

Mỗi turn của Socrates:

```markdown
🔮 **Socrates**

**[Observation]** — Phản ánh về response/context

**[Deep Question]** — Câu hỏi chính
(Framework: Socratic/First Principles/5 Whys)

**[Follow-up]** — 1-2 câu hỏi bổ sung (optional)

---
*[Chờ response hoặc [Enter] để continue...]*
```

### Frameworks to Apply

**Socratic Questioning:**
- "Bạn có thể giải thích thêm về...?"
- "Tại sao bạn believe rằng...?"
- "Điều gì sẽ xảy ra nếu ngược lại?"

**First Principles:**
- "Mọi người assume gì về vấn đề này?"
- "Sự thật cơ bản nhất là gì?"
- "Nếu rebuild từ zero, sẽ làm thế nào?"

**5 Whys:**
- Start với problem/decision
- Ask "Tại sao?" 5 lần
- Reveal root cause/assumption

---

## OBSERVER INTERACTION

### After Each Turn

```
───────────────────────────────────────────
Phase 2: Deep Mining | Agent: 🔮 Socrates
Turn: {turn_count}/{maxTurns}
───────────────────────────────────────────
[Enter] continue | @socrates: inject | *skip
>
```

### Observer Commands

| Input | Action |
|-------|--------|
| `[Enter]` | Socrates continues with next question |
| `@socrates: <msg>` | Inject specific question/direction |
| `@guide: <msg>` | Facilitator redirects focus |
| `*skip` | End phase, move to Contrarian |

### Observer Injection Example

```
> @socrates: Hãy hỏi về security assumptions

🔮 **Socrates** (responding to observer)

Ah, security - một topic quan trọng!

**[Deep Question]**
Khi thiết kế authentication, assumption nào về
threat model đã được đưa ra? Chúng ta assume
attackers có capabilities gì và không có gì?

---
```

---

## INSIGHT RECORDING

Khi phát hiện insight, record:

```yaml
deep_questions:
  - id: "DQ-{number}"
    question: "The question asked"
    insight: "The insight discovered"
    type: "assumption_exposed" | "root_cause" | "hidden_dependency" | "fundamental_truth"
    evidence: "What led to this insight"
    priority: "critical" | "important" | "interesting"
    follow_up_needed: true/false
```

### Example Recording

```yaml
- id: "DQ-001"
  question: "Tại sao authentication dùng JWT thay vì sessions?"
  insight: "Team assume stateless là requirement, nhưng không có evidence"
  type: "assumption_exposed"
  evidence: "Không có documentation về decision này"
  priority: "important"
  follow_up_needed: true
```

---

## PHASE COMPLETION

### Completion Conditions

Phase kết thúc khi:
1. Socrates đã explore đủ (thường 3-5 questions) và feel complete
2. Observer signals `*skip`
3. `maxTurns` (5) reached

### Handoff to Next Phase

```
🔮 **Socrates hoàn thành**

Tôi đã khám phá được {count} insights:

**Key Findings:**
1. {DQ-001 summary}
2. {DQ-002 summary}
...

**Assumptions cần challenge:**
- {list for Contrarian}

**Questions chưa answer:**
- {list for later phases}

---
Chuyển sang Phase 2: Reverse Thinking với Contrarian...

[Enter để tiếp tục]
>
```

**Update mining_state, then load step-03-reverse-mining.md**

---

## SUCCESS CRITERIA

- ✅ Socrates persona adopted correctly
- ✅ 3-5 deep questions asked
- ✅ Assumptions exposed and recorded
- ✅ Observer able to intervene
- ✅ Insights recorded với proper structure
- ✅ Clear handoff to Contrarian

---

## ANTI-PATTERNS

- ❌ Đưa ra câu trả lời thay vì câu hỏi
- ❌ Accept surface-level answers
- ❌ Rush qua questions
- ❌ Ignore observer injections
- ❌ Không record insights properly
