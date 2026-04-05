---
phase: 06
slug: frontend-backend-integration
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-05
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Vitest 4.0.17 |
| **Config file** | `vitest.config.ts` |
| **Quick run command** | `cd price_recorder_vue && pnpm test:unit` |
| **Full suite command** | `cd price_recorder_vue && pnpm test:unit` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Manual UI verification via `pnpm dev`
- **After every plan wave:** Full API testing via UI
- **Before `/gsd:verify-work`:** All CRUD operations verified end-to-end
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | BLOG-04/05 | build | `go build ./cmd/blog` | ❌ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | BLOG-04/05 | proto | `make api` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 2 | BLOG-01 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 2 | BLOG-03 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-02-03 | 02 | 2 | BLOG-04 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-02-04 | 02 | 2 | BLOG-05 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-03-01 | 03 | 3 | DEBT-01 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-03-02 | 03 | 3 | DEBT-02 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-03-03 | 03 | 3 | DEBT-04 | integration | Manual UI check | ❌ W0 | ⬜ pending |
| 06-03-04 | 03 | 3 | DEBT-05 | integration | Manual UI check | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `src/stores/blogStore.ts` - Pinia store for blog state
- [ ] `src/stores/debtStore.ts` - Pinia store for debt state
- [ ] `src/api/blog.ts` - Blog API client
- [ ] `src/api/debt.ts` - Debt API client
- [ ] `src/components/Modal.vue` - Reusable modal component
- [ ] `vue-toastification` package installation

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Blog list display | BLOG-01 | Requires UI rendering | 1. Navigate to Blog page<br>2. Verify posts appear as cards<br>3. Check pagination works |
| Create blog post | BLOG-03 | Requires UI interaction | 1. Click "新建博文"<br>2. Fill form<br>3. Submit<br>4. Verify appears in list |
| Edit blog post | BLOG-04 | Requires UI interaction | 1. Click "编辑" on card<br>2. Modify content<br>3. Save<br>4. Verify changes |
| Delete blog post | BLOG-05 | Requires UI interaction | 1. Click "删除" on card<br>2. Confirm<br>3. Verify removed from list |
| Debt CRUD | DEBT-01/02/04/05 | Requires UI interaction | Same pattern as Blog |
| Toast notifications | INT-01 | Visual feedback | Verify toast appears on create/update/delete |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
