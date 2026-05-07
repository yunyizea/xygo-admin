<!-- 会员管理页面 -->
<template>
  <div class="member-page art-full-height">
    <!-- 搜索栏 -->
    <MemberSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></MemberSearch>

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'add'" @click="showDialog('add')" v-ripple>新增会员</ElButton>
            <ElButton v-auth="'batchDel'" type="danger" :disabled="selectedRows.length === 0" @click="handleBatchDelete">批量删除</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 会员弹窗 -->
      <MemberDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :member-data="currentMemberData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    getMemberList,
    deleteMember,
    updateMemberStatus,
    type MemberItem
  } from '@/api/backend/member'
  import MemberSearch from './modules/member-search.vue'
  import MemberDialog from './modules/member-dialog.vue'
  import { ElTag, ElMessageBox, ElAvatar, ElSwitch } from 'element-plus'
  import { DialogType } from '@/types'

  defineOptions({ name: 'MemberManage' })
  const { hasAuth } = useAuth()

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentMemberData = ref<Partial<MemberItem>>({})

  // 选中行
  const selectedRows = ref<MemberItem[]>([])

  // 搜索表单
  const searchForm = ref({
    username: undefined,
    mobile: undefined,
    email: undefined,
    status: undefined,
    groupId: undefined
  })


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
      apiFn: getMemberList,
      apiParams: {
        page: 1,
        pageSize: 20,
        ...searchForm.value
      },
      paginationKey: {
        current: 'page',
        size: 'pageSize'
      },
      columnsFactory: () => [
        { type: 'selection', width: 50 },
        { prop: 'id', label: 'ID', width: 80 },
        {
          prop: 'avatar',
          label: '头像',
          width: 80,
          formatter: (row: MemberItem) => h(ElAvatar, { size: 40, src: row.avatar }, () => row.nickname?.charAt(0) || 'U')
        },
        { prop: 'username', label: '用户名', minWidth: 120 },
        { prop: 'nickname', label: '昵称', minWidth: 120 },
        { prop: 'mobile', label: '手机号', minWidth: 130 },
        { prop: 'email', label: '邮箱', minWidth: 180 },
        {
          prop: 'gender',
          label: '性别',
          width: 80,
          formatter: (row: MemberItem) => {
            const genderMap: Record<number, string> = { 0: '未知', 1: '男', 2: '女' }
            return genderMap[row.gender] || '未知'
          }
        },
        { prop: 'groupName', label: '会员分组', width: 100 },
        { prop: 'score', label: '积分', width: 80 },
        { prop: 'money', label: '余额', width: 100 },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: MemberItem) =>
            h(ElSwitch, {
              modelValue: row.status === 1,
              activeText: '正常',
              inactiveText: '禁用',
              onChange: () => handleStatusChange(row)
            })
        },
        { prop: 'loginCount', label: '登录次数', width: 100 },
        { prop: 'lastLoginAt', label: '最后登录', width: 160 },
        { prop: 'createdAt', label: '注册时间', width: 160 },
        {
          prop: 'action',
          label: '操作',
          width: 200,
          fixed: 'right',
          formatter: (row: MemberItem) =>
            h('div', { class: 'table-actions' }, [
              hasAuth('edit') ? h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }) : null,
              hasAuth('delete') ? h(ArtButtonTable, {
                type: 'delete',
                onClick: () => handleDelete(row)
              }) : null,
            ].filter(Boolean))
        }
      ]
    }
  })

  // 搜索
  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 选择变化
  const handleSelectionChange = (rows: MemberItem[]) => {
    selectedRows.value = rows
  }

  // 显示弹窗
  const showDialog = (type: DialogType, row?: MemberItem) => {
    dialogType.value = type
    currentMemberData.value = row ? { ...row } : {}
    dialogVisible.value = true
  }

  // 弹窗提交
  const handleDialogSubmit = () => {
    refreshData()
  }

  // 状态变更
  const handleStatusChange = async (row: MemberItem) => {
    const newStatus = row.status === 1 ? 0 : 1
    const statusText = newStatus === 1 ? '启用' : '禁用'

    try {
      await ElMessageBox.confirm(`确定要${statusText}该会员吗？`, '提示', { type: 'warning' })
      await updateMemberStatus(row.id, newStatus)
      ElMessage.success(`${statusText}成功`)
      refreshData()
    } catch {
      // 取消操作
    }
  }

  // 删除
  const handleDelete = async (row: MemberItem) => {
    try {
      await ElMessageBox.confirm('确定要删除该会员吗？删除后不可恢复', '警告', { type: 'warning' })
      await deleteMember([row.id])
      ElMessage.success('删除成功')
      refreshData()
    } catch {
      // 取消操作
    }
  }

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRows.value.length === 0) return

    try {
      await ElMessageBox.confirm(`确定要删除选中的 ${selectedRows.value.length} 个会员吗？`, '警告', { type: 'warning' })
      const ids = selectedRows.value.map(row => row.id)
      await deleteMember(ids)
      ElMessage.success('删除成功')
      refreshData()
    } catch {
      // 取消操作
    }
  }
</script>
