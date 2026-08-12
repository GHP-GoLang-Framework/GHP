module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat',     // Nova funcionalidade
        'fix',      // Correção de bug
        'docs',     // Documentação
        'style',    // Formatação, gofmt, etc.
        'refactor', // Refatoração de código
        'perf',     // Melhorias de performance
        'test',     // Adição ou ajuste de testes
        'build',    // Mudanças no sistema de build (go.mod, Dockerfile, Makefile)
        'ci',       // Mudanças em arquivos de CI/CD
        'chore',    // Outras mudanças que não modificam src ou test
        'revert',   // Reverter um commit anterior
      ],
    ],
    'type-case': [2, 'always', 'lower-case'],
    'type-empty': [2, 'never'],
    'subject-empty': [2, 'never'],
    'subject-full-stop': [2, 'never', '.'],
    'header-max-length': [2, 'always', 100],
  },
};
