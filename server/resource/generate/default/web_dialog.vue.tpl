<!-- {{.TableComment}} 编辑弹窗 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="type === 'add' ? '新增{{.TableComment}}' : '编辑{{.TableComment}}'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="100px">
{{- range .EditColumns}}
{{- if eq .DesignType "pk"}}
      <!-- 主键隐藏，不在表单中显示 -->
{{- else if eq .DesignType "switch"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElSwitch v-model="formData.{{.TsName}}" :active-value="1" :inactive-value="0" />
      </ElFormItem>
{{- else if eq .DesignType "radio"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
{{- if .HasOptions}}
        <ElRadioGroup v-model="formData.{{.TsName}}">
{{- range .Options}}
          <ElRadio :value="{{jsValue .Value}}">{{.Label}}</ElRadio>
{{- end}}
        </ElRadioGroup>
{{- else}}
{{- if eq .TsType "number"}}
        <ElInputNumber v-model="formData.{{.TsName}}" controls-position="right" />
{{- else}}
        <ElInput v-model="formData.{{.TsName}}" placeholder="请输入{{.Label}}" />
{{- end}}
{{- end}}
      </ElFormItem>
{{- else if eq .DesignType "checkbox"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
{{- if .HasOptions}}
        <ElCheckboxGroup v-model="formData.{{.TsName}}">
{{- range .Options}}
          <ElCheckbox :value="{{jsValue .Value}}">{{.Label}}</ElCheckbox>
{{- end}}
        </ElCheckboxGroup>
{{- else}}
        <ElInput v-model="formData.{{.TsName}}" placeholder="请输入{{.Label}}" />
{{- end}}
      </ElFormItem>
{{- else if eq .DesignType "remoteSelect"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElSelect
          v-model="formData.{{.TsName}}"
          filterable
          remote
          :remote-method="(q: string) => load{{.RelationName}}Options(q)"
          placeholder="请选择{{.Label}}"
          clearable
          :loading="{{.RelationAlias}}Loading"
        >
          <ElOption
            v-for="opt in {{.RelationAlias}}Options"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </ElSelect>
      </ElFormItem>
{{- else if eq .DesignType "remoteSelects"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElSelect
          v-model="formData.{{.TsName}}"
          filterable
          remote
          multiple
          :remote-method="(q: string) => load{{.RelationName}}Options(q)"
          placeholder="请选择{{.Label}}"
          clearable
          :loading="{{.RelationAlias}}Loading"
        >
          <ElOption
            v-for="opt in {{.RelationAlias}}Options"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </ElSelect>
      </ElFormItem>
{{- else if or (eq .DesignType "select") (eq .DesignType "selects")}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
{{- if .HasOptions}}
        <ElSelect v-model="formData.{{.TsName}}" placeholder="请选择{{.Label}}" clearable{{if eq .DesignType "selects"}} multiple{{end}}>
{{- range .Options}}
          <ElOption :value="{{jsValue .Value}}" label="{{.Label}}" />
{{- end}}
        </ElSelect>
{{- else}}
{{- if eq .TsType "number"}}
        <ElInputNumber v-model="formData.{{.TsName}}" controls-position="right" />
{{- else}}
        <ElInput v-model="formData.{{.TsName}}" placeholder="请输入{{.Label}}" />
{{- end}}
{{- end}}
      </ElFormItem>
{{- else if eq .DesignType "textarea"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInput v-model="formData.{{.TsName}}" type="textarea" :rows="3" placeholder="请输入{{.Label}}" />
      </ElFormItem>
{{- else if eq .DesignType "password"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInput v-model="formData.{{.TsName}}" type="password" show-password placeholder="请输入{{.Label}}" />
      </ElFormItem>
{{- else if or (eq .DesignType "number") (eq .DesignType "float") (eq .DesignType "weigh")}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInputNumber v-model="formData.{{.TsName}}" controls-position="right" />
      </ElFormItem>
{{- else if or (eq .DesignType "datetime") (eq .DesignType "timestamp")}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElDatePicker v-model="formData.{{.TsName}}" type="datetime" value-format="X" placeholder="请选择{{.Label}}" />
      </ElFormItem>
{{- else if eq .DesignType "date"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElDatePicker v-model="formData.{{.TsName}}" type="date" value-format="X" placeholder="请选择{{.Label}}" />
      </ElFormItem>
{{- else if eq .DesignType "time"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElTimePicker v-model="formData.{{.TsName}}" placeholder="请选择{{.Label}}" />
      </ElFormItem>
{{- else if eq .DesignType "image"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtFileSelector v-model="formData.{{.TsName}}" file-type="image" />
      </ElFormItem>
{{- else if eq .DesignType "images"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtFileSelector v-model="formData.{{.TsName}}" file-type="image" :max-number="9" />
      </ElFormItem>
{{- else if eq .DesignType "file"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtFileSelector v-model="formData.{{.TsName}}" />
      </ElFormItem>
{{- else if eq .DesignType "files"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtFileSelector v-model="formData.{{.TsName}}" :max-number="5" />
      </ElFormItem>
{{- else if eq .DesignType "editor"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtWangEditor v-model="formData.{{.TsName}}" placeholder="请输入{{.Label}}" />
      </ElFormItem>
{{- else if eq .DesignType "color"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElColorPicker v-model="formData.{{.TsName}}" />
      </ElFormItem>
{{- else if eq .DesignType "icon"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtIconSelector v-model="formData.{{.TsName}}" />
      </ElFormItem>
{{- else}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInput v-model="formData.{{.TsName}}" placeholder="请输入{{.Label}}" />
      </ElFormItem>
{{- end}}
{{- end}}
    </ElForm>

    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" :loading="loading" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import type { DialogType } from '@/types'
  import { adminRequest } from '@/utils/http'
  import ArtFileSelector from '@/components/core/forms/art-file-selector/index.vue'
  import ArtIconSelector from '@/components/core/forms/art-icon-selector/index.vue'
  import ArtWangEditor from '@/components/core/forms/art-wang-editor/index.vue'

  const props = defineProps<{
    visible: boolean
    type: DialogType
    editData?: Record<string, any>
  }>()

  const emit = defineEmits<{
    (e: 'update:visible', v: boolean): void
    (e: 'submit', data: Record<string, any>): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (val: boolean) => emit('update:visible', val)
  })

  const formRef = ref<FormInstance>()
  const loading = ref(false)

  const defaultForm = (): Record<string, any> => ({
{{- range .EditColumns}}
{{- if eq .DesignType "remoteSelect"}}
    {{.TsName}}: undefined,
{{- else if eq .DesignType "remoteSelects"}}
    {{.TsName}}: [],
{{- else}}
    {{.TsName}}: {{.DefaultValue}},
{{- end}}
{{- end}}
  })

  const formData = reactive(defaultForm())

  const rules = reactive<FormRules>({
{{- range .EditColumns}}
{{- if .Required}}
    {{.TsName}}: [{ required: true, message: '{{.Label}}不能为空', trigger: 'blur' }],
{{- end}}
{{- end}}
  })

{{- if .HasRelations}}

  // ==================== 远程下拉选项 ====================
{{- range $rel := .Relations}}
  const {{$rel.RelationAlias}}Options = ref<{ value: any; label: string }[]>([])
  const {{$rel.RelationAlias}}Loading = ref(false)
  const load{{$rel.RelationName}}Options = async (query: string) => {
    {{$rel.RelationAlias}}Loading.value = true
    try {
      const res = await adminRequest.get<any>({
        url: '/{{$rel.RelationApiPath}}/list',
        params: { pageSize: 50, {{$rel.RemoteField}}: query || undefined }
      })
      {{$rel.RelationAlias}}Options.value = (res.list || []).map((item: any) => ({
        value: item.{{$rel.RemotePk}},
        label: item.{{$rel.RemoteField}},
      }))
    } catch { /* ignore */ }
    {{$rel.RelationAlias}}Loading.value = false
  }
{{- end}}
{{- end}}

  watch(() => props.visible, (val) => {
    if (val && props.type === 'edit' && props.editData) {
      Object.assign(formData, props.editData)
{{- if .HasRelations}}
      // 编辑时加载已选关联项
{{- range $rel := .Relations}}
      load{{$rel.RelationName}}Options('')
{{- end}}
{{- end}}
    } else if (val) {
      Object.assign(formData, defaultForm())
{{- if .HasRelations}}
{{- range $rel := .Relations}}
      load{{$rel.RelationName}}Options('')
{{- end}}
{{- end}}
    }
  })

  const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate()
    emit('submit', { ...formData })
  }

  const handleClose = () => {
    formRef.value?.resetFields()
    Object.assign(formData, defaultForm())
    dialogVisible.value = false
  }
</script>
