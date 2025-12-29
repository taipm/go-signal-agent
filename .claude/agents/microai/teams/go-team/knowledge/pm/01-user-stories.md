# User Stories - PM Agent Knowledge

**Version:** 1.0.0
**Agent:** PM Agent

---

## TL;DR

- Format: "As a [role], I want [feature], so that [benefit]"
- Acceptance criteria dùng Given-When-Then
- INVEST: Independent, Negotiable, Valuable, Estimable, Small, Testable
- Mỗi story phải có clear DoD (Definition of Done)

---

## 1. User Story Format

### Standard Template

```markdown
## User Story: [Short Title]

**As a** [type of user]
**I want** [some goal]
**So that** [some reason/benefit]

### Acceptance Criteria

**Given** [precondition]
**When** [action]
**Then** [expected result]

### Technical Notes
- [Implementation hints]
- [Dependencies]

### Out of Scope
- [What is NOT included]
```

### Example

```markdown
## User Story: User Registration

**As a** new visitor
**I want** to create an account with email and password
**So that** I can access personalized features

### Acceptance Criteria

**AC1: Successful Registration**
Given I am on the registration page
When I enter a valid email "user@example.com" and password "SecurePass123!"
Then my account is created
And I receive a confirmation email
And I am redirected to the dashboard

**AC2: Duplicate Email**
Given an account with email "existing@example.com" already exists
When I try to register with the same email
Then I see error "Email already registered"
And no new account is created

**AC3: Weak Password**
Given I am on the registration page
When I enter password "123"
Then I see error "Password must be at least 8 characters with uppercase, lowercase, and number"
And no account is created

### Technical Notes
- Password hashing: bcrypt with cost 12
- Email verification required before login
- Rate limit: 5 registration attempts per IP per hour

### Out of Scope
- Social login (OAuth)
- Phone number registration
- Two-factor authentication setup
```

---

## 2. INVEST Criteria

### Independent

```markdown
# ✅ GOOD - Independent stories
Story A: User can register with email
Story B: User can login with email
Story C: User can reset password

# ❌ BAD - Coupled stories
Story A: User can register (requires Story B)
Story B: Setup email service (blocks Story A and C)
Story C: User can reset password (requires Story B)
```

### Negotiable

```markdown
# ✅ GOOD - Leaves room for discussion
"User can view their order history"
- How many orders to show? (discuss with team)
- Pagination or infinite scroll? (discuss UX)
- Filter options? (can be separate story)

# ❌ BAD - Too prescriptive
"Display exactly 10 orders per page with left-aligned blue pagination buttons at 12px font"
```

### Valuable

```markdown
# ✅ GOOD - Clear value
"As a customer, I want to track my order so that I know when it will arrive"
→ Reduces support calls, improves customer satisfaction

# ❌ BAD - No clear value
"As a developer, I want to refactor the database layer"
→ Should be technical task, not user story
```

### Estimable

```markdown
# ✅ GOOD - Can be estimated
"User can search products by name"
- Clear scope
- Team understands requirements
- Similar to previous features

# ❌ BAD - Cannot estimate
"Improve system performance"
- What does "improve" mean?
- Which part of the system?
- What's the target?
```

### Small

```markdown
# ✅ GOOD - Fits in a sprint
"User can add item to cart"
- Single focused feature
- Clear boundaries
- Testable in isolation

# ❌ BAD - Too large (Epic)
"User can complete entire checkout process"
→ Break into:
- Add to cart
- View cart
- Update quantities
- Apply coupon
- Enter shipping
- Select payment
- Place order
- Receive confirmation
```

### Testable

```markdown
# ✅ GOOD - Testable
Given I have items in cart
When I click "Place Order"
Then order is created with status "pending"
And I see order confirmation number

# ❌ BAD - Not testable
"System should be fast"
"User experience should be good"
"Application should be secure"
```

---

## 3. Acceptance Criteria Patterns

### Given-When-Then (BDD Style)

```markdown
**Scenario: Valid Login**
Given I am a registered user with email "user@example.com"
And my password is "correct_password"
When I submit the login form
Then I am redirected to the dashboard
And I see welcome message "Hello, User"
And my session is created with 24h expiry

**Scenario: Invalid Password**
Given I am a registered user
When I enter wrong password 3 times
Then my account is locked for 15 minutes
And I receive security alert email
```

### Rule-Based

```markdown
**Rules:**
1. Password minimum 8 characters
2. Must contain uppercase, lowercase, number
3. Cannot be same as email
4. Cannot be in common password list

**Examples:**
| Password       | Valid | Reason                    |
|---------------|-------|---------------------------|
| Secret123     | ✅    | Meets all requirements    |
| secret123     | ❌    | No uppercase              |
| SECRET123     | ❌    | No lowercase              |
| SecretABC     | ❌    | No number                 |
| Sec1          | ❌    | Too short                 |
| user@test.com | ❌    | Same as email             |
| Password123   | ❌    | Common password           |
```

### Checklist Style

```markdown
**Acceptance Checklist:**
- [ ] Form validates email format
- [ ] Form shows password strength indicator
- [ ] Submit button disabled until form valid
- [ ] Loading state shown during submission
- [ ] Success redirects to dashboard
- [ ] Error displays friendly message
- [ ] Works on mobile (responsive)
- [ ] Accessible (screen reader compatible)
```

---

## 4. API Contract Definition

### REST Endpoint Spec

```markdown
## API: Create User

**Endpoint:** `POST /api/v1/users`

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "id": "usr_abc123",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
| Status | Code | Message |
|--------|------|---------|
| 400 | INVALID_EMAIL | "Invalid email format" |
| 400 | WEAK_PASSWORD | "Password does not meet requirements" |
| 409 | EMAIL_EXISTS | "Email already registered" |
| 429 | RATE_LIMITED | "Too many requests" |
```

---

## 5. Story Mapping

### Feature Breakdown

```
Epic: User Authentication
│
├── Story: Registration
│   ├── Task: Create registration form
│   ├── Task: Implement email validation
│   ├── Task: Implement password hashing
│   └── Task: Send confirmation email
│
├── Story: Login
│   ├── Task: Create login form
│   ├── Task: Implement JWT generation
│   └── Task: Implement session management
│
├── Story: Password Reset
│   ├── Task: Create forgot password form
│   ├── Task: Generate reset token
│   └── Task: Implement password update
│
└── Story: Logout
    └── Task: Invalidate session/token
```

### Priority Matrix

```
                    High Value
                        │
    Quick Wins    │    Do First
    ─────────────────────────────────
    Low Effort ────┼──── High Effort
    ─────────────────────────────────
    Fill Ins      │    Major Projects
                        │
                    Low Value
```

---

## 6. Definition of Done (DoD)

### Standard DoD

```markdown
## Definition of Done

A story is DONE when:

**Code:**
- [ ] Code written and follows style guide
- [ ] Unit tests written (80% coverage)
- [ ] Integration tests passing
- [ ] No lint errors
- [ ] Code reviewed and approved

**Documentation:**
- [ ] API documented (if applicable)
- [ ] README updated (if applicable)
- [ ] Changelog entry added

**Quality:**
- [ ] Tested on staging environment
- [ ] No critical bugs
- [ ] Performance acceptable
- [ ] Security review passed (if applicable)

**Deployment:**
- [ ] Merged to main branch
- [ ] Deployed to staging
- [ ] Smoke tests passing
```

---

## 7. Questions to Ask

### Requirements Clarification

```markdown
**User Context:**
- Who are the target users?
- What problem are we solving?
- What's the expected user journey?

**Functional:**
- What are the input validations?
- What are the success/error scenarios?
- Are there any business rules?
- What are the edge cases?

**Non-Functional:**
- What's the expected response time?
- How many concurrent users?
- Any security requirements?
- Accessibility requirements?

**Scope:**
- What's MVP vs nice-to-have?
- Any related features to consider?
- What's explicitly out of scope?
```

---

## Quick Reference

| Element | Template |
|---------|----------|
| User Story | As a [role], I want [goal], so that [benefit] |
| AC Format | Given [context], When [action], Then [result] |
| API Spec | Method, Endpoint, Request, Response, Errors |
| DoD | Code + Tests + Docs + QA + Deploy |

---

## Related Knowledge

- [02-api-contracts.md](./02-api-contracts.md) - API specification details
- [../shared/01-go-fundamentals.md](../shared/01-go-fundamentals.md) - Go patterns
