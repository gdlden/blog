# Phase 7: Backend Unit Tests - Context

**Gathered:** 2026-04-05
**Status:** Ready for planning

<domain>
## Phase Boundary

Create unit tests for critical backend services (blog/post and debt) to fulfill QUAL-01 requirement: "Critical blog and debt flows have backend tests or verification coverage strong enough to catch common regressions."

Scope includes:
- Post service (all CRUD operations)
- Debt service (all CRUD operations + business constraints)
- All three layers: service, biz (use cases), and data

Out of scope:
- Frontend tests (covered by QUAL-02)
- Integration tests with real database
- Performance/load tests

</domain>

<decisions>
## Implementation Decisions

### Test Scope (Layers)
- **D-01:** Test all three layers: biz (use cases), data (repositories), and service (handlers)
- **D-02:** Follow existing pattern in `blog/internal/data/post_test.go` for data layer tests

### Debt Test Coverage
- **D-03:** Full CRUD coverage for Debt (Create, Get, List, Update, Delete)
- **D-04:** Test business constraints:
  - Ownership check: users can only access their own debts
  - Delete protection: cannot delete debt that has associated details
- **D-05:** Test user context extraction from requests

### Auth Handling
- **D-06:** Mock context with user ID using the same mechanism as real auth middleware
- **D-07:** Context key pattern: use the same key that `utils.CurrentUserId()` expects

### Coverage Target
- **D-08:** Happy path tests for all operations
- **D-09:** Key error cases: invalid input, not found, permission denied
- **D-10:** Does NOT need: exhaustive edge cases, performance tests, concurrent access tests

### Test Infrastructure
- **D-11:** Use SQLite in-memory database for data layer tests (fast, isolated)
- **D-12:** Use `testify/assert` for assertions (existing pattern)
- **D-13:** Each test creates fresh database instance to avoid state leakage

### File Organization
- **D-14:** Place tests adjacent to implementation: `*_test.go` in same package
- **D-15:** Name pattern: `Test<Service>_<Operation>` (e.g., `TestDebtUsecase_CreateDebt`)

### Claude's Discretion
- Exact assertion style (table-driven vs individual tests)
- Test helper function design (repeated setup code)
- Order of test implementation (which service first)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Existing Test Pattern
- `blog/internal/data/post_test.go` — Reference implementation for data layer tests using SQLite

### Implementation Files to Test
- `blog/internal/biz/post.go` — Post use cases
- `blog/internal/biz/debt.go` — Debt use cases
- `blog/internal/data/post.go` — Post repository (partial tests exist)
- `blog/internal/data/debt.go` — Debt repository
- `blog/internal/service/post.go` — Post service handlers
- `blog/internal/service/debt.go` — Debt service handlers

### Auth Utilities
- `blog/internal/utils/userutil.go` — User context extraction (must understand context key)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `setupTestDB(t *testing.T) *gorm.DB` pattern from `post_test.go`
- `testify/assert` already imported and used
- SQLite driver already configured for tests

### Established Patterns
- Data layer: Direct repo struct initialization with injected test DB
- Biz layer: Use case struct with mocked repo interface
- Service layer: Likely needs context with user ID for auth

### Integration Points
- Tests must align with GORM model definitions in `blog/internal/model/`
- Debt tests need DebtDetail model for "delete with details" constraint
- Context key for user ID must match `utils.CurrentUserId()` expectation

</code_context>

<specifics>
## Specific Ideas

### Critical Paths to Cover

**Post (already has data tests, add biz + service):**
1. CreatePost — success
2. GetPostPage — pagination works
3. GetPostById — found and not found
4. UpdatePost — success
5. DeletePost — success

**Debt (no existing tests):**
1. CreateDebt — success, extracts user from context
2. GetDebt — success, ownership check
3. ListDebt — pagination, filtering
4. Edit — success, ownership verification
5. Delete — success, ownership, blocked when details exist

### Key Error Cases
- Invalid ID format (parse errors)
- Record not found
- Delete debt with existing details (business rule violation)
- Access other user's debt (ownership violation)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 07-backend-unit-tests*
*Context gathered: 2026-04-05*
