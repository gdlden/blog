# Phase 13: Fishing Spot Map - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

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
