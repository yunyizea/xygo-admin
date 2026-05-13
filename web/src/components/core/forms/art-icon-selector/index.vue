<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 图标选择器组件 -->
<template>
  <div class="art-icon-selector">
    <ElInput
      :model-value="modelValue"
      placeholder="请选择图标"
      readonly
      @click="showDialog = true"
    >
      <template #prepend>
        <div class="icon-preview">
          <ArtSvgIcon v-if="modelValue" :icon="modelValue" :size="18" />
          <ArtSvgIcon v-else icon="ri:image-line" :size="18" />
        </div>
      </template>
      <template #append>
        <ElButton v-if="modelValue && clearable" @click.stop="handleClear">
          <ArtSvgIcon icon="ri:close-line" :size="16" />
        </ElButton>
        <div v-else class="search-icon-append">
          <ArtSvgIcon icon="ri:search-line" :size="16" />
        </div>
      </template>
    </ElInput>

    <!-- 图标选择弹窗 -->
    <ElDialog
      v-model="showDialog"
      title="选择图标"
      width="900px"
      :close-on-click-modal="false"
      class="icon-selector-dialog"
    >
      <div class="icon-selector-content">
        <!-- 搜索和筛选 -->
        <div class="selector-header">
          <div class="input-group">
            <ElInput
              v-model="searchText"
              placeholder="搜索图标名称..."
              clearable
              class="search-input"
            >
              <template #prefix>
                <ArtSvgIcon icon="ri:search-line" :size="16" />
              </template>
            </ElInput>
            <ElButton type="primary" plain @click="showCustomInput = !showCustomInput">
              <ArtSvgIcon icon="ri:edit-line" :size="16" style="margin-right: 4px;" />
              {{ showCustomInput ? '选择图标' : '自定义输入' }}
            </ElButton>
          </div>
          
          <ElRadioGroup v-if="!showCustomInput" v-model="currentCategory" class="category-tabs" size="default">
            <ElRadioButton value="local">Remix</ElRadioButton>
            <ElRadioButton value="mdi">Material</ElRadioButton>
            <ElRadioButton value="awe">Awesome</ElRadioButton>
            <ElRadioButton value="ali">Ant Design</ElRadioButton>
          </ElRadioGroup>

          <div v-if="!showCustomInput" class="sub-category-bar">
            <ElSelect v-if="subCategories.length > 1" v-model="currentSubCategory" size="small" placeholder="全部分类" style="width: 180px;">
              <ElOption label="全部分类" value="all" />
              <ElOption v-for="cat in subCategories" :key="cat" :label="cat" :value="cat" />
            </ElSelect>
            <span v-else />
            <span class="icon-stats">
              <span class="icon-count">共 {{ filteredIcons.length }} 个</span>
              <ElButton v-if="hasMore" type="primary" link size="small" @click="displayLimit = filteredIcons.length">
                显示全部（当前 {{ displayLimit }}）
              </ElButton>
            </span>
          </div>

          <!-- 自定义输入区域 -->
          <div v-if="showCustomInput" class="custom-input-area">
            <ElInput
              v-model="customIconName"
              placeholder="输入完整图标名称，如：ri:wechat-line"
              clearable
            >
              <template #prepend>
                <ArtSvgIcon v-if="customIconName" :icon="customIconName" :size="18" />
                <ArtSvgIcon v-else icon="ri:quill-pen-line" :size="18" />
              </template>
            </ElInput>
            <div class="custom-hint">
              <span>提示：访问 </span>
              <a href="https://icon-sets.iconify.design/" target="_blank" class="icon-link">
                Iconify 图标库
              </a>
              <span> 查找更多图标</span>
            </div>
          </div>
        </div>

        <!-- 图标列表 -->
        <div v-if="!showCustomInput" class="icon-list-wrapper">
          <div v-if="iconLoading" class="loading-state">
            <ArtSvgIcon icon="svg-spinners:3-dots-fade" :size="32" class="text-theme" />
            <span>正在加载图标库...</span>
          </div>
          <ElScrollbar v-else height="420px">
            <div v-if="displayIcons.length > 0" class="icon-list">
              <div
                v-for="icon in displayIcons"
                :key="icon"
                class="icon-item"
                :class="{ active: tempSelectedIcon === icon }"
                @click="handleSelectIcon(icon)"
              >
                <ArtSvgIcon :icon="icon" :size="24" />
                <div class="icon-name">{{ formatIconName(icon) }}</div>
              </div>
            </div>
            <div v-if="hasMore" class="load-more" @click="displayLimit += 500">
              点击加载更多（剩余 {{ filteredIcons.length - displayLimit }} 个）
            </div>
            <ElEmpty v-if="!iconLoading && displayIcons.length === 0" description="未找到匹配的图标" :image-size="120" />
          </ElScrollbar>
        </div>

        <!-- 自定义输入预览 -->
        <div v-else class="custom-preview-area">
          <div class="preview-title">预览效果</div>
          <div class="preview-box">
            <div class="preview-icon-large">
              <ArtSvgIcon v-if="customIconName" :icon="customIconName" :size="64" />
              <span v-else class="preview-placeholder">输入图标名称查看预览</span>
            </div>
            <div v-if="customIconName" class="preview-icon-name">{{ customIconName }}</div>
          </div>
          <div class="preview-sizes">
            <div class="size-item">
              <span class="size-label">小</span>
              <ArtSvgIcon v-if="customIconName" :icon="customIconName" :size="16" />
            </div>
            <div class="size-item">
              <span class="size-label">中</span>
              <ArtSvgIcon v-if="customIconName" :icon="customIconName" :size="24" />
            </div>
            <div class="size-item">
              <span class="size-label">大</span>
              <ArtSvgIcon v-if="customIconName" :icon="customIconName" :size="32" />
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <div class="selected-info">
            <template v-if="modelValue">
              已选择: <strong>{{ modelValue }}</strong>
            </template>
          </div>
          <div>
            <ElButton @click="showDialog = false">取消</ElButton>
            <ElButton type="primary" @click="handleConfirm">确定</ElButton>
          </div>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { addCollection } from '@iconify/vue'
import riData from '@iconify-json/ri/icons.json'
import fa6Data from '@iconify-json/fa6-solid/icons.json'
import antData from '@iconify-json/ant-design/icons.json'
import mdiData from '@iconify-json/mdi/icons.json'

defineOptions({ name: 'ArtIconSelector' })

interface Props {
  modelValue?: string
  clearable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  clearable: true
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'change': [value: string]
}>()

const showDialog = ref(false)
const searchText = ref('')
const currentCategory = ref('local')
const tempSelectedIcon = ref('')
const showCustomInput = ref(false)
const customIconName = ref('')
const currentSubCategory = ref('all')
const displayLimit = ref(500)

// 注册图标集到本地（离线可用）
addCollection(riData as any)
addCollection(fa6Data as any)
addCollection(antData as any)
addCollection(mdiData as any)

interface IconSetData {
  prefix: string
  categories: Record<string, string[]>
  allIcons: string[]
}

function parseIconSet(data: any): IconSetData {
  const prefix = data.prefix as string
  const names = Object.keys(data.icons)
  const result: IconSetData = { prefix, categories: {}, allIcons: [] }

  if (data.categories) {
    const categorized = new Set<string>()
    for (const [cat, items] of Object.entries(data.categories)) {
      const prefixed = (items as string[]).map(n => `${prefix}:${n}`)
      result.categories[cat] = prefixed
      ;(items as string[]).forEach(n => categorized.add(n))
    }
    result.allIcons = names.map(n => `${prefix}:${n}`)
  } else {
    result.allIcons = names.map(n => `${prefix}:${n}`)
  }
  return result
}

const iconSetsMap: Record<string, IconSetData> = {
  local: parseIconSet(riData),
  mdi: parseIconSet(mdiData),
  awe: parseIconSet(fa6Data),
  ali: parseIconSet(antData),
}

const subCategories = computed(() => {
  const data = iconSetsMap[currentCategory.value]
  return data ? Object.keys(data.categories) : []
})

const filteredIcons = computed(() => {
  const data = iconSetsMap[currentCategory.value]
  if (!data) return []
  let icons: string[]
  if (currentSubCategory.value !== 'all' && data.categories[currentSubCategory.value]) {
    icons = data.categories[currentSubCategory.value]
  } else {
    icons = data.allIcons
  }
  if (searchText.value) {
    const s = searchText.value.toLowerCase()
    icons = icons.filter(i => i.toLowerCase().includes(s))
  }
  return icons
})

const displayIcons = computed(() => filteredIcons.value.slice(0, displayLimit.value))
const hasMore = computed(() => filteredIcons.value.length > displayLimit.value)
const iconLoading = ref(false)

function formatIconName(icon: string): string {
  const idx = icon.indexOf(':')
  return idx >= 0 ? icon.substring(idx + 1) : icon
}

// 根据图标集大小决定初始显示量
function getDefaultLimit(key: string): number {
  const total = iconSetsMap[key]?.allIcons.length || 0
  return total <= 1000 ? total : 500
}

// 切换 Tab 时重置分类和分页
watch(currentCategory, (val) => {
  currentSubCategory.value = 'all'
  displayLimit.value = getDefaultLimit(val)
  searchText.value = ''
})

// 切换子分类时重置分页
watch(currentSubCategory, () => {
  displayLimit.value = getDefaultLimit(currentCategory.value)
})

// 搜索时显示全部匹配结果
watch(searchText, () => {
  displayLimit.value = 9999
})

// 监听弹窗打开，同步当前值
watch(showDialog, (val) => {
  if (val) {
    tempSelectedIcon.value = props.modelValue
    customIconName.value = props.modelValue
    showCustomInput.value = false
    displayLimit.value = getDefaultLimit(currentCategory.value)
  }
})

// 监听自定义输入，同步到临时选中值
watch(customIconName, (val) => {
  if (showCustomInput.value && val) {
    tempSelectedIcon.value = val
  }
})

// 选择图标
const handleSelectIcon = (icon: string) => {
  tempSelectedIcon.value = icon
  customIconName.value = icon
}

// 确认选择
const handleConfirm = () => {
  const iconValue = showCustomInput.value ? customIconName.value : tempSelectedIcon.value
  emit('update:modelValue', iconValue)
  emit('change', iconValue)
  showDialog.value = false
}

// 清空选择
const handleClear = () => {
  emit('update:modelValue', '')
  emit('change', '')
}
</script>

<style scoped lang="scss">
.art-icon-selector {
  width: 100%;
  
  :deep(.el-input) {
    cursor: pointer;
  }
  
  :deep(.el-input__inner) {
    cursor: pointer;
  }
}

.icon-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  color: var(--art-gray-600);
}

.search-icon-append {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
  color: var(--art-gray-400);
  cursor: pointer;
}

.icon-selector-dialog {
  :deep(.el-dialog__body) {
    padding: 0;
  }
}

.icon-selector-content {
  display: flex;
  flex-direction: column;
  max-height: 600px;
  overflow: hidden;
}

.selector-header {
  padding: 20px 20px 16px;
  border-bottom: 1px solid var(--art-gray-200);
  background: var(--art-gray-50);
  
  .input-group {
    display: flex;
    gap: 12px;
    margin-bottom: 16px;
    
    .search-input {
      flex: 1;
    }
  }
  
  .category-tabs {
    display: flex;
    justify-content: center;
    
    :deep(.el-radio-button__inner) {
      padding: 8px 16px;
    }
  }

  .sub-category-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 12px;

    .icon-stats {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .icon-count {
      font-size: 12px;
      color: var(--art-gray-500);
    }
  }

  .custom-input-area {
    margin-top: 16px;
    
    .custom-hint {
      margin-top: 8px;
      font-size: 12px;
      color: var(--art-gray-500);
      text-align: center;
      
      .icon-link {
        color: var(--theme-color);
        text-decoration: none;
        
        &:hover {
          text-decoration: underline;
        }
      }
    }
  }
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  gap: 16px;
  color: var(--art-gray-500);
  font-size: 14px;
}

.load-more {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  margin-top: 12px;
  font-size: 13px;
  color: var(--theme-color);
  cursor: pointer;
  border: 1px dashed var(--art-gray-300);
  border-radius: 8px;
  transition: all 0.2s;

  &:hover {
    background: var(--theme-color-alpha-5);
    border-color: var(--theme-color);
  }
}

.icon-list-wrapper {
  flex: 1;
  padding: 20px;
  overflow: hidden;
}

.icon-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 12px;
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px 8px;
  border: 1px solid var(--art-gray-200);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: var(--default-box-color);
  
  &:hover {
    border-color: var(--theme-color);
    background: var(--theme-color-alpha-5);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.08);
  }
  
  &.active {
    border-color: var(--theme-color);
    background: var(--theme-color-alpha-10);
    
    .icon-name {
      color: var(--theme-color);
      font-weight: 600;
    }
  }
  
  .icon-name {
    margin-top: 8px;
    font-size: 11px;
    color: var(--art-gray-600);
    text-align: center;
    line-height: 1.2;
    word-break: break-all;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }
}

.custom-preview-area {
  padding: 40px 20px;
  text-align: center;
  
  .preview-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--art-gray-700);
    margin-bottom: 24px;
  }
  
  .preview-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    background: var(--art-gray-50);
    border: 2px dashed var(--art-gray-300);
    border-radius: 12px;
    margin-bottom: 32px;
    min-height: 200px;
  }
  
  .preview-icon-large {
    color: var(--theme-color);
    margin-bottom: 16px;
  }
  
  .preview-placeholder {
    font-size: 14px;
    color: var(--art-gray-400);
  }
  
  .preview-icon-name {
    font-size: 13px;
    color: var(--art-gray-600);
    font-family: monospace;
    background: var(--art-gray-100);
    padding: 4px 12px;
    border-radius: 4px;
  }
  
  .preview-sizes {
    display: flex;
    justify-content: center;
    gap: 48px;
    
    .size-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 12px;
      
      .size-label {
        font-size: 12px;
        color: var(--art-gray-500);
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  
  .selected-info {
    font-size: 14px;
    color: var(--art-gray-600);
    
    strong {
      color: var(--theme-color);
      margin-left: 4px;
    }
  }
}

// 暗黑模式适配
.dark {
  .icon-preview {
    color: var(--art-gray-400);
  }

  .search-icon-append {
    color: var(--art-gray-500);
  }
  
  .selector-header {
    background: var(--art-gray-900);
    border-bottom-color: var(--art-gray-700);
  }
  
  .icon-item {
    background: var(--art-gray-800);
    border-color: var(--art-gray-700);
    
    &:hover {
      background: var(--theme-color-alpha-10);
    }
    
    &.active {
      background: var(--theme-color-alpha-20);
    }
  }

  .custom-preview-area {
    .preview-box {
      background: var(--art-gray-800);
      border-color: var(--art-gray-700);
    }
    
    .preview-icon-name {
      background: var(--art-gray-700);
    }
  }
}
</style>
