# 仓库规范

## 项目结构与模块组织
本仓库包含两个项目：
- `blog/`：基于 Kratos 的 Go 后端服务。入口文件为 `blog/cmd/blog/main.go`。核心分层位于 `blog/internal/{service,biz,data,server}`。API 与契约文件位于 `blog/api/**` 和 `blog/openapi.yaml`。
- `price_recorder_vue/`：基于 Vue 3 + Vite 的前端。应用启动文件为 `price_recorder_vue/src/main.ts`，页面位于 `src/view/`，API 客户端位于 `src/api/`，状态管理位于 `src/stores/`，测试位于 `src/__tests__/`。

生成的 protobuf 依赖以 vendor 方式放在 `blog/third_party/` 下。

## 构建、测试与开发命令
后端（`cd blog`）：
- `make run`：在本地运行 Kratos 服务。
- `make build`：构建二进制文件到 `blog/bin/`。
- `make api && make config`：重新生成 API/内部 protobuf 输出。
- `go test ./...`：运行全部 Go 测试。

前端（`cd price_recorder_vue`）：
- `pnpm install`：安装依赖。
- `pnpm dev`：启动 Vite 开发服务器。
- `pnpm build`：执行类型检查并构建生产资源。
- `pnpm test:unit`：运行 Vitest 单元测试。
- `pnpm lint` / `pnpm format`：对前端源码执行 lint 与格式化。

## 代码风格与命名规范
- Go：提交前使用 `gofmt`（或 `go fmt ./...`）格式化；包名保持小写；遵循清晰的分层边界（`service -> biz -> data`）。
- Vue/TS：遵循 ESLint + Prettier（`eslint.config.ts`），使用 2 空格缩进，组件文件使用 PascalCase 命名（例如 `Article.vue`）。
- API/proto 更新应保持带版本的路径（例如 `api/post/v1/post.proto`）。

## 函数复用与索引维护（仅后端 `blog/`）
- 新增函数前，必须先在后端代码与函数索引中检索可复用函数；若存在可复用实现，优先复用，避免重复实现。
- 后端函数索引文件固定为 `blog/FUNCTION_INDEX.md`。
- 以下变更必须同步更新 `blog/FUNCTION_INDEX.md`：新增函数、修改函数签名或行为、删除函数。
- 提交 PR 时，需在描述中明确确认：`已检查复用并更新 blog/FUNCTION_INDEX.md`。


## 测试规范
- Go 测试文件与实现文件放在相邻位置，命名为 `*_test.go`（示例见 `blog/internal/data/`）。
- 前端测试使用 Vitest，路径为 `src/**/__tests__/*.spec.ts`。
- 涉及跨项目改动时，同时运行 `go test ./...` 和 `pnpm test:unit`。

## Commit 与 Pull Request 规范
近期提交信息通常简短、祈使语气（例如 `add debtDetail server`、`fix bug`），有时会中英混合。请保持提交聚焦且描述清晰，推荐使用 `scope: action` 风格，例如 `blog: add debt detail endpoint`。

PR 应包含：
- 行为变更的清晰摘要。
- 涉及的模块/路径（backend/frontend/proto）。
- 验证依据（测试命令与结果）。
- 若涉及 UI 或契约行为变更，提供 API 截图或请求/响应示例。
