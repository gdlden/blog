# Blog Debt Hub

## What This Is

Blog Debt Hub is a personal full-stack site that combines a blog area and a debt-management area under one account system. It is intended for a single user to write and manage blog content at any time, while also tracking personal debt records and reviewing debt statistics in the same place.

The v1.0 MVP is shipped and functional: users can authenticate reliably, navigate between blog and debt areas, perform full CRUD on both blog posts and debt records, and benefit from backend unit-test coverage for safe iteration.

## Core Value

At any time, the user can reliably record and manage their own blog content, then use the same site to review and manage personal debt information.

## Requirements

### Validated (v1.0)

- [x] **FND-01**: User can log in and remain authenticated across normal page navigation without stale route-guard state — v1.0 (Phase 1)
- [x] **FND-02**: User can log out or otherwise clear session state cleanly and return to the login page — v1.0 (Phase 1)
- [x] **FND-03**: User can navigate between the blog area and debt area within one authenticated site shell — v1.0 (Phase 1)
- [x] **BLOG-01**: User can view a list of their blog posts in the frontend — v1.0 (Phase 6)
- [x] **BLOG-02**: User can view the full content of a selected blog post — v1.0 (Phase 6)
- [x] **BLOG-03**: User can create a blog post with title and content — v1.0 (Phase 6)
- [x] **BLOG-04**: User can edit an existing blog post they own — v1.0 (Phase 6)
- [x] **BLOG-05**: User can delete an existing blog post they own — v1.0 (Phase 6)
- [x] **BLOG-06**: User can manage blog posts from an authenticated interface rather than raw API-only access — v1.0 (Phase 6)
- [x] **DEBT-01**: User can create a debt record with the current core debt fields — v1.0 (Phase 6)
- [x] **DEBT-02**: User can view a paginated list of debt records — v1.0 (Phase 6)
- [x] **DEBT-03**: User can view the details of an individual debt record — v1.0 (Phase 6)
- [x] **DEBT-04**: User can update an existing debt record — v1.0 (Phase 6)
- [x] **DEBT-05**: User can delete an existing debt record — v1.0 (Phase 6)
- [x] **DEBT-06**: User can view debt summaries including total debt, repaid amount, outstanding amount, and per-record breakdown — v1.0 (Phase 6)
- [x] **DEBT-07**: User can view and manage debt-detail records needed to support debt history and summary accuracy — v1.0 (Phase 6)
- [x] **QUAL-01**: Critical blog and debt flows have backend tests or verification coverage strong enough to catch common regressions — v1.0 (Phase 7)
- [x] **CFG-01**: Database source can be configured via environment variable for flexible deployment — v1.0 (Phase 5)

### Active

- [ ] **QUAL-02**: Critical frontend auth and navigation flows have automated tests covering current app behavior.
- [ ] **QUAL-03**: Project setup documents the generated-code or verification workflow well enough that contributors can run core checks reliably.

### Out of Scope

- Multi-user collaboration or shared debt management - this project is for personal use.
- Payment execution or automatic repayment deduction - not part of the current management goal.
- Reminder/notification workflows - explicitly deferred from v1.
- Large-screen analytics or advanced visualization dashboards - deferred until the core flows are stable.

## Context

This is a brownfield repository with two application roots:

- `blog/`: a Go backend built with Kratos, protobuf-first contracts, GORM, and PostgreSQL-oriented data access.
- `price_recorder_vue/`: a Vue 3 + Vite frontend with Vue Router, Pinia, and Axios-based API calls.

**Current state after v1.0:**
- Authentication is stable: reactive Pinia store, aligned Axios interceptor, and proper router guards.
- Blog module is fully functional end to end: list, view, create, edit, delete with toast feedback.
- Debt module is fully functional end to end: list, view, create, edit, delete with summary statistics and toast feedback.
- Database configuration supports environment variable overrides for flexible deployment.
- Backend has comprehensive unit tests for debt (data, biz, service) and post (biz, service) layers.
- Frontend automated tests and contributor documentation remain as the primary gaps for v1.1.

## Constraints

- **Tech stack**: Continue with the current Go Kratos backend and Vue 3 frontend.
- **Product scope**: Personal single-user workflow only.
- **Architecture**: Blog and debt remain in one site with one account system.
- **Codebase reality**: Frontend test coverage and generated-code workflow documentation are the next priorities.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Keep blog and debt in one product | One account system with two functional areas in the same site | Validated in v1.0 |
| Prioritize blog work before debt expansion | Blog is the primary value stream | Confirmed; both are now stable |
| Continue on the existing Kratos + Vue stack | Working foundations already existed | Confirmed; v1.0 delivered on this stack |
| Keep debt v1 focused on records plus summary statistics | Explicitly excluded reminders, payments, collaboration, advanced visualization | Validated in v1.0 |
| Use SQLite in-memory for backend unit tests | Fast, isolated execution without Docker | Validated in Phase 7 |
| Use manual mock repositories instead of testify/mock | Simplicity and clarity in test code | Validated in Phases 6-7 |

---
*Last updated: 2026-04-13 after v1.0 milestone completion*
