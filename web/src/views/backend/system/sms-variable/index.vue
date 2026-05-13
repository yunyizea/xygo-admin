<template>
  <div class="sms-variable-page art-full-height">
    <ArtSearchBar v-model="searchForm" :items="searchItems as any" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton type="primary" @click="showDialog('add')" v-ripple>
            <ArtSvgIcon icon="ri:add-line" class="text-sm mr-1" />
            新增变量
          </ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data as any"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- 新增/编辑弹窗 -->
    <ElDialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增变量' : '编辑变量'" width="580px" :close-on-click-modal="false" @close="resetForm">
      <ElForm ref="formRef" :model="formData" :rules="rules" label-width="100px">
        <ElFormItem label="变量标题" prop="title">
          <ElInput v-model="formData.title" placeholder="如：联系人手机号" />
        </ElFormItem>
        <ElFormItem label="变量名" prop="name">
          <ElInput v-model="formData.name" placeholder="如：usermobile" :disabled="dialogType === 'edit'" />
        </ElFormItem>
        <ElFormItem label="来源类型" prop="sourceType">
          <ElSelect v-model="formData.sourceType" placeholder="选择来源">
            <ElOption :value="1" label="字段提取" />
            <ElOption :value="2" label="SQL查询" />
            <ElOption :value="3" label="内置Helper" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem v-if="formData.sourceType === 2" label="SQL查询">
          <ElInput v-model="formData.sqlQuery" type="textarea" :rows="3" placeholder="SELECT value FROM ..." />
        </ElFormItem>
        <ElFormItem v-if="formData.sourceType === 3" label="Helper方法">
          <ElInput v-model="formData.methodName" placeholder="如：sms.GetUserMobile" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSwitch v-model="formData.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchSmsVariableList, fetchSmsVariableSave, fetchSmsVariableDelete } from '@/api/backend/system/sms'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { ElTag, ElMessageBox } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'SmsVariableManage' })

  const searchForm = ref({ name: '' })

  const searchItems = computed((): any[] => [
    { label: '变量名', key: 'name', type: 'input', props: { clearable: true, placeholder: '关键词' } }
  ])

  const sourceTypeMap: Record<number, string> = { 1: '字段提取', 2: 'SQL查询', 3: '内置Helper' }

  const { columns, columnChecks, data, loading, pagination, getData, searchParams, resetSearchParams, handleSizeChange, handleCurrentChange, refreshData } = useTable({
    core: {
      apiFn: fetchSmsVariableList,
      apiParams: { page: 1, size: 20, ...searchForm.value },
      paginationKey: { current: 'page', size: 'size' },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        { prop: 'title', label: '变量标题', minWidth: 160 },
        { prop: 'name', label: '变量名', width: 160 },
        {
          prop: 'sourceType', label: '来源类型', width: 110, align: 'center',
          formatter: (row: any) => h(ElTag, { size: 'small', effect: 'light' }, () => sourceTypeMap[row.sourceType] || '未知')
        },
        {
          prop: 'status', label: '状态', width: 80, align: 'center',
          formatter: (row: any) => h(ElTag, { type: row.status === 1 ? 'success' : 'danger', size: 'small' }, () => row.status === 1 ? '启用' : '禁用')
        },
        { prop: 'sharedCount', label: '共通数', width: 80, align: 'center' },
        { prop: 'createTime', label: '创建时间', width: 170, formatter: (row: any) => formatTimestamp(row.createTime) },
        {
          prop: 'operation', label: '操作', width: 140, fixed: 'right',
          formatter: (row: any) => h('div', { class: 'flex items-center gap-1' }, [
            h(ArtButtonTable, { type: 'edit', onClick: () => showDialog('edit', row) }),
            h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
          ])
        }
      ]
    }
  })

  const handleSearch = () => { Object.assign(searchParams, searchForm.value); getData() }

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const formRef = ref<FormInstance>()
  const submitLoading = ref(false)

  const defaultForm = () => ({ id: 0, title: '', name: '', sourceType: 1, sqlQuery: '', methodName: '', status: 1 })
  const formData = reactive(defaultForm())
  const rules: FormRules = {
    title: [{ required: true, message: '变量标题不能为空', trigger: 'blur' }],
    name: [{ required: true, message: '变量名不能为空', trigger: 'blur' }],
  }

  const showDialog = (type: 'add' | 'edit', row?: any) => {
    dialogType.value = type
    if (type === 'edit' && row) Object.assign(formData, row)
    else Object.assign(formData, defaultForm())
    dialogVisible.value = true
  }

  const resetForm = () => {
    formRef.value?.resetFields()
    Object.assign(formData, defaultForm())
  }

  const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate()
    submitLoading.value = true
    try {
      await fetchSmsVariableSave({ ...formData })
      ElMessage.success(dialogType.value === 'add' ? '新增成功' : '编辑成功')
      dialogVisible.value = false
      refreshData()
    } catch (e) { console.error(e) }
    submitLoading.value = false
  }

  const handleDelete = async (row: any) => {
    try {
      await ElMessageBox.confirm(`确定删除变量「${row.title}」？`, '删除确认', { type: 'warning' })
      await fetchSmsVariableDelete(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch { /* cancel */ }
  }
</script>
