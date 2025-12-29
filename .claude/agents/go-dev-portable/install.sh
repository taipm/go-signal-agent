#!/usr/bin/env bash
# install.sh - Portable Go Dev Agent Installer
#
# Installs go-dev-agent to any project or global Claude Code config.
#
# Usage:
#   ./install.sh                    # Install to current project
#   ./install.sh --global           # Install to ~/.claude/agents/
#   ./install.sh --project /path    # Install to specific project
#   ./install.sh --uninstall        # Remove installation

set -euo pipefail

# === Configuration ===
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly AGENT_NAME="go-dev"
readonly VERSION="1.0.0"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# === Logging ===
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

# === Installation Targets ===
get_global_dir() {
    echo "${HOME}/.claude/agents"
}

get_project_dir() {
    local project="${1:-.}"
    echo "${project}/.claude/agents"
}

# === Install Functions ===
install_agent() {
    local target_dir="$1"
    local agent_dir="${target_dir}/${AGENT_NAME}"

    log_info "Installing go-dev-agent to: ${agent_dir}"

    # Create directories
    mkdir -p "${agent_dir}/knowledge/learning/"{raw,pending,archive,tools}

    # Copy agent definition
    cp "${SCRIPT_DIR}/agent.md" "${agent_dir}/agent.md"

    # Copy knowledge base
    cp "${SCRIPT_DIR}/knowledge/"*.md "${agent_dir}/knowledge/" 2>/dev/null || true
    cp "${SCRIPT_DIR}/knowledge/"*.yaml "${agent_dir}/knowledge/" 2>/dev/null || true

    # Copy learning tools
    cp "${SCRIPT_DIR}/knowledge/learning/tools/"* "${agent_dir}/knowledge/learning/tools/" 2>/dev/null || true
    chmod +x "${agent_dir}/knowledge/learning/tools/"*.sh 2>/dev/null || true

    # Copy learning config
    cp "${SCRIPT_DIR}/knowledge/learning/config.md" "${agent_dir}/knowledge/learning/" 2>/dev/null || true

    # Initialize learning directories
    touch "${agent_dir}/knowledge/learning/raw/.gitkeep"
    touch "${agent_dir}/knowledge/learning/archive/.gitkeep"

    # Create initial review queue if not exists
    if [[ ! -f "${agent_dir}/knowledge/learning/pending/review-queue.md" ]]; then
        cat > "${agent_dir}/knowledge/learning/pending/review-queue.md" << 'EOF'
# Learning Review Queue

**Last Updated:** $(date +%Y-%m-%d)
**Pending Items:** 0
**Agent:** go-dev-agent

---

## Queue Empty

No pending learnings. Use `*learn-capture` to add new learnings.
EOF
    fi

    # Create version file
    echo "${VERSION}" > "${agent_dir}/VERSION"

    log_success "Agent installed successfully!"
}

install_skill() {
    local project_dir="$1"
    local commands_dir="${project_dir}/.claude/commands"

    mkdir -p "${commands_dir}"

    # Create skill command
    cat > "${commands_dir}/${AGENT_NAME}.md" << 'EOF'
---
description: Go development specialist - implement, debug, refactor, optimize Go code
---

$ARGUMENTS

Load agent from: @.claude/agents/go-dev/agent.md
EOF

    log_success "Skill /${AGENT_NAME} registered"
}

uninstall_agent() {
    local target_dir="$1"
    local agent_dir="${target_dir}/${AGENT_NAME}"

    if [[ -d "${agent_dir}" ]]; then
        rm -rf "${agent_dir}"
        log_success "Agent removed from: ${agent_dir}"
    else
        log_warn "Agent not found at: ${agent_dir}"
    fi

    # Remove skill
    local skill_file="${target_dir%/agents}/.claude/commands/${AGENT_NAME}.md"
    if [[ -f "${skill_file}" ]]; then
        rm "${skill_file}"
        log_success "Skill removed: ${skill_file}"
    fi
}

show_help() {
    cat << EOF
go-dev-agent Portable Installer v${VERSION}

Usage:
  $0 [OPTIONS]

Options:
  --global          Install to ~/.claude/agents/ (available in all projects)
  --project PATH    Install to specific project's .claude/agents/
  --uninstall       Remove installation
  --help            Show this help

Examples:
  $0                           # Install to current project
  $0 --global                  # Install globally
  $0 --project ~/my-go-app     # Install to specific project
  $0 --uninstall --global      # Remove global installation

After Installation:
  Use /go-dev in Claude Code to invoke the agent.
EOF
}

# === Main ===
main() {
    local mode="project"
    local target_path="."
    local uninstall=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --global)
                mode="global"
                shift
                ;;
            --project)
                mode="project"
                target_path="$2"
                shift 2
                ;;
            --uninstall)
                uninstall=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Determine target directory
    local target_dir
    if [[ "$mode" == "global" ]]; then
        target_dir=$(get_global_dir)
    else
        target_dir=$(get_project_dir "$target_path")
    fi

    # Execute action
    if $uninstall; then
        uninstall_agent "$target_dir"
    else
        install_agent "$target_dir"

        # Install skill for project mode
        if [[ "$mode" == "project" ]]; then
            install_skill "$target_path"
        fi

        echo ""
        log_info "Installation complete!"
        echo ""
        echo "Usage:"
        echo "  1. Open Claude Code in your project"
        echo "  2. Type: /go-dev <your task>"
        echo ""
        echo "Tools available in: ${target_dir}/${AGENT_NAME}/knowledge/learning/tools/"
        echo "  - parallel-validate.sh  - Parallel quality gates"
        echo "  - select-knowledge.sh   - Relevance-based loading"
    fi
}

main "$@"
