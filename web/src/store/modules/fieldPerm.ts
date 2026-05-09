/**
 * 字段权限状态管理
 *
 * 存储当前登录用户的字段权限配置（合并所有角色，取最高权限）。
 * 在用户登录成功 / 路由初始化时调用 loadFieldPerms() 拉取一次。
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getFieldPermMineApi } from '@/api/backend/system/fieldPerm'

export const useFieldPermStore = defineStore('fieldPermStore', () => {
  // resource -> field -> permType (0=不可见, 1=只读, 2=可编辑)
  const fieldPerms = ref<Record<string, Record<string, number>>>({})
  const loaded = ref(false)

  const loadFieldPerms = async () => {
    try {
      const res = await getFieldPermMineApi()
      fieldPerms.value = res?.fieldPerms || {}
      loaded.value = true
    } catch {
      fieldPerms.value = {}
      loaded.value = true
    }
  }

  const reset = () => {
    fieldPerms.value = {}
    loaded.value = false
  }

  return {
    fieldPerms,
    loaded,
    loadFieldPerms,
    reset
  }
})
