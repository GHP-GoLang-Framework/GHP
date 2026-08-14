package router

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// Page describes one .ghp file discovered by Scan: the route it maps to,
// the name its handler function will have, and the filename its
// generated Go source will be written to.
type Page struct {
	GhpPath  string // caminho do .ghp relativo ao diretorio escaneado, sempre com '/'
	Route    string // padrao de rota, ex.: "/blog/{slug}"
	FuncName string // nome da funcao handler, ex.: "BlogSlug"
	GoFile   string // nome do .go a ser gerado, ex.: "blog_slug.go"
}

// RouteConflictError reports that two different .ghp files derived the
// same route.
type RouteConflictError struct {
	Route         string
	First, Second string
}

func (e *RouteConflictError) Error() string {
	return fmt.Sprintf("rota %q definida por duas paginas: %s e %s", e.Route, e.First, e.Second)
}

// Scan walks dir recursively for *.ghp files and derives a Page for each
// one, following the convention documented on deriveRoute. dir is
// whatever directory the caller wants scanned (e.g. the ghp CLI's --dir
// flag) - this package has no opinion on what it's named or where it
// lives, and never falls back to a default of its own.
//
// It returns a *RouteConflictError if two different files map to the
// same route.
func Scan(dir string) ([]Page, error) {
	var pages []Page
	byRoute := make(map[string]string)

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

		page := Page{
			GhpPath:  rel,
			Route:    deriveRoute(rel),
			FuncName: deriveFuncName(rel),
			GoFile:   deriveGoFile(rel),
		}

		if other, ok := byRoute[page.Route]; ok {
			return &RouteConflictError{Route: page.Route, First: other, Second: page.GhpPath}
		}
		byRoute[page.Route] = page.GhpPath

		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pages, nil
}
