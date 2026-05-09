/**
 * useFieldPerm - 字段权限控制
 *
 * 根据当前用户的角色配置，控制页面表格列和表单字段的可见性与可编辑性。
 * permType: 0=不可见, 1=只读, 2=可编辑（默认）
 *
 * @example
 * ```ts
 * const { isFieldVisible, isFieldReadonly, filterColumns } = useFieldPerm('admin_user')
 *
 * // 过滤表格列
 * const visibleColumns = filterColumns(allColumns)
 * ```
 */

import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFieldPermStore } from '@/store/modules/fieldPerm'
import { useUserStore } from '@/store/modules/user'

export const useFieldPerm = (resource?: string) => {
  const route = useRoute()
  const fieldPermStore = useFieldPermStore()
  const userStore = useUserStore()
  const { fieldPerms } = storeToRefs(fieldPermStore)

  const currentResource = computed(() => {
    return resource || (route.meta as any)?.resource || ''
  })

  const getPermType = (fieldName: string): number => {
    if (userStore.info?.isSuper) return 2
    const res = currentResource.value
    if (!res) return 2
    const resPerms = fieldPerms.value[res]
    if (!resPerms) return 2
    return resPerms[fieldName] ?? 2
  }

  const isFieldVisible = (fieldName: string): boolean => {
    return getPermType(fieldName) > 0
  }

  const isFieldReadonly = (fieldName: string): boolean => {
    return getPermType(fieldName) === 1
  }

  const filterColumns = <T extends { prop?: string; field?: string }>(columns: T[]): T[] => {
    return columns.filter((col) => {
      const prop = col.prop || col.field
      if (!prop) return true
      return isFieldVisible(prop)
    })
  }

  return {
    currentResource,
    getPermType,
    isFieldVisible,
    isFieldReadonly,
    filterColumns
  }
}
