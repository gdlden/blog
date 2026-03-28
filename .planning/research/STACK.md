# Stack Research

**Project:** Blog Debt Hub
**Date:** 2026-03-26
**Confidence:** High

## Recommended Stack Direction

Stay on the current stack rather than re-platforming:

- **Backend:** Go + Kratos + protobuf contracts + GORM
- **Frontend:** Vue 3 + Vite + Vue Router + Pinia + Axios
- **Data:** Relational persistence for blog posts, users, debts, and debt details
- **Testing:** Expand Go unit/service tests and Vitest coverage instead of adding a second test stack

## Why This Stack Fits

- The codebase already uses clear backend layering (`service -> biz -> data`) and versioned API contracts.
- Blog and debt are both CRUD-heavy domains, which fit the current transport and data model well.
- The main bottleneck is not stack capability; it is incomplete flows, fragile auth state handling, and low regression coverage.

## Prescriptive Choices

- Keep protobuf-first backend APIs and continue versioned paths under `blog/api/**`.
- Keep one shared auth/session model for blog and debt.
- Use the existing Vue SPA as the single frontend shell for both functional areas.
- Favor incremental hardening over broad architecture changes.

## Avoid

- Rewriting to a new frontend or backend framework before the current product goals are stabilized.
- Adding complex infrastructure such as real-time messaging, job orchestration, or analytics platforms before core CRUD and testability are solid.
- Splitting the product into separate blog and debt applications.

## Confidence Notes

- **High:** existing stack already matches current product scope.
- **Medium:** production hardening needs additional work around config, secrets, and generated code workflow.
