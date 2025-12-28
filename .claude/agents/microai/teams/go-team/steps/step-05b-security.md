---
stepNumber: 5.5
nextStep: './step-06-review-loop.md'
previousStep: './step-05-testing.md'
agent: security-agent
hasBreakpoint: true
isGate: true
blockOnCritical: true
checkpoint:
  enabled: true
  id_format: "cp-05b-security"
  auto_create: true
---

# Step 05b: Security Audit

## STEP GOAL

Security Agent performs comprehensive security audit including SAST, dependency scanning, and secrets detection. This step acts as a **security gate** - critical/high vulnerabilities block the workflow.

---

## AGENT ACTIVATION

Load persona từ `../agents/security-agent.md`

---

## EXECUTION SEQUENCE

### 1. Security Agent Introduction

```
[Security Agent]

🔒 Starting Security Audit for "{topic}"...

I will check for:
- Vulnerabilities (OWASP Top 10)
- Hardcoded secrets
- Insecure dependencies
- Go-specific security issues

Running security tools...
```

### 2. Run Security Scans

#### 2.1 Static Analysis (SAST)

```bash
# Primary: gosec
gosec -fmt json -out gosec-report.json ./...

# If gosec not available, use grep patterns
# Check for SQL injection
grep -rn "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT" --include="*.go" .

# Check for command injection
grep -rn "exec.Command.*\+" --include="*.go" .
```

#### 2.2 Dependency Vulnerability Scan

```bash
# Primary: govulncheck
govulncheck -json ./... > vulncheck-report.json

# Backup: nancy
go list -m -json all | nancy sleuth
```

#### 2.3 Secret Detection

```bash
# Pattern-based secret scan
grep -rn \
  -e "password\s*=\s*[\"'][^\"']*[\"']" \
  -e "secret\s*=\s*[\"'][^\"']*[\"']" \
  -e "api_key\s*=\s*[\"'][^\"']*[\"']" \
  -e "token\s*=\s*[\"'][^\"']*[\"']" \
  -e "-----BEGIN.*PRIVATE KEY-----" \
  --include="*.go" --include="*.yaml" --include="*.json" .

# Check for common secret patterns
grep -rn \
  -e "AKIA[0-9A-Z]{16}" \
  -e "ghp_[a-zA-Z0-9]{36}" \
  -e "sk-[a-zA-Z0-9]{48}" \
  --include="*.go" --include="*.env" .
```

#### 2.4 Go-Specific Security Checks

```bash
# Check for unsafe package
grep -rn "\"unsafe\"" --include="*.go" .

# Check for weak crypto
grep -rn "crypto/md5\|crypto/sha1\|crypto/des" --include="*.go" .

# Check for math/rand (should use crypto/rand for security)
grep -rn "math/rand" --include="*.go" .
```

### 3. Manual Security Review

Security Agent reviews code for:

#### Authentication & Authorization
- [ ] Auth endpoints have rate limiting
- [ ] Password hashing uses bcrypt/argon2
- [ ] JWT secrets from environment
- [ ] Token expiration enforced

#### Input Validation
- [ ] All user input validated
- [ ] SQL queries parameterized
- [ ] No command injection possible
- [ ] Path traversal prevented

#### Data Protection
- [ ] No sensitive data in logs
- [ ] TLS for external connections
- [ ] Secure random generation

#### Error Handling
- [ ] No stack traces in responses
- [ ] Generic error messages to users

### 4. Generate Security Report

```markdown
## Security Audit Report - {topic}

### Executive Summary

| Metric | Value |
|--------|-------|
| Risk Level | {CRITICAL/HIGH/MEDIUM/LOW} |
| Total Issues | {count} |
| Critical | {count} |
| High | {count} |
| Medium | {count} |
| Low | {count} |

### Tool Results

#### gosec
{output or "Not available - manual checks performed"}

#### govulncheck
{output or "Not available"}

#### Secret Scan
{findings or "No secrets detected"}

### Vulnerabilities Found

#### Critical Issues

[SEC-CRIT-001] {title}
- File: {path}:{line}
- Category: {OWASP category}
- Description: {description}
- Fix: {remediation}

#### High Issues
...

#### Medium Issues
...

#### Low Issues
...

### Dependency Vulnerabilities

| Package | Vulnerable Version | Fixed Version | CVE |
|---------|-------------------|---------------|-----|
| {pkg}   | {version}         | {fixed}       | {cve} |

### Recommendations

1. Immediate (Critical/High): {actions}
2. Short-term (Medium): {actions}
3. Long-term (Low): {actions}
```

### 5. Security Gate Decision

```
IF critical_count > 0 OR high_count > 0:

  ⛔ SECURITY GATE FAILED
  ═══════════════════════════════════════════

  {critical_count} critical, {high_count} high vulnerabilities found.

  BLOCKING ISSUES:
  {list of critical/high issues}

  Action Required:
  - Return to Go Coder Agent
  - Fix all critical and high issues
  - Re-run security audit

  Commands:
  - [Enter] → Return to Implementation (fix issues)
  - *force-continue → Continue anyway (NOT RECOMMENDED)
  - *details → Show detailed vulnerability info

  ═══════════════════════════════════════════

ELSE IF medium_count > 0 OR low_count > 0:

  ⚠️ SECURITY REVIEW PASSED WITH WARNINGS
  ═══════════════════════════════════════════

  {medium_count} medium, {low_count} low issues found.

  Recommendation: Fix before production release.

  ═══════════════════════════════════════════
  ═══════════════ BREAKPOINT ═══════════════
  ═══════════════════════════════════════════

  Commands:
  - [Enter] → Continue to Review Loop
  - *fix → Return to fix medium/low issues first
  - *details → Show detailed findings

ELSE:

  ✅ SECURITY REVIEW PASSED
  ═══════════════════════════════════════════

  No security vulnerabilities detected!

  Code is secure for deployment.

  ═══════════════════════════════════════════
  ═══════════════ BREAKPOINT ═══════════════
  ═══════════════════════════════════════════

  Press [Enter] to continue to Review Loop...
```

---

## CHECKPOINT INTEGRATION

### Pre-Step

```markdown
Before security audit:

1. Verify previous checkpoint exists
   - cp-05-testing should exist

2. Display status
   ```
   Security Audit Prerequisites:
   - Tests passing: ✓
   - Code complete: ✓
   - Ready for security scan
   ```
```

### Post-Step

```markdown
After security audit complete:

1. Create checkpoint
   ```yaml
   checkpoint_data:
     step: 5.5
     step_name: "security"
     state:
       phase: "security"
       security_status: "PASS|FAIL|WARN"
     outputs:
       security_report: "{full report}"
       vulnerabilities:
         critical: []
         high: []
         medium: []
         low: []
       gate_result: "PASSED|BLOCKED|PASSED_WITH_WARNINGS"
   ```

2. Git commit
   ```bash
   git add -A
   git commit -m "checkpoint: step-05b - security audit

   Status: {PASSED|BLOCKED|WARNINGS}
   Critical: {count}
   High: {count}
   Medium: {count}
   Low: {count}
   "
   ```

3. Display confirmation
   ```
   ✓ Security checkpoint saved: cp-05b-security

   Gate Result: {result}
   Vulnerabilities: C:{n} H:{n} M:{n} L:{n}

   Rollback available: *rollback:5b
   ```
```

---

## OUTPUT

```yaml
outputs:
  security_report:
    risk_level: "CRITICAL|HIGH|MEDIUM|LOW"
    vulnerabilities:
      critical: [{...}]
      high: [{...}]
      medium: [{...}]
      low: [{...}]
    tool_results:
      gosec: "{output}"
      govulncheck: "{output}"
      secrets: "{output}"
    gate_result: "PASSED|BLOCKED|PASSED_WITH_WARNINGS"
    recommendations: [...]
```

---

## SUCCESS CRITERIA

- [ ] All security tools executed (or manual checks performed)
- [ ] Security report generated
- [ ] No critical vulnerabilities (gate requirement)
- [ ] No high vulnerabilities (gate requirement)
- [ ] Medium/low issues documented
- [ ] Checkpoint created
- [ ] Observer reviewed findings (breakpoint)

---

## FAILURE HANDLING

### If Security Tools Not Available

```
[Security Agent]

⚠️ Some security tools not installed.

Missing: {list}

Performing manual security checks instead...

For full scanning, install:
- gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest
- govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest
```

### If Scan Fails

```
[Security Agent]

⚠️ Security scan encountered errors.

Error: {error_message}

Options:
- *retry → Retry scan
- *manual → Perform manual review only
- *skip → Skip security (NOT RECOMMENDED)
```

---

## NEXT STEP

- If PASSED or PASSED_WITH_WARNINGS: Load `./step-06-review-loop.md`
- If BLOCKED: Return to `./step-04-implementation.md` for fixes
