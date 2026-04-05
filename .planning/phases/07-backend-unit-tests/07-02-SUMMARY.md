---
phase: 07-backend-unit-tests
plan: 02
type: summary
subsystem: backend
tags: [testing, unit-tests, biz-layer, mocks]
dependency_graph:
  requires:
    - 07-01
  provides:
    - Post biz layer unit tests
    - Debt biz layer unit tests
  affects: []
tech_stack:
  added: []
  patterns:
    - Manual mock repositories (function field pattern)
    - JWT context mocking with jwt.NewContext
    - Table-driven tests with subtests
key_files:
  created:
    - blog/internal/biz/post_test.go
    - blog/internal/biz/debt_test.go
  modified: []
decisions:
  - Used manual mock pattern with function fields instead of testify/mock for simplicity and clarity
  - Created withUser helper for consistent JWT context mocking across tests
  - Verified UserId is set correctly by inspecting mock function arguments
  - Tested both success and error paths for all use case methods
metrics:
  duration: 5m
  completed_date: "2026-04-05"
---

# Phase 07 Plan 02: Post and Debt Biz Layer Unit Tests

## Summary

Comprehensive unit tests for Post and Debt business layer use cases using manual mock repositories. Tests cover CRUD operations, user context extraction, ownership checks, and delete protection logic.

## What Was Built

### Post Biz Layer Tests (`blog/internal/biz/post_test.go`)

**329 lines, 5 test functions**

| Test Function | Coverage |
|--------------|----------|
| `TestPostUsecase_CreatePost` | Create post success and error cases |
| `TestPostUsecase_GetPostPage` | Pagination with total count |
| `TestPostUsecase_GetPostById` | Existing post and not-found cases |
| `TestPostUsecase_UpdatePost` | Update success and error cases |
| `TestPostUsecase_DeletePost` | Delete success and error cases |

### Debt Biz Layer Tests (`blog/internal/biz/debt_test.go`)

**401 lines, 10 test functions**

| Test Function | Coverage |
|--------------|----------|
| `TestDebtUsecase_CreateDebt` | Create debt success and error cases |
| `TestDebtUsecase_CreateDebt_Unauthorized` | Auth check (no user in context) |
| `TestDebtUsecase_Edit` | Edit success with ownership verification |
| `TestDebtUsecase_Edit_NotOwner` | Ownership check failure |
| `TestDebtUsecase_Delete` | Delete success with detail count check |
| `TestDebtUsecase_Delete_WithDetails` | Delete blocked when details exist |
| `TestDebtUsecase_Delete_NotOwner` | Ownership check failure |
| `TestDebtUsecase_GetDebt` | Get debt success and error cases |
| `TestDebtUsecase_ListDebt` | List with pagination |

## Key Patterns Used

### Manual Mock Pattern
```go
type mockPostRepo struct {
    saveFunc       func(context.Context, *Post) (*Post, error)
    updateFunc     func(context.Context, *Post) (*Post, error)
    // ... function fields for each interface method
}
```

### JWT Context Mocking
```go
func withUser(ctx context.Context, userId string) context.Context {
    return jwt.NewContext(ctx, jwtv5.MapClaims{
        "userId": userId,
    })
}
```

### UserId Verification in Mocks
```go
mockFn: func(ctx context.Context, d *Debt) (uint, error) {
    // Verify UserId was set by the usecase
    if d.UserId != "user-123" {
        t.Errorf("expected UserId to be 'user-123', got %s", d.UserId)
    }
    return 1, nil
},
```

## Truths Validated

- [x] Post use cases work correctly with mocked repositories
- [x] Debt use cases work correctly with mocked repositories
- [x] User context extraction works in business layer
- [x] Ownership checks prevent unauthorized operations
- [x] Delete blocked when debt has details

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- [x] File `blog/internal/biz/post_test.go` exists (329 lines)
- [x] File `blog/internal/biz/debt_test.go` exists (401 lines)
- [x] File contains `type mockPostRepo struct`
- [x] File contains `type mockDebtRepo struct`
- [x] File contains `func withUser(ctx context.Context, userId string) context.Context`
- [x] All 5 Post test functions present
- [x] All 10 Debt test functions present
- [x] All tests pass (`go test ./internal/biz -run "TestPostUsecase|TestDebtUsecase" -v`)
- [x] Tests use manual mocks (not testify/mock)
- [x] Tests use jwt.NewContext for auth context

## Commits

| Hash | Message |
|------|---------|
| `36971ef` | test(07-02): add Post biz layer unit tests with manual mocks |
| `96a5f93` | test(07-02): add Debt biz layer unit tests with manual mocks |
