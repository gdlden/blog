---
name: sync-function-index
description: 扫描 blog/internal/ 变更并自动同步 FUNCTION_INDEX.md
---

# Sync Function Index

扫描 `blog/internal/` 目录下 Go 文件的变更，自动更新 `blog/FUNCTION_INDEX.md`。

## 触发条件

当 `blog/internal/` 下任何 `.go` 文件发生变更后，检查是否需要更新索引。

## 扫描规则

### 需要添加到索引的函数
- `blog/internal/biz/` 中导出的 usecase 方法和构造函数
- `blog/internal/data/` 中导出的 repo 构造函数
- `blog/internal/service/` 中导出的 service 构造函数
- 签名包含 `func New*` 或 `func (uc *XxxUsecase)` 模式

### 索引条目格式

```markdown
| `FunctionName` | `package_path/file.go:line` | 简短描述 | 完整签名 | 调用方 | active | YYYY-MM-DD |
```

### 需要更新索引的情况
- **新增函数**：添加新行，status=active
- **删除函数**：标记 status=deprecated
- **修改签名**：更新签名列和 last_updated 日期
- **修改行为**：更新 responsibility 描述

## 执行步骤

1. `git diff --name-only` 查看变更文件
2. 对每个变更的 `.go` 文件，提取导出函数列表
3. 对比 FUNCTION_INDEX.md 现有条目
4. 新增/更新/废弃对应条目
5. 更新文件头部的 last_updated（如有）

## 注意事项
- 不索引测试文件（`*_test.go`）
- 不索引 `internal/model/` 中的纯模型 struct（只索引带方法的）
- `last_updated` 使用当前日期 `YYYY-MM-DD` 格式
