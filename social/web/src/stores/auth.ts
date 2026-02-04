import { defineStore } from 'pinia'
import { ref } from 'vue'

import { clearSession, getToken, getUserCache, setToken as persistToken, setUserCache } from '../lib/storage'
import { authLoginWithOpenId, getMe } from '../api'

type UserProfile = {
  id?: number
  openId?: string
  nickname?: string
  avatarUrl?: string
  [k: string]: unknown
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(getToken())
  const user = ref<UserProfile | null>((getUserCache() as UserProfile | null) || null)

  const setUser = (u: UserProfile | null) => {
    user.value = u
    setUserCache(u)
  }

  const setToken = (t: string) => {
    token.value = t
    persistToken(t)
  }

  const clear = () => {
    token.value = ''
    user.value = null
    clearSession()
  }

  const refreshProfile = async () => {
    const profile = (await getMe()) as UserProfile
    setUser(profile)
    return profile
  }

  const loginWithOpenId = async (openId: string) => {
    const { token: t, user: u } = await authLoginWithOpenId(openId)
    if (!t) throw new Error('登录失败')
    setToken(t)
    if (u) {
      const profile = u as UserProfile
      setUser(profile)
      return profile
    }
    return await refreshProfile()
  }

  const logout = () => clear()

  const loginWithWechatCode = loginWithOpenId

  return { token, user, setUser, setToken, clear, logout, refreshProfile, loginWithOpenId, loginWithWechatCode }
})
