---
phase: 08
slug: api-response-standardization
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-14
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go test + Vitest |
| **Config file** | `blog/go.mod` / `price_recorder_vue/vitest.config.ts` |
| **Quick run command** | `cd blog && go test ./internal/server/...` + `cd price_recorder_vue && pnpm test:unit --run` |
| **Full suite command** | `cd blog && go test ./...` + `cd price_recorder_vue && pnpm test:unit --run` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run quick run command for modified subsystem
- **After every plan wave:** Run full suite command
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | API-04 | — | Encoder wraps without leaking raw errors | unit | `cd blog && go test ./internal/server/...` | ✅ | ⬜ pending |
| 08-01-02 | 01 | 1 | API-01, API-02 | — | All endpoints return `{code, message, data}` | integration | `curl` or API test | ✅ | ⬜ pending |
| 08-02-01 | 02 | 1 | API-05 | — | Interceptor unwraps `res.data.data` | unit | `cd price_recorder_vue && pnpm test:unit --run` | ✅ | ⬜ pending |
| 08-02-02 | 02 | 1 | API-05 | — | API clients no longer double-extract `.data` | static | `grep -r "res.data" src/api/` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| End-to-end response shape | API-01, API-02 | Requires running backend + frontend together | Start backend (`make run`), verify HTTP response shape via browser DevTools or `curl` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
