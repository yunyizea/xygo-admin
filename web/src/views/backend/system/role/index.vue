<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 角色管理页面 -->
<template>
  <div class="art-full-height">
    <RoleSearch
      v-show="showSearchBar"
      v-model="searchForm"
      @search="handleSearch"
      @reset="resetSearchParams"
    ></RoleSearch>

    <ElCard
      class="art-table-card"
      shadow="never"
      :style="{ 'margin-top': showSearchBar ? '12px' : '0' }"
    >
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showDialog('add')" v-ripple>新增角色</ElButton>
            <ElButton @click="toggleExpand" v-ripple type="primary">
              {{ isExpanded ? '收起' : '展开' }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        ref="tableRef"
        rowKey="id"
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        :stripe="false"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="true"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>
    </ElCard>

    <!-- 角色编辑弹窗 -->
    <RoleEditDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :role-data="currentRoleData"
      @success="refreshData"
    />

    <!-- 菜单权限弹窗 -->
    <RolePermissionDialog
      v-model="permissionDialog"
      :role-data="currentRoleData"
      @success="refreshData"
    />

    <!-- 数据权限弹窗 -->
    <RoleDataScopeDialog
      v-model="dataScopeDialog"
      :role-data="currentRoleData"
      @success="refreshData"
    />

    <!-- 字段权限弹窗 -->
    <RoleFieldPermDialog
      v-model="fieldPermDialog"
      :role-data="currentRoleData"
      @success="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetRoleList } from '@/api/backend/system'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import RoleSearch from './modules/role-search.vue'
  import RoleEditDialog from './modules/role-edit-dialog.vue'
  import RolePermissionDialog from './modules/role-permission-dialog.vue'
  import RoleDataScopeDialog from './modules/role-data-scope-dialog.vue'
  import RoleFieldPermDialog from './modules/role-field-perm-dialog.vue'
  import { ElTag, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'Role' })

  type RoleListItem = Api.SystemManage.RoleListItem

  // 搜索表单
  const searchForm = ref({
    roleName: undefined,
    roleCode: undefined,
    description: undefined,
    enabled: undefined,
    daterange: undefined
  })

  const showSearchBar = ref(false)
  const isExpanded = ref(true)
  const tableRef = ref()

  const dialogVisible = ref(false)
  const permissionDialog = ref(false)
  const dataScopeDialog = ref(false)
  const fieldPermDialog = ref(false)
  const currentRoleData = ref<RoleListItem | undefined>(undefined)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    // 核心配置
    core: {
      apiFn: fetchGetRoleList,
      apiParams: {
        current: 1,
        size: 20
      },
      excludeParams: ['daterange'],
      columnsFactory: () => [
        {
          prop: 'id',
          label: '角色ID',
          width: 100
        },
        {
          prop: 'name',
          label: '角色名称',
          minWidth: 120
        },
        {
          prop: 'key',
          label: '角色标识',
          minWidth: 120
        },
        {
          prop: 'remark',
          label: '角色描述',
          minWidth: 150,
          showOverflowTooltip: true,
          formatter: (row: RoleListItem) => row.remark || '-'
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          align: 'center',
          formatter: (row: RoleListItem) => {
            return h(ElTag, {
              type: row.status === 1 ? 'success' : 'danger'
            }, () => row.status === 1 ? '启用' : '禁用')
          }
        },
        {
          prop: 'createdAt',
          label: '创建时间',
          width: 180,
          sortable: true,
          formatter: (row: RoleListItem) => row.createdAt || '-'
        },
        {
          prop: 'operation',
          label: '操作',
          width: 130,
          fixed: 'right',
          align: 'right',
          formatter: (row: RoleListItem) => {
            const isSuperAdmin = row.key === 'super_admin'
            const menuList: any[] = []
            
            if (!isSuperAdmin) {
              menuList.push(
                {
                  key: 'permission',
                  label: '菜单权限',
                  icon: 'ri:menu-line'
                },
                {
                  key: 'dataScope',
                  label: '数据权限',
                  icon: 'ri:shield-check-line'
                },
                {
                  key: 'fieldPerm',
                  label: '字段权限',
                  icon: 'ri:table-line'
                }
              )
            }
            
            menuList.push({
              key: 'edit',
              label: '编辑角色',
              icon: 'ri:edit-2-line'
            })
            
            if (!isSuperAdmin) {
              menuList.push({
                key: 'delete',
                label: '删除角色',
                icon: 'ri:delete-bin-4-line',
                color: '#f56c6c'
              })
            }
            
            return h('div', { class: 'flex items-center justify-end' }, [
              h(ArtButtonTable, {
                type: 'add',
                title: '添加子角色',
                onClick: () => showAddChildRole(row)
              }),
              h(ArtButtonMore, {
                list: menuList,
                onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
              })
            ])
          }
        }
      ]
    }
  })

  const dialogType = ref<'add' | 'edit'>('add')

  const showDialog = (type: 'add' | 'edit', row?: RoleListItem) => {
    dialogType.value = type
    currentRoleData.value = row
    dialogVisible.value = true
  }

  const showAddChildRole = (parent: RoleListItem) => {
    showDialog('add', parent)
  }

  const handleSearch = (params: Record<string, any>) => {
    const { daterange, ...filtersParams } = params
    const [startTime, endTime] = Array.isArray(daterange) ? daterange : [null, null]

    Object.assign(searchParams, { ...filtersParams, startTime, endTime })
    getData()
  }

  const buttonMoreClick = (item: ButtonMoreItem, row: RoleListItem) => {
    switch (item.key) {
      case 'permission':
        showPermissionDialog(row)
        break
      case 'dataScope':
        showDataScopeDialog(row)
        break
      case 'fieldPerm':
        showFieldPermDialog(row)
        break
      case 'edit':
        showDialog('edit', row)
        break
      case 'delete':
        deleteRole(row)
        break
    }
  }

  const showPermissionDialog = (row?: RoleListItem) => {
    permissionDialog.value = true
    currentRoleData.value = row
  }

  const showDataScopeDialog = (row?: RoleListItem) => {
    dataScopeDialog.value = true
    currentRoleData.value = row
  }

  const showFieldPermDialog = (row?: RoleListItem) => {
    fieldPermDialog.value = true
    currentRoleData.value = row
  }

  const deleteRole = (row: RoleListItem) => {
    ElMessageBox.confirm(`确定删除角色"${row.name}"吗？此操作不可恢复！`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => {
        // TODO: 调用删除接口
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch(() => {
        ElMessage.info('已取消删除')
      })
  }

  const toggleExpand = (): void => {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      if (tableRef.value?.elTableRef && data.value) {
        const processRows = (rows: RoleListItem[]) => {
          rows.forEach((row) => {
            if (row.children?.length) {
              tableRef.value.elTableRef.toggleRowExpansion(row, isExpanded.value)
              processRows(row.children)
            }
          })
        }
        processRows(data.value)
      }
    })
  }
</script>
