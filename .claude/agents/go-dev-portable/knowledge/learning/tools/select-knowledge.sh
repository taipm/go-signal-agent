#!/usr/bin/env bash
# select-knowledge.sh - Relevance-Based Knowledge Loading for go-dev-agent
#
# Selects relevant knowledge files based on task description keywords.
# Configuration is externalized to knowledge-keywords.conf
#
# Usage:
#   ./select-knowledge.sh "implement http server with graceful shutdown"
#   ./select-knowledge.sh --files "worker pool"
#   ./select-knowledge.sh --stats

set -uo pipefail

# === Configuration (externalized, not hardcoded) ===
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly KEYWORDS_FILE="${SCRIPT_DIR}/knowledge-keywords.conf"
readonly TOTAL_KNOWLEDGE_FILES=10

# Core files: always loaded (configurable)
readonly CORE_FILES=(
    "08-anti-patterns.md"
    "10-learned-anti-patterns.md"
)

# Source shared utilities (DRY principle)
source "${SCRIPT_DIR}/common.sh" || {
    echo "Error: Cannot source common.sh" >&2
    exit $EXIT_SYSTEM_ERROR
}

# === Single-Responsibility Functions ===

# Load keywords from config file
load_keywords() {
    if [[ ! -f "$KEYWORDS_FILE" ]]; then
        log_error "Keywords file not found: $KEYWORDS_FILE"
        exit $EXIT_SYSTEM_ERROR
    fi
    load_config "$KEYWORDS_FILE"
}

# Check if file is a core file
is_core_file() {
    local file="$1"
    for core in "${CORE_FILES[@]}"; do
        [[ "$file" == "$core" ]] && return 0
    done
    return 1
}

# Match keywords against task description
match_keywords() {
    local task="$1"
    local normalized
    normalized=$(normalize_text "$task")

    local matched_files=()
    local matched_info=()

    while IFS='|' read -r keyword file priority; do
        [[ -z "$keyword" ]] && continue

        if echo " $normalized " | grep -q " $keyword "; then
            # Check if file already matched
            local already_matched=false
            for f in "${matched_files[@]:-}"; do
                [[ "$f" == "$file" ]] && already_matched=true && break
            done

            if ! $already_matched; then
                matched_files+=("$file")
                matched_info+=("$keyword:$file")
            fi
        fi
    done <<< "$(load_keywords)"

    # Output matched files (newline separated for piping)
    printf '%s\n' "${matched_files[@]:-}"
}

# Get keywords that matched a specific file
get_matched_keywords() {
    local file="$1"
    local task="$2"
    local normalized
    normalized=$(normalize_text "$task")
    local keywords=""

    while IFS='|' read -r keyword f priority; do
        [[ "$f" != "$file" ]] && continue
        if echo " $normalized " | grep -q " $keyword "; then
            [[ -n "$keywords" ]] && keywords+=", "
            keywords+="$keyword"
        fi
    done <<< "$(load_keywords)"

    echo "$keywords"
}

# === Output Functions ===

display_results() {
    local task="$1"

    print_box_header "RELEVANCE-BASED KNOWLEDGE LOADING" 62
    echo ""
    echo -e "${COLOR_BLUE}TASK:${COLOR_RESET} $task"
    echo ""

    # Show core files
    echo -e "${COLOR_GREEN}CORE FILES (always loaded):${COLOR_RESET}"
    for f in "${CORE_FILES[@]}"; do
        echo "  ✓ $f"
    done
    echo ""

    # Get matched files
    local matched
    matched=$(match_keywords "$task")

    if [[ -z "$matched" ]]; then
        echo -e "${COLOR_YELLOW}RELEVANT FILES:${COLOR_RESET} None found (using core only)"
        echo ""
        echo "FILES TO LOAD: ${#CORE_FILES[@]} (core only)"
        echo "REDUCTION: $((100 - ${#CORE_FILES[@]} * 100 / TOTAL_KNOWLEDGE_FILES))% (${#CORE_FILES[@]}/$TOTAL_KNOWLEDGE_FILES files)"
    else
        echo -e "${COLOR_GREEN}RELEVANT FILES:${COLOR_RESET}"

        local extra_count=0
        while IFS= read -r file; do
            [[ -z "$file" ]] && continue
            is_core_file "$file" && continue

            local keywords
            keywords=$(get_matched_keywords "$file" "$task")
            echo "  ✓ $file (matched: $keywords)"
            ((extra_count++))
        done <<< "$matched"

        echo ""
        local total=$((${#CORE_FILES[@]} + extra_count))
        local reduction=$((100 - total * 100 / TOTAL_KNOWLEDGE_FILES))
        echo "FILES TO LOAD: $total"
        echo "REDUCTION: ${reduction}% ($total/$TOTAL_KNOWLEDGE_FILES files)"
    fi
}

get_file_list() {
    local task="$1"

    # Output core files
    printf '%s\n' "${CORE_FILES[@]}"

    # Output matched files (excluding duplicates handled by sort -u)
    match_keywords "$task"
}

show_stats() {
    print_box_header "KNOWLEDGE INDEX STATISTICS" 62
    echo ""

    local keyword_count
    keyword_count=$(load_keywords | wc -l | tr -d ' ')

    echo "Total knowledge files: $TOTAL_KNOWLEDGE_FILES"
    echo "Total keywords indexed: $keyword_count"
    echo "Core files (always loaded): ${#CORE_FILES[@]}"
    echo ""

    echo "Files by priority:"
    echo "  Priority 1 (high): 06-concurrency.md"
    echo "  Priority 2 (medium): 02, 03, 04, 05, 09"
    echo "  Priority 3 (low): 07, 11"
    echo "  Core (always): 08, 10"
    echo ""

    echo "Example reductions:"
    echo ""

    for example in "implement http server" "worker pool with channels" "openai chat completion"; do
        echo "Task: '$example'"
        get_file_list "$example" | sort -u | head -5
        echo ""
    done
}

show_help() {
    cat <<EOF
Usage: $0 [OPTIONS] "task description"

Options:
  --files     Output only file names (for scripting)
  --stats     Show index statistics
  --help      Show this help

Examples:
  $0 "implement http server with graceful shutdown"
  $0 --files "worker pool with channels"

Configuration:
  Keywords are defined in: $KEYWORDS_FILE
  Edit this file to add/modify keyword mappings.
EOF
}

# === Main Entry Point ===

main() {
    case "${1:-}" in
        --stats)
            show_stats
            ;;
        --files)
            shift
            [[ -z "${1:-}" ]] && {
                log_error "Task description required"
                exit $EXIT_USER_ERROR
            }
            get_file_list "$*" | sort -u | grep -v "^$"
            ;;
        --help|-h)
            show_help
            ;;
        "")
            log_error "Task description required"
            echo "Usage: $0 \"task description\""
            exit $EXIT_USER_ERROR
            ;;
        *)
            display_results "$*"
            ;;
    esac
}

main "$@"
