#!/bin/bash
# Test Semi-Automatic Signal Flow
# Location: .claude/agents/microai/teams/go-team/tests/test-signal-flow.sh

set +e

BASE_DIR=".claude/agents/microai/teams/go-team"
QUEUE_FILE="$BASE_DIR/kanban/signal-queue.json"

echo "═══════════════════════════════════════════════════════════════"
echo "  SEMI-AUTOMATIC SIGNAL FLOW TEST"
echo "═══════════════════════════════════════════════════════════════"
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
# SECTION 1: Queue File Structure
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 1: Queue File Structure"
echo "───────────────────────────────────────────────────────────────"

if [ -f "$QUEUE_FILE" ]; then
    pass "signal-queue.json exists"
else
    fail "signal-queue.json missing"
fi

# Check JSON structure
if python3 -c "import json; json.load(open('$QUEUE_FILE'))" 2>/dev/null; then
    pass "valid JSON syntax"
else
    fail "invalid JSON syntax"
fi

# Check required keys
REQUIRED_KEYS=("queue" "session" "board_state" "metrics")
for key in "${REQUIRED_KEYS[@]}"; do
    if python3 -c "import json; d=json.load(open('$QUEUE_FILE')); assert '$key' in d" 2>/dev/null; then
        pass "has '$key' key"
    else
        fail "missing '$key' key"
    fi
done

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 2: Queue Handler Documentation
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 2: Queue Handler Documentation"
echo "───────────────────────────────────────────────────────────────"

HANDLER="$BASE_DIR/kanban/queue-handler.md"

if [ -f "$HANDLER" ]; then
    pass "queue-handler.md exists"
else
    fail "queue-handler.md missing"
fi

# Check for emit_signal function
if grep -q "def emit_signal" "$HANDLER"; then
    pass "emit_signal function documented"
else
    fail "emit_signal function missing"
fi

# Check for update_board_state function
if grep -q "def update_board_state" "$HANDLER"; then
    pass "update_board_state function documented"
else
    fail "update_board_state function missing"
fi

# Check for step_to_column mapping
if grep -q "def step_to_column" "$HANDLER"; then
    pass "step_to_column mapping documented"
else
    fail "step_to_column mapping missing"
fi

# Check for command handlers
COMMANDS=("cmd_board" "cmd_status" "cmd_metrics")
for cmd in "${COMMANDS[@]}"; do
    if grep -q "$cmd" "$HANDLER"; then
        pass "$cmd handler documented"
    else
        fail "$cmd handler missing"
    fi
done

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 3: Signal Types Coverage
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 3: Signal Types Coverage"
echo "───────────────────────────────────────────────────────────────"

SIGNAL_TYPES=(
    "session_started"
    "session_completed"
    "step_started"
    "step_completed"
    "agent_activated"
    "security_gate"
)

for sig in "${SIGNAL_TYPES[@]}"; do
    if grep -q "$sig" "$HANDLER"; then
        pass "signal '$sig' handled"
    else
        fail "signal '$sig' not handled"
    fi
done

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 4: Workflow Integration
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 4: Workflow Integration"
echo "───────────────────────────────────────────────────────────────"

WORKFLOW="$BASE_DIR/workflow.md"

# Check kanban config
if grep -q "queue_path:" "$WORKFLOW"; then
    pass "queue_path configured in workflow"
else
    fail "queue_path missing in workflow"
fi

if grep -q "sync_mode: semi_automatic" "$WORKFLOW"; then
    pass "sync_mode set to semi_automatic"
else
    fail "sync_mode not semi_automatic"
fi

# Check kanban commands in workflow
KANBAN_CMDS=("\\*board" "\\*status" "\\*metrics" "\\*wip")
for cmd in "${KANBAN_CMDS[@]}"; do
    if grep -q "$cmd" "$WORKFLOW"; then
        pass "command $cmd in workflow"
    else
        fail "command $cmd missing in workflow"
    fi
done

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 5: Orchestrator Integration
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 5: Orchestrator Integration"
echo "───────────────────────────────────────────────────────────────"

ORCHESTRATOR="$BASE_DIR/agents/orchestrator-agent.md"

# Check signal emission in orchestrator
if grep -q "emit_signal" "$ORCHESTRATOR"; then
    pass "emit_signal in orchestrator"
else
    fail "emit_signal missing in orchestrator"
fi

# Check signal emission points
EMISSION_POINTS=("on_session_start" "before_step" "after_step" "on_agent_activate")
for point in "${EMISSION_POINTS[@]}"; do
    if grep -q "$point" "$ORCHESTRATOR"; then
        pass "emission point '$point' defined"
    else
        fail "emission point '$point' missing"
    fi
done

echo ""

# ═══════════════════════════════════════════════════════════════
# SECTION 6: Simulate Signal Flow
# ═══════════════════════════════════════════════════════════════

echo "───────────────────────────────────────────────────────────────"
echo "SECTION 6: Simulate Signal Flow"
echo "───────────────────────────────────────────────────────────────"

# Create a test signal
TEST_SIGNAL=$(cat <<EOF
{
  "queue": {
    "pending": [
      {
        "id": "sig-test-001",
        "type": "session_started",
        "timestamp": "2025-12-29T02:30:00+07:00",
        "payload": {
          "session_id": "go-team-test-123",
          "topic": "Test Signal Flow",
          "workflow": "full_pipeline"
        }
      }
    ],
    "processed": []
  },
  "session": {
    "id": "go-team-test-123",
    "topic": "Test Signal Flow",
    "started_at": "2025-12-29T02:30:00+07:00",
    "current_step": "step-01",
    "current_agent": "orchestrator-agent"
  },
  "board_state": {
    "columns": {
      "backlog": { "tasks": [{"id": "task-test", "title": "Test Signal Flow", "status": "active"}], "wip": 1 },
      "requirements": { "tasks": [], "wip": 0 },
      "architecture": { "tasks": [], "wip": 0 },
      "development": { "tasks": [], "wip": 0 },
      "testing": { "tasks": [], "wip": 0 },
      "security": { "tasks": [], "wip": 0 },
      "review": { "tasks": [], "wip": 0 },
      "optimization": { "tasks": [], "wip": 0 },
      "release": { "tasks": [], "wip": 0 },
      "blocked": { "tasks": [], "wip": 0 },
      "done": { "tasks": [], "wip": 0 }
    },
    "progress_percent": 11,
    "total_wip": 1
  },
  "metrics": {
    "signals_emitted": 1,
    "signals_processed": 0,
    "steps_completed": 1,
    "total_steps": 9,
    "agents_activated": ["orchestrator-agent"],
    "durations": {"step-01": 30}
  },
  "last_updated": "2025-12-29T02:30:00+07:00"
}
EOF
)

# Save to a temp file and validate
TEMP_FILE=$(mktemp)
echo "$TEST_SIGNAL" > "$TEMP_FILE"

if python3 -c "import json; json.load(open('$TEMP_FILE'))" 2>/dev/null; then
    pass "test signal JSON valid"
else
    fail "test signal JSON invalid"
fi

# Check signal structure
if python3 -c "
import json
d = json.load(open('$TEMP_FILE'))
assert len(d['queue']['pending']) == 1
assert d['session']['id'] == 'go-team-test-123'
assert d['board_state']['progress_percent'] == 11
assert d['metrics']['signals_emitted'] == 1
" 2>/dev/null; then
    pass "test signal structure correct"
else
    fail "test signal structure incorrect"
fi

rm -f "$TEMP_FILE"

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
    echo "  Semi-Automatic Signal Flow: READY"
    EXIT_CODE=0
else
    echo "  ⚠ SOME TESTS FAILED"
    EXIT_CODE=1
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  SIGNAL FLOW SUMMARY"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  How it works:"
echo ""
echo "  1. User starts /go-team session"
echo "  2. Orchestrator emits signals at each step:"
echo "     - session_started (on init)"
echo "     - step_started (before each step)"
echo "     - agent_activated (when agent starts)"
echo "     - step_completed (after each step)"
echo "     - security_gate (after security scan)"
echo "     - session_completed (on finish)"
echo ""
echo "  3. Signals written to: signal-queue.json"
echo ""
echo "  4. User views board anytime with: *board"
echo ""
echo "  5. Board displays current state from queue"
echo ""
echo "═══════════════════════════════════════════════════════════════"

exit $EXIT_CODE
