/**
 * 字段权限管理 API
 */
import { adminRequest } from '@/utils/http'

// 查询字段权限列表
export function getFieldPermListApi(params: {
  roleId: number
  module?: string
  resource?: string
}) {
  return adminRequest.get<any>({
    url: '/fieldPerm/list',
    params
  })
}

// 批量保存字段权限
export function batchSaveFieldPermApi(data: {
  roleId: number
  resource: string
  fields: Array<{
    fieldName: string
    fieldLabel: string
    permType: number
  }>
}) {
  return adminRequest.post<any>({
    url: '/fieldPerm/batchSave',
    data
  })
}

// 获取角色字段权限映射
export function getFieldPermByRoleApi(params: {
  roleId: number
  resource?: string
}) {
  return adminRequest.get<any>({
    url: '/fieldPerm/getByRole',
    params
  })
}

// 获取资源字段列表
export function getResourceFieldsApi(params: {
  resource: string
}) {
  return adminRequest.get<any>({
    url: '/fieldPerm/resourceFields',
    params
  })
}

// 获取当前用户的字段权限（合并所有角色）
export function getFieldPermMineApi() {
  return adminRequest.get<{ fieldPerms: Record<string, Record<string, number>> }>({
    url: '/fieldPerm/mine'
  })
}
