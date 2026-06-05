<!-- {{.TableComment}} 编辑弹窗(树表) -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="type === 'add' ? '新增{{.TableComment}}' : '编辑{{.TableComment}}'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="100px">
      <!-- 父级选择 -->
      <ElFormItem label="上级" prop="{{.TreePidTsColumn}}">
        <ElTreeSelect
          v-model="formData.{{.TreePidTsColumn}}"
          :data="parentTreeData"
          :props="{ label: '{{.TreeTitleTsColumn}}', value: '{{.PkTsName}}', children: 'children' }"
          placeholder="选择上级（留空为顶级）"
          clearable
          check-strictly
          :render-after-expand="false"
          default-expand-all
          style="width: 100%"
        />
      </ElFormItem>
{{- range .EditColumns}}
{{- if and (ne .Name $.TreePidColumn) (ne .Name $.PkColumn)}}
{{- if eq .DesignType "switch"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElSwitch v-model="formData.{{.TsName}}" :active-value="1" :inactive-value="0" />
      </ElFormItem>
{{- else if eq .DesignType "radio"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElRadioGroup v-model="formData.{{.TsName}}">
{{- if .RadioOptions}}
{{- range .RadioOptions}}
          <ElRadio :value="{{jsValue .Value}}">{{.Label}}</ElRadio>
{{- end}}
{{- else}}
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
{{- end}}
        </ElRadioGroup>
      </ElFormItem>
{{- else if or (eq .DesignType "number") (eq .DesignType "float") (eq .DesignType "weigh")}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInputNumber v-model="formData.{{.TsName}}" :min="0" controls-position="right" />
      </ElFormItem>
{{- else if eq .DesignType "textarea"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInput v-model="formData.{{.TsName}}" type="textarea" :rows="3" placeholder="请输入{{.Label}}" />
      </ElFormItem>
{{- else if or (eq .DesignType "select") (eq .DesignType "selects")}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElSelect v-model="formData.{{.TsName}}" placeholder="请选择{{.Label}}" clearable{{if eq .DesignType "selects"}} multiple{{end}}>
{{- if .HasOptions}}
{{- range .Options}}
          <ElOption :value="{{jsValue .Value}}" label="{{.Label}}" />
{{- end}}
{{- end}}
        </ElSelect>
      </ElFormItem>
{{- else if eq .DesignType "icon"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ArtIconSelector v-model="formData.{{.TsName}}" />
      </ElFormItem>
{{- else if eq .DesignType "color"}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElColorPicker v-model="formData.{{.TsName}}" />
      </ElFormItem>
{{- else if not .IsTimeField}}
      <ElFormItem label="{{.Label}}" prop="{{.TsName}}">
        <ElInput v-model="formData.{{.TsName}}" placeholder="请输入{{.Label}}" />
      </ElFormItem>
{{- end}}
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
  import ArtFileSelector from '@/components/core/forms/art-file-selector/index.vue'
  import ArtIconSelector from '@/components/core/forms/art-icon-selector/index.vue'

  const props = defineProps<{
    visible: boolean
    type: 'add' | 'edit'
    editData?: Record<string, any>
    treeData: any[]
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

  // 构建父级树数据（加一个顶级选项）
  const parentTreeData = computed(() => [
    { {{.PkTsName}}: 0, {{.TreeTitleTsColumn}}: '顶级', children: props.treeData || [] }
  ])

  const defaultForm = (): Record<string, any> => ({
    {{.TreePidTsColumn}}: 0,
{{- range .EditColumns}}
{{- if ne .Name $.TreePidColumn}}
{{- if eq .DesignType "remoteSelect"}}
    {{.TsName}}: undefined,
{{- else if eq .DesignType "remoteSelects"}}
    {{.TsName}}: [],
{{- else}}
    {{.TsName}}: {{.DefaultValue}},
{{- end}}
{{- end}}
{{- end}}
  })

  const formData = reactive(defaultForm())

  const rules = reactive<FormRules>({
{{- range .EditColumns}}
{{- if and .Required (ne .Name $.TreePidColumn)}}
    {{.TsName}}: [{ required: true, message: '{{.Label}}不能为空', trigger: 'blur' }],
{{- end}}
{{- end}}
  })

  watch(() => props.visible, (val) => {
    if (val && props.type === 'edit' && props.editData) {
      Object.assign(formData, props.editData)
    } else if (val && props.type === 'add' && props.editData) {
      // 新增子项时继承父ID
      const newData = defaultForm()
      if (props.editData.{{.TreePidTsColumn}} !== undefined) {
        newData.{{.TreePidTsColumn}} = props.editData.{{.TreePidTsColumn}}
      }
      Object.assign(formData, newData)
    } else if (val) {
      Object.assign(formData, defaultForm())
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
