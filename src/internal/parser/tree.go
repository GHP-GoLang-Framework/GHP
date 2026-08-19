package parser

import (
	"fmt"
	"ghp/src/internal/ast"
	"strings"
)

var tagKindNames = map[tagKind]string{
	tagNone:        "text",
	tagImport:      "<go:import/>",
	tagStatement:   "<go/>",
	tagEcho:        "<go=/>",
	tagIf:          "<go:if/>",
	tagElse:        "<go:else/>",
	tagElif:        "<go:elif/>",
	tagSwitch:      "<go:switch/>",
	tagCase:        "<go:case/>",
	tagDefault:     "<go:default/>",
	tagFor:         "<go:for/>",
	tagCloseIf:     "<go:endif/>",
	tagCloseSwitch: "<go:endswitch/>",
	tagCloseFor:    "<go:endfor/>",
}

func (k tagKind) String() string {
	if name, ok := tagKindNames[k]; ok {
		return name
	}
	return "unknown tag"
}

// parseNodes parses a sequence of nodes until EOF or one of the stop kinds
// is reached. The stop tag itself is consumed from the scanner but not
// turned into a node - it's returned to the caller instead, so callers
// like buildIf can tell a <go:else/> apart from a <go:endif/>. A nil stop
// map means "run to EOF" (used at the top level), in which case the
// returned token is always the zero tagToken.
func parseNodes(s *scanner, stop map[tagKind]bool) ([]ast.Node, tagToken, error) {
	var nodes []ast.Node

	for {
		if text, ok := s.nextText(); ok {
			nodes = append(nodes, text)
		}

		if s.eof() {
			return nodes, tagToken{}, nil
		}

		tok, err := s.nextTagToken()
		if err != nil {
			return nil, tagToken{}, err
		}

		if stop[tok.kind] {
			return nodes, tok, nil
		}

		node, err := buildNode(s, tok)
		if err != nil {
			return nil, tagToken{}, err
		}
		nodes = append(nodes, node)
	}
}

// buildNode consumes the body (and any nested tags) belonging to tok and
// returns the finished node. tok has already been read from s; buildNode
// is responsible for reading everything up to and including tok's
// matching close tag, for tags that have one.
func buildNode(s *scanner, tok tagToken) (ast.Node, error) {
	switch tok.kind {
	case tagImport:
		paths, err := parseImportPaths(tok.payload)
		if err != nil {
			return nil, &SyntaxError{Line: tok.line, Message: err.Error()}
		}
		return ast.NewImport(paths, tok.line), nil

	case tagStatement:
		return ast.NewStatement(tok.payload, tok.line), nil

	case tagEcho:
		return ast.NewEcho(tok.payload, tok.line), nil

	case tagIf:
		return buildIf(s, tok)

	case tagSwitch:
		return buildSwitch(s, tok)

	case tagFor:
		body, close, err := parseNodes(s, map[tagKind]bool{tagCloseFor: true})
		if err != nil {
			return nil, err
		}
		if close.kind == tagNone {
			return nil, &SyntaxError{Line: tok.line, Message: "<go:for> missing <go:endfor/>"}
		}
		return ast.NewFor(tok.payload, body, tok.line), nil

	default:
		return nil, &SyntaxError{Line: tok.line, Message: fmt.Sprintf("unexpected %s here", tok.kind)}
	}
}

// buildIf handles a <go:if> that already had its head read (tok). It reads
// the "then" body up to <go:elif/>, <go:else/>, or <go:endif/>; when it
// stops at <go:elif/>, it collects the elif condition and body, then loops
// until a non-elif stop tag is reached.  Elifs, Else stay nil (not empty
// slices) when there are no <go:elif/> or <go:else/> tags respectively,
// matching ast.If's contract.
func buildIf(s *scanner, tok tagToken) (ast.Node, error) {
	stop := map[tagKind]bool{tagElif: true, tagElse: true, tagCloseIf: true}

	then, close, err := parseNodes(s, stop)
	if err != nil {
		return nil, err
	}
	if close.kind == tagNone {
		return nil, &SyntaxError{Line: tok.line, Message: "<go:if> missing <go:endif/>"}
	}

	var elifs []ast.ElseIf
	for close.kind == tagElif {
		if close.payload == "" {
			return nil, &SyntaxError{Line: close.line, Message: "<go:elif> requires a condition"}
		}
		elifBody, after, err := parseNodes(s, stop)
		if err != nil {
			return nil, err
		}
		elifs = append(elifs, ast.ElseIf{Cond: close.payload, Body: elifBody, Line: close.line})
		close = after
	}

	if close.kind != tagElse && close.kind != tagCloseIf {
		return nil, &SyntaxError{Line: tok.line, Message: "<go:if> missing <go:endif/>"}
	}

	var els []ast.Node
	if close.kind == tagElse {
		if close.payload != "" {
			return nil, &SyntaxError{Line: close.line, Message: "<go:else> does not take a condition - chained else-if isn't supported, use nested <go:if> instead"}
		}

		els, close, err = parseNodes(s, map[tagKind]bool{tagCloseIf: true})
		if err != nil {
			return nil, err
		}
		if close.kind == tagNone {
			return nil, &SyntaxError{Line: tok.line, Message: "<go:if> missing <go:endif/>"}
		}
	}

	return ast.NewIf(tok.payload, then, elifs, els, tok.line), nil
}

// buildSwitch handles a <go:switch> that already had its head read (tok).
// Its body is a run of <go:case/>/<go:default/> blocks up to
// <go:endswitch/>; any text sitting directly between <go:switch> and the
// first case (e.g. stray whitespace) is discarded rather than kept, since
// GHP's grammar doesn't give it anywhere to live on the Switch node.
func buildSwitch(s *scanner, tok tagToken) (ast.Node, error) {
	sw := ast.NewSwitch(tok.payload, nil, nil, tok.line)
	stop := map[tagKind]bool{tagCase: true, tagDefault: true, tagCloseSwitch: true}

	_, next, err := parseNodes(s, stop)
	if err != nil {
		return nil, err
	}

	sawDefault := false
	for next.kind != tagCloseSwitch {
		switch next.kind {
		case tagCase:
			caseTok := next
			body, after, err := parseNodes(s, stop)
			if err != nil {
				return nil, err
			}
			sw.Cases = append(sw.Cases, ast.Case{Value: caseTok.payload, Body: body, Line: caseTok.line})
			next = after

		case tagDefault:
			if sawDefault {
				return nil, &SyntaxError{Line: next.line, Message: "duplicate <go:default>"}
			}
			sawDefault = true
			body, after, err := parseNodes(s, stop)
			if err != nil {
				return nil, err
			}
			sw.Default = body
			next = after

		default:
			return nil, &SyntaxError{Line: tok.line, Message: "<go:switch> missing <go:endswitch/>"}
		}
	}
	return sw, nil
}

// parseImportPaths splits a <go:import (...)/> payload into its individual
// paths. GHP mirrors Go's own import syntax: parentheses hold one path per
// line (or comma-separated), each optionally prefixed with an alias.
func parseImportPaths(payload string) ([]ast.ImportPath, error) {
	inner := strings.TrimSpace(payload)
	inner = strings.TrimPrefix(inner, "(")
	inner = strings.TrimSuffix(inner, ")")

	var paths []ast.ImportPath
	for _, line := range strings.FieldsFunc(inner, func(r rune) bool {
		return r == '\n' || r == ','
	}) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var alias, path string
		if idx := strings.LastIndexByte(line, ' '); idx != -1 {
			alias = line[:idx]
			path = strings.TrimSpace(line[idx+1:])
		} else {
			path = line
		}

		paths = append(paths, ast.ImportPath{Alias: alias, Path: strings.Trim(path, `"`)})
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("empty <go:import>")
	}
	return paths, nil
}
