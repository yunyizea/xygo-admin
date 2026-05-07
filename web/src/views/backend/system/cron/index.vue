<!-- 定时任务管理页面 -->
<template>
  <div class="cron-page art-full-height">
    <!-- 搜索栏 -->
    <ElCard shadow="never" class="mb-4">
      <ElForm :model="searchForm" inline>
        <ElFormItem label="分组">
          <ElSelect v-model="searchForm.groupId" clearable placeholder="全部分组" style="width: 160px">
            <ElOption v-for="g in groupOptions" :key="g.id" :label="g.name" :value="g.id" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="searchForm.status" clearable placeholder="全部" style="width: 120px">
            <ElOption label="启用" :value="1" />
            <ElOption label="禁用" :value="0" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="任务标识">
          <ElInput v-model="searchForm.name" placeholder="模糊搜索" clearable style="width: 180px" />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="loadList">搜索</ElButton>
          <ElButton @click="resetSearch">重置</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <ElCard shadow="never">
      <!-- 工具栏 -->
      <div class="mb-4 flex-cb">
        <div class="flex-c gap-2">
          <ElButton v-auth="'add'" type="primary" @click="handleAdd">新增任务</ElButton>
          <ElButton @click="showGroupDialog = true">分组管理</ElButton>
          <ElButton @click="showLogDialog = true">执行日志</ElButton>
        </div>
        <ElButton :icon="Refresh" circle @click="loadList" />
      </div>

      <!-- 任务列表 -->
      <ElTable :data="tableData" v-loading="loading" border stripe row-key="id">
        <ElTableColumn prop="id" label="ID" width="70" align="center" />
        <ElTableColumn prop="title" label="任务标题" min-width="140" show-overflow-tooltip />
        <ElTableColumn prop="name" label="任务标识" min-width="120">
          <template #default="{ row }">
            <code class="text-xs text-blue-600">{{ row.name }}</code>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="groupName" label="分组" width="100" />
        <ElTableColumn prop="pattern" label="Cron表达式" width="140">
          <template #default="{ row }">
            <code class="text-xs">{{ row.pattern }}</code>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="policy" label="策略" width="90" align="center">
          <template #default="{ row }">
            <ElTag :type="policyTagType(row.policy)" size="small">{{ policyLabel(row.policy) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <ElSwitch
              :model-value="row.status === 1"
              @change="(val: any) => handleStatusChange(row, !!val)"
              inline-prompt
              active-text="启"
              inactive-text="停"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="200" align="right" fixed="right">
          <template #default="{ row }">
            <div class="flex flex-nowrap items-center justify-end">
              <ArtButtonTable v-if="hasAuth('exec')" icon="ri:play-line" icon-color="#67c23a" button-bg-color="rgba(103,194,58,0.12)" title="立即执行" @click="handleExec(row)" />
              <ArtButtonTable v-if="hasAuth('view')" type="view" title="执行日志" @click="handleViewLog(row)" />
              <ArtButtonTable v-if="hasAuth('edit')" type="edit" title="编辑" @click="handleEdit(row)" />
              <ArtButtonTable v-if="hasAuth('delete')" type="delete" title="删除" @click="handleDelete(row)" />
            </div>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="mt-4 flex justify-end">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadList"
        />
      </div>
    </ElCard>

    <!-- 新增/编辑弹窗 -->
    <CronDialog
      v-model:visible="showEditDialog"
      :data="editData"
      :group-options="groupOptions"
      @success="loadList"
    />

    <!-- 分组管理弹窗 -->
    <CronGroupDialog v-model:visible="showGroupDialog" @change="loadGroupOptions" />

    <!-- 执行日志弹窗 -->
    <CronLogDialog v-model:visible="showLogDialog" :cron-id="logCronId" :cron-title="logCronTitle" />
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useAuth } from '@/hooks/core/useAuth'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import {
    fetchCronList,
    fetchCronDelete,
    fetchCronStatus,
    fetchCronOnlineExec,
    fetchCronGroupSelect,
    type CronItem
  } from '@/api/backend/system/cron'
  import CronDialog from './modules/cron-dialog.vue'
  import CronGroupDialog from './modules/cron-group-dialog.vue'
  import CronLogDialog from './modules/cron-log-dialog.vue'

  defineOptions({ name: 'CronManage' })
  const { hasAuth } = useAuth()

  const loading = ref(false)
  const tableData = ref<CronItem[]>([])
  const groupOptions = ref<{ id: number; name: string }[]>([])

  const searchForm = reactive({ groupId: undefined as number | undefined, status: undefined as number | undefined, name: '' })
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

  const showEditDialog = ref(false)
  const editData = ref<CronItem | null>(null)
  const showGroupDialog = ref(false)
  const showLogDialog = ref(false)
  const logCronId = ref(0)
  const logCronTitle = ref('')

  const policyLabel = (p: number) => ({ 1: '并行', 2: '单例', 3: '单次', 4: '固定次数' }[p] || '未知')
  const policyTagType = (p: number) => ({ 1: '', 2: 'warning', 3: 'info', 4: 'success' }[p] || 'info') as any

  const loadList = async () => {
    loading.value = true
    try {
      const res = await fetchCronList({
        ...searchForm,
        page: pagination.page,
        pageSize: pagination.pageSize
      }) as any
      tableData.value = res?.list || []
      pagination.total = res?.total || 0
    } catch { /* ignore */ } finally {
      loading.value = false
    }
  }

  const loadGroupOptions = async () => {
    try {
      const res = await fetchCronGroupSelect() as any
      groupOptions.value = res?.list || []
    } catch { /* ignore */ }
  }

  const resetSearch = () => {
    searchForm.groupId = undefined
    searchForm.status = undefined
    searchForm.name = ''
    pagination.page = 1
    loadList()
  }

  const handleAdd = () => {
    editData.value = null
    showEditDialog.value = true
  }

  const handleEdit = (row: CronItem) => {
    editData.value = { ...row }
    showEditDialog.value = true
  }

  const handleDelete = async (row: CronItem) => {
    await ElMessageBox.confirm(`确定要删除任务「${row.title}」吗？`, '删除确认', { type: 'warning' })
    await fetchCronDelete(row.id)
    ElMessage.success('删除成功')
    loadList()
  }

  const handleStatusChange = async (row: CronItem, val: boolean) => {
    const status = val ? 1 : 0
    await fetchCronStatus({ id: row.id, status })
    ElMessage.success(val ? '已启用' : '已禁用')
    loadList()
  }

  const handleExec = async (row: CronItem) => {
    await ElMessageBox.confirm(`确定要立即执行「${row.title}」吗？`, '手动执行', { type: 'info' })
    try {
      const res = await fetchCronOnlineExec(row.id) as any
      ElMessage.success(`执行完成：${res?.output || '无输出'}`)
    } catch (e: any) {
      ElMessage.error(`执行失败：${e?.message || '未知错误'}`)
    }
  }

  const handleViewLog = (row: CronItem) => {
    logCronId.value = row.id
    logCronTitle.value = row.title
    showLogDialog.value = true
  }

  onMounted(() => {
    loadGroupOptions()
    loadList()
  })
</script>
