---
stepNumber: "8b"
nextStep: './step-09-synthesis.md'
agent: doc-agent
hasBreakpoint: false
---

# Step 08b: Documentation

## STEP GOAL

Doc Agent creates/updates comprehensive project documentation including README.md, API docs, godoc comments, CHANGELOG, and Architecture Decision Records.

## AGENT ACTIVATION

Load persona từ `../agents/doc-agent.md`

Input context:
- Project structure từ implementation
- API endpoints (if any)
- Public functions và types
- Git history for CHANGELOG

## EXECUTION SEQUENCE

### 1. Doc Agent Introduction

```
[Doc Agent]

Generating documentation for "{topic}"...

Documentation targets:
1. README.md (overview, install, examples)
2. API documentation (if applicable)
3. Godoc comments (exported functions)
4. CHANGELOG.md
5. Architecture Decision Records (if significant decisions)
```

### 2. Analyze Codebase for Documentation

```bash
# Get module name
go list -m

# List all packages
go list ./...

# Count exported functions
go doc -all ./... 2>&1 | grep -c "^func"

# Check current documentation coverage
for pkg in $(go list ./...); do
    go doc "$pkg" 2>/dev/null | head -3
done
```

### 3. Generate README.md

Create or update README.md with:

```markdown
# {Project Name}

{Brief description in 1-2 sentences}

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

- Feature 1: Description
- Feature 2: Description
- Feature 3: Description

## Installation

```bash
go get github.com/{user}/{project}
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "{project}"
)

func main() {
    // Initialize
    client := project.New()

    // Use the library
    result, err := client.DoSomething(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %v", result)
}
```

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `Timeout` | `time.Duration` | `30s` | Request timeout |
| `RetryCount` | `int` | `3` | Number of retries |

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/{user}/{project})

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
```

### 4. Add/Update Godoc Comments

Check and add godoc comments for all exported functions:

```go
// FunctionName does X.
//
// It accepts Y and returns Z. If an error occurs,
// it returns a wrapped error with context.
//
// Example:
//
//	result, err := FunctionName(ctx, "input")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result)
func FunctionName(ctx context.Context, input string) (string, error)
```

### 5. Generate CHANGELOG

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of {feature}
- {List features from git log}

### Changed
- {List changes from git log}

### Fixed
- {List fixes from git log}
```

### 6. API Documentation (if applicable)

For HTTP APIs, create `docs/API.md`:

```markdown
# API Documentation

## Base URL

`http://localhost:8080/api/v1`

## Authentication

{Authentication method}

## Endpoints

### GET /endpoint

Description of the endpoint.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| param | string | Yes | Description |

**Response:**

```json
{
  "status": "success",
  "data": {}
}
```

**Error Codes:**

| Code | Description |
|------|-------------|
| 400 | Bad Request |
| 401 | Unauthorized |
| 500 | Internal Server Error |
```

### 7. Report Results

```
[Doc Agent]

Documentation complete:

### Files Created/Updated

| File | Action | Status |
|------|--------|--------|
| README.md | Created/Updated | ✅ |
| docs/API.md | Created | ✅ |
| CHANGELOG.md | Updated | ✅ |
| {N} godoc comments | Added | ✅ |

### Documentation Coverage

| Package | Exported | Documented | Coverage |
|---------|----------|------------|----------|
| handler | 12 | 12 | 100% |
| service | 8 | 8 | 100% |
| repo | 6 | 5 | 83% |
| **Total** | **26** | **25** | **96%** |

### Verification

```bash
$ go doc -all ./... 2>&1 | wc -l
156 lines of documentation

$ go test -run=Example ./...
ok  	example/handler	0.003s
```

Ready for final synthesis.
```

## OUTPUT

```yaml
outputs:
  documentation:
    readme: "README.md"
    api_docs: "docs/API.md"  # if applicable
    changelog: "CHANGELOG.md"
    godoc_comments: {count}
  coverage:
    total_exported: {count}
    total_documented: {count}
    percentage: {value}%
  verification:
    godoc_valid: true
    examples_compile: true
```

## QUALITY GATES

| Requirement | Threshold | Blocking |
|-------------|-----------|----------|
| README exists | Yes | Yes |
| README has install section | Yes | Yes |
| README has example | Yes | Yes |
| Exported functions documented | 80% | No |
| Package docs exist | 100% | No |

## SUCCESS CRITERIA

- [ ] README.md created/updated with all sections
- [ ] API documentation (if applicable)
- [ ] Godoc comments on exported functions (≥80%)
- [ ] CHANGELOG.md updated
- [ ] Examples compile and run

## KANBAN INTEGRATION

```yaml
on_step_start:
  signal: step_started
  payload:
    step: "step-08b"
    step_name: "documentation"
    agent: "doc-agent"

on_step_complete:
  signal: step_completed
  payload:
    step: "step-08b"
    agent: "doc-agent"
    outputs:
      - README.md
      - docs/*.md
      - CHANGELOG.md
    metrics:
      files_created: count
      coverage_percent: value
```

## CHECKPOINT

```yaml
checkpoint_config:
  id_format: "cp-08b-doc"
  includes:
    - documentation_report
    - files_created
    - coverage_metrics
  git_commit_message: "docs: update documentation"
```

## NEXT STEP

Load `./step-09-synthesis.md`
