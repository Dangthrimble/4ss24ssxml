# 4ss24ssxml

CLI that converts forScore **native** `.4ss` setlists (gzip-compressed binary plist / NSKeyedArchiver) into forScore **Open Setlist XML** (`.4ss` with an XML body) for forScore 12+.

## Requirements

- Go 1.24+

## Install / build

```bash
go build -o 4ss24ssxml ./cmd/4ss24ssxml
```

Or:

```bash
go install github.com/Dangthrimble/4ss24ssxml/cmd/4ss24ssxml@latest
```

## Usage

```bash
4ss24ssxml <file.4ss>
```

Writes a sibling file named `* (forScore12plus).4ss`, prints the output path, and exits:

- `2` — bad usage
- `1` — read / convert / write error
- `0` — success

## What it converts

Setlist items are classified as:

| Element | Source cues |
|---------|-------------|
| `score` | PDF `FilePath` |
| `bookmark` | `Bookmark` / type fields |
| `placeholder` | `Placeholder` |

## Layout

| Path | Role |
|------|------|
| [`cmd/4ss24ssxml`](cmd/4ss24ssxml) | CLI entrypoint |
| [`internal/convert`](internal/convert) | Conversion library (gzip → plist → Open Setlist XML) |

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
