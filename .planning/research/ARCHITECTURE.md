# Architecture Research

**Project:** Blog Debt Hub
**Date:** 2026-03-26

## Suggested Product Structure

Keep the current two-part product inside one application shell:

1. **Account and session foundation**
2. **Blog area**
3. **Debt area**
4. **Shared navigation and API client behavior**
5. **Verification and hardening layer**

## Component Boundaries

- **Backend auth/user component:** login, account identity, token handling, current-user extraction
- **Backend blog component:** post contracts, post service/use case/repository, article management endpoints
- **Backend debt component:** debt and debt-detail contracts, debt aggregation, debt record ownership
- **Frontend shell component:** router, navigation, auth guard, global request behavior
- **Frontend blog component:** article list/detail/editor/manage views
- **Frontend debt component:** debt list/detail/statistics views

## Data Flow

- User authenticates once, then accesses both blog and debt areas through the same frontend shell.
- Frontend routes call domain-specific API wrappers through a shared Axios instance.
- Backend services map requests into business use cases and then repositories for persistence.
- Debt summaries should be derived from persisted debt and debt-detail data, then surfaced through API responses and simple frontend views.

## Suggested Build Order

1. Stabilize auth/session behavior shared by both product areas
2. Make blog management end to end and pleasant to use
3. Close debt CRUD gaps and expose reliable debt summaries
4. Increase automated test coverage around both domains
5. Revisit richer debt analysis only after the core workflows are dependable

## Phase Implications

- Phase boundaries should follow user-visible workflows instead of technical layer boundaries.
- Blog should receive earlier phases than debt because it is the declared v1 priority.
- Technical-debt reduction should be attached to feature phases where it removes concrete product risk.
