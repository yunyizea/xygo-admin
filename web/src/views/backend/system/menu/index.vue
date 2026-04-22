<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 菜单管理页面 -->
<template>
  <div class="menu-page art-full-height">
    <!-- 搜索栏 -->
    <ArtSearchBar
      v-model="formFilters"
      :items="formItems"
      :showExpand="false"
      @reset="handleReset"
      @search="handleSearch"
    />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader
        :showZebra="false"
        :loading="loading"
        v-model:columns="columnChecks"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElButton @click="handleAddMenu" v-ripple>添加菜单</ElButton>
          <ElButton @click="toggleExpand" v-ripple type="primary">
            {{ isExpanded ? '收起' : '展开' }}
          </ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        rowKey="id"
        :loading="loading"
        :columns="columns"
        :data="filteredTableData"
        :stripe="false"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
      />

      <!-- 菜单弹窗 -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :editData="editData"
        :lockType="lockMenuType"
        :parentMenu="parentMenu"
        :menuTree="tableData"
        @submit="handleSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { formatMenuTitle } from '@/utils/router'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import type { AppRouteRecord } from '@/types/router'
  import MenuDialog from './modules/menu-dialog.vue'
  import { fetchGetMenuTree, fetchSaveMenu, fetchDeleteMenu } from '@/api/backend/system'
  import { ElTag, ElMessageBox, ElSwitch } from 'element-plus'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'Menus' })

  // 状态管理
  const loading = ref(false)
  const isExpanded = ref(false)
  const tableRef = ref()

  // 弹窗相关
  const dialogVisible = ref(false)
  const dialogType = ref<'directory' | 'menu' | 'button'>('menu')
  const editData = ref<AppRouteRecord | any>(null)
  const lockMenuType = ref(false)
  const parentMenu = ref<any>(null) // 当前操作的父菜单

  // 搜索相关
  const initialSearchState = {
    name: '',
    route: ''
  }

  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })

  const formItems = computed(() => [
    {
      label: '菜单名称',
      key: 'name',
      type: 'input',
      props: { clearable: true }
    },
    {
      label: '路由地址',
      key: 'route',
      type: 'input',
      props: { clearable: true }
    }  
  ])

  onMounted(() => {
    getMenuList()
  })

  /**
   * 获取菜单列表数据
   */
  const getMenuList = async (): Promise<void> => {
    loading.value = true

    try {
      const list = await fetchGetMenuTree()  // ✅ 改为获取菜单树（管理后台配置）
      tableData.value = list
    } catch (error) {
      throw error instanceof Error ? error : new Error('获取菜单失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取菜单类型标签颜色（已废弃，改用直接映射）
   * @param row 菜单行数据
   * @returns 标签颜色类型
   */
  const getMenuTypeTag = (
    row: any
  ): 'primary' | 'success' | 'warning' | 'info' | 'danger' => {
    const typeMap: Record<number, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
      1: 'info',
      2: 'primary',
      3: 'danger'
    }
    return typeMap[row.type] || 'info'
  }

  /**
   * 获取菜单类型文本（已废弃，改用直接映射）
   * @param row 菜单行数据
   * @returns 菜单类型文本
   */
  const getMenuTypeText = (row: any): string => {
    const typeMap: Record<number, string> = {
      1: '目录',
      2: '菜单',
      3: '按钮'
    }
    return typeMap[row.type] || '未知'
  }

  // 表格列配置
  const { columnChecks, columns } = useTableColumns(() => [
    {
      prop: 'title',
      label: '菜单名称',
     
      formatter: (row: any) => formatMenuTitle(row.title)
    },
    {
      prop: 'icon',
      label: '图标',
      
      align: 'center',
      formatter: (row: any) => {
        if (!row.icon) return '-'
        return h(ArtSvgIcon, { 
          icon: row.icon,
          style: 'font-size: 24px;'
        })
      }
    },
    {
      prop: 'type',
      label: '菜单类型',
      
      formatter: (row: any) => {
        const typeMap: Record<number, { text: string; type: 'info' | 'primary' | 'danger' }> = {
          1: { text: '目录', type: 'info' },
          2: { text: '菜单', type: 'primary' },
          3: { text: '按钮', type: 'danger' }
        }
        const config = typeMap[row.type] || { text: '未知', type: 'info' }
        return h(ElTag, { type: config.type }, () => config.text)
      }
    },
    {
      prop: 'path',
      label: '路由',
      
      formatter: (row: any) => {
        if (row.type === 3) return '' // 按钮类型无路由
        return row.frameSrc || row.path || ''
      }
    },
    {
      prop: 'perms',
      label: '权限标识',
      
      formatter: (row: any) => {
        if (!row.perms || !row.perms.trim()) return ''
        
        // 如果是 JSON 数组格式，显示数量
        if (row.perms.trim().startsWith('[')) {
          try {
            const parsed = JSON.parse(row.perms)
            if (Array.isArray(parsed) && parsed.length > 0) {
              return `${parsed.length} 个权限`
            }
          } catch (e) {
            // ignore
          }
        }
        
        // 如果是简单字符串格式，直接显示（截断过长内容）
        const permsStr = row.perms.toString().trim()
        return permsStr.length > 30 ? permsStr.substring(0, 30) + '...' : permsStr
      }
    },
    {
      prop: 'update_time',
      label: '更新时间',
      
      formatter: (row: any) => formatTimestamp(row.update_time)
    },
    {
      prop: 'status',
      label: '是否启用',
      
      align: 'center',
      formatter: (row: any) =>
        h(ElSwitch, {
          modelValue: row.status === 1,
          activeColor: '#13ce66',
          inactiveColor: '#ff4949',
          onChange: (val) => handleToggleField(row, 'status', val ? 1 : 0)
        })
    },
    {
      prop: 'hidden',
      label: '隐藏菜单',
      
      align: 'center',
      formatter: (row: any) =>
        h(ElSwitch, {
          modelValue: row.hidden === 1,
          activeColor: '#13ce66',
          inactiveColor: '#ff4949',
          onChange: (val) => handleToggleField(row, 'hidden', val ? 1 : 0)
        })
    },
    // {
    //   prop: 'keepAlive',
    //   label: '页面缓存',
    //   width: 90,
    //   align: 'center',
    //   formatter: (row: any) =>
    //     h(ElSwitch, {
    //       modelValue: row.keepAlive === 1,
    //       disabled: true,
    //       activeColor: '#13ce66',
    //       inactiveColor: '#ff4949'
    //     })
    // },
    {
      prop: 'operation',
      label: '操作',
      width: 180,
      align: 'right',
      formatter: (row: any) => {
        const buttonStyle = { style: 'text-align: right' }

        // 按钮类型（type = 3）
        if (row.type === 3) {
          return h('div', buttonStyle, [
            h(ArtButtonTable, {
              type: 'edit',
              onClick: () => handleEditAuth(row)
            }),
            h(ArtButtonTable, {
              type: 'delete',
              onClick: () => handleDeleteAuth(row)
            })
          ])
        }

        return h('div', buttonStyle, [
          h(ArtButtonTable, {
            type: 'add',
            onClick: () => handleAddAuth(row),
            title: '添加子级'
          }),
          h(ArtButtonTable, {
            type: 'edit',
            onClick: () => handleEditMenu(row)
          }),
          h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDeleteMenu(row)
          })
        ])
      }
    }
  ])

  // 数据相关（后端原始格式）
  const tableData = ref<any[]>([])

  /**
   * 重置搜索条件
   */
  const handleReset = (): void => {
    Object.assign(formFilters, { ...initialSearchState })
    Object.assign(appliedFilters, { ...initialSearchState })
    getMenuList()
  }

  /**
   * 执行搜索
   */
  const handleSearch = (): void => {
    Object.assign(appliedFilters, { ...formFilters })
    getMenuList()
  }

  /**
   * 刷新菜单列表
   */
  const handleRefresh = (): void => {
    getMenuList()
  }

  /**
   * 深度克隆对象
   * @param obj 要克隆的对象
   * @returns 克隆后的对象
   */
  const deepClone = <T,>(obj: T): T => {
    if (obj === null || typeof obj !== 'object') return obj
    if (obj instanceof Date) return new Date(obj) as T
    if (Array.isArray(obj)) return obj.map((item) => deepClone(item)) as T

    const cloned = {} as T
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        cloned[key] = deepClone(obj[key])
      }
    }
    return cloned
  }

  /**
   * 将权限列表转换为子节点（后端数据格式）
   * @param items 菜单项数组
   * @returns 转换后的菜单项数组
   */
  const convertAuthListToChildren = (items: any[]): any[] => {
    return items.map((item) => {
      const clonedItem = deepClone(item)

      if (clonedItem.children?.length) {
        clonedItem.children = convertAuthListToChildren(clonedItem.children)
      }

      return clonedItem
    })
  }

  /**
   * 搜索菜单（后端数据格式）
   * @param items 菜单项数组
   * @returns 搜索结果数组
   */
  const searchMenu = (items: any[]): any[] => {
    const results: any[] = []

    for (const item of items) {
      const searchName = appliedFilters.name?.toLowerCase().trim() || ''
      const searchRoute = appliedFilters.route?.toLowerCase().trim() || ''
      const menuTitle = formatMenuTitle(item.title || '').toLowerCase()
      const menuPath = (item.path || '').toLowerCase()
      const nameMatch = !searchName || menuTitle.includes(searchName)
      const routeMatch = !searchRoute || menuPath.includes(searchRoute)

      if (item.children?.length) {
        const matchedChildren = searchMenu(item.children)
        if (matchedChildren.length > 0) {
          const clonedItem = deepClone(item)
          clonedItem.children = matchedChildren
          results.push(clonedItem)
          continue
        }
      }

      if (nameMatch && routeMatch) {
        results.push(deepClone(item))
      }
    }

    return results
  }

  // 过滤后的表格数据
  const filteredTableData = computed(() => {
    const searchedData = searchMenu(tableData.value)
    return convertAuthListToChildren(searchedData)
  })

  /**
   * 添加顶级菜单
   */
  const handleAddMenu = (): void => {
    dialogType.value = 'menu'
    editData.value = null
    parentMenu.value = null
    lockMenuType.value = false // ✅ 允许切换类型
    dialogVisible.value = true
  }

  /**
   * 添加子菜单
   * @param row 父菜单行数据
   */
  const handleAddAuth = (row: any): void => {
    dialogType.value = 'menu'
    editData.value = null
    parentMenu.value = row
    lockMenuType.value = false
    dialogVisible.value = true
  }

  /**
   * 编辑菜单
   * @param row 菜单行数据
   */
  const handleEditMenu = (row: any): void => {
    dialogType.value = row.type === 1 ? 'directory' : 'menu'
    
    // ✅ 后端返回的是parentId（驼峰），不是parent_id
    const parentIdValue = row.parentId || row.parent_id || 0
    
    // 如果有上级菜单，需要找到上级菜单数据
    if (parentIdValue > 0) {
      const findParent = (list: any[], pid: number): any => {
        for (const item of list) {
          if (item.id === pid) return item
          if (item.children) {
            const found = findParent(item.children, pid)
            if (found) return found
          }
        }
        return null
      }
      parentMenu.value = findParent(tableData.value, parentIdValue)
    } else {
      parentMenu.value = null
    }
    
    // 转换后端数据为前端表单格式
    editData.value = {
      id: row.id,
      parentId: parentIdValue > 0 ? parentIdValue : null,
      path: row.path,
      name: row.name,
      component: row.component,
      resource: row.resource || '',
      perms: row.perms || '',
      sort: row.sort,
      meta: {
        title: row.title,
        icon: row.icon,
        isHide: row.hidden === 1,
        isHideTab: row.hideTab === 1,
        keepAlive: row.keepAlive === 1,
        fixedTab: row.affix === 1,
        link: row.frameSrc,
        isIframe: row.isFrame === 1,
        showBadge: row.showBadge === 1,
        showTextBadge: row.badgeText,
        activePath: row.activePath,
        isFullPage: row.isFullPage === 1,
        isEnable: row.status === 1,
        sort: row.sort
      }
    }
    lockMenuType.value = true
    dialogVisible.value = true
  }

  /**
   * 编辑权限按钮
   * @param row 权限行数据
   */
  const handleEditAuth = (row: any): void => {
    dialogType.value = 'button'
    const parentIdValue = row.parentId || row.parent_id || 0
    editData.value = {
      id: row.id,
      parentId: parentIdValue > 0 ? parentIdValue : null,
      title: row.title,
      authMark: row.name,
      perms: row.perms,
      sort: row.sort,
      keepAlive: row.keepAlive
    }
    lockMenuType.value = false
    dialogVisible.value = true
  }

  /**
   * 处理开关状态变化
   * @param row 菜单行数据
   * @param field 字段名
   * @param value 新值
   */
  const handleToggleField = async (row: any, field: string, newValue: number): Promise<void> => {
    const oldValue = row[field]
    try {
      row[field] = newValue

      const findOriginal = (list: any[], id: number): any => {
        for (const item of list) {
          if (item.id === id) return item
          if (item.children) {
            const found = findOriginal(item.children, id)
            if (found) return found
          }
        }
        return null
      }
      const original = findOriginal(tableData.value, row.id)
      if (original) original[field] = newValue

      await fetchSaveMenu({
        id: row.id,
        parentId: row.parentId || row.parent_id || 0,
        type: row.type,
        title: row.title,
        name: row.name,
        path: row.path || '',
        component: row.component || '',
        icon: row.icon || '',
        resource: row.resource || '',
        hidden: field === 'hidden' ? newValue : (row.hidden || 0),
        hideTab: row.hideTab || 0,
        keepAlive: row.keepAlive || 0,
        redirect: row.redirect || '',
        frameSrc: row.frameSrc || '',
        perms: row.perms || '',
        isFrame: row.isFrame || 0,
        affix: row.affix || 0,
        showBadge: row.showBadge || 0,
        badgeText: row.badgeText || '',
        activePath: row.activePath || '',
        isFullPage: row.isFullPage || 0,
        sort: row.sort || 1,
        status: field === 'status' ? newValue : (row.status || 0),
        remark: row.remark || ''
      })

      ElMessage.success('更新成功')
    } catch (error) {
      row[field] = oldValue
      const original = (function find(list: any[], id: number): any {
        for (const item of list) {
          if (item.id === id) return item
          if (item.children) { const f = find(item.children, id); if (f) return f }
        }
        return null
      })(tableData.value, row.id)
      if (original) original[field] = oldValue
      ElMessage.error('更新失败')
    }
  }

  /**
   * 菜单表单数据类型
   */
  interface MenuFormData {
    name: string
    path: string
    component?: string
    icon?: string
    roles?: string[]
    sort?: number
    [key: string]: any
  }

  /**
   * 提交表单数据
   * @param formData 表单数据
   */
  const handleSubmit = async (formData: any): Promise<void> => {
    try {
      // 转换前端表单格式为后端格式
      const params: any = {
        id: formData.id || 0,
        parentId: formData.parentId || 0,  // null/undefined转为0（顶级）
        type: formData.menuType === 'button' ? 3 : (formData.menuType === 'directory' ? 1 : 2),
        title: formData.menuType === 'button' ? formData.authName : formData.name,
        name: formData.menuType === 'button' ? formData.authLabel : formData.label,
        path: formData.path || '',
        component: formData.component || '',
        icon: formData.icon || '',
        resource: formData.resource || '',  // 关联数据表名
        hidden: formData.isHide ? 1 : 0,
        hideTab: formData.isHideTab ? 1 : 0,
        keepAlive: formData.keepAlive ? 1 : 0,
        redirect: formData.redirect || '',
        frameSrc: formData.link || '',
        perms: formData.perms || '',
        isFrame: formData.isIframe ? 1 : 0,
        affix: formData.fixedTab ? 1 : 0,
        showBadge: formData.showBadge ? 1 : 0,
        badgeText: formData.showTextBadge || '',
        activePath: formData.activePath || '',
        isFullPage: formData.isFullPage ? 1 : 0,
        sort: formData.sort || formData.authSort || 1,
        status: formData.isEnable ? 1 : 0,
        remark: formData.remark || ''
      }

      await fetchSaveMenu(params)
      const action = params.id ? '编辑' : '添加'
      ElMessage.success(`${action}成功，请刷新页面查看左侧菜单更新`)
      dialogVisible.value = false
      await getMenuList()  // 刷新当前页面菜单列表
    } catch (error) {
      console.error('保存菜单失败:', error)
    }
  }

  /**
   * 删除菜单
   * @param row 菜单行数据
   */
  const handleDeleteMenu = async (row: any): Promise<void> => {
    try {
      await ElMessageBox.confirm('确定要删除该菜单吗？删除后无法恢复', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      await fetchDeleteMenu(row.id)
      ElMessage.success('删除成功')
      getMenuList()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除菜单失败:', error)
      }
    }
  }

  /**
   * 删除权限按钮
   */
  const handleDeleteAuth = async (row: any): Promise<void> => {
    try {
      await ElMessageBox.confirm('确定要删除该权限吗？删除后无法恢复', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await fetchDeleteMenu(row.id)
      ElMessage.success('删除成功')
      getMenuList()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error('删除失败')
      }
    }
  }

  /**
   * 切换展开/收起所有菜单
   */
  const toggleExpand = (): void => {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      if (tableRef.value?.elTableRef && filteredTableData.value) {
        const processRows = (rows: AppRouteRecord[]) => {
          rows.forEach((row) => {
            if (row.children?.length) {
              tableRef.value.elTableRef.toggleRowExpansion(row, isExpanded.value)
              processRows(row.children)
            }
          })
        }
        processRows(filteredTableData.value)
      }
    })
  }
</script>
