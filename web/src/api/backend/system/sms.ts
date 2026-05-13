/**
 * 短信管理 API（模板、变量、日志）
 */
import { adminRequest } from '@/utils/http'

// ===================== 短信模板 =====================

export interface SmsTemplateItem {
  id: number
  title: string
  code: string
  content: string
  providerTemplateId: string
  variables: any
  relatedVariableId: number
  status: number
  sort: number
  remark: string
  createTime: number
  updateTime: number
}

export function fetchSmsTemplateList(params: any) {
  return adminRequest.get<{ list: SmsTemplateItem[]; total: number }>({
    url: '/sms/template/list',
    params
  })
}

export function fetchSmsTemplateSave(params: any) {
  return adminRequest.post({ url: '/sms/template/save', params })
}

export function fetchSmsTemplateDelete(id: number) {
  return adminRequest.post({ url: '/sms/template/delete', params: { id } })
}

export function fetchSmsTemplateTest(params: { id: number; phone: string }) {
  return adminRequest.post<{ success: boolean; requestId: string; message: string }>({
    url: '/sms/template/test',
    params
  })
}

// ===================== 短信变量 =====================

export interface SmsVariableItem {
  id: number
  title: string
  name: string
  sourceType: number
  sqlQuery: string
  methodName: string
  sharedCount: number
  status: number
  createTime: number
  updateTime: number
}

export function fetchSmsVariableList(params: any) {
  return adminRequest.get<{ list: SmsVariableItem[]; total: number }>({
    url: '/sms/variable/list',
    params
  })
}

export function fetchSmsVariableSave(params: any) {
  return adminRequest.post({ url: '/sms/variable/save', params })
}

export function fetchSmsVariableDelete(id: number) {
  return adminRequest.post({ url: '/sms/variable/delete', params: { id } })
}

// ===================== 短信日志 =====================

export interface SmsLogItem {
  id: number
  phone: string
  templateCode: string
  driver: string
  content: string
  params: any
  status: number
  requestId: string
  errorMsg: string
  createTime: number
}

export function fetchSmsLogList(params: any) {
  return adminRequest.get<{ list: SmsLogItem[]; total: number }>({
    url: '/sms/log/list',
    params
  })
}
