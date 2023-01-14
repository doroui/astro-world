module.exports = {
  env: {
    browser: true,
    es2021: true,
  },
  extends: ['eslint:recommended', 'plugin:react/recommended'],
  overrides: [],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  plugins: ['react'],
  rules: {
    'react/prop-types': 'off', // Enable again
    'prefer-arrow-callback': 'error',
    'arrow-body-style': 'error',
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
};
