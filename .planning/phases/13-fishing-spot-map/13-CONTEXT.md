# Phase 13: Fishing Spot Map - Context

**Gathered:** 2026-07-05
**Status:** Ready for planning

<domain>
## Phase Boundary

Save and manage GPS-tagged fishing spots — capture current coordinates with one tap, name and describe the spot, attach photos and tags, and navigate back to it later via Gaode Maps. Adds a new "地图" top-level tab alongside "博客" and "债务" with a map-primary full-screen view showing all saved spots as pins.

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

### Frontend Architecture
- `price_recorder_vue/src/router/index.ts` — Vue Router setup, route guards, existing route patterns
- `price_recorder_vue/src/utils/request.ts` — Axios instance with JWT interceptor
- `price_recorder_vue/src/App.vue` — App shell with navigation tabs
- `price_recorder_vue/src/stores/userStore.ts` — Pinia auth store pattern
- `price_recorder_vue/vite.config.ts` — Vite config with `@` alias

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

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 13-Fishing Spot Map*
*Context gathered: 2026-07-05*
