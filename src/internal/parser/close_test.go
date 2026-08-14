package parser

import "testing"

func TestFindTagClose(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{
			// Aspas com escape: o '"' escapado (\") nao pode ser lido
			// como o fim da string, senao o '>' logo depois dele
			// fecharia a tag cedo demais.
			name: "aspas com escape",
			src:  ` x == "a\"b" >resto`,
			want: len(` x == "a\"b" `), // ate o '>' logo apos a aspa de fechamento real
		},
		{
			name: "parenteses protegem o > interno",
			src:  ` (a > b) >resto`,
			want: len(` (a > b) `),
		},
		{
			// Ao contrario de '(' e '[', '{' nao pausa o fechamento -
			// um <go ...> pode abrir uma chave que so fecha numa tag
			// posterior (ver comentario em findTagClose), entao o '>'
			// logo depois dela tem que fechar a tag normalmente.
			name: "chave nao protege o > interno",
			src:  ` if x {>resto`,
			want: len(` if x {`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findTagClose(tt.src, 0)
			if got != tt.want {
				t.Errorf("findTagClose(%q, 0) = %d, want %d", tt.src, got, tt.want)
			}
		})
	}
}
