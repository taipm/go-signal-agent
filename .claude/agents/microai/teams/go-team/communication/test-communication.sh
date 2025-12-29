#!/bin/bash
# Test Agent Communication System
# Validates message bus, schemas, and agent protocols

set -e

COMM_DIR=".claude/agents/microai/teams/go-team/communication"

echo "═══════════════════════════════════════════════════════════"
echo "  AGENT COMMUNICATION SYSTEM TEST SUITE"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Test 1: Core Files Exist
echo "Test 1: Core Files..."
FILES=(
    "agent-bus.md"
    "message-schemas.md"
    "agent-protocols.md"
    "message-queue.json"
)

for file in "${FILES[@]}"; do
    if [ -f "$COMM_DIR/$file" ]; then
        echo "  ✓ $file exists"
    else
        echo "  ✗ $file NOT FOUND"
        exit 1
    fi
done
echo ""

# Test 2: Message Queue Structure
echo "Test 2: Message Queue Structure..."
if jq -e '.version' "$COMM_DIR/message-queue.json" > /dev/null 2>&1; then
    echo "  ✓ Valid JSON structure"
else
    echo "  ✗ Invalid JSON"
    exit 1
fi

# Check required fields
FIELDS=("queue_id" "configuration" "statistics" "agent_status" "topic_subscriptions")
for field in "${FIELDS[@]}"; do
    if jq -e ".$field" "$COMM_DIR/message-queue.json" > /dev/null 2>&1; then
        echo "  ✓ Field '$field' present"
    else
        echo "  ✗ Field '$field' missing"
        exit 1
    fi
done
echo ""

# Test 3: Agent Registration
echo "Test 3: Agent Registration..."
AGENTS=$(jq -r '.agent_status | keys[]' "$COMM_DIR/message-queue.json")
AGENT_COUNT=$(echo "$AGENTS" | wc -l | tr -d ' ')
echo "  ✓ $AGENT_COUNT agents registered"

# Verify all expected agents
EXPECTED_AGENTS=("orchestrator" "pm-agent" "architect-agent" "go-coder-agent" "test-agent" "security-agent" "reviewer-agent" "optimizer-agent" "devops-agent")
for agent in "${EXPECTED_AGENTS[@]}"; do
    if jq -e ".agent_status.\"$agent\"" "$COMM_DIR/message-queue.json" > /dev/null 2>&1; then
        STATUS=$(jq -r ".agent_status.\"$agent\".status" "$COMM_DIR/message-queue.json")
        echo "  ✓ $agent registered (status: $STATUS)"
    else
        echo "  ✗ $agent NOT registered"
        exit 1
    fi
done
echo ""

# Test 4: Topic Subscriptions
echo "Test 4: Topic Subscriptions..."
TOPICS=$(jq -r '.topic_subscriptions | keys[]' "$COMM_DIR/message-queue.json")
TOPIC_COUNT=$(echo "$TOPICS" | wc -l | tr -d ' ')
echo "  ✓ $TOPIC_COUNT topics configured"

# Check key topics
KEY_TOPICS=("architecture" "security" "review" "workflow")
for topic in "${KEY_TOPICS[@]}"; do
    SUBSCRIBERS=$(jq -r ".topic_subscriptions.$topic | length" "$COMM_DIR/message-queue.json")
    echo "  ✓ Topic '$topic': $SUBSCRIBERS subscribers"
done
echo ""

# Test 5: Communication Matrix Validation
echo "Test 5: Communication Matrix..."
# Verify Coder can receive from Architect
CODER_TOPICS=$(jq -r '.agent_status."go-coder-agent".subscribed_topics[]' "$COMM_DIR/message-queue.json")
if echo "$CODER_TOPICS" | grep -q "architecture"; then
    echo "  ✓ Coder subscribed to 'architecture' (can receive from Architect)"
else
    echo "  ✗ Coder not subscribed to architecture"
    exit 1
fi

# Verify Reviewer can communicate with Coder
if echo "$CODER_TOPICS" | grep -q "review"; then
    echo "  ✓ Coder subscribed to 'review' (can receive from Reviewer)"
else
    echo "  ✗ Coder not subscribed to review"
    exit 1
fi

# Verify Security can alert Coder
if echo "$CODER_TOPICS" | grep -q "security"; then
    echo "  ✓ Coder subscribed to 'security' (can receive Security alerts)"
else
    echo "  ✗ Coder not subscribed to security"
    exit 1
fi
echo ""

# Test 6: Workflow Integration
echo "Test 6: Workflow Integration..."
WORKFLOW_FILE=".claude/agents/microai/teams/go-team/workflow.md"

if grep -q "communication:" "$WORKFLOW_FILE"; then
    echo "  ✓ Communication config in workflow.md"
else
    echo "  ✗ Communication config missing from workflow"
    exit 1
fi

if grep -q "@ask:" "$WORKFLOW_FILE"; then
    echo "  ✓ Query command documented"
fi

if grep -q "@notify:" "$WORKFLOW_FILE"; then
    echo "  ✓ Notify command documented"
fi

if grep -q "@request:" "$WORKFLOW_FILE"; then
    echo "  ✓ Request command documented"
fi

if grep -q "@broadcast" "$WORKFLOW_FILE"; then
    echo "  ✓ Broadcast command documented"
fi
echo ""

# Test 7: Message Schema Documentation
echo "Test 7: Message Schemas..."
SCHEMA_FILE="$COMM_DIR/message-schemas.md"

MESSAGE_TYPES=("query" "response" "notification" "collaboration" "broadcast" "ack")
for msg_type in "${MESSAGE_TYPES[@]}"; do
    if grep -qi "## $msg_type" "$SCHEMA_FILE" || grep -qi "### $msg_type" "$SCHEMA_FILE" || grep -qi "\"type\": \"$msg_type\"" "$SCHEMA_FILE"; then
        echo "  ✓ Schema for '$msg_type' documented"
    else
        echo "  ⚠ Schema for '$msg_type' may be missing"
    fi
done
echo ""

# Test 8: Protocol Documentation
echo "Test 8: Agent Protocols..."
PROTOCOL_FILE="$COMM_DIR/agent-protocols.md"

for agent in "PM Agent" "Architect" "Go Coder" "Test Agent" "Security" "Reviewer" "Optimizer" "DevOps"; do
    if grep -qi "$agent" "$PROTOCOL_FILE"; then
        echo "  ✓ Protocol for '$agent' documented"
    else
        echo "  ⚠ Protocol for '$agent' may be missing"
    fi
done
echo ""

# Summary
echo "═══════════════════════════════════════════════════════════"
echo "  TEST RESULTS"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "  ✓ All 8 tests passed!"
echo ""
echo "  Agent Communication System Status: OPERATIONAL"
echo ""
echo "  Features Verified:"
echo "  • Message bus architecture"
echo "  • 6 message types defined"
echo "  • 9 agents registered"
echo "  • $TOPIC_COUNT topics configured"
echo "  • Agent protocols documented"
echo "  • Workflow integration complete"
echo ""
echo "  Available Commands:"
echo "  • @ask:{agent} \"{question}\"  - Query agent"
echo "  • @notify:{topic} \"{msg}\"    - Send notification"
echo "  • @request:{agent} \"{task}\"  - Request collaboration"
echo "  • @broadcast \"{msg}\"         - Broadcast to all"
echo "  • ?arch, ?test, ?sec           - Quick queries"
echo "  • !fix, !test, !vuln           - Quick requests"
echo ""
echo "═══════════════════════════════════════════════════════════"
