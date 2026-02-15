# How To Add Config Options

This guide covers adding a new setting that is:

- persisted in `.commiter.json`
- available at runtime
- editable via `commiter config`

## 1. Add the field to config types

Edit `internal/config/types.go` and add the new field to `Config` with a JSON tag.

Example:

```go
EnableFoo bool `json:"enable_foo,omitempty"`
```

## 2. Set defaults + upgrade behavior

Edit `internal/config/config.go` in `Load()`:

- when loading existing config, backfill default values if missing
- set a sane default in the "Return default config" block

For non-trivial values, add a safe accessor method in `internal/config/config.go` (pattern: `GetHookTimeoutSeconds()`).

## 3. Wire runtime usage

Read the setting where behavior is executed:

- CLI startup and bypass paths: `cmd/commiter/start.go`
- interactive model behavior: `internal/ui/model.go`
- provider/runtime plumbing as needed

Keep runtime logic using accessors for fallback safety where possible.

## 4. Expose it in config TUI

Edit `internal/ui/configui.go`:

- add a menu row in `configItems(...)`
- add read mapping in `getFieldValue(...)`
- add write validation/mapping in `setFieldValue(...)`
- if needed, add a dedicated editor flow (see hook list handling as example)

## 5. Update user-facing docs

If relevant, update:

- `README.md` configuration JSON example
- command docs describing the behavior

## 6. Add/adjust tests

Recommended coverage:

- `internal/config/config_test.go`: load/save defaults and migration behavior
- UI model/config tests where the setting changes behavior (for example `internal/ui/model_test.go`)

Then run:

```bash
go test ./...
```

