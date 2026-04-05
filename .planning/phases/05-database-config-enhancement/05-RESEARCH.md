# Phase 5: Database Configuration Enhancement - Research

**Researched:** 2026-04-05
**Domain:** Kratos v2 Configuration System
**Confidence:** HIGH

## Summary

This phase enables database connection configuration via environment variables for flexible deployment. The implementation leverages Kratos v2's built-in `config/env` package to add an environment variable source alongside the existing file source.

Kratos configuration uses a source-based architecture where multiple sources (file, env, etc.) are merged in order, with later sources overriding earlier ones. The `config.New()` function accepts multiple sources via `config.WithSource()`, and the internal merge uses `mergo.WithOverride`, meaning **sources specified later take precedence**.

For environment variable mapping, Kratos uses a prefix-based filter. When creating an env source with `env.NewSource("DATA_DATABASE")`, only environment variables starting with that prefix are captured, and the prefix is stripped. The key `DATA_DATABASE_SOURCE` becomes `SOURCE` in the config, which maps to the protobuf field `Data.Database.Source`.

**Primary recommendation:** Add `env.NewSource("DATA_DATABASE")` as a second source after the file source in `main.go`, ensuring env vars override YAML config while maintaining backward compatibility.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Use Kratos built-in `config.NewEnvSource()` for environment variable configuration
- **D-02:** Wire the env source into the bootstrap configuration alongside the existing file source
- **D-03:** Leverage Kratos configuration merge behavior where later sources override earlier ones
- **D-04:** Primary variable: `DATABASE_URL`
- **D-05:** Follow standard connection string format: `postgres://user:password@host:port/dbname?sslmode=disable`
- **D-06:** Kratos env source prefix handling: map to `DATA_DATABASE_SOURCE` (following Kratos env naming convention for nested protobuf fields)
- **D-07:** Environment variables take precedence over YAML config file when both are present
- **D-08:** Load order: file source first, then env source (so env overrides file)
- **D-09:** Maintain backward compatibility—config.yaml works as before when env var is not set

### Claude's Discretion
- Exact implementation of Wire provider changes in `cmd/blog/wire.go`
- Whether to support additional env vars for partial override (DB_HOST, DB_PORT, etc.)
- Documentation format and location

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CFG-01 | Enable database connection source to be injected via environment variables for flexible deployment | Kratos env source with prefix `DATA_DATABASE` maps to `Data.Database.Source` field; load order file then env ensures override behavior |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-kratos/kratos/v2/config | v2.x | Configuration management | Framework-native, supports multiple sources |
| go-kratos/kratos/v2/config/env | v2.x | Environment variable source | Built-in, no external dependencies |
| go-kratos/kratos/v2/config/file | v2.x | File source (YAML) | Already in use |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| mergo | indirect | Map merging with override | Used internally by Kratos for source merging |

## Architecture Patterns

### Recommended Configuration Flow

```
main.go
  └── config.New(
        config.WithSource(
          file.NewSource(flagconf),     // 1. Load YAML (base config)
          env.NewSource("DATA_DATABASE") // 2. Load env vars (overrides)
        )
      )
          │
          ▼
      config.Load() ──► Merge sources (later overrides earlier)
          │
          ▼
      config.Scan(&bc) ──► conf.Bootstrap{Data: {Database: {Source: "..."}}}
          │
          ▼
      wireApp(bc.Server, bc.Data, logger)
          │
          ▼
      data.NewDb(bc.Data) ──► gorm.Open(postgres.Open(c.Database.Source))
```

### Pattern 1: Multiple Config Sources with Override
**What:** Load configuration from multiple sources where later sources override earlier ones
**When to use:** When you need environment-specific overrides of base configuration
**Example:**
```go
// Source: Kratos config/config.go - mergo.WithOverride behavior
c := config.New(
    config.WithSource(
        file.NewSource("configs/config.yaml"),  // Base config
        env.NewSource("DATA_DATABASE"),          // Env overrides
    ),
)
```

### Pattern 2: Environment Variable Prefix Mapping
**What:** Use prefixed env vars to map to nested config structures
**When to use:** When mapping flat env vars to hierarchical config (protobuf structs)
**Example:**
```go
// Source: Kratos config/env/env.go
// Environment variable: DATA_DATABASE_SOURCE=postgres://...
// Maps to: conf.Data.Database.Source

src := env.NewSource("DATA_DATABASE")
// Loads: DATA_DATABASE_* vars
// Strips prefix: DATA_DATABASE_SOURCE → SOURCE
// Maps to: Data.Database.Source (via config reader)
```

### Anti-Patterns to Avoid
- **Single source only:** Do not replace file source with env source—use both for fallback
- **Wrong load order:** Do not put env source before file source (would be overridden)
- **Missing prefix:** Do not use `env.NewSource()` without prefix—loads all env vars
- **Manual env var parsing:** Do not use `os.Getenv()` directly when Kratos env source exists

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Env var to config mapping | Custom `os.Getenv()` parsing | `config/env` package | Prefix filtering, automatic key stripping, watch support |
| Config merging | Manual map merging | Kratos `config.Load()` with multiple sources | Uses `mergo.WithOverride`, handles edge cases, type safety |
| Hot reload | Custom file watcher | `config.Watch()` | Built-in watcher per source, async updates |
| Nested field access | Custom struct traversal | `config.Value()` | Path-based access, type conversion |

**Key insight:** Kratos config system is designed for exactly this use case. Custom implementations would miss edge cases like proper merging semantics, watch support, and type-safe value access.

## Common Pitfalls

### Pitfall 1: Incorrect Environment Variable Name
**What goes wrong:** Setting `DATABASE_URL` but expecting it to map to `Data.Database.Source` without proper prefix configuration
**Why it happens:** Kratos env source requires the full prefixed name to match and strip
**How to avoid:** Use `DATA_DATABASE_SOURCE` as the env var name when using prefix `DATA_DATABASE`
**Warning signs:** Config scan succeeds but database connection uses YAML value instead of env var

### Pitfall 2: Wrong Source Order
**What goes wrong:** Env vars not overriding YAML config despite being set
**Why it happens:** Sources are merged in order; earlier sources are overridden by later ones
**How to avoid:** Always specify file source first, then env source: `file.NewSource(...), env.NewSource(...)`
**Warning signs:** Changes to env vars have no effect on running application

### Pitfall 3: Connection String Format Mismatch
**What goes wrong:** `postgres://` URL format not accepted by GORM postgres driver
**Why it happens:** GORM's `postgres.Open()` expects DSN format (`host=... user=...`) not URL format
**How to avoid:** Either use DSN format in env var, or use a DSN parser to convert URL to DSN
**Warning signs:** Database connection fails with "invalid connection string" error

### Pitfall 4: Missing Import
**What goes wrong:** Build fails with undefined `env` package
**Why it happens:** `config/env` is a separate subpackage requiring explicit import
**How to avoid:** Add `"github.com/go-kratos/kratos/v2/config/env"` to imports in `main.go`
**Warning signs:** Compilation error: `undefined: env`

## Code Examples

### Current Configuration Setup (main.go)
```go
// Source: blog/cmd/blog/main.go (current state)
import (
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/file"
)

c := config.New(
    config.WithSource(
        file.NewSource(flagconf),
    ),
)
```

### Enhanced Configuration with Env Source
```go
// Source: Kratos config patterns + project requirements
import (
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/env"
    "github.com/go-kratos/kratos/v2/config/file"
)

c := config.New(
    config.WithSource(
        file.NewSource(flagconf),           // Base config from YAML
        env.NewSource("DATA_DATABASE"),      // Override with env vars
    ),
)
```

### Environment Variable Usage
```bash
# Set the environment variable (DATA_DATABASE prefix + _SOURCE suffix)
export DATA_DATABASE_SOURCE="host=prod-db.example.com user=app password=secret dbname=blog port=5432 sslmode=require"

# Or using postgres URL format (requires conversion if GORM doesn't support it)
export DATA_DATABASE_SOURCE="postgres://app:secret@prod-db.example.com:5432/blog?sslmode=require"
```

### Protobuf Configuration Schema
```protobuf
// Source: blog/internal/conf/conf.proto
message Data {
  message Database {
    string driver = 1;
    string source = 2;  // Maps to DATA_DATABASE_SOURCE env var
  }
  Database database = 1;
}
```

### Database Initialization (No Changes Required)
```go
// Source: blog/internal/data/data.go (unchanged)
func NewDb(c *conf.Data) *gorm.DB {
    db, err := gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    // ...
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single file source | Multiple sources with merge | Kratos v2 | Flexible deployment configs |
| `os.Getenv()` parsing | `config/env` package | Kratos v2 | Type-safe, prefix-filtered, watchable |
| Manual config structs | Protobuf-generated configs | Kratos v2 | Schema validation, code generation |

**Deprecated/outdated:**
- Direct `os.Getenv()` calls for configuration: Use Kratos config sources instead
- Custom config merging: Use Kratos `config.Load()` with multiple sources

## Open Questions

1. **Connection string format compatibility**
   - What we know: GORM postgres driver accepts DSN format (`host=... user=...`)
   - What's unclear: Whether it also accepts `postgres://` URL format directly
   - Recommendation: Use DSN format in `DATA_DATABASE_SOURCE` to avoid conversion complexity; document this requirement

2. **Additional environment variables for partial override**
   - What we know: Current approach uses single `DATA_DATABASE_SOURCE` for full DSN
   - What's unclear: Whether to support `DB_HOST`, `DB_PORT`, etc. as separate env vars
   - Recommendation: Start with single `DATA_DATABASE_SOURCE` (matches CONTEXT.md D-04/D-05); add partial overrides later if needed

3. **Wire provider changes necessity**
   - What we know: `wireApp` receives `*conf.Data` which is populated by `config.Scan()`
   - What's unclear: Whether any Wire changes are actually needed
   - Recommendation: Likely no Wire changes needed—only `main.go` changes; verify during implementation

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Kratos config/env | Env var configuration | ✓ | v2 (via go.mod) | — |
| PostgreSQL | Database connection testing | ✓ | Running locally | — |
| Go compiler | Build and test | ✓ | 1.21+ | — |

**Missing dependencies with no fallback:**
- None

**Missing dependencies with fallback:**
- None

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify |
| Config file | `blog/configs/config.yaml` (fallback) |
| Quick run command | `cd blog && go test ./internal/data/... -v` |
| Full suite command | `cd blog && go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CFG-01 | Database config loads from env var | integration | Manual verification | ❌ Wave 0 |
| CFG-01 | Env var overrides YAML config | integration | Manual verification | ❌ Wave 0 |
| CFG-01 | Backward compatibility (YAML only) | integration | `go test ./...` | ✅ Existing |

### Sampling Rate
- **Per task commit:** `go test ./internal/data/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] Env var configuration test — verify `DATA_DATABASE_SOURCE` is loaded
- [ ] Override behavior test — verify env takes precedence over YAML
- [ ] Documentation — `.env.example` or README update

## Sources

### Primary (HIGH confidence)
- Kratos config/env source: `github.com/go-kratos/kratos/config/env/env.go` - Prefix handling, key stripping
- Kratos config merging: `github.com/go-kratos/kratos/config/config.go` - `mergo.WithOverride` behavior
- Project codebase: `blog/cmd/blog/main.go` - Current configuration setup
- Project codebase: `blog/internal/conf/conf.proto` - Configuration schema

### Secondary (MEDIUM confidence)
- Kratos documentation patterns - Standard multi-source configuration approach
- GORM postgres driver documentation - DSN format expectations

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Kratos built-in packages, well-documented patterns
- Architecture: HIGH - Clear source ordering and merge semantics
- Pitfalls: MEDIUM-HIGH - Based on common Kratos usage patterns and protobuf mapping behavior

**Research date:** 2026-04-05
**Valid until:** 2026-05-05 (Kratos v2 is stable, low churn expected)

---

## Implementation Checklist for Planner

### Files to Modify
1. `blog/cmd/blog/main.go` - Add env source import and configuration
2. `blog/configs/config.yaml` - No changes (remains as fallback)
3. `blog/internal/data/data.go` - No changes (already receives `*conf.Data`)
4. `blog/cmd/blog/wire.go` - Likely no changes (verify during implementation)

### Code Changes Required
```go
// In blog/cmd/blog/main.go:
// 1. Add import: "github.com/go-kratos/kratos/v2/config/env"
// 2. Modify config.New() call to include env source after file source
```

### Testing Approach
1. Unit test: Verify config loads correctly with only YAML (backward compatibility)
2. Integration test: Set `DATA_DATABASE_SOURCE` env var, verify it overrides YAML
3. Manual verification: Run service with env var, confirm database connection uses env value

### Documentation Updates
1. Add `.env.example` file showing `DATA_DATABASE_SOURCE` format
2. Update README or deployment docs with environment variable configuration instructions
