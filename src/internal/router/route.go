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
	"strings"
)

// deriveRoute converts relPath (a .ghp path relative to the pages
// directory, using '/' as separator regardless of OS) into the route
// pattern it maps to:
//
//	index.ghp        -> /
//	sobre.ghp         -> /sobre
//	blog/index.ghp    -> /blog
//	blog/[slug].ghp   -> /blog/{slug}
//
// A trailing "index" segment is dropped (it's what makes a folder's own
// route work); any other segment named "[name]" becomes a dynamic
// "{name}" segment, using the same syntax net/http's ServeMux (Go 1.22+)
// already understands - no custom pattern matching needed.
func deriveRoute(relPath string) string {
	segments := pathSegments(relPath)

	if len(segments) > 0 && segments[len(segments)-1] == "index" {
		segments = segments[:len(segments)-1]
	}
	for i, seg := range segments {
		if param, ok := strings.CutPrefix(seg, "["); ok {
			if name, ok := strings.CutSuffix(param, "]"); ok {
				segments[i] = "{" + name + "}"
			}
		}
	}

	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

// deriveFuncName converts the same relative path into a valid, exported
// Go identifier for the page's handler function, e.g.
// "blog/[slug].ghp" -> "BlogSlug". Unlike deriveRoute, it keeps "index"
// segments - the function name should still say what file it came from,
// so "blog.ghp" and "blog/index.ghp" (which do collide on their route,
// see Scan) at least don't also collide on their func name.
func deriveFuncName(relPath string) string {
	var b strings.Builder
	for _, seg := range pathSegments(relPath) {
		seg = strings.Trim(seg, "[]")
		for _, word := range strings.FieldsFunc(seg, isWordBoundary) {
			b.WriteString(strings.ToUpper(word[:1]))
			b.WriteString(word[1:])
		}
	}
	return b.String()
}

// deriveGoFile converts relPath into a flat, valid filename for the
// generated .go file, e.g. "blog/[slug].ghp" -> "blog_slug.go". It
// flattens directories into the name instead of preserving them because
// a Go package can't span subdirectories, and whether generated pages
// share one package or several is a decision for whoever assembles the
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
