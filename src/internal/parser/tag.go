package parser

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// tagKind identifies which GHP tag begins at a given `<` in the source
type tagKind int

const (
	//tagNone means the `<` does not start a GHP tag - it's plain markup
	// (HTML, or a false match like "<google-map>") and should be treated
	// as ordinary text.
	tagNone tagKind = iota

	tagImport
	tagStatement
	tagEcho
	tagIf
	tagElse
	tagSwitch
	tagCase
	tagDefault
	tagFor
	tagCloseIf
	tagCloseSwitch
	tagCloseFor
)

var tagHeads = []struct {
	open string
	kind tagKind
}{
	{"go:import", tagImport},
	{"go:if", tagIf},
	{"go:else", tagElse},
	{"go:switch", tagSwitch},
	{"go:case", tagCase},
	{"go:default", tagDefault},
	{"go:for", tagFor},
	{"go=", tagEcho},
	{"go", tagStatement},
}

var closeTagHeads = []struct {
	open string
	kind tagKind
}{
	{"go:if", tagCloseIf},
	{"go:switch", tagCloseSwitch},
	{"go:for", tagCloseFor},
}

// matchTagHead looks at s (the text right after a `<`) and reports which
// tag it opens, plus how many bytes the head itself takes up — the caller
// resumes scanning right after that many bytes for the tag's payload.
// It returns (tagNone, 0) when s isn't a GHP tag at all.
func matchTagHead(s string) (kind tagKind, headLen int) {
	if after, ok := strings.CutPrefix(s, "/"); ok {
		for _, h := range closeTagHeads {
			if rest, ok := strings.CutPrefix(after, h.open); ok && boundary(rest) {
				return h.kind, 1 + len(h.open)
			}
		}
		return tagNone, 0
	}

	for _, h := range tagHeads {
		rest, ok := strings.CutPrefix(s, h.open)
		if !ok {
			continue
		}
		matched := boundary(rest)
		if h.kind == tagStatement {
			matched = statementBoundary(rest)
		}
		if matched {
			return h.kind, len(h.open)
		}
	}
	return tagNone, 0
}

// boundary reports whether rest may legally follow a tag head that already
// ends on a fixed keyword (e.g. "go:if", "go="). The head only matches if
// rest doesn't continue the identifier — otherwise "go:iffy" would be read
// as "go:if" plus garbage.
func boundary(rest string) bool {
	if rest == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return !isIdentRune(r)
}

// statementBoundary is the same check for the bare "go" statement head,
// which additionally must not be followed by ':' or '=' — those belong to
// "go:..." and "go=" respectively, not to a statement.
func statementBoundary(rest string) bool {
	if rest == "" {
		return true
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return !isIdentRune(r) && r != ':' && r != '='
}

func isIdentRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
