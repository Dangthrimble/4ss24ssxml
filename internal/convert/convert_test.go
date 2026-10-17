package convert

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

func TestOutputPath(t *testing.T) {
	dir := t.TempDir()
	got := OutputPath(filepath.Join(dir, "2026-05-16 Mount Pleasant Baptist.4ss"))
	want := filepath.Join(dir, "2026-05-16 Mount Pleasant Baptist (forScore12plus).4ss")
	if got != want {
		t.Fatalf("outputPath: got %q want %q", got, want)
	}
}

func TestConvertSample(t *testing.T) {
	path := filepath.Join("testdata", "2026-05-16 Mount Pleasant Baptist.4ss")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Convert(data, SetlistTitle(path))
	if err != nil {
		t.Fatal(err)
	}

	type item struct {
		XMLName xml.Name
		Title   string `xml:"title,attr"`
		Path    string `xml:"path,attr"`
	}
	type doc struct {
		Kind    string `xml:"kind,attr"`
		Version string `xml:"version,attr"`
		Title   string `xml:"title,attr"`
		Items   []item `xml:",any"`
	}
	var parsed doc
	if err := xml.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("unmarshal result: %v\n%s", err, out)
	}
	if parsed.Kind != "setlist" || parsed.Version != "1.0" {
		t.Fatalf("root attrs: kind=%q version=%q", parsed.Kind, parsed.Version)
	}
	if parsed.Title != "2026-05-16 Mount Pleasant Baptist" {
		t.Fatalf("title: %q", parsed.Title)
	}
	if len(parsed.Items) != 16 {
		t.Fatalf("item count: got %d want 16", len(parsed.Items))
	}

	first := parsed.Items[0]
	if first.XMLName.Local != "placeholder" || first.Title != "Hymn - Lord, For The Years" {
		t.Fatalf("first item: %+v", first)
	}

	second := parsed.Items[1]
	if second.XMLName.Local != "score" {
		t.Fatalf("second item kind: %s", second.XMLName.Local)
	}
	wantTitle := "For The Beauty Of The Earth ~JoRu~ [2b] SATB+Piano{Scanned}"
	wantPath := wantTitle + ".pdf"
	if second.Title != wantTitle || second.Path != wantPath {
		t.Fatalf("second item: title=%q path=%q", second.Title, second.Path)
	}

	var sawSwingin bool
	placeholderCount, scoreCount := 0, 0
	for _, it := range parsed.Items {
		switch it.XMLName.Local {
		case "placeholder":
			placeholderCount++
			if it.Path != "" {
				t.Fatalf("placeholder has path: %+v", it)
			}
		case "score":
			scoreCount++
		case "bookmark":
			t.Fatalf("unexpected bookmark: %+v", it)
		default:
			t.Fatalf("unknown element %s", it.XMLName.Local)
		}
		if it.Title == "Swingin' With The Saints ~TrAd, arr. MaHa~ [3b] SATB+Piano{S}" {
			sawSwingin = true
			if it.XMLName.Local != "score" {
				t.Fatalf("Swingin' item is %s", it.XMLName.Local)
			}
		}
	}
	if !sawSwingin {
		t.Fatal("missing Swingin' With The Saints score")
	}
	if placeholderCount != 6 || scoreCount != 10 {
		t.Fatalf("counts: placeholders=%d scores=%d", placeholderCount, scoreCount)
	}

	if !bytes.HasPrefix(out, []byte(`<?xml version="1.0" encoding="UTF-8" ?>`)) {
		t.Fatalf("missing XML declaration:\n%s", out[:80])
	}
	if !bytes.Contains(out, []byte(`<placeholder title="Hymn - Lord, For The Years" />`)) {
		t.Fatalf("expected self-closing placeholder:\n%s", out)
	}
}

func TestConvertRejectsXML(t *testing.T) {
	_, err := Convert([]byte(`<?xml version="1.0"?><forScore kind="setlist" version="1.0" title="x"/>`), "x")
	if err == nil {
		t.Fatal("expected error for Open Setlist XML input")
	}
}
