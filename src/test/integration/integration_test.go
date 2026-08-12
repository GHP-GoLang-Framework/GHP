//go:build integration

// Package integration reúne testes que exercitam múltiplos pacotes juntos
// (parser+codegen+project) ou chamam o `go build` de verdade, sem precisar
// do binário `ghp` compilado.
package integration

import "testing"

func TestPlaceholder(t *testing.T) {
	t.Skip("sem testes de integração ainda — ver GHP-13 no Linear")
}
