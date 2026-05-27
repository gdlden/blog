---
name: kratos-crud
description: 根据 proto 定义生成 Kratos 分层 CRUD 代码（proto → service → biz → data → wire）
disable-model-invocation: true
---

# Kratos CRUD Generator

根据领域名和字段定义，生成完整的 Kratos 分层 CRUD 代码。

## 输入

用户提供：
- 领域名（如 `post`、`debt`、`fuel`）
- 实体字段定义
- 是否需要分页列表

## 生成步骤

### 1. Proto 定义

在 `blog/api/<domain>/v1/<domain>.proto` 中定义：
- 实体 message
- CRUD RPC（Create/Update/Delete/Get/List）
- 对应的 Request/Reply message

### 2. Service 层

在 `blog/internal/service/<domain>.go` 中：
- 实现 protobuf handler
- 映射请求参数到 biz 层输入
- 调用 biz usecase 方法
- 映射返回值到 protobuf reply

参考现有 service 文件（如 `blog/internal/service/debt.go`）了解模式。

### 3. Biz 层

在 `blog/internal/biz/<domain>.go` 中：
- 定义领域模型 struct（带 GORM 标签）
- 定义 repo 接口（如 `PostRepo`）
- 定义 usecase struct 及构造函数 `New<Domain>Usecase`
- 实现 CRUD 方法，每个方法：
  - 通过 `userutil.GetUserID(ctx)` 获取当前用户
  - 所有权校验（update/delete 时）
  - 调用 repo 接口

### 4. Data 层

在 `blog/internal/data/<domain>.go` 中：
- 实现 biz 层定义的 repo 接口
- 构造函数 `New<Domain>Repo`
- 使用 GORM 操作数据库

### 5. Wire 注入

更新以下文件的 Wire provider：
- `blog/internal/data/data.go` — 添加 repo 构造函数
- `blog/internal/biz/biz.go` — 添加 usecase 构造函数
- `blog/internal/service/service.go` — 添加 service 构造函数

### 6. 运行代码生成

```bash
cd blog && make api
cd blog/cmd/blog && wire
```

### 7. 更新 FUNCTION_INDEX.md

将新增的导出函数添加到 `blog/FUNCTION_INDEX.md`。

## 输出

生成所有文件后，给出文件清单和下一步操作（make api / wire / make run）。
