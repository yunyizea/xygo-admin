<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- ArtFileSelector 全局文件选择器（对标 HotGo FileChooser） -->
<template>
  <div class="art-file-selector">
    <!-- 已选文件预览 -->
    <div class="selector-preview" v-if="fileList.length">
      <div
        v-for="(url, idx) in fileList"
        :key="idx"
        class="preview-card"
        :style="{ width: `${width}px`, height: `${height}px` }"
      >
        <ElImage
          v-if="isImage(url)"
          :src="url"
          :preview-src-list="fileList.filter(isImage)"
          :initial-index="fileList.filter(isImage).indexOf(url)"
          fit="cover"
          preview-teleported
          class="preview-img"
        />
        <div v-else class="preview-file">
          <ArtSvgIcon :icon="getFileTypeIcon(url)" class="text-2xl" />
          <span class="preview-ext">{{ getExt(url) }}</span>
        </div>
        <div class="preview-actions" @click.stop="removeFile(idx)">
          <ArtSvgIcon icon="ri:delete-bin-line" class="action-icon" />
        </div>
      </div>
    </div>

    <!-- 选择按钮 -->
    <ElButton @click="dialogVisible = true" v-ripple>
      <ArtSvgIcon icon="ri:add-line" class="text-sm mr-1" />
      {{ buttonText }}
    </ElButton>

    <!-- 文件选择弹窗 -->
    <ElDialog
      v-model="dialogVisible"
      title="文件选择"
      width="900px"
      top="5vh"
      :close-on-click-modal="false"
      destroy-on-close
      class="file-selector-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <span class="dialog-title">文件选择</span>
          <ElButton type="primary" size="small" @click="triggerUpload">
            <ArtSvgIcon icon="ri:upload-2-line" class="text-sm mr-1" />
            上传文件
          </ElButton>
          <input ref="uploadInputRef" type="file" :accept="acceptStr" :multiple="maxNumber > 1" class="hidden" @change="handleFileChange" />
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
            <ElInput v-model="searchName" placeholder="搜索文件名" clearable size="small" style="width: 200px" @clear="loadFiles" @keyup.enter="loadFiles" />
            <ElButton size="small" type="primary" @click="loadFiles">搜索</ElButton>
            <ElButton size="small" @click="searchName = ''; loadFiles()">重置</ElButton>
          </div>

          <!-- 文件网格 -->
          <div class="chooser-grid" v-loading="filesLoading">
            <div v-if="fileItems.length === 0 && !filesLoading" class="chooser-empty">
              <ArtSvgIcon icon="ri:inbox-line" class="text-5xl text-gray-300" />
              <p>无数据</p>
            </div>
            <div
              v-for="item in fileItems"
              :key="item.id"
              class="file-card"
              :class="{ selected: isSelected(item) }"
              @click="toggleSelect(item)"
            >
              <div class="file-card__preview">
                <ElImage
                  v-if="item.mimetype?.startsWith('image/')"
                  :src="item.url"
                  fit="cover"
                  class="file-card__img"
                  :preview-src-list="[]"
                />
                <div v-else class="file-card__icon">
                  <ArtSvgIcon :icon="getMimeIcon(item.mimetype)" class="text-3xl" />
                </div>
                <!-- 选中角标 -->
                <div v-if="isSelected(item)" class="file-card__check">
                  <ArtSvgIcon icon="ri:check-line" class="text-white text-sm" />
                </div>
              </div>
              <div class="file-card__name" :title="item.name">{{ item.name }}</div>
              <div class="file-card__meta">{{ formatSize(item.size) }}</div>
            </div>
          </div>

          <!-- 分页 -->
          <div class="chooser-pagination">
            <ElPagination
              v-model:current-page="page"
              v-model:page-size="pageSize"
              :total="total"
              :page-sizes="[12, 24, 36, 48]"
              layout="total, sizes, prev, pager, next"
              small
              @current-change="loadFiles"
              @size-change="loadFiles"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <span class="selected-count" v-if="tempSelected.length">已选 {{ tempSelected.length }} / {{ maxNumber }} 个</span>
          <ElButton @click="dialogVisible = false">取消</ElButton>
          <ElButton type="primary" :disabled="!tempSelected.length" @click="confirmSelect">确定</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { uploadFileApi } from '@/api/backend/common/upload'
  import { fetchAttachmentList } from '@/api/backend/common/attachment'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'

  export interface FileItem {
    id: number
    url: string
    name: string
    size: number
    mimetype: string
    topic: string
  }

  const props = withDefaults(defineProps<{
    modelValue?: string | string[]
    maxNumber?: number
    fileType?: 'image' | 'doc' | 'audio' | 'video' | 'archive' | 'all'
    width?: number
    height?: number
  }>(), {
    maxNumber: 1,
    fileType: 'all',
    width: 100,
    height: 100,
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', val: string | string[]): void
  }>()

  const dialogVisible = ref(false)
  const uploadInputRef = ref<HTMLInputElement>()

  // 当前已选文件列表（从 v-model 解析）
  const fileList = computed<string[]>(() => {
    if (!props.modelValue) return []
    if (Array.isArray(props.modelValue)) return props.modelValue.filter(Boolean)
    return props.modelValue.split(',').map(s => s.trim()).filter(Boolean)
  })

  const buttonText = computed(() => {
    const map: Record<string, string> = {
      image: '选择图片', doc: '选择文档', audio: '选择音频',
      video: '选择视频', archive: '选择文件', all: '选择文件'
    }
    return map[props.fileType] || '选择文件'
  })

  const acceptStr = computed(() => {
    const map: Record<string, string> = {
      image: 'image/*', audio: 'audio/*', video: 'video/*',
      doc: '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt',
      archive: '.zip,.rar,.7z,.tar,.gz',
    }
    return map[props.fileType] || '*/*'
  })

  // 分类
  const categories = [
    { value: '', label: '全部', icon: 'ri:apps-line' },
    { value: 'image', label: '图片', icon: 'ri:image-line' },
    { value: 'doc', label: '文档', icon: 'ri:file-text-line' },
    { value: 'audio', label: '音频', icon: 'ri:music-line' },
    { value: 'video', label: '视频', icon: 'ri:video-line' },
    { value: 'archive', label: '压缩包', icon: 'ri:file-zip-line' },
    { value: 'other', label: '其他', icon: 'ri:add-circle-line' },
  ]

  const activeCategory = ref(props.fileType === 'all' ? '' : props.fileType)
  const searchName = ref('')
  const page = ref(1)
  const pageSize = ref(24)
  const total = ref(0)
  const filesLoading = ref(false)
  const fileItems = ref<FileItem[]>([])
  const tempSelected = ref<FileItem[]>([])

  // 加载文件列表
  const loadFiles = async () => {
    filesLoading.value = true
    try {
      const params: any = { page: page.value, pageSize: pageSize.value }
      if (activeCategory.value) params.topic = activeCategory.value
      if (searchName.value) params.name = searchName.value
      const res = await fetchAttachmentList(params)
      fileItems.value = (res as any).list || []
      total.value = (res as any).total || 0
    } catch { /* ignore */ }
    filesLoading.value = false
  }

  const selectCategory = (cat: string) => {
    activeCategory.value = cat
    page.value = 1
    loadFiles()
  }

  // 弹窗打开时加载
  watch(dialogVisible, (v) => {
    if (v) {
      tempSelected.value = []
      loadFiles()
    }
  })

  // 选择/取消
  const isSelected = (item: FileItem) => tempSelected.value.some(s => s.id === item.id)

  const toggleSelect = (item: FileItem) => {
    const idx = tempSelected.value.findIndex(s => s.id === item.id)
    if (idx >= 0) {
      tempSelected.value.splice(idx, 1)
    } else {
      if (tempSelected.value.length >= props.maxNumber) {
        if (props.maxNumber === 1) {
          tempSelected.value = [item]
        } else {
          ElMessage.warning(`最多选择 ${props.maxNumber} 个文件`)
        }
        return
      }
      tempSelected.value.push(item)
    }
  }

  const confirmSelect = () => {
    const urls = tempSelected.value.map(f => f.url)
    const existing = fileList.value.slice()
    const merged = [...existing, ...urls].slice(0, props.maxNumber)

    if (props.maxNumber === 1) {
      emit('update:modelValue', merged[0] || '')
    } else {
      emit('update:modelValue', merged.join(','))
    }
    dialogVisible.value = false
  }

  const removeFile = (idx: number) => {
    const list = [...fileList.value]
    list.splice(idx, 1)
    if (props.maxNumber === 1) {
      emit('update:modelValue', '')
    } else {
      emit('update:modelValue', list.join(','))
    }
  }

  // 上传
  const triggerUpload = () => uploadInputRef.value?.click()

  const handleFileChange = async (e: Event) => {
    const input = e.target as HTMLInputElement
    if (!input.files?.length) return
    for (const file of Array.from(input.files)) {
      try {
        const res = await uploadFileApi(file)
        if ((res as any)?.url) {
          ElMessage.success(`${file.name} 上传成功`)
          loadFiles() // 刷新列表
        }
      } catch { ElMessage.error(`${file.name} 上传失败`) }
    }
    input.value = '' // 重置
  }

  // 工具函数
  const isImage = (url: string) => /\.(jpg|jpeg|png|gif|webp|bmp|svg)(\?.*)?$/i.test(url)
  const getExt = (url: string) => url.split('.').pop()?.split('?')[0]?.toUpperCase() || 'FILE'

  const getFileTypeIcon = (url: string) => {
    if (isImage(url)) return 'ri:image-line'
    const ext = getExt(url).toLowerCase()
    if (['pdf'].includes(ext)) return 'ri:file-pdf-2-line'
    if (['doc', 'docx'].includes(ext)) return 'ri:file-word-line'
    if (['xls', 'xlsx'].includes(ext)) return 'ri:file-excel-line'
    if (['zip', 'rar', '7z'].includes(ext)) return 'ri:file-zip-line'
    if (['mp4', 'avi', 'mov'].includes(ext)) return 'ri:video-line'
    if (['mp3', 'wav', 'flac'].includes(ext)) return 'ri:music-line'
    return 'ri:file-line'
  }

  const getMimeIcon = (mime: string) => {
    if (mime?.startsWith('video/')) return 'ri:video-line'
    if (mime?.startsWith('audio/')) return 'ri:music-line'
    if (mime?.includes('pdf')) return 'ri:file-pdf-2-line'
    if (mime?.includes('word')) return 'ri:file-word-line'
    if (mime?.includes('sheet') || mime?.includes('excel')) return 'ri:file-excel-line'
    if (mime?.includes('zip') || mime?.includes('rar')) return 'ri:file-zip-line'
    return 'ri:file-line'
  }

  const formatSize = (size: number): string => {
    if (!size) return '0B'
    const k = 1024
    const s = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(size) / Math.log(k))
    return (size / Math.pow(k, i)).toFixed(1) + s[i]
  }
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .art-file-selector {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: 8px;
  }

  /* 预览卡片 */
  .selector-preview {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .preview-card {
    position: relative;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid var(--el-border-color-lighter);
    cursor: pointer;
  }

  .preview-card:hover .preview-actions {
    opacity: 1;
  }

  .preview-img {
    width: 100%;
    height: 100%;
  }

  .preview-file {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: var(--el-fill-color-lighter);
    color: var(--el-text-color-secondary);
  }

  .preview-ext {
    font-size: 10px;
    margin-top: 2px;
  }

  .preview-actions {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.5);
    opacity: 0;
    transition: opacity 0.2s;
  }

  .action-icon {
    color: #fff;
    font-size: 18px;
    cursor: pointer;
  }

  /* 弹窗头部 */
  .dialog-header {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .dialog-title {
    font-size: 16px;
    font-weight: 600;
  }

  .hidden {
    display: none;
  }

  /* 弹窗布局 */
  .chooser-layout {
    display: flex;
    gap: 16px;
    height: 500px;
  }

  /* 左侧分类 */
  .chooser-sidebar {
    width: 120px;
    flex-shrink: 0;
    border-right: 1px solid var(--el-border-color-lighter);
    padding-right: 12px;
    overflow-y: auto;
  }

  .cat-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 13px;
    color: var(--el-text-color-regular);
    transition: all 0.15s;
    margin-bottom: 2px;
  }

  .cat-item:hover {
    background: var(--el-fill-color-lighter);
  }

  .cat-item.active {
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-weight: 500;
  }

  .cat-icon {
    font-size: 16px;
    flex-shrink: 0;
  }

  /* 右侧主体 */
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
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
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
    height: 300px;
    color: var(--el-text-color-placeholder);
  }

  .chooser-empty p {
    margin-top: 8px;
    font-size: 13px;
  }

  /* 文件卡片 */
  .file-card {
    border: 2px solid var(--el-border-color-lighter);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.15s;
  }

  .file-card:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }

  .file-card.selected {
    border-color: var(--el-color-primary);
    box-shadow: 0 0 0 1px var(--el-color-primary);
  }

  .file-card__preview {
    position: relative;
    height: 100px;
    background: var(--el-fill-color-lighter);
  }

  .file-card__img {
    width: 100%;
    height: 100%;
    display: block;
  }

  .file-card__icon {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-text-color-placeholder);
  }

  .file-card__check {
    position: absolute;
    top: 4px;
    right: 4px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--el-color-primary);
    display: flex;
    align-items: center;
    justify-content: center;
  }

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

  /* 分页 */
  .chooser-pagination {
    flex-shrink: 0;
    padding-top: 12px;
    display: flex;
    justify-content: flex-end;
  }

  /* 弹窗底部 */
  .dialog-footer {
    display: flex;
    align-items: center;
    gap: 12px;
    justify-content: flex-end;
  }

  .selected-count {
    font-size: 13px;
    color: var(--el-color-primary);
    margin-right: auto;
  }
</style>
