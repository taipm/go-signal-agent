# Security Audit Report

## Project Information

| Field | Value |
|-------|-------|
| **Project** | {{topic}} |
| **Date** | {{date}} |
| **Auditor** | Security Agent |
| **Session** | {{session_id}} |

---

## Executive Summary

| Metric | Value |
|--------|-------|
| **Overall Risk Level** | {{risk_level}} |
| **Gate Status** | {{gate_status}} |
| **Total Vulnerabilities** | {{total_count}} |
| **Critical** | {{critical_count}} |
| **High** | {{high_count}} |
| **Medium** | {{medium_count}} |
| **Low** | {{low_count}} |

### Risk Assessment

```
{{risk_level}} RISK

{{#if critical_count}}
⛔ {{critical_count}} CRITICAL vulnerabilities require immediate attention
{{/if}}
{{#if high_count}}
🔴 {{high_count}} HIGH vulnerabilities should be fixed before release
{{/if}}
{{#if medium_count}}
🟠 {{medium_count}} MEDIUM vulnerabilities should be tracked
{{/if}}
{{#if low_count}}
🟡 {{low_count}} LOW/informational findings
{{/if}}
```

---

## Tool Results

### gosec (Static Analysis)

```
{{gosec_output}}
```

**Summary:** {{gosec_summary}}

### govulncheck (Dependency Vulnerabilities)

```
{{govulncheck_output}}
```

**Summary:** {{govulncheck_summary}}

### Secret Scan

```
{{secret_scan_output}}
```

**Summary:** {{secret_scan_summary}}

---

## Vulnerability Details

### Critical Vulnerabilities

{{#each critical_vulnerabilities}}

#### [SEC-CRIT-{{@index}}] {{this.title}}

| Field | Value |
|-------|-------|
| **Severity** | 🔴 CRITICAL |
| **Category** | {{this.category}} |
| **CWE** | {{this.cwe}} |
| **File** | `{{this.file}}:{{this.line}}` |
| **CVSS** | {{this.cvss}} |

**Description:**
{{this.description}}

**Vulnerable Code:**

```go
{{this.vulnerable_code}}
```

**Remediation:**

```go
{{this.fixed_code}}
```

**References:**
{{#each this.references}}
- {{this}}
{{/each}}

---

{{/each}}

### High Vulnerabilities

{{#each high_vulnerabilities}}

#### [SEC-HIGH-{{@index}}] {{this.title}}

| Field | Value |
|-------|-------|
| **Severity** | 🟠 HIGH |
| **Category** | {{this.category}} |
| **CWE** | {{this.cwe}} |
| **File** | `{{this.file}}:{{this.line}}` |

**Description:**
{{this.description}}

**Remediation:**
{{this.remediation}}

---

{{/each}}

### Medium Vulnerabilities

{{#each medium_vulnerabilities}}

#### [SEC-MED-{{@index}}] {{this.title}}

| Field | Value |
|-------|-------|
| **Severity** | 🟡 MEDIUM |
| **Category** | {{this.category}} |
| **File** | `{{this.file}}:{{this.line}}` |

**Description:**
{{this.description}}

**Remediation:**
{{this.remediation}}

---

{{/each}}

### Low / Informational

{{#each low_vulnerabilities}}

#### [SEC-LOW-{{@index}}] {{this.title}}

- **Category:** {{this.category}}
- **File:** `{{this.file}}`
- **Description:** {{this.description}}
- **Suggestion:** {{this.suggestion}}

{{/each}}

---

## Dependency Vulnerabilities

| Package | Current Version | Vulnerable Range | Fixed Version | CVE | Severity |
|---------|-----------------|------------------|---------------|-----|----------|
{{#each dependency_vulnerabilities}}
| {{this.package}} | {{this.current}} | {{this.vulnerable_range}} | {{this.fixed}} | {{this.cve}} | {{this.severity}} |
{{/each}}

---

## Security Checklist

### Authentication & Authorization

- [{{auth_rate_limiting}}] Rate limiting on auth endpoints
- [{{password_hashing}}] Secure password hashing (bcrypt/argon2)
- [{{jwt_secrets}}] JWT secrets from environment
- [{{token_expiration}}] Token expiration implemented

### Input Validation

- [{{input_validation}}] All user input validated
- [{{sql_parameterized}}] SQL queries parameterized
- [{{command_injection}}] Command injection prevented
- [{{path_traversal}}] Path traversal prevented

### Data Protection

- [{{sensitive_logs}}] No sensitive data in logs
- [{{tls_enabled}}] TLS for external connections
- [{{crypto_rand}}] Secure random generation (crypto/rand)

### Error Handling

- [{{no_stack_traces}}] No stack traces in responses
- [{{generic_errors}}] Generic error messages to users

### HTTP Security

- [{{cors_configured}}] CORS properly configured
- [{{security_headers}}] Security headers present
- [{{https_enforced}}] HTTPS enforced

---

## Recommendations

### Immediate Actions (Critical/High)

{{#each immediate_actions}}
{{@index}}. {{this}}
{{/each}}

### Short-term Improvements (Medium)

{{#each short_term_actions}}
{{@index}}. {{this}}
{{/each}}

### Long-term Hardening (Low)

{{#each long_term_actions}}
{{@index}}. {{this}}
{{/each}}

---

## Compliance Notes

| Standard | Status | Notes |
|----------|--------|-------|
| OWASP Top 10 | {{owasp_status}} | {{owasp_notes}} |
| CWE Top 25 | {{cwe_status}} | {{cwe_notes}} |
| No Hardcoded Secrets | {{secrets_status}} | {{secrets_notes}} |
| Dependencies Updated | {{deps_status}} | {{deps_notes}} |

---

## Gate Decision

```
{{#if gate_passed}}
✅ SECURITY GATE PASSED

No critical or high vulnerabilities detected.
Code is approved for deployment.
{{else if gate_passed_with_warnings}}
⚠️ SECURITY GATE PASSED WITH WARNINGS

{{medium_count}} medium and {{low_count}} low issues found.
Recommend fixing before production release.
Code may proceed to next step.
{{else}}
⛔ SECURITY GATE BLOCKED

{{critical_count}} critical and {{high_count}} high vulnerabilities found.
Code MUST NOT proceed until issues are resolved.

Required Actions:
{{#each blocking_issues}}
- Fix: {{this.title}} ({{this.file}})
{{/each}}
{{/if}}
```

---

## Appendix

### Security Tools Versions

| Tool | Version | Status |
|------|---------|--------|
| gosec | {{gosec_version}} | {{gosec_status}} |
| govulncheck | {{govulncheck_version}} | {{govulncheck_status}} |
| Go | {{go_version}} | - |

### Scan Configuration

```yaml
scan_config:
  exclude_paths:
    - vendor/
    - testdata/
    - *_test.go
  severity_threshold: HIGH
  fail_on_severity: [CRITICAL, HIGH]
```

---

*Generated by Security Agent on {{date}}*
*Session: {{session_id}}*
