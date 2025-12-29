# Token & Cost Tracking System

**Version:** 1.0.0

Theo dõi token usage và chi phí cho Go Team sessions.

---

## Quick Start

```bash
# Show token usage summary
*tokens

# Show detailed breakdown
*tokens:detail

# Show cost estimate
*cost

# Export metrics
*tokens:export
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    TOKEN TRACKING ENGINE                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Token Counter   │  │ Cost Calculator │  │ Metrics     │ │
│  │                 │  │                 │  │ Aggregator  │ │
│  │ • Input tokens  │  │ • Per-model     │  │ • Per-step  │ │
│  │ • Output tokens │  │ • Per-agent     │  │ • Per-agent │ │
│  │ • Cache hits    │  │ • Total cost    │  │ • Trends    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      TRACKING LAYER                          │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  PM  │ ARCH │ CODER │ TEST │ SEC │ REV │ OPT │ DEV     ││
│  │  ↓   │  ↓   │   ↓   │  ↓   │  ↓  │  ↓  │  ↓  │  ↓      ││
│  │ [Token meters attached to each agent invocation]        ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

---

## Token Tracking

### Per-Agent Tracking

```json
{
  "agent_tokens": {
    "pm-agent": {
      "invocations": 3,
      "input_tokens": 15000,
      "output_tokens": 8000,
      "cached_tokens": 5000,
      "total_tokens": 23000
    },
    "architect-agent": {
      "invocations": 2,
      "input_tokens": 20000,
      "output_tokens": 12000,
      "cached_tokens": 8000,
      "total_tokens": 32000
    },
    "go-coder-agent": {
      "invocations": 8,
      "input_tokens": 45000,
      "output_tokens": 35000,
      "cached_tokens": 15000,
      "total_tokens": 80000
    }
  }
}
```

### Per-Step Tracking

```json
{
  "step_tokens": {
    "step-01-init": {
      "tokens": 5000,
      "duration_seconds": 30,
      "agents_involved": ["orchestrator"]
    },
    "step-02-requirements": {
      "tokens": 25000,
      "duration_seconds": 180,
      "agents_involved": ["pm-agent"]
    },
    "step-03-architecture": {
      "tokens": 35000,
      "duration_seconds": 240,
      "agents_involved": ["architect-agent"]
    },
    "step-04-implementation": {
      "tokens": 85000,
      "duration_seconds": 600,
      "agents_involved": ["go-coder-agent"]
    },
    "step-05-testing": {
      "tokens": 40000,
      "duration_seconds": 300,
      "agents_involved": ["test-agent"]
    },
    "step-05b-security": {
      "tokens": 15000,
      "duration_seconds": 120,
      "agents_involved": ["security-agent"]
    },
    "step-06-review": {
      "tokens": 60000,
      "duration_seconds": 450,
      "iterations": 2,
      "agents_involved": ["reviewer-agent", "go-coder-agent", "test-agent"]
    }
  }
}
```

---

## Cost Estimation

### Model Pricing (as of Dec 2024)

```yaml
pricing:
  claude-3-5-sonnet:
    input_per_1m: 3.00      # $3 per 1M input tokens
    output_per_1m: 15.00    # $15 per 1M output tokens
    cached_per_1m: 0.30     # $0.30 per 1M cached tokens

  claude-3-opus:
    input_per_1m: 15.00     # $15 per 1M input tokens
    output_per_1m: 75.00    # $75 per 1M output tokens
    cached_per_1m: 1.50     # $1.50 per 1M cached tokens

  claude-3-haiku:
    input_per_1m: 0.25      # $0.25 per 1M input tokens
    output_per_1m: 1.25     # $1.25 per 1M output tokens
    cached_per_1m: 0.03     # $0.03 per 1M cached tokens
```

### Cost Calculation

```python
def calculate_cost(tokens, model="claude-3-5-sonnet"):
    pricing = get_pricing(model)

    input_cost = (tokens.input / 1_000_000) * pricing.input_per_1m
    output_cost = (tokens.output / 1_000_000) * pricing.output_per_1m
    cached_cost = (tokens.cached / 1_000_000) * pricing.cached_per_1m

    return {
        "input_cost": input_cost,
        "output_cost": output_cost,
        "cached_cost": cached_cost,
        "total_cost": input_cost + output_cost + cached_cost,
        "savings_from_cache": (tokens.cached / 1_000_000) * (pricing.input_per_1m - pricing.cached_per_1m)
    }
```

---

## Session Metrics

### Complete Session Record

```json
{
  "session_id": "go-team-session-001",
  "topic": "user-authentication-service",
  "started_at": "2025-12-28T22:00:00Z",
  "completed_at": "2025-12-28T23:30:00Z",
  "duration_minutes": 90,

  "token_summary": {
    "total_input": 180000,
    "total_output": 120000,
    "total_cached": 45000,
    "total_tokens": 300000,
    "cache_hit_rate": 0.15
  },

  "cost_summary": {
    "model": "claude-3-5-sonnet",
    "input_cost": 0.54,
    "output_cost": 1.80,
    "cached_cost": 0.014,
    "total_cost": 2.354,
    "savings_from_cache": 0.122
  },

  "efficiency_metrics": {
    "tokens_per_minute": 3333,
    "cost_per_minute": 0.026,
    "lines_of_code_generated": 850,
    "cost_per_loc": 0.0028
  },

  "by_agent": {
    "pm-agent": { "tokens": 25000, "cost": 0.20, "percentage": 8.3 },
    "architect-agent": { "tokens": 35000, "cost": 0.28, "percentage": 11.7 },
    "go-coder-agent": { "tokens": 120000, "cost": 0.96, "percentage": 40.0 },
    "test-agent": { "tokens": 45000, "cost": 0.36, "percentage": 15.0 },
    "security-agent": { "tokens": 20000, "cost": 0.16, "percentage": 6.7 },
    "reviewer-agent": { "tokens": 40000, "cost": 0.32, "percentage": 13.3 },
    "optimizer-agent": { "tokens": 10000, "cost": 0.08, "percentage": 3.3 },
    "devops-agent": { "tokens": 5000, "cost": 0.04, "percentage": 1.7 }
  },

  "by_step": {
    "init": { "tokens": 5000, "percentage": 1.7 },
    "requirements": { "tokens": 25000, "percentage": 8.3 },
    "architecture": { "tokens": 35000, "percentage": 11.7 },
    "implementation": { "tokens": 120000, "percentage": 40.0 },
    "testing": { "tokens": 45000, "percentage": 15.0 },
    "security": { "tokens": 20000, "percentage": 6.7 },
    "review": { "tokens": 40000, "percentage": 13.3 },
    "optimization": { "tokens": 10000, "percentage": 3.3 }
  }
}
```

---

## Commands

### Basic Commands

| Command | Description |
|---------|-------------|
| `*tokens` | Show token usage summary |
| `*tokens:detail` | Show detailed breakdown by agent/step |
| `*tokens:agent:{name}` | Show tokens for specific agent |
| `*tokens:step:{N}` | Show tokens for specific step |

### Cost Commands

| Command | Description |
|---------|-------------|
| `*cost` | Show cost estimate |
| `*cost:detail` | Show detailed cost breakdown |
| `*cost:compare` | Compare with previous sessions |
| `*cost:budget` | Show budget status (if set) |

### Budget Commands

| Command | Description |
|---------|-------------|
| `*budget:set {amount}` | Set session budget limit |
| `*budget:warn {percentage}` | Set warning threshold (default: 80%) |
| `*budget:status` | Show budget status |
| `*budget:clear` | Clear budget limit |

### Export Commands

| Command | Description |
|---------|-------------|
| `*tokens:export` | Export to JSON |
| `*tokens:export:csv` | Export to CSV |
| `*tokens:history` | Show historical usage |

---

## Budget Management

### Setting Budget

```bash
# Set $5 budget for session
*budget:set 5.00

# Warn at 80% usage
*budget:warn 80
```

### Budget Alerts

```
═══════════════════════════════════════════════════════════
⚠️  BUDGET WARNING
═══════════════════════════════════════════════════════════

Budget: $5.00
Used: $4.12 (82.4%)
Remaining: $0.88

Estimated to complete: $1.50 more needed

Options:
- *budget:add 2.00  → Increase budget
- *continue         → Continue anyway
- *pause            → Pause for decision
- *abort            → Abort session

═══════════════════════════════════════════════════════════
```

### Budget Exceeded

```
═══════════════════════════════════════════════════════════
🛑 BUDGET EXCEEDED
═══════════════════════════════════════════════════════════

Budget: $5.00
Used: $5.23 (104.6%)
Over by: $0.23

Current step: 6 (Review Loop)
Progress: 78% complete

Options:
- *budget:add 2.00  → Add to budget and continue
- *save             → Save progress and exit
- *abort            → Abort session

═══════════════════════════════════════════════════════════
```

---

## Display Formats

### Summary View (*tokens)

```
═══════════════════════════════════════════════════════════
📊 TOKEN USAGE SUMMARY
═══════════════════════════════════════════════════════════

Session: go-team-session-001
Topic: user-authentication-service
Duration: 45 minutes

┌─────────────────────────────────────────────────────────┐
│ TOKENS                                                   │
├─────────────────────────────────────────────────────────┤
│ Input:    180,000 tokens                                │
│ Output:   120,000 tokens                                │
│ Cached:    45,000 tokens (15% cache hit)                │
│ ─────────────────────────────────────────────────       │
│ Total:    300,000 tokens                                │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ COST (claude-3-5-sonnet)                                │
├─────────────────────────────────────────────────────────┤
│ Input:    $0.54                                         │
│ Output:   $1.80                                         │
│ Cached:   $0.01                                         │
│ ─────────────────────────────────────────────────       │
│ Total:    $2.35                                         │
│ Saved:    $0.12 (from cache)                            │
└─────────────────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════
```

### Detail View (*tokens:detail)

```
═══════════════════════════════════════════════════════════
📊 TOKEN USAGE - DETAILED BREAKDOWN
═══════════════════════════════════════════════════════════

BY AGENT:
┌─────────────────┬──────────┬─────────┬────────┬─────────┐
│ Agent           │ Input    │ Output  │ Total  │ Cost    │
├─────────────────┼──────────┼─────────┼────────┼─────────┤
│ go-coder-agent  │  80,000  │ 40,000  │120,000 │  $0.96  │
│ test-agent      │  30,000  │ 15,000  │ 45,000 │  $0.36  │
│ reviewer-agent  │  25,000  │ 15,000  │ 40,000 │  $0.32  │
│ architect-agent │  20,000  │ 15,000  │ 35,000 │  $0.28  │
│ pm-agent        │  15,000  │ 10,000  │ 25,000 │  $0.20  │
│ security-agent  │  12,000  │  8,000  │ 20,000 │  $0.16  │
│ optimizer-agent │   6,000  │  4,000  │ 10,000 │  $0.08  │
│ devops-agent    │   3,000  │  2,000  │  5,000 │  $0.04  │
├─────────────────┼──────────┼─────────┼────────┼─────────┤
│ TOTAL           │ 191,000  │109,000  │300,000 │  $2.35  │
└─────────────────┴──────────┴─────────┴────────┴─────────┘

BY STEP:
┌─────────────────┬──────────┬─────────┬────────┬─────────┐
│ Step            │ Tokens   │    %    │ Time   │ Cost    │
├─────────────────┼──────────┼─────────┼────────┼─────────┤
│ 4. Implement    │ 120,000  │  40.0%  │ 15m    │  $0.96  │
│ 5. Testing      │  45,000  │  15.0%  │  8m    │  $0.36  │
│ 6. Review       │  40,000  │  13.3%  │ 10m    │  $0.32  │
│ 3. Architecture │  35,000  │  11.7%  │  6m    │  $0.28  │
│ 2. Requirements │  25,000  │   8.3%  │  4m    │  $0.20  │
│ 5b. Security    │  20,000  │   6.7%  │  3m    │  $0.16  │
│ 7. Optimization │  10,000  │   3.3%  │  2m    │  $0.08  │
│ 1. Init         │   5,000  │   1.7%  │  1m    │  $0.04  │
└─────────────────┴──────────┴─────────┴────────┴─────────┘

═══════════════════════════════════════════════════════════
```

---

## Integration with Workflow

### State Extension

```yaml
go_team_state:
  # ... existing fields ...

  metrics:
    tokens:
      total_input: 0
      total_output: 0
      total_cached: 0
      by_agent: {}
      by_step: {}

    cost:
      model: "claude-3-5-sonnet"
      total: 0.0
      budget: null
      budget_warning: 0.8

    timing:
      started_at: null
      last_updated: null
      total_duration_seconds: 0
```

### Tracking Hook

```yaml
on_agent_complete:
  1. Capture token usage from response
  2. Update agent metrics
  3. Update step metrics
  4. Calculate running cost
  5. Check budget (if set)
  6. Display running total (optional)
```

---

## Historical Tracking

### Session History

```json
{
  "history": [
    {
      "session_id": "session-001",
      "date": "2025-12-28",
      "topic": "user-auth",
      "tokens": 300000,
      "cost": 2.35,
      "duration_min": 45
    },
    {
      "session_id": "session-002",
      "date": "2025-12-27",
      "topic": "payment-api",
      "tokens": 450000,
      "cost": 3.52,
      "duration_min": 60
    }
  ],
  "totals": {
    "sessions": 2,
    "tokens": 750000,
    "cost": 5.87,
    "avg_cost_per_session": 2.94
  }
}
```

### Trend Analysis

```
═══════════════════════════════════════════════════════════
📈 USAGE TRENDS (Last 7 days)
═══════════════════════════════════════════════════════════

Daily Cost:
Dec 22: $2.15  ████████░░░░░░░░
Dec 23: $3.45  █████████████░░░
Dec 24: $1.80  ██████░░░░░░░░░░
Dec 25: $0.00  ░░░░░░░░░░░░░░░░
Dec 26: $4.20  ████████████████
Dec 27: $3.52  █████████████░░░
Dec 28: $2.35  █████████░░░░░░░

Total: $17.47
Avg:   $2.50/day

Most expensive step: Implementation (40%)
Most efficient: Init (1.7%)

═══════════════════════════════════════════════════════════
```

---

## Benefits

1. **Cost Visibility:** Know exactly how much each session costs
2. **Budget Control:** Set limits and receive warnings
3. **Optimization Insights:** Identify expensive agents/steps
4. **Historical Analysis:** Track trends over time
5. **ROI Calculation:** Cost per line of code generated

---

## Limitations

1. Token counts are estimates (actual may vary slightly)
2. Cached token detection depends on API response
3. Cost is based on published pricing (may change)
4. Does not track external tool costs (e.g., GitHub Actions)
