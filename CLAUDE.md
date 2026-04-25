# CLAUDE.md

本文件为 Claude Code（claude.ai/code）在此仓库中工作时提供指引。

## 仓库结构

这是一个 monorepo，包含两个项目：
- `blog/` — 使用 Kratos 框架的 Go 后端服务
- `price_recorder_vue/` — Vue 3 + Vite 前端 SPA

后端遵循严格的 Kratos 分层模式：`server`（传输层）→ `service` → `biz`（业务逻辑）→ `data`（持久化）。依赖通过 `blog/cmd/blog/wire.go` 中的 Google Wire 注入。

## 后端开发（blog/）

### 常用命令
```bash
cd blog

# 初始化：安装 proto / 代码生成工具
make init

# 从 proto 文件生成代码
make api          # 生成 API protobuf（同时生成 OpenAPI 到 openapi.yaml）
make config       # 生成内部配置 protobuf
make all          # 运行 api + config + generate

# 运行和构建
make run          # 运行 Kratos 服务（自动加载 .env）
make build        # 构建二进制文件到 blog/bin/
go test ./...     # 运行 Go 测试

# 依赖注入
cd cmd/blog && wire  # 生成 Wire 装配代码
```

### 本地开发配置

后端通过环境变量 `DATA_DATABASE_SOURCE` 连接数据库。为避免将真实密码暴露给 Claude，采用以下流程：

1. 复制 `blog/.env.example` 为 `blog/.env`，填入你的真实数据库连接信息
2. `blog/.env` 已被 `.gitignore` 排除，不会提交到仓库
3. `make run` 会自动加载 `blog/.env` 中的环境变量
4. 调试前后端时，Claude 只需执行 `cd blog && make run`，无需知道密码内容

### 架构分层

**Transport（`blog/internal/server/`）**：HTTP/gRPC 服务器配置、中间件、handler 注册。认证中间件为公开路由配置了白名单。

**Service（`blog/internal/service/*.go`）**：实现 protobuf handler，映射请求/响应，调用用例。每个领域一个 service 文件（post.go、user.go、debt.go、price.go 等）。

**Business（`blog/internal/biz/*.go`）**：领域模型、用例、仓库接口。用例包含业务规则，调用 repo 接口。

**Data（`blog/internal/data/*.go`）**：仓库实现、GORM 数据库初始化、持久化模型。后端测试放在 `*_test.go` 文件中。

**Model（`blog/internal/model/`）**：跨 repo 共享的 GORM 模型（如 User）。

### 新增后端功能

1. 在 `blog/api/<domain>/v1/<domain>.proto` 中定义或更新 protobuf 契约
2. 运行 `make api` 生成 API 绑定并更新 `blog/openapi.yaml`
3. 在 `blog/internal/service/<domain>.go` 中实现 service handler
4. 在 `blog/internal/biz/<domain>.go` 中添加用例和 repo 接口
5. 在 `blog/internal/data/<domain>.go` 或 `blog/internal/model/` 中实现 repo（共享模型）
6. 如新增构造函数，更新 `blog/internal/{data,biz,service}/*.go` 中的 Wire provider

### 函数复用规则

在 `blog/internal` 中新增后端函数前，先查看 `blog/FUNCTION_INDEX.md` 是否有可复用的现有函数。添加/删除/修改函数时，在同一次变更中同步更新索引。PR 描述中应确认已执行此操作。

## 前端开发（price_recorder_vue/）

### 常用命令
```bash
cd price_recorder_vue

# 安装
pnpm install

# 开发
pnpm dev          # 启动 Vite 开发服务器

# 构建和测试
pnpm build        # 类型检查并构建生产资源
pnpm test:unit    # 运行 Vitest 单元测试
pnpm lint         # 运行 ESLint 并自动修复
pnpm format       # 运行 Prettier
```

### 架构

**入口**：`src/main.ts` 挂载 Vue 应用，集成 Pinia 和 Vue Router。

**路由（`src/router/index.ts`）**：路由定义和认证守卫。未认证用户重定向到 `/login`。

**视图（`src/view/*.vue`）**：路由级页面。组件使用 PascalCase 文件名（如 Article.vue、Login.vue）。

**API 客户端（`src/api/*.ts`）**：薄封装层，导入共享的 Axios 实例并调用后端接口。

**共享传输层（`src/utils/request.ts`）**：集中式 Axios 实例，带请求/响应拦截器。从 localStorage 读取 token 并添加 `Authorization: Bearer <token>` 请求头。

**Store（`src/stores/*.ts`）**：Pinia store 用于跨视图持久化状态（如 userStore.ts 镜像 localStorage 中的认证信息）。

### 新增前端功能

1. 在 `src/view/<FeatureName>.vue` 中添加路由级 UI
2. 在 `src/router/index.ts` 中注册路由
3. 如需，在 `src/api/<FeatureName>.ts` 中添加 API 封装
4. 跨视图共享状态使用 `src/stores/`
5. 仅在涉及跨模块传输变更时修改 `src/utils/request.ts`

## 代码规范

**后端**：Go 文件使用小写文件名（post.go、user.go）。Proto 文件使用版本化路径（api/post/v1/post.proto）。提交前使用 `go fmt ./...` 格式化。

**前端**：Vue 组件使用 PascalCase 文件名。使用 `eslint.config.ts` 和 `.prettierrc.json` 中的 ESLint + Prettier 配置。

**API 契约**：protobuf 文件维护在 `blog/api/` 下。修改后运行 `make api` 重新生成绑定和 `blog/openapi.yaml`。

## 跨领域关注点

**认证**：后端 JWT 中间件在 `blog/internal/server/http.go` 中，带路由白名单。前端 token 存在 localStorage，通过 `src/utils/request.ts` 中的 Axios 拦截器附加。用户身份通过 `blog/internal/utils/userutil.go` 提取。

**数据库**：`blog/internal/data/data.go` 启用 GORM 自动迁移。使用 PostgreSQL 驱动。Decimal 字段使用 `github.com/shopspring/decimal`。

**配置**：后端通过 Kratos file source 读取 `blog/configs/config.yaml`。Schema 定义在 `blog/internal/conf/conf.proto` 中。

## 测试

- 后端测试：放在 `blog/internal/**/*_test.go`，与实现相邻。使用 `go test ./...` 运行。
- 前端测试：放在 `src/__tests__/*.spec.ts`。使用 `pnpm test:unit` 运行。

## 重要参考

- `blog/FUNCTION_INDEX.md` — 后端函数复用索引
- `blog/openapi.yaml` — HTTP 端点生成的 OpenAPI 规范
- `AGENTS.md` — 详细的仓库规格说明（中文）
- `.planning/codebase/` — 架构和结构文档
