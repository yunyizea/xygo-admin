/**
 * 前台会员认证 API
 *
 * 使用 Xy-User-Token 进行认证
 */
import { memberRequest } from '@/utils/http'

/** 登录参数 */
export interface MemberLoginParams {
  username: string
  password: string
  captcha?: string
  captchaId?: string
}

/** 登录响应 */
export interface MemberLoginResult {
  token: string
  expiresIn: number
  refreshToken: string
  refreshExpiresIn: number
}

/** 注册参数 */
export interface MemberRegisterParams {
  username: string
  password: string
  mobile?: string
  email?: string
  code?: string
}

/** 注册响应 */
export interface MemberRegisterResult {
  id: number
}

/**
 * 会员登录
 */
export function memberLogin(params: MemberLoginParams) {
  return memberRequest.post<MemberLoginResult>({
    url: '/auth/login',
    data: params
  })
}

/**
 * 会员注册
 */
export function memberRegister(params: MemberRegisterParams) {
  return memberRequest.post<MemberRegisterResult>({
    url: '/auth/register',
    data: params
  })
}

/**
 * 会员刷新访问令牌
 */
export function memberRefreshToken(refreshToken: string) {
  return memberRequest.post<{ accessToken: string; expiresIn: number }>({
    url: '/auth/refresh',
    data: { refreshToken }
  })
}

/**
 * 会员退出登录
 */
export function memberLogout() {
  return memberRequest.post<void>({
    url: '/auth/logout'
  })
}

/** 点选验证码响应 */
export interface ClickCaptchaResult {
  id: string
  text: string[]
  base64: string
  width: number
  height: number
}

/**
 * 获取点选验证码
 */
export function getClickCaptcha() {
  return memberRequest.get<ClickCaptchaResult>({
    url: '/auth/captcha'
  })
}

/**
 * 校验点选验证码
 */
export function checkClickCaptcha(id: string, info: string) {
  return memberRequest.post<void>({
    url: '/auth/checkCaptcha',
    data: { id, info }
  })
}
