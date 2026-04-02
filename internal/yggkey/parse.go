package yggkey

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ParsedKey struct {
	Canonical    string    `json:"canonical"`
	KindPath     []string  `json:"kindPath"`
	Scope        []Segment `json:"scope"`
	Principal    *Segment  `json:"principal,omitempty"`
	Root         Segment   `json:"root"`
	Path         []Segment `json:"path,omitempty"`
	TerminalKind string    `json:"terminalKind"`
}

type Segment struct {
	Label string `json:"label"`
	Value string `json:"value,omitempty"`
	Kind  string `json:"kind"`
}

type DerivedKind struct {
	Hierarchy []string `json:"hierarchy"`
}

type parser struct {
	tokens []string
	index  int
}

func Parse(keyID string) (ParsedKey, error) {
	if keyID == "" {
		return ParsedKey{}, fmt.Errorf("invalid key: empty key")
	}

	tokens := strings.Split(keyID, ":")
	if len(tokens) < 2 {
		return ParsedKey{}, fmt.Errorf("invalid key: incomplete key")
	}

	p := parser{tokens: tokens}
	out := ParsedKey{Canonical: keyID}

	level1, err := p.consumeIDLabel("level1", "id", "tenant", "department")
	if err != nil {
		return ParsedKey{}, err
	}
	out.Scope = append(out.Scope, level1)

	if next := p.peek(); next == "group" || next == "team" || next == "region" {
		level2, err := p.consumeIDLabel("level2", "id", "group", "team", "region")
		if err != nil {
			return ParsedKey{}, err
		}
		out.Scope = append(out.Scope, level2)
	}

	if next := p.peek(); next == "user" || next == "member" || next == "subscriber" {
		principal, err := p.consumeIDLabel("principal", "id", "user", "member", "subscriber")
		if err != nil {
			return ParsedKey{}, err
		}
		out.Principal = &principal
	}

	root, err := p.consumeIDLabel("root", "id", "dashboard", "profile")
	if err != nil {
		return ParsedKey{}, err
	}
	out.Root = root

	for p.hasMore() {
		label := p.peek()
		switch label {
		case "note", "comment":
			seg, err := p.consumeIDLabel("path", "id", "note", "comment")
			if err != nil {
				return ParsedKey{}, err
			}
			out.Path = append(out.Path, seg)
		case "like", "language", "thumbnail":
			seg, err := p.consumeBranchLabel(label)
			if err != nil {
				return ParsedKey{}, err
			}
			out.Path = append(out.Path, seg)
		case "text":
			seg, err := p.consumeTerminalLabel("leaf", "text")
			if err != nil {
				return ParsedKey{}, err
			}
			out.Path = append(out.Path, seg)
			out.TerminalKind = seg.Kind
		case "count":
			seg, err := p.consumeTerminalLabel("derived", "count")
			if err != nil {
				return ParsedKey{}, err
			}
			out.Path = append(out.Path, seg)
			out.TerminalKind = seg.Kind
		case "user", "member", "subscriber":
			if len(out.Path) == 0 || out.Path[len(out.Path)-1].Label != "like" {
				return ParsedKey{}, fmt.Errorf("invalid key: principal label %q must follow like", label)
			}
			seg, err := p.consumeIDLabel("like-principal", "id", "user", "member", "subscriber")
			if err != nil {
				return ParsedKey{}, err
			}
			if seg.Value != "_" {
				return ParsedKey{}, fmt.Errorf("invalid key: principal label %q under like must use _", label)
			}
			out.Path = append(out.Path, seg)
			out.TerminalKind = seg.Kind
		case "_":
			if len(out.Path) == 0 {
				return ParsedKey{}, fmt.Errorf("invalid key: alias _ must follow an explicit label")
			}
			prev := out.Path[len(out.Path)-1]
			if prev.Label != "language" && prev.Label != "thumbnail" {
				return ParsedKey{}, fmt.Errorf("invalid key: alias _ is not allowed after %q", prev.Label)
			}
			p.index++
			out.TerminalKind = prev.Kind
		default:
			return ParsedKey{}, fmt.Errorf("invalid key: unsupported label %q", label)
		}

		if out.TerminalKind != "" && p.hasMore() {
			return ParsedKey{}, fmt.Errorf("invalid key: terminal segment %q cannot have children", out.Path[len(out.Path)-1].Label)
		}
	}

	if out.TerminalKind == "" {
		out.TerminalKind = out.Root.Kind
	}
	out.KindPath = append(out.KindPath, labelList(out.Scope)...)
	if out.Principal != nil {
		out.KindPath = append(out.KindPath, out.Principal.Label)
	}
	out.KindPath = append(out.KindPath, out.Root.Label)
	out.KindPath = append(out.KindPath, labelList(out.Path)...)
	return out, nil
}

func MustJSON(parsed ParsedKey) string {
	data, err := json.Marshal(parsed)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (p ParsedKey) DerivedKind() DerivedKind {
	hierarchy := []string{p.Root.Label}
	for _, segment := range p.Path {
		hierarchy = append(hierarchy, segment.Label)
	}
	return DerivedKind{Hierarchy: hierarchy}
}

func labelList(parts []Segment) []string {
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		labels = append(labels, part.Label)
	}
	return labels
}

func (p *parser) hasMore() bool {
	return p.index < len(p.tokens)
}

func (p *parser) peek() string {
	if !p.hasMore() {
		return ""
	}
	return p.tokens[p.index]
}

func (p *parser) consumeIDLabel(scope, kind string, labels ...string) (Segment, error) {
	label := p.peek()
	if !contains(labels, label) {
		return Segment{}, fmt.Errorf("invalid key: expected %s label, got %q", scope, label)
	}
	p.index++
	if !p.hasMore() {
		return Segment{}, fmt.Errorf("invalid key: label %q requires an id value", label)
	}
	value := p.tokens[p.index]
	if value == "" {
		return Segment{}, fmt.Errorf("invalid key: label %q requires a non-empty id value", label)
	}
	p.index++
	return Segment{Label: label, Value: value, Kind: kind}, nil
}

func (p *parser) consumeBranchLabel(label string) (Segment, error) {
	if p.peek() != label {
		return Segment{}, fmt.Errorf("invalid key: expected branch label %q", label)
	}
	p.index++
	if !p.hasMore() {
		return Segment{}, fmt.Errorf("invalid key: branch label %q must be followed by a child", label)
	}
	return Segment{Label: label, Kind: "branch"}, nil
}

func (p *parser) consumeTerminalLabel(kind, label string) (Segment, error) {
	if p.peek() != label {
		return Segment{}, fmt.Errorf("invalid key: expected terminal label %q", label)
	}
	p.index++
	return Segment{Label: label, Kind: kind}, nil
}

func contains(items []string, item string) bool {
	for _, candidate := range items {
		if candidate == item {
			return true
		}
	}
	return false
}
