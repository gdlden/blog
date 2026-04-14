---
status: complete
---

# Quick Task 260414-w9n: 都正常保存了怎么还是提示新建失败

## What Changed

Fixed `price_recorder_vue/src/utils/request.ts` Axios response interceptor.

**Before:**
```js
if (res.data.code === "200") {
  return res.data;
}
return Promise.reject(res.data)
```

**After:**
```js
if (!res.data.code || res.data.code === "200") {
  return res.data;
}
return Promise.reject(res.data)
```

## Why

Backend Kratos protobuf endpoints return data directly (e.g., `{id, title, content}`) without a `{code: "200"}` wrapper. The interceptor rejected these successful responses, causing the frontend to show "创建失败" even though the data was saved.

## Verification

- `npm run test:unit` — 19 passed
