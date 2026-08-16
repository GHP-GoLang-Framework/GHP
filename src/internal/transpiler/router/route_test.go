package router

import "testing"

func TestDerive(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		wantPath string
		wantFunc string
		wantFile string
	}{
		{name: "root", relPath: "index.ghp", wantPath: "/", wantFunc: "Index", wantFile: "index.go"},
		{name: "static route", relPath: "about.ghp", wantPath: "/about", wantFunc: "About", wantFile: "about.go"},
		{name: "subfolder index", relPath: "blog/index.ghp", wantPath: "/blog", wantFunc: "BlogIndex", wantFile: "blog_index.go"},
		{name: "dynamic route", relPath: "blog/[slug].ghp", wantPath: "/blog/{slug}", wantFunc: "BlogSlug", wantFile: "blog_slug.go"},
		{name: "nested without index", relPath: "blog/2024/summary.ghp", wantPath: "/blog/2024/summary", wantFunc: "Blog2024Summary", wantFile: "blog_2024_summary.go"},
		{name: "multiple dynamic parameters", relPath: "store/[category]/[product].ghp", wantPath: "/store/{category}/{product}", wantFunc: "StoreCategoryProduct", wantFile: "store_category_product.go"},
		{name: "hyphen is a word boundary", relPath: "my-page.ghp", wantPath: "/my-page", wantFunc: "MyPage", wantFile: "my-page.go"},
		{name: "underscore is a word boundary", relPath: "other_page.ghp", wantPath: "/other_page", wantFunc: "OtherPage", wantFile: "other_page.go"},
		{name: "first multi-byte letter capitalized by rune, not byte", relPath: "órgão.ghp", wantPath: "/órgão", wantFunc: "Órgão", wantFile: "órgão.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotFunc, gotFile := derive(tt.relPath)
			if gotPath != tt.wantPath {
				t.Errorf("derive(%q) route = %q, want %q", tt.relPath, gotPath, tt.wantPath)
			}
			if gotFunc != tt.wantFunc {
				t.Errorf("derive(%q) func name = %q, want %q", tt.relPath, gotFunc, tt.wantFunc)
			}
			if gotFile != tt.wantFile {
				t.Errorf("derive(%q) go file = %q, want %q", tt.relPath, gotFile, tt.wantFile)
			}
		})
	}
}
