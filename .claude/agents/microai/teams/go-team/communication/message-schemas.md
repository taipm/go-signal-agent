# Message Schemas

**Version:** 1.0.0

Chi tiết schemas cho tất cả message types trong Agent Communication System.

---

## Base Message Schema

Tất cả messages đều extend từ base schema này:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["id", "type", "from", "to", "timestamp"],
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^msg-[a-f0-9]{8}$",
      "description": "Unique message identifier"
    },
    "type": {
      "type": "string",
      "enum": ["query", "response", "notification", "collaboration", "broadcast", "ack"],
      "description": "Message type"
    },
    "from": {
      "type": "string",
      "enum": ["orchestrator", "pm-agent", "architect-agent", "go-coder-agent", "test-agent", "security-agent", "reviewer-agent", "optimizer-agent", "devops-agent"],
      "description": "Sender agent"
    },
    "to": {
      "type": ["string", "array"],
      "description": "Recipient agent(s) or '*' for broadcast"
    },
    "topic": {
      "type": "string",
      "enum": ["requirements", "architecture", "code_change", "testing", "security", "review", "performance", "release", "workflow"],
      "description": "Message topic for routing"
    },
    "priority": {
      "type": "string",
      "enum": ["critical", "high", "normal", "low"],
      "default": "normal",
      "description": "Message priority level"
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 timestamp"
    },
    "session_id": {
      "type": "string",
      "description": "Current workflow session ID"
    },
    "correlation_id": {
      "type": "string",
      "description": "ID linking related messages"
    }
  }
}
```

---

## Query Message

Request for information from another agent.

```json
{
  "type": "query",
  "properties": {
    "payload": {
      "type": "object",
      "required": ["question"],
      "properties": {
        "question": {
          "type": "string",
          "maxLength": 500,
          "description": "The question being asked"
        },
        "context": {
          "type": "object",
          "properties": {
            "file": {"type": "string"},
            "line": {"type": "integer"},
            "function": {"type": "string"},
            "related_files": {
              "type": "array",
              "items": {"type": "string"}
            }
          }
        },
        "expected_format": {
          "type": "string",
          "enum": ["text", "code", "json", "yaml"],
          "default": "text"
        }
      }
    },
    "timeout_ms": {
      "type": "integer",
      "minimum": 1000,
      "maximum": 30000,
      "default": 5000
    },
    "requires_response": {
      "type": "boolean",
      "default": true
    }
  }
}
```

### Example Queries

**Coder → Architect:**
```json
{
  "id": "msg-a1b2c3d4",
  "type": "query",
  "from": "go-coder-agent",
  "to": "architect-agent",
  "topic": "architecture",
  "payload": {
    "question": "What methods should UserRepository interface have?",
    "context": {
      "file": "internal/repo/interfaces.go",
      "line": 10
    },
    "expected_format": "code"
  },
  "priority": "normal",
  "timestamp": "2025-12-28T22:00:00Z",
  "timeout_ms": 5000,
  "requires_response": true
}
```

**Test → Coder:**
```json
{
  "id": "msg-e5f6g7h8",
  "type": "query",
  "from": "test-agent",
  "to": "go-coder-agent",
  "topic": "testing",
  "payload": {
    "question": "What are the expected error types for CreateUser?",
    "context": {
      "file": "internal/service/user_service.go",
      "function": "CreateUser"
    },
    "expected_format": "json"
  },
  "priority": "normal",
  "timestamp": "2025-12-28T22:05:00Z"
}
```

---

## Response Message

Reply to a query message.

```json
{
  "type": "response",
  "required": ["in_reply_to", "payload"],
  "properties": {
    "in_reply_to": {
      "type": "string",
      "description": "ID of the query being answered"
    },
    "payload": {
      "type": "object",
      "properties": {
        "answer": {
          "type": "string",
          "description": "The answer content"
        },
        "code": {
          "type": "string",
          "description": "Code snippet if requested"
        },
        "reference": {
          "type": "string",
          "description": "Reference to documentation/file"
        },
        "confidence": {
          "type": "number",
          "minimum": 0,
          "maximum": 1,
          "description": "Confidence level of answer"
        },
        "alternatives": {
          "type": "array",
          "items": {"type": "string"},
          "description": "Alternative answers if applicable"
        }
      }
    },
    "status": {
      "type": "string",
      "enum": ["success", "partial", "unable", "error"],
      "default": "success"
    }
  }
}
```

### Example Response

```json
{
  "id": "msg-a1b2c3d4-resp",
  "type": "response",
  "from": "architect-agent",
  "to": "go-coder-agent",
  "in_reply_to": "msg-a1b2c3d4",
  "payload": {
    "answer": "UserRepository should follow repository pattern with CRUD operations",
    "code": "type UserRepository interface {\n\tCreate(ctx context.Context, user *User) error\n\tGetByID(ctx context.Context, id string) (*User, error)\n\tGetByEmail(ctx context.Context, email string) (*User, error)\n\tUpdate(ctx context.Context, user *User) error\n\tDelete(ctx context.Context, id string) error\n}",
    "reference": "docs/architecture.md:45-60",
    "confidence": 0.95
  },
  "status": "success",
  "timestamp": "2025-12-28T22:00:02Z"
}
```

---

## Notification Message

One-way notification that doesn't require response.

```json
{
  "type": "notification",
  "properties": {
    "payload": {
      "type": "object",
      "required": ["event"],
      "properties": {
        "event": {
          "type": "string",
          "enum": [
            "file_created",
            "file_modified",
            "build_started",
            "build_completed",
            "build_failed",
            "test_started",
            "test_completed",
            "test_failed",
            "vulnerability_found",
            "review_started",
            "review_issue",
            "phase_change",
            "checkpoint_created",
            "error_occurred"
          ]
        },
        "severity": {
          "type": "string",
          "enum": ["info", "warning", "error", "critical"]
        },
        "details": {
          "type": "object",
          "additionalProperties": true
        },
        "affected_files": {
          "type": "array",
          "items": {"type": "string"}
        },
        "action_required": {
          "type": "boolean",
          "default": false
        }
      }
    },
    "requires_response": {
      "const": false
    }
  }
}
```

### Example Notifications

**Security Alert:**
```json
{
  "id": "msg-sec-001",
  "type": "notification",
  "from": "security-agent",
  "to": ["go-coder-agent", "reviewer-agent"],
  "topic": "security",
  "payload": {
    "event": "vulnerability_found",
    "severity": "critical",
    "details": {
      "type": "SQL Injection",
      "cwe": "CWE-89",
      "file": "internal/repo/user_repo.go",
      "line": 42,
      "code": "query := fmt.Sprintf(\"SELECT * FROM users WHERE id = '%s'\", id)",
      "recommendation": "Use parameterized query with $1 placeholder"
    },
    "action_required": true
  },
  "priority": "critical",
  "timestamp": "2025-12-28T22:10:00Z"
}
```

**Build Status:**
```json
{
  "id": "msg-build-001",
  "type": "notification",
  "from": "go-coder-agent",
  "to": "*",
  "topic": "workflow",
  "payload": {
    "event": "build_completed",
    "severity": "info",
    "details": {
      "duration_ms": 3500,
      "binary_size": "12MB",
      "go_version": "1.21.5"
    },
    "affected_files": ["cmd/app/main.go"]
  },
  "priority": "low",
  "timestamp": "2025-12-28T22:15:00Z"
}
```

---

## Collaboration Message

Request for another agent to perform a task.

```json
{
  "type": "collaboration",
  "properties": {
    "payload": {
      "type": "object",
      "required": ["request", "description"],
      "properties": {
        "request": {
          "type": "string",
          "enum": [
            "fix_code",
            "add_test",
            "update_test",
            "fix_vulnerability",
            "add_documentation",
            "refactor",
            "optimize",
            "review_change"
          ]
        },
        "description": {
          "type": "string",
          "maxLength": 1000
        },
        "target": {
          "type": "object",
          "properties": {
            "file": {"type": "string"},
            "function": {"type": "string"},
            "line_start": {"type": "integer"},
            "line_end": {"type": "integer"}
          }
        },
        "acceptance_criteria": {
          "type": "array",
          "items": {"type": "string"}
        },
        "deadline_step": {
          "type": "integer",
          "description": "Must be completed before this step"
        },
        "blocking": {
          "type": "boolean",
          "default": false,
          "description": "Whether this blocks the sender"
        }
      }
    },
    "requires_response": {
      "type": "boolean",
      "default": true
    }
  }
}
```

### Example Collaboration

**Reviewer → Coder:**
```json
{
  "id": "msg-collab-001",
  "type": "collaboration",
  "from": "reviewer-agent",
  "to": "go-coder-agent",
  "topic": "review",
  "payload": {
    "request": "fix_code",
    "description": "Error not wrapped with context, making debugging difficult",
    "target": {
      "file": "internal/service/user_service.go",
      "function": "CreateUser",
      "line_start": 45,
      "line_end": 48
    },
    "acceptance_criteria": [
      "Wrap error with fmt.Errorf and %w",
      "Include operation name in error message",
      "Tests still pass after change"
    ],
    "blocking": true
  },
  "priority": "high",
  "timestamp": "2025-12-28T22:20:00Z"
}
```

**Collaboration Response:**
```json
{
  "id": "msg-collab-001-resp",
  "type": "response",
  "from": "go-coder-agent",
  "to": "reviewer-agent",
  "in_reply_to": "msg-collab-001",
  "payload": {
    "answer": "Fixed error wrapping in CreateUser",
    "code": "if err != nil {\n\treturn fmt.Errorf(\"create user %s: %w\", user.Email, err)\n}",
    "reference": "internal/service/user_service.go:45-48"
  },
  "status": "success",
  "timestamp": "2025-12-28T22:21:30Z"
}
```

---

## Broadcast Message

Message sent to all agents.

```json
{
  "type": "broadcast",
  "properties": {
    "to": {
      "const": "*"
    },
    "payload": {
      "type": "object",
      "required": ["event"],
      "properties": {
        "event": {
          "type": "string",
          "enum": [
            "session_start",
            "session_end",
            "phase_change",
            "checkpoint_created",
            "rollback_initiated",
            "emergency_stop",
            "configuration_change"
          ]
        },
        "context": {
          "type": "object",
          "additionalProperties": true
        },
        "instructions": {
          "type": "string",
          "description": "Optional instructions for all agents"
        }
      }
    },
    "requires_response": {
      "const": false
    }
  }
}
```

### Example Broadcast

```json
{
  "id": "msg-broadcast-001",
  "type": "broadcast",
  "from": "orchestrator",
  "to": "*",
  "topic": "workflow",
  "payload": {
    "event": "phase_change",
    "context": {
      "from_phase": "implementation",
      "to_phase": "testing",
      "step": 5,
      "files_ready": [
        "internal/handler/user_handler.go",
        "internal/service/user_service.go",
        "internal/repo/user_repo.go"
      ],
      "metrics": {
        "lines_of_code": 450,
        "files_created": 8
      }
    },
    "instructions": "Implementation complete. Test Agent should begin test creation."
  },
  "priority": "normal",
  "timestamp": "2025-12-28T22:30:00Z"
}
```

---

## Acknowledgment Message

Confirm receipt of notification or broadcast.

```json
{
  "type": "ack",
  "properties": {
    "in_reply_to": {
      "type": "string"
    },
    "payload": {
      "type": "object",
      "properties": {
        "received": {
          "type": "boolean",
          "const": true
        },
        "will_act": {
          "type": "boolean",
          "description": "Whether agent will take action"
        },
        "notes": {
          "type": "string"
        }
      }
    }
  }
}
```

---

## Message Validation Rules

### Required Fields by Type

| Type | Required Fields |
|------|----------------|
| query | id, type, from, to, payload.question |
| response | id, type, from, to, in_reply_to, payload |
| notification | id, type, from, to, payload.event |
| collaboration | id, type, from, to, payload.request, payload.description |
| broadcast | id, type, from, to=*, payload.event |
| ack | id, type, from, to, in_reply_to |

### Priority Routing

| Priority | Max Latency | Queue Position |
|----------|-------------|----------------|
| critical | 100ms | Front |
| high | 500ms | After critical |
| normal | 2000ms | FIFO |
| low | 5000ms | Back |

---

## Error Schemas

### Delivery Error

```json
{
  "type": "error",
  "error_type": "delivery_failed",
  "original_message_id": "msg-xxx",
  "reason": "agent_unavailable",
  "retry_count": 2,
  "max_retries": 3,
  "next_retry_at": "2025-12-28T22:35:00Z"
}
```

### Timeout Error

```json
{
  "type": "error",
  "error_type": "response_timeout",
  "original_message_id": "msg-xxx",
  "waited_ms": 5000,
  "recipient": "architect-agent"
}
```

### Validation Error

```json
{
  "type": "error",
  "error_type": "validation_failed",
  "original_message_id": "msg-xxx",
  "errors": [
    {"field": "payload.question", "error": "required field missing"},
    {"field": "priority", "error": "invalid value: urgent"}
  ]
}
```
