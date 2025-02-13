module.exports = {
  root: true,
  env: {
    node: true,
    browser: true,
    es6: true
  },
  extends: [
    'plugin:vue/essential',
    'eslint:recommended'
  ],
  parserOptions: {
    parser: '@babel/eslint-parser',
    ecmaVersion: 2020,
    sourceType: 'module',
    requireConfigFile: false
  },
  rules: {
    'no-console': 'off',
    'no-debugger': 'off',
    'vue/no-unused-components': 'warn',
    'no-unused-vars': 'warn',
    'vue/multi-word-component-names': 'off'
  }
} 