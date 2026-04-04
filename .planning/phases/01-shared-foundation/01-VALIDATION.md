---
phase: 1
slug: shared-foundation
status: approved
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-04
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Vitest 4.0.18 |
| **Config file** | price_recorder_vue/vitest.config.ts |
| **Quick run command** | `cd price_recorder_vue && pnpm test:unit` |
| **Full suite command** | `cd price_recorder_vue && pnpm test:unit && pnpm lint && pnpm type-check` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd price_recorder_vue && pnpm test:unit`
- **After every plan wave:** Run `cd price_recorder_vue && pnpm test:unit && pnpm lint && pnpm type-check`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 1 | FND-01 | integration | `pnpm test:unit` | ❌ W0 | ⬜ pending |
| 1-01-02 | 01 | 1 | FND-02 | unit | `pnpm test:unit` | ❌ W0 | ⬜ pending |
| 1-02-01 | 02 | 1 | FND-03 | component | `pnpm test:unit` | ❌ W0 | ⬜ pending |
| 1-02-02 | 02 | 1 | FND-01 | unit | `pnpm test:unit` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `price_recorder_vue/src/__tests__/stores/userStore.spec.ts` — covers FND-01, FND-02 store behavior
- [ ] `price_recorder_vue/src/__tests__/router/index.spec.ts` — covers FND-01, FND-02, FND-03 navigation guard behavior
- [ ] `price_recorder_vue/src/__tests__/components/AppLayout.spec.ts` — covers FND-03 navigation menu and logout
- [ ] `price_recorder_vue/src/__tests__/view/Login.spec.ts` — covers FND-01 login flow with redirect handling
- [ ] Test utilities for mocking localStorage and Vue Router

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Page refresh restores session | FND-01 | localStorage + browser refresh behavior requires real browser | 1. Log in. 2. Refresh page. 3. Verify still authenticated and on intended page. |
| Visual layout of navigation shell | FND-03 | Pixel-perfect layout verification | 1. Log in. 2. Verify sidebar and main content render correctly at common viewport sizes. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
