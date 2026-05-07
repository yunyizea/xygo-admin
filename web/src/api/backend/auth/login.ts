/**
 * 认证相关 API
 * @module api/backend/auth/login
 */
import { adminRequest } from '@/utils/http'

/**
 * 登录
 * @param params 登录参数
 * @returns 登录响应
 */
export function fetchLogin(params: Api.Auth.LoginParams) {
  return adminRequest.post<Api.Auth.LoginResponse>({
    url: '/auth/login',
    params
  })
}

/**
 * 获取用户信息
 * @returns 用户信息
 */
export function fetchGetUserInfo() {
  return adminRequest.get<Api.Auth.UserInfo>({
    url: '/auth/profile'
  })
}

/**
 * 刷新访问令牌
 * @param refreshToken 刷新令牌
 * @returns 新的访问令牌
 */
export function fetchRefreshToken(refreshToken: string) {
  return adminRequest.post<{ accessToken: string; expiresIn: number }>({
    url: '/auth/refresh',
    params: { refreshToken }
  })
}

/**
 * 退出登录
 */
export function fetchLogout() {
  return adminRequest.post({
    url: '/auth/logout'
  })
}

/**
 * 更新用户信息
 */
export function fetchUpdateProfile(params: {
  nickname: string
  realName?: string
  avatar?: string
  email: string
  mobile?: string
  address?: string
  gender: number
  remark?: string
}) {
  return adminRequest.post({
    url: '/auth/updateProfile',
    params
  })
}

/**
 * 修改密码
 */
export function fetchChangePassword(params: {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}) {
  return adminRequest.post({
    url: '/auth/changePassword',
    params
  })
}
