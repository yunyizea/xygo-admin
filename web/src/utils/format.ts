// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

/**
 * 格式化文件大小
 * @param bytes 字节数
 * @returns 格式化后的文件大小
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

/**
 * 将多图/多文件字段值标准化为 URL 字符串数组。
 * 兼容多种历史存储格式：
 * - 数组：['/a.png', '/b.png']
 * - 逗号分隔字符串：'/a.png,/b.png'
 * - JSON 数组字符串：'["/a.png","/b.png"]'
 * - 含多余引号的字符串：'"/a.png"'
 * @param val 字段值
 * @returns 干净的 URL 数组
 */
export function toUrlList(val: unknown): string[] {
  if (Array.isArray(val)) return val.map((v) => String(v)).filter(Boolean)
  const raw = String(val ?? '').trim()
  if (!raw) return []
  if (raw.startsWith('[')) {
    try {
      const arr = JSON.parse(raw)
      if (Array.isArray(arr)) return arr.map((v) => String(v)).filter(Boolean)
    } catch {
      /* 不是合法 JSON，走逗号分隔回退 */
    }
  }
  return raw
    .replace(/^\[|\]$/g, '')
    .split(',')
    .map((s) => s.trim().replace(/^["']+|["']+$/g, ''))
    .filter(Boolean)
}

/**
 * 格式化时间戳
 * @param timestamp Unix时间戳（秒）
 * @returns 格式化后的时间字符串
 */
export function formatDate(timestamp: number): string {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

/**
 * 格式化日期（不含时间）
 * @param timestamp Unix时间戳（秒）
 * @returns 格式化后的日期字符串
 */
export function formatDateOnly(timestamp: number): string {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
