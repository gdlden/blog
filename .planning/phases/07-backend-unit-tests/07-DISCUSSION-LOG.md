# Phase 7: Backend Unit Tests - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-05
**Phase:** 07-backend-unit-tests
**Areas discussed:** Test scope, Debt test complexity, User context mocking, Coverage target

---

## Test Scope — Which layers to cover?

| Option | Description | Selected |
|--------|-------------|----------|
| Biz layer only | Test use cases. Skip service layer. | |
| Biz + Data layers | Test use cases and complete data layer | |
| Biz + Data + Service layers | Full coverage including service handlers | ✓ |
| You decide | Claude recommends | |

**User's choice:** Full coverage (Biz + Data + Service layers)
**Notes:** User wants comprehensive testing despite service layer being thin mapping.

---

## Debt Test Complexity — Which flows?

| Option | Description | Selected |
|--------|-------------|----------|
| Full CRUD + all constraints | All operations plus ownership and delete checks | ✓ |
| Happy path + critical constraints | Core flows + key business rules | |
| Critical paths only | Create, List, Delete only | |
| You decide | Claude recommends | |

**User's choice:** Full CRUD + all constraints
**Notes:** Include ownership verification and "cannot delete with details" constraint.

---

## User Context Mocking — How to handle auth?

| Option | Description | Selected |
|--------|-------------|----------|
| Mock context with user ID | Create context with embedded user ID | ✓ |
| Mock utils.CurrentUserId function | Replace function during tests | |
| Test at repo layer only | Skip auth-dependent use cases | |
| You decide | Claude recommends | |

**User's choice:** Mock context with user ID
**Notes:** Matches real auth middleware pattern.

---

## Coverage Target — What's "enough"?

| Option | Description | Selected |
|--------|-------------|----------|
| Happy path + key errors | Success cases + most likely errors | ✓ |
| Comprehensive | All code paths including edge cases | |
| Minimal viable | Critical paths only | |
| You decide | Claude recommends | |

**User's choice:** Happy path + key errors
**Notes:** Sufficient for QUAL-01 requirement.

---

## Claude's Discretion

**Areas where user deferred to Claude:**
- Exact assertion style (table-driven vs individual tests)
- Test helper function design
- Order of test implementation

## Deferred Ideas

None — discussion stayed within phase scope.
