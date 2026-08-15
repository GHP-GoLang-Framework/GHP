// Teste da gramática TextMate do GHP.
//
// Carrega ghp.tmLanguage.json no mesmo motor que o VSCode usa (vscode-textmate
// + oniguruma) e verifica os escopos atribuídos a cada trecho. Sem isso, a
// única forma de validar a gramática seria abrir o editor e olhar as cores.
//
// `source.go` e `text.html.basic` são gramáticas built-in do VSCode, não
// empacotadas nesta extensão. Aqui elas são registradas como stubs vazios —
// isso NÃO é cosmético: se um `include` não resolve, o vscode-textmate
// descarta a regra inteira que o contém (a tag deixa de ser reconhecida por
// completo, sem degradação). Os stubs reproduzem o ambiente real, onde os
// dois escopos existem, e mantêm o teste focado no que é nosso: o
// reconhecimento das tags GHP e a marcação das regiões
// `meta.embedded.block.go` que fazem o VSCode injetar o Go de verdade.

const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert');

const oniguruma = require('vscode-oniguruma');
const textmate = require('vscode-textmate');

const GRAMMAR_PATH = path.join(__dirname, '..', 'syntaxes', 'ghp.tmLanguage.json');

async function loadGrammar() {
  const wasmPath = path.join(
    require.resolve('vscode-oniguruma'),
    '..',
    '..',
    'release',
    'onig.wasm',
  );
  await oniguruma.loadWASM(fs.readFileSync(wasmPath).buffer);

  const registry = new textmate.Registry({
    onigLib: Promise.resolve({
      createOnigScanner: (patterns) => new oniguruma.OnigScanner(patterns),
      createOnigString: (s) => new oniguruma.OnigString(s),
    }),
    loadGrammar: async (scopeName) => {
      if (scopeName === 'source.ghp') {
        const raw = fs.readFileSync(GRAMMAR_PATH, 'utf8');
        return textmate.parseRawGrammar(raw, GRAMMAR_PATH);
      }
      // Stubs das gramáticas built-in do VSCode — ver comentário no topo.
      if (scopeName === 'source.go' || scopeName === 'text.html.basic') {
        return { scopeName, patterns: [] };
      }
      return null;
    },
  });

  const grammar = await registry.loadGrammar('source.ghp');
  assert.ok(grammar, 'gramática source.ghp não carregou');
  return grammar;
}

// tokenize roda a gramática sobre `source` e devolve [{text, scopes}], com o
// estado de linha encadeado — necessário para tags multi-linha como <go .../>.
function tokenize(grammar, source) {
  const out = [];
  let ruleStack = textmate.INITIAL;

  for (const line of source.split('\n')) {
    const result = grammar.tokenizeLine(line, ruleStack);
    for (const token of result.tokens) {
      const text = line.substring(token.startIndex, token.endIndex);
      if (text.trim() === '') continue;
      // O texto é comparado sem as bordas em branco: a gramática costuma
      // devolver o conteúdo Go junto com o espaço que o separa da tag
      // (ex.: `<go:if cond/>` produz " cond"), e esse espaço não é relevante.
      out.push({ text: text.trim(), scopes: token.scopes });
    }
    ruleStack = result.ruleStack;
  }
  return out;
}

// assertScope encontra o primeiro token cujo texto casa e exige que ele tenha
// o escopo esperado.
function assertScope(tokens, text, expectedScope) {
  const token = tokens.find((t) => t.text === text);
  assert.ok(token, `token ${JSON.stringify(text)} não foi encontrado`);
  assert.ok(
    token.scopes.includes(expectedScope),
    `token ${JSON.stringify(text)} tem escopos [${token.scopes.join(', ')}], esperado incluir ${expectedScope}`,
  );
}

// assertEmbeddedGo exige que o trecho esteja marcado como Go embutido — é essa
// marca que o VSCode usa para colorir o conteúdo como Go de verdade.
function assertEmbeddedGo(tokens, text) {
  assertScope(tokens, text, 'meta.embedded.block.go');
}

const tests = {
  'reconhece <go:import> e marca o conteúdo como Go'(grammar) {
    const tokens = tokenize(grammar, '<go:import ("fmt")/>');
    assertScope(tokens, 'go:import', 'keyword.control.import.ghp');
    assertEmbeddedGo(tokens, '("fmt")');
  },

  'reconhece <go= ...> como echo'(grammar) {
    const tokens = tokenize(grammar, '<title><go= expression /></title>');
    assertScope(tokens, 'go=', 'keyword.control.echo.ghp');
    assertEmbeddedGo(tokens, 'expression');
  },

  'reconhece bloco <go .../> multi-linha'(grammar) {
    const tokens = tokenize(grammar, '<go\n    items := []string{"a"}\n/>');
    assertScope(tokens, 'go', 'keyword.control.ghp');
    assertEmbeddedGo(tokens, 'items := []string{"a"}');
  },

  'reconhece go:if / go:else / fechamento'(grammar) {
    const tokens = tokenize(
      grammar,
      '<go:if variable == value/>\n  x\n<go:else/>\n  y\n<go:endif/>',
    );
    assertScope(tokens, 'go:if', 'keyword.control.ghp');
    assertEmbeddedGo(tokens, 'variable == value');
    assertScope(tokens, 'go:else', 'keyword.control.ghp');
    // O fechamento é uma tag própria com o mesmo escopo de controle.
    assertScope(tokens, 'go:endif', 'keyword.control.ghp');
  },

  'reconhece switch/case/default'(grammar) {
    const tokens = tokenize(
      grammar,
      '<go:switch variable/>\n<go:case value/>\n<go:default/>\n<go:endswitch/>',
    );
    assertScope(tokens, 'go:switch', 'keyword.control.ghp');
    assertScope(tokens, 'go:case', 'keyword.control.ghp');
    assertScope(tokens, 'go:default', 'keyword.control.ghp');
    assertEmbeddedGo(tokens, 'variable');
  },

  'reconhece go:for'(grammar) {
    const tokens = tokenize(grammar, '<go:for _, item := range items/>\n<go:endfor/>');
    assertScope(tokens, 'go:for', 'keyword.control.ghp');
    assertEmbeddedGo(tokens, '_, item := range items');
  },

  'não confunde tag HTML que começa com "go" com bloco Go'(grammar) {
    const tokens = tokenize(grammar, '<google-maps zoom="3"></google-maps>');
    const wrong = tokens.find((t) => t.scopes.includes('meta.tag.statement.ghp'));
    assert.ok(
      !wrong,
      `<google-maps> foi tratado como bloco Go (token ${JSON.stringify(wrong?.text)})`,
    );
  },

  'não trata go:import como tag de controle'(grammar) {
    const tokens = tokenize(grammar, '<go:import ("fmt")/>');
    const token = tokens.find((t) => t.text === 'go:import');
    assert.ok(
      !token.scopes.includes('meta.tag.control.ghp'),
      'go:import caiu no padrão de tag de controle',
    );
  },

  'processa docs/template.ghp inteiro sem cair para texto puro'(grammar) {
    const templatePath = path.join(__dirname, '..', '..', '..', 'docs', 'template.ghp');
    const source = fs.readFileSync(templatePath, 'utf8');
    const tokens = tokenize(grammar, source);

    // Toda tag presente no template de referência precisa ser reconhecida.
    for (const tag of [
      'go:import',
      'go=',
      'go:if',
      'go:else',
      'go:endif',
      'go:switch',
      'go:case',
      'go:default',
      'go:endswitch',
      'go:for',
      'go:endfor',
    ]) {
      const found = tokens.find(
        (t) => t.text === tag && t.scopes.some((s) => s.startsWith('keyword.control')),
      );
      assert.ok(found, `tag ${tag} do template de referência não foi reconhecida`);
    }
  },
};

(async () => {
  const grammar = await loadGrammar();
  let failed = 0;

  for (const [name, fn] of Object.entries(tests)) {
    try {
      fn(grammar);
      console.log(`  ok  ${name}`);
    } catch (err) {
      failed++;
      console.error(`FAIL  ${name}\n      ${err.message}`);
    }
  }

  const total = Object.keys(tests).length;
  console.log(`\n${total - failed}/${total} testes passaram`);
  process.exit(failed === 0 ? 0 : 1);
})();
