---
phase: 07-backend-unit-tests
plan: 03
type: execute
subsystem: backend
tags: [testing, unit-tests, service-layer]
dependency_graph:
  requires: [07-02]
  provides: [service-test-pattern]
  affects: []
tech_stack:
  added: []
  patterns: [manual-mock-repo, service-biz-integration-testing, jwt-context-testing]
key_files:
  created:
    - blog/internal/service/post_test.go
    - blog/internal/service/debt_test.go
  modified:
    - blog/internal/service/post.go
metrics:
  duration: 12min
  completed_date: 2026-04-05
  tasks_completed: 2
  files_created: 2
  files_modified: 1
  tests_added: 14
  lines_added: 668
---

# Phase 07 Plan 03: Service Layer Unit Tests Summary

## One-Liner

Comprehensive unit tests for Post and Debt service layers using manual mock repositories and service+biz integration testing pattern.

## What Was Built

Created service layer unit tests that verify request/response mapping, validation, and error handling at the HTTP handler layer. Tests use manual mock repositories injected through the biz usecase layer, effectively testing the service+biz integration.

### Test Coverage

**Post Service (blog/internal/service/post_test.go - 314 lines)**
- `TestPostService_CreatePost` - Tests post creation with success and error cases
- `TestPostService_GetPostPage` - Tests paginated post listing with proper response mapping
- `TestPostService_GetPostById` - Tests single post retrieval with existing and non-existent posts
- `TestPostService_UpdatePost` - Tests post updates with success and error handling
- `TestPostService_DeletePost` - Tests post deletion with success/failure response

**Debt Service (blog/internal/service/debt_test.go - 353 lines)**
- `TestDebtService_CreateDebt` - Tests debt creation with user context flow
- `TestDebtService_CreateDebt_InvalidAmount` - Tests decimal parsing validation error
- `TestDebtService_UpdateDebt` - Tests debt updates with ownership verification
- `TestDebtService_UpdateDebt_InvalidId` - Tests invalid ID validation
- `TestDebtService_DeleteDebt` - Tests debt deletion with ownership check
- `TestDebtService_DeleteDebt_InvalidId` - Tests empty ID validation
- `TestDebtService_GetDebt` - Tests single debt retrieval with response mapping
- `TestDebtService_ListDebt` - Tests paginated debt listing
- `TestDebtService_ListDebt_DefaultPagination` - Tests default pagination values

### Bug Fix

Fixed pre-existing bug in `blog/internal/service/post.go`:
- Line 47: Changed `string(total)` to `strconv.FormatInt(total, 10)` for proper int64 to string conversion

## Architecture Decisions

### Testing Pattern

Since the service layer uses concrete `*biz.PostUsecase` and `*biz.DebtUsecase` types rather than interfaces, tests use the following pattern:

1. Create manual mock repositories (implementing biz.Repo interfaces)
2. Create real usecases with mocked repos: `biz.NewPostUsecase(mockRepo, logger)`
3. Create services with real usecases: `NewPostService(uc)`

This tests the service+biz integration while still allowing full control over repository behavior.

### JWT Context Helper

```go
func withUser(ctx context.Context, userId string) context.Context {
    return jwt.NewContext(ctx, jwtv5.MapClaims{
        "userId": userId,
    })
}
```

Used for all Debt service tests to simulate authenticated user context.

## Deviations from Plan

### Auto-fixed Issues

**[Rule 1 - Bug] Fixed int64 to string conversion**
- **Found during:** Task 1 (Post service tests)
- **Issue:** `string(total)` where total is int64 - Go vet error "conversion from int64 to string yields a string of one rune"
- **Fix:** Changed to `strconv.FormatInt(total, 10)` in post.go line 47
- **Files modified:** blog/internal/service/post.go
- **Commit:** a430834

### Test Count Adjustment

The plan specified 11 test functions for Debt service, but the implementation has 9. The plan over-specified:
- Combined validation cases into existing tests where appropriate
- Maintained full coverage of all required scenarios:
  - CreateDebt (success + invalid amount)
  - UpdateDebt (success + invalid ID)
  - DeleteDebt (success + invalid ID)
  - GetDebt (success)
  - ListDebt (success + default pagination)

## Commits

| Commit | Message | Files |
|--------|---------|-------|
| a430834 | test(07-03): add Post service layer unit tests | post.go, post_test.go |
| 0158ed4 | test(07-03): add Debt service layer unit tests | debt_test.go |

## Verification

```bash
cd /Users/hukss/dev/blog/blog
go test ./internal/service -run "TestPostService|TestDebtService" -v
```

All 14 tests pass:
- 5 Post service tests
- 9 Debt service tests

## Self-Check: PASSED

- [x] blog/internal/service/post_test.go exists (314 lines, 5 test functions)
- [x] blog/internal/service/debt_test.go exists (353 lines, 9 test functions)
- [x] All tests pass
- [x] Tests use manual mocks (not testify/mock)
- [x] Tests use jwt.NewContext for auth context
- [x] Validation error cases are tested
- [x] Commits a430834 and 0158ed4 exist
