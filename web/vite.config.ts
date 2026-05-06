import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'
import vueDevTools from 'vite-plugin-vue-devtools'
import viteCompression from 'vite-plugin-compression'
import Components from 'unplugin-vue-components/vite'
import AutoImport from 'unplugin-auto-import/vite'
import ElementPlus from 'unplugin-element-plus/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import tailwindcss from '@tailwindcss/vite'
// import { visualizer } from 'rollup-plugin-visualizer'

export default ({ mode }: { mode: string }) => {
  const root = process.cwd()
  const env = loadEnv(mode, root)
  const { VITE_VERSION, VITE_PORT, VITE_BASE_URL, VITE_API_URL, VITE_API_PROXY_URL, VITE_OUT_DIR } = env

  console.log(`🚀 API_URL = ${VITE_API_URL}`)
  console.log(`🚀 VERSION = ${VITE_VERSION}`)

  return defineConfig({
    define: {
      __APP_VERSION__: JSON.stringify(VITE_VERSION)
    },
    base: VITE_BASE_URL,
    server: {
      port: Number(VITE_PORT),
      // Windows 下外部进程（如 Go 代码生成器）批量创建/覆盖文件时，
      // 原生 fs 事件偶发会丢失或合并，启用轮询可显著提升稳定性。
      watch: {
        usePolling: process.platform === 'win32',
        interval: 120,
        awaitWriteFinish: {
          stabilityThreshold: 180,
          pollInterval: 40
        }
      },
      proxy: {
        '/admin': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        },
        // 静态资源代理（上传的文件）
        '/attachment': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        },
        // 前台会员接口代理
        '/member': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        },
        // 公共验证码接口代理
        '/captcha': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        },
        // 站点公开接口代理
        '/site': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        },
        // 系统公开接口代理
        '/system': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        },
        // WebSocket 代理
        '/socket': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true,
          ws: true
        }
      },
      host: true
    },
    // 路径别名
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        // BuildAdmin 风格别名
        '@api': resolvePath('src/api'),
        '@backend': resolvePath('src/views/backend'),
        '@frontend': resolvePath('src/views/frontend'),
        '@modules': resolvePath('src/modules'),
        // 原有别名
        '@views': resolvePath('src/views'),
        '@imgs': resolvePath('src/assets/images'),
        '@icons': resolvePath('src/assets/icons'),
        '@utils': resolvePath('src/utils'),
        '@stores': resolvePath('src/stores'),
        '@styles': resolvePath('src/assets/styles'),
        '@addons': resolvePath('src/addons')
      }
    },
    build: {
      target: 'es2015',
      outDir: VITE_OUT_DIR || 'dist',
      emptyOutDir: true,
      chunkSizeWarningLimit: 2000,
      cssMinify: 'esbuild',
      minify: 'terser',
      terserOptions: {
        compress: {
          drop_console: true,
          drop_debugger: true
        }
      },
      rollupOptions: {
        output: {
          manualChunks: {
            'vendor-vue': ['vue', 'vue-router', 'pinia'],
            'vendor-element': ['element-plus'],
            'vendor-echarts': ['echarts/core', 'echarts/charts', 'echarts/components', 'echarts/renderers'],
            'vendor-utils': ['axios', 'crypto-js', 'xlsx', 'file-saver'],
            'vendor-editor': ['md-editor-v3'],
          }
        }
      },
      dynamicImportVarsOptions: {
        warnOnError: true,
        exclude: [],
        include: ['src/views/**/*.vue']
      }
    },
    plugins: [
      vue(),
      tailwindcss(),
      // 自动按需导入 API
      AutoImport({
        imports: ['vue', 'vue-router', 'pinia', '@vueuse/core'],
        dts: 'src/types/import/auto-imports.d.ts',
        resolvers: [ElementPlusResolver()],
        eslintrc: {
          enabled: true,
          filepath: './.auto-import.json',
          globalsPropValue: true
        }
      }),
      // 自动按需导入组件
      Components({
        dts: 'src/types/import/components.d.ts',
        resolvers: [ElementPlusResolver()]
      }),
      // 按需定制主题配置
      ElementPlus({
        useSource: true
      }),
      // 压缩
      viteCompression({
        verbose: false, // 是否在控制台输出压缩结果
        disable: false, // 是否禁用
        algorithm: 'gzip', // 压缩算法
        ext: '.gz', // 压缩后的文件名后缀
        threshold: 10240, // 只有大小大于该值的资源会被处理 10240B = 10KB
        deleteOriginFile: false // 压缩后是否删除原文件
      }),
      vueDevTools()
      // 打包分析
      // visualizer({
      //   open: true,
      //   gzipSize: true,
      //   brotliSize: true,
      //   filename: 'dist/stats.html' // 分析图生成的文件名及路径
      // }),
    ],
    // 依赖预构建：避免运行时重复请求与转换，提升首次加载速度
    optimizeDeps: {
      include: [
        'echarts/core',
        'echarts/charts',
        'echarts/components',
        'echarts/renderers',
        'xlsx',
        'xgplayer',
        'crypto-js',
        'file-saver',
        'vue-img-cutter',
        'element-plus/es',
        'element-plus/es/components/*/style/css',
        'element-plus/es/components/*/style/index'
      ]
    },
    css: {
      preprocessorOptions: {
        // sass variable and mixin
        scss: {
          additionalData: `
            @use "@styles/core/el-light.scss" as *; 
            @use "@styles/core/mixin.scss" as *;
          `
        }
      },
      postcss: {
        plugins: [
          {
            postcssPlugin: 'internal:charset-removal',
            AtRule: {
              charset: (atRule) => {
                if (atRule.name === 'charset') {
                  atRule.remove()
                }
              }
            }
          }
        ]
      }
    }
  })
}

function resolvePath(paths: string) {
  return path.resolve(__dirname, paths)
}
