package router

import "testing"

func TestDeriveRoute(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{"root", "index.ghp", "/"},
		{"static route", "about.ghp", "/about"},
		{"subfolder index", "blog/index.ghp", "/blog"},
		{"dynamic route", "blog/[slug].ghp", "/blog/{slug}"},
		{"nested without index", "blog/2024/summary.ghp", "/blog/2024/summary"},
		{"multiple dynamic parameters", "store/[category]/[product].ghp", "/store/{category}/{product}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveRoute(tt.relPath); got != tt.want {
				t.Errorf("deriveRoute(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestDeriveFuncName(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{"root", "index.ghp", "Index"},
		{"static route", "about.ghp", "About"},
		{"first multi-byte letter capitalized by rune, not byte", "órgão.ghp", "Órgão"},
		{"subfolder index keeps Index in the name", "blog/index.ghp", "BlogIndex"},
		{"dynamic route without brackets in the name", "blog/[slug].ghp", "BlogSlug"},
		{"hyphen becomes a word boundary", "my-page.ghp", "MyPage"},
		{"underscore becomes a word boundary", "other_page.ghp", "OtherPage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveFuncName(tt.relPath); got != tt.want {
				t.Errorf("deriveFuncName(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestDeriveGoFile(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{"root", "index.ghp", "index.go"},
		{"static route", "about.ghp", "about.go"},
		{"subfolder flattened with underscore", "blog/index.ghp", "blog_index.go"},
		{"brackets removed", "blog/[slug].ghp", "blog_slug.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveGoFile(tt.relPath); got != tt.want {
				t.Errorf("deriveGoFile(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}
