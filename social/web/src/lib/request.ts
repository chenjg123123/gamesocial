import { clearSession, getToken } from './storage'

const DEFAULT_BASE_URL = 'https://www.u835587.nyat.app:26462'

export const BASE_URL = (() => {
  const v = (import.meta as unknown as { env?: Record<string, string | undefined> }).env?.VITE_API_BASE
  return (v || '').trim() || DEFAULT_BASE_URL
})()

const normalizeErrorMessage = (data: unknown) => {
  if (!data || typeof data !== 'object') return '请求失败'
  const d = data as { message?: unknown; msg?: unknown }
  return (typeof d.message === 'string' && d.message) || (typeof d.msg === 'string' && d.msg) || '请求失败'
}

const extractData = (data: unknown) => {
  if (!data || typeof data !== 'object') return data
  const d = data as { data?: unknown }
  if (Object.prototype.hasOwnProperty.call(d, 'data')) return d.data
  return data
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: unknown
  headers?: Record<string, string>
  timeout?: number
}

export const request = async <T = unknown>(path: string, options?: RequestOptions): Promise<T> => {
  const opt = options || {}

  const controller = new AbortController()
  const timeoutMs = typeof opt.timeout === 'number' ? opt.timeout : 15000
  const timer = window.setTimeout(() => controller.abort(), timeoutMs)

  try {
    const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(opt.headers || {}) }
    const token = getToken()
    if (token && !headers.Authorization) headers.Authorization = `Bearer ${token}`

    const res = await fetch(`${BASE_URL}${path}`, {
      method: opt.method || 'GET',
      headers,
      body: opt.data === undefined ? undefined : JSON.stringify(opt.data),
      signal: controller.signal,
      credentials: 'include',
    })

    const status = res.status
    let payload: unknown = null
    const text = await res.text()
    if (text) {
      try {
        payload = JSON.parse(text) as unknown
      } catch {
        payload = text
      }
    }

    if (status < 200 || status >= 300) {
      if (status === 401) {
        clearSession()
        throw new Error('unauthorized')
      }
      if (status === 403) throw new Error('forbidden')
      const msg = normalizeErrorMessage(payload)
      throw new Error(msg === '请求失败' ? `HTTP ${status}` : msg)
    }

    const envelope = payload as { code?: unknown; message?: unknown; msg?: unknown }
    const hasBizCode =
      envelope && typeof envelope === 'object' && Object.prototype.hasOwnProperty.call(envelope, 'code')
    if (hasBizCode) {
      const code = Number(envelope.code)
      if (code === 200) return extractData(payload) as T
      if (code === 1001) {
        clearSession()
        throw new Error('unauthorized')
      }
      if (code === 401) {
        clearSession()
        throw new Error('unauthorized')
      }
      if (code === 403) throw new Error('forbidden')
      throw new Error(
        (typeof envelope.message === 'string' && envelope.message) ||
          (typeof envelope.msg === 'string' && envelope.msg) ||
          '请求失败'
      )
    }

    return extractData(payload) as T
  } catch (err) {
    const e = err as { name?: unknown }
    if (e && e.name === 'AbortError') throw new Error('请求超时')
    throw err
  } finally {
    window.clearTimeout(timer)
  }
}
