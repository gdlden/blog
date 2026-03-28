# Blog Debt Hub

## What This Is

Blog Debt Hub is a personal full-stack site that combines a blog area and a debt-management area under one account system. It is intended for a single user to write and manage blog content at any time, while also tracking personal debt records and reviewing debt statistics in the same place.

The current codebase already contains backend Kratos services and a Vue frontend for login, article access, and debt-related APIs. This project initialization defines the next focused direction: stabilize the blog experience first, then continue improving debt record and statistics workflows on top of the existing foundation.

## Core Value

At any time, the user can reliably record and manage their own blog content, then use the same site to review and manage personal debt information.

## Requirements

### Validated

- [x] User can log in to the site with an existing account flow backed by the current backend and frontend implementation.
- [x] User can access existing blog post APIs for creating posts, listing posts, and viewing post details through the current backend contracts.
- [x] User can access existing debt APIs to create, update, delete, query, and list debt records in the current backend structure.
- [x] User can access debt-detail APIs and supporting backend domain structure, although some CRUD paths still need completion and hardening.

### Active

- [ ] Stabilize the personal blog workflow so writing, managing, and browsing posts becomes the strongest and most reliable part of the site.
- [ ] Keep the blog area and debt area under one consistent account system and navigation structure.
- [ ] Improve the debt area so core debt record details and summary statistics are usable for day-to-day personal tracking.
- [ ] Add enough test and verification coverage that the existing blog and debt flows can be changed safely.

### Out of Scope

- Multi-user collaboration or shared debt management - this project is for personal use.
- Payment execution or automatic repayment deduction - not part of the current management goal.
- Reminder/notification workflows - explicitly deferred from v1.
- Large-screen analytics or advanced visualization dashboards - deferred until the core flows are stable.

## Context

This is a brownfield repository with two application roots:

- `blog/`: a Go backend built with Kratos, protobuf-first contracts, GORM, and PostgreSQL-oriented data access.
- `price_recorder_vue/`: a Vue 3 + Vite frontend with Vue Router, Pinia, and Axios-based API calls.

Existing codebase mapping shows the following realities:

- The backend already exposes blog, user, debt, debt detail, price, OCR, and file-related service layers.
- The frontend already contains login, article, and basic placeholder pages, but its current UX is still thin and some auth behavior is fragile.
- Debt-detail update/delete behavior is incomplete, generated protobuf outputs are not fully committed, and automated test coverage is currently weak.
- The current route structure and feature surface suggest the repository is best treated as an evolving personal operations site rather than a clean-slate blog product.

## Constraints

- **Tech stack**: Continue with the current Go Kratos backend and Vue 3 frontend - the user explicitly wants to evolve the existing stack rather than replace it.
- **Product scope**: Personal single-user workflow only - no need to design for collaboration, teams, or multi-tenant behavior right now.
- **Priority**: Blog comes first in v1 - debt functionality should remain available, but the roadmap should favor blog usability and stability before deeper debt expansion.
- **Architecture**: Blog and debt remain in one site with one account system - the roadmap should avoid splitting them into separate apps.
- **Codebase reality**: Planning must account for existing technical debt in auth flow, debt-detail CRUD completion, generated code workflow, and weak test coverage.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Keep blog and debt in one product | The user wants one account system with two functional areas in the same site | Pending |
| Prioritize blog work before debt expansion | The blog is the primary value stream and should become stable first | Pending |
| Continue on the existing Kratos + Vue stack | The repository already contains working foundations in both backend and frontend | Pending |
| Keep debt v1 focused on records plus summary statistics | The user wants debt management, but not reminders, payments, collaboration, or advanced visualization | Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `$gsd-transition`):
1. Requirements invalidated? Move to Out of Scope with reason
2. Requirements validated? Move to Validated with phase reference
3. New requirements emerged? Add to Active
4. Decisions to log? Add to Key Decisions
5. "What This Is" still accurate? Update if drifted

**After each milestone** (via `$gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check - still the right priority?
3. Audit Out of Scope - reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-26 after initialization*
