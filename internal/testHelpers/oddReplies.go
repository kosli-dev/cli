package testHelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/kosli-dev/cli/internal/types"
)

// NullVariant is a copy of a JSON document with the value at Path set to null.
// A Path ending in "[]" names a list emptied instead.
type NullVariant struct {
	Path string
	Body []byte
}

// NullVariants returns one copy of doc per value it holds, with that value
// made null, plus an emptied copy of each non-empty list. The document itself
// is Path "."; its fields have no leading dot. Key order is kept; a variant
// means the same as doc except where its Path says. Paths assume no key holds
// "." or "[", or two variants could share one.
func NullVariants(doc []byte) ([]NullVariant, error) {
	root, err := parseOrdered(doc)
	if err != nil {
		return nil, err
	}
	variants := nullAt(root, root, ".")
	var walk func(n *jsonNode, path string)
	walk = func(n *jsonNode, path string) {
		switch n.kind {
		case kindObject:
			for _, f := range n.fields {
				child := joinPath(path, f.key)
				variants = append(variants, nullAt(root, f.value, child)...)
				walk(f.value, child)
			}
		case kindArray:
			for i, item := range n.items {
				child := fmt.Sprintf("%s[%d]", path, i)
				variants = append(variants, nullAt(root, item, child)...)
				walk(item, child)
			}
		}
	}
	walk(root, "")
	return variants, nil
}

// nullAt gives root rendered with n made null and, for a non-empty list,
// rendered with it emptied. An already-null n gives nothing.
func nullAt(root, n *jsonNode, path string) []NullVariant {
	if n.kind == kindNull {
		return nil
	}
	saved := *n
	defer func() { *n = saved }()

	*n = jsonNode{kind: kindNull}
	out := []NullVariant{{Path: path, Body: root.render()}}
	if saved.kind == kindArray && len(saved.items) > 0 {
		*n = jsonNode{kind: kindArray}
		out = append(out, NullVariant{Path: path + "[]", Body: root.render()})
	}
	return out
}

func joinPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// Outcome runs fn and says what happened: "ok", "error", or "panic: " and the
// panic value. Error text is left out because it carries test server ports.
func Outcome(fn func() error) (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = fmt.Sprintf("panic: %v", r)
		}
	}()
	if err := fn(); err != nil {
		return "error"
	}
	return "ok"
}

// EvidenceOutcome runs fetch and says what happened, as Outcome does. When it
// succeeds it compares each pull request's API payload with baseline's, key by
// key, so it sees what the server would. "lost:" when a value the baseline had
// came back missing or null, or content it had came back empty or shorter;
// "changed:" for any other difference. Only top-level keys are compared, so a
// value lost inside a list (a review's username) shows as "changed:".
func EvidenceOutcome(baseline []*types.PREvidence, fetch func() ([]*types.PREvidence, error)) string {
	var got []*types.PREvidence
	out := Outcome(func() error {
		var err error
		got, err = fetch()
		return err
	})
	if out != "ok" {
		return out
	}
	if len(got) != len(baseline) {
		return fmt.Sprintf("ok lost:PullRequests(%d->%d)", len(baseline), len(got))
	}
	for i := range baseline {
		if got[i] == nil {
			if baseline[i] != nil {
				out += fmt.Sprintf(" lost:PullRequest%d", i)
			}
			continue
		}
		want, have := payloadFields(baseline[i]), payloadFields(got[i])
		keys := []string{}
		for k := range want {
			keys = append(keys, k)
		}
		for k := range have {
			if _, ok := want[k]; !ok {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			w, h := want[k], have[k]
			if reflect.DeepEqual(w, h) {
				continue
			}
			kind := "changed:"
			if (w != nil && h == nil) || (!isEmptyValue(w) && (isEmptyValue(h) || shorter(h, w))) {
				kind = "lost:"
			}
			out += " " + kind + k
		}
	}
	return out
}

// payloadFields is e as the API receives it, one entry per top-level key.
func payloadFields(e *types.PREvidence) map[string]any {
	raw, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	fields := map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&fields); err != nil {
		panic(err)
	}
	return fields
}

func isEmptyValue(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case json.Number:
		return x == "0"
	case []any:
		return len(x) == 0
	}
	return false
}

func shorter(have, want any) bool {
	h, ok1 := have.([]any)
	w, ok2 := want.([]any)
	return ok1 && ok2 && len(h) < len(w)
}

type jsonKind int

const (
	kindScalar jsonKind = iota
	kindNull
	kindObject
	kindArray
)

// jsonNode keeps object keys in document order, which encoding/json maps lose.
type jsonNode struct {
	kind   jsonKind
	scalar json.RawMessage
	fields []jsonField
	items  []*jsonNode
}

type jsonField struct {
	key   string
	value *jsonNode
}

func parseOrdered(doc []byte) (*jsonNode, error) {
	dec := json.NewDecoder(bytes.NewReader(doc))
	dec.UseNumber()
	n, err := decodeNode(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing data after JSON document")
	}
	return n, nil
}

func decodeNode(dec *json.Decoder) (*jsonNode, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			n := &jsonNode{kind: kindObject}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				value, err := decodeNode(dec)
				if err != nil {
					return nil, err
				}
				n.fields = append(n.fields, jsonField{key: keyTok.(string), value: value})
			}
			_, err := dec.Token()
			return n, err
		case '[':
			n := &jsonNode{kind: kindArray}
			for dec.More() {
				item, err := decodeNode(dec)
				if err != nil {
					return nil, err
				}
				n.items = append(n.items, item)
			}
			_, err := dec.Token()
			return n, err
		}
		return nil, fmt.Errorf("unexpected delimiter %v", v)
	case nil:
		return &jsonNode{kind: kindNull}, nil
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return &jsonNode{kind: kindScalar, scalar: raw}, nil
	}
}

func (n *jsonNode) render() []byte {
	var b bytes.Buffer
	n.writeTo(&b)
	return b.Bytes()
}

func (n *jsonNode) writeTo(b *bytes.Buffer) {
	switch n.kind {
	case kindNull:
		b.WriteString("null")
	case kindScalar:
		b.Write(n.scalar)
	case kindObject:
		b.WriteByte('{')
		for i, f := range n.fields {
			if i > 0 {
				b.WriteByte(',')
			}
			key, _ := json.Marshal(f.key)
			b.Write(key)
			b.WriteByte(':')
			f.value.writeTo(b)
		}
		b.WriteByte('}')
	case kindArray:
		b.WriteByte('[')
		for i, item := range n.items {
			if i > 0 {
				b.WriteByte(',')
			}
			item.writeTo(b)
		}
		b.WriteByte(']')
	}
}
