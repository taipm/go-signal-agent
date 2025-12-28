---
step: 1
name: Session Initialization
nextStep: './step-02-deep-mining.md'
---

# Step 1: Session Initialization

## STEP GOAL

Khởi tạo mining session: chào đón observer, nhận topic, load project context, và chuẩn bị cho các phases tiếp theo.

---

## EXECUTION SEQUENCE

### 1. Welcome Observer

```
=== 🔮🔄🔍🚀 MINING TEAM SESSION ===

Chào mừng đến với Mining Team!

Tôi là Mining Facilitator, sẽ điều phối session hôm nay với team:

  🔮 Socrates     - Deep Questions (First Principles, Socratic Method)
  🔄 Contrarian   - Reverse Thinking (Inversion, Pre-mortem)
  🔍 Sherlock     - Codebase Explorer (Technical Analysis)
  🚀 Ops          - Production Readiness (Real-world Scenarios)

Mỗi agent sẽ "đào" từ góc độ khác nhau để tìm insights mà
có thể bị bỏ qua khi review một mình.
```

### 2. Get Topic from Observer

```
📋 **Bạn muốn mining về topic/project gì hôm nay?**

Ví dụ:
- "Authentication system của project này"
- "Performance của API endpoints"
- "Architecture decision cho feature X"
- "Toàn bộ codebase" (general review)

> Topic: _
```

**Wait for observer input.**

### 3. Auto-Load Project Context

Sau khi nhận topic, tự động load context:

```bash
# Check for README
if [ -f "README.md" ]; then
  Read README.md
fi

# Check for docs
Glob "docs/**/*.md"

# Check for dependencies
if [ -f "go.mod" ]; then
  Read go.mod
elif [ -f "package.json" ]; then
  Read package.json
fi

# Recent git history
git log --oneline -10
```

Display loaded context:
```
📂 **Project Context Loaded:**
- README.md ✅
- Docs: {count} files
- Dependencies: go.mod / package.json
- Recent commits: 10 entries

🔍 Relevant files for "{topic}":
- {list of potentially relevant files}
```

### 4. Check Previous Mining Sessions

```bash
# Check for previous sessions on similar topic
ls -la ./logs/*{topic_slug}*.md 2>/dev/null
```

If previous sessions exist:
```
📜 **Previous Mining Sessions Found:**
- {date}-{topic}.md - {summary}

Bạn muốn:
1. Xem insights từ session trước
2. Bắt đầu fresh session
3. Continue từ session trước

> Chọn (1/2/3): _
```

### 5. Initialize Mining State

```yaml
session:
  date: "{current_date}"
  topic: "{user_provided_topic}"
  topic_slug: "{kebab-case-topic}"
  observer_mode: true
  current_phase: "initialization"

mining_state:
  accumulated_insights:
    deep_questions: []
    challenges: []
    technical_findings: []
    production_concerns: []
  cross_references: []
  priority_issues: []
  context_loaded:
    readme: true/false
    docs_count: N
    dependencies: "go.mod" | "package.json" | null
    git_history: true/false
```

### 6. Explain Observer Controls

```
🎮 **Observer Controls:**

Trong suốt session, bạn có thể:

| Command | Effect |
|---------|--------|
| [Enter] | Tiếp tục với agent hiện tại |
| @socrates: <msg> | Hỏi/chỉ đạo Socrates |
| @contrarian: <msg> | Hỏi/chỉ đạo Contrarian |
| @sherlock: <msg> | Yêu cầu Sherlock investigate |
| @ops: <msg> | Hỏi Ops về production |
| @guide: <msg> | Redirect conversation |
| *skip | Bỏ qua phase hiện tại |
| *synthesize | Nhảy đến tổng hợp |
| *exit | Kết thúc session |

💡 Tip: Bạn là observer - có thể can thiệp bất cứ lúc nào!
```

### 7. Confirm and Proceed

```
✅ **Session Initialized**

Topic: {topic}
Date: {date}
Context: Loaded

Bắt đầu Phase 1: Deep Questions với Socrates...

───────────────────────────────────────────
[Enter để bắt đầu, hoặc gõ lệnh]
>
```

**Wait for observer confirmation, then load step-02-deep-mining.md**

---

## SUCCESS CRITERIA

- ✅ Observer welcomed và hiểu format
- ✅ Topic received và validated
- ✅ Project context auto-loaded
- ✅ Previous sessions checked
- ✅ Mining state initialized
- ✅ Observer controls explained
- ✅ Ready to proceed to Phase 1

---

## ERROR HANDLING

**No README found:**
```
⚠️ Không tìm thấy README.md.
Bạn có thể mô tả ngắn về project không?
```

**No relevant files for topic:**
```
⚠️ Không tìm thấy files rõ ràng liên quan đến "{topic}".
Sherlock sẽ explore codebase để tìm relevant areas.
```

**Observer wants to exit early:**
```
Đã lưu session state. Bạn có thể resume sau bằng cách
chạy /mine và chọn "Continue từ session trước".
```
