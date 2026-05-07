<!-- 附件管理页面 -->
<template>
  <div class="attachment-page art-full-height">
    <!-- 搜索栏 -->
    <AttachmentSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'add'" @click="showUploadDialog" type="primary" v-ripple>
              <ArtSvgIcon icon="ri:upload-2-line" class="text-sm mr-1" />
              上传文件
            </ElButton>
            <ElButton
              v-if="selectedRows.length"
              v-auth="'batchDel'"
              type="danger"
              @click="handleBatchDelete"
              v-ripple
            >
              批量删除 ({{ selectedRows.length }})
            </ElButton>
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
      />
    </ElCard>

    <!-- 上传/选择文件弹窗（复用 ArtFileSelector 的弹窗） -->
    <ElDialog
      v-model="uploadDialogVisible"
      title="文件管理"
      width="920px"
      align-center
      :close-on-click-modal="false"
      destroy-on-close
    >
      <template #header>
        <div class="upload-dialog-header">
          <span class="upload-dialog-title">文件管理</span>
          <ElButton type="primary" size="small" @click="triggerUpload">
            <ArtSvgIcon icon="ri:upload-2-line" class="text-sm mr-1" />
            上传文件
          </ElButton>
          <input ref="uploadInputRef" type="file" multiple class="hidden" @change="handleFileChange" />
        </div>
      </template>

      <div class="chooser-layout">
        <!-- 左侧分类 -->
        <div class="chooser-sidebar">
          <div
            v-for="cat in categories"
            :key="cat.value"
            class="cat-item"
            :class="{ active: activeCategory === cat.value }"
            @click="selectCategory(cat.value)"
          >
            <ArtSvgIcon :icon="cat.icon" class="cat-icon" />
            <span>{{ cat.label }}</span>
          </div>
        </div>

        <!-- 右侧内容 -->
        <div class="chooser-main">
          <!-- 搜索栏 -->
          <div class="chooser-search">
            <ElInput v-model="dialogSearchName" placeholder="搜索文件名" clearable size="small" style="width: 200px" @clear="loadDialogFiles" @keyup.enter="loadDialogFiles" />
            <ElButton size="small" type="primary" @click="loadDialogFiles">搜索</ElButton>
            <ElButton size="small" @click="dialogSearchName = ''; loadDialogFiles()">重置</ElButton>
          </div>

          <!-- 文件网格 -->
          <div class="chooser-grid" v-loading="dialogLoading">
            <div v-if="dialogFiles.length === 0 && !dialogLoading" class="chooser-empty">
              <ArtSvgIcon icon="ri:inbox-line" class="text-5xl text-gray-300" />
              <p>无数据</p>
            </div>
            <div
              v-for="item in dialogFiles"
              :key="item.id"
              class="file-card"
            >
              <div class="file-card__preview">
                <ElImage
                  v-if="item.mimetype?.startsWith('image/')"
                  :src="item.url"
                  fit="cover"
                  class="file-card__img"
                  :preview-src-list="[item.url]"
                  preview-teleported
                />
                <div v-else class="file-card__icon">
                  <ArtSvgIcon :icon="getMimeIcon(item.mimetype)" class="text-3xl" />
                </div>
                <!-- 操作浮层 -->
                <div class="file-card__actions">
                  <ArtSvgIcon icon="ri:file-copy-line" class="action-btn" title="复制链接" @click.stop="handleCopyUrl(item)" />
                  <ArtSvgIcon icon="ri:download-line" class="action-btn" title="下载" @click.stop="handleDownload(item)" />
                  <ArtSvgIcon icon="ri:delete-bin-line" class="action-btn action-btn--danger" title="删除" @click.stop="handleDialogDelete(item)" />
                </div>
              </div>
              <div class="file-card__name" :title="item.name">{{ item.name }}</div>
              <div class="file-card__meta">{{ formatSize(item.size) }}</div>
            </div>
          </div>

          <!-- 分页 -->
          <div class="chooser-pagination">
            <ElPagination
              v-model:current-page="dialogPage"
              v-model:page-size="dialogPageSize"
              :total="dialogTotal"
              :page-sizes="[12, 24, 36, 48]"
              layout="total, sizes, prev, pager, next"
              small
              @current-change="loadDialogFiles"
              @size-change="loadDialogFiles"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <ElButton @click="uploadDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchAttachmentList, fetchDeleteAttachment } from '@/api/backend/common/attachment'
  import { uploadFileApi } from '@/api/backend/common/upload'
  import AttachmentSearch from './modules/attachment-search.vue'
  import { ElTag, ElMessageBox, ElImage, ElButton } from 'element-plus'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { formatTimestamp } from '@/utils/time'

  defineOptions({ name: 'Attachment' })
  const { hasAuth } = useAuth()

  interface AttachmentItem {
    id: number
    topic: string
    userId: number
    url: string
    name: string
    size: number
    mimetype: string
    storage: string
    sha1: string
    quote: number
    width: number
    height: number
    createTime: number
    updateTime: number
  }

  // 搜索表单
  const searchForm = ref({ topic: '', storage: '' })

  // 选中行
  const selectedRows = ref<AttachmentItem[]>([])

  // ==================== 列表 ====================
  const {
    columns, columnChecks, data, loading, pagination,
    getData, searchParams, resetSearchParams,
    handleSizeChange, handleCurrentChange, refreshData
  } = useTable({
    core: {
      apiFn: fetchAttachmentList,
      apiParams: { page: 1, pageSize: 20, ...searchForm.value },
      paginationKey: { current: 'page', size: 'pageSize' },
      columnsFactory: () => [
        { type: 'selection' },
        { type: 'index', width: 60, label: '序号' },
        {
          prop: 'preview', label: '预览', width: 80, align: 'center',
          formatter: (row: AttachmentItem) =>
            row.mimetype?.startsWith('image/')
              ? h(ElImage, { class: 'w-12 h-12 rounded', src: row.url, previewSrcList: [row.url], previewTeleported: true, fit: 'cover' })
              : h('span', { class: 'text-gray-400 text-xs' }, getFileEmoji(row.mimetype))
        },
        { prop: 'name', label: '文件名', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'topic', label: '分类', width: 90, align: 'center',
          formatter: (row: AttachmentItem) => {
            const m: Record<string, { label: string; type: any }> = {
              image: { label: '图片', type: 'success' }, video: { label: '视频', type: 'warning' },
              audio: { label: '音频', type: 'info' }, doc: { label: '文档', type: 'primary' },
              archive: { label: '压缩包', type: 'danger' }, other: { label: '其他', type: '' }
            }
            const c = m[row.topic] || m.other
            return h(ElTag, { type: c.type, size: 'small', effect: 'light', round: true }, () => c.label)
          }
        },
        { prop: 'size', label: '大小', width: 100, align: 'center', formatter: (row: AttachmentItem) => formatSize(row.size) },
        {
          prop: 'storage', label: '存储', width: 80, align: 'center',
          formatter: (row: AttachmentItem) => ({ local: '本地', oss: '阿里云', cos: '腾讯云', qiniu: '七牛' }[row.storage] || row.storage || '-')
        },
        { prop: 'createTime', label: '上传时间', width: 170, formatter: (row: AttachmentItem) => formatTimestamp(row.createTime) },
        {
          prop: 'operation', label: '操作', width: 200, fixed: 'right',
          formatter: (row: AttachmentItem) => h('div', { class: 'flex items-center gap-1' }, [
            h(ElButton, { type: 'primary', link: true, size: 'small', onClick: () => handleCopyUrl(row) }, () => '复制'),
            h(ElButton, { type: 'success', link: true, size: 'small', onClick: () => handleDownload(row) }, () => '下载'),
            hasAuth('delete') ? h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) }) : null,
          ].filter(Boolean))
        }
      ]
    }
  })

  // ==================== 上传弹窗 ====================
  const uploadDialogVisible = ref(false)
  const uploadInputRef = ref<HTMLInputElement>()
  const dialogLoading = ref(false)
  const dialogFiles = ref<AttachmentItem[]>([])
  const dialogPage = ref(1)
  const dialogPageSize = ref(24)
  const dialogTotal = ref(0)
  const dialogSearchName = ref('')
  const activeCategory = ref('')

  const categories = [
    { value: '', label: '全部', icon: 'ri:apps-line' },
    { value: 'image', label: '图片', icon: 'ri:image-line' },
    { value: 'doc', label: '文档', icon: 'ri:file-text-line' },
    { value: 'audio', label: '音频', icon: 'ri:music-line' },
    { value: 'video', label: '视频', icon: 'ri:video-line' },
    { value: 'archive', label: '压缩包', icon: 'ri:file-zip-line' },
    { value: 'other', label: '其他', icon: 'ri:add-circle-line' },
  ]

  const showUploadDialog = () => {
    uploadDialogVisible.value = true
    loadDialogFiles()
  }

  const loadDialogFiles = async () => {
    dialogLoading.value = true
    try {
      const params: any = { page: dialogPage.value, pageSize: dialogPageSize.value }
      if (activeCategory.value) params.topic = activeCategory.value
      if (dialogSearchName.value) params.name = dialogSearchName.value
      const res = await fetchAttachmentList(params)
      dialogFiles.value = (res as any).list || []
      dialogTotal.value = (res as any).total || 0
    } catch { /* ignore */ }
    dialogLoading.value = false
  }

  const selectCategory = (cat: string) => {
    activeCategory.value = cat
    dialogPage.value = 1
    loadDialogFiles()
  }

  const triggerUpload = () => uploadInputRef.value?.click()

  const handleFileChange = async (e: Event) => {
    const input = e.target as HTMLInputElement
    if (!input.files?.length) return
    for (const file of Array.from(input.files)) {
      try {
        const res = await uploadFileApi(file)
        if ((res as any)?.url) ElMessage.success(`${file.name} 上传成功`)
      } catch { ElMessage.error(`${file.name} 上传失败`) }
    }
    input.value = ''
    loadDialogFiles()
    refreshData() // 同步刷新列表
  }

  const handleDialogDelete = async (item: AttachmentItem) => {
    try {
      await ElMessageBox.confirm(`确定删除"${item.name}"？`, '删除', { type: 'warning' })
      await fetchDeleteAttachment(item.id)
      ElMessage.success('删除成功')
      loadDialogFiles()
      refreshData()
    } catch { /* cancel */ }
  }

  // ==================== 通用操作 ====================
  const handleSearch = () => { Object.assign(searchParams, searchForm.value); getData() }
  const handleSelectionChange = (rows: any[]) => { selectedRows.value = rows }

  const handleCopyUrl = async (row: AttachmentItem) => {
    try {
      const url = row.url.startsWith('http') ? row.url : window.location.origin + row.url
      await navigator.clipboard.writeText(url)
      ElMessage.success('链接已复制')
    } catch { ElMessage.error('复制失败') }
  }

  const handleDownload = (row: AttachmentItem) => {
    const url = row.url.startsWith('http') ? row.url : window.location.origin + row.url
    const a = document.createElement('a')
    a.href = url; a.download = row.name || 'download'; a.target = '_blank'
    document.body.appendChild(a); a.click(); document.body.removeChild(a)
  }

  const handleDelete = async (row: AttachmentItem) => {
    try {
      await ElMessageBox.confirm(`确定删除"${row.name}"？`, '删除确认', { type: 'warning' })
      await fetchDeleteAttachment(row.id)
      ElMessage.success('删除成功')
      refreshData()
    } catch (e) { if (e !== 'cancel') console.error(e) }
  }

  const handleBatchDelete = async () => {
    try {
      await ElMessageBox.confirm(`确定删除选中的 ${selectedRows.value.length} 个文件？`, '批量删除', { type: 'warning' })
      for (const row of selectedRows.value) await fetchDeleteAttachment(row.id)
      ElMessage.success(`已删除 ${selectedRows.value.length} 个文件`)
      selectedRows.value = []
      refreshData()
    } catch (e) { if (e !== 'cancel') console.error(e) }
  }

  // ==================== 工具 ====================
  const formatSize = (size: number): string => {
    if (!size) return '0 B'
    const k = 1024, s = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(size) / Math.log(k))
    return (size / Math.pow(k, i)).toFixed(1) + ' ' + s[i]
  }

  const getFileEmoji = (mime: string): string => {
    if (mime?.startsWith('video/')) return '🎬'
    if (mime?.startsWith('audio/')) return '🎵'
    if (mime?.includes('pdf')) return '📄'
    if (mime?.includes('word') || mime?.includes('document')) return '📝'
    if (mime?.includes('sheet') || mime?.includes('excel')) return '📊'
    if (mime?.includes('zip') || mime?.includes('rar') || mime?.includes('7z')) return '📦'
    return '📎'
  }

  const getMimeIcon = (mime: string): string => {
    if (mime?.startsWith('video/')) return 'ri:video-line'
    if (mime?.startsWith('audio/')) return 'ri:music-line'
    if (mime?.includes('pdf')) return 'ri:file-pdf-2-line'
    if (mime?.includes('word')) return 'ri:file-word-line'
    if (mime?.includes('sheet') || mime?.includes('excel')) return 'ri:file-excel-line'
    if (mime?.includes('zip') || mime?.includes('rar')) return 'ri:file-zip-line'
    return 'ri:file-line'
  }
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .hidden { display: none; }

  .upload-dialog-header {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .upload-dialog-title {
    font-size: 16px;
    font-weight: 600;
  }

  .chooser-layout {
    display: flex;
    gap: 16px;
    height: 520px;
  }

  .chooser-sidebar {
    width: 110px;
    flex-shrink: 0;
    border-right: 1px solid var(--el-border-color-lighter);
    padding-right: 12px;
    overflow-y: auto;
  }

  .cat-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 13px;
    color: var(--el-text-color-regular);
    transition: all 0.15s;
    margin-bottom: 2px;
  }

  .cat-item:hover { background: var(--el-fill-color-lighter); }

  .cat-item.active {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-weight: 500;
  }

  .cat-icon { font-size: 16px; flex-shrink: 0; }

  .chooser-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .chooser-search {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
    flex-shrink: 0;
  }

  .chooser-grid {
    flex: 1;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
    gap: 10px;
    align-content: start;
    padding: 2px;
  }

  .chooser-empty {
    grid-column: 1 / -1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 300px;
    color: var(--el-text-color-placeholder);
  }

  .chooser-empty p { margin-top: 8px; font-size: 13px; }

  .file-card {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    overflow: hidden;
    transition: all 0.15s;
  }

  .file-card:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }

  .file-card:hover .file-card__actions { opacity: 1; }

  .file-card__preview {
    position: relative;
    height: 110px;
    background: var(--el-fill-color-lighter);
  }

  .file-card__img { width: 100%; height: 100%; display: block; }

  .file-card__icon {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-text-color-placeholder);
  }

  .file-card__actions {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    background: rgba(0, 0, 0, 0.45);
    opacity: 0;
    transition: opacity 0.2s;
  }

  .action-btn {
    color: #fff;
    font-size: 18px;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: background 0.15s;
  }

  .action-btn:hover { background: rgba(255, 255, 255, 0.2); }
  .action-btn--danger:hover { background: rgba(245, 108, 108, 0.6); }

  .file-card__name {
    padding: 4px 8px 0;
    font-size: 12px;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .file-card__meta {
    padding: 0 8px 6px;
    font-size: 11px;
    color: var(--el-text-color-placeholder);
  }

  .chooser-pagination {
    flex-shrink: 0;
    padding-top: 12px;
    display: flex;
    justify-content: flex-end;
  }
</style>
