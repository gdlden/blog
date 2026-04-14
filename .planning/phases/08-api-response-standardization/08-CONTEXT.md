# Phase 8: API Response Standardization - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Unify all backend HTTP responses under a consistent `{code, message, data}` wrapper via Kratos response encoding, and align the frontend Axios interceptor.

</domain>

<decisions>
## Implementation Decisions

### Wrapper structure
- **D-01:** Response wrapper uses `{ code: number, message: string, data: any }`
- **D-02:** `code` is numeric, not string

### Implementation approach
- **D-03:** Wrapper is applied via Kratos HTTP response encoder at the transport layer
- **D-04:** Individual handlers should NOT manually construct the wrapper

### Error code strategy
- **D-05:** Use custom business codes rather than raw HTTP status codes
- **D-06:** `200` means success; other codes indicate specific failures (e.g., `1001` = param error, `1002` = not found)

### Frontend transition
- **D-07:** Big-bang update — frontend Axios interceptor will immediately expect the unified wrapper on all responses
- **D-08:** All backend endpoints in this phase must return the wrapped format together

### Claude's Discretion
- Exact business code enum values and mapping strategy
- Whether to also wrap gRPC responses or only HTTP
- Specific error message formatting

</decisions>

<specifics>
## Specific Ideas

- Kratos response encoder can be overridden in `blog/internal/server/http.go` or via server options
- Frontend interceptor lives in `price_recorder_vue/src/utils/request.ts`
- Existing `LoginReply` already has `code` and `msg` fields, but the new standard uses `message`

</specifics>

<canonical_refs>
## Canonical References

### Backend
- `blog/internal/server/http.go` — HTTP server setup where encoder can be wired
- `blog/internal/service/user.go:151` — Existing `LoginReply` with `Code: "200"` pattern

### Frontend
- `price_recorder_vue/src/utils/request.ts` — Axios interceptor to update
- `price_recorder_vue/src/api/blog.ts` — Blog API client expectations
- `price_recorder_vue/src/api/debt.ts` — Debt API client expectations

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `blog/internal/server/http.go`: Central HTTP server factory — ideal place to register a custom response encoder
- `price_recorder_vue/src/utils/request.ts`: Central Axios instance — single point of change for frontend parsing

### Established Patterns
- Kratos uses protobuf-first handlers; transport layer wraps them for HTTP
- Frontend already uses `res.data.code === "200"` in the interceptor (from `LoginReply`)
- `blog/internal/service/` handlers return protobuf replies directly today

### Integration Points
- New encoder must wrap protobuf reply messages into `{code, message, data}` without breaking gRPC (if gRPC stays raw)
- All `src/api/*.ts` clients currently expect `.then((res) => res.data)` to be the payload directly
</code_context>

<deferred>
## Deferred Ideas

- None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-api-response-standardization*
*Context gathered: 2026-04-14*
