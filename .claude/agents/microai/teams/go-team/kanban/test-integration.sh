#!/bin/bash
# Test script for Go-Team Kanban Integration
# Location: .claude/agents/microai/teams/go-team/kanban/test-integration.sh

set -e

BOARD_PATH=".claude/agents/microai/teams/go-team/kanban/go-team-board.yaml"
TEST_SESSION_ID="test-session-$(date +%s)"
TEST_TOPIC="Kanban Integration Test"

echo "═══════════════════════════════════════════════════════════════════"
echo "       GO-TEAM KANBAN INTEGRATION TEST"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
echo "Test Session ID: $TEST_SESSION_ID"
echo "Board Path: $BOARD_PATH"
echo ""

# Test 1: Board file exists
echo "Test 1: Board file exists"
if [ -f "$BOARD_PATH" ]; then
    echo "  ✓ PASS: Board file exists"
else
    echo "  ✗ FAIL: Board file not found"
    exit 1
fi

# Test 2: Board structure valid
echo "Test 2: Board structure validation"
if command -v yq &> /dev/null; then
    BOARD_NAME=$(yq '.board.name' "$BOARD_PATH")
    COLUMNS_COUNT=$(yq '.columns | keys | length' "$BOARD_PATH")
    AGENTS_COUNT=$(yq '.agents | keys | length' "$BOARD_PATH")

    echo "  Board Name: $BOARD_NAME"
    echo "  Columns: $COLUMNS_COUNT"
    echo "  Agents: $AGENTS_COUNT"

    if [ "$COLUMNS_COUNT" -ge 10 ]; then
        echo "  ✓ PASS: Columns count >= 10"
    else
        echo "  ✗ FAIL: Expected >= 10 columns, got $COLUMNS_COUNT"
    fi

    if [ "$AGENTS_COUNT" -ge 9 ]; then
        echo "  ✓ PASS: Agents count >= 9"
    else
        echo "  ✗ FAIL: Expected >= 9 agents, got $AGENTS_COUNT"
    fi
else
    echo "  ⚠ SKIP: yq not installed, skipping YAML validation"
fi

# Test 3: Integration doc exists
echo "Test 3: Integration documentation"
INTEGRATION_DOC=".claude/agents/microai/teams/go-team/kanban/integration.md"
if [ -f "$INTEGRATION_DOC" ]; then
    echo "  ✓ PASS: Integration doc exists"
else
    echo "  ✗ FAIL: Integration doc not found"
fi

# Test 4: Workflow kanban config
echo "Test 4: Workflow kanban configuration"
WORKFLOW_PATH=".claude/agents/microai/teams/go-team/workflow.md"
if grep -q "kanban:" "$WORKFLOW_PATH"; then
    echo "  ✓ PASS: Kanban config found in workflow.md"
else
    echo "  ✗ FAIL: Kanban config not found in workflow.md"
fi

if grep -q "board_path:" "$WORKFLOW_PATH"; then
    echo "  ✓ PASS: Board path configured"
else
    echo "  ✗ FAIL: Board path not configured"
fi

# Test 5: Orchestrator kanban integration
echo "Test 5: Orchestrator kanban integration"
ORCHESTRATOR_PATH=".claude/agents/microai/teams/go-team/agents/orchestrator-agent.md"
if grep -q "Kanban Integration" "$ORCHESTRATOR_PATH"; then
    echo "  ✓ PASS: Kanban integration section found in orchestrator"
else
    echo "  ✗ FAIL: Kanban integration not found in orchestrator"
fi

if grep -q "WIP Limit Enforcement" "$ORCHESTRATOR_PATH"; then
    echo "  ✓ PASS: WIP limit enforcement documented"
else
    echo "  ✗ FAIL: WIP limit enforcement not found"
fi

# Test 6: Signal types defined
echo "Test 6: Signal types validation"
SIGNALS=("session_started" "step_started" "step_completed" "agent_activated" "security_gate" "session_completed")
for signal in "${SIGNALS[@]}"; do
    if grep -q "$signal" "$ORCHESTRATOR_PATH"; then
        echo "  ✓ Signal: $signal"
    else
        echo "  ✗ Missing signal: $signal"
    fi
done

# Test 7: Column-Step mapping
echo "Test 7: Column-Step mapping"
COLUMNS=("requirements" "architecture" "development" "testing" "security" "review" "optimization" "release")
for col in "${COLUMNS[@]}"; do
    if grep -q "$col:" "$BOARD_PATH"; then
        echo "  ✓ Column: $col"
    else
        echo "  ✗ Missing column: $col"
    fi
done

# Test 8: Kanban commands in workflow
echo "Test 8: Kanban commands in workflow"
COMMANDS=("*board" "*board:full" "*wip" "*metrics:kanban")
for cmd in "${COMMANDS[@]}"; do
    if grep -q "$cmd" "$WORKFLOW_PATH"; then
        echo "  ✓ Command: $cmd"
    else
        echo "  ✗ Missing command: $cmd"
    fi
done

echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo "       TEST SUMMARY"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
echo "All critical tests completed."
echo ""
echo "To run a full integration test:"
echo "  1. Start go-team session: /go-team"
echo "  2. Request a simple feature"
echo "  3. Monitor kanban updates with *board"
echo "  4. Check metrics with *metrics:kanban"
echo ""
echo "═══════════════════════════════════════════════════════════════════"
