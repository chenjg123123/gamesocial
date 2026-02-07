import { request } from '../lib/request'

export type AnyRecord = Record<string, unknown>

export const listFrom = <T,>(res: unknown): T[] => {
  if (Array.isArray(res)) return res as T[]
  const items = res && typeof res === 'object' ? (res as { items?: unknown }).items : undefined
  return (Array.isArray(items) ? items : []) as T[]
}

export const authLoginWithOpenId = async (openId: string) => {
  const res = await request<{ token?: unknown; user?: unknown }>('/api/auth/wechat/login', {
    method: 'POST',
    data: { openId },
  })
  const token = res && typeof res.token === 'string' ? res.token : ''
  const user = (res && res.user && typeof res.user === 'object' ? (res.user as AnyRecord) : null) || null
  return { token, user }
}

export const getMe = async () => {
  return await request<AnyRecord>('/api/users/me')
}

export const updateMe = async (payload: { nickname: string; avatarUrl?: string; file?: File | null }) => {
  const nn = String(payload.nickname || '').trim()
  const file = payload.file || null
  if (file) {
    const fd = new FormData()
    fd.append('nickname', nn)
    fd.append('file', file)
    return await request<AnyRecord>('/api/users/me', { method: 'PUT', data: fd })
  }
  const av = String(payload.avatarUrl || '').trim()
  return await request<AnyRecord>('/api/users/me', { method: 'PUT', data: { nickname: nn, avatarUrl: av } })
}

export const apiMediaUpload = async (file: File) => {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await request<AnyRecord>('/api/media/upload', { method: 'POST', data: fd })
    const url = typeof res.url === 'string' ? res.url : typeof res.path === 'string' ? res.path : ''
    const path = typeof res.path === 'string' ? res.path : ''
    const createdAt = typeof res.createdAt === 'string' ? res.createdAt : ''
    return { url, path, createdAt }
  } catch (e) {
    const err = e as { message?: unknown }
    const msg = typeof err.message === 'string' ? err.message : ''
    if (msg.startsWith('HTTP 404')) {
      const res = await request<AnyRecord>('/admin/media/upload', { method: 'POST', data: fd })
      const url = typeof res.url === 'string' ? res.url : typeof res.path === 'string' ? res.path : ''
      const path = typeof res.path === 'string' ? res.path : ''
      const createdAt = typeof res.createdAt === 'string' ? res.createdAt : ''
      return { url, path, createdAt }
    }
    throw e
  }
}

export const getVipStatus = async () => {
  return await request<AnyRecord>('/api/vip/status')
}

export const getPointsBalance = async () => {
  return await request<AnyRecord>('/api/points/balance')
}

export const getPointsLedgers = async (offset: number, limit: number) => {
  const res = await request<unknown>(`/api/points/ledgers?offset=${offset}&limit=${limit}`)
  return listFrom<AnyRecord>(res)
}

export const getTasks = async () => {
  const res = await request<unknown>('/api/tasks')
  return listFrom<AnyRecord>(res)
}

export const taskCheckin = async () => {
  await request('/api/tasks/checkin', { method: 'POST' })
}

export const taskClaim = async (taskCode: string) => {
  const code = encodeURIComponent(taskCode)
  await request(`/api/tasks/${code}/claim`, { method: 'POST' })
}

export const listGoods = async (offset = 0, limit = 50) => {
  const res = await request<unknown>(`/api/goods?offset=${offset}&limit=${limit}`)
  return listFrom<AnyRecord>(res)
}

export const listRedeemOrders = async (offset = 0, limit = 50) => {
  const res = await request<unknown>(`/api/redeem/orders?offset=${offset}&limit=${limit}`)
  return listFrom<AnyRecord>(res)
}

export const getRedeemOrder = async (id: number) => {
  return await request<AnyRecord>(`/api/redeem/orders/${id}`)
}

export const createRedeemOrder = async (items: Array<{ goodsId: number; quantity: number; pointsPrice: number }>) => {
  return await request<AnyRecord>('/api/redeem/orders', { method: 'POST', data: { items } })
}

export const cancelRedeemOrder = async (id: number) => {
  await request(`/api/redeem/orders/${id}/cancel`, { method: 'PUT' })
}

export const listTournaments = async (offset = 0, limit = 50) => {
  const res = await request<unknown>(`/api/tournaments?offset=${offset}&limit=${limit}`)
  return listFrom<AnyRecord>(res)
}

export const getTournament = async (id: number) => {
  return await request<AnyRecord>(`/api/tournaments/${id}`)
}

export const joinTournament = async (id: number) => {
  await request(`/api/tournaments/${id}/join`, { method: 'POST' })
}

export const cancelTournamentJoin = async (id: number) => {
  await request(`/api/tournaments/${id}/cancel`, { method: 'PUT' })
}

export const getTournamentResults = async (id: number, offset = 0, limit = 50) => {
  const res = await request<AnyRecord>(`/api/tournaments/${id}/results?offset=${offset}&limit=${limit}`)
  const items = listFrom<AnyRecord>(res)
  const my = res && typeof res === 'object' ? ((res as { my?: unknown }).my as unknown) : undefined
  return { items, my }
}

export const listJoinedTournaments = async (offset = 0, limit = 50) => {
  const res = await request<unknown>(`/api/tournaments/joined?offset=${offset}&limit=${limit}`)
  return listFrom<AnyRecord>(res)
}

export const adminListGoods = async () => {
  const res = await request<unknown>('/admin/goods?offset=0&limit=50&status=0')
  return listFrom<AnyRecord>(res)
}

export const adminGetGoods = async (id: number) => {
  return await request<AnyRecord>(`/admin/goods/${id}`)
}

export const adminUpsertGoods = async (id: number | null, payload: AnyRecord) => {
  const file = payload.file instanceof File ? (payload.file as File) : null
  if (file) {
    const fd = new FormData()
    for (const [k, v] of Object.entries(payload)) {
      if (k === 'file' || k === 'coverUrl') continue
      if (v === undefined || v === null) continue
      fd.append(k, String(v))
    }
    fd.append('file', file)
    if (id) {
      await request(`/admin/goods/${id}`, { method: 'PUT', data: fd })
      return
    }
    await request('/admin/goods', { method: 'POST', data: fd })
    return
  }
  if (id) {
    await request(`/admin/goods/${id}`, { method: 'PUT', data: payload })
    return
  }
  await request('/admin/goods', { method: 'POST', data: payload })
}

export const adminDeleteGoods = async (id: number) => {
  await request(`/admin/goods/${id}`, { method: 'DELETE' })
}

export const adminListTournaments = async () => {
  const res = await request<unknown>('/admin/tournaments?offset=0&limit=50')
  return listFrom<AnyRecord>(res)
}

export const adminGetTournament = async (id: number) => {
  return await request<AnyRecord>(`/admin/tournaments/${id}`)
}

export const adminUpsertTournament = async (id: number | null, payload: AnyRecord) => {
  const file = payload.file instanceof File ? (payload.file as File) : null
  if (file) {
    const fd = new FormData()
    for (const [k, v] of Object.entries(payload)) {
      if (k === 'file' || k === 'coverUrl') continue
      if (v === undefined || v === null) continue
      fd.append(k, String(v))
    }
    fd.append('file', file)
    if (id) {
      await request(`/admin/tournaments/${id}`, { method: 'PUT', data: fd })
      return
    }
    await request('/admin/tournaments', { method: 'POST', data: fd })
    return
  }
  if (id) {
    await request(`/admin/tournaments/${id}`, { method: 'PUT', data: payload })
    return
  }
  await request('/admin/tournaments', { method: 'POST', data: payload })
}

export const adminDeleteTournament = async (id: number) => {
  await request(`/admin/tournaments/${id}`, { method: 'DELETE' })
}

export const adminListTaskDefs = async () => {
  const res = await request<unknown>('/admin/task-defs?offset=0&limit=50&status=0')
  return listFrom<AnyRecord>(res)
}

export const adminGetTaskDef = async (id: number) => {
  return await request<AnyRecord>(`/admin/task-defs/${id}`)
}

export const adminUpsertTaskDef = async (id: number | null, payload: AnyRecord) => {
  if (id) {
    await request(`/admin/task-defs/${id}`, { method: 'PUT', data: payload })
    return
  }
  await request('/admin/task-defs', { method: 'POST', data: payload })
}

export const adminDeleteTaskDef = async (id: number) => {
  await request(`/admin/task-defs/${id}`, { method: 'DELETE' })
}

export const adminListUsers = async () => {
  const res = await request<unknown>('/admin/users?offset=0&limit=50')
  return listFrom<AnyRecord>(res)
}

export const adminGetUser = async (id: number) => {
  return await request<AnyRecord>(`/admin/users/${id}`)
}

export const adminUpdateUser = async (id: number, payload: AnyRecord) => {
  const file = payload.file instanceof File ? (payload.file as File) : null
  if (file) {
    const fd = new FormData()
    for (const [k, v] of Object.entries(payload)) {
      if (k === 'file' || k === 'avatarUrl') continue
      if (v === undefined || v === null) continue
      fd.append(k, String(v))
    }
    fd.append('file', file)
    await request(`/admin/users/${id}`, { method: 'PUT', data: fd })
    return
  }
  await request(`/admin/users/${id}`, { method: 'PUT', data: payload })
}

export const adminListRedeemOrders = async () => {
  const res = await request<unknown>('/admin/redeem/orders?offset=0&limit=50')
  return listFrom<AnyRecord>(res)
}

export const adminUseRedeemOrder = async (id: number) => {
  await request(`/admin/redeem/orders/${id}/use`, { method: 'PUT' })
}

export const adminCancelRedeemOrder = async (id: number) => {
  await request(`/admin/redeem/orders/${id}/cancel`, { method: 'PUT' })
}

export const adminCreateRedeemOrder = async (items: Array<{ goodsId: number; quantity: number; pointsPrice: number }>) => {
  return await request<AnyRecord>('/admin/redeem/orders', { method: 'POST', data: { items } })
}

export const adminListAuditLogs = async (offset = 0, limit = 50) => {
  const res = await request<unknown>(`/admin/audit/logs?offset=${offset}&limit=${limit}`)
  return listFrom<AnyRecord>(res)
}

export const adminMediaUpload = async (file: File) => {
  const fd = new FormData()
  fd.append('file', file)
  const res = await request<AnyRecord>('/admin/media/upload', { method: 'POST', data: fd })
  const url = typeof res.url === 'string' ? res.url : ''
  const key = typeof res.key === 'string' ? res.key : ''
  const createdAt = typeof res.createdAt === 'string' ? res.createdAt : ''
  return { url, key, createdAt }
}
