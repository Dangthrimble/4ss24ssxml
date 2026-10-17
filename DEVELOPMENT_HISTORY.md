# Development history

| When | What |
|------|------|
| 2026-10-17 | Initial checkpoint: converter CLI (`main.go`, `convert.go`), tests, Mount Pleasant Baptist fixture; module originally named `4ss2xml`. |
| 2026-10-17 | Module renamed to `4ss24ssxml`. |
| 2026-10-17 | Scrubbed `.DS_Store` from git history on `main`. |
| 2026-10-17 | Restructure to `cmd/4ss24ssxml` + `internal/convert`; module path `github.com/Dangthrimble/4ss24ssxml`; README, backlog, MIT license; fixture hygiene. |
| 2026-10-17 | Renamed CLI to `4ss2open4ss`; added `txt2open4ss` (non-blank lines → placeholders); shared `(forScore12plus)` output marker. |

## Design notes

- Native `.4ss` input is treated as gzip → binary plist → NSKeyedArchiver (`$objects` / `$top.setlist`).
- Already-XML / Open Setlist inputs are rejected.
- Output file name: input stem + ` (forScore12plus)` + `.4ss`.
- Dependency: [`howett.net/plist`](https://pkg.go.dev/howett.net/plist).
