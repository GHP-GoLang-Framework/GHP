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
// value for one of Route or FuncName - either collision breaks the site
// (a duplicate route shadows a page, a duplicate func name fails to
// compile), so Scan checks both, not just Route.
type ConflictError struct {
	What          string // full, grammatically correct phrase, e.g. "the same route"
	Value         string
	First, Second string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s and %s share %s: %q", e.First, e.Second, e.What, e.Value)
}

// site tracks the routes and func names seen so far, rejecting collisions
// as pages are added - the two checks Scan has to run for every file.
type site struct {
	routes map[string]string // route -> first .ghp path that claimed it
	funcs  map[string]string // func name -> first .ghp path that claimed it
	pages  []Page
}

// add records page, or fails if another page already claimed its route
// or func name. A derived GoFile is deliberately not checked: it's a
// function of the same joined segments that define FuncName, so any GoFile
// collision is necessarily a FuncName collision and is caught above.
func (s *site) add(page Page) error {
	if other, ok := s.routes[page.Route]; ok {
		return &ConflictError{What: "the same route", Value: page.Route, First: other, Second: page.GhpPath}
	}
	if other, ok := s.funcs[page.FuncName]; ok {
		return &ConflictError{What: "the same function name", Value: page.FuncName, First: other, Second: page.GhpPath}
	}
	s.routes[page.Route] = page.GhpPath
	s.funcs[page.FuncName] = page.GhpPath
	s.pages = append(s.pages, page)
	return nil
}

// Scan walks dir recursively for *.ghp files and derives a Page for each
// one, following the convention documented on derive. dir is whatever
// directory the caller wants scanned (e.g. the ghp CLI's --dir flag) -
// this package has no opinion on what it's named or where it lives.
//
// It returns a *ConflictError if two different files derive the same
// Route or FuncName, and a plain error if a derived FuncName isn't a
// valid Go identifier (e.g. a page named "2024.ghp" on its own).
func Scan(dir string) ([]Page, error) {
	s := &site{routes: make(map[string]string), funcs: make(map[string]string)}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".ghp" {
			return nil
		}

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

		page := Page{GhpPath: rel}
		page.Route, page.FuncName, page.GoFile = derive(rel)

		if !token.IsIdentifier(page.FuncName) {
			return fmt.Errorf("%s: invalid function name derived from file: %q", rel, page.FuncName)
		}

		return s.add(page)
	})
	if err != nil {
		return nil, err
	}

	return s.pages, nil
}
