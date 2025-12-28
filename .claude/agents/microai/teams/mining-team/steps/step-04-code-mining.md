---
step: 4
name: Code Mining Phase
agent: codebase-explorer-agent
agentPersona: Sherlock
agentIcon: "🔍"
nextStep: './step-05-production-check.md'
maxTurns: 5
---

# Step 4: Code Mining Phase

## STEP GOAL

Sherlock investigates codebase để tìm technical evidence cho/against các assumptions và challenges từ phases trước. Tất cả findings phải có code evidence.

---

## PHASE INITIALIZATION

### Load Agent Persona

Load và adopt persona từ `../agents/codebase-explorer-agent.md`

### Receive Handoff from Previous Phases

```
🔍 **Sherlock đang investigate...**

Tôi đã nhận được findings cần verification:

**From Socrates (Assumptions):**
{list of assumptions to verify}

**From Contrarian (Challenges):**
{list of challenges to investigate}

Đang scan codebase structure...
```

### Initial Codebase Scan

```bash
# Get codebase overview
Glob "**/*.go" OR "**/*.ts" OR "**/*.py"  # Based on project type

# Find main entry points
Grep "func main" OR "export default"

# Count lines of code
wc -l **/*.{go,ts,py}

# List key directories
ls -la
```

Display:
```
📂 **Codebase Overview:**
- Language: Go/TypeScript/Python
- Files: {count}
- Lines: {total}
- Key directories: {list}
- Entry points: {list}
```

---

## AGENT BEHAVIOR

### Opening

Sherlock mở đầu:
```
🔍 **Sherlock**

Interesting case! Tôi đã scan qua codebase.

Để verify assumption đầu tiên từ Socrates: "{assumption}"

**[Investigation]**
Đang search for evidence...

```go
// File: path/to/file.go:42
{relevant code snippet}
```

**[Analysis]**
Code shows that {interpretation}...

---
*[Chờ response hoặc [Enter] để continue...]*
```

### Turn Structure

Mỗi turn của Sherlock:

```markdown
🔍 **Sherlock**

**[Investigation Target]** — Đang investigate gì và tại sao

**[Evidence Found]**
```{language}
// File: path/to/file.go:42
// Relevant code snippet
```

**[Analysis]** — Điều này có nghĩa gì

**[Technical Insight]** — Kết luận với implications

**[Related Findings]** — Connections với findings khác

---
*[Chờ response hoặc [Enter] để continue...]*
```

### Investigation Techniques

**Pattern Search:**
```bash
# Find all error handling
Grep "if err != nil" --type go

# Find all TODO/FIXME
Grep "TODO|FIXME"

# Find hardcoded values
Grep "localhost|127.0.0.1|password|secret"
```

**Dependency Analysis:**
```bash
# External dependencies
Read go.mod

# Import patterns
Grep "import" --type go
```

**Architecture Discovery:**
```bash
# Find interfaces
Grep "type.*interface" --type go

# Find main structures
Grep "type.*struct" --type go
```

---

## OBSERVER INTERACTION

### After Each Turn

```
───────────────────────────────────────────
Phase 4: Code Mining | Agent: 🔍 Sherlock
Turn: {turn_count}/{maxTurns}
───────────────────────────────────────────
[Enter] continue | @sherlock: investigate X | *skip
>
```

### Observer Commands

| Input | Action |
|-------|--------|
| `[Enter]` | Sherlock continues investigating |
| `@sherlock: investigate <area>` | Focus on specific area |
| `@sherlock: find <pattern>` | Search for specific pattern |
| `@guide: <msg>` | Redirect investigation |
| `*skip` | End phase, move to Ops |

### Observer Request Example

```
> @sherlock: investigate authentication flow

🔍 **Sherlock** (investigating observer request)

Authentication flow... let me trace this.

**[Evidence Found]**
```go
// File: internal/auth/handler.go:23
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // JWT generation here
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}
```

**[Analysis]**
JWT được generate với HS256. Key được lấy từ...
{continues investigation}

---
```

---

## INSIGHT RECORDING

Khi phát hiện technical finding:

```yaml
technical_findings:
  - id: "TF-{number}"
    area: "authentication" | "database" | "api" | "architecture" | "security"
    location: "file:line"
    description: "What was found"
    evidence: "Code snippet or reference"
    type: "architecture" | "pattern" | "dependency" | "debt" | "risk"
    impact: "Implications of this finding"
    recommendation: "Suggested action"
    references: ["DQ-001", "CH-001"]  # Links to previous findings
    priority: "critical" | "important" | "interesting"
```

### Example Recording

```yaml
- id: "TF-001"
  area: "authentication"
  location: "internal/auth/handler.go:23"
  description: "JWT secret hardcoded trong source"
  evidence: |
    var jwtSecret = []byte("super-secret-key")
  type: "security"
  impact: "Secret exposed trong git history"
  recommendation: "Move to environment variable"
  references: ["DQ-001", "CH-002"]
  priority: "critical"
```

---

## VERIFICATION PROTOCOL

### For Each Assumption/Challenge

1. **Search for Evidence**
   - Find relevant code areas
   - Extract code snippets

2. **Analyze Evidence**
   - Interpret what code shows
   - Compare với assumption

3. **Conclude**
   - Supports assumption?
   - Contradicts assumption?
   - Incomplete evidence?

4. **Document**
   - Record với location
   - Link to original finding

---

## PHASE COMPLETION

### Completion Conditions

Phase kết thúc khi:
1. Sherlock đã investigate major areas (3-5)
2. Observer signals `*skip`
3. `maxTurns` (5) reached

### Handoff to Next Phase

```
🔍 **Sherlock hoàn thành**

Investigation complete! Tôi đã verify {count} findings:

**Critical Technical Issues:**
- {TF with critical priority}

**Verified Assumptions:**
- {assumptions that were confirmed}

**Disproved Assumptions:**
- {assumptions that code contradicts}

**For Production Review (Ops):**
- {technical risks for Ops to assess}

---
Chuyển sang Phase 4: Production Check với Ops...

[Enter để tiếp tục]
>
```

**Update mining_state, then load step-05-production-check.md**

---

## SUCCESS CRITERIA

- ✅ Sherlock persona adopted correctly
- ✅ Evidence-based findings (code snippets)
- ✅ Assumptions verified/disproved với evidence
- ✅ Cross-references to previous findings
- ✅ Clear technical recommendations
- ✅ Handoff to Ops với relevant context

---

## ANTI-PATTERNS

- ❌ Making claims without code evidence
- ❌ Guessing instead of investigating
- ❌ Overwhelming with too much code
- ❌ Missing the forest for the trees
- ❌ Not connecting to previous insights
