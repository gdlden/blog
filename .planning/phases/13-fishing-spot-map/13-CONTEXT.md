# Phase 13: Fishing Spot Map - Context

**Gathered:** 2026-07-05
**Re-discussed:** 2026-07-06 (frontend resumption — Plan 13-03)
**Status:** Backend shipped (13-01/13-02); frontend ready for planning (Plan 13-03)

<domain>
## Phase Boundary

Save and manage GPS-tagged fishing spots — capture current coordinates with one tap, name and describe the spot, attach photos and tags, and navigate back to it later via Gaode Maps. Adds a new "地图" top-level tab to the existing 5-tab AppLayout nav (博文/债务/油耗/价格/版本) with a map-primary full-screen view showing all saved spots as pins.

**Current state:** Backend half (proto contract + GORM data layer + biz usecase + Gaode Web regeo + MapService handler + JWT-protected HTTP routes + Wire wiring) shipped via plans 13-01 and 13-02 and compiles cleanly. The 4 user-facing ROADMAP success criteria are BLOCKED pending the frontend (Plan 13-03). This resumption defines the frontend implementation decisions for Plan 13-03.

</domain>

<decisions>
## Implementation Decisions

### Map Provider
- **D-01:** Use Gaode (Amap) Maps. Native GCJ-02 coordinate system support — no offset for China GPS. Gaode JS API for webview display, deep-link to Gaode App for turn-by-turn navigation.
- **D-02:** Map-primary view — main screen is a full-screen Gaode map with pins for all saved spots. No separate list view in v1.
- **D-03:** "Navigate" button deep-links to Gaode App with the spot's coordinates. Fallback to web if app not installed.

### Data Model
- **D-04:** Spot entity fields: name (required), latitude, longitude, notes (optional), tags (comma-separated free text), photos (multiple image URLs), created_at, updated_at.
- **D-05:** Tags are free-text comma-separated input. No preset categories. User types whatever they want (e.g. "carp, night fishing, river").
- **D-06:** Multiple photos per spot. Use existing file upload infrastructure (`blog/api/file/v1/file.proto`). Store image URLs in DB.

### Capture UX
- **D-07:** One-tap "save current location" button on the map. Calls browser Geolocation API to get WGS-84 coordinates, opens a form to enter name/notes/tags/photos.
- **D-08:** Reverse geocode coordinates via Gaode Geocoding API to show a human-readable address (e.g. "广东省XX市XX镇XX河") alongside coordinates in the capture form.
- **D-09:** Only name is required when first saving. Notes, tags, and photos are optional at capture time — user can edit later to add details.

### View & Navigation
- **D-10:** Tapping a pin on the map opens a bottom sheet panel showing: name, address, tags, photo thumbnails, notes. Action buttons: Edit, Delete (with confirm), Navigate (open Gaode App).
- **D-11:** Edit button opens a full edit page/form. Delete requires confirmation dialog.
- **D-12:** Tag filter bar + search box at top of map view. Filter pins by tag selection, search by name.
- **D-13:** New top-level "地图" tab in the app navigation bar alongside "博客" and "债务". Uses the existing AppLayout pattern.

### Agent's Discretion
- Coordinate storage format (store WGS-84 from browser GPS, convert to GCJ-02 for Gaode JS API display on the frontend)
- Proto message design and API endpoint structure for the spot domain
- Database table schema for spots and photo associations
- Gaode API key configuration approach (env var or config)
- Pin clustering behavior when many spots exist on the map
- GPS failure/denied permission error handling
- Bottom sheet component implementation (custom or library)
- Whether to implement coordinate conversion on frontend or backend

### Frontend Resumption (Plan 13-03) — gathered 2026-07-06

The gray-area questions below were unresolved at plan-13-01/13-02 time and are now locked for Plan 13-03.

- **D-14 (AppLayout nav expansion):** Add "地图" as the 6th entry at the end of the existing `navItems` array in `price_recorder_vue/src/components/AppLayout.vue` — `{ name: 'map', path: '/map', label: '地图' }`. Reuse the existing `router-link` active-state pattern. Do NOT collapse other tabs into a "更多" dropdown; do NOT replace "版本". The prior D-13 wording ("alongside 博客/债务") referenced the original 3-tab layout and is superseded by current code reality (5 tabs).
- **D-15 (Capture form layout):** The spot-capture form is rendered as a bottom sheet sliding up from the bottom of the map, covering ~60–70% of the screen with the underlying map still visible. The same bottom-sheet component is reused (visually consistent) for spot detail on pin tap (D-10). A custom bottom-sheet is preferred over pulling an external library — reuse the modal/transition pattern already in `DebtDetail.vue` and adapt to bottom-anchored swipe semantics.
- **D-16 (Photo upload UX):** Multi-file picker → parallel `POST /file/v1` per image, each with its own progress bar; completed uploads display as a thumbnail list with delete-x. Do NOT defer upload to spot-submit; do NOT limit to a single main image. Uploads are triggered on file selection, not on form submit, so a network failure during typing still leaves user with a working spot draft. `mapStore` holds the array of successfully uploaded image URLs; only those URLs are sent in `CreateSpotRequest`.
- **D-17 (Capture form field priority):** First screen of the capture bottom sheet shows only the name input + submit button — always visible. notes / tags / photos are hidden behind an expandable chevron (collapsed by default) so a one-handed field save takes two taps. Aligns with D-09 (only name required at capture time) — user can save the spot bare-bones and edit later. The expanded form never auto-expands even if GPS has reverse-geocoded an address; address shows as a small read-only chip below name when available.
- **D-18 (Filter + search overlay):** Tag filter + search is NOT a persistent top bar. Expose via a FAB button (top-right of map) that opens a slide-down overlay containing the search input and tag chip row. Closing the overlay returns the map to full screen. Rationale: the map is the primary surface; mobile screen budget does not afford a permanent 30–40 px chrome bar.
- **D-19 (Filter action semantics):** Selecting a tag chip or submitting search hides markers whose spots don't match — the map viewport (center/zoom) does NOT auto-pan or refit. User retains manual control via standard AMap pan/zoom; if the user wants to fit-to-results, a "fit view" toolbar button (`map.setFitView()`) is exposed separately, NOT auto-triggered.
- **D-20 (Empty / initial state):** When `mapStore` has zero spots (initial state), the map centers on the user's current location via `AMap.Geolocation` plugin (not a hardcoded city). A soft tooltip bubble ("点右上按钮保存当前钓点") sits on the map until first spot save or first FAB open. Fallback (GPS denied or unsupported): center on the user's last-known spot if any, otherwise default to the RESEARCH.md example Guangzhou center `[113.264385, 23.129112]` (GCJ-02 directly — no conversion needed for this static fallback).
- **D-21 (Vitest test scope):** Plan 13-03 ships `src/__tests__/stores/mapStore.spec.ts`, `src/__tests__/api/map.spec.ts`, and `src/__tests__/view/Map.spec.ts`. Store and API tests follow the existing `fuelStore.spec.ts` / `fuel.spec.ts` / `request.spec.ts` patterns (mock fetch, assert endpoint/payload, assert store mutation). For the component spec, mock `window.AMapLoader`, `window.AMap`, `navigator.geolocation`, and `URL.createObjectURL` — exercise the template wiring, store integration, and event handlers without instantiating a real `AMap.Map`. Behaviors that genuinely require a live map instance (`new AMap.Map(dom)` actually attaching to a DOM node, real tile rendering, real touch gestures on markers) are gated with `it.skipIf(typeof AMap.Map !== 'function')` or equivalent `vi.skipIf` and explicitly enumerated in a comment so they are not silently lost. Reuse happy-dom (already configured in `vitest.config.ts`) — do NOT introduce jsdom as a second environment.
- **D-22 (Tests marked jsdom/happy-dom-skip):** Document the skipped test list in `13-03-PLAN.md` `must_haves` and re-route them as manually-verified UAT items in `13-VALIDATION.md` (so the verifier does not re-flag them as gaps). Examples: pin clustering under 50+ pins, touch-pinch zoom on a real device, Geolocation-denied fallback flow on a real phone.
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Planning
- `.planning/ROADMAP.md` — Phase 13 definition and success criteria (MAP-01 through MAP-03)
- `.planning/PROJECT.md` — Project overview, constraints, single-user scope, tech stack decisions
- `.planning/REQUIREMENTS.md` — v1.1 requirements (no MAP entries yet — Phase 13 is a new addition)

### Backend Architecture
- `blog/api/file/v1/file.proto` — Existing file upload protobuf (reuse for spot photo uploads)
- `blog/internal/server/http.go` — HTTP server setup, middleware, JWT auth pattern
- `blog/internal/server/server.go` — Provider set for DI
- `blog/cmd/blog/main.go` — Application bootstrap and config loading

### Backend Shipped in 13-01/13-02 (frontend must target these contracts — DO NOT modify)
- `blog/api/map/v1/map.proto` — Map service API: 6 RPCs (`CreateSpot` / `ListSpots` / `UpdateSpot` / `DeleteSpot` / `GetSpot` / `ReverseGeocode`) and 8 messages incl. `SpotEntity` with all 10 D-04 fields. HTTP paths: `/map/save/v1`, `/map/list/v1`, `/map/update/v1`, `/map/delete/v1`, `/map/get/{id}`, `/map/reverse-geocode/v1`. All JWT-protected (no whitelist entry in `http.go`).
- `blog/internal/biz/map.go` — `Spot` domain struct (10 string-typed fields), `MapRepo` interface, `MapUsecase` with reverse-geocode via Gaode Web API (reads `GAODE_WEB_API_KEY` env). ReverseGeocode HTTP order is **lng then lat** — frontend MUST send `longitude` before `latitude` in the request body to match.
- `blog/internal/service/map.go` — Proto handler; all `s.mu.*` errors silently swallowed and empty replies returned (consistent with `post.go` convention). Frontend should treat empty `data` as an error and surface a toast.
- `blog/internal/data/map.go` — GORM `Spot` model + `mapRepo` implementing the 5 interface methods. Spot ID is a uint autoincrement (gorm.Model); frontend receives it as `id` integer.
- `blog/internal/data/data.go` — Auto-migration registered (`&Spot{}` line 40), provider set wired (`NewMapRepo` line 15).

### Frontend Architecture
- `price_recorder_vue/src/router/index.ts` — Vue Router setup, route guards, existing route patterns; the children array under `/` (AppLayout) is the mount point for the new `/map` route
- `price_recorder_vue/src/utils/request.ts` — Axios instance with JWT interceptor
- `price_recorder_vue/src/App.vue` — App shell (renders `<router-view>` only; nav lives in AppLayout)
- `price_recorder_vue/src/components/AppLayout.vue` — Current 5-tab nav (`navItems` array at lines 12–18); append `{ name: 'map', path: '/map', label: '地图' }` as the 6th entry (per D-14)
- `price_recorder_vue/src/view/DebtDetail.vue` — Existing modal-with-transition pattern; reference for the spot detail / capture bottom sheet (D-10, D-15)
- `price_recorder_vue/src/stores/fuelStore.ts` — Closest Pinia-store analog to mapStore (list/detail/create/update/delete actions); mirror its action signatures for `mapStore.ts`
- `price_recorder_vue/src/api/fuel.ts` — Closest axios-client analog; mirror its function shape (separate exported functions per backend RPC) for `map.ts`
- `price_recorder_vue/src/__tests__/stores/fuelStore.spec.ts`, `price_recorder_vue/src/__tests__/api/fuel.spec.ts`, `price_recorder_vue/src/__tests__/utils/request.spec.ts` — Test patterns to mirror for `mapStore.spec.ts`, `map.spec.ts`, and the `Map.vue` component spec (per D-21)
- `price_recorder_vue/vite.config.ts` — Vite config with `@` alias; env-var injection point for `VITE_GAODE_JS_API_KEY` + `VITE_GAODE_JS_SECURITY_CODE`

### API Conventions
- `blog/api/post/v1/post.proto` — Example proto contract (follow this pattern for map API)
- `blog/api/debt/v1/debt.proto` — Another API contract example
- `blog/openapi.yaml` — Generated HTTP contract

### Codebase Maps
- `.planning/codebase/ARCHITECTURE.md` — Layering, data flow, entry points
- `.planning/codebase/STACK.md` — Go 1.24, Kratos v2.8, Vue 3.5, GORM, PostgreSQL
- `.planning/codebase/CONVENTIONS.md` — Backend layering, proto-first, frontend patterns
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **File upload infrastructure** (`blog/api/file/v1/file.proto`): Existing multipart file upload API — reuse for spot photo uploads.
- **Auth system** (`blog/internal/server/http.go` + `price_recorder_vue/src/stores/userStore.ts`): JWT login/logout, route guards, Axios interceptor — all spots are user-scoped.
- **Navigation shell** (`price_recorder_vue/src/App.vue`): Existing top-tab layout with "博客" and "债务" — add "地图" as third tab.
- **Toast notifications** (`price_recorder_vue/src/view/`): Existing toast pattern for CRUD success/error — reuse for spot save/delete feedback.

### Established Patterns
- **Proto-first API**: Define `api/map/v1/map.proto`, run `make api`, implement service handler in `blog/internal/service/`.
- **Kratos layering**: New `map` domain follows `server → service → biz → data` with Wire DI.
- **Vue SFC + Pinia**: New `src/view/Map.vue` + `src/api/map.ts` + `src/stores/mapStore.ts`.
- **Modal + Toast**: Existing debt detail uses modals with transition animations and toast feedback — follow for spot edit page.
- **Chinese UI**: Toast messages, labels, and UI text are in Chinese, consistent with blog/debt modules.

### Integration Points
- **New proto domain**: `blog/api/map/v1/` — define Spot, CreateSpot, UpdateSpot, DeleteSpot, ListSpots, GetSpot messages and service.
- **New backend layers**: `blog/internal/biz/map.go` (repo interface + use case), `blog/internal/data/map.go` (GORM implementation), `blog/internal/service/map.go` (proto handler).
- **New frontend module**: `price_recorder_vue/src/view/Map.vue` (main map view), `price_recorder_vue/src/api/map.ts` (API client), `price_recorder_vue/src/stores/mapStore.ts` (Pinia store).
- **Nav tab addition**: `price_recorder_vue/src/App.vue` — add "地图" tab with `router-link`.
- **Route registration**: `price_recorder_vue/src/router/index.ts` — add `/map` route (auth-guarded).
- **Gaode JS API**: Needs to be loaded in `index.html` or via dynamic import. Requires Gaode API key (JS API + Geocoding API).
- **Wire providers**: `blog/internal/data/data.go` (Data struct), `blog/internal/biz/biz.go` (biz provider set), `blog/internal/service/service.go` (service provider set), `blog/cmd/blog/wire.go` (wire assembly).
</code_context>

<specifics>
## Specific Ideas

- User is in China — GPS accuracy and GCJ-02 coordinate handling are critical. Browser Geolocation API returns WGS-84, which must be converted to GCJ-02 for correct display on Gaode Maps.
- Primary use case is fishing — spots will be near water (rivers, lakes, reservoirs), often in rural areas with varying GPS signal quality.
- Mobile-first — the capture flow must work well in the field with one hand. Big touch targets, minimal typing.
- Reverse geocoding gives context: knowing the spot is at "XX河" is more useful than just coordinates.
- Gaode deep-link navigation format: `androidamap://navi?sourceApplication=blog&lat=X&lon=Y&dev=0` or `iosamap://navi?sourceApplication=blog&lat=X&lon=Y&dev=0`.

### Frontend Resumption Specifics (2026-07-06)
- The two Gaode API keys are distinct and MUST NOT be cross-used: `GAODE_JS_API_KEY` + `GAODE_JS_SECURITY_CODE` (frontend-exposed, "Web端 JS API" type, set via `VITE_*` env vars so Vite injects them at build / dev time) vs `GAODE_WEB_API_KEY` (server-side only, "Web服务" type, already wired in `blog/internal/biz/map.go`). The frontend MUST NOT embed `GAODE_WEB_API_KEY` — reverse geocode goes through the backend endpoint, NEVER called directly from the browser.
- AMap key domain whitelist must include the dev origin (commonly `http://localhost:5173` for Vite dev) and production origin in the Gaode console, or the map returns blank tiles.
- The capture FAB (capture button) and the filter FAB are SEPARATE buttons — capture lives on the bottom edge of the map (above the bottom-sheet region when collapsed) with a "+" icon; filter lives top-right. Don't merge them into one FAB menu (D-15 bottom sheet vs D-18 filter overlay use different anchors).
- Plan 13-03 should register MAP-01, MAP-02, MAP-03 in `REQUIREMENTS.md` as an ACTIVE-section row pointing to Phase 13 — VERIFICATION flagged this as a broken traceability hop and the planner should close it as part of this plan.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 13-Fishing Spot Map*
*Context gathered: 2026-07-05*
