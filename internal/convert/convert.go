package convert

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

type setlistItemKind int

const (
	kindScore setlistItemKind = iota
	kindBookmark
	kindPlaceholder
)

type setlistItem struct {
	Kind  setlistItemKind
	Title string
	Path  string
}

const forScore12PlusSuffix = " (forScore12plus)"

func OutputPath(input string) string {
	dir := filepath.Dir(input)
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, stem+forScore12PlusSuffix+ext)
}

// TextOutputPath returns a sibling Open Setlist path: stem (forScore12plus).4ss.
func TextOutputPath(input string) string {
	dir := filepath.Dir(input)
	base := filepath.Base(input)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join(dir, stem+forScore12PlusSuffix+".4ss")
}

func SetlistTitle(input string) string {
	base := filepath.Base(input)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// FromText builds an Open Setlist XML document with one placeholder per non-blank line.
func FromText(data []byte, title string) ([]byte, error) {
	var items []setlistItem
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" {
			continue
		}
		items = append(items, setlistItem{Kind: kindPlaceholder, Title: line})
	}
	return marshalOpenSetlist(title, items)
}

func Convert(data []byte, title string) ([]byte, error) {
	if isOpenSetlistXML(data) {
		return nil, fmt.Errorf("file is already an Open Setlist XML document")
	}
	if !isGzip(data) {
		return nil, fmt.Errorf("not a gzip-compressed forScore setlist (expected native .4ss)")
	}

	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}

	if !bytes.HasPrefix(raw, []byte("bplist")) {
		return nil, fmt.Errorf("not a binary plist after gzip decompress")
	}

	var archive map[string]any
	if _, err := plist.Unmarshal(raw, &archive); err != nil {
		return nil, fmt.Errorf("plist: %w", err)
	}

	items, err := itemsFromArchive(archive)
	if err != nil {
		return nil, err
	}

	return marshalOpenSetlist(title, items)
}

func isGzip(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

func isOpenSetlistXML(data []byte) bool {
	s := strings.TrimSpace(string(data))
	return strings.HasPrefix(s, "<?xml") || strings.HasPrefix(s, "<forScore")
}

func itemsFromArchive(archive map[string]any) ([]setlistItem, error) {
	objects, ok := archive["$objects"].([]any)
	if !ok {
		return nil, fmt.Errorf("NSKeyedArchiver: missing $objects")
	}
	top, ok := archive["$top"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("NSKeyedArchiver: missing $top")
	}
	resolved, err := resolve(objects, top["setlist"])
	if err != nil {
		return nil, fmt.Errorf("NSKeyedArchiver: %w", err)
	}
	if resolved == nil {
		return nil, fmt.Errorf("NSKeyedArchiver: missing $top.setlist")
	}
	rawItems, ok := resolved.([]any)
	if !ok {
		return nil, fmt.Errorf("NSKeyedArchiver: $top.setlist is not an array")
	}

	out := make([]setlistItem, 0, len(rawItems))
	for i, raw := range rawItems {
		dict, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("setlist item %d is not a dictionary", i+1)
		}
		item, err := classifyItem(dict)
		if err != nil {
			return nil, fmt.Errorf("setlist item %d: %w", i+1, err)
		}
		out = append(out, item)
	}
	return out, nil
}

func classifyItem(dict map[string]any) (setlistItem, error) {
	title, _ := dict["Title"].(string)
	path, _ := dict["FilePath"].(string)

	if isTruthy(dict["Placeholder"]) {
		if title == "" {
			return setlistItem{}, fmt.Errorf("placeholder is missing Title")
		}
		return setlistItem{Kind: kindPlaceholder, Title: title}, nil
	}

	if isBookmark(dict) {
		if title == "" || path == "" {
			return setlistItem{}, fmt.Errorf("bookmark requires Title and FilePath")
		}
		return setlistItem{Kind: kindBookmark, Title: title, Path: path}, nil
	}

	if strings.HasSuffix(strings.ToLower(path), ".pdf") {
		return setlistItem{Kind: kindScore, Title: title, Path: path}, nil
	}

	return setlistItem{}, fmt.Errorf("unrecognized item (Title=%q FilePath=%q)", title, path)
}

func isBookmark(dict map[string]any) bool {
	if _, ok := dict["Bookmark"]; ok {
		if !isTruthy(dict["Bookmark"]) {
			if s, ok := dict["Bookmark"].(string); ok {
				return s != ""
			}
			return false
		}
		return true
	}
	for _, key := range []string{"Type", "Kind"} {
		if s, ok := dict[key].(string); ok && strings.EqualFold(s, "bookmark") {
			return true
		}
	}
	return false
}

func isTruthy(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case uint64:
		return t != 0
	case int:
		return t != 0
	case float64:
		return t != 0
	case string:
		return strings.EqualFold(t, "true") || t == "1"
	default:
		return false
	}
}

func resolve(objects []any, v any) (any, error) {
	switch t := v.(type) {
	case plist.UID:
		idx := int(t)
		if idx < 0 || idx >= len(objects) {
			return nil, fmt.Errorf("UID %d out of range", idx)
		}
		return resolve(objects, objects[idx])
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			r, err := resolve(objects, e)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil
	case map[string]any:
		if _, ok := t["$classname"]; ok {
			return t, nil
		}
		keys, hasKeys := t["NS.keys"]
		objs, hasObjs := t["NS.objects"]
		if hasKeys && hasObjs {
			rk, err := resolve(objects, keys)
			if err != nil {
				return nil, err
			}
			rv, err := resolve(objects, objs)
			if err != nil {
				return nil, err
			}
			keyList, ok1 := rk.([]any)
			valList, ok2 := rv.([]any)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("NSDictionary keys/objects are not arrays")
			}
			if len(keyList) != len(valList) {
				return nil, fmt.Errorf("NSDictionary key/value count mismatch")
			}
			out := make(map[string]any, len(keyList))
			for i, k := range keyList {
				ks, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("NSDictionary key is not a string")
				}
				out[ks] = valList[i]
			}
			return out, nil
		}
		if hasObjs {
			return resolve(objects, objs)
		}
		out := make(map[string]any, len(t))
		for k, val := range t {
			if k == "$class" {
				continue
			}
			r, err := resolve(objects, val)
			if err != nil {
				return nil, err
			}
			out[k] = r
		}
		return out, nil
	default:
		return v, nil
	}
}

func marshalOpenSetlist(title string, items []setlistItem) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" ?>`)
	buf.WriteByte('\n')
	buf.WriteString(`<forScore kind="setlist" version="1.0" title="`)
	if err := xml.EscapeText(&buf, []byte(title)); err != nil {
		return nil, err
	}
	buf.WriteString(`">`)
	buf.WriteByte('\n')
	for _, it := range items {
		name := "score"
		switch it.Kind {
		case kindBookmark:
			name = "bookmark"
		case kindPlaceholder:
			name = "placeholder"
		}
		buf.WriteByte('\t')
		buf.WriteByte('<')
		buf.WriteString(name)
		buf.WriteString(` title="`)
		if err := xml.EscapeText(&buf, []byte(it.Title)); err != nil {
			return nil, err
		}
		buf.WriteByte('"')
		if it.Kind != kindPlaceholder {
			buf.WriteString(` path="`)
			if err := xml.EscapeText(&buf, []byte(it.Path)); err != nil {
				return nil, err
			}
			buf.WriteByte('"')
		}
		buf.WriteString(" />")
		buf.WriteByte('\n')
	}
	buf.WriteString("</forScore>\n")
	return buf.Bytes(), nil
}
