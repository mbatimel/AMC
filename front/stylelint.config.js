/** @type {import('stylelint').Config} */
export default {
  extends: ['stylelint-config-standard'],
  plugins: ['stylelint-use-logical'],
  ignoreFiles: [
    '**/node_modules/**',
    '**/.next/**',
    '**/dist/**',
    '**/build/**',
    '**/out/**',
    '**/vi-portal/**',
  ],
  rules: {
    'at-rule-no-unknown': [
      true,
      {
        ignoreAtRules: ['define-mixin', 'mixin', 'mixin-content'],
      },
    ],
    'csstools/use-logical': 'always',
    'import-notation': 'string',
    'media-feature-range-notation': 'context',
    'property-no-unknown': [
      true,
      {
        ignoreProperties: ['composes'],
      },
    ],
    'selector-pseudo-class-no-unknown': [
      true,
      {
        ignorePseudoClasses: ['global'],
      },
    ],
  },
  overrides: [
    {
      files: ['src/**/*.module.css'],
      rules: {
        'selector-class-pattern': [
          '^[a-z][a-zA-Z0-9]*$',
          {
            message: 'Expected class selector to be camelCase',
          },
        ],
      },
    },
    {
      files: ['src/core/shared/styles/mixins/**/*.css'],
      rules: {
        'nesting-selector-no-missing-scoping-root': null,
      },
    },
    {
      files: ['src/core/shared/styles/normalize.css'],
      rules: {
        'csstools/use-logical': null,
      },
    },
  ],
};
