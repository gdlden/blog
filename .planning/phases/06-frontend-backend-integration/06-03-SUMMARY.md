---
phase: 06-frontend-backend-integration
plan: "03"
subsystem: frontend
tags: [debt, frontend, pinia, api, crud]
requires: [06-02]
provides: [debt-module-complete]
affects: [price_recorder_vue/src/api/debt.ts, price_recorder_vue/src/stores/debtStore.ts, price_recorder_vue/src/view/DebtList.vue]
tech-stack:
  added: []
  patterns: [pinia-setup-store, api-client-pattern, card-grid-ui, modal-form]
key-files:
  created:
    - price_recorder_vue/src/api/debt.ts
    - price_recorder_vue/src/stores/debtStore.ts
  modified:
    - price_recorder_vue/src/view/DebtList.vue
  deleted: []
decisions:
  - Used '已结清' and 'repaid' status values for repaid amount calculation to support both Chinese and English status values
  - Implemented 2-column grid layout in modal for better form organization on desktop
  - Added status color coding (green for repaid, yellow for active, gray for unknown)
  - Used number input types with step="0.01" for monetary fields (amount, apr, fee)
  - Summary cards show formatted amounts with ¥ prefix and Chinese locale formatting
metrics:
  duration: 2 min
  completed-date: "2026-04-05"
  commits: 3
  files-changed: 3
---

# Phase 06 Plan 03: Frontend Debt Module with CRUD and Summary

## One-Liner

Complete debt module frontend with full CRUD operations, summary statistics, and responsive card grid UI.

## What Was Built

### 1. Debt API Client (`price_recorder_vue/src/api/debt.ts`)

Created a domain-specific API client following the same pattern as the blog API:

- **Debt interface**: Full TypeScript interface matching the backend protobuf contract (id, name, bankName, bankAccount, applyTime, endTime, amount, status, remark, apr, fee, tenor)
- **DebtPageResponse interface**: For paginated list responses (page, total, list)
- **getDebts(page?, pageSize?)**: GET /debt/page/v1 with pagination
- **getDebtById(id)**: GET /debt/get/v1 for single debt retrieval
- **createDebt(data)**: POST /debt/save/v1 for creating new debts
- **updateDebt(data)**: POST /debt/update/v1 for updating existing debts
- **deleteDebt(id)**: POST /debt/delete/v1 returning boolean flag

### 2. Debt Pinia Store (`price_recorder_vue/src/stores/debtStore.ts`)

Implemented a setup store (Composition API style) with state, computed getters, and actions:

**State:**
- `debts`: Array of debt records
- `currentDebt`: Currently selected debt for detail view
- `loading`: Loading state indicator
- `error`: Error message storage
- `total`, `currentPage`, `pageSize`: Pagination state

**Getters (including summary statistics per DEBT-06):**
- `debtCount`: Number of debts in current page
- `totalPages`: Calculated total pages
- `totalDebt`: Sum of all debt amounts (parseFloat with fallback to 0)
- `repaidAmount`: Sum of debts with status '已结清' or 'repaid'
- `outstandingAmount`: Sum of debts not marked as repaid

**Actions:**
- `fetchDebts(page?, size?)`: Load paginated debt list
- `createDebt(data)`: Create and refresh list
- `updateDebt(data)`: Update and refresh list
- `deleteDebt(id)`: Delete and remove from local array
- `setCurrentDebt(debt)`: Set current selection

### 3. DebtList View (`price_recorder_vue/src/view/DebtList.vue`)

Replaced placeholder with complete CRUD UI:

**Header:**
- "债务" title with "新建债务" button (blue, top-right)

**Summary Cards (3-column grid):**
- 总债务 (Total Debt): Gray/black text
- 已还款 (Repaid): Green text for positive indicator
- 未还金额 (Outstanding): Red text for attention

**Card Grid (responsive):**
- 1 column on mobile, 2 on tablet, 3 on desktop
- Each card shows: name (truncated), bankName, amount, status (color-coded), remark preview
- Action buttons: 编辑 (blue), 删除 (red)

**Modal Form:**
- 2-column grid layout for desktop, single column on mobile
- Fields: name, bankName, bankAccount, amount, applyTime, endTime, status (dropdown), apr, fee, tenor, remark
- Date inputs for applyTime/endTime
- Select dropdown for status (进行中/已结清)
- Number inputs with decimal support for monetary fields
- Cancel and Save buttons with loading state

**Pagination:**
- Previous/Next buttons with disabled states
- Page number buttons with active state highlighting

**States:**
- Loading: "加载中..." centered text
- Empty: "暂无内容" with guidance

## Deviations from Plan

None - plan executed exactly as written.

## Implementation Notes

### Status Handling
The summary calculations check for both Chinese ('已结清') and English ('repaid') status values to be resilient to backend data variations. The status dropdown in the form uses Chinese values ('进行中', '已结清') as the primary interface language.

### Form Reset
The modal form resets all fields to empty strings when closing or switching between create/edit modes to prevent data leakage between operations.

### Amount Formatting
Summary amounts are formatted using `toLocaleString('zh-CN')` with 2 decimal places for consistent currency display.

## Commits

| Hash | Type | Message |
|------|------|---------|
| eb6ce22 | feat | create debt API client with CRUD operations |
| fedf8a6 | feat | create debt Pinia store with summary statistics |
| 8861b74 | feat | implement DebtList.vue with CRUD UI and summary statistics |

## Verification

- [x] Debt API client exports all CRUD functions (getDebts, getDebtById, createDebt, updateDebt, deleteDebt)
- [x] Debt store manages state with summary computed properties (totalDebt, repaidAmount, outstandingAmount)
- [x] DebtList displays debts in card grid layout with summary cards
- [x] Modal opens for create and edit operations
- [x] Delete has confirmation dialog
- [x] Pagination controls implemented
- [x] Summary shows accurate totals with proper filtering

## Self-Check: PASSED

- [x] Created files exist: price_recorder_vue/src/api/debt.ts, price_recorder_vue/src/stores/debtStore.ts
- [x] Modified file exists: price_recorder_vue/src/view/DebtList.vue
- [x] All commits verified: eb6ce22, fedf8a6, 8861b74
