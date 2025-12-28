# Code Review Report

**Project:** {project_name}
**Iteration:** {iteration_number}
**Date:** {date}
**Reviewer:** Reviewer Agent

---

## Summary

| Category | Count |
|----------|-------|
| Critical Issues | {count} |
| Warnings | {count} |
| Suggestions | {count} |

**Overall Status:** PASS | NEEDS_FIX | BLOCKED

---

## Tool Results

### go vet

```
{output}
```

**Status:** PASS | FAIL

### golangci-lint

```
{output}
```

**Status:** PASS | FAIL

### go test -race

```
{output}
```

**Status:** PASS | FAIL

### Test Coverage

```
{coverage output}
```

**Coverage:** {percentage}%
**Target:** 80%
**Status:** PASS | FAIL

---

## Critical Issues (Must Fix)

### [C1] {Issue Title}

**File:** `{file_path}:{line_number}`

**Category:** Race Condition | Error Handling | Security | Resource Leak

**Issue:**
{Description of the problem}

**Code:**
```go
// Current (problematic)
{code snippet}
```

**Fix:**
```go
// Suggested fix
{fixed code snippet}
```

**Impact:** HIGH

---

### [C2] {Issue Title}

**File:** `{file_path}:{line_number}`

**Category:** {category}

**Issue:**
{Description}

**Fix:**
{Suggestion}

---

## Warnings (Should Fix)

### [W1] {Issue Title}

**File:** `{file_path}:{line_number}`

**Category:** Performance | Style | Maintainability

**Issue:**
{Description}

**Suggestion:**
{Recommendation}

---

### [W2] {Issue Title}

**File:** `{file_path}:{line_number}`

**Issue:**
{Description}

---

## Suggestions (Nice to Have)

### [S1] {Suggestion Title}

**File:** `{file_path}`

**Suggestion:**
{Description of improvement}

---

## Checklist

### Concurrency
- [ ] Shared state protected by mutex or channels
- [ ] Context passed to all blocking operations
- [ ] Goroutines have clear termination conditions
- [ ] No unbounded goroutine creation

### Error Handling
- [ ] All errors checked
- [ ] Errors wrapped with context (fmt.Errorf %w)
- [ ] No swallowed errors
- [ ] Appropriate error messages

### Resource Management
- [ ] Files/connections closed with defer
- [ ] HTTP response bodies closed
- [ ] Contexts cancelled when done
- [ ] Timeouts on external calls

### Security
- [ ] No hardcoded credentials
- [ ] Input validated before use
- [ ] SQL uses parameterized queries
- [ ] Sensitive data not logged

### Style
- [ ] gofmt applied
- [ ] Meaningful variable names
- [ ] Functions under 50 lines
- [ ] No dead code

---

## Metrics

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| Build | {pass/fail} | PASS | {status} |
| Tests | {passed}/{total} | 100% | {status} |
| Coverage | {percentage}% | 80% | {status} |
| Lint | {issues} | 0 | {status} |
| Race Detection | {pass/fail} | PASS | {status} |

---

## Next Steps

{If NEEDS_FIX:}
1. Coder Agent fixes critical issues
2. Re-run review (iteration {n+1})

{If PASS:}
1. Proceed to Optimization phase

---

## History

| Iteration | Date | Critical | Warnings | Status |
|-----------|------|----------|----------|--------|
| 1 | {date} | {count} | {count} | {status} |
| 2 | {date} | {count} | {count} | {status} |
