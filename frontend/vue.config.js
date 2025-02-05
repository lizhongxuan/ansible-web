const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    port: 3000  // 修改前端端口为3000，避免与后端冲突
  },
  lintOnSave: false
}) 