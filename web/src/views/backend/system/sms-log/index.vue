<template>
  <div class="sms-log-page art-full-height">
    <ArtSearchBar v-model="searchForm" :items="searchItems as any" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data as any"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchSmsLogList } from '@/api/backend/system/sms'
  import { ElTag } from 'element-plus'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'SmsLogList' })

  const searchForm = ref({ phone: '', templateCode: '', status: undefined as number | undefined, driver: '' })

  const searchItems = computed((): any[] => [
    { label: '手机号', key: 'phone', type: 'input', props: { clearable: true, placeholder: '手机号' } },
    { label: '模板标识', key: 'templateCode', type: 'input', props: { clearable: true, placeholder: '模板标识' } },
    {
      label: '状态', key: 'status', type: 'select',
      props: { clearable: true, placeholder: '全部' },
      options: [{ label: '成功', value: 1 }, { label: '失败', value: 0 }]
    },
    {
      label: '驱动', key: 'driver', type: 'select',
      props: { clearable: true, placeholder: '全部' },
      options: [{ label: '阿里云', value: 'aliyun' }, { label: '腾讯云', value: 'tencent' }]
    }
  ])

  const { columns, columnChecks, data, loading, pagination, getData, searchParams, resetSearchParams, handleSizeChange, handleCurrentChange, refreshData } = useTable({
    core: {
      apiFn: fetchSmsLogList,
      apiParams: { page: 1, size: 20, ...searchForm.value },
      paginationKey: { current: 'page', size: 'size' },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        { prop: 'phone', label: '手机号', width: 130 },
        { prop: 'templateCode', label: '模板标识', width: 150, showOverflowTooltip: true },
        {
          prop: 'driver', label: '驱动', width: 100, align: 'center',
          formatter: (row: any) => {
            const m: Record<string, string> = { aliyun: '阿里云', tencent: '腾讯云' }
            return h(ElTag, { size: 'small', effect: 'light' }, () => m[row.driver] || row.driver)
          }
        },
        {
          prop: 'status', label: '状态', width: 80, align: 'center',
          formatter: (row: any) => h(ElTag, { type: row.status === 1 ? 'success' : 'danger', size: 'small' }, () => row.status === 1 ? '成功' : '失败')
        },
        { prop: 'content', label: '发送内容', minWidth: 200, showOverflowTooltip: true },
        { prop: 'requestId', label: 'RequestId', width: 180, showOverflowTooltip: true },
        { prop: 'errorMsg', label: '错误信息', width: 200, showOverflowTooltip: true },
        { prop: 'createTime', label: '发送时间', width: 170, formatter: (row: any) => formatTimestamp(row.createTime) }
      ]
    }
  })

  const handleSearch = () => {
    Object.assign(searchParams, {
      ...searchForm.value,
      status: searchForm.value.status ?? -1
    })
    getData()
  }
</script>
