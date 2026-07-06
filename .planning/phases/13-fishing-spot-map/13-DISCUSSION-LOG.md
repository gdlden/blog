# Phase 13: Fishing Spot Map - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

---

# Re-Discussion: Frontend Resumption (Plan 13-03)

**Date:** 2026-07-06
**Mode:** default (interactive) — single-question turns per area
**Prior state:** Backend shipped via 13-01/13-02; verification report marked 4/8 user-facing truths FAILED because frontend never built. User invoked `/gsd-discuss-phase 13 把前端也写了` to lock the remaining frontend gray areas.
**Areas discussed:** AppLayout 导航扩展, Spot capture 交互, Tag filter + Search UX, Vitest 测试覆盖深度

## Pre-flight: continue or scrap?

| Option | Description | Selected |
|---|---|---|
| 追加前端 CONTEXT + 新 plan | Append frontend decisions to existing CONTEXT.md; regenerate Plan 13-03 via /gsd-plan-phase | ✓ |
| 查看现有 CONTEXT/Plans | Read 13-CONTEXT.md + 13-01/13-02 SUMMARY first | |
| 重写整个 CONTEXT | Discard existing CONTEXT, regenerate from scratch | |

**Notes:** Append preserves the locked D-01..D-13 backend + frontend-directional decisions, only adds D-14..D-22 for Plan 13-03 specifics.

## Pre-flight: gray-area selection

Multi-select of 4 phase-specific gray areas. User chose all 4.

## Area 1: AppLayout 导航扩展

State drift identified: existing CONTEXT.md D-13 implied only 博客/债务; actual `AppLayout.vue` `navItems` already has 5 entries.

### Q1.1 — Where does "地图" go?

| Option | Description | Selected |
|---|---|---|
| 末尾第 6 项"地图" | Append as 6th entry; lowest-diff, follows existing router-link pattern | ✓ |
| 折叠"更多" dropdown | Collapse 价格/版本 into "更多", keep 4 main tabs | |
| 替换"版本"主入口 | Move 版本管理 to /settings sub-page; heaviest change | |

### Q1.2 — Chinese label text?

| Option | Description | Selected |
|---|---|---|
| 地图 | Generic short label, matches original CONTEXT D-13 | ✓ |
| 钓点 | Closer to use case but narrows future scope | |
| 钓位地图 | More explicit but slightly long | |

**Notes:** Recorded as D-14; supersedes D-13's "alongside 博客/债务" wording.

## Area 2: Spot capture 交互

### Q2.1 — Form placement after GPS captured?

| Option | Description | Selected |
|---|---|---|
| Bottom sheet 半遮地图 | Apple/Google Maps habit; reuses D-10 detail sheet visually | ✓ |
| 全屏覆盖表单 | Simpler but loses "see the point you're saving" context | |
| 侧滑抽屉 modal | Reuses DebtDetail modal pattern but breaks full-screen map theme | |

### Q2.2 — Multi-photo upload UX?

| Option | Description | Selected |
|---|---|---|
| 多选即时上传带进度条 | Each file POST /file/v1 immediately, parallel, progress per file | ✓ |
| 多选后提交时一起上传 | Deferred upload to spot-submit; risky if interrupted | |
| 仅一张主图 + 详情页补 | Lighter but doesn't match D-06 "many photos" intent | |

### Q2.3 — Form field priority on mobile?

| Option | Description | Selected |
|---|---|---|
| 只 name 必填，其他折叠 | name + submit always on screen; notes/tags/photos behind chevron | ✓ |
| name + notes 默认展开 | Two-field first screen | |
| 所有字段并列展开 | Largest height pressure | |

**Notes:** D-15 + D-16 + D-17 locked. Reaffirms D-09. Recommended reusing DebtDetail.vue modal/transition pattern for the custom bottom sheet.

## Area 3: Tag filter + Search UX

### Q3.1 — Default presentation?

| Option | Description | Selected |
|---|---|---|
| 默认隐藏，FAB 唤出 | Top-right FAB → slide-down overlay; full map canvas by default | ✓ |
| 默认隐藏 + modal 弹层 | Different animation syntax | |
| 始终置顶部 sticky bar | Eats 30–40 px vertical on mobile | |
| 折叠 chip 展开 | Chip row auto-collapses; search hidden in menu | |

### Q3.2 — Filter selection behavior?

| Option | Description | Selected |
|---|---|---|
| 隐藏不匹配 pin + 地图不变 | Just hide markers; user keeps manual pan/zoom | ✓ |
| 隐藏 + 自动平移到首匹配 | Machine-driven panning risks disorienting | |
| 不隐藏，颜色高亮匹配 | Higher implementation complexity | |

### Q3.3 — Empty / initial state?

| Option | Description | Selected |
|---|---|---|
| 用户当前 GPS 为中心 + soft tooltip | Most useful first view; uses AMap.Geolocation | ✓ |
| 固定市级中心 + 提示 | Works without GPS permission (Guangzhou example) | |
| 空白地图无任何暗示 | User doesn't know next step | |

**Notes:** D-18 + D-19 + D-20 locked. "fit view" toolbar button (`map.setFitView()`) kept as separate non-auto-triggered control. Capture FAB (bottom edge "+") and filter FAB (top-right) MUST be separate buttons.

## Area 4: Vitest 测试覆盖深度

### Q4.1 — Test scope?

| Option | Description | Selected |
|---|---|---|
| 只要 store + api client | Mirror existing fuel.spec / fuelStore.spec / request.spec patterns | |
| store + api + 组件挂载 mock AMap | Plus Map.vue mount with mocked window.AMapLoader, AMap, geolocation, URL.createObjectURL | ✓ |
| 完全不加测试 | Defer to future QUAL-02 | |

### Q4.2 — Live map behaviors that jsdom/happy-dom can't exercise?

| Option | Description | Selected |
|---|---|---|
| 跳过用 vi.skipIf / it.skip + 注释 | Explicit enumerated skip with rationale | ✓ |
| 测逻辑 path 不测地图实例 | mapStore holds marker set as plain data | |
| 引入 happy-dom 替代 jsdom | Repo already uses happy-dom; no switch needed | |

**Notes:** D-21 + D-22 locked. Tests that need a live map instance (real tile rendering, real touch-pinch gestures, Geolocation-denied fallback flow on a real device) are explicitly enumerated as skipped and re-routed to `13-VALIDATION.md` manual UAT so the verifier does not re-flag them as gaps.

## the agent's Discretion (still open for planner)

Original "Agent's Discretion" items NOT user-facing and remain planner's call:
- Pin clustering (`AMap.MarkerClusterer`) — planner decides include vs defer
- DB schema already shipped
- Coordinate conversion location locked to frontend by RESEARCH
- GPS failure handling: D-20 defines empty-state fallback; planner choreographs toast copy
- Gaode key config: `VITE_GAODE_JS_API_KEY` + `VITE_GAODE_JS_SECURITY_CODE` env vars; planner adds `.env.example` and README docs

## Deferred Ideas

None. Discussion stayed strictly within Phase 13 frontend scope; no scope-creep surfaced.

## Reviewer-flagged follow-ups (closed within Plan 13-03, not deferred)

After 13-VERIFICATION flagged broken requirement traceability (MAP-01/02/03 exist only in ROADMAP, never registered in REQUIREMENTS.md), this re-discussion explicitly tasks Plan 13-03 with closing that gap. Recorded in CONTEXT.md `<specifics>` section.

---

**Date:** 2026-07-05
**Phase:** 13-Fishing Spot Map
**Areas discussed:** Map provider, Data model scope, Capture UX flow, Navigation & list view

---

## Map provider

| Option | Description | Selected |
|--------|-------------|----------|
| Gaode (Amap) | Native GCJ-02 support, no offset for China GPS, JS API for web, deep-link for navigation | ✓ |
| Baidu Maps | BD-09 coordinate system, popular alternative | |
| Leaflet + free tiles | Open source, no API key, but OSM tiles show offset in China | |

**User's choice:** Gaode (Amap)
**Notes:** Best accuracy for China — coordinates match real GPS positions without offset correction.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Map primary | Full-screen map with pins, most intuitive for location data | ✓ |
| List primary, tap for map | Simpler, list-first approach | |
| Both — tabbed | Two tabs: list and map, most flexible but more code | |

**User's choice:** Map primary
**Notes:** Feels like a proper map app. Tap pins for details.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Open Gaode app | Deep-link to Amap native app for turn-by-turn navigation | ✓ |
| In-app directions | Use Gaode JS API for route drawing in web | |
| Both options | Offer both app deep-link and static route | |

**User's choice:** Open Gaode app
**Notes:** Best UX for actually finding the spot again — tap to launch navigation.

---

## Data model scope

| Option | Description | Selected |
|--------|-------------|----------|
| Full structure | name + lat/lng + notes + timestamps + photos + tags | ✓ |
| With photos | name + lat/lng + notes + timestamps + photos | |
| Basic | name + lat/lng + notes + timestamps | |

**User's choice:** Full structure with photos and tags
**Notes:** Most expressive — covers all the information needed to find and remember a spot.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Free-text comma input | User types custom tags like "carp, night fishing, river" | ✓ |
| Fixed preset tags | Predefined categories to pick from | |
| Both | Presets as quick-select + free-text field | |

**User's choice:** Free-text comma input
**Notes:** Maximum flexibility — no constraints on what categories the user wants.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Multiple photos, backend upload | Upload 1-N photos, reuse existing file infrastructure | ✓ |
| Single photo | Simpler, one photo per spot | |
| No photos for now | Defer to future phase | |

**User's choice:** Multiple photos, backend upload
**Notes:** Existing `blog/api/file/v1/file.proto` can be reused.

---

## Capture UX flow

| Option | Description | Selected |
|--------|-------------|----------|
| One-tap "save here" button | GPS current location, then open form | ✓ |
| Long-press on map | Drop pin anywhere, confirm | |
| Both | Button as primary, long-press as alternative | |

**User's choice:** One-tap button
**Notes:** Fastest for field use when actually standing at the spot.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, show address | Gaode Geocoding API to resolve coordinates into address | ✓ |
| No, only coordinates | Just lat/lng numbers | |
| Address as default name | Use geocoded address as the spot name | |

**User's choice:** Yes, show address
**Notes:** Helps user verify they saved the right spot — seeing "XX河" is more meaningful than raw coordinates.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Only name required | Name + coords required, notes/tags/photos optional at capture | ✓ |
| Name + tags required | Forces categorization upfront | |
| Coordinates only | Save first, fill details later | |

**User's choice:** Only name required
**Notes:** Respects mobile typing friction — user can fill in details later at home.

---

## Navigation & list view

| Option | Description | Selected |
|--------|-------------|----------|
| Bottom sheet panel | Slide-up panel with details, edit/delete/navigate buttons | ✓ |
| Detail page | Full page transition to spot detail | |
| Popover then choose | Small bubble with name + address, buttons to drill deeper | |

**User's choice:** Bottom sheet panel
**Notes:** Native-feeling mobile UX. Immediate access to actions without full page navigation.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Bottom sheet buttons | Edit and delete buttons directly on the sheet | ✓ |
| Swipe/long-press gestures | Gesture-based, less suitable for map view | |
| Only in detail page | All management in separate page | |

**User's choice:** Bottom sheet buttons
**Notes:** Quick access — don't make the user navigate to another page just to delete.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Tag filter + search box | Filter bar by tags + name search at top of map | ✓ |
| Search only | Name search only | |
| None for now | No filter/search in v1 | |

**User's choice:** Tag filter + search box
**Notes:** Having free-text tags makes tag filtering valuable for finding spots by category.

---

| Option | Description | Selected |
|--------|-------------|----------|
| New top-level "地图" tab | New tab alongside "博客" and "债务" | ✓ |
| Under blog section | As blog sub-page | |
| User menu entry | Inside user avatar dropdown | |

**User's choice:** New top-level "地图" tab
**Notes:** Independent, prominent entry point — fishing spots are their own domain.

---

## Agent's Discretion

- Coordinate storage format (WGS-84 vs GCJ-02 conversion strategy)
- Proto message design and API endpoint structure
- Database table schema
- Gaode API key configuration
- Pin clustering on crowded maps
- GPS failure/denied permission error handling
- Bottom sheet component implementation approach
- Coordinate conversion — frontend or backend

## Deferred Ideas

None — discussion stayed within phase scope.
