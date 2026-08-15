package router

import (
	"fmt"
	"go/token"
	"io/fs"
	"path/filepath"
)

// Page describes one .ghp file discovered by Scan: the route it maps to,
// the name its handler function will have, and the filename its
// generated Go source will be written to.
type Page struct {
	GhpPath  string // .ghp path relative to the scanned directory, always with '/'
	Route    string // route pattern, e.g. "/blog/{slug}"
	FuncName string // handler function name, e.g. "BlogSlug"
	GoFile   string // generated .go filename, e.g. "blog_slug.go"
}

// ConflictError reports that two different .ghp files derived the same
// value for one of Route, FuncName or GoFile - any of the three
// colliding breaks the site (a duplicate route shadows a page, a
// duplicate func name fails to compile, a duplicate filename overwrites
// a sibling's generated source), so Scan checks all three, not just
// Route.
type ConflictError struct {
	What          string // full, grammatically correct phrase, e.g. "the same route"
	Value         string
	First, Second string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s and %s share %s: %q", e.First, e.Second, e.What, e.Value)
}

// Scan walks dir recursively for *.ghp files and derives a Page for each
// one, following the convention documented on deriveRoute. dir is
// whatever directory the caller wants scanned (e.g. the ghp CLI's --dir
// flag) - this package has no opinion on what it's named or where it
// lives, and never falls back to a default of its own.
//
// It returns a *ConflictError if two different files derive the same
// Route, FuncName or GoFile, and a plain error if a derived FuncName
// isn't a valid Go identifier (e.g. a page named "2024.ghp" on its own,
// with no other path segment to give it a leading letter).
func Scan(dir string) ([]Page, error) {
	var pages []Page
	byRoute := make(map[string]string)
	byFuncName := make(map[string]string)
	byGoFile := make(map[string]string)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".ghp" {
			return nil
		}

		// path always comes from walking dir itself, so it's always
		// relative to it - this error is unreachable in practice, kept
		// only because filepath.Rel returns one.
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		for _, seg := range pathSegments(rel) {
			if err := validateSegment(seg); err != nil {
				return fmt.Errorf("%s: %w", rel, err)
			}
		}

		page := Page{
			GhpPath:  rel,
			Route:    deriveRoute(rel),
			FuncName: deriveFuncName(rel),
			GoFile:   deriveGoFile(rel),
		}

		if !token.IsIdentifier(page.FuncName) {
			return fmt.Errorf("%s: invalid function name derived from file: %q", rel, page.FuncName)
		}

		if other, ok := byRoute[page.Route]; ok {
			return &ConflictError{What: "the same route", Value: page.Route, First: other, Second: page.GhpPath}
		}
		if other, ok := byFuncName[page.FuncName]; ok {
			return &ConflictError{What: "the same function name", Value: page.FuncName, First: other, Second: page.GhpPath}
		}
		if other, ok := byGoFile[page.GoFile]; ok {
			return &ConflictError{What: "the same .go file", Value: page.GoFile, First: other, Second: page.GhpPath}
		}
		byRoute[page.Route] = page.GhpPath
		byFuncName[page.FuncName] = page.GhpPath
		byGoFile[page.GoFile] = page.GhpPath

		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pages, nil
}
