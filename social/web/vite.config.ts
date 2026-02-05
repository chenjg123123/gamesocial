import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  // 关键配置：适配腾讯云托管的静态资源路径
  base: './', // 相对路径（推荐），适配所有部署路径；如果部署到根路径也可以用 '/'
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  build: {
    // 确保打包产物目录正确（默认dist，和你之前的构建配置匹配）
    outDir: 'dist',
    // 可选：修复静态资源文件名哈希可能的路径问题
    assetsDir: 'assets'
  }
})
