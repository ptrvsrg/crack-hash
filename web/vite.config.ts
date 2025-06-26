import { defineConfig, UserConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig(({ _, command }) => {
  const config: UserConfig = {
    build: {
      target: 'esnext',
    },
    plugins: [vue(), vueDevTools()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    css: {
      preprocessorOptions: {
        less: {
          javascriptEnabled: true,
        },
      },
    },
    server: {
      watch: {
        useFsEvents: true,
      },
      warmup: {
        clientFiles: ['./src/**/*.vue', './src/**/*.ts', './src/**/*.css', './src/**/*.svg'],
      },
      open: true,
      host: true,
      port: 3000,
      cors: {
        origin: 'http://localhost:8080',
      },
    },
    define:
      command === 'serve'
        ? {
            global: {},
          }
        : undefined,
  }

  return config
})
