<template>
  <div class="sms-template-page art-full-height">
    <ArtSearchBar v-model="searchForm" :items="searchItems as any" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton type="primary" @click="showDialog('add')" v-ripple>
            <ArtSvgIcon icon="ri:add-line" class="text-sm mr-1" />
            新增模板
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
    <ElDialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增模板' : '编辑模板'" width="680px" :close-on-click-modal="false" @close="resetForm">
      <ElForm ref="formRef" :model="formData" :rules="rules" label-width="120px">
        <ElFormItem label="模板标题" prop="title">
          <ElInput v-model="formData.title" placeholder="如：用户注册验证码" />
        </ElFormItem>
        <ElFormItem label="唯一标识" prop="code">
          <ElInput v-model="formData.code" placeholder="如：user_register" :disabled="dialogType === 'edit'" />
          <div class="text-xs text-gray-400 mt-1">可在业务代码中使用唯一标识调取本模板发送短信</div>
        </ElFormItem>
        <ElFormItem label="短信内容" prop="content">
          <ElInput v-model="formData.content" type="textarea" :rows="4" placeholder="短信文案，变量用 ${var} 占位" />
          <div class="text-xs text-gray-400 mt-1">可使用模板变量：${var_name}</div>
        </ElFormItem>
        <ElFormItem label="服务商模板ID" prop="providerTemplateId">
          <ElInput v-model="formData.providerTemplateId" placeholder="有的服务商需要使用模板ID来发送短信，请按需填写" />
          <div class="text-xs text-gray-400 mt-1">有的服务商需要使用模板ID来发送短信，请按需填写</div>
        </ElFormItem>
        <ElFormItem label="模板变量">
          <ElSelect
            v-model="formData.selectedVarIds"
            multiple
            filterable
            placeholder="选择关联的模板变量"
            style="width: 100%"
            :loading="variableLoading"
          >
            <ElOption
              v-for="v in variableOptions"
              :key="v.id"
              :label="v.title"
              :value="v.id"
            />
          </ElSelect>
          <div class="text-xs text-gray-400 mt-1">选择此模板使用的变量，也可直接在短信内容中使用模板变量的变量</div>
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="formData.sort" :min="0" controls-position="right" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSwitch v-model="formData.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput v-model="formData.remark" type="textarea" :rows="2" placeholder="可选" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">确定</ElButton>
      </template>
    </ElDialog>

    <!-- 测试发送弹窗 -->
    <ElDialog v-model="testDialogVisible" title="测试发送" width="420px" :close-on-click-modal="false">
      <ElForm label-width="80px">
        <ElFormItem label="手机号">
          <ElInput v-model="testPhone" placeholder="输入接收手机号" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="testDialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="testLoading" @click="handleTest">发送</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchSmsTemplateList, fetchSmsTemplateSave, fetchSmsTemplateDelete, fetchSmsTemplateTest, fetchSmsVariableList } from '@/api/backend/system/sms'
  import type { SmsVariableItem } from '@/api/backend/system/sms'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { ElButton, ElTag, ElMessageBox } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { formatTimestamp } from '@/utils/time'

  import { onMounted } from 'vue'

  defineOptions({ name: 'SmsTemplateManage' })

  const searchForm = ref({ status: undefined as number | undefined, code: '', title: '' })

  const searchItems = computed((): any[] => [
    { label: '模板标题', key: 'title', type: 'input', props: { clearable: true, placeholder: '关键词' } },
    { label: '模板标识', key: 'code', type: 'input', props: { clearable: true, placeholder: '如 user_register' } },
    {
      label: '状态', key: 'status', type: 'select',
      props: { clearable: true, placeholder: '全部' },
      options: [{ label: '启用', value: 1 }, { label: '禁用', value: 0 }]
    }
  ])

  const variableOptions = ref<SmsVariableItem[]>([])
  const variableLoading = ref(false)

  const loadVariableOptions = async () => {
    variableLoading.value = true
    try {
      const res = await fetchSmsVariableList({ page: 1, size: 200, status: 1 })
      variableOptions.value = res.list || []
    } catch { variableOptions.value = [] }
    variableLoading.value = false
  }

  const { columns, columnChecks, data, loading, pagination, getData, searchParams, resetSearchParams, handleSizeChange, handleCurrentChange, refreshData } = useTable({
    core: {
      apiFn: fetchSmsTemplateList,
      apiParams: { page: 1, size: 20, ...searchForm.value, status: -1 },
      paginationKey: { current: 'page', size: 'size' },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        { prop: 'title', label: '模板标题', minWidth: 160, showOverflowTooltip: true },
        { prop: 'code', label: '模板标识', width: 160, showOverflowTooltip: true },
        { prop: 'providerTemplateId', label: '服务商模板ID', width: 160, showOverflowTooltip: true },
        {
          prop: 'variables', label: '模板变量', minWidth: 200,
          formatter: (row: any) => {
            const vars: string[] = Array.isArray(row.variables) ? row.variables : []
            if (!vars.length) return h('span', { class: 'text-gray-400' }, '—')
            return h('div', { class: 'flex flex-wrap gap-1' },
              vars.map(name => {
                const matched = variableOptions.value.find(v => v.name === name)
                return h(ElTag, { size: 'small', closable: false }, () => matched ? matched.title : name)
              })
            )
          }
        },
        {
          prop: 'status', label: '状态', width: 80, align: 'center',
          formatter: (row: any) => h(ElTag, { type: row.status === 1 ? 'success' : 'danger', size: 'small' }, () => row.status === 1 ? '启用' : '禁用')
        },
        { prop: 'sort', label: '排序', width: 70, align: 'center' },
        { prop: 'createTime', label: '创建时间', width: 170, formatter: (row: any) => formatTimestamp(row.createTime) },
        {
          prop: 'operation', label: '操作', width: 180, fixed: 'right',
          formatter: (row: any) => h('div', { class: 'flex items-center gap-1' }, [
            h(ElButton, { size: 'small', type: 'success', text: true, onClick: () => showTestDialog(row) }, () => '测试'),
            h(ArtButtonTable, { type: 'edit', onClick: () => showDialog('edit', row) }),
            h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
          ])
        }
      ]
    }
  })

  onMounted(() => {
    loadVariableOptions()
  })

  const handleSearch = () => {
    Object.assign(searchParams, {
      ...searchForm.value,
      status: searchForm.value.status ?? -1
    })
    getData()
  }

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const formRef = ref<FormInstance>()
  const submitLoading = ref(false)

  const defaultForm = () => ({ id: 0, title: '', code: '', content: '', providerTemplateId: '', variables: '', selectedVarIds: [] as number[], relatedVariableId: 0, status: 1, sort: 0, remark: '' })
  const formData = reactive(defaultForm())
  const rules: FormRules = {
    title: [{ required: true, message: '模板标题不能为空', trigger: 'blur' }],
    code: [{ required: true, message: '模板标识不能为空', trigger: 'blur' }],
  }

  const showDialog = async (type: 'add' | 'edit', row?: any) => {
    dialogType.value = type
    await loadVariableOptions()

    if (type === 'edit' && row) {
      const varNames: string[] = Array.isArray(row.variables) ? row.variables : []
      const selectedIds = variableOptions.value
        .filter(v => varNames.includes(v.name))
        .map(v => v.id)
      Object.assign(formData, {
        ...row,
        variables: varNames.length ? JSON.stringify(varNames) : '',
        selectedVarIds: selectedIds
      })
    } else {
      Object.assign(formData, defaultForm())
    }
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
      const varNames = variableOptions.value
        .filter(v => formData.selectedVarIds.includes(v.id))
        .map(v => v.name)
      const payload = {
        ...formData,
        variables: varNames.length ? JSON.stringify(varNames) : '[]'
      }
      delete (payload as any).selectedVarIds
      await fetchSmsTemplateSave(payload)
      ElMessage.success(dialogType.value === 'add' ? '新增成功' : '编辑成功')
      dialogVisible.value = false
      refreshData()
    } catch (e) { console.error(e) }
    submitLoading.value = false
  }

  const handleDelete = async (row: any) => {
    try {
      await ElMessageBox.confirm(`确定删除模板「${row.title}」？`, '删除确认', { type: 'warning' })
      await fetchSmsTemplateDelete(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch { /* cancel */ }
  }

  const testDialogVisible = ref(false)
  const testPhone = ref('')
  const testLoading = ref(false)
  const testTemplateId = ref(0)

  const showTestDialog = (row: any) => {
    testTemplateId.value = row.id
    testPhone.value = ''
    testDialogVisible.value = true
  }

  const handleTest = async () => {
    if (!testPhone.value) { ElMessage.warning('请输入手机号'); return }
    testLoading.value = true
    try {
      const res = await fetchSmsTemplateTest({ id: testTemplateId.value, phone: testPhone.value })
      if (res.success) {
        ElMessage.success(`发送成功，RequestId: ${res.requestId}`)
      } else {
        ElMessage.error(`发送失败: ${res.message}`)
      }
      testDialogVisible.value = false
    } catch (e) { console.error(e) }
    testLoading.value = false
  }
</script>
