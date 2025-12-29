# Go Refactor Agent - Portable Edition

> Self-learning Go Refactoring Specialist with 2-layer knowledge system.
> Portable across projects and machines.

## Features

- **5W2H Workflow**: Phân tích từng issue với What/Why/Where/When/Who/How/HowMuch
- **Interactive Mode**: Xử lý từng issue, hỏi user xác nhận trước khi tiếp tục
- **2-Layer Knowledge**: GLOBAL patterns (cross-project) + PROJECT conventions (local)
- **Self-Learning**: Tích lũy patterns và anti-patterns qua mỗi session

---

## 📋 Step-by-Step Installation Guide

### Scenario A: Cài đặt lần đầu trên máy mới

```bash
# Step 1: Copy folder go-refactor-portable vào máy mới
# (qua USB, scp, git clone, hoặc download)
scp -r go-refactor-portable user@new-machine:~/tools/

# Step 2: SSH vào máy mới (hoặc mở terminal)
ssh user@new-machine

# Step 3: Di chuyển vào folder
cd ~/tools/go-refactor-portable

# Step 4: Cấp quyền thực thi cho script
chmod +x install.sh

# Step 5: Chạy installer
./install.sh

# ✅ Kết quả:
# - GLOBAL agent được cài vào ~/.claude/agents/go-refactor/
# - PROJECT knowledge được tạo trong thư mục hiện tại
# - Slash command /go-refactor sẵn sàng sử dụng
```

### Scenario B: Thêm agent vào một Go project mới

```bash
# Step 1: Đảm bảo đã cài GLOBAL (Scenario A)
ls ~/.claude/agents/go-refactor/
# Phải thấy: agent.md, knowledge/

# Step 2: Di chuyển vào Go project của bạn
cd ~/my-awesome-go-project

# Step 3: Chạy installer với flag --project-only
~/tools/go-refactor-portable/install.sh --project-only

# ✅ Kết quả:
# - .claude/go-refactor/ được tạo với conventions.md, learnings.md, metrics.md
# - .claude/commands/go-refactor.md được tạo
# - settings.local.json được kiểm tra/tạo
```

### Scenario C: Cài cho project ở đường dẫn khác

```bash
# Không cần cd vào project, chỉ định path trực tiếp
~/tools/go-refactor-portable/install.sh --project /path/to/another-project

# Hoặc cài GLOBAL + nhiều projects cùng lúc
~/tools/go-refactor-portable/install.sh --global-only
~/tools/go-refactor-portable/install.sh --project-only --project ~/project-1
~/tools/go-refactor-portable/install.sh --project-only --project ~/project-2
~/tools/go-refactor-portable/install.sh --project-only --project ~/project-3
```

---

## 🚀 Step-by-Step Usage Guide

### Step 1: Mở Claude Code trong project

```bash
cd ~/my-go-project
claude
```

### Step 2: Gọi lệnh refactor

```bash
# Refactor một package
/go-refactor pkg/handlers/

# Refactor một file cụ thể
/go-refactor internal/service/user.go

# Refactor với mô tả task
/go-refactor "simplify error handling in auth package"
```

### Step 3: Agent sẽ thực hiện 5-Phase Workflow

```
Phase 1: ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Agent đọc code và phát hiện TẤT CẢ issues

Phase 2: 5W2H TODO LIST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Agent tạo todo list với mỗi issue có đầy đủ:
• WHAT:     Vấn đề là gì?
• WHY:      Tại sao cần fix?
• WHERE:    Vị trí code (file:line)
• WHEN:     Khi nào xảy ra?
• WHO:      Ai bị ảnh hưởng?
• HOW:      Cách fix?
• HOW MUCH: Ước tính impact?

→ Agent HỎI BẠN xác nhận thứ tự ưu tiên

Phase 3: EXECUTION (từng issue một)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Với mỗi issue:
1. Agent show code BEFORE
2. Agent giải thích sẽ làm gì
3. Agent apply fix
4. Agent show code AFTER
5. Agent validate (go build, go vet)
6. → Agent HỎI BẠN: "Đồng ý không?"
7. Nếu đồng ý → ghi learning → next issue

Phase 4: VALIDATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
go build, go vet, go test

Phase 5: LEARNING CAPTURE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Agent tự động cập nhật:
• Go-universal insight → GLOBAL knowledge
• Project-specific → PROJECT knowledge
```

### Step 4: Review learnings

```bash
# Xem GLOBAL patterns đã học
cat ~/.claude/agents/go-refactor/knowledge/patterns.md

# Xem PROJECT learnings
cat .claude/go-refactor/learnings.md

# Xem metrics
cat .claude/go-refactor/metrics.md
```

---

## 📁 Quick Reference

### Install Commands

| Command | Mô tả |
|---------|-------|
| `./install.sh` | Cài GLOBAL + setup project hiện tại |
| `./install.sh --global-only` | Chỉ cài GLOBAL (1 lần/máy) |
| `./install.sh --project-only` | Chỉ setup project (cần GLOBAL trước) |
| `./install.sh --project <path>` | Setup project ở path khác |
| `./install.sh --help` | Xem help |

### Usage Commands

| Command | Mô tả |
|---------|-------|
| `/go-refactor <package>` | Refactor một package |
| `/go-refactor <file.go>` | Refactor một file |
| `/go-refactor "<task>"` | Refactor theo mô tả task |

## Directory Structure

After installation:

```
~/.claude/agents/go-refactor/          # GLOBAL (shared across all projects)
├── agent.md                           # Agent definition
└── knowledge/
    ├── go-idioms.md                   # Go best practices
    ├── patterns.md                    # Refactoring patterns discovered
    └── anti-patterns.md               # Code smells to avoid

$PROJECT/.claude/go-refactor/          # PROJECT (per-project)
├── conventions.md                     # Project-specific coding standards
├── learnings.md                       # Session insights for this project
└── metrics.md                         # Refactoring metrics

$PROJECT/.claude/commands/go-refactor.md  # Slash command
```

## Knowledge Layers

### GLOBAL Layer (`~/.claude/agents/go-refactor/knowledge/`)

- Shared across ALL projects on this machine
- Contains universal Go patterns and idioms
- Updated when you discover Go-universal insights:
  - "Go methods cannot have type parameters"
  - "Use strings.Builder for O(n) concatenation"
  - "Always check context.Done() in loops"

### PROJECT Layer (`$PROJECT/.claude/go-refactor/`)

- Specific to each project
- Contains project conventions and metrics
- Updated when you discover project-specific patterns:
  - "This project uses zap for logging"
  - "Error format: pkg: action: detail"
  - "All handlers return JSON with status field"

## 5W2H Workflow

The agent follows a strict 5W2H framework for each issue:

```
Issue #N: [Short name]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• WHAT:     What is the problem?
• WHY:      Why does it need fixing?
• WHERE:    File:line location
• WHEN:     When does it trigger?
• WHO:      Who/what is affected?
• HOW:      How to fix it?
• HOW MUCH: Impact estimate (lines, complexity)
```

## Requirements

- macOS or Linux (Windows: use WSL or Git Bash)
- Bash 4.0+
- Claude Code CLI

## Uninstall

```bash
# Remove global agent
rm -rf ~/.claude/agents/go-refactor

# Remove from a project
rm -rf $PROJECT/.claude/go-refactor
rm $PROJECT/.claude/commands/go-refactor.md
```

## Updating

To update the agent:

```bash
# Re-run installer (will overwrite global, preserve project knowledge)
./install.sh --global-only

# Knowledge files in ~/.claude/agents/go-refactor/knowledge/ will be replaced
# Project knowledge (.claude/go-refactor/) is NOT touched
```

## Troubleshooting

### Command not recognized

Make sure the skill is registered in Claude Code settings. Add to `.claude/settings.local.json`:

```json
{
  "permissions": {
    "allow": [
      "Skill(go-refactor)"
    ]
  }
}
```

### Permission denied on install.sh

```bash
chmod +x install.sh
```

### Windows support

Use WSL (Windows Subsystem for Linux) or Git Bash. Native Windows is not supported.

## Version

- **Version**: 1.0.0
- **Last Updated**: 2025-12-29
- **Author**: go-signal-agent team

## License

MIT
