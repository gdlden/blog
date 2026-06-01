---
name: pr-check
description: PR 提交前检查 — 验证代码变更符合仓库 PR 规范（函数索引、测试、commit 格式等）
disable-model-invocation: true
---

# PR Check

在提交 PR 前全面验证代码变更，确保符合本仓库的所有规范要求。

## 检查流程

### 1. 后端变更检查（如有 blog/ 变更）

```bash
cd blog
go fmt ./...          # 格式化
go build ./...        # 编译
go test ./...         # 测试
```

**函数索引**: 检查 `blog/FUNCTION_INDEX.md` 是否同步更新：
- 新增函数 → 已添加条目
- 修改签名 → 已更新对应条目
- 删除函数 → 已移除条目

### 2. 前端变更检查（如有 price_recorder_vue/ 变更）

```bash
cd price_recorder_vue
pnpm lint            # ESLint
pnpm format          # Prettier
pnpm type-check      # TypeScript 类型检查
pnpm test:unit       # Vitest 测试
pnpm build           # 完整构建
```

### 3. Commit 信息检查

- 格式: `<scope>: <action>`（如 `blog: add debt detail endpoint`）
- 简洁、祈使语气

### 4. PR 完整性检查

PR 描述须包含：
- [ ] 变更摘要
- [ ] 涉及模块/路径 (backend/frontend/proto)
- [ ] 验证依据 (测试命令与结果)
- [ ] UI/契约变更时附截图或请求响应示例
- [ ] 后端变更时确认 `已检查复用并更新 blog/FUNCTION_INDEX.md`

### 5. 输出

汇总所有检查结果，标注通过/未通过项。未通过项给出具体修复建议。
