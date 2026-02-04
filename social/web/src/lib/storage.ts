const STORAGE_KEYS = {
  openId: 'gs_openid',
  token: 'gs_token',
  user: 'gs_user',
} as const

export type UserCache = Record<string, unknown> | null

export const getOpenId = () => window.localStorage.getItem(STORAGE_KEYS.openId) || ''

export const setOpenId = (openId: string) => {
  window.localStorage.setItem(STORAGE_KEYS.openId, openId)
}

export const clearOpenId = () => {
  window.localStorage.removeItem(STORAGE_KEYS.openId)
}

export const getToken = () => window.localStorage.getItem(STORAGE_KEYS.token) || ''

export const setToken = (token: string) => {
  window.localStorage.setItem(STORAGE_KEYS.token, token)
}

export const clearToken = () => {
  window.localStorage.removeItem(STORAGE_KEYS.token)
}

export const getUserCache = (): UserCache => {
  const raw = window.localStorage.getItem(STORAGE_KEYS.user)
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserCache
  } catch {
    return null
  }
}

export const setUserCache = (user: unknown) => {
  window.localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(user ?? null))
}

export const clearUserCache = () => {
  window.localStorage.removeItem(STORAGE_KEYS.user)
}

export const clearSession = () => {
  clearOpenId()
  clearToken()
  clearUserCache()
}
