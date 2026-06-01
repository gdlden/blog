---
name: gen-test
description: 为后端 Go 代码或前端 Vue/TS 代码生成符合仓库规范的测试文件
disable-model-invocation: true
---

# Generate Tests

为项目代码生成测试，严格遵循仓库测试规范。

## 后端测试（Go）

```bash
# 先运行现有测试了解模式
cd blog && go test ./... -v -count=1
```

**规则**:
- 测试文件与实现文件相邻放置，命名为 `<name>_test.go`
- 使用标准 `testing` 包 + `github.com/stretchr/testify`（已安装）
- 遵循 `service → biz → data` 分层，mock 下一层依赖
- 参考 `blog/internal/data/*_test.go` 中的现有测试模式

**生成内容**:
- 为每个公开函数生成至少一个正常路径用例
- 为有错误返回的函数生成错误路径用例
- 如涉及数据库，使用测试数据库或 mock

## 前端测试（Vue/TS）

```bash
# 先运行现有测试了解模式
cd price_recorder_vue && pnpm test:unit --run
```

**规则**:
- 测试文件放在 `src/__tests__/` 下，命名为 `<name>.spec.ts`
- 使用 Vitest + `@vue/test-utils`（已安装）
- Vue 组件测试：mount + 交互验证
- API 客户端测试：mock Axios 响应
- Store 测试：直接测试 Pinia store

**生成内容**:
- 组件: 渲染 + 关键交互测试
- API: 请求参数 + 响应处理测试
- Store: 状态变更测试

## 执行后

1. 运行相应测试套件确认通过
2. 如有测试失败，分析原因并修复
