package router

import "testing"

func TestDeriveRoute(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{"raiz", "index.ghp", "/"},
		{"rota estatica", "sobre.ghp", "/sobre"},
		{"index de subpasta", "blog/index.ghp", "/blog"},
		{"rota dinamica", "blog/[slug].ghp", "/blog/{slug}"},
		{"aninhada sem index", "blog/2024/resumo.ghp", "/blog/2024/resumo"},
		{"multiplos parametros dinamicos", "loja/[categoria]/[produto].ghp", "/loja/{categoria}/{produto}"},
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
		{"raiz", "index.ghp", "Index"},
		{"rota estatica", "sobre.ghp", "Sobre"},
		{"index de subpasta mantem Index no nome", "blog/index.ghp", "BlogIndex"},
		{"rota dinamica sem colchetes no nome", "blog/[slug].ghp", "BlogSlug"},
		{"hifen vira fronteira de palavra", "minha-pagina.ghp", "MinhaPagina"},
		{"underscore vira fronteira de palavra", "outra_pagina.ghp", "OutraPagina"},
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
		{"raiz", "index.ghp", "index.go"},
		{"rota estatica", "sobre.ghp", "sobre.go"},
		{"subpasta achatada com underscore", "blog/index.ghp", "blog_index.go"},
		{"colchetes removidos", "blog/[slug].ghp", "blog_slug.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveGoFile(tt.relPath); got != tt.want {
				t.Errorf("deriveGoFile(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}
