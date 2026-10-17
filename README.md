# 4ss24ssxml

Tools that produce forScore **Open Setlist** `.4ss` files (XML body) for **forScore 12+**.

| Command | Purpose |
|---------|---------|
| `4ss2open4ss` | Native `.4ss` setlist (gzip / binary plist) → Open Setlist |
| `txt2open4ss` | Text file (one title per non-blank line) → Open Setlist placeholders |

## Requirements

- Go 1.24+

## Install / build

```bash
go build -o 4ss2open4ss ./cmd/4ss2open4ss
go build -o txt2open4ss ./cmd/txt2open4ss
```

Or:

```bash
go install github.com/Dangthrimble/4ss24ssxml/cmd/4ss2open4ss@latest
go install github.com/Dangthrimble/4ss24ssxml/cmd/txt2open4ss@latest
```

## Usage

```bash
4ss2open4ss <file.4ss>
txt2open4ss <file.txt>
```

Each writes a sibling file named `* (forScore12plus).4ss` (Open Setlist is only readable in forScore 12+), prints the output path, and exits:

- `2` — bad usage
- `1` — read / convert / write error
- `0` — success

### `4ss2open4ss`

Setlist items are classified as:

| Element | Source cues |
|---------|-------------|
| `score` | PDF `FilePath` |
| `bookmark` | `Bookmark` / type fields |
| `placeholder` | `Placeholder` |

### `txt2open4ss`

Blank and whitespace-only lines are skipped. Each non-blank line becomes a `<placeholder title="..." />`.

## Layout

| Path | Role |
|------|------|
| [`cmd/4ss2open4ss`](cmd/4ss2open4ss) | Native `.4ss` → Open Setlist |
| [`cmd/txt2open4ss`](cmd/txt2open4ss) | Text lines → Open Setlist placeholders |
| [`internal/convert`](internal/convert) | Conversion library |

## Tests

```bash
go test ./...
```

Fixture: [`internal/convert/testdata`](internal/convert/testdata).

## License

[MIT](LICENSE)

## Related docs

- [DEVELOPMENT_HISTORY.md](DEVELOPMENT_HISTORY.md)
- [BACKLOG.md](BACKLOG.md)
