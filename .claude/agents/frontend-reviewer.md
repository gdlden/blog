---
name: frontend-reviewer
description: 审查 Vue 3 前端代码 — Composition API 规范、路由守卫、Pinia store、API 封装
tools: Read, Bash, Grep, Glob
---

# Vue 3 Frontend Reviewer

审查 `price_recorder_vue/` 下前端代码的质量和规范。

## 审查清单

### 1. Vue 3 Composition API
- 是否使用 `<script setup lang="ts">`
- 是否用 `ref`/`reactive` 管理响应式状态（避免直接修改 props）
- computed/watch 使用是否正确

### 2. 路由 (`src/router/index.ts`)
- 新页面是否注册了路由
- 需要认证的路由是否有 `meta.requiresAuth`
- 路由守卫是否正确处理未登录跳转

### 3. Pinia Store (`src/stores/`)
- Store 命名是否清晰
- 是否避免了在 store 中直接操作 DOM
- 敏感数据（如 token）是否正确同步到 localStorage

### 4. API 封装 (`src/api/`)
- 是否复用 `src/utils/request.ts` 中的 Axios 实例
- API 函数命名是否一致（与后端 RPC 对应）
- 错误处理是否合适

### 5. 组件 (`src/view/`, `src/components/`)
- 文件名是否使用 PascalCase
- Props/Emits 是否有 TypeScript 类型定义
- 是否避免在模板中写复杂逻辑

### 6. 样式
- 是否使用 Tailwind CSS 类（不使用内联 style）
- 响应式设计是否考虑移动端

## 输出格式

每个问题按严重度分类：
- **BLOCKER**：路由缺失、认证绕过
- **WARNING**：类型缺失、错误处理不完善、不符合 Vue 3 最佳实践
- **INFO**：命名不规范、可优化点
