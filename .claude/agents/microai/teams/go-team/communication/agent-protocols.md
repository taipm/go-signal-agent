# Agent Communication Protocols

**Version:** 1.0.0

Định nghĩa protocols giao tiếp cho từng agent trong Go Team.

---

## Agent Communication Matrix

```
              PM    ARCH   CODER  TEST   SEC    REV    OPT    DEV
PM Agent      -     →      →      -      -      -      -      -
Architect     ←     -      →      →      -      -      -      -
Go Coder      ←     ←      -      ↔      ←      ←      ←      -
Test Agent    -     ←      ↔      -      -      ←      -      -
Security      -     -      →      →      -      →      -      -
Reviewer      -     →      →      →      ←      -      -      -
Optimizer     -     ←      →      -      -      ←      -      -
DevOps        -     -      ←      -      -      ←      ←      -

Legend: → sends to, ← receives from, ↔ bidirectional
```

---

## PM Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Architect | `notification` | requirements | Spec completed |
| Coder | `notification` | requirements | Scope clarification |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Orchestrator | `broadcast` | No |
| Any | `query` about requirements | Yes |

### Communication Template

```yaml
pm_agent_protocol:
  sends:
    - type: notification
      to: architect-agent
      topic: requirements
      on_event: spec_approved
      payload:
        event: requirements_complete
        spec_location: "./docs/spec.md"
        key_features: [...]

  receives:
    - type: query
      topic: requirements
      respond_with:
        format: text
        include: [scope, acceptance_criteria, constraints]
```

---

## Architect Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Coder | `notification`, `query` | architecture | Design completed, clarification |
| Test | `notification` | architecture | Interface definitions |
| Reviewer | `query` | architecture | Design review request |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| PM | `notification` | No |
| Coder | `query` | Yes |
| Optimizer | `query` | Yes |

### Communication Template

```yaml
architect_agent_protocol:
  sends:
    - type: notification
      to: [go-coder-agent, test-agent]
      topic: architecture
      on_event: architecture_approved
      payload:
        event: design_complete
        interfaces: [...]
        package_structure: {...}

  receives:
    - type: query
      from: go-coder-agent
      topic: architecture
      expected_questions:
        - interface_signatures
        - package_placement
        - dependency_direction
      respond_with:
        format: code
        include_reference: true

  queries:
    - type: query
      to: reviewer-agent
      topic: architecture
      question: "Review proposed interface design"
      await_response: true
```

---

## Go Coder Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Architect | `query` | architecture | Need design clarification |
| Test | `notification`, `collaboration` | code_change | File created/modified |
| All | `notification` | workflow | Build status |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Architect | `notification`, `response` | No, N/A |
| Test | `query`, `collaboration` | Yes |
| Security | `collaboration` | Yes |
| Reviewer | `collaboration` | Yes |
| Optimizer | `collaboration` | Yes |

### Communication Template

```yaml
go_coder_agent_protocol:
  sends:
    # Query to Architect
    - type: query
      to: architect-agent
      topic: architecture
      shortcut: "?arch"
      payload_template:
        question: "{question}"
        context:
          file: "{current_file}"
          line: "{current_line}"

    # Notify Test of new code
    - type: notification
      to: test-agent
      topic: code_change
      on_event: file_created
      payload:
        event: file_created
        file: "{file_path}"
        functions: [...]

    # Broadcast build status
    - type: notification
      to: "*"
      topic: workflow
      on_event: build_complete
      payload:
        event: build_completed
        status: "{pass|fail}"
        errors: [...]

  receives:
    # Fix requests from Reviewer
    - type: collaboration
      from: reviewer-agent
      request: fix_code
      action: |
        1. Read issue details from payload
        2. Locate target file/function
        3. Apply fix
        4. Verify build passes
        5. Send response with code change

    # Vulnerability fixes from Security
    - type: collaboration
      from: security-agent
      request: fix_vulnerability
      priority: critical
      action: |
        1. Read vulnerability details
        2. Apply recommended fix
        3. Verify fix resolves issue
        4. Notify security-agent

    # Test questions
    - type: query
      from: test-agent
      topic: testing
      respond_with:
        error_types: [...]
        expected_behavior: "..."
```

### Quick Commands

| Command | Action |
|---------|--------|
| `?arch {q}` | Query architect about {q} |
| `?test {q}` | Query test agent about {q} |
| `!done` | Notify code complete |
| `!build` | Broadcast build status |

---

## Test Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Coder | `query`, `collaboration` | testing | Need impl details, request fix |
| Reviewer | `notification` | testing | Coverage report |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Architect | `notification` | No |
| Coder | `notification`, `response` | No |
| Reviewer | `collaboration` | Yes |

### Communication Template

```yaml
test_agent_protocol:
  sends:
    # Query Coder for implementation details
    - type: query
      to: go-coder-agent
      topic: testing
      shortcut: "?coder"
      payload_template:
        question: "What error types does {function} return?"

    # Request test helper from Coder
    - type: collaboration
      to: go-coder-agent
      topic: testing
      request: add_test_helper
      payload_template:
        description: "Need mock implementation for {interface}"

    # Report coverage
    - type: notification
      to: reviewer-agent
      topic: testing
      on_event: tests_complete
      payload:
        event: test_completed
        coverage: "{percentage}"
        passed: "{count}"
        failed: "{count}"

  receives:
    # Add test requests from Reviewer
    - type: collaboration
      from: reviewer-agent
      request: add_test
      action: |
        1. Read test requirement
        2. Create test case
        3. Verify test passes
        4. Send confirmation
```

---

## Security Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Coder | `notification`, `collaboration` | security | Vulnerability found |
| Test | `notification` | security | Security test recommendations |
| Reviewer | `notification` | security | Audit complete |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Any | `query` about security | Yes |
| Orchestrator | `broadcast` | No |

### Communication Template

```yaml
security_agent_protocol:
  sends:
    # Alert on critical vulnerability
    - type: collaboration
      to: go-coder-agent
      topic: security
      priority: critical
      request: fix_vulnerability
      payload_template:
        event: vulnerability_found
        severity: "{CRITICAL|HIGH|MEDIUM|LOW}"
        details:
          type: "{vuln_type}"
          cwe: "{cwe_id}"
          file: "{file}"
          line: "{line}"
          recommendation: "{fix}"
      blocking: true

    # Recommend security tests
    - type: notification
      to: test-agent
      topic: security
      payload:
        event: security_test_needed
        test_cases:
          - type: input_validation
            target: "{function}"
          - type: auth_bypass
            target: "{handler}"

    # Report to Reviewer
    - type: notification
      to: reviewer-agent
      topic: security
      on_event: scan_complete
      payload:
        event: security_audit_complete
        findings:
          critical: 0
          high: 0
          medium: 2
          low: 5
        report_location: "./docs/security-report.md"

  receives:
    - type: query
      topic: security
      respond_with:
        vulnerability_status: [...]
        recommendations: [...]
        compliance: {...}
```

### Alert Severity Actions

| Severity | Action | Blocking |
|----------|--------|----------|
| CRITICAL | Immediate collaboration request | Yes |
| HIGH | Collaboration request | Yes |
| MEDIUM | Notification with recommendation | No |
| LOW | Batch notification | No |

---

## Reviewer Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Coder | `collaboration` | review | Issue found |
| Test | `collaboration` | review | Test needed |
| Architect | `query` | review | Design question |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Coder | `notification`, `response` | No |
| Test | `notification` | No |
| Security | `notification` | No |
| Optimizer | `query` | Yes |

### Communication Template

```yaml
reviewer_agent_protocol:
  sends:
    # Code fix request
    - type: collaboration
      to: go-coder-agent
      topic: review
      shortcut: "!fix"
      request: fix_code
      payload_template:
        description: "{issue_description}"
        target:
          file: "{file}"
          function: "{function}"
          line_start: "{start}"
          line_end: "{end}"
        acceptance_criteria: [...]
        severity: "{CRITICAL|WARNING}"
      blocking: true

    # Test addition request
    - type: collaboration
      to: test-agent
      topic: review
      shortcut: "!test"
      request: add_test
      payload_template:
        description: "Missing test for {scenario}"
        target:
          function: "{function}"
        acceptance_criteria: [...]

    # Architecture clarification
    - type: query
      to: architect-agent
      topic: review
      shortcut: "?arch"
      question: "Is this pattern consistent with design?"

  receives:
    - type: notification
      from: [go-coder-agent, test-agent, security-agent]
      action: |
        1. Log incoming notification
        2. Update review checklist
        3. Queue for next review iteration

  review_loop_protocol:
    on_issue_found:
      1. Send collaboration to appropriate agent
      2. Wait for response
      3. Verify fix
      4. Update issue status

    on_all_clear:
      1. Broadcast review_complete
      2. Update metrics
```

---

## Optimizer Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| Coder | `collaboration` | performance | Optimization needed |
| Architect | `query` | performance | Design impact question |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Reviewer | `notification` | No |
| DevOps | `query` | Yes |

### Communication Template

```yaml
optimizer_agent_protocol:
  sends:
    # Optimization request
    - type: collaboration
      to: go-coder-agent
      topic: performance
      request: optimize
      payload_template:
        description: "Optimize {target}"
        optimization_type: "{memory|cpu|concurrency}"
        current_metrics:
          allocations: "{n}"
          time: "{ms}"
        target_metrics:
          allocations: "{n}"
          time: "{ms}"
        suggested_changes: [...]

    # Check architecture constraints
    - type: query
      to: architect-agent
      topic: performance
      question: "Can I add caching layer to {component}?"

  receives:
    - type: query
      from: devops-agent
      topic: performance
      respond_with:
        benchmarks: [...]
        recommendations: [...]
```

---

## DevOps Agent Protocol

### Outbound Communications

| To | Message Types | Topics | Triggers |
|----|--------------|--------|----------|
| None directly | - | - | - |

### Inbound Communications

| From | Expected Messages | Response Required |
|------|------------------|-------------------|
| Coder | `notification` | No |
| Reviewer | `notification` | No |
| Optimizer | `query`, `notification` | Yes, No |

### Communication Template

```yaml
devops_agent_protocol:
  sends:
    # Only broadcast release status
    - type: notification
      to: "*"
      topic: release
      on_event: release_ready
      payload:
        event: release_artifacts_ready
        docker_image: "{tag}"
        ci_status: "{pass|fail}"

  receives:
    - type: notification
      from: [go-coder-agent, reviewer-agent, optimizer-agent]
      action: |
        1. Log status updates
        2. Update release checklist
        3. Trigger CI pipeline if all ready

    - type: query
      from: optimizer-agent
      topic: release
      respond_with:
        build_flags: [...]
        docker_config: {...}
```

---

## Protocol Enforcement

### Pre-Send Validation

```python
def validate_message(message, sender_protocol):
    # 1. Check if sender can send to recipient
    if message.to not in sender_protocol.allowed_recipients:
        raise ProtocolViolation("Unauthorized recipient")

    # 2. Check if message type is allowed
    if message.type not in sender_protocol.allowed_types:
        raise ProtocolViolation("Unauthorized message type")

    # 3. Check topic subscription
    if message.topic not in sender_protocol.topics:
        raise ProtocolViolation("Not subscribed to topic")

    return True
```

### Post-Receive Validation

```python
def validate_response(response, original_query):
    # 1. Check response matches query
    if response.in_reply_to != original_query.id:
        raise ProtocolViolation("Response mismatch")

    # 2. Check response format
    if original_query.expected_format:
        validate_format(response.payload, original_query.expected_format)

    return True
```

---

## Error Recovery

### Agent Unavailable

```yaml
on_agent_unavailable:
  1. Queue message with retry
  2. Notify orchestrator if critical
  3. After max_retries:
     - If blocking: halt and escalate
     - If non-blocking: log and continue
```

### Response Timeout

```yaml
on_response_timeout:
  1. Retry with extended timeout (2x)
  2. If still timeout:
     - Query orchestrator for alternative
     - Or proceed with best-effort answer
  3. Log for debugging
```

### Protocol Violation

```yaml
on_protocol_violation:
  1. Log violation details
  2. Notify orchestrator
  3. Block message
  4. Return error to sender
```
