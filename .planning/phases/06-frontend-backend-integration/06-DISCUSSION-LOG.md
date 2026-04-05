# Phase 6: Frontend-Backend Integration - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-05
**Phase:** 06-frontend-backend-integration
**Areas discussed:** Blog Edit/Delete API, API Client Organization, State Management, List Page Design, Form Interaction, Error Handling

---

## Blog Edit/Delete API

| Option | Description | Selected |
|--------|-------------|----------|
| 启用完整 CRUD | Uncomment proto APIs, implement Update/Delete in service/biz/data | ✓ |
| 仅只读展示 | Keep proto as-is, frontend only does Create/List/View | |

**User's choice:** 启用完整 CRUD（推荐）
**Notes:** Backend needs to add Update and Delete methods to PostService and PostUsecase. Proto file has EditPostById and DeletePostById commented out.

---

## API Client Organization

| Option | Description | Selected |
|--------|-------------|----------|
| 按领域分文件 | blog.ts, debt.ts — each domain has its own API file | ✓ |
| 按操作类型分 | list.ts, detail.ts — mixed domains by operation type | |

**User's choice:** 按领域分文件（推荐）
**Notes:** Clean separation, matches backend service boundaries

---

## State Management Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Pinia Stores | useBlogStore, useDebtStore — centralized state management | ✓ |
| 组件级状态 | Each component manages its own with ref/reactive | |

**User's choice:** Pinia Stores（推荐）
**Notes:** Project already has Pinia configured, consistent with existing userStore pattern

---

## List Page Design

| Option | Description | Selected |
|--------|-------------|----------|
| 表格布局 | Column-based, sortable, admin-style | |
| 卡片网格 | Visual cards, content-focused, lower density | ✓ |

**User's choice:** 卡片网格
**Notes:** Better for content display, more visual appeal

---

## Form Interaction

| Option | Description | Selected |
|--------|-------------|----------|
| Modal 对话框 | Edit within list page, context preserved, quick | ✓ |
| 独立页面路由 | /blog/create, /blog/edit/:id — shareable URLs | |
| 行内编辑 | Edit directly in list row — fastest but complex UI | |

**User's choice:** Modal 对话框（推荐）
**Notes:** Best balance of UX simplicity and context preservation

---

## Error Handling & Feedback

| Option | Description | Selected |
|--------|-------------|----------|
| Toast 通知 | Auto-dismiss, non-blocking, modern UX | ✓ |
| Modal 确认 | Blocking, requires user action | |

**User's choice:** Toast 通知（推荐）
**Notes:** Standard modern pattern, doesn't interrupt flow

---

## Claude's Discretion

- Exact card styling and spacing
- Toast library selection
- Modal component implementation
- Form validation rules

## Deferred Ideas

- Image upload for blog posts — future enhancement
- Advanced filtering/search — future enhancement
- Bulk operations — future enhancement
- Real-time updates via WebSocket — out of scope
