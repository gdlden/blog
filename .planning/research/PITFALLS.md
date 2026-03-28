# Pitfalls Research

**Project:** Blog Debt Hub
**Date:** 2026-03-26

## Pitfall 1: Treating existing endpoints as fully complete

- **Warning signs:** APIs exist on paper, but UI flows are missing or service methods are stubbed
- **Prevention:** verify each user flow end to end before calling it complete
- **Phase mapping:** blog workflow phase, debt hardening phase

## Pitfall 2: Leaving auth/session behavior fragile

- **Warning signs:** route guards cache login state once, logout is unclear, failed login paths do not propagate cleanly
- **Prevention:** centralize auth state, recompute from store/state changes, and test login/logout/error flows
- **Phase mapping:** foundation phase

## Pitfall 3: Expanding debt features before blog is stable

- **Warning signs:** roadmap spends early phases on advanced debt analysis while basic blog management remains weak
- **Prevention:** keep roadmap aligned to the declared core value and v1 priority
- **Phase mapping:** roadmap design, phase sequencing

## Pitfall 4: Planning around idealized architecture instead of current repo reality

- **Warning signs:** roadmap assumes clean generated-code workflow, strong tests, and complete CRUD behavior that the repo does not actually have
- **Prevention:** incorporate codebase concerns directly into requirements and success criteria
- **Phase mapping:** all phases

## Pitfall 5: Deferring testing until after feature work

- **Warning signs:** repeated regressions in login, routing, or debt logic
- **Prevention:** add targeted backend and frontend regression tests as each core workflow is stabilized
- **Phase mapping:** every implementation phase
