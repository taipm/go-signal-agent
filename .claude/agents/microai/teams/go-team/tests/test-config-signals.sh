#!/bin/bash
# Test Config System and Kanban Signal Integration
# Location: .claude/agents/microai/teams/go-team/tests/test-config-signals.sh

# Don't exit on error, we handle failures manually
set +e

BASE_DIR=".claude/agents/microai/teams/go-team"
CONFIG_FILE="$BASE_DIR/config/config.yaml"
CONFIG_LOADER="$BASE_DIR/config/config-loader.md"
SIGNAL_EMITTER="$BASE_DIR/kanban/signal-emitter.md"
ORCHESTRATOR="$BASE_DIR/agents/orchestrator-agent.md"
BOARD="$BASE_DIR/kanban/go-team-board.yaml"
INTEGRATION="$BASE_DIR/kanban/integration.md"

echo "═══════════════════════════════════════════════════════════════"
echo "  GO TEAM CONFIG & SIGNAL INTEGRATION TEST SUITE"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "Testing Config System and Kanban Signal emission..."
echo ""

PASSED=0
FAILED=0

pass() {
    echo "  ✓ $1"
    ((PASSED++))
}

fail() {
    echo "  ✗ $1"
    ((FAILED++))
}

# ═══════════════════════════════════════════════════════════════
# SECTION 1: CONFIG FILE TESTS
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 1: Config File Structure"
echo "───────────────────────────────────────────────────────────────"

# Test 1.1: Config file exists
if [ -f "$CONFIG_FILE" ]; then
    pass "config.yaml exists"
else
    fail "config.yaml missing"
fi

# Test 1.2: Workflow config section
if grep -q "^workflow:" "$CONFIG_FILE"; then
    pass "workflow section exists"
else
    fail "workflow section missing"
fi

# Test 1.3: Iterations config
if grep -q "iterations:" "$CONFIG_FILE" && grep -q "default: 3" "$CONFIG_FILE"; then
    pass "iterations config with default 3"
else
    fail "iterations config missing or wrong default"
fi

# Test 1.4: Coverage config
if grep -q "coverage:" "$CONFIG_FILE" && grep -q "default: 80" "$CONFIG_FILE"; then
    pass "coverage config with default 80"
else
    fail "coverage config missing or wrong default"
fi

# Test 1.5: Security config
if grep -q "^security:" "$CONFIG_FILE"; then
    pass "security section exists"
else
    fail "security section missing"
fi

# Test 1.6: Security gate blocking
if grep -q "block_on_critical: true" "$CONFIG_FILE" && grep -q "block_on_high: true" "$CONFIG_FILE"; then
    pass "security gate blocking configured"
else
    fail "security gate blocking not configured"
fi

# Test 1.7: Kanban config
if grep -q "^kanban:" "$CONFIG_FILE"; then
    pass "kanban section exists"
else
    fail "kanban section missing"
fi

# Test 1.8: Signal emission config
if grep -q "emit_step_start: true" "$CONFIG_FILE" && grep -q "emit_step_complete: true" "$CONFIG_FILE"; then
    pass "signal emission configured"
else
    fail "signal emission not configured"
fi

# Test 1.9: WIP limits
if grep -q "wip_limits:" "$CONFIG_FILE"; then
    pass "WIP limits section exists"
else
    fail "WIP limits missing"
fi

# Test 1.10: Metrics config
if grep -q "^metrics:" "$CONFIG_FILE"; then
    pass "metrics section exists"
else
    fail "metrics section missing"
fi

# Test 1.11: Token tracking
if grep -q "track: true" "$CONFIG_FILE" && grep -q "budget:" "$CONFIG_FILE"; then
    pass "token tracking configured"
else
    fail "token tracking not configured"
fi

# Test 1.12: Agent models
if grep -q "orchestrator: \"opus\"" "$CONFIG_FILE" && grep -q "fixer: \"sonnet\"" "$CONFIG_FILE"; then
    pass "agent models configured"
else
    fail "agent models not configured"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 2: CONFIG LOADER TESTS
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 2: Config Loader"
echo "───────────────────────────────────────────────────────────────"

# Test 2.1: Config loader exists
if [ -f "$CONFIG_LOADER" ]; then
    pass "config-loader.md exists"
else
    fail "config-loader.md missing"
fi

# Test 2.2: get_config function
if grep -q "def get_config" "$CONFIG_LOADER"; then
    pass "get_config function defined"
else
    fail "get_config function missing"
fi

# Test 2.3: set_config function
if grep -q "def set_config" "$CONFIG_LOADER"; then
    pass "set_config function defined"
else
    fail "set_config function missing"
fi

# Test 2.4: load_config function
if grep -q "def load_config" "$CONFIG_LOADER"; then
    pass "load_config function defined"
else
    fail "load_config function missing"
fi

# Test 2.5: Validation rules
if grep -q "validators:" "$CONFIG_LOADER"; then
    pass "validation rules defined"
else
    fail "validation rules missing"
fi

# Test 2.6: Command handlers
if grep -q "\\*config" "$CONFIG_LOADER" && grep -q "\\*iterations" "$CONFIG_LOADER"; then
    pass "command handlers documented"
else
    fail "command handlers missing"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 3: SIGNAL EMITTER TESTS
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 3: Signal Emitter"
echo "───────────────────────────────────────────────────────────────"

# Test 3.1: Signal emitter exists
if [ -f "$SIGNAL_EMITTER" ]; then
    pass "signal-emitter.md exists"
else
    fail "signal-emitter.md missing"
fi

# Test 3.2: emit_signal function
if grep -q "def emit_signal" "$SIGNAL_EMITTER"; then
    pass "emit_signal function defined"
else
    fail "emit_signal function missing"
fi

# Test 3.3: Signal types documented
SIGNAL_TYPES=("session_started" "session_completed" "step_started" "step_completed" "agent_activated" "security_gate" "review_iteration" "breakpoint_hit")
for sig in "${SIGNAL_TYPES[@]}"; do
    if grep -q "$sig" "$SIGNAL_EMITTER"; then
        pass "signal type '$sig' documented"
    else
        fail "signal type '$sig' missing"
    fi
done

# Test 3.4: Signal handlers
if grep -q "handle_session_started" "$SIGNAL_EMITTER" && grep -q "handle_step_completed" "$SIGNAL_EMITTER"; then
    pass "signal handlers defined"
else
    fail "signal handlers missing"
fi

# Test 3.5: Emission points
if grep -q "Emission Points in Workflow" "$SIGNAL_EMITTER"; then
    pass "emission points documented"
else
    fail "emission points missing"
fi

# Test 3.6: Queue management
if grep -q "signal_queue" "$SIGNAL_EMITTER"; then
    pass "signal queue management documented"
else
    fail "signal queue management missing"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 4: ORCHESTRATOR INTEGRATION TESTS
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 4: Orchestrator Integration"
echo "───────────────────────────────────────────────────────────────"

# Test 4.1: Config system integration
if grep -q "Configuration System Integration" "$ORCHESTRATOR"; then
    pass "config system integrated in orchestrator"
else
    fail "config system not integrated in orchestrator"
fi

# Test 4.2: Config loading protocol
if grep -q "Config Loading Protocol" "$ORCHESTRATOR"; then
    pass "config loading protocol documented"
else
    fail "config loading protocol missing"
fi

# Test 4.3: Config access functions
if grep -q "get_config" "$ORCHESTRATOR" && grep -q "set_config" "$ORCHESTRATOR"; then
    pass "config access functions in orchestrator"
else
    fail "config access functions missing in orchestrator"
fi

# Test 4.4: Signal emission system
if grep -q "Signal Emission System" "$ORCHESTRATOR"; then
    pass "signal emission system in orchestrator"
else
    fail "signal emission system missing in orchestrator"
fi

# Test 4.5: Emission protocol
if grep -q "Emission Protocol" "$ORCHESTRATOR"; then
    pass "emission protocol documented"
else
    fail "emission protocol missing"
fi

# Test 4.6: Signal emission points
if grep -q "Signal Emission Points" "$ORCHESTRATOR"; then
    pass "signal emission points in orchestrator"
else
    fail "signal emission points missing in orchestrator"
fi

# Test 4.7: Board update integration
if grep -q "Board Update Integration" "$ORCHESTRATOR"; then
    pass "board update integration documented"
else
    fail "board update integration missing"
fi

# Test 4.8: WIP limit checking
if grep -q "check_wip_limit" "$ORCHESTRATOR"; then
    pass "WIP limit checking in orchestrator"
else
    fail "WIP limit checking missing in orchestrator"
fi

# Test 4.9: Config commands
if grep -q "\\*config" "$ORCHESTRATOR" && grep -q "\\*iterations" "$ORCHESTRATOR" && grep -q "\\*coverage" "$ORCHESTRATOR"; then
    pass "config commands in orchestrator"
else
    fail "config commands missing in orchestrator"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 5: KANBAN BOARD INTEGRATION TESTS
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 5: Kanban Board Integration"
echo "───────────────────────────────────────────────────────────────"

# Test 5.1: Board file exists
if [ -f "$BOARD" ]; then
    pass "go-team-board.yaml exists"
else
    fail "go-team-board.yaml missing"
fi

# Test 5.2: Signal queue in board
if grep -q "signals:" "$BOARD" && grep -q "pending:" "$BOARD"; then
    pass "signal queue in board"
else
    fail "signal queue missing in board"
fi

# Test 5.3: WIP limits in board
if grep -q "wip_limit:" "$BOARD"; then
    pass "WIP limits in board"
else
    fail "WIP limits missing in board"
fi

# Test 5.4: Agent registry in board
if grep -q "agents:" "$BOARD" && grep -q "orchestrator-agent:" "$BOARD"; then
    pass "agent registry in board"
else
    fail "agent registry missing in board"
fi

# Test 5.5: Metrics in board
if grep -q "metrics:" "$BOARD"; then
    pass "metrics section in board"
else
    fail "metrics section missing in board"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 6: INTEGRATION DOCUMENT TESTS
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 6: Integration Document"
echo "───────────────────────────────────────────────────────────────"

# Test 6.1: Integration doc exists
if [ -f "$INTEGRATION" ]; then
    pass "integration.md exists"
else
    fail "integration.md missing"
fi

# Test 6.2: Signal types documented
if grep -q "Signal Types" "$INTEGRATION"; then
    pass "signal types documented in integration"
else
    fail "signal types missing in integration"
fi

# Test 6.3: Workflow hooks
if grep -q "Workflow Hooks" "$INTEGRATION"; then
    pass "workflow hooks documented"
else
    fail "workflow hooks missing"
fi

# Test 6.4: Board columns mapping
if grep -q "Board Columns Mapping" "$INTEGRATION"; then
    pass "board columns mapping documented"
else
    fail "board columns mapping missing"
fi

# Test 6.5: WIP limit enforcement
if grep -q "WIP Limit Enforcement" "$INTEGRATION"; then
    pass "WIP limit enforcement documented"
else
    fail "WIP limit enforcement missing"
fi

# Test 6.6: Implementation checklist
if grep -q "Implementation Checklist" "$INTEGRATION"; then
    pass "implementation checklist exists"
else
    fail "implementation checklist missing"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 7: YAML SYNTAX VALIDATION
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 7: YAML Syntax Validation"
echo "───────────────────────────────────────────────────────────────"

# Check if python3 is available for YAML validation
if command -v python3 &> /dev/null; then
    # Test 7.1: config.yaml syntax
    if python3 -c "import yaml; yaml.safe_load(open('$CONFIG_FILE'))" 2>/dev/null; then
        pass "config.yaml valid YAML syntax"
    else
        fail "config.yaml has YAML syntax errors"
    fi

    # Test 7.2: go-team-board.yaml syntax
    if python3 -c "import yaml; yaml.safe_load(open('$BOARD'))" 2>/dev/null; then
        pass "go-team-board.yaml valid YAML syntax"
    else
        fail "go-team-board.yaml has YAML syntax errors"
    fi
else
    echo "  ⚠ python3 not available, skipping YAML syntax validation"
fi

echo ""

# ═══════════════════════════════════════════════════════════════
# SUMMARY
# ═══════════════════════════════════════════════════════════════

echo "═══════════════════════════════════════════════════════════════"
echo "  TEST SUMMARY"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  Passed: $PASSED"
echo "  Failed: $FAILED"
echo "  Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "  ✅ ALL TESTS PASSED!"
    echo ""
    echo "  Config System: OPERATIONAL"
    echo "  Signal Emitter: OPERATIONAL"
    echo "  Orchestrator Integration: COMPLETE"
    echo "  Kanban Integration: READY"
    EXIT_CODE=0
else
    echo "  ⚠ SOME TESTS FAILED"
    echo "  Please review the failures above."
    EXIT_CODE=1
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  FEATURE SUMMARY"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  Config System:"
echo "  • config.yaml with all parameters"
echo "  • config-loader.md with functions"
echo "  • Validation rules for ranges"
echo "  • Runtime commands (*config, *iterations, *coverage)"
echo ""
echo "  Signal System:"
echo "  • signal-emitter.md with emit protocol"
echo "  • 8 signal types defined"
echo "  • Emission points for all workflow events"
echo "  • Signal handlers for board updates"
echo ""
echo "  Integration:"
echo "  • Orchestrator has config loading"
echo "  • Orchestrator has signal emission"
echo "  • WIP limit checking integrated"
echo "  • Board update handlers ready"
echo ""
echo "═══════════════════════════════════════════════════════════════"

exit $EXIT_CODE
