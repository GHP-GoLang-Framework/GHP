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

// derive computes a page's route, handler name and filename from its
// relative .ghp path, in one pass over the path segments.
//
//	index.ghp        -> "/",               "Index",    "index.go"
//	about.ghp        -> "/about",          "About",    "about.go"
//	blog/index.ghp   -> "/blog",           "BlogIndex", "blog_index.go"
//	blog/[slug].ghp  -> "/blog/{slug}",    "BlogSlug", "blog_slug.go"
//
// A trailing "index" segment is dropped from the route (it's the folder's
// own page); a "[name]" segment becomes "{name}", matching ServeMux (Go
// 1.22+). The func name capitalizes each word and keeps index segments so
// blog.ghp and blog/index.ghp collide only on route, not also on name;
// it isn't guaranteed to be a valid identifier (e.g. "2024.ghp" -> "2024")
// - Scan checks that before using it.
func derive(relPath string) (route, funcName, goFile string) {
	segs := pathSegments(relPath)

	var routeSegs []string
	for i, seg := range segs {
		name := strings.Trim(seg, "[]")

		if i > 0 {
			goFile += "_"
		}
		goFile += name

		for _, word := range strings.FieldsFunc(name, isWordBoundary) {
			funcName += capitalizeFirst(word)
		}

		if seg == "index" && i == len(segs)-1 {
			continue
		}
		if strings.HasPrefix(seg, "[") {
			seg = "{" + seg[1:len(seg)-1] + "}"
		}
		routeSegs = append(routeSegs, seg)
	}

	if len(routeSegs) == 0 {
		route = "/"
	} else {
		route = "/" + strings.Join(routeSegs, "/")
	}
	goFile += ".go"
	return route, funcName, goFile
}

// capitalizeFirst uppercases word's first rune, leaving the rest alone.
// It works on runes, not bytes, so a non-ASCII first letter (e.g. the
// "ó" in "órgão.ghp") isn't split into invalid UTF-8.
func capitalizeFirst(word string) string {
	r, size := utf8.DecodeRuneInString(word)
	return string(unicode.ToUpper(r)) + word[size:]
}

// isWordBoundary marks the characters that split a segment into words,
// so "my-page.ghp" and "my_page.ghp" both derive the func name "MyPage".
func isWordBoundary(r rune) bool {
	return r == '-' || r == '_'
}

// pathSegments splits relPath into segments, dropping the ".ghp" suffix
// first.
func pathSegments(relPath string) []string {
	return strings.Split(strings.TrimSuffix(relPath, ".ghp"), "/")
}

// validateSegment reports whether seg is safe to derive a route from: an
// opening '[' must be paired with a closing ']' (a typo like "[slug.ghp"
// would otherwise leak a literal '[' into the route), and "[]" is rejected
// because it would silently become an empty "{}" segment.
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
