/**
 * 路由全局前置守卫模块
 *
 * 提供完整的路由导航守卫功能
 *
 * ## 主要功能
 *
 * - 登录状态验证和重定向
 * - 动态路由注册和权限控制
 * - 菜单数据获取和处理（前端/后端模式）
 * - 用户信息获取和缓存
 * - 页面标题设置
 * - 工作标签页管理
 * - 进度条和加载动画控制
 * - 静态路由识别和处理
 * - 错误处理和异常跳转
 *
 * ## 使用场景
 *
 * - 路由跳转前的权限验证
 * - 动态菜单加载和路由注册
 * - 用户登录状态管理
 * - 页面访问控制
 * - 路由级别的加载状态管理
 *
 * ## 工作流程
 *
 * 1. 检查登录状态，未登录跳转到登录页
 * 2. 首次访问时获取用户信息和菜单数据
 * 3. 根据权限动态注册路由
 * 4. 设置页面标题和工作标签页
 * 5. 处理根路径重定向到首页
 * 6. 未匹配路由跳转到 404 页面
 *
 * @module router/guards/beforeEach
 * @author Art Design Pro Team
 */
import type { Router, RouteLocationNormalized, NavigationGuardNext } from 'vue-router'
import { nextTick } from 'vue'
import NProgress from 'nprogress'
import { useSettingStore } from '@/store/modules/setting'
import { useUserStore } from '@/store/modules/user'
import { useMemberStore } from '@/store/modules/member'
import { useMenuStore } from '@/store/modules/menu'
import { setWorktab } from '@/utils/navigation'
import { setPageTitle } from '@/utils/router'
import { ADMIN_BASE_PATH, ADMIN_LOGIN_PATH } from '../routesAlias'
// staticRoutes 不再需要导入（isStaticRoute 已弃用，前后台靠 /admin 前缀隔离）
import { loadingService } from '@/utils/ui'
import { useCommon } from '@/hooks/core/useCommon'
import { useWorktabStore } from '@/store/modules/worktab'
import { useFieldPermStore } from '@/store/modules/fieldPerm'
import { fetchGetUserInfo } from '@/api/backend/auth'
import { ApiStatus } from '@/utils/http/status'
import { isHttpError } from '@/utils/http/error'
import { RouteRegistry, MenuProcessor, IframeRouteManager, RoutePermissionValidator } from '../core'
import type { AppRouteRecord } from '@/types/router'
import { loadFrontendRoutes, isFrontendRoutesLoaded } from '../frontend/loader'

// 路由注册器实例
let routeRegistry: RouteRegistry | null = null

// 菜单处理器实例
const menuProcessor = new MenuProcessor()

// 跟踪是否需要关闭 loading
let pendingLoading = false

// 路由初始化失败标记，防止死循环
// 一旦设置为 true，只有刷新页面或重新登录才能重置
let routeInitFailed = false

// 路由初始化进行中标记，防止并发请求
let routeInitInProgress = false

/**
 * 获取 pendingLoading 状态
 */
export function getPendingLoading(): boolean {
  return pendingLoading
}

/**
 * 重置 pendingLoading 状态
 */
export function resetPendingLoading(): void {
  pendingLoading = false
}

/**
 * 获取路由初始化失败状态
 */
export function getRouteInitFailed(): boolean {
  return routeInitFailed
}

/**
 * 重置路由初始化状态（用于重新登录场景）
 */
export function resetRouteInitState(): void {
  routeInitFailed = false
  routeInitInProgress = false
}

/**
 * 设置路由全局前置守卫
 */
export function setupBeforeEachGuard(router: Router): void {
  // 初始化路由注册器
  routeRegistry = new RouteRegistry(router)

  router.beforeEach(
    async (
      to: RouteLocationNormalized,
      from: RouteLocationNormalized,
      next: NavigationGuardNext
    ) => {
      try {
        await handleRouteGuard(to, from, next, router)
      } catch (error) {
        console.error('[RouteGuard] 路由守卫处理失败:', error)
        closeLoading()
        next({ name: 'Exception500' })
      }
    }
  )
}

/**
 * 关闭 loading 效果
 */
function closeLoading(): void {
  if (pendingLoading) {
    nextTick(() => {
      loadingService.hideLoading()
      pendingLoading = false
    })
  }
}

/**
 * 处理路由守卫逻辑
 */
async function handleRouteGuard(
  to: RouteLocationNormalized,
  from: RouteLocationNormalized,
  next: NavigationGuardNext,
  router: Router
): Promise<void> {
  // 清理历史非版本化用户存储，避免旧 accessToken 被误当作当前登录态。
  localStorage.removeItem('user')

  const settingStore = useSettingStore()
  const userStore = useUserStore()

  // 启动进度条
  if (settingStore.showNprogress) {
    NProgress.start()
  }

  // 1. 检查登录状态
  if (!handleLoginStatus(to, userStore, next)) {
    return
  }

  // 2. 检查路由初始化是否已失败（防止死循环）
  if (routeInitFailed) {
    // 已经失败过，直接放行到错误页面，不再重试
    if (to.matched.length > 0) {
      next()
    } else {
      // 未匹配到路由，跳转到 500 页面
      next({ name: 'Exception500', replace: true })
    }
    return
  }

  // 3. 处理动态路由注册（仅后台 /admin 路由需要）
  if (isAdminRoute(to) && !routeRegistry?.isRegistered() && userStore.isLogin) {
    // 防止并发请求（快速连续导航场景）
    if (routeInitInProgress) {
      // 正在初始化中，等待完成后重新导航
      next(false)
      return
    }
    await handleDynamicRoutes(to, next, router)
    return
  }

  // 4. 处理根路径重定向
  if (handleRootPathRedirect(to, next)) {
    return
  }

  // 5. 前台动态路由：如果命中了全局 404 catch-all，但实际可能是尚未注册的前台页面
  //    先加载前台菜单并注册路由，再重新导航到目标路径
  //    仅对非 admin 路由生效，不影响后台任何逻辑
  if (to.name === 'Exception404' && !isAdminRoute(to) && !isFrontendRoutesLoaded()) {
    loadingService.showLoading()
    try {
      const loaded = await loadFrontendRoutes(router)
      if (loaded) {
        next({ path: to.fullPath, replace: true })
        return
      }
    } finally {
      loadingService.hideLoading()
    }
  }

  // 6. 处理已匹配的路由
  if (to.matched.length > 0) {
    setWorktab(to)
    setPageTitle(to)
    next()
    return
  }

  // 7. 未匹配到路由，跳转到 404
  next({ name: 'Exception404' })
}

/**
 * 检查是否为 Frontend 门户路由（无需后台权限）
 * 前台页面全部在 FrontendLayout (/) 下静态注册，无 catch-all
 */
function isFrontendRoute(to: RouteLocationNormalized): boolean {
  return to.matched.length > 0 && to.matched[0].name === 'FrontendLayout'
}

/**
 * 检查是否为后台管理路由（对齐 BuildAdmin 的 isAdminApp）
 *
 * 设计原则：所有后台路由全部在 /admin 前缀下，只需一个简单的路径前缀判断。
 * 不再需要维护一个冗长的后台路径前缀列表。
 */
function isAdminRoute(to: RouteLocationNormalized): boolean {
  return to.path.startsWith(ADMIN_BASE_PATH)
}

/**
 * 检查路由是否需要登录验证
 */
function requiresAuth(to: RouteLocationNormalized): boolean {
  // 检查 meta 中是否标记需要登录
  return to.matched.some(record => record.meta?.requiresAuth === true)
}

/**
 * 处理登录状态
 * @returns true 表示可以继续，false 表示已处理跳转
 */
function handleLoginStatus(
  to: RouteLocationNormalized,
  userStore: ReturnType<typeof useUserStore>,
  next: NavigationGuardNext
): boolean {
  // 1. Frontend 门户路由特殊处理（使用 memberStore）
  if (isFrontendRoute(to)) {
    const memberStore = useMemberStore()

    // 已登录会员访问登录/注册页时，重定向到用户中心（避免重复登录）
    const memberAuthPages = ['/user/login', '/user/register']
    if (memberAuthPages.includes(to.path) && memberStore.isLogin) {
      const redirect = (to.query.redirect as string) || '/user'
      next({ path: redirect, replace: true })
      return false
    }

    // 如果需要登录验证但会员未登录，跳转到前台登录页
    if (requiresAuth(to) && !memberStore.isLogin) {
      next({
        path: '/user/login',
        query: { redirect: to.fullPath }
      })
      return false
    }
    // 门户路由直接放行（不需要动态路由注册）
    return true
  }

  // 2. 非后台路由（公共错误页等），直接放行
  if (!isAdminRoute(to)) {
    return true
  }

  // ===== 以下逻辑仅处理后台 /admin 路由 =====

  // 3. 后台登录/注册/忘记密码页本身
  if (to.path === ADMIN_LOGIN_PATH || to.path === `${ADMIN_BASE_PATH}/register` || to.path === `${ADMIN_BASE_PATH}/forget-password`) {
    if (userStore.isLogin) {
      // 已登录，重定向到后台首页
      const { homePath } = useCommon()
      const redirect = (to.query.redirect as string) || homePath.value || `${ADMIN_BASE_PATH}/dashboard/console`
      next({ path: redirect, replace: true })
      return false
    }
    // 未登录，放行到登录/注册页
    return true
  }

  // 4. 其他后台页面：已登录放行
  if (userStore.isLogin) {
    return true
  }

  // 5. 确实未登录，跳转到后台登录页
  next({
    name: 'Login',
    query: { redirect: to.fullPath }
  })
  return false
}

// isStaticRoute 已弃用——前后台靠 /admin 路径前缀隔离，不再需要遍历静态路由表判断

/**
 * 处理动态路由注册
 */
async function handleDynamicRoutes(
  to: RouteLocationNormalized,
  next: NavigationGuardNext,
  router: Router
): Promise<void> {
  // 标记初始化进行中
  routeInitInProgress = true

  // 显示 loading
  pendingLoading = true
  loadingService.showLoading()

  try {
    // 1. 获取用户信息
    await fetchUserInfo()

    // 1.5 加载字段权限
    const fieldPermStore = useFieldPermStore()
    if (!fieldPermStore.loaded) {
      await fieldPermStore.loadFieldPerms()
    }

    // 2. 获取菜单数据
    const menuList = await menuProcessor.getMenuList()

    // 3. 验证菜单数据
    if (!menuProcessor.validateMenuList(menuList)) {
      throw new Error('获取菜单列表失败，请重新登录')
    }

    // 4. 注册动态路由（RouteRegistry 内部自动为路由添加 /admin 前缀）
    routeRegistry?.register(menuList)

    // 5. 保存菜单数据到 store（路径也需要加 /admin 前缀，用于侧边栏导航和 WorkTab）
    const menuStore = useMenuStore()
    const prefixedMenuList = addAdminPrefixToMenuList(menuList)
    menuStore.setMenuList(prefixedMenuList)
    menuStore.addRemoveRouteFns(routeRegistry?.getRemoveRouteFns() || [])

    // 6. 保存 iframe 路由
    IframeRouteManager.getInstance().save()

    // 7. 验证工作标签页
    useWorktabStore().validateWorktabs(router)

    // 8. 验证目标路径权限（使用已加前缀的菜单列表，与 to.path 一致）
    const { homePath } = useCommon()
    const { path: validatedPath, hasPermission } = RoutePermissionValidator.validatePath(
      to.path,
      prefixedMenuList,
      homePath.value || `${ADMIN_BASE_PATH}/dashboard/console`
    )

    // 初始化成功，重置进行中标记
    routeInitInProgress = false

    // 9. 重新导航到目标路由
    if (!hasPermission) {
      // 无权限访问，跳转到首页
      closeLoading()

      // 输出警告信息
      console.warn(`[RouteGuard] 用户无权限访问路径: ${to.path}，已跳转到首页`)

      // 直接跳转到首页
      next({
        path: validatedPath,
        replace: true
      })
    } else {
      // 有权限，正常导航
      next({
        path: to.path,
        query: to.query,
        hash: to.hash,
        replace: true
      })
    }
  } catch (error) {
    console.error('[RouteGuard] 动态路由注册失败:', error)

    // 关闭 loading
    closeLoading()

    // 后台初始化时 accessToken 失效：不能 next(false) 回滚到来源页，
    // 否则从门户首页直接访问后台时会被带回前台登录/首页。
    if (isUnauthorizedError(error)) {
      routeInitInProgress = false
      await useUserStore().logOut({ callApi: false, redirect: false })
      next({
        path: ADMIN_LOGIN_PATH,
        query: { redirect: to.fullPath },
        replace: true
      })
      return
    }

    // 标记初始化失败，防止死循环
    routeInitFailed = true
    routeInitInProgress = false

    // 输出详细错误信息，便于排查
    if (isHttpError(error)) {
      console.error(`[RouteGuard] 错误码: ${error.code}, 消息: ${error.message}`)
    }

    // 跳转到 500 页面，使用 replace 避免产生历史记录
    next({ name: 'Exception500', replace: true })
  }
}

/**
 * 获取用户信息
 */
async function fetchUserInfo(): Promise<void> {
  const userStore = useUserStore()
  const data = await fetchGetUserInfo()
  userStore.setUserInfo(data)
  // 检查并清理工作台标签页（如果是不同用户登录）
  userStore.checkAndClearWorktabs()
}

/**
 * 重置路由相关状态
 */
export function resetRouterState(delay: number): void {
  setTimeout(() => {
    routeRegistry?.unregister()
    IframeRouteManager.getInstance().clear()

    const menuStore = useMenuStore()
    menuStore.removeAllDynamicRoutes()
    menuStore.setMenuList([])

    useFieldPermStore().reset()

    // 重置路由初始化状态，允许重新登录后再次初始化
    resetRouteInitState()

    // 前台菜单 Store 在 memberStore.logOut 中清理
  }, delay)
}

/**
 * 处理根路径重定向到首页
 * @returns true 表示已处理跳转，false 表示无需跳转
 */
function handleRootPathRedirect(to: RouteLocationNormalized, next: NavigationGuardNext): boolean {
  // 根路径现在是门户首页，不需要重定向
  // 后台入口通过 /admin 访问
  if (to.path === '/') {
    return false
  }

  return false
}

/**
 * 判断是否为未授权错误（401）
 */
function isUnauthorizedError(error: unknown): boolean {
  return isHttpError(error) && error.code === ApiStatus.unauthorized
}

/**
 * 为菜单列表递归添加 /admin 前缀
 * 用于 menuStore，确保侧边栏导航路径与实际注册路由一致
 */
function addAdminPrefixToMenuList(menuList: AppRouteRecord[]): AppRouteRecord[] {
  return menuList.map((item) => {
    const newItem = { ...item }
    if (newItem.path && newItem.path.startsWith('/') && !newItem.path.startsWith(ADMIN_BASE_PATH)) {
      newItem.path = ADMIN_BASE_PATH + newItem.path
    }
    if (newItem.redirect && typeof newItem.redirect === 'string'
        && newItem.redirect.startsWith('/') && !newItem.redirect.startsWith(ADMIN_BASE_PATH)) {
      newItem.redirect = ADMIN_BASE_PATH + newItem.redirect
    }
    if (newItem.children?.length) {
      newItem.children = addAdminPrefixToMenuList(newItem.children)
    }
    return newItem
  })
}
