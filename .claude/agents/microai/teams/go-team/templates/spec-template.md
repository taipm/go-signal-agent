# Specification: {Project Title}

**Date:** {date}
**Author:** PM Agent
**Status:** Draft | Review | Approved

---

## Overview

{Brief description of the project/feature}

---

## User Stories

### US-01: {Story Title}

**As a** {user persona},
**I want** {capability/feature},
**So that** {business value/benefit}.

#### Acceptance Criteria

**AC1: {Criterion Title}**
- **Given:** {precondition}
- **When:** {action}
- **Then:** {expected result}

**AC2: {Criterion Title}**
- **Given:** {precondition}
- **When:** {action}
- **Then:** {expected result}

---

### US-02: {Story Title}

**As a** {user persona},
**I want** {capability/feature},
**So that** {business value/benefit}.

#### Acceptance Criteria

**AC1: {Criterion Title}**
- **Given:** {precondition}
- **When:** {action}
- **Then:** {expected result}

---

## Scope

### In Scope
- {Feature 1}
- {Feature 2}
- {Feature 3}

### Out of Scope (Deferred)
- {Feature X} - Reason: {reason}
- {Feature Y} - Reason: {reason}

### Assumptions
- {Assumption 1}
- {Assumption 2}

### Dependencies
- {External system/service}
- {Library/package}

---

## API Contract

### Endpoints

#### GET /api/v1/{resource}/{id}

**Description:** Retrieve a {resource} by ID

**Request:**
```
GET /api/v1/{resource}/123
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "id": "123",
  "name": "Example",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
| Status | Description |
|--------|-------------|
| 400 | Invalid request |
| 401 | Unauthorized |
| 404 | Not found |
| 500 | Internal error |

---

#### POST /api/v1/{resource}

**Description:** Create a new {resource}

**Request:**
```json
{
  "name": "Example",
  "description": "Optional description"
}
```

**Response (201 Created):**
```json
{
  "id": "124",
  "name": "Example",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

## Data Model

### {Entity}

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | Yes | Unique identifier |
| name | string | Yes | Display name |
| description | string | No | Optional description |
| created_at | datetime | Yes | Creation timestamp |
| updated_at | datetime | Yes | Last update timestamp |

---

## Non-Functional Requirements

### Performance
- Response time: < 200ms (p95)
- Throughput: 100 requests/second

### Security
- Authentication: Bearer token
- Authorization: Role-based

### Reliability
- Availability: 99.9%
- Error rate: < 0.1%

---

## Open Questions

- [ ] {Question 1}?
- [ ] {Question 2}?

---

## Approval

| Role | Name | Date | Status |
|------|------|------|--------|
| Observer | | | Pending |
| Architect | | | Pending |
