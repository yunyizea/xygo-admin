<!-- 操作日志管理页面 -->
<template>
  <div class="art-full-height">
    <LogSearch
      v-show="showSearchBar"
      v-model="searchForm"
      @search="handleSearch"
      @reset="resetSearchParams"
    ></LogSearch>

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
            <ElButton
              v-auth="'batchDel'"
              type="danger"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
              v-ripple
            >
              批量删除
            </ElButton>
            <ElButton v-auth="'clear'" type="warning" @click="handleClear" v-ripple>清空日志</ElButton>
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
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
        @selection-change="handleSelectionChange"
      >
      </ArtTable>
    </ElCard>

    <!-- 详情抽屉 -->
    <LogDetailDrawer
      v-model="detailVisible"
      :log-id="currentLogId"
    />
  </div>
</template>

<script setup lang="ts">
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    getOperationLogList,
    deleteOperationLog,
    clearOperationLog,
    type OperationLogItem
  } from '@/api/backend/monitor/operationLog'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import LogSearch from './modules/log-search.vue'
  import LogDetailDrawer from './modules/log-detail-drawer.vue'
  import { ElTag, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'OperationLog' })
  const { hasAuth } = useAuth()

  const searchForm = ref({
    username: undefined,
    module: undefined,
    status: undefined,
    dateRange: undefined
  })

  const showSearchBar = ref(false)
  const tableRef = ref()
  const selectedIds = ref<number[]>([])
  const detailVisible = ref(false)
  const currentLogId = ref<number>()

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
    core: {
      apiFn: getOperationLogList,
      apiParams: {
        page: 1,
        pageSize: 20
      },
      paginationKey: {
        current: 'page',
        size: 'pageSize'
      },
      columnsFactory: () => [
        {
          type: 'selection',
          width: 50,
          align: 'center'
        },
        {
          prop: 'id',
          label: 'ID',
          width: 80
        },
        {
          prop: 'username',
          label: '操作人',
          minWidth: 100
        },
        {
          prop: 'module',
          label: '模块',
          minWidth: 120,
          showOverflowTooltip: true
        },
        {
          prop: 'title',
          label: '操作',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'method',
          label: '请求方式',
          width: 90,
          align: 'center',
          formatter: (row: OperationLogItem) => {
            const typeMap: Record<string, string> = {
              GET: 'info',
              POST: 'success',
              PUT: 'warning',
              DELETE: 'danger'
            }
            return h(ElTag, {
              type: (typeMap[row.method] || 'info') as any,
              size: 'small'
            }, () => row.method)
          }
        },
        {
          prop: 'url',
          label: '请求URL',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'ip',
          label: 'IP',
          width: 130
        },
        {
          prop: 'status',
          label: '状态',
          width: 80,
          align: 'center',
          formatter: (row: OperationLogItem) => {
            return h(ElTag, {
              type: row.status === 1 ? 'success' : 'danger',
              size: 'small'
            }, () => row.status === 1 ? '成功' : '失败')
          }
        },
        {
          prop: 'elapsed',
          label: '耗时',
          width: 90,
          align: 'center',
          formatter: (row: OperationLogItem) => {
            const color = row.elapsed > 1000 ? '#f56c6c' : row.elapsed > 500 ? '#e6a23c' : '#67c23a'
            return h('span', { style: { color } }, `${row.elapsed}ms`)
          }
        },
        {
          prop: 'createdAt',
          label: '操作时间',
          width: 180,
          sortable: true
        },
        {
          prop: 'operation',
          label: '操作',
          width: 80,
          fixed: 'right',
          formatter: (row: OperationLogItem) => {
            const menuList: any[] = [
              {
                key: 'detail',
                label: '查看详情',
                icon: 'ri:eye-line'
              },
              {
                key: 'delete',
                label: '删除',
                icon: 'ri:delete-bin-4-line',
                color: '#f56c6c'
              }
            ]
            const filteredMenuList = menuList.filter((m: any) => hasAuth(m.key))

            return filteredMenuList.length ? h('div', [
              h(ArtButtonMore, {
                list: filteredMenuList,
                onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
              })
            ]) : null
          }
        }
      ]
    }
  })

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  const handleSelectionChange = (rows: OperationLogItem[]) => {
    selectedIds.value = rows.map(r => r.id)
  }

  const buttonMoreClick = (item: ButtonMoreItem, row: OperationLogItem) => {
    switch (item.key) {
      case 'detail':
        showDetail(row)
        break
      case 'delete':
        handleDelete(row)
        break
    }
  }

  const showDetail = (row: OperationLogItem) => {
    currentLogId.value = row.id
    detailVisible.value = true
  }

  const handleDelete = (row: OperationLogItem) => {
    ElMessageBox.confirm('确定删除该条操作日志？', '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        await deleteOperationLog([row.id])
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch(() => {})
  }

  const handleBatchDelete = () => {
    if (selectedIds.value.length === 0) return
    ElMessageBox.confirm(`确定删除选中的 ${selectedIds.value.length} 条日志？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        await deleteOperationLog(selectedIds.value)
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch(() => {})
  }

  const handleClear = () => {
    ElMessageBox.confirm('确定清空所有操作日志？此操作不可恢复！', '清空确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        await clearOperationLog()
        ElMessage.success('清空成功')
        refreshData()
      })
      .catch(() => {})
  }
</script>
