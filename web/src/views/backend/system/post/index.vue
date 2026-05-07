<!-- 岗位管理页面 -->
<template>
  <div class="post-page art-full-height">
    <!-- 搜索栏 -->
    <PostSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton v-auth="'add'" @click="showDialog('add')" v-ripple>新增岗位</ElButton>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <!-- 岗位弹窗 -->
      <PostDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :post-data="currentPostData"
        @submit="handleDialogSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetPostList, fetchSavePost, fetchDeletePost } from '@/api/backend/system'
  import PostSearch from './modules/post-search.vue'
  import PostDialog from './modules/post-dialog.vue'
  import { ElTag, ElMessageBox, ElSwitch } from 'element-plus'
  import { DialogType } from '@/types'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'Post' })
  const { hasAuth } = useAuth()

  interface PostListItem {
    id: number
    code: string
    name: string
    sort: number
    status: number
    remark: string
    create_time: number
    update_time: number
  }

  // 弹窗相关
  const dialogType = ref<DialogType>('add')
  const dialogVisible = ref(false)
  const currentPostData = ref<Partial<PostListItem>>({})

  // 搜索表单
  const searchForm = ref({
    name: undefined,
    code: undefined,
    status: undefined
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
    core: {
      apiFn: fetchGetPostList,
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
        { type: 'index', width: 60, label: '序号' },
        {
          prop: 'code',
          label: '岗位编码',
          minWidth: 120
        },
        {
          prop: 'name',
          label: '岗位名称',
          minWidth: 150
        },
        {
          prop: 'sort',
          label: '排序',
          width: 80,
          align: 'center'
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          align: 'center',
          formatter: (row: PostListItem) =>
            h(ElSwitch, {
              modelValue: row.status === 1,
              activeColor: '#13ce66',
              inactiveColor: '#ff4949',
              onChange: (val) => handleStatusChange(row, val)
            })
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 150,
          showOverflowTooltip: true,
          formatter: (row: PostListItem) => row.remark || '-'
        },
        {
          prop: 'create_time',
          label: '创建时间',
          width: 180,
          sortable: true,
          formatter: (row: PostListItem) => formatTimestamp(row.create_time)
        },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row: PostListItem) =>
            h('div', { class: 'flex items-center gap-1' }, [
              hasAuth('edit') ? h(ArtButtonTable, {
                type: 'edit',
                onClick: () => showDialog('edit', row)
              }) : null,
              hasAuth('delete') ? h(ArtButtonTable, {
                type: 'delete',
                onClick: () => deletePost(row)
              }) : null,
            ].filter(Boolean))
        }
      ]
    }
  })

  /**
   * 搜索处理
   */
  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  /**
   * 显示弹窗
   */
  const showDialog = (type: DialogType, row?: PostListItem): void => {
    dialogType.value = type
    currentPostData.value = row || {}
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /**
   * 处理状态切换
   */
  const handleStatusChange = async (row: PostListItem, value: string | number | boolean): Promise<void> => {
    try {
      const boolValue = !!value
      await fetchSavePost({
        id: row.id,
        code: row.code,
        name: row.name,
        sort: row.sort,
        status: boolValue ? 1 : 0,
        remark: row.remark
      })
      row.status = boolValue ? 1 : 0
      ElMessage.success('更新成功')
    } catch (error) {
      row.status = value ? 0 : 1
      ElMessage.error('更新失败')
    }
  }

  /**
   * 删除岗位
   */
  const deletePost = async (row: PostListItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除岗位"${row.name}"吗？删除后无法恢复`, '删除岗位', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })

      await fetchDeletePost(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除岗位失败:', error)
      }
    }
  }

  /**
   * 处理弹窗提交
   */
  const handleDialogSubmit = async (formData: any) => {
    try {
      await fetchSavePost(formData)
      ElMessage.success(formData.id ? '编辑成功' : '添加成功')
      dialogVisible.value = false
      currentPostData.value = {}
      refreshData()
    } catch (error) {
      console.error('保存岗位失败:', error)
    }
  }
</script>
