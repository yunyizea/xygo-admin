<!-- {{.TableComment}} 详情页（全屏） -->
<template>
  <div class="{{.CssClass}}-detail art-full-height">
    <ElCard shadow="never" class="art-table-card">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-bold text-lg">{{.TableComment}}详情</span>
          <ElButton @click="goBack">
            <ArtSvgIcon icon="ri:arrow-left-line" class="text-sm mr-1" />
            返回列表
          </ElButton>
        </div>
      </template>

      <div v-if="loading" class="flex justify-center py-20">
        <ElIcon class="is-loading" :size="32"><Loading /></ElIcon>
      </div>

      <ElDescriptions v-else-if="detail" :column="2" border class="detail-descriptions">
{{- range .ListColumns}}
{{- if or (eq .DesignType "image") (eq .Render "image")}}
        <ElDescriptionsItem label="{{.Label}}">
          <ElImage v-if="detail.{{.TsName}}" :src="detail.{{.TsName}}" style="width:120px;height:120px" fit="cover" :preview-src-list="[detail.{{.TsName}}]" />
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if or (eq .DesignType "images") (eq .Render "images")}}
        <ElDescriptionsItem label="{{.Label}}" :span="2">
          <div v-if="detail.{{.TsName}}" style="display:flex;gap:8px;flex-wrap:wrap">
            <ElImage v-for="(img, i) in (Array.isArray(detail.{{.TsName}}) ? detail.{{.TsName}} : String(detail.{{.TsName}} || '').split(',').filter(Boolean))" :key="i" :src="img" style="width:80px;height:80px" fit="cover" :preview-src-list="Array.isArray(detail.{{.TsName}}) ? detail.{{.TsName}} : String(detail.{{.TsName}} || '').split(',').filter(Boolean)" />
          </div>
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if or (eq .DesignType "switch") (eq .Render "switch")}}
        <ElDescriptionsItem label="{{.Label}}">
          <ElTag :type="detail.{{.TsName}} === 1 ? 'success' : 'danger'" size="small">
            {{"{{"}} detail.{{.TsName}} === 1 ? '启用' : '禁用' {{"}}"}}
          </ElTag>
        </ElDescriptionsItem>
{{- else if or (eq .Render "tag") (and (eq .Render "") (or (eq .DesignType "radio") (eq .DesignType "select")))}}
        <ElDescriptionsItem label="{{.Label}}">
{{- if .RadioOptions}}
          <ElTag size="small">{{"{{"}} ({ {{range .RadioOptions}}'{{.Value}}': '{{.Label}}', {{end}} })[String(detail.{{.TsName}})] || detail.{{.TsName}} {{"}}"}}</ElTag>
{{- else}}
          <ElTag size="small">{{"{{"}} detail.{{.TsName}} ?? '-' {{"}}"}}</ElTag>
{{- end}}
        </ElDescriptionsItem>
{{- else if or (eq .DesignType "color") (eq .Render "color")}}
        <ElDescriptionsItem label="{{.Label}}">
          <span v-if="detail.{{.TsName}}" :style="`display:inline-block;width:24px;height:24px;border-radius:4px;background:${detail.{{.TsName}}}`"></span>
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if or (eq .Render "url") (and (eq .Render "") (eq .DesignType "file"))}}
        <ElDescriptionsItem label="{{.Label}}">
          <a v-if="detail.{{.TsName}}" :href="detail.{{.TsName}}" target="_blank" style="color:var(--el-color-primary)">{{"{{"}} detail.{{.TsName}} {{"}}"}}</a>
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if eq .DesignType "editor"}}
        <ElDescriptionsItem label="{{.Label}}" :span="2">
          <div v-if="detail.{{.TsName}}" v-html="detail.{{.TsName}}" class="max-h-[400px] overflow-auto"></div>
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if or (eq .DesignType "timestamp") (eq .DesignType "datetime") (eq .Render "datetime")}}
        <ElDescriptionsItem label="{{.Label}}">{{"{{"}} formatTimestamp(detail.{{.TsName}}) {{"}}"}}</ElDescriptionsItem>
{{- else if or (eq .Render "icon") (and (eq .Render "") (eq .DesignType "icon"))}}
        <ElDescriptionsItem label="{{.Label}}">
          <ArtSvgIcon v-if="detail.{{.TsName}}" :icon="detail.{{.TsName}}" class="text-lg" />
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if or (eq .Render "tags") (and (eq .Render "") (or (eq .DesignType "checkbox") (eq .DesignType "selects")))}}
        <ElDescriptionsItem label="{{.Label}}">
          <div v-if="detail.{{.TsName}}" style="display:flex;gap:4px;flex-wrap:wrap">
            <ElTag v-for="(v, i) in (Array.isArray(detail.{{.TsName}}) ? detail.{{.TsName}} : String(detail.{{.TsName}} || '').split(',').filter(Boolean))" :key="i" size="small">{{"{{"}} v {{"}}"}}</ElTag>
          </div>
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else if and (eq .Render "") (eq .DesignType "files")}}
        <ElDescriptionsItem label="{{.Label}}">
          <div v-if="detail.{{.TsName}}" style="display:flex;flex-direction:column;gap:4px">
            <a v-for="(f, i) in (Array.isArray(detail.{{.TsName}}) ? detail.{{.TsName}} : String(detail.{{.TsName}} || '').split(',').filter(Boolean))" :key="i" :href="f" target="_blank" style="color:var(--el-color-primary);font-size:13px">文件 {{"{{"}} i + 1 {{"}}"}}</a>
          </div>
          <span v-else>-</span>
        </ElDescriptionsItem>
{{- else}}
        <ElDescriptionsItem label="{{.Label}}">{{"{{"}} detail.{{.TsName}} ?? '-' {{"}}"}}</ElDescriptionsItem>
{{- end}}
{{- end}}
      </ElDescriptions>

      <div v-else class="py-20 text-center text-gray-400">
        数据加载失败
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Loading } from '@element-plus/icons-vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { fetch{{.VarName}}View } from '{{.WebApiImportPath}}'
  import { formatTimestamp } from '@/utils/time'
  import { useRoute, useRouter } from 'vue-router'

  defineOptions({ name: '{{.VarName}}Detail' })

  const route = useRoute()
  const router = useRouter()
  const loading = ref(false)
  const detail = ref<Record<string, any> | null>(null)

  const goBack = () => {
    router.back()
  }

  onMounted(async () => {
    const id = Number(route.query.id || route.params.id)
    if (!id) return
    loading.value = true
    try {
      detail.value = await fetch{{.VarName}}View(id) as any
    } catch { detail.value = null }
    loading.value = false
  })
</script>

<style scoped>
  .detail-descriptions {
    :deep(.el-descriptions__label) {
      width: 140px;
      font-weight: 600;
    }
  }
</style>
