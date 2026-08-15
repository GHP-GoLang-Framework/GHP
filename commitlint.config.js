module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat',     // New feature
        'fix',      // Bug fix
        'docs',     // Documentation
        'style',    // Formatting, gofmt, etc.
        'refactor', // Code refactoring
        'perf',     // Performance improvements
        'test',     // Adding or adjusting tests
        'build',    // Changes to the build system (go.mod, Dockerfile, Makefile)
        'ci',       // Changes to CI/CD files
        'chore',    // Other changes that do not touch src or test
        'revert',   // Revert a previous commit
      ],
    ],
    'type-case': [2, 'always', 'lower-case'],
    'type-empty': [2, 'never'],
    'subject-empty': [2, 'never'],
    'subject-full-stop': [2, 'never', '.'],
    'header-max-length': [2, 'always', 100],
  },
};
