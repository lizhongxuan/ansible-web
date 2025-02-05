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
    // 添加一些宽松的规则，避免开发时出现太多警告
    'vue/no-unused-components': 'warn',
    'no-unused-vars': 'warn',
    'vue/multi-word-component-names': 'off'
  },
  ignorePatterns: ['dist/', 'node_modules/']
} 