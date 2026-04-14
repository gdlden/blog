# Phase 8: API Response Standardization - Research

**Researched:** 2026-04-14
**Domain:** Kratos HTTP transport encoding, Go backend, Vue 3 + Axios frontend
**Confidence:** HIGH

## Summary

This phase unifies all backend HTTP responses under a consistent `{code, message, data}` wrapper via Kratos transport-layer encoding, and aligns the frontend Axios interceptor to parse it. Research confirms Kratos v2.8.0 provides `http.ResponseEncoder()` and `http.ErrorEncoder()` as `ServerOption` hooks that can globally override how success and error responses are serialized to HTTP. This is the idiomatic mechanism — individual handlers continue returning protobuf replies directly, and the encoder wraps them transparently.

The frontend currently has a mixed parsing landscape: the Axios interceptor in `src/utils/request.ts` already handles `code`/`msg` for login (returning `res.data` when `code === "200"`), but all other API clients (`blog.ts`, `debt.ts`, `Article.ts`) expect `res.data` to be the direct payload. A big-bang update (per D-07/D-08) requires changing the interceptor to unwrap `response.data.data` for successes and updating every API client and store to consume the unwrapped payload. gRPC can remain raw because the encoder is HTTP-specific.

**Primary recommendation:** Implement a custom `EncodeResponseFunc` and `EncodeErrorFunc` in `blog/internal/server/http.go`, register them via `http.ResponseEncoder` and `http.ErrorEncoder` server options, and update all frontend `src/api/*.ts` clients to expect the wrapper while making the interceptor the single unwrapping point.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Response wrapping | API / Backend (Kratos HTTP transport) | — | Kratos `ResponseEncoder` is a transport-layer concern; handlers stay agnostic |
| Error code mapping | API / Backend | — | Business errors map to custom codes in the error encoder |
| Response unwrapping | Browser / Client (Axios interceptor) | — | Interceptor is the single frontend gate for parsing the wrapper |
| API client alignment | Browser / Client (`src/api/*.ts`) | — | Clients must expect unwrapped payloads after interceptor processing |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Kratos | v2.8.0 | Go microservices framework | Verified module version; provides `http.ResponseEncoder`/`ErrorEncoder` hooks [VERIFIED: `go list -m github.com/go-kratos/kratos/v2`] |
| Axios | ^1.13.3 | Frontend HTTP client | Already used project-wide for API calls [VERIFIED: `price_recorder_vue/package.json`] |
| Vue | ^3.5.26 | Frontend framework | Existing SPA stack [VERIFIED: `price_recorder_vue/package.json`] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/go-kratos/kratos/v2/errors` | v2.8.0 | Structured error creation (`errors.New`, `errors.BadRequest`, etc.) | Map Go errors to wrapper `code`/`message` in custom error encoder |
| `github.com/go-kratos/kratos/v2/transport/http` | v2.8.0 | `CodecForRequest`, `DefaultResponseEncoder`, `DefaultErrorEncoder` | Reference implementations for custom encoder |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Kratos `ResponseEncoder` | Middleware that wraps at handler level | Violates D-04 (handlers should not manually construct wrapper); more boilerplate per handler |
| Big-bang frontend update | Gradual endpoint-by-endpoint migration | Violates D-07/D-08; requires maintaining dual parsing logic |

**Installation:** No new dependencies required — all libraries are already in `go.mod` and `package.json`.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────┐     HTTP      ┌─────────────────────────────────────────┐
│  Vue Frontend   │ ─────────────▶│  Kratos HTTP Server                     │
│                 │               │  ┌─────────────────────────────────────┐│
│  Axios          │               │  │ Transport Layer                     ││
│  ├─ interceptor │               │  │ ┌─────────────┐   ┌─────────────┐   ││
│  │  (unwraps    │               │  │ │  Handler    │   │  Custom     │   ││
│  │   data)       │               │  │ │  (protobuf  │──▶│  Response   │   ││
│  │               │               │  │ │  reply)     │   │  Encoder    │   ││
│  └─ api clients │               │  │ └─────────────┘   └──────┬──────┘   ││
│     expect      │               │  │                          │          ││
│     payload     │               │  │ ┌─────────────┐   ┌──────▼──────┐   ││
│                 │◀──────────────│  │ │  Error      │──▶│  Custom     │   ││
│                 │  {code,msg,data}│ │ │  (kratos)   │   │  Error      │   ││
│                 │               │  │ │             │   │  Encoder    │   ││
│                 │               │  │ └─────────────┘   └─────────────┘   ││
│                 │               │  └─────────────────────────────────────┘│
│                 │               └─────────────────────────────────────────┘
│                 │                              │
│                 │                              │ gRPC (raw, unchanged)
│                 │                              ▼
│                 │               ┌─────────────────────────────────────────┐
│                 │               │  Kratos gRPC Server                     │
│                 │               │  (no encoder changes)                   │
└─────────────────┘               └─────────────────────────────────────────┘
```

### Recommended Project Structure

Backend changes:
```
blog/internal/server/
├── http.go              # Add custom response/error encoder functions
└── response.go          # [optional] Wrapper struct and code constants
```

Frontend changes:
```
price_recorder_vue/src/
├── utils/
│   └── request.ts       # Update interceptor to unwrap wrapper
├── api/
│   ├── blog.ts          # Update to expect payload directly from interceptor
│   ├── debt.ts          # Update to expect payload directly from interceptor
│   ├── Login.ts         # Update login parsing (remove manual res handling)
│   └── Article.ts       # Update to expect payload directly
└── stores/
    ├── blogStore.ts     # Update response.data access
    └── debtStore.ts     # Update response.list / response.total access
```

### Pattern 1: Custom Kratos Response Encoder
**What:** Override `DefaultResponseEncoder` to wrap every non-error HTTP response in `{code, message, data}`.
**When to use:** When you need a uniform envelope for all success responses without touching handlers.
**Example:**
```go
// Source: Kratos v2.8.0 transport/http/codec.go + project conventions
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func CustomResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(Response{Code: 200, Message: "success", Data: nil})
	}
	// Preserve redirects if any
	if rd, ok := v.(http.Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return nil
	}
	codec, _ := http.CodecForRequest(r, "Accept")
	w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
	return codec.Marshal(w, Response{Code: 200, Message: "success", Data: v})
}
```

### Pattern 2: Custom Kratos Error Encoder
**What:** Override `DefaultErrorEncoder` to map `*kratos/errors.Error` into the same `{code, message, data}` envelope.
**When to use:** When you want error responses to share the same shape as success responses.
**Example:**
```go
// Source: Kratos v2.8.0 transport/http/codec.go + project conventions
func CustomErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := kratoserrors.FromError(err)
	// Map HTTP/gRPC status to business code, or use a custom code field
	code := int(se.Code)
	if code == 0 {
		code = 500
	}
	// For business-level codes, you may prefer a lookup table:
	// e.g., 400 -> 1001 (param error), 404 -> 1002 (not found)
	codec, _ := http.CodecForRequest(r, "Accept")
	body, _ := codec.Marshal(Response{
		Code:    code,
		Message: se.Message,
		Data:    nil,
	})
	w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
	w.WriteHeader(code)
	w.Write(body)
}
```

### Pattern 3: Axios Interceptor Unwrapping
**What:** Make the response interceptor return `res.data.data` on success and `Promise.reject({code, message})` on error.
**When to use:** Centralized unwrapping so every API client gets the payload directly.
**Example:**
```typescript
// Source: project conventions + Axios docs
instance.interceptors.response.use(
  (res) => {
    const body = res.data;
    if (body && body.code === 200) {
      return body.data;
    }
    return Promise.reject(new Error(body.message || "请求失败"));
  },
  (err) => {
    const msg = err.response?.data?.message || err.message || "网络错误";
    return Promise.reject(new Error(msg));
  }
);
```

### Anti-Patterns to Avoid
- **Handler-level wrapping:** Returning `Wrapper{Code: 200, Data: reply}` from individual service handlers violates D-04 and duplicates logic.
- **Partial frontend update:** Updating only some API clients to expect the wrapper while leaving others will break the big-bang guarantee (D-08).
- **String `code` in wrapper:** D-02 specifies numeric `code`; keeping string codes (like existing `LoginReply.Code`) in the new wrapper would violate the decision.
- **Wrapping gRPC:** gRPC uses protobuf status codes natively; applying the HTTP JSON wrapper to gRPC is unnecessary and non-idiomatic.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Per-handler response wrapping | Manual wrapper construction in every `service/*.go` | Kratos `ResponseEncoder` | Single point of change, zero handler boilerplate |
| Error code extraction from wrapped errors | Custom error unwrapping logic | `kratoserrors.FromError(err)` | Handles wrapped errors, extracts `*errors.Error` automatically |
| JSON marshaling per endpoint | `json.Marshal` in every handler | `http.CodecForRequest(r, "Accept")` | Respects content negotiation, supports proto-JSON consistently |

**Key insight:** Kratos already solved transport-layer encoding. Hand-rolling it in handlers is a common trap that leads to inconsistent wrappers and makes future changes expensive.

## Runtime State Inventory

> This phase involves a code-only change with no persistent data renames or migrations.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — no database schema or stored keys change | None |
| Live service config | None — no external UI-managed configs reference response shapes | None |
| OS-registered state | None — no OS registrations depend on response format | None |
| Secrets/env vars | None — no secrets keyed by response structure | None |
| Build artifacts | None — response shape is runtime behavior, not build artifact | None |

**Nothing found in category:** All categories verified as none.

## Common Pitfalls

### Pitfall 1: Existing `LoginReply` conflicts with the new wrapper
**What goes wrong:** `LoginReply` already has `code` and `msg` fields. After the encoder wraps it, the frontend will see `{code: 200, message: "success", data: {msg: "...", code: "200", token: "...", user: ...}}`. The login page currently checks `response.token` directly because the interceptor returns `res.data` (which today is the raw `LoginReply` for that endpoint, since its `code` is `"200"`). After the change, the interceptor will return `res.data.data`, so `response` becomes the inner `LoginReply` object — but `Login.vue` expects `response.token`, which is still present on the inner object. However, `Login.ts` currently returns `res` (the Axios response object) instead of `res.data`, so `Login.vue` accesses `response.token` from the Axios response wrapper, which will break.

**Why it happens:** `Login.ts` does `.then(res => res)` instead of `.then(res => res.data)`, and `Login.vue` reads `response.token`. After the interceptor change, `res.data` becomes the wrapper, and the interceptor returns `res.data.data`. If `Login.ts` bypasses the interceptor pattern, it will break.

**How to avoid:** Update `Login.ts` to remove the custom `.then(res => res)` and let the interceptor return the unwrapped payload. Update `Login.vue` to expect the unwrapped `LoginReply` shape. Also update `Login.spec.ts` mocks.

**Warning signs:** Login page fails with `token is undefined` after backend encoder is deployed.

### Pitfall 2: `blogStore.ts` accesses `response.data`
**What goes wrong:** `blogStore.fetchPosts` does `posts.value = response.data || []`. After the API change, `getPosts` will return the unwrapped `PostPageReply` directly (via interceptor), so `response.data` will be `undefined` and the store will clear the post list.

**Why it happens:** `PostPageReply` already has a field named `data` (the list of posts). The store confuses the wrapper's `data` with the reply's `data`. After unwrapping, the returned object IS the `PostPageReply`, so the correct access is `response.data` (the reply field) — wait, no: after unwrapping, the interceptor returns `res.data.data` (the inner `PostPageReply`). The API client `getPosts` returns that inner object. So `response` in the store will be the `PostPageReply`, which has a `data` field. Therefore `response.data` is still correct. BUT the store also accesses `response.total`, which exists on `PostPageReply`. So `blogStore.ts` may actually work unchanged IF the API client returns the unwrapped reply.

Let me re-verify: `blogStore.fetchPosts` calls `blogApi.getPosts(...)`, which currently does `.then((res) => res.data)`. After interceptor change, the interceptor returns `res.data.data` (the unwrapped `PostPageReply`). Then `getPosts` returns that. So `response` in the store is the `PostPageReply`. `response.data` is the post list field inside `PostPageReply`. `response.total` is the total field. So `blogStore.ts` does NOT need changes for `fetchPosts`.

However, `debtStore.ts` calls `debtApi.getDebts(...)`, which returns `res.data`. After interceptor change, it returns the unwrapped `ListDebtReply`. `ListDebtReply` has fields `list` and `total`. The store does `debts.value = response.list || []` and `total.value = parseInt(response.total || "0", 10)`. So `debtStore.ts` also does NOT need changes.

The real frontend changes are in the API clients (`blog.ts`, `debt.ts`, `Login.ts`, `Article.ts`) because they currently do `.then((res) => res.data)`. After the interceptor unwraps, they should simply `return await instance.get(...)` without the `.then` (or with `.then(res => res)` if they want, but the interceptor already returns the payload). Actually, if the interceptor returns `res.data.data`, then `instance.get(...)` resolves to the payload directly. So the `.then((res) => res.data)` in the API clients will try to access `.data` on the payload, which is wrong for most replies.

Wait, let me trace carefully:
- Today: Axios gets HTTP body `{id: "1", title: "..."}` (for AddPostReply). `res.data` = that object. API client does `.then(res => res.data)` and returns it. Store gets the object.
- After encoder: Axios gets HTTP body `{code: 200, message: "success", data: {id: "1", title: "..."}}`. `res.data` = the wrapper. Interceptor returns `res.data.data` = `{id: "1", title: "..."}`. API client does `.then(res => res.data)` where `res` is what the interceptor returned. But the interceptor's return value becomes the resolved value of the promise. So inside `.then(res => ...)`, `res` is `{id: "1", title: "..."}`. Then `res.data` is `undefined` (unless the reply happens to have a `data` field, like `PostPageReply`).

**Conclusion:** All API clients that do `.then((res) => res.data)` must be changed. For replies that DON'T have a `data` field, `res.data` will be undefined. For `PostPageReply` which DOES have a `data` field, it would accidentally work but return the post list instead of the whole reply object — which would break `response.total` access in the store.

Actually wait: `blogApi.getPosts` does `.then((res) => res.data)`. After interceptor, `res` is the `PostPageReply` object. `res.data` is the post list array. The store expects `response.data` to be the list and `response.total` to be the total. But if `getPosts` returns `res.data` (the list), then `blogStore.fetchPosts` gets an array, and `response.data` is undefined, `response.total` is undefined. This WILL break.

So ALL API clients must remove the `.then((res) => res.data)` pattern and simply return `instance.get/post(...)` directly, letting the interceptor do the unwrapping.

Similarly, `debtApi.deleteDebt` does `return response.data.flag;`. After interceptor, `response` is the `DeleteDebtReply`. So it should become `return response.flag;`.

And `api/Article.ts` does `.then(res => res.data)` — must be removed.

And `api/Login.ts` does `.then(res => res)` — this returns the raw Axios response. After interceptor, `res` is the unwrapped `LoginReply`. But `Login.vue` does `response.token`. If `login()` returns the unwrapped `LoginReply`, then `response.token` works. However, `Login.vue` also checks `if (response && response.token)`. That still works. BUT `Login.ts` has a `.catch(err => { console.log(err); })` which swallows errors and returns `undefined`. That catch should probably be removed so errors propagate.

**How to avoid:** Systematically update every `src/api/*.ts` to remove manual `res.data` extraction and rely on the interceptor. Update `Login.vue` and `Login.spec.ts` to work with the unwrapped reply.

### Pitfall 3: Numeric `code` vs existing string `code`
**What goes wrong:** `LoginReply.Code` is a `string` (`"200"`). The new wrapper uses numeric `code` (`200`). The frontend interceptor currently checks `res.data.code === "200"`. After the change, it must check against the number `200`.

**How to avoid:** Update the interceptor to use numeric comparison (`=== 200`) or coerce both sides (`Number(body.code) === 200`).

### Pitfall 4: Error encoder breaks HTTP status codes
**What goes wrong:** If the custom error encoder writes the business code as the HTTP status (e.g., `w.WriteHeader(1001)`), proxies/browsers may treat it as an invalid HTTP status.

**How to avoid:** Keep HTTP status as a standard code (e.g., 400 or 500) and put the business code inside the JSON body. Or if you must align them, ensure the code is a valid HTTP status. D-05/D-06 suggest custom business codes like `1001`, so they must live in the JSON body, not the HTTP status line.

### Pitfall 5: gRPC accidentally affected
**What goes wrong:** If the wrapper struct or encoder is placed in a shared package and accidentally used for gRPC replies.

**How to avoid:** Register the custom encoder only on the HTTP server in `internal/server/http.go`. Leave `internal/server/grpc.go` untouched.

## Code Examples

### Backend: Register custom encoders in `http.go`
```go
// Source: Kratos v2.8.0 docs + verified source code
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService,
    postService *service.PostService, userService *service.UserService,
    ocrService *service.AiocrService, priceService *service.PriceService,
    debtService *service.DebtService, detailService *service.DebtDetailService,
    logger log.Logger) *http.Server {
    var opts = []http.ServerOption{
        http.ResponseEncoder(CustomResponseEncoder),
        http.ErrorEncoder(CustomErrorEncoder),
        http.Middleware(
            recovery.Recovery(),
            tracing.Server(),
            logging.Server(logger),
            selector.Server(
                jwt.Server(func(token *jwtv5.Token) (interface{}, error) {
                    return []byte("dfsdsjikldsfkdfjdkls"), nil
                }),
            ).Match(NewWhiteListMatcher()).Build(),
        ),
    }
    // ... rest unchanged
}
```

### Backend: Wrapper struct and encoder
```go
// Source: project conventions + Kratos patterns
type UnifiedResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func CustomResponseEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, v interface{}) error {
    if v == nil {
        w.Header().Set("Content-Type", "application/json")
        return json.NewEncoder(w).Encode(UnifiedResponse{Code: 200, Message: "success"})
    }
    if rd, ok := v.(http.Redirector); ok {
        url, code := rd.Redirect()
        stdhttp.Redirect(w, r, url, code)
        return nil
    }
    codec, _ := http.CodecForRequest(r, "Accept")
    w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
    return codec.Marshal(w, UnifiedResponse{Code: 200, Message: "success", Data: v})
}

func CustomErrorEncoder(w stdhttp.ResponseWriter, r *stdhttp.Request, err error) {
    se := kratoserrors.FromError(err)
    businessCode := mapHTTPToBusinessCode(int(se.Code))
    codec, _ := http.CodecForRequest(r, "Accept")
    body, _ := codec.Marshal(UnifiedResponse{
        Code:    businessCode,
        Message: se.Message,
    })
    w.Header().Set("Content-Type", httputil.ContentType(codec.Name()))
    // Keep HTTP status as a valid code to avoid proxy issues
    httpStatus := int(se.Code)
    if httpStatus < 100 || httpStatus > 599 {
        httpStatus = 500
    }
    w.WriteHeader(httpStatus)
    w.Write(body)
}

func mapHTTPToBusinessCode(httpCode int) int {
    switch httpCode {
    case 400:
        return 1001 // param error
    case 401:
        return 1002 // unauthorized
    case 404:
        return 1003 // not found
    case 500:
        return 1004 // internal error
    default:
        return httpCode
    }
}
```

### Frontend: Updated Axios interceptor
```typescript
// Source: project conventions + Axios interceptor docs
instance.interceptors.response.use(
  (res) => {
    const body = res.data as { code?: number; message?: string; data?: any } | undefined;
    if (!body || body.code === 200) {
      return body?.data;
    }
    return Promise.reject(new Error(body.message || "请求失败"));
  },
  (err) => {
    const msg = err.response?.data?.message || err.message || "网络错误";
    alert(msg);
    return Promise.reject(new Error(msg));
  }
);
```

### Frontend: Updated API client (`blog.ts`)
```typescript
export async function getPosts(
  current?: string,
  size?: string
): Promise<PostPageResponse> {
  const params: Record<string, string> = {};
  if (current) params.current = current;
  if (size) params.size = size;
  return await instance.get("/post/page/v1", { params });
}

export async function getPostById(id: string): Promise<Post> {
  return await instance.get(`/post/get/${id}`);
}

export async function createPost(data: { title: string; content: string }): Promise<Post> {
  return await instance.post("/post/add/v1", data);
}

export async function updatePost(
  id: string,
  data: { title: string; content: string }
): Promise<Post> {
  return await instance.post("/post/edit/v1", { id, ...data });
}

export async function deletePost(id: string): Promise<void> {
  await instance.post("/post/delete/v1", { id });
}
```

### Frontend: Updated `Login.ts`
```typescript
export async function login(data: userReq) {
  return await instance.post("/user/login/v1", data);
}
```

### Frontend: Updated `Login.vue` usage
```typescript
const response = await login(user);
if (response && response.token) {
  userStore.setUserInfo(response);
  // ...
}
```
(Note: `response` is now the unwrapped `LoginReply` object with `token`, `user`, etc.)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Raw protobuf replies over HTTP | Transport-layer wrapper via `ResponseEncoder` | This phase | Unified contract, no handler changes |
| Frontend `res.data` direct access | Interceptor unwrapping `res.data.data` | This phase | Single parsing gate, consistent error handling |
| `LoginReply` string `code` | Wrapper numeric `code` + inner `LoginReply` unchanged | This phase | `LoginReply` can remain as-is; wrapper provides the standard envelope |

**Deprecated/outdated:**
- Manual wrapper construction in service handlers: violates D-04 and is unnecessary given Kratos encoder hooks.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | gRPC responses should remain raw (not wrapped) because the encoder is HTTP-specific and gRPC uses its own status mechanism | Architecture Patterns | If gRPC clients also need the wrapper, additional research into gRPC interceptors/middleware is required |
| A2 | Business code mapping (e.g., 400 -> 1001) is acceptable as a simple switch table in the error encoder | Code Examples | If the user wants a more elaborate mapping strategy (e.g., reason-based), the encoder logic will need expansion |
| A3 | `Article.vue` and `Article.ts` are still in use despite not being linked from the main router | Code Examples | If they are dead code, updating them is harmless; if they are used by direct navigation, they must be updated |

## Open Questions

1. **Should we also update `LoginReply` to use `message` instead of `msg`?**
   - What we know: The new standard wrapper uses `message`. `LoginReply` currently uses `msg`.
   - What's unclear: Whether to modify the protobuf (which triggers `make api` and regenerates bindings) or leave `LoginReply` as-is since the wrapper already provides `message` at the envelope level.
   - Recommendation: Leave `LoginReply` unchanged. The wrapper provides the standard `message` field for all endpoints. The inner `LoginReply.msg` becomes an implementation detail. This avoids proto regeneration churn.

2. **What should the exact business code enum be?**
   - What we know: D-06 gives examples (`1001` = param error, `1002` = not found).
   - What's unclear: Whether there is an existing enum or if we should invent a small mapping table.
   - Recommendation: Create a minimal mapping in the error encoder (1001=bad request, 1002=unauthorized, 1003=not found, 1004=internal) and document it in code. This is within Claude's discretion per D-05/D-06.

3. **Do any other frontend files import `res.data` directly?**
   - What we know: Grep found `res.data` usages in `request.ts`, `api/*.ts`, `view/Article.vue`, and `stores/blogStore.ts` (`response.data`).
   - What's unclear: Whether there are dynamic imports or lazy-loaded modules not caught by static grep.
   - Recommendation: After updating the known files, run `pnpm build` and `go test ./...` to catch any remaining mismatches.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend build/tests | ✓ | 1.24+ | — |
| Kratos | Backend encoder | ✓ | v2.8.0 | — |
| Node.js | Frontend build/tests | ✓ | v22.21.1 | — |
| Vitest | Frontend unit tests | ✓ | v4.0.17 | — |

**Missing dependencies with no fallback:** None.

**Missing dependencies with fallback:** None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest v4.0.17 |
| Config file | `vitest.config.ts` (implied by working `npx vitest run`) |
| Quick run command | `cd price_recorder_vue && ./node_modules/.bin/vitest run` |
| Full suite command | Same as quick run (19 tests across 6 files) |

Backend:
| Property | Value |
|----------|-------|
| Framework | Go testing |
| Quick run command | `cd blog && go test ./internal/...` |
| Full suite command | `cd blog && go test ./...` (note: `cmd/blog/main.go` has a `fmt.Printf` lint failure unrelated to this phase) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| API-01 | All HTTP endpoints return `{code, message, data}` | integration/smoke | Manual curl or new Go HTTP test | ❌ Need to add `internal/server/http_test.go` |
| API-02 | Success responses have `code: 200` and payload in `data` | integration | Same as above | ❌ |
| API-03 | Error responses include non-success code and message | integration | Same as above | ❌ |
| API-04 | Wrapper applied at transport layer | code review / unit | Verify `http.ResponseEncoder` is registered in `http.go` | ❌ |
| API-05 | Frontend interceptor parses unified wrapper | unit | `cd price_recorder_vue && ./node_modules/.bin/vitest run` (update + run existing) | ✅ Existing tests pass; need to update mocks |

### Sampling Rate
- **Per task commit:** `cd blog && go test ./internal/...` and `cd price_recorder_vue && ./node_modules/.bin/vitest run`
- **Per wave merge:** Same commands
- **Phase gate:** Both green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `blog/internal/server/http_test.go` — covers API-01 through API-04 with a lightweight HTTP round-trip test
- [ ] Update `price_recorder_vue/src/__tests__/view/Login.spec.ts` mocks to return unwrapped `LoginReply` shape

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No changes to auth mechanism |
| V3 Session Management | No | No changes to session handling |
| V4 Access Control | No | No changes to access control |
| V5 Input Validation | Yes | Existing Kratos request decoding + biz-layer validation remains |
| V6 Cryptography | No | No crypto changes |

### Known Threat Patterns for Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Information disclosure via error messages | Information Disclosure | Error encoder should not leak internal stack traces; use `se.Message` or a sanitized message |
| HTTP status injection | Tampering | Error encoder validates HTTP status is in 100-599 range before calling `w.WriteHeader` |

## Sources

### Primary (HIGH confidence)
- Kratos v2.8.0 source: `/Users/hukss/go/pkg/mod/github.com/go-kratos/kratos/v2@v2.8.0/transport/http/codec.go` — `DefaultResponseEncoder`, `DefaultErrorEncoder`, `CodecForRequest`
- Kratos v2.8.0 source: `/Users/hukss/go/pkg/mod/github.com/go-kratos/kratos/v2@v2.8.0/transport/http/server.go` — `ResponseEncoder`, `ErrorEncoder` server options
- Kratos v2.8.0 source: `/Users/hukss/go/pkg/mod/github.com/go-kratos/kratos/v2@v2.8.0/errors/errors.pb.go` — `Status` struct with `Code`, `Reason`, `Message`
- Project code: `blog/internal/server/http.go`, `blog/internal/service/user.go`, `price_recorder_vue/src/utils/request.ts`, `price_recorder_vue/src/api/*.ts`

### Secondary (MEDIUM confidence)
- Kratos documentation patterns for custom encoders (derived from source code analysis and community conventions)
- Axios interceptor behavior verified by running existing Vitest suite (`19 passed`)

### Tertiary (LOW confidence)
- None — all claims were verified against source code or running tests.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified Kratos version and APIs via `go doc` and source code
- Architecture: HIGH — encoder hooks are the idiomatic Kratos mechanism
- Pitfalls: HIGH — traced every frontend `res.data` usage and identified exact breakage points

**Research date:** 2026-04-14
**Valid until:** 2026-05-14 (Kratos is stable; only risk is minor project code drift)
