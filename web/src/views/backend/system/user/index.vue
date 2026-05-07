<!-- 用户管理页面 -->
<!-- art-full-height 自动计算出页面剩余高度 -->
<!-- art-table-card 一个符合系统样式的 class，同时自动撑满剩余高度 -->
<!-- 更多 useTable 使用示例请移步至 功能示例 下面的高级表格示例或者查看官方文档 -->
<!-- useTable 文档：https://www.artd.pro/docs/zh/guide/hooks/use-table.html -->
<template>
  <div class="user-page art-full-height">
    <!-- 搜索栏 -->
    <UserSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams"></UserSearch>

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'add'" @click="showDialog('add')" v-ripple>新增用户</ElButton>
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

      <!-- 用户弹窗 -->
      <UserDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :user-data="currentUserData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { ACCOUNT_TABLE_DATA } from '@/mock/temp/formData'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetUserList, fetchSaveUser, fetchDeleteUser, fetchKickUser } from '@/api/backend/system'
  import UserSearch from './modules/user-search.vue'
  import UserDialog from './modules/user-dialog.vue'
  import { ElTag, ElMessageBox, ElImage, ElButton } from 'element-plus'
  import { DialogType } from '@/types'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'User' })
  const { hasAuth } = useAuth()

  type UserListItem = Api.SystemManage.UserListItem

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentUserData = ref<Partial<UserListItem>>({})

  // 选中行
  const selectedRows = ref<UserListItem[]>([])

  // 搜索表单
  const searchForm = ref({
    username: undefined,
    gender: undefined,
    mobile: undefined,
    email: undefined,
    status: 1
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
      apiFn: fetchGetUserList,
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
        { type: 'selection' }, // 勾选列
        { type: 'index', width: 60, label: '序号' }, // 序号
        {
          prop: 'userInfo',
          label: '用户信息',
          minWidth: 220,
          formatter: (row: UserListItem) => {
            return h('div', { class: 'user flex-c' }, [
              h(ElImage, {
                class: 'size-9.5 rounded-md',
                src: row.avatar,
                previewSrcList: [row.avatar],
                previewTeleported: true
              }),
              h('div', { class: 'ml-2' }, [
                h('p', { class: 'user-name' }, row.username),
                h('p', { class: 'email' }, row.email || '-')
              ])
            ])
          }
        },
        {
          prop: 'nickname',
          label: '昵称',
          formatter: (row: UserListItem) => row.nickname || '-'
        },
        {
          prop: 'gender',
          label: '性别',
          width: 80,
          formatter: (row: UserListItem) => {
            const genderMap: Record<string, string> = {
              '1': '男',
              '2': '女',
              '0': '未知'
            }
            return genderMap[row.gender] || '-'
          }
        },
        { 
          prop: 'mobile', 
          label: '手机号',
          formatter: (row: UserListItem) => row.mobile || '-'
        },
        {
          prop: 'roles',
          label: '角色',
          formatter: (row: UserListItem) => {
            if (!row.roleNames || row.roleNames.length === 0) return '-'
            return h('div', { class: 'flex flex-wrap gap-1' }, 
              row.roleNames.map((name: string) => 
                h(ElTag, { size: 'small', type: 'primary' }, () => name)
              )
            )
          }
        },
        {
          prop: 'status',
          label: '状态',
          width: 80,
          align: 'center',
          formatter: (row: UserListItem) => {
            return h(ElTag, { 
              type: row.status === 1 ? 'success' : 'danger' 
            }, () => row.status === 1 ? '启用' : '禁用')
          }
        },
        {
          prop: 'create_time',
          label: '创建时间',
          width: 180,
          sortable: true,
          formatter: (row: UserListItem) => formatTimestamp(row.create_time)
        },
        {
          prop: 'operation',
          label: '操作',
          width: 180,
          fixed: 'right', // 固定列
          formatter: (row: UserListItem) =>
            h('div', { class: 'flex items-center gap-1' }, [
              hasAuth('edit') ? h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }) : null,
              hasAuth('delete') ? h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deleteUser(row)
              }) : null,
              hasAuth('kick') ? h(ElButton, {
                type: 'warning',
                size: 'small',
                link: true,
                onClick: () => kickUser(row)
              }, () => '下线') : null,
            ].filter(Boolean))
        }
      ]
    }
  })

  /**
   * 搜索处理
   * @param params 参数
   */
  const handleSearch = (params: Record<string, any>) => {
    console.log(params)
    // 搜索参数赋值
    Object.assign(searchParams, params)
    getData()
  }

  /**
   * 显示用户弹窗
   */
  const showDialog = (type: DialogType, row?: UserListItem): void => {
    console.log('打开弹窗:', { type, row })
    dialogType.value = type
    currentUserData.value = row || {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /**
   * 强制用户下线
   */
  const kickUser = async (row: UserListItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要将用户"${row.username}"强制下线吗？`,
        '强制下线',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
      await fetchKickUser(row.id)
      ElMessage.success('已将该用户强制下线')
    } catch (error) {
      if (error !== 'cancel') {
        console.error('强制下线失败:', error)
      }
    }
  }

  /**
   * 删除用户
   */
  const deleteUser = async (row: UserListItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除用户"${row.username}"吗？删除后无法恢复`, '删除用户', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      
      await fetchDeleteUser(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除用户失败:', error)
      }
    }
  }

  /**
   * 处理弹窗提交事件
   */
  const handleDialogSubmit = async (formData: any) => {
    try {
      await fetchSaveUser(formData)
      ElMessage.success(formData.id ? '编辑成功' : '添加成功')
      dialogVisible.value = false
      currentUserData.value = {}
      refreshData()
    } catch (error) {
      console.error('保存用户失败:', error)
    }
  }

  /**
   * 处理表格行选择变化
   */
  const handleSelectionChange = (selection: UserListItem[]): void => {
    selectedRows.value = selection
    console.log('选中行数据:', selectedRows.value)
  }
</script>
