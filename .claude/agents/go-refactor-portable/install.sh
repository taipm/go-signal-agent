#!/bin/bash
# go-refactor-agent installer
# Installs the portable Go Refactoring Agent with 2-layer knowledge system
# Version: 1.0.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Version
VERSION=$(cat "$(dirname "${BASH_SOURCE[0]}")/VERSION" 2>/dev/null || echo "1.0.0")

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}   Go Refactor Agent - Portable Edition v${VERSION}${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# =============================================================================
# Pre-flight checks
# =============================================================================

# Check OS
case "$(uname -s)" in
    Linux*)     OS=Linux;;
    Darwin*)    OS=Mac;;
    CYGWIN*|MINGW*|MSYS*) OS=Windows;;
    *)          OS="Unknown"
esac

if [[ "$OS" == "Unknown" ]]; then
    echo -e "${RED}Error: Unsupported operating system${NC}"
    echo "Supported: Linux, macOS, Windows (via WSL/Git Bash)"
    exit 1
fi

# Check bash version
if [[ "${BASH_VERSION%%.*}" -lt 4 ]]; then
    echo -e "${YELLOW}Warning: Bash version ${BASH_VERSION} detected.${NC}"
    echo "Some features may not work. Recommended: Bash 4.0+"
fi

# Determine HOME directory (cross-platform)
if [[ "$OS" == "Windows" ]]; then
    HOME_DIR="${USERPROFILE:-$HOME}"
else
    HOME_DIR="$HOME"
fi

# Determine script directory (where agent files are)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GLOBAL_DIR="$HOME_DIR/.claude/agents/go-refactor"
PROJECT_DIR=""

# Parse arguments
INSTALL_MODE="both"
while [[ $# -gt 0 ]]; do
    case $1 in
        --global-only)
            INSTALL_MODE="global"
            shift
            ;;
        --project-only)
            INSTALL_MODE="project"
            PROJECT_DIR="${2:-.}"
            shift 2
            ;;
        --project)
            PROJECT_DIR="$2"
            shift 2
            ;;
        --help)
            echo "Usage: install.sh [options]"
            echo ""
            echo "Options:"
            echo "  --global-only     Only install global agent (no project setup)"
            echo "  --project-only    Only setup project knowledge (requires global)"
            echo "  --project <path>  Specify project directory (default: current)"
            echo "  --help            Show this help"
            echo ""
            echo "Examples:"
            echo "  ./install.sh                    # Install both global and current project"
            echo "  ./install.sh --global-only      # Only install global agent"
            echo "  ./install.sh --project ~/myapp  # Setup for specific project"
            exit 0
            ;;
        *)
            PROJECT_DIR="$1"
            shift
            ;;
    esac
done

# Default project to current directory if not specified
if [[ -z "$PROJECT_DIR" ]]; then
    PROJECT_DIR="$(pwd)"
fi

# =============================================================================
# STEP 1: Install GLOBAL agent
# =============================================================================

install_global() {
    echo -e "${YELLOW}[1/2] Installing GLOBAL agent...${NC}"

    # Create global directory
    mkdir -p "$GLOBAL_DIR/knowledge"

    # Copy agent definition
    cp "$SCRIPT_DIR/agent.md" "$GLOBAL_DIR/"
    echo -e "  ${GREEN}✓${NC} Copied agent.md"

    # Copy knowledge files
    cp "$SCRIPT_DIR/knowledge/go-idioms.md" "$GLOBAL_DIR/knowledge/"
    cp "$SCRIPT_DIR/knowledge/patterns.md" "$GLOBAL_DIR/knowledge/"
    cp "$SCRIPT_DIR/knowledge/anti-patterns.md" "$GLOBAL_DIR/knowledge/"
    echo -e "  ${GREEN}✓${NC} Copied knowledge files (go-idioms, patterns, anti-patterns)"

    echo -e "${GREEN}✓ GLOBAL agent installed at: $GLOBAL_DIR${NC}"
    echo ""
}

# =============================================================================
# STEP 2: Setup PROJECT knowledge
# =============================================================================

setup_project() {
    local project_path="$1"
    local project_knowledge="$project_path/.claude/go-refactor"

    echo -e "${YELLOW}[2/2] Setting up PROJECT knowledge...${NC}"
    echo -e "  Project: ${BLUE}$project_path${NC}"

    # Create project knowledge directory
    mkdir -p "$project_knowledge"

    # Create project-specific files (empty templates)
    if [[ ! -f "$project_knowledge/conventions.md" ]]; then
        cat > "$project_knowledge/conventions.md" << 'EOF'
# Project Conventions

> Project-specific coding standards discovered during refactoring sessions.
> These conventions override global patterns when there's a conflict.

## Naming Conventions

<!-- Add project-specific naming rules here -->

## Error Handling Style

<!-- Add project-specific error handling patterns here -->

## Package Structure

<!-- Add project-specific package organization here -->

## Testing Patterns

<!-- Add project-specific testing patterns here -->

---

*Auto-generated by go-refactor installer*
EOF
        echo -e "  ${GREEN}✓${NC} Created conventions.md"
    else
        echo -e "  ${YELLOW}⊘${NC} conventions.md already exists, skipping"
    fi

    if [[ ! -f "$project_knowledge/learnings.md" ]]; then
        cat > "$project_knowledge/learnings.md" << 'EOF'
# Session Learnings

> Insights and lessons learned from refactoring THIS project.
> Project-specific learnings that don't apply universally.

---

*Auto-generated by go-refactor installer*
EOF
        echo -e "  ${GREEN}✓${NC} Created learnings.md"
    else
        echo -e "  ${YELLOW}⊘${NC} learnings.md already exists, skipping"
    fi

    if [[ ! -f "$project_knowledge/metrics.md" ]]; then
        cat > "$project_knowledge/metrics.md" << 'EOF'
# Refactoring Metrics

> Track improvements in THIS project over time.

## Session History

| Date | Files | Lines Changed | Complexity Δ | Notes |
|------|-------|---------------|--------------|-------|

---

*Auto-generated by go-refactor installer*
EOF
        echo -e "  ${GREEN}✓${NC} Created metrics.md"
    else
        echo -e "  ${YELLOW}⊘${NC} metrics.md already exists, skipping"
    fi

    echo -e "${GREEN}✓ PROJECT knowledge setup at: $project_knowledge${NC}"
    echo ""
}

# =============================================================================
# STEP 3: Create/update command file in project
# =============================================================================

setup_command() {
    local project_path="$1"
    local commands_dir="$project_path/.claude/commands"

    echo -e "${YELLOW}[3/3] Setting up slash command...${NC}"

    mkdir -p "$commands_dir"

    cat > "$commands_dir/go-refactor.md" << 'EOF'
---
description: Go refactoring specialist - 5W2H issue-by-issue workflow with 2-layer learning
argument-hint: "[file/package path]"
---

You must fully embody this agent's persona. NEVER break character.

<agent-activation CRITICAL="TRUE">
1. LOAD the agent from ~/.claude/agents/go-refactor/agent.md (GLOBAL)
2. READ its entire contents for persona, methodology, and behavioral guidelines
3. LOAD GLOBAL knowledge from ~/.claude/agents/go-refactor/knowledge/
   - go-idioms.md - Go best practices (universal)
   - patterns.md - Refactoring patterns (cross-project)
   - anti-patterns.md - Code smells to avoid (universal)
4. LOAD PROJECT knowledge from .claude/go-refactor/
   - conventions.md - Project-specific coding standards
   - learnings.md - Project-specific session insights
   - metrics.md - Project improvement tracking
5. Execute refactoring based on arguments: $ARGUMENTS
6. After EACH issue:
   - Go-universal insight? → Update GLOBAL knowledge
   - Project-specific? → Update PROJECT knowledge
</agent-activation>

## 2-Layer Knowledge System

```text
GLOBAL (~/.claude/agents/go-refactor/)     ← Shared across ALL projects
  └── knowledge/
      ├── go-idioms.md      (Go best practices)
      ├── patterns.md       (Universal patterns)
      └── anti-patterns.md  (Common mistakes)

PROJECT (.claude/go-refactor/)              ← THIS project only
  ├── conventions.md  (Project coding style)
  ├── learnings.md    (Project insights)
  └── metrics.md      (Project metrics)
```

## 5-Phase Interactive Workflow

### Phase 1: Analysis
Đọc code, phát hiện TẤT CẢ issues

### Phase 2: 5W2H Todo List (BẮT BUỘC)
Tạo TodoWrite với MỖI issue theo 5W2H format

### Phase 3: Xử Lý Từng Issue Với User
1. Mark in_progress → 2. Show BEFORE → 3. Explain → 4. Fix
5. Show AFTER → 6. Validate → 7. **HỎI USER** → 8. Update knowledge → 9. Next

### Phase 4: Validation
go vet, go build, tests

### Phase 5: Learning Capture
- Go-universal? → GLOBAL
- Project-specific? → PROJECT

## Usage

```bash
/go-refactor ollama/              # Refactor package
/go-refactor pkg/signal/engine.go # Refactor file
```
EOF

    echo -e "  ${GREEN}✓${NC} Created $commands_dir/go-refactor.md"
    echo -e "${GREEN}✓ Slash command ready: /go-refactor${NC}"
    echo ""
}

# =============================================================================
# STEP 4: Register skill in settings (optional)
# =============================================================================

register_skill() {
    local project_path="$1"
    local settings_file="$project_path/.claude/settings.local.json"

    echo -e "${YELLOW}[4/4] Checking skill registration...${NC}"

    # Check if settings file exists
    if [[ ! -f "$settings_file" ]]; then
        echo -e "  ${YELLOW}⊘${NC} No settings.local.json found, creating..."
        mkdir -p "$(dirname "$settings_file")"
        cat > "$settings_file" << 'EOF'
{
  "permissions": {
    "allow": [
      "Skill(go-refactor)"
    ]
  }
}
EOF
        echo -e "  ${GREEN}✓${NC} Created settings.local.json with go-refactor skill"
    else
        # Check if skill already registered
        if grep -q "Skill(go-refactor)" "$settings_file" 2>/dev/null; then
            echo -e "  ${GREEN}✓${NC} Skill already registered in settings.local.json"
        else
            echo -e "  ${YELLOW}!${NC} Skill not registered. Add manually to $settings_file:"
            echo -e "      ${BLUE}\"Skill(go-refactor)\"${NC}"
        fi
    fi
    echo ""
}

# =============================================================================
# Main execution
# =============================================================================

case $INSTALL_MODE in
    "global")
        install_global
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}   Installation complete!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo "Next: Run './install.sh --project-only' in each project"
        ;;
    "project")
        if [[ ! -d "$GLOBAL_DIR" ]]; then
            echo -e "${RED}Error: Global agent not installed.${NC}"
            echo "Run './install.sh --global-only' first."
            exit 1
        fi
        setup_project "$PROJECT_DIR"
        setup_command "$PROJECT_DIR"
        register_skill "$PROJECT_DIR"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}   Project setup complete!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        ;;
    "both")
        install_global
        setup_project "$PROJECT_DIR"
        setup_command "$PROJECT_DIR"
        register_skill "$PROJECT_DIR"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}   Installation complete!${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "GLOBAL agent: ${BLUE}$GLOBAL_DIR${NC}"
        echo -e "PROJECT knowledge: ${BLUE}$PROJECT_DIR/.claude/go-refactor${NC}"
        echo ""
        echo -e "Usage: ${YELLOW}/go-refactor <file-or-package>${NC}"
        ;;
esac
