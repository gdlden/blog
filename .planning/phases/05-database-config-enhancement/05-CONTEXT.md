# Phase 5: Database Configuration Enhancement - Context

**Gathered:** 2026-04-05
**Status:** Ready for planning

<domain>
## Phase Boundary

Enable database connection source to be injected via environment variables for flexible deployment and configuration management. This phase focuses on the backend Kratos service configuration only—no frontend changes required.

</domain>

<decisions>
## Implementation Decisions

### Kratos Integration Strategy
- **D-01:** Use Kratos built-in `config.NewEnvSource()` for environment variable configuration
- **D-02:** Wire the env source into the bootstrap configuration alongside the existing file source
- **D-03:** Leverage Kratos configuration merge behavior where later sources override earlier ones

### Environment Variable Naming
- **D-04:** Primary variable: `DATABASE_URL`
- **D-05:** Follow standard connection string format: `postgres://user:password@host:port/dbname?sslmode=disable`
- **D-06:** Kratos env source prefix handling: map to `DATA_DATABASE_SOURCE` (following Kratos env naming convention for nested protobuf fields)

### Configuration Priority
- **D-07:** Environment variables take precedence over YAML config file when both are present
- **D-08:** Load order: file source first, then env source (so env overrides file)
- **D-09:** Maintain backward compatibility—config.yaml works as before when env var is not set

### Claude's Discretion
- Exact implementation of Wire provider changes in `cmd/blog/wire.go`
- Whether to support additional env vars for partial override (DB_HOST, DB_PORT, etc.)
- Documentation format and location

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Configuration
- `blog/internal/conf/conf.proto` — Protobuf configuration schema (Data.Database.source field)
- `blog/configs/config.yaml` — Current configuration file with hardcoded database source

### Data Layer
- `blog/internal/data/data.go` — NewDb() function that receives conf.Data and opens GORM connection
- `blog/cmd/blog/wire.go` — Wire injection setup where configuration sources are wired

### Kratos Framework
- No external specs — use standard Kratos v2 configuration patterns

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `conf.Data` struct already defined with `Database.Source` field (string)
- `NewDb(c *conf.Data)` function in `data.go` ready to receive env-injected config
- Existing Wire setup in `cmd/blog/wire.go` can be extended with additional config sources

### Established Patterns
- Kratos configuration uses protobuf-generated structs
- Configuration loading happens in `main.go` or `wire.go` via `config.NewFileSource()`
- GORM opens connection using `postgres.Open(c.Database.Source)` with DSN format string

### Integration Points
- Wire provider set in `cmd/blog/wire.go` — needs env source added
- `configs/config.yaml` — remains as fallback/default configuration
- Docker/deployment contexts — will now support `DATABASE_URL` injection

</code_context>

<specifics>
## Specific Ideas

- Standard `DATABASE_URL` format for compatibility with cloud providers (Railway, Render, AWS RDS, etc.)
- No changes needed to `data.go` — the same `c.Database.Source` field receives the value
- Backward compatibility: existing `make run` workflow continues to work with config.yaml

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-database-config-enhancement*
*Context gathered: 2026-04-05*
