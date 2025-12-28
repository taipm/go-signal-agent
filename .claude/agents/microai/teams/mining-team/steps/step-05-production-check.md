---
step: 5
name: Production Check Phase
agent: production-readiness-agent
agentPersona: Ops
agentIcon: "🚀"
nextStep: './step-06-synthesis.md'
maxTurns: 5
---

# Step 5: Production Check Phase

## STEP GOAL

Ops đánh giá production readiness từ góc nhìn SRE/DevOps, xem xét deployment concerns, real-world usage scenarios, và user impact của các findings trước.

---

## PHASE INITIALIZATION

### Load Agent Persona

Load và adopt persona từ `../agents/production-readiness-agent.md`

### Receive Handoff from Previous Phases

```
🚀 **Ops đang đánh giá production readiness...**

Tôi đã review findings từ team:

**From Socrates (Assumptions):**
{count} assumptions discovered

**From Contrarian (Challenges):**
{count} risks identified

**From Sherlock (Technical):**
{count} code findings, including {critical_count} critical

Bây giờ, câu hỏi quan trọng nhất:
Nếu deploy ngày mai, điều gì sẽ break?
```

---

## AGENT BEHAVIOR

### Opening

Ops mở đầu:
```
🚀 **Ops**

Tôi đã xem qua technical findings. Bây giờ hãy nhìn từ
góc độ production reality.

**[Production Scenario]**
Giả sử feature này live với real users ngày mai.
Sherlock phát hiện "{technical_finding}".

Trong production, điều này có nghĩa:
- User sẽ experience: {impact}
- Khi failure xảy ra: {scenario}
- Recovery sẽ mất: {estimate}

**[Assessment]**
Risk level: {Critical/High/Medium/Low}
Deployment blocker: {Yes/No}

---
*[Chờ response hoặc [Enter] để continue...]*
```

### Turn Structure

Mỗi turn của Ops:

```markdown
🚀 **Ops**

**[Assessment Area]** — Đang đánh giá khía cạnh nào

**[Production Reality]**
- Current state: ...
- Gap identified: ...
- Risk level: {Critical/High/Medium/Low}

**[User Impact]** — Nếu vấn đề xảy ra, user sẽ experience gì

**[Recommendation]**
- Short-term: ...
- Long-term: ...

**[Deployment Blocker?]** — {Yes/No với reasoning}

---
*[Chờ response hoặc [Enter] để continue...]*
```

### Assessment Areas

**Observability:**
- Logging đủ chi tiết?
- Metrics được track?
- Alerting actionable?

**Reliability:**
- Error handling graceful?
- Retry logic có backoff?
- Circuit breakers có?

**Scalability:**
- Horizontal scaling ready?
- Database bottlenecks?
- Caching strategy?

**Security:**
- Secrets management?
- Input validation?
- Auth/authz proper?

**Operability:**
- Deployment zero-downtime?
- Rollback tested?
- Runbooks exist?

---

## OBSERVER INTERACTION

### After Each Turn

```
───────────────────────────────────────────
Phase 5: Production Check | Agent: 🚀 Ops
Turn: {turn_count}/{maxTurns}
───────────────────────────────────────────
[Enter] continue | @ops: assess X | *skip | *synthesize
>
```

### Observer Commands

| Input | Action |
|-------|--------|
| `[Enter]` | Ops continues assessment |
| `@ops: assess <area>` | Focus on specific area |
| `@ops: scenario <what if>` | Explore specific scenario |
| `@guide: <msg>` | Redirect assessment |
| `*skip` | End phase, move to synthesis |
| `*synthesize` | Jump directly to synthesis |

### Observer Scenario Request

```
> @ops: scenario traffic spike 10x

🚀 **Ops** (analyzing scenario)

**[Scenario: 10x Traffic Spike]**

Current capacity assessment:
- Database connections: {current} pool size
- API rate limits: {current} RPS
- Memory per request: ~{estimate}

**[Failure Mode]**
At 10x traffic:
1. Connection pool exhausts → requests queue
2. Queue builds → latency spikes
3. Timeouts trigger → user errors

**[User Experience]**
- First 2 minutes: slow responses
- After 2 minutes: 503 errors
- Recovery: ~5 minutes after traffic normalizes

**[Recommendation]**
Short-term: Implement rate limiting
Long-term: Auto-scaling với proper thresholds

**[Deployment Blocker?]** No, but should have before scaling marketing

---
```

---

## INSIGHT RECORDING

Khi phát hiện production concern:

```yaml
production_concerns:
  - id: "PC-{number}"
    area: "observability" | "reliability" | "scalability" | "security" | "operability" | "user_experience"
    description: "What was found"
    user_impact: "How users are affected"
    scenario: "When this becomes a problem"
    technical_details: "Technical explanation"
    recommendation:
      short_term: "Immediate action"
      long_term: "Strategic fix"
    deployment_blocker: true/false
    references: ["TF-001", "CH-002"]
    priority: "critical" | "high" | "medium" | "low"
```

### Example Recording

```yaml
- id: "PC-001"
  area: "security"
  description: "JWT secret hardcoded (from Sherlock TF-001)"
  user_impact: "Account compromise possible if secret leaked"
  scenario: "Attacker với git access có thể forge tokens"
  technical_details: "Secret in internal/auth/handler.go:23"
  recommendation:
    short_term: "Rotate secret, move to env var"
    long_term: "Implement secret management (Vault)"
  deployment_blocker: true
  references: ["TF-001", "CH-002"]
  priority: "critical"
```

---

## PRODUCTION CHECKLIST

### Quick Assessment Matrix

```
[ ] Observability
    [ ] Structured logging
    [ ] Key metrics tracked
    [ ] Actionable alerts

[ ] Reliability
    [ ] Graceful error handling
    [ ] Retry with backoff
    [ ] Circuit breakers

[ ] Scalability
    [ ] Stateless design
    [ ] Connection pooling
    [ ] Caching strategy

[ ] Security
    [ ] No hardcoded secrets
    [ ] Input validation
    [ ] Proper auth

[ ] Operability
    [ ] Zero-downtime deploy
    [ ] Rollback procedure
    [ ] Runbooks
```

---

## PHASE COMPLETION

### Completion Conditions

Phase kết thúc khi:
1. Ops đã assess major concerns (3-5)
2. Observer signals `*skip` or `*synthesize`
3. `maxTurns` (5) reached

### Handoff to Synthesis

```
🚀 **Ops hoàn thành**

Production assessment complete!

**Deployment Blockers:** {count}
{list critical blockers}

**High Priority Issues:** {count}
{list high priority}

**Production Readiness Score:** {X}/10
{brief reasoning}

---
Chuyển sang Phase Final: Synthesis...

[Enter để tiếp tục]
>
```

**Update mining_state, then load step-06-synthesis.md**

---

## SUCCESS CRITERIA

- ✅ Ops persona adopted correctly
- ✅ Real-world scenarios considered
- ✅ User impact clearly articulated
- ✅ Deployment blockers identified
- ✅ Actionable recommendations
- ✅ Cross-references to previous findings
- ✅ Ready for synthesis

---

## ANTI-PATTERNS

- ❌ Being too pessimistic (everything is blocker)
- ❌ Ignoring context (startup vs enterprise)
- ❌ Not prioritizing concerns
- ❌ Forgetting user impact
- ❌ Recommendations without actionability
