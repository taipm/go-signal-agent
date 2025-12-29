#!/usr/bin/env bash
# parallel-validate.sh - Parallel Quality Gates for go-dev-agent
#
# Optimizes validation by running independent checks in parallel.
# Configuration is externalized to validate.conf
#
# Usage: ./parallel-validate.sh [--tier 1|2|3] [--verbose]

set -uo pipefail

# === Configuration ===
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly CONFIG_FILE="${SCRIPT_DIR}/validate.conf"

# Source shared utilities (DRY principle)
source "${SCRIPT_DIR}/common.sh" || {
    echo "Error: Cannot source common.sh" >&2
    exit 2
}

# === Configurable Settings ===
TIER="${TIER:-2}"
VERBOSE="${VERBOSE:-false}"
TEMP_DIR=""

# === Cleanup ===
cleanup() {
    [[ -n "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# === Argument Parsing ===
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --tier)
                TIER="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE="true"
                shift
                ;;
            --help|-h)
                show_help
                exit $EXIT_SUCCESS
                ;;
            *)
                log_error "Unknown option: $1"
                exit $EXIT_USER_ERROR
                ;;
        esac
    done

    # Validate tier
    if [[ ! "$TIER" =~ ^[123]$ ]]; then
        log_error "Tier must be 1, 2, or 3"
        exit $EXIT_USER_ERROR
    fi
}

show_help() {
    cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --tier 1|2|3   Validation tier (default: 2)
                 1 = Quick (build, vet, test)
                 2 = Standard (+ fmt, race)
                 3 = Comprehensive (+ staticcheck, gosec)
  --verbose      Show detailed output
  --help         Show this help

Configuration:
  Checks are defined in: $CONFIG_FILE
EOF
}

# === Check Execution ===

# Run a single check in background
run_check() {
    local name="$1"
    local cmd="$2"

    (
        local start_time
        start_time=$(date +%s.%N 2>/dev/null || date +%s)

        if eval "$cmd" > "$TEMP_DIR/${name}.out" 2>&1; then
            echo "0" > "$TEMP_DIR/${name}.status"
        else
            echo "1" > "$TEMP_DIR/${name}.status"
        fi

        local end_time
        end_time=$(date +%s.%N 2>/dev/null || date +%s)

        # Calculate elapsed (handle both GNU and BSD date)
        if command -v bc >/dev/null 2>&1; then
            echo "scale=2; $end_time - $start_time" | bc
        else
            echo "$((${end_time%.*} - ${start_time%.*}))"
        fi > "$TEMP_DIR/${name}.time"
    ) &
    echo $!
}

# Report result of a check
report_check() {
    local name="$1"

    local status
    status=$(cat "$TEMP_DIR/${name}.status" 2>/dev/null || echo "1")

    local elapsed
    elapsed=$(cat "$TEMP_DIR/${name}.time" 2>/dev/null || echo "?")

    if [[ "$status" == "0" ]]; then
        log_success "$name (${elapsed}s)"
        return 0
    else
        log_error "$name"
        if [[ "$VERBOSE" == "true" ]]; then
            cat "$TEMP_DIR/${name}.out" 2>/dev/null | head -20
        else
            cat "$TEMP_DIR/${name}.out" 2>/dev/null | head -5
        fi
        return 1
    fi
}

# === Phase Execution ===

# Phase 1: Independent checks (parallel)
run_phase1() {
    local failed=0
    local pids=()
    local names=()

    log_info "Phase 1: Running independent checks in parallel..."
    echo ""

    # Build (always run)
    pids+=("$(run_check "build" "go build ./...")")
    names+=("build")

    # Vet (always run)
    pids+=("$(run_check "vet" "go vet ./...")")
    names+=("vet")

    # Fmt (tier 2+)
    if [[ "$TIER" -ge 2 ]]; then
        pids+=("$(run_check "fmt" "test -z \"\$(gofmt -l . 2>/dev/null | grep -v vendor)\"")")
        names+=("fmt")
    fi

    # Wait and report
    for i in "${!pids[@]}"; do
        wait "${pids[$i]}" 2>/dev/null || true
        report_check "${names[$i]}" || ((failed++))
    done

    return $failed
}

# Phase 2: Dependent checks (after build passes)
run_phase2() {
    local failed=0
    local pids=()
    local names=()

    log_info "Phase 2: Running tests (build passed)..."
    echo ""

    # Test (always run in tier 1+)
    pids+=("$(run_check "test" "go test ./...")")
    names+=("test")

    # Race (tier 2+)
    if [[ "$TIER" -ge 2 ]]; then
        pids+=("$(run_check "race" "go test -race ./...")")
        names+=("race")
    fi

    # Wait and report
    for i in "${!pids[@]}"; do
        wait "${pids[$i]}" 2>/dev/null || true
        report_check "${names[$i]}" || ((failed++))
    done

    return $failed
}

# Phase 3: Extended checks (tier 3 only)
run_phase3() {
    local failed=0

    log_info "Phase 3: Running extended checks..."
    echo ""

    # Staticcheck
    if command -v staticcheck >/dev/null 2>&1; then
        local pid
        pid=$(run_check "staticcheck" "staticcheck ./...")
        wait "$pid" 2>/dev/null || true
        report_check "staticcheck" || ((failed++))
    else
        log_warn "staticcheck not installed, skipping"
    fi

    # Gosec
    if command -v gosec >/dev/null 2>&1; then
        local pid
        pid=$(run_check "gosec" "gosec -quiet ./...")
        wait "$pid" 2>/dev/null || true
        report_check "gosec" || ((failed++))
    else
        log_warn "gosec not installed, skipping"
    fi

    return $failed
}

# === Main ===

main() {
    parse_args "$@"

    # Create temp directory
    TEMP_DIR=$(mktemp -d)

    local start_total
    start_total=$(date +%s)
    local total_failed=0

    echo ""
    print_box_header "PARALLEL QUALITY GATES - TIER $TIER" 62
    echo ""

    # Phase 1
    run_phase1
    local phase1_failed=$?
    ((total_failed += phase1_failed))

    # Check if build passed
    local build_status
    build_status=$(cat "$TEMP_DIR/build.status" 2>/dev/null || echo "1")

    if [[ "$build_status" != "0" ]]; then
        log_error "Build failed - skipping tests"
        echo ""
        print_box_header "VALIDATION FAILED" 62
        exit $EXIT_USER_ERROR
    fi

    # Phase 2
    echo ""
    run_phase2
    ((total_failed += $?))

    # Phase 3 (tier 3 only)
    if [[ "$TIER" -ge 3 ]]; then
        echo ""
        run_phase3
        ((total_failed += $?))
    fi

    # Summary
    local end_total
    end_total=$(date +%s)
    local total_elapsed=$((end_total - start_total))

    echo ""
    echo "══════════════════════════════════════════════════════════════"

    if [[ "$total_failed" -eq 0 ]]; then
        echo -e "${COLOR_GREEN}✓ ALL CHECKS PASSED${COLOR_RESET} (${total_elapsed}s total)"
        echo ""
        print_box_header "VALIDATION PASSED" 62
        exit $EXIT_SUCCESS
    else
        echo -e "${COLOR_RED}✗ $total_failed CHECK(S) FAILED${COLOR_RESET} (${total_elapsed}s total)"
        echo ""
        print_box_header "VALIDATION FAILED" 62
        exit $EXIT_USER_ERROR
    fi
}

main "$@"
