<!-- 会员菜单管理页面（对齐 BuildAdmin user/rule） -->
<template>
  <div class="member-menu-page art-full-height">
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
          <ElButton v-auth="'add'" @click="handleAdd" v-ripple>添加菜单</ElButton>
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

      <!-- 菜单弹窗（新版：对齐 BuildAdmin） -->
      <MenuDialog
        v-model:visible="dialogVisible"
        :editData="editData"
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
  import { useAuth } from '@/hooks/core/useAuth'
  import MenuDialog from './modules/menu-dialog.vue'
  import {
    getMemberMenuTree,
    saveMemberMenu,
    deleteMemberMenu,
    type MemberMenuItem,
    type RuleType
  } from '@/api/backend/member/menu'
  import { ElTag, ElMessageBox, ElSwitch } from 'element-plus'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'

  defineOptions({ name: 'MemberMenu' })
  const { hasAuth } = useAuth()

  // 类型标签配色
  const typeTagMap: Record<string, { text: string; type: 'primary' | 'success' | 'warning' | 'danger' | 'info' }> = {
    route:         { text: '普通路由',     type: 'info' },
    menu_dir:      { text: '菜单目录',     type: 'success' },
    menu:          { text: '菜单项',       type: 'danger' },
    nav:           { text: '顶栏菜单',     type: 'warning' },
    nav_user_menu: { text: '顶栏下拉',     type: 'primary' },
    button:        { text: '按钮',         type: 'info' },
  }

  // 菜单类型标签
  const menuTypeMap: Record<string, string> = {
    tab: '标签卡',
    link: '链接(站外)',
    iframe: 'Iframe',
  }

  // 状态管理
  const loading = ref(false)
  const isExpanded = ref(false)
  const tableRef = ref()

  // 弹窗相关
  const dialogVisible = ref(false)
  const editData = ref<MemberMenuItem | null>(null)
  const parentMenu = ref<MemberMenuItem | null>(null)

  // 搜索
  const initialSearchState = { name: '', route: '' }
  const formFilters = reactive({ ...initialSearchState })
  const appliedFilters = reactive({ ...initialSearchState })

  const formItems = computed(() => [
    { label: '菜单名称', key: 'name', type: 'input', props: { clearable: true } },
    { label: '路由地址', key: 'route', type: 'input', props: { clearable: true } },
  ])

  onMounted(() => { getMenuList() })

  const getMenuList = async () => {
    loading.value = true
    try {
      const res = await getMemberMenuTree()
      tableData.value = res.list || []
    } catch (error) {
      console.error('获取菜单失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 表格列配置
  const { columnChecks, columns } = useTableColumns(() => [
    {
      prop: 'title',
      label: '菜单名称',
      minWidth: 200,
      formatter: (row: any) => formatMenuTitle(row.title)
    },
    {
      prop: 'icon',
      label: '图标',
      width: 60,
      align: 'center',
      formatter: (row: any) => {
        if (!row.icon) return ''
        return h(ArtSvgIcon, { icon: row.icon, style: 'font-size: 20px;' })
      }
    },
    {
      prop: 'type',
      label: '菜单类型',
      width: 120,
      formatter: (row: any) => {
        const cfg = typeTagMap[row.type] || { text: row.type, type: 'info' }
        return h(ElTag, { type: cfg.type, size: 'small' }, () => cfg.text)
      }
    },
    {
      prop: 'name',
      label: '名称',
      minWidth: 150,
      formatter: (row: any) => row.name || '-'
    },
    {
      prop: 'path',
      label: '路由',
      minWidth: 150,
      formatter: (row: any) => {
        if (row.type === 'button') return ''
        return row.path || ''
      }
    },
    {
      prop: 'menuType',
      label: '打开方式',
      width: 100,
      formatter: (row: any) => {
        if (row.type === 'button' || row.type === 'menu_dir') return ''
        return menuTypeMap[row.menuType] || row.menuType || ''
      }
    },
    {
      prop: 'sort',
      label: '排序',
      width: 70,
      align: 'center',
    },
    {
      prop: 'status',
      label: '是否启用',
      width: 90,
      align: 'center',
      formatter: (row: any) =>
        h(ElSwitch, {
          modelValue: row.status === 1,
          activeColor: '#13ce66',
          inactiveColor: '#ff4949',
          onChange: (val: any) => handleStatusChange(row, val)
        })
    },
    {
      prop: 'operation',
      label: '操作',
      width: 160,
      align: 'right',
      formatter: (row: any) => {
        const btns: any[] = []
        if (row.type !== 'button' && hasAuth('add')) {
          btns.push(h(ArtButtonTable, {
            type: 'add',
            onClick: () => handleAddChild(row),
            title: '添加子级'
          }))
        }
        if (hasAuth('edit')) {
          btns.push(h(ArtButtonTable, {
            type: 'edit',
            onClick: () => handleEdit(row)
          }))
        }
        if (hasAuth('delete')) {
          btns.push(h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDelete(row)
          }))
        }
        return h('div', { style: 'text-align: right' }, btns)
      }
    }
  ])

  const tableData = ref<MemberMenuItem[]>([])

  // 搜索过滤
  const deepClone = <T,>(obj: T): T => {
    if (obj === null || typeof obj !== 'object') return obj
    if (obj instanceof Date) return new Date(obj) as T
    if (Array.isArray(obj)) return obj.map(item => deepClone(item)) as T
    const cloned = {} as T
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        cloned[key] = deepClone(obj[key])
      }
    }
    return cloned
  }

  const searchMenu = (items: MemberMenuItem[]): MemberMenuItem[] => {
    const results: MemberMenuItem[] = []
    for (const item of items) {
      const sName = appliedFilters.name?.toLowerCase().trim() || ''
      const sRoute = appliedFilters.route?.toLowerCase().trim() || ''
      const titleMatch = !sName || (item.title || '').toLowerCase().includes(sName)
      const pathMatch = !sRoute || (item.path || '').toLowerCase().includes(sRoute)

      if (item.children?.length) {
        const matched = searchMenu(item.children)
        if (matched.length > 0) {
          const c = deepClone(item)
          c.children = matched
          results.push(c)
          continue
        }
      }
      if (titleMatch && pathMatch) {
        results.push(deepClone(item))
      }
    }
    return results
  }

  const filteredTableData = computed(() => searchMenu(tableData.value))

  const handleReset = () => {
    Object.assign(formFilters, { ...initialSearchState })
    Object.assign(appliedFilters, { ...initialSearchState })
    getMenuList()
  }
  const handleSearch = () => { Object.assign(appliedFilters, { ...formFilters }) }
  const handleRefresh = () => { getMenuList() }

  // 添加（顶级）
  const handleAdd = () => {
    editData.value = null
    parentMenu.value = null
    dialogVisible.value = true
  }

  // 添加子级
  const handleAddChild = (row: MemberMenuItem) => {
    editData.value = null
    parentMenu.value = row
    dialogVisible.value = true
  }

  // 编辑
  const handleEdit = (row: MemberMenuItem) => {
    editData.value = row
    parentMenu.value = null
    dialogVisible.value = true
  }

  // 删除
  const handleDelete = (row: MemberMenuItem) => {
    ElMessageBox.confirm(`确定删除"${row.title}"吗？此操作不可恢复！`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
      .then(async () => {
        await deleteMemberMenu(row.id)
        ElMessage.success('删除成功')
        getMenuList()
      })
      .catch(() => {})
  }

  // 状态切换
  const handleStatusChange = async (row: MemberMenuItem, value: any) => {
    const newStatus = value ? 1 : 0
    try {
      await saveMemberMenu({ ...row, status: newStatus })
      row.status = newStatus
      ElMessage.success('更新成功')
    } catch {
      ElMessage.error('更新失败')
    }
  }

  // 弹窗提交后刷新
  const handleSubmit = () => { getMenuList() }

  // 展开/收起
  const toggleExpand = () => {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      if (tableRef.value?.elTableRef && filteredTableData.value) {
        const processRows = (rows: MemberMenuItem[]) => {
          rows.forEach(row => {
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
