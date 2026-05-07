<!-- 部门管理页面 -->
<template>
  <div class="dept-page art-full-height">
    <!-- 搜索栏 -->
    <DeptSearch v-model="searchForm" @search="handleSearch" @reset="handleReset" />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader :showZebra="false" :loading="loading" v-model:columns="columnChecks" @refresh="getDeptList">
        <template #left>
          <ElButton v-auth="'add'" @click="handleAddDept" v-ripple>添加部门</ElButton>
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
        :default-expand-all="true"
      />

      <!-- 部门弹窗 -->
      <DeptDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :dept-data="editData"
        :parent-dept="parentDept"
        @submit="handleSubmit"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import { useAuth } from '@/hooks/core/useAuth'
  import DeptSearch from './modules/dept-search.vue'
  import DeptDialog from './modules/dept-dialog.vue'
  import { fetchGetDeptList, fetchSaveDept, fetchDeleteDept } from '@/api/backend/system'
  import { ElTag, ElMessageBox, ElSwitch } from 'element-plus'
  import { DialogType } from '@/types'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'Dept' })
  const { hasAuth } = useAuth()

  interface DeptListItem {
    id: number
    parentId: number
    name: string
    sort: number
    status: number
    remark: string
    create_time: number
    update_time: number
    children?: DeptListItem[]
  }

  const loading = ref(false)
  const isExpanded = ref(true) // 默认展开
  const tableRef = ref()

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const editData = ref<DeptListItem | null>(null)
  const parentDept = ref<DeptListItem | null>(null)

  const searchForm = ref({
    name: '',
    status: undefined
  })

  const appliedFilters = reactive({ ...searchForm.value })

  const tableData = ref<DeptListItem[]>([])

  onMounted(() => {
    getDeptList()
  })

  /**
   * 获取部门列表
   */
  const getDeptList = async (): Promise<void> => {
    loading.value = true
    try {
      const list = await fetchGetDeptList(appliedFilters)
      tableData.value = list
    } catch (error) {
      ElMessage.error('获取部门列表失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 搜索
   */
  const handleSearch = (): void => {
    Object.assign(appliedFilters, { ...searchForm.value })
    getDeptList()
  }

  /**
   * 重置
   */
  const handleReset = (): void => {
    searchForm.value = { name: '', status: undefined }
    Object.assign(appliedFilters, { ...searchForm.value })
    getDeptList()
  }

  // 表格列配置
  const { columnChecks, columns } = useTableColumns(() => [
    {
      prop: 'name',
      label: '部门名称',
      minWidth: 200
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
      formatter: (row: DeptListItem) =>
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
      formatter: (row: DeptListItem) => row.remark || '-'
    },
    {
      prop: 'create_time',
      label: '创建时间',
      width: 180,
      formatter: (row: DeptListItem) => formatTimestamp(row.create_time)
    },
    {
      prop: 'operation',
      label: '操作',
      width: 180,
      align: 'right',
      formatter: (row: DeptListItem) =>
        h('div', { style: 'text-align: right' }, [
          hasAuth('add') ? h(ArtButtonTable, {
            type: 'add',
            onClick: () => handleAddSubDept(row),
            title: '添加子部门'
          }) : null,
          hasAuth('edit') ? h(ArtButtonTable, {
            type: 'edit',
            onClick: () => handleEditDept(row)
          }) : null,
          hasAuth('delete') ? h(ArtButtonTable, {
            type: 'delete',
            onClick: () => handleDeleteDept(row)
          }) : null,
        ].filter(Boolean))
    }
  ])

  const filteredTableData = computed(() => tableData.value)

  /**
   * 添加顶级部门
   */
  const handleAddDept = (): void => {
    dialogType.value = 'add'
    editData.value = null
    parentDept.value = null
    dialogVisible.value = true
  }

  /**
   * 添加子部门
   */
  const handleAddSubDept = (row: DeptListItem): void => {
    dialogType.value = 'add'
    editData.value = null
    parentDept.value = row
    dialogVisible.value = true
  }

  /**
   * 编辑部门
   */
  const handleEditDept = (row: DeptListItem): void => {
    dialogType.value = 'edit'
    editData.value = row
    parentDept.value = null
    dialogVisible.value = true
  }

  /**
   * 删除部门
   */
  const handleDeleteDept = async (row: DeptListItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除部门"${row.name}"吗？删除后无法恢复`, '删除部门', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })

      await fetchDeleteDept(row.id)
      ElMessage.success('删除成功')
      getDeptList()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除部门失败:', error)
      }
    }
  }

  /**
   * 提交表单
   */
  const handleSubmit = async (formData: any): Promise<void> => {
    try {
      await fetchSaveDept(formData)
      ElMessage.success(formData.id ? '编辑成功' : '添加成功')
      getDeptList()
    } catch (error) {
      console.error('保存部门失败:', error)
    }
  }

  /**
   * 处理状态切换
   */
  const handleStatusChange = async (row: DeptListItem, value: string | number | boolean): Promise<void> => {
    try {
      const boolValue = !!value
      await fetchSaveDept({
        id: row.id,
        parentId: row.parentId,
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
   * 切换展开/收起
   */
  const toggleExpand = (): void => {
    isExpanded.value = !isExpanded.value
    nextTick(() => {
      if (tableRef.value?.elTableRef && filteredTableData.value) {
        const processRows = (rows: DeptListItem[]) => {
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
