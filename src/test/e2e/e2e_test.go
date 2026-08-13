//go:build e2e

// Package e2e roda o binário `ghp` compilado contra projetos de exemplo
// (ghp init/dev/build) e valida respostas HTTP reais.
package e2e

import "testing"

func TestPlaceholder(t *testing.T) {
	t.Skip("sem testes e2e ainda — ver GHP-14 / GHP-15 no Linear")
}
