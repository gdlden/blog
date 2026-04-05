# Phase 7: Backend Unit Tests - Research

**Researched:** 2026-04-05
**Domain:** Go unit testing with Kratos/GORM layered architecture
**Confidence:** HIGH

## Summary

This phase implements unit tests for critical backend services (blog/post and debt) to fulfill QUAL-01 requirement. The codebase uses a clean architecture pattern with three layers: service (transport/handlers), biz (use cases), and data (repositories).

The existing test infrastructure in `blog/internal/data/post_test.go` provides a proven pattern using SQLite in-memory databases for data layer tests. This research confirms the approach and extends it to cover all three layers for both Post and Debt domains.

**Primary recommendation:** Use SQLite in-memory for data layer tests, manual mock structs for biz layer tests, and JWT context injection for service layer tests. Follow table-driven patterns for multiple similar test cases.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Test all three layers: biz (use cases), data (repositories), and service (handlers)
- **D-02:** Follow existing pattern in `blog/internal/data/post_test.go` for data layer tests
- **D-03:** Full CRUD coverage for Debt (Create, Get, List, Update, Delete)
- **D-04:** Test business constraints:
  - Ownership check: users can only access their own debts
  - Delete protection: cannot delete debt that has associated details
- **D-05:** Test user context extraction from requests
- **D-06:** Mock context with user ID using the same mechanism as real auth middleware
- **D-07:** Context key pattern: use the same key that `utils.CurrentUserId()` expects
- **D-08:** Happy path tests for all operations
- **D-09:** Key error cases: invalid input, not found, permission denied
- **D-10:** Does NOT need: exhaustive edge cases, performance tests, concurrent access tests
- **D-11:** Use SQLite in-memory database for data layer tests (fast, isolated)
- **D-12:** Use `testify/assert` for assertions (existing pattern)
- **D-13:** Each test creates fresh database instance to avoid state leakage
- **D-14:** Place tests adjacent to implementation: `*_test.go` in same package
- **D-15:** Name pattern: `Test<Service>_<Operation>` (e.g., `TestDebtUsecase_CreateDebt`)

### Claude's Discretion
- Exact assertion style (table-driven vs individual tests)
- Test helper function design (repeated setup code)
- Order of test implementation (which service first)

### Deferred Ideas (OUT OF SCOPE)
- None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| QUAL-01 | Critical blog and debt flows have backend tests or verification coverage strong enough to catch common regressions | All findings in this document enable implementation |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| testing | Go stdlib | Test framework | Native Go testing, no dependencies |
| testify/assert | v1.9.0 | Assertion helpers | Clean syntax, widely adopted, already in project |
| gorm.io/driver/sqlite | v1.6.0 | In-memory test DB | Fast, isolated, no external services |
| gorm.io/gorm | v1.25.12 | ORM | Already used in production code |
| jwt (kratos) | v2.8.0 | Context injection | `jwt.NewContext()` for auth mocking |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/golang-jwt/jwt/v5 | v5.1.0 | JWT claims | Creating mock tokens in tests |
| github.com/shopspring/decimal | v1.4.0 | Decimal arithmetic | Debt amount fields, already in use |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| testify/mock | Manual mock structs | testify/mock adds complexity; manual mocks are clearer for simple interfaces |
| PostgreSQL test container | SQLite in-memory | SQLite is faster and sufficient for unit tests; PG container for integration tests |
| gomock | testify/mock | gomock has code generation; manual mocks preferred for small interfaces |

**Installation:** Already installed via go.mod

## Architecture Patterns

### Recommended Test Organization
```
blog/internal/
├── biz/
│   ├── post.go              # Post use cases
│   ├── post_test.go         # Post biz layer tests (NEW)
│   ├── debt.go              # Debt use cases
│   └── debt_test.go         # Debt biz layer tests (NEW)
├── data/
│   ├── post.go              # Post repository
│   ├── post_test.go         # Post data tests (EXISTING - reference)
│   ├── debt.go              # Debt repository
│   └── debt_test.go         # Debt data tests (NEW)
└── service/
    ├── post.go              # Post service handlers
    ├── post_test.go         # Post service tests (NEW)
    ├── debt.go              # Debt service handlers
    └── debt_test.go         # Debt service tests (NEW)
```

### Pattern 1: Data Layer Test (SQLite In-Memory)
**What:** Direct repository testing with real database operations using SQLite in-memory
**When to use:** Data layer tests where SQL logic needs verification
**Example:**
```go
// Source: blog/internal/data/post_test.go (existing pattern)
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open test database: %v", err)
    }
    err = db.AutoMigrate(&Debt{}, &DebtDetail{}) // Migrate all needed models
    if err != nil {
        t.Fatalf("failed to migrate test database: %v", err)
    }
    return db
}

func TestDebtRepo_Save(t *testing.T) {
    db := setupTestDB(t)
    repo := &DebtRepo{
        data: &Data{db: db},
        log:  log.NewHelper(log.DefaultLogger),
    }

    ctx := context.Background()
    debt := &biz.Debt{
        Name:      "Test Debt",
        BankName:  "Test Bank",
        Amount:    decimal.NewFromInt(1000),
        ApplyTime: "2024-01-01 00:00:00",
        EndTime:   "2024-12-31 00:00:00",
        UserId:    "user-123",
    }

    id, err := repo.Save(ctx, debt)

    assert.NoError(t, err)
    assert.NotZero(t, id)
}
```

### Pattern 2: Biz Layer Test (Manual Mock)
**What:** Use case testing with manually created mock repositories
**When to use:** Testing business logic independent of database
**Example:**
```go
// Manual mock implementing biz.DebtRepo interface
type mockDebtRepo struct {
    saveFunc    func(context.Context, *biz.Debt) (uint, error)
    findByIdFunc func(context.Context, string, uint) (*biz.Debt, error)
    // ... other methods
}

func (m *mockDebtRepo) Save(ctx context.Context, d *biz.Debt) (uint, error) {
    return m.saveFunc(ctx, d)
}
// ... implement all interface methods

func TestDebtUsecase_CreateDebt(t *testing.T) {
    repo := &mockDebtRepo{
        saveFunc: func(ctx context.Context, d *biz.Debt) (uint, error) {
            return 1, nil
        },
    }
    uc := biz.NewDebtUsecase(repo, log.DefaultLogger)
    
    // Create context with user ID
    ctx := jwt.NewContext(context.Background(), jwtv5.MapClaims{"userId": "user-123"})
    
    debt := &biz.Debt{Name: "Test"}
    id, err := uc.CreateDebt(ctx, debt)
    
    assert.NoError(t, err)
    assert.Equal(t, uint(1), id)
    assert.Equal(t, "user-123", debt.UserId) // Verify user ID was set
}
```

### Pattern 3: Service Layer Test (Protobuf + Context)
**What:** Handler testing with protobuf requests and JWT context
**When to use:** Testing request/response mapping and validation
**Example:**
```go
func TestDebtService_CreateDebt(t *testing.T) {
    // Setup mock use case
    mockUC := &mockDebtUsecase{}
    svc := service.NewDebtService(mockUC)
    
    // Create authenticated context
    ctx := jwt.NewContext(context.Background(), jwtv5.MapClaims{"userId": "user-123"})
    
    req := &pb.DebtEntity{
        Name:   "Test Debt",
        Amount: "1000.00",
        // ... other fields
    }
    
    resp, err := svc.CreateDebt(ctx, req)
    
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, "save success", resp.Message)
}
```

### Pattern 4: Table-Driven Tests
**What:** Single test function covering multiple scenarios with a table of inputs/expected outputs
**When to use:** Multiple similar test cases with different inputs (validation, error cases)
**Example:**
```go
func TestDebtService_CreateDebt_Validation(t *testing.T) {
    tests := []struct {
        name      string
        req       *pb.DebtEntity
        wantError bool
        errorMsg  string
    }{
        {
            name: "valid debt",
            req:  &pb.DebtEntity{Name: "Test", Amount: "1000"},
            wantError: false,
        },
        {
            name: "invalid amount format",
            req:  &pb.DebtEntity{Name: "Test", Amount: "invalid"},
            wantError: true,
            errorMsg:  "invalid decimal",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := service.NewDebtService(&mockDebtUsecase{})
            ctx := jwt.NewContext(context.Background(), jwtv5.MapClaims{"userId": "user-1"})
            
            _, err := svc.CreateDebt(ctx, tt.req)
            
            if tt.wantError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Anti-Patterns to Avoid
- **Global test database:** Each test must create its own isolated database instance
- **Test order dependencies:** Tests must pass regardless of execution order
- **Mocking external libraries:** Mock interfaces you own, not third-party libraries
- **Testing implementation details:** Test behavior, not internal structure
- **Skipping error assertions:** Always assert error conditions explicitly

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Test database | PostgreSQL container | SQLite in-memory | 100x faster, no Docker needed |
| Mock framework | testify/mock or gomock | Manual mock structs | Simpler for small interfaces, clearer intent |
| Assertion library | Custom assert helpers | testify/assert | Standard, well-documented, comprehensive |
| Context with auth | Custom context keys | `jwt.NewContext()` | Matches production middleware exactly |
| Decimal parsing | Float arithmetic | shopspring/decimal | Exact precision for financial data |

**Key insight:** The existing `post_test.go` already demonstrates the right patterns. Don't introduce new testing frameworks or complex mock generators when simple manual mocks and testify/assert suffice.

## Common Pitfalls

### Pitfall 1: Context Key Mismatch
**What goes wrong:** Tests create context with wrong key type, `utils.CurrentUserId()` fails to extract user ID
**Why it happens:** `jwt.NewContext()` uses specific key types that must match what `jwt.FromContext()` expects
**How to avoid:** Always use `jwt.NewContext(ctx, jwtv5.MapClaims{"userId": "..."})` exactly as shown
**Warning signs:** Tests fail with "未登录" error even when context appears set

### Pitfall 2: Database State Leakage
**What goes wrong:** Tests share database state, causing flaky tests that pass/fail based on order
**Why it happens:** Using global test DB or not cleaning up between tests
**How to avoid:** Each test creates fresh DB via `setupTestDB(t)`; defer cleanup
**Warning signs:** Tests pass individually but fail in suite; intermittent failures

### Pitfall 3: Missing Model Migration
**What goes wrong:** SQLite tests fail with "no such table" errors
**Why it happens:** `AutoMigrate` only migrates specified models; related models (e.g., DebtDetail) not included
**How to avoid:** Include all related models in migration: `db.AutoMigrate(&Debt{}, &DebtDetail{})`
**Warning signs:** SQL errors about missing tables in test output

### Pitfall 4: Decimal String Parsing
**What goes wrong:** Tests pass decimal as float, lose precision
**Why it happens:** `decimal.Decimal` requires string constructor for exact precision
**How to avoid:** Use `decimal.NewFromString("1000.50")` not `decimal.NewFromFloat(1000.50)`
**Warning signs:** Assertion failures on amount comparisons; off-by-cent errors

### Pitfall 5: Time Format Mismatches
**What goes wrong:** Debt time fields don't parse correctly
**Why it happens:** Repository expects `"2006-01-02 15:04:05"` format
**How to avoid:** Always use exact format string in test data
**Warning signs:** "parsing time" errors in test output

### Pitfall 6: Repository Interface Drift
**What goes wrong:** Mock doesn't implement updated interface, compile fails
**Why it happens:** Interface changed, mocks not updated
**How to avoid:** Run `go build ./...` after interface changes; consider compile-time interface checks
**Warning signs:** Compilation errors in test files after biz layer changes

## Code Examples

### Creating Authenticated Context
```go
import (
    "context"
    "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
    jwtv5 "github.com/golang-jwt/jwt/v5"
)

func withUser(ctx context.Context, userId string) context.Context {
    return jwt.NewContext(ctx, jwtv5.MapClaims{
        "userId": userId,
    })
}

// Usage in test:
ctx := withUser(context.Background(), "user-123")
```

### Complete Data Layer Test (Debt)
```go
func TestDebtRepo_FindByUserIdAndDebtId(t *testing.T) {
    db := setupTestDB(t)
    repo := &DebtRepo{
        data: &Data{db: db},
        log:  log.NewHelper(log.DefaultLogger),
    }
    ctx := context.Background()
    
    // Create test debt
    debt := &biz.Debt{
        Name:      "Test",
        BankName:  "Bank",
        Amount:    decimal.NewFromInt(1000),
        ApplyTime: "2024-01-01 00:00:00",
        EndTime:   "2024-12-31 00:00:00",
        UserId:    "user-123",
    }
    id, _ := repo.Save(ctx, debt)
    
    // Test ownership check
    found, err := repo.FindByUserIdAndDebtId(ctx, "user-123", id)
    assert.NoError(t, err)
    assert.NotNil(t, found)
    
    // Test wrong user
    notFound, err := repo.FindByUserIdAndDebtId(ctx, "user-456", id)
    assert.Error(t, err)
    assert.Nil(t, notFound)
}
```

### Complete Biz Layer Test (Ownership Check)
```go
func TestDebtUsecase_Delete_WithDetails(t *testing.T) {
    repo := &mockDebtRepo{
        findByIdFunc: func(ctx context.Context, uid string, id uint) (*biz.Debt, error) {
            return &biz.Debt{Id: int64(id), UserId: uid}, nil
        },
        countDetailFunc: func(ctx context.Context, id uint) (int64, error) {
            return 2, nil // Has 2 details
        },
    }
    
    uc := biz.NewDebtUsecase(repo, log.DefaultLogger)
    ctx := jwt.NewContext(context.Background(), jwtv5.MapClaims{"userId": "user-1"})
    
    err := uc.Delete(ctx, 1)
    
    assert.Error(t, err)
    assert.Equal(t, "存在明细，禁止删除", err.Error())
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| External test DB | SQLite in-memory | Already adopted | Faster, isolated tests |
| Custom assertions | testify/assert | Already adopted | Standard, maintainable |
| Mocking frameworks | Manual mocks | Recommended | Simpler for Go interfaces |

**Deprecated/outdated:**
- None identified for this phase

## Open Questions

1. **Test Coverage Target**
   - What we know: QUAL-01 requires "coverage strong enough to catch common regressions"
   - What's unclear: Exact percentage or coverage tool to use
   - Recommendation: Focus on critical paths (CRUD + business rules), measure with `go test -cover`

2. **CI Integration**
   - What we know: Tests must run in CI pipeline
   - What's unclear: Whether CI has SQLite support
   - Recommendation: SQLite is embedded, no external dependency needed

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | All tests | ✓ | 1.24.0 | — |
| SQLite driver | Data tests | ✓ | v1.6.0 | — |
| testify/assert | Assertions | ✓ | v1.9.0 | — |
| GORM | Data tests | ✓ | v1.25.12 | — |

**Missing dependencies with no fallback:**
- None

**Missing dependencies with fallback:**
- None

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify/assert v1.9.0 |
| Config file | None — standard Go testing |
| Quick run command | `go test ./internal/biz/... ./internal/data/... ./internal/service/... -v` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QUAL-01 | Post CRUD operations work | unit | `go test ./internal/data -run TestPostRepo -v` | ✅ |
| QUAL-01 | Debt data layer CRUD | unit | `go test ./internal/data -run TestDebtRepo -v` | ❌ Wave 0 |
| QUAL-01 | Debt biz layer use cases | unit | `go test ./internal/biz -run TestDebtUsecase -v` | ❌ Wave 0 |
| QUAL-01 | Debt service handlers | unit | `go test ./internal/service -run TestDebtService -v` | ❌ Wave 0 |
| QUAL-01 | Post biz layer use cases | unit | `go test ./internal/biz -run TestPostUsecase -v` | ❌ Wave 0 |
| QUAL-01 | Post service handlers | unit | `go test ./internal/service -run TestPostService -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/{package} -v -run Test{Specific}`
- **Per wave merge:** `go test ./internal/... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `blog/internal/data/debt_test.go` — Debt data layer tests
- [ ] `blog/internal/biz/debt_test.go` — Debt biz layer tests
- [ ] `blog/internal/biz/post_test.go` — Post biz layer tests
- [ ] `blog/internal/service/debt_test.go` — Debt service tests
- [ ] `blog/internal/service/post_test.go` — Post service tests

## Sources

### Primary (HIGH confidence)
- `blog/internal/data/post_test.go` - Existing test pattern (verified working)
- `blog/internal/utils/userutil.go` - Context extraction mechanism
- `blog/go.mod` - Dependency versions
- Go documentation: `go doc github.com/go-kratos/kratos/v2/middleware/auth/jwt` - Context injection API

### Secondary (MEDIUM confidence)
- Go Testing Patterns (training knowledge) - Table-driven tests, subtests
- Kratos examples (training knowledge) - Service testing patterns

### Tertiary (LOW confidence)
- None required — all patterns verified from codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All dependencies verified in go.mod and working
- Architecture: HIGH - Existing post_test.go proves pattern works
- Pitfalls: MEDIUM - Based on codebase analysis and common Go patterns

**Research date:** 2026-04-05
**Valid until:** 30 days (stable stack)
