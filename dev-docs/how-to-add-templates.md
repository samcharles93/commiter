# How To Add Templates

Templates are loaded from config and used to guide LLM commit message generation.

## Fast path: add a built-in default template

Edit `internal/config/templates.go` and append a `CommitTemplate` entry in `GetDefaultTemplates()`.

Use this shape:

```go
{
    Key:    "my-template",
    Name:   "My Template",
    Format: "{subject}",
    Prompt: "Generate a commit message ...",
}
```

## Field expectations

- `Key`: stable machine value used in config (`default_template`), lowercase slug format.
- `Name`: display label shown in the UI.
- `Format`: display hint only.
- `Prompt`: instruction block injected into message generation.

## How selection works

Template lookup is handled by `internal/config/config.go` (`FindTemplate` / `ResolveDefaultTemplate`):

- accepts key values (for example `simple`, `conventional`)
- accepts display names (for example `Simple`)
- accepts `default` as alias to the first template in the list

`default_template` should store canonical key values.

## Config examples

In `.commiter.json`:

```json
{
  "default_template": "conventional"
}
```

Other valid values depend on configured templates, for example `simple`, `my-template`, or `default`.

## Add custom templates without code changes

Users can define templates directly in `.commiter.json` under `templates`.  
If you add custom templates this way, ensure each one has a unique key.

## Verification checklist

1. `commiter config` shows new template in accepted values.
2. Setting `default_template` to the new key works.
3. Interactive flow starts without template prompt when default is set.
4. `-y` bypass mode uses the selected default template.

Run:

```bash
go test ./...
```

