package router

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writePages creates one empty file per path (relative to dir) under
// dir, creating any parent directories needed. Scan only looks at paths,
// never at file contents, so an empty file is enough to test it.
func writePages(t *testing.T, dir string, relPaths ...string) {
	t.Helper()
	for _, rel := range relPaths {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", full, err)
		}
	}
}

func TestScanDerivesPagesForEachFile(t *testing.T) {
	dir := t.TempDir()
	writePages(t, dir, "index.ghp", "sobre.ghp", "blog/[slug].ghp")

	pages, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(pages) != 3 {
		t.Fatalf("len(pages) = %d, want 3: %+v", len(pages), pages)
	}

	byGhpPath := make(map[string]Page)
	for _, p := range pages {
		byGhpPath[p.GhpPath] = p
	}

	if got := byGhpPath["index.ghp"]; got.Route != "/" || got.FuncName != "Index" {
		t.Errorf("index.ghp = %+v", got)
	}
	if got := byGhpPath["sobre.ghp"]; got.Route != "/sobre" || got.FuncName != "Sobre" {
		t.Errorf("sobre.ghp = %+v", got)
	}
	if got := byGhpPath["blog/[slug].ghp"]; got.Route != "/blog/{slug}" || got.FuncName != "BlogSlug" {
		t.Errorf("blog/[slug].ghp = %+v", got)
	}
}

func TestScanIgnoresNonGhpFiles(t *testing.T) {
	dir := t.TempDir()
	writePages(t, dir, "index.ghp")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	pages, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("len(pages) = %d, want 1: %+v", len(pages), pages)
	}
}

func TestScanEmptyDir(t *testing.T) {
	pages, err := Scan(t.TempDir())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("len(pages) = %d, want 0", len(pages))
	}
}

func TestScanPropagatesWalkErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("rodando como root, permissao 0000 nao bloqueia a leitura")
	}

	dir := t.TempDir()
	blocked := filepath.Join(dir, "sem-permissao")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	writePages(t, blocked, "index.ghp")
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(blocked, 0o755) })

	if _, err := Scan(dir); err == nil {
		t.Fatal("Scan() = nil error, want erro de permissao propagado")
	}
}

func TestScanDetectsRouteConflict(t *testing.T) {
	// blog.ghp -> /blog e blog/index.ghp -> /blog: duas paginas
	// diferentes, mesma rota - consequencia direta da convencao de
	// dropar "index", nao um bug isolado.
	dir := t.TempDir()
	writePages(t, dir, "blog.ghp", "blog/index.ghp")

	_, err := Scan(dir)
	if err == nil {
		t.Fatal("Scan() = nil error, want conflito de rota")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("erro = %T, want *ConflictError", err)
	}
	if conflict.Value != "/blog" {
		t.Errorf("Value = %q, want %q", conflict.Value, "/blog")
	}

	// filepath.WalkDir visita "blog/" (diretorio) antes de "blog.ghp" no
	// mesmo nivel - "blog" e prefixo de "blog.ghp", entao vem primeiro
	// na ordem lexical - e desce nela por completo antes de continuar,
	// entao "blog/index.ghp" e descoberto primeiro.
	want := `blog/index.ghp e blog.ghp compartilham a mesma rota: "/blog"`
	if conflict.Error() != want {
		t.Errorf("Error() = %q, want %q", conflict.Error(), want)
	}
}

func TestScanDetectsFuncNameConflict(t *testing.T) {
	// blog-post.ghp e blog_post.ghp derivam rotas diferentes
	// (/blog-post e /blog_post, sem conflito ali), mas hifen e
	// underscore viram a mesma fronteira de palavra em deriveFuncName -
	// as duas colidem em "BlogPost". Sem checar FuncName tambem, o
	// segundo <go:import> gerado nunca compilaria (func redeclarada).
	dir := t.TempDir()
	writePages(t, dir, "blog-post.ghp", "blog_post.ghp")

	_, err := Scan(dir)
	if err == nil {
		t.Fatal("Scan() = nil error, want conflito de nome de funcao")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("erro = %T, want *ConflictError", err)
	}
	if conflict.Value != "BlogPost" {
		t.Errorf("Value = %q, want %q", conflict.Value, "BlogPost")
	}
}

func TestScanDetectsGoFileConflict(t *testing.T) {
	// blog_post.ghp (raiz) e blog/post.ghp (subpasta) derivam rotas
	// diferentes (/blog_post e /blog/post, sem conflito ali), mas
	// deriveGoFile achata os dois pro mesmo "blog_post.go" - sem checar
	// GoFile tambem, um dos dois arquivos gerados sobrescreveria o
	// outro em silencio. Esse mesmo par tambem colide em FuncName (as
	// duas derivacoes tratam "_" como fronteira de palavra do mesmo
	// jeito, entao um "_" dentro de um segmento e indistinguivel de uma
	// barra separando dois segmentos) - Scan reporta o primeiro
	// conflito que encontra, que aqui e o de nome de funcao. O ponto
	// deste teste nao e qual campo especifico e citado, e sim que o
	// par e detectado como conflito - antes desta checagem, nenhum dos
	// dois disparava erro nenhum.
	dir := t.TempDir()
	writePages(t, dir, "blog_post.ghp", "blog/post.ghp")

	_, err := Scan(dir)
	if err == nil {
		t.Fatal("Scan() = nil error, want conflito (func name ou go file)")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("erro = %T, want *ConflictError", err)
	}
}

func TestScanRejectsInvalidFuncName(t *testing.T) {
	// "2024.ghp" sozinho (sem outro segmento de caminho pra dar uma
	// letra inicial) deriva FuncName "2024" - um literal numerico
	// valido em Go, mas nao um identificador valido. Sem essa checagem,
	// Register geraria "mux.HandleFunc(\"/2024\", 2024)": sintaxe Go
	// valida, mas semanticamente quebrada, so descoberta la na frente
	// com `go build`.
	dir := t.TempDir()
	writePages(t, dir, "2024.ghp")

	if _, err := Scan(dir); err == nil {
		t.Fatal("Scan() = nil error, want erro de nome de funcao invalido")
	}
}

func TestScanRejectsMalformedDynamicSegment(t *testing.T) {
	tests := []string{
		"blog/[slug.ghp", // falta o ']'
		"blog/slug].ghp", // falta o '['
		"blog/[].ghp",    // nome do parametro vazio
	}

	for _, rel := range tests {
		t.Run(rel, func(t *testing.T) {
			dir := t.TempDir()
			writePages(t, dir, rel)

			if _, err := Scan(dir); err == nil {
				t.Fatalf("Scan() = nil error para %q, want erro de segmento mal formado", rel)
			}
		})
	}
}
