<!-- 登录日志管理页面 -->
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
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    getLoginLogList,
    deleteLoginLog,
    clearLoginLog,
    type LoginLogItem
  } from '@/api/backend/monitor/loginLog'
  import LogSearch from './modules/log-search.vue'
  import { ElTag, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'LoginLog' })
  const { hasAuth } = useAuth()

  const searchForm = ref({
    username: undefined,
    ip: undefined,
    status: undefined,
    dateRange: undefined
  })

  const showSearchBar = ref(false)
  const tableRef = ref()
  const selectedIds = ref<number[]>([])

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
      apiFn: getLoginLogList,
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
          label: '登录账号',
          minWidth: 120
        },
        {
          prop: 'ip',
          label: '登录IP',
          minWidth: 130
        },
        {
          prop: 'location',
          label: '登录地点',
          minWidth: 130,
          formatter: (row: LoginLogItem) => row.location || '-'
        },
        {
          prop: 'browser',
          label: '浏览器',
          width: 100
        },
        {
          prop: 'os',
          label: '操作系统',
          width: 100
        },
        {
          prop: 'status',
          label: '状态',
          width: 80,
          align: 'center',
          formatter: (row: LoginLogItem) => {
            return h(ElTag, {
              type: row.status === 1 ? 'success' : 'danger',
              size: 'small'
            }, () => row.status === 1 ? '成功' : '失败')
          }
        },
        {
          prop: 'message',
          label: '提示消息',
          minWidth: 150,
          showOverflowTooltip: true
        },
        {
          prop: 'createdAt',
          label: '登录时间',
          width: 180,
          sortable: true
        }
      ]
    }
  })

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  const handleSelectionChange = (rows: LoginLogItem[]) => {
    selectedIds.value = rows.map(r => r.id)
  }

  const handleBatchDelete = () => {
    if (selectedIds.value.length === 0) return
    ElMessageBox.confirm(`确定删除选中的 ${selectedIds.value.length} 条日志？`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        await deleteLoginLog(selectedIds.value)
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch(() => {})
  }

  const handleClear = () => {
    ElMessageBox.confirm('确定清空所有登录日志？此操作不可恢复！', '清空确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(async () => {
        await clearLoginLog()
        ElMessage.success('清空成功')
        refreshData()
      })
      .catch(() => {})
  }
</script>
