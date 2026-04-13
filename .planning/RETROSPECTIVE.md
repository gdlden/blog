# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 — Blog Debt Hub MVP

**Shipped:** 2026-04-13
**Phases:** 4 executed (01, 05, 06, 07) | **Plans:** 11 | **Tasks:** 11
**Timeline:** 8 days (2026-04-05 → 2026-04-13)
**Git activity:** 50 commits, 64 files changed, +10,805/-214 lines

### What Was Built

- Reactive authentication foundation with Pinia as the live source of truth, aligned Axios interceptor, and proper Vue Router guards.
- Unified authenticated app shell with sidebar navigation connecting Blog and Debt functional areas.
- Environment-based database configuration enabling flexible deployment via env var overrides.
- Full frontend-backend integration for blog and debt modules: list, view, create, edit, delete, plus toast notifications and debt summary statistics.
- Comprehensive backend unit tests across data, biz, and service layers using SQLite in-memory and manual mock repositories.

### What Worked

- Executing phases in yolo mode kept overhead low and allowed rapid iteration.
- Reusing the existing Kratos + Vue stack rather than rewriting saved significant time.
- Manual mock pattern for Go unit tests proved simpler and clearer than pulling in testify/mock.
- SQLite in-memory tests provided fast, isolated execution without container overhead.

### What Was Inefficient

- Phases 2, 3, and 4 were planned in the original roadmap but never executed as separate phases; their requirements were instead absorbed into Phases 5-7. This created a mismatch between roadmap numbering and actual execution history.
- The requirements traceability table in REQUIREMENTS.md became stale because debt and blog features were delivered under Phase 6 rather than the originally mapped Phase 2 and Phase 3.

### Patterns Established

- **Backend layering**: Strict Kratos pattern (server → service → biz → data) with Wire dependency injection.
- **Test pattern**: Manual mock repos with function fields, JWT context mocking via `jwt.NewContext`, and SQLite in-memory for data layer tests.
- **Frontend state**: Pinia stores mirror localStorage for auth, with Axios interceptors reading from the reactive store.
- **Commit granularity**: One commit per plan task, atomic and descriptive.

### Key Lessons

1. **Roadmap phases should be executable units**. Planning phases that never get executed as separate units creates bookkeeping debt at milestone time.
2. **Requirements traceability must be updated as scope shifts**. When features move between phases, the requirements file needs to be kept in sync to avoid confusion during milestone audits.
3. **Brownfield evolution works when constraints are respected**. Building on the existing stack and accepting current architecture allowed the team to ship functional CRUD quickly.
4. **Backend tests on SQLite catch real GORM issues**. Using a real (in-memory) database for repository tests validated actual query behavior, not just interface contracts.

### Cost Observations

- Model mix: primarily sonnet for execution and verification, opus for planning
- Sessions: compact, multi-phase sessions with good context reuse
- Notable: Low token overhead due to yolo mode and inline execution for smaller phases

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.0 | ~5 | 4 executed | Established yolo-mode execution and manual mock testing patterns |

### Cumulative Quality

| Milestone | Backend Tests | Frontend Tests | Coverage Notes |
|-----------|---------------|----------------|----------------|
| v1.0 | 55+ (data/biz/service) | 0 | Backend covered; frontend tests are the v1.1 priority |

### Top Lessons (Verified Across Milestones)

1. Test against real database behavior where possible — mocks alone miss GORM and SQL edge cases.
2. Keep roadmaps aligned with actual execution rhythm; avoid placeholder phases that get skipped.

---
