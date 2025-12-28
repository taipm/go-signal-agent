---
story_id: "{{story_id}}"
created_date: "{{date}}"
session_subject: "{{subject}}"
status: "{{status}}"
participants:
  - Solo Developer
  - End User
  - Observer: {{observer_name}}
---

# User Story: {{story_title}}

## Story Definition

**As a** {{user_persona}},
**I want** {{capability}},
**So that** {{business_value}}.

---

## Acceptance Criteria

{{#each acceptance_criteria}}
### AC{{@index_plus_one}}: {{this.title}}

| Element | Description |
|---------|-------------|
| **Given** | {{this.given}} |
| **When** | {{this.when}} |
| **Then** | {{this.then}} |
{{#if this.and}}
| **And** | {{this.and}} |
{{/if}}

{{/each}}

---

## Scope Notes

### In Scope
{{#each in_scope}}
- {{this}}
{{/each}}

### Out of Scope (Deferred)
{{#each out_of_scope}}
- {{this}}
{{/each}}

---

## Technical Notes

{{technical_notes}}

---

## Dependencies

{{#if dependencies}}
{{#each dependencies}}
- {{this}}
{{/each}}
{{else}}
No external dependencies identified.
{{/if}}

---

## Estimation

| Attribute | Value |
|-----------|-------|
| **Complexity** | {{complexity}} |
| **Estimated Effort** | {{effort}} |

---

## Sign-Off

| Role | Status | Notes |
|------|--------|-------|
| End User | {{#if enduser_signoff}}✅ Approved{{else}}⏳ Pending{{/if}} | {{enduser_notes}} |
| Solo Developer | {{#if dev_signoff}}✅ Confirmed{{else}}⏳ Pending{{/if}} | {{dev_notes}} |
| Observer | {{#if observer_signoff}}✅ Approved{{else}}⏳ Pending{{/if}} | {{observer_notes}} |

---

## Metadata

| Field | Value |
|-------|-------|
| Story ID | {{story_id}} |
| Created | {{date}} |
| Session Topic | {{subject}} |
| Status | {{status}} |

---

*Generated from dev-user team session on {{date}}*
