// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

/**
 * 组件加载器
 *
 * 负责动态加载 Vue 组件
 * 支持 BuildAdmin 风格的 backend/frontend/common 目录结构
 *
 * @module router/core/ComponentLoader
 * @author Art Design Pro Team
 */

import { h } from 'vue'

export class ComponentLoader {
  private modules: Record<string, () => Promise<any>>
  private addonModules: Record<string, () => Promise<any>>

  constructor() {
    // 动态导入 views 目录下所有 .vue 组件
    this.modules = import.meta.glob('../../views/**/*.vue')
    // 动态导入 addons 目录下所有 .vue 组件
    this.addonModules = import.meta.glob('../../addons/*/views/**/*.vue')
  }

  // 需要映射到 /common 目录的路径前缀
  private commonPrefixes = ['/exception', '/result']

  /**
   * 加载组件
   * 支持以下路径格式：
   * - /backend/system/user -> views/backend/system/user/index.vue
   * - /system/user -> views/backend/system/user/index.vue (自动添加 backend 前缀)
   * - /common/exception/404 -> views/common/exception/404/index.vue
   * - /exception/404 -> views/common/exception/404/index.vue (自动映射到 common)
   */
  load(componentPath: string): () => Promise<any> {
    if (!componentPath) {
      return this.createEmptyComponent()
    }

    // 特殊别名：Layout
    if (componentPath === 'Layout') {
      return this.loadLayout()
    }

    // 扩展组件：@addons/ 开头
    if (componentPath.startsWith('@addons/') || componentPath.startsWith('/addons/')) {
      return this.loadAddonComponent(componentPath)
    }

    // 尝试直接加载
    let module = this.tryLoadComponent(componentPath)
    if (module) return module

    // 如果路径不以 /backend、/frontend、/common 开头
    if (!componentPath.startsWith('/backend') && 
        !componentPath.startsWith('/frontend') && 
        !componentPath.startsWith('/common')) {
      
      // 检查是否是需要映射到 /common 的路径
      const isCommonPath = this.commonPrefixes.some(prefix => componentPath.startsWith(prefix))
      
      if (isCommonPath) {
        // 映射到 /common 目录
        module = this.tryLoadComponent(`/common${componentPath}`)
        if (module) return module
      } else {
        // 映射到 /backend 目录
        module = this.tryLoadComponent(`/backend${componentPath}`)
        if (module) return module
      }
    }

    console.error(`[ComponentLoader] 未找到组件: ${componentPath}`)
    return this.createErrorComponent(componentPath)
  }

  /**
   * 尝试加载组件
   */
  private tryLoadComponent(componentPath: string): (() => Promise<any>) | null {
    // 构建可能的路径
    const paths = [
      `../../views${componentPath}.vue`,
      `../../views${componentPath}/index.vue`
    ]

    for (const path of paths) {
      if (this.modules[path]) {
        return this.modules[path]
      }
    }

    return null
  }

  /**
   * 加载布局组件
   */
  loadLayout(): () => Promise<any> {
    // 优先使用 backend/index，兼容旧路径
    const backendLayout = this.modules['../../views/backend/index/index.vue']
    if (backendLayout) {
      return backendLayout
    }
    return () => import('@/views/backend/index/index.vue')
  }

  /**
   * 加载 iframe 组件
   */
  loadIframe(): () => Promise<any> {
    // 优先使用 common/outside，兼容旧路径
    const commonIframe = this.modules['../../views/common/outside/Iframe.vue']
    if (commonIframe) {
      return commonIframe
    }
    return () => import('@/views/common/outside/Iframe.vue')
  }

  /**
   * 创建空组件
   */
  private createEmptyComponent(): () => Promise<any> {
    return () =>
      Promise.resolve({
        render() {
          return h('div', {})
        }
      })
  }

  /**
   * 加载扩展组件
   * 路径格式: @addons/{name}/views/xxx 或 /addons/{name}/views/xxx
   */
  private loadAddonComponent(componentPath: string): () => Promise<any> {
    const cleaned = componentPath.replace(/^[@/]addons\//, '')
    const paths = [
      `../../addons/${cleaned}.vue`,
      `../../addons/${cleaned}/index.vue`
    ]

    for (const path of paths) {
      if (this.addonModules[path]) {
        return this.addonModules[path]
      }
    }

    console.error(`[ComponentLoader] 未找到扩展组件: ${componentPath}`)
    return this.createErrorComponent(componentPath)
  }

  /**
   * 创建错误提示组件
   */
  private createErrorComponent(componentPath: string): () => Promise<any> {
    return () =>
      Promise.resolve({
        render() {
          return h('div', { class: 'route-error' }, `组件未找到: ${componentPath}`)
        }
      })
  }
}
