// Package router maps .ghp files inside a directory to HTTP routes,
// file-based-routing style (like PHP or Next.js): the file's path decides
// its URL, with no route registered by hand.
//
// This package never opens or parses the .ghp files themselves - it only
// looks at paths. It also never decides which directory holds pages:
// every function here takes that directory as an explicit argument, so
// whoever calls this package (the ghp CLI, via a flag like --dir) is free
// to point it at any folder, not just one named "pages".
package router

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// deriveRoute converts relPath (a .ghp path relative to the pages
// directory, using '/' as separator regardless of OS) into the route
// pattern it maps to:
//
//	index.ghp        -> /
//	about.ghp        -> /about
//	blog/index.ghp   -> /blog
//	blog/[slug].ghp  -> /blog/{slug}
//
// A trailing "index" segment is dropped (it's what makes a folder's own
// route work); any other segment named "[name]" becomes a dynamic
// "{name}" segment, using the same syntax net/http's ServeMux (Go 1.22+)
// already understands - no custom pattern matching needed.
//
// relPath's segments are assumed already validated (see validateSegment)
// - this function doesn't re-check bracket pairing.
func deriveRoute(relPath string) string {
	segments := pathSegments(relPath)

	if len(segments) > 0 && segments[len(segments)-1] == "index" {
		segments = segments[:len(segments)-1]
	}
	for i, seg := range segments {
		if strings.HasPrefix(seg, "[") {
			segments[i] = "{" + seg[1:len(seg)-1] + "}"
		}
	}

	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

// deriveFuncName converts the same relative path into a Go identifier for
// the page's handler function, e.g. "blog/[slug].ghp" -> "BlogSlug".
// Unlike deriveRoute, it keeps "index" segments - the function name
// should still say what file it came from, so "blog.ghp" and
// "blog/index.ghp" at least don't also collide on their func name just
// because they collide on their route (Scan checks both, along with
// GoFile, since none of the three is allowed to collide - see Scan).
//
// The result isn't guaranteed to be a valid Go identifier on its own
// (e.g. a lone numeric segment like "2024.ghp" produces "2024"); Scan
// checks that before using it.
func deriveFuncName(relPath string) string {
	var b strings.Builder
	for _, seg := range pathSegments(relPath) {
		seg = strings.Trim(seg, "[]")
		for _, word := range strings.FieldsFunc(seg, isWordBoundary) {
			b.WriteString(capitalizeFirst(word))
		}
	}
	return b.String()
}

// capitalizeFirst uppercases word's first rune and leaves the rest of it
// untouched. It operates on runes, not bytes: slicing word[:1] instead
// would split a multi-byte UTF-8 character in half whenever the first
// character isn't plain ASCII (e.g. an accented letter in a Portuguese
// filename like "órgão.ghp"), corrupting it into invalid UTF-8.
func capitalizeFirst(word string) string {
	r, size := utf8.DecodeRuneInString(word)
	return string(unicode.ToUpper(r)) + word[size:]
}

// deriveGoFile converts relPath into a flat filename for the generated
// .go file, e.g. "blog/[slug].ghp" -> "blog_slug.go". It flattens
// directories into the name instead of preserving them because a Go
// package can't span subdirectories, and whether generated pages share
// one package or several is a decision for whoever assembles the
// project, not this package.
func deriveGoFile(relPath string) string {
	segments := pathSegments(relPath)
	for i, seg := range segments {
		segments[i] = strings.Trim(seg, "[]")
	}
	return strings.Join(segments, "_") + ".go"
}

// pathSegments splits relPath into its path segments, stripping the
// ".ghp" extension first.
func pathSegments(relPath string) []string {
	trimmed := strings.TrimSuffix(relPath, ".ghp")
	return strings.Split(trimmed, "/")
}

func isWordBoundary(r rune) bool {
	return r == '-' || r == '_'
}

// validateSegment reports whether seg is well-formed enough to derive a
// route/func name/filename from:
//   - if it starts with '[' it also has to end with ']', and vice versa -
//     a typo like "blog/[slug.ghp" (missing "]") would otherwise leak a
//     literal '[' into the route instead of the "{slug}" the developer
//     almost certainly meant
//   - "[]" (an empty parameter name) isn't allowed either, for the same
//     reason: it would silently become an empty "{}" segment
func validateSegment(seg string) error {
	starts := strings.HasPrefix(seg, "[")
	ends := strings.HasSuffix(seg, "]")

	switch {
	case starts != ends:
		return fmt.Errorf("malformed route segment: %q (unpaired bracket)", seg)
	case starts && len(seg) == 2:
		return fmt.Errorf("malformed route segment: %q (empty parameter name)", seg)
	}
	return nil
}
