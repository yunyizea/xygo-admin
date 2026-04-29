<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<template>
  <ElDialog
    v-model="visible"
    :title="dialogTitle"
    width="30%"
    align-center
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="120px" class="mt-4">
      <ElFormItem label="角色名称" prop="name" class="mb-6">
        <ElInput v-model="form.name" placeholder="请输入角色名称" />
      </ElFormItem>
      <ElFormItem label="角色标识" prop="key" class="mb-6">
        <ElInput v-model="form.key" placeholder="请输入角色标识（如：admin, manager）" />
      </ElFormItem>
      <ElFormItem label="描述" prop="remark" class="mb-6">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="3"
          placeholder="请输入角色描述"
        />
      </ElFormItem>
      <ElFormItem label="排序" prop="sort" class="mb-6">
        <ElInputNumber v-model="form.sort" :min="0" controls-position="right" style="width: 100%" />
      </ElFormItem>
      <ElFormItem label="启用" class="mb-4">
        <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchSaveRole } from '@/api/backend/system/role'

  type RoleListItem = Api.SystemManage.RoleListItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    roleData?: RoleListItem
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    roleData: undefined
  })

  const emit = defineEmits<Emits>()

  const formRef = ref<FormInstance>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const dialogTitle = computed(() => {
    if (props.dialogType === 'edit') return '编辑角色'
    if (props.roleData?.id) return '新增子角色'
    return '新增角色'
  })

  const rules = reactive<FormRules>({
    name: [
      { required: true, message: '请输入角色名称', trigger: 'blur' },
      { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
    ],
    key: [
      { required: true, message: '请输入角色标识', trigger: 'blur' },
      { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' },
      { pattern: /^[a-z_]+$/, message: '只能包含小写字母和下划线', trigger: 'blur' }
    ]
  })

  const form = reactive({
    id: 0,
    name: '',
    key: '',
    pid: 0,
    sort: 0,
    status: 1,
    remark: ''
  })

  watch(
    () => props.modelValue,
    (newVal) => {
      if (newVal) initForm()
    },
    { flush: 'post' }
  )

  watch(
    () => props.roleData,
    (newData) => {
      if (newData && props.modelValue) initForm()
    },
    { deep: true, flush: 'post' }
  )

  const initForm = () => {
    if (props.dialogType === 'edit' && props.roleData) {
      form.id = props.roleData.id || 0
      form.name = props.roleData.name || ''
      form.key = props.roleData.key || ''
      form.pid = props.roleData.pid || 0
      form.sort = props.roleData.sort || 0
      form.status = props.roleData.status || 1
      form.remark = props.roleData.remark || ''
    } else {
      form.id = 0
      form.name = ''
      form.key = ''
      form.pid =
        props.dialogType === 'add' && props.roleData?.id ? props.roleData.id : 0
      form.sort = 0
      form.status = 1
      form.remark = ''
    }
  }

  const handleClose = () => {
    visible.value = false
    formRef.value?.resetFields()
  }

  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
    } catch {
      return
    }

    try {
      await fetchSaveRole({ ...form })
      const message =
        props.dialogType === 'edit'
          ? '修改成功'
          : form.pid
            ? '新增子角色成功'
            : '新增成功'
      ElMessage.success(message)
      emit('success')
      handleClose()
    } catch (error) {
      console.error('保存角色失败:', error)
    }
  }
</script>
