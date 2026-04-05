# Phase 5: Database Configuration Enhancement - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-05
**Phase:** 05-database-config-enhancement
**Areas discussed:** Kratos Integration Strategy, Environment Variable Naming, Configuration Priority

---

## Kratos Integration Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Kratos 内置 env 配置源 | Use Kratos framework's `config.NewEnvSource()` for standard env variable loading | ✓ |
| 代码手动解析注入 | Manual `os.Getenv()` parsing and injection into conf.Data | |

**User's choice:** Kratos 内置 env 配置源（推荐）
**Notes:** Standard approach with minimal code changes, leverages existing Kratos configuration infrastructure

---

## Environment Variable Naming

| Option | Description | Selected |
|--------|-------------|----------|
| DATABASE_URL | Standard convention used by cloud providers | ✓ |
| DB_SOURCE | Matches conf.proto field name | |
| BLOG_DATABASE_URL | Application-specific prefix | |

**User's choice:** DATABASE_URL（标准惯例）
**Notes:** Chosen for compatibility with cloud deployment platforms (Railway, Render, AWS RDS, etc.)

---

## Configuration Priority

| Option | Description | Selected |
|--------|-------------|----------|
| 环境变量优先 | Environment variables override config file | ✓ |
| 配置文件优先 | Config file takes precedence | |
| 报错冲突 | Error if both sources provide conflicting values | |

**User's choice:** 环境变量优先（推荐）
**Notes:** Standard pattern for containerized deployments, enables 12-factor app compliance

---

## Claude's Discretion

- Wire provider implementation details
- Additional partial override support (DB_HOST, DB_PORT, etc.)
- Documentation formatting

## Deferred Ideas

None — discussion stayed within phase scope
