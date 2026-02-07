<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import {
  adminCancelRedeemOrder,
  adminCreateRedeemOrder,
  adminDeleteGoods,
  adminDeleteTaskDef,
  adminDeleteTournament,
  adminGetGoods,
  adminGetTaskDef,
  adminGetTournament,
  adminGetUser,
  adminListGoods,
  adminListRedeemOrders,
  adminListTaskDefs,
  adminListTournaments,
  adminListUsers,
  adminListAuditLogs,
  adminUpdateUser,
  adminUpsertGoods,
  adminUpsertTaskDef,
  adminUpsertTournament,
  adminUseRedeemOrder,
} from '../../api'
import { useToastStore } from '../../stores/toast'

type TabKey = 'goods' | 'tournaments' | 'taskDefs' | 'users' | 'orders' | 'audit'

const toast = useToastStore()

const tab = ref<TabKey>('goods')
const loading = ref(false)

const goods = ref<Array<Record<string, unknown>>>([])
const tournaments = ref<Array<Record<string, unknown>>>([])
const taskDefs = ref<Array<Record<string, unknown>>>([])
const users = ref<Array<Record<string, unknown>>>([])
const orders = ref<Array<Record<string, unknown>>>([])
const auditLogs = ref<Array<Record<string, unknown>>>([])

const goodsForm = ref({ id: '', name: '', coverUrl: '', pointsPrice: '', stock: '', status: '1' })
const tournamentForm = ref({
  id: '',
  title: '',
  content: '',
  coverUrl: '',
  startAt: '',
  endAt: '',
  status: 'DRAFT',
  createdByAdminId: '1',
})
const taskDefForm = ref({
  id: '',
  taskCode: '',
  name: '',
  periodType: 'DAILY',
  targetCount: '1',
  rewardPoints: '1',
  status: '1',
})
const userForm = ref({ id: '', nickname: '', avatarUrl: '', status: '1' })
const orderForm = ref({ goodsId: '', quantity: '1', pointsPrice: '' })

const goodsCoverFileEl = ref<HTMLInputElement | null>(null)
const tournamentCoverFileEl = ref<HTMLInputElement | null>(null)
const userAvatarFileEl = ref<HTMLInputElement | null>(null)

const goodsCoverFile = ref<File | null>(null)
const tournamentCoverFile = ref<File | null>(null)
const userAvatarFile = ref<File | null>(null)

const goodsCoverPreviewUrl = ref('')
const tournamentCoverPreviewUrl = ref('')
const userAvatarPreviewUrl = ref('')

const MAX_IMAGE_BYTES = 5 * 1024 * 1024

const validateImageFile = (file: File) => {
  if (!file) return '请选择图片文件'
  if (!file.type || !file.type.startsWith('image/')) return '仅支持 image/* 图片文件'
  if (file.size > MAX_IMAGE_BYTES) return '图片大小不能超过 5MB'
  return ''
}

const pickFileFromChange = (e: Event) => {
  const input = e.target as HTMLInputElement | null
  const f = input?.files && input.files.length > 0 ? input.files[0] : null
  if (input) input.value = ''
  return f
}

const triggerPick = (el: HTMLInputElement | null) => {
  el?.click()
}

const clearPreviewUrl = (url: string) => {
  if (!url) return
  URL.revokeObjectURL(url)
}

const onPickGoodsCover = (e: Event) => {
  const file = pickFileFromChange(e)
  if (!file) return
  const msg = validateImageFile(file)
  if (msg) {
    toast.show(msg, 'error')
    return
  }
  goodsCoverFile.value = file
  clearPreviewUrl(goodsCoverPreviewUrl.value)
  goodsCoverPreviewUrl.value = URL.createObjectURL(file)
  toast.show('已选择封面，提交时上传', 'success')
}

const onPickTournamentCover = (e: Event) => {
  const file = pickFileFromChange(e)
  if (!file) return
  const msg = validateImageFile(file)
  if (msg) {
    toast.show(msg, 'error')
    return
  }
  tournamentCoverFile.value = file
  clearPreviewUrl(tournamentCoverPreviewUrl.value)
  tournamentCoverPreviewUrl.value = URL.createObjectURL(file)
  toast.show('已选择封面，提交时上传', 'success')
}

const onPickUserAvatar = (e: Event) => {
  const file = pickFileFromChange(e)
  if (!file) return
  const msg = validateImageFile(file)
  if (msg) {
    toast.show(msg, 'error')
    return
  }
  userAvatarFile.value = file
  clearPreviewUrl(userAvatarPreviewUrl.value)
  userAvatarPreviewUrl.value = URL.createObjectURL(file)
  toast.show('已选择头像，提交时上传', 'success')
}

const title = computed(() => {
  if (tab.value === 'goods') return '商品管理'
  if (tab.value === 'tournaments') return '赛事管理'
  if (tab.value === 'taskDefs') return '任务定义'
  if (tab.value === 'users') return '用户管理'
  if (tab.value === 'orders') return '订单管理'
  return '审计日志'
})

const asNumber = (v: unknown) => {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

const safeStr = (v: unknown) => (v === undefined || v === null ? '' : String(v))

const withLoading = async <T>(fn: () => Promise<T>) => {
  loading.value = true
  try {
    return await fn()
  } finally {
    loading.value = false
  }
}

const refreshCurrent = async () => {
  if (tab.value === 'goods') return loadGoods()
  if (tab.value === 'tournaments') return loadTournaments()
  if (tab.value === 'taskDefs') return loadTaskDefs()
  if (tab.value === 'users') return loadUsers()
  if (tab.value === 'orders') return loadOrders()
  return loadAuditLogs()
}

const loadGoods = async () => {
  try {
    await withLoading(async () => {
      goods.value = await adminListGoods()
    })
  } catch (e) {
    const err = e as { message?: unknown }
    goods.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const editGoods = async (id: number) => {
  try {
    await withLoading(async () => {
      const item = await adminGetGoods(id)
      goodsForm.value = {
        id: safeStr(item.id),
        name: safeStr(item.name),
        coverUrl: safeStr(item.coverUrl),
        pointsPrice: safeStr(item.pointsPrice),
        stock: safeStr(item.stock),
        status: safeStr(item.status) || '1',
      }
    })
    goodsCoverFile.value = null
    clearPreviewUrl(goodsCoverPreviewUrl.value)
    goodsCoverPreviewUrl.value = ''
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const resetGoodsForm = () => {
  goodsForm.value = { id: '', name: '', coverUrl: '', pointsPrice: '', stock: '', status: '1' }
  goodsCoverFile.value = null
  clearPreviewUrl(goodsCoverPreviewUrl.value)
  goodsCoverPreviewUrl.value = ''
}

const submitGoods = async () => {
  const form = goodsForm.value
  const name = String(form.name || '').trim()
  if (!name) {
    toast.show('请输入商品名', 'error')
    return
  }
  const id = asNumber(form.id)
  try {
    await withLoading(async () => {
      const payload = {
        name: String(goodsForm.value.name || '').trim(),
        pointsPrice: asNumber(goodsForm.value.pointsPrice),
        stock: asNumber(goodsForm.value.stock),
        status: asNumber(String(goodsForm.value.status || '1').trim() || '1'),
        file: goodsCoverFile.value || undefined,
      }
      await adminUpsertGoods(id || null, payload)
    })
    toast.show(id ? '已更新' : '已创建', 'success')
    resetGoodsForm()
    await loadGoods()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '提交失败', 'error')
  }
}

const deleteGoods = async (id: number) => {
  try {
    await withLoading(async () => {
      await adminDeleteGoods(id)
    })
    toast.show('已删除', 'success')
    resetGoodsForm()
    await loadGoods()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '删除失败', 'error')
  }
}

const loadTournaments = async () => {
  try {
    await withLoading(async () => {
      tournaments.value = await adminListTournaments()
    })
  } catch (e) {
    const err = e as { message?: unknown }
    tournaments.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const editTournament = async (id: number) => {
  try {
    await withLoading(async () => {
      const item = await adminGetTournament(id)
      tournamentForm.value = {
        id: safeStr(item.id),
        title: safeStr(item.title),
        content: safeStr(item.content),
        coverUrl: safeStr(item.coverUrl),
        startAt: safeStr(item.startAt),
        endAt: safeStr(item.endAt),
        status: safeStr(item.status) || 'DRAFT',
        createdByAdminId: safeStr(item.createdByAdminId) || '1',
      }
    })
    tournamentCoverFile.value = null
    clearPreviewUrl(tournamentCoverPreviewUrl.value)
    tournamentCoverPreviewUrl.value = ''
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const resetTournamentForm = () => {
  tournamentForm.value = {
    id: '',
    title: '',
    content: '',
    coverUrl: '',
    startAt: '',
    endAt: '',
    status: 'DRAFT',
    createdByAdminId: '1',
  }
  tournamentCoverFile.value = null
  clearPreviewUrl(tournamentCoverPreviewUrl.value)
  tournamentCoverPreviewUrl.value = ''
}

const submitTournament = async () => {
  const form = tournamentForm.value
  const title = String(form.title || '').trim()
  if (!title) {
    toast.show('请输入标题', 'error')
    return
  }
  const id = asNumber(form.id)
  try {
    await withLoading(async () => {
      const payload = {
        title: String(tournamentForm.value.title || '').trim(),
        content: String(tournamentForm.value.content || '').trim(),
        startAt: String(tournamentForm.value.startAt || '').trim(),
        endAt: String(tournamentForm.value.endAt || '').trim(),
        status: String(tournamentForm.value.status || '').trim() || 'DRAFT',
        createdByAdminId: asNumber(tournamentForm.value.createdByAdminId) || 1,
        file: tournamentCoverFile.value || undefined,
      }
      await adminUpsertTournament(id || null, payload)
    })
    toast.show(id ? '已更新' : '已创建', 'success')
    resetTournamentForm()
    await loadTournaments()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '提交失败', 'error')
  }
}

const deleteTournament = async (id: number) => {
  try {
    await withLoading(async () => {
      await adminDeleteTournament(id)
    })
    toast.show('已删除', 'success')
    resetTournamentForm()
    await loadTournaments()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '删除失败', 'error')
  }
}

const loadTaskDefs = async () => {
  try {
    await withLoading(async () => {
      taskDefs.value = await adminListTaskDefs()
    })
  } catch (e) {
    const err = e as { message?: unknown }
    taskDefs.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const editTaskDef = async (id: number) => {
  try {
    await withLoading(async () => {
      const item = await adminGetTaskDef(id)
      taskDefForm.value = {
        id: safeStr(item.id),
        taskCode: safeStr(item.taskCode),
        name: safeStr(item.name),
        periodType: safeStr(item.periodType) || 'DAILY',
        targetCount: safeStr(item.targetCount),
        rewardPoints: safeStr(item.rewardPoints),
        status: safeStr(item.status) || '1',
      }
    })
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const resetTaskDefForm = () => {
  taskDefForm.value = {
    id: '',
    taskCode: '',
    name: '',
    periodType: 'DAILY',
    targetCount: '1',
    rewardPoints: '1',
    status: '1',
  }
}

const submitTaskDef = async () => {
  const form = taskDefForm.value
  const id = asNumber(form.id)
  const taskCode = String(form.taskCode || '').trim()
  const name = String(form.name || '').trim()
  const payload = id
    ? {
        name,
        periodType: String(form.periodType || '').trim() || 'DAILY',
        targetCount: asNumber(form.targetCount),
        rewardPoints: asNumber(form.rewardPoints),
        status: asNumber(String(form.status || '1').trim() || '1'),
        ruleJson: {},
      }
    : {
        taskCode,
        name,
        periodType: String(form.periodType || '').trim() || 'DAILY',
        targetCount: asNumber(form.targetCount),
        rewardPoints: asNumber(form.rewardPoints),
        status: asNumber(String(form.status || '1').trim() || '1'),
        ruleJson: {},
      }
  if (!payload.name) {
    toast.show('请输入任务名', 'error')
    return
  }
  if (!id && !taskCode) {
    toast.show('请输入 taskCode', 'error')
    return
  }
  try {
    await withLoading(async () => {
      await adminUpsertTaskDef(id || null, payload as Record<string, unknown>)
    })
    toast.show(id ? '已更新' : '已创建', 'success')
    resetTaskDefForm()
    await loadTaskDefs()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '提交失败', 'error')
  }
}

const deleteTaskDef = async (id: number) => {
  try {
    await withLoading(async () => {
      await adminDeleteTaskDef(id)
    })
    toast.show('已删除', 'success')
    resetTaskDefForm()
    await loadTaskDefs()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '删除失败', 'error')
  }
}

const loadUsers = async () => {
  try {
    await withLoading(async () => {
      users.value = await adminListUsers()
    })
  } catch (e) {
    const err = e as { message?: unknown }
    users.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const editUser = async (id: number) => {
  try {
    await withLoading(async () => {
      const item = await adminGetUser(id)
      userForm.value = {
        id: safeStr(item.id),
        nickname: safeStr(item.nickname),
        avatarUrl: safeStr(item.avatarUrl),
        status: safeStr(item.status) || '1',
      }
    })
    userAvatarFile.value = null
    clearPreviewUrl(userAvatarPreviewUrl.value)
    userAvatarPreviewUrl.value = ''
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const resetUserForm = () => {
  userForm.value = { id: '', nickname: '', avatarUrl: '', status: '1' }
  userAvatarFile.value = null
  clearPreviewUrl(userAvatarPreviewUrl.value)
  userAvatarPreviewUrl.value = ''
}

const submitUser = async () => {
  const form = userForm.value
  const id = asNumber(form.id)
  if (!id) {
    toast.show('请选择用户', 'error')
    return
  }
  try {
    await withLoading(async () => {
      const payload = {
        nickname: String(userForm.value.nickname || '').trim(),
        avatarUrl: String(userForm.value.avatarUrl || '').trim(),
        status: asNumber(String(userForm.value.status || '1').trim() || '1'),
        file: userAvatarFile.value || undefined,
      }
      await adminUpdateUser(id, payload)
    })
    toast.show('已更新', 'success')
    resetUserForm()
    await loadUsers()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '提交失败', 'error')
  }
}

const loadOrders = async () => {
  try {
    await withLoading(async () => {
      orders.value = await adminListRedeemOrders()
    })
  } catch (e) {
    const err = e as { message?: unknown }
    orders.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const loadAuditLogs = async () => {
  try {
    await withLoading(async () => {
      auditLogs.value = await adminListAuditLogs()
    })
  } catch (e) {
    const err = e as { message?: unknown }
    auditLogs.value = []
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  }
}

const useOrder = async (id: number) => {
  try {
    await withLoading(async () => {
      await adminUseRedeemOrder(id)
    })
    toast.show('已核销', 'success')
    await loadOrders()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '操作失败', 'error')
  }
}

const cancelOrder = async (id: number) => {
  try {
    await withLoading(async () => {
      await adminCancelRedeemOrder(id)
    })
    toast.show('已取消', 'success')
    await loadOrders()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '操作失败', 'error')
  }
}

const resetOrderForm = () => {
  orderForm.value = { goodsId: '', quantity: '1', pointsPrice: '' }
}

const submitOrder = async () => {
  const form = orderForm.value
  const goodsId = asNumber(form.goodsId)
  const quantity = asNumber(form.quantity)
  const pointsPriceStr = safeStr(form.pointsPrice).trim()
  const pointsPrice = asNumber(pointsPriceStr)
  if (!goodsId || !quantity || pointsPriceStr === '' || pointsPrice < 0) {
    toast.show('请填写 goodsId/数量/pointsPrice', 'error')
    return
  }
  const items = [{ goodsId, quantity, pointsPrice }]
  try {
    const res = await withLoading(async () => {
      return await adminCreateRedeemOrder(items)
    })
    const orderNo = safeStr(res.orderNo || res.id)
    toast.show(orderNo ? `创建成功：${orderNo}` : '创建成功', 'success')
    resetOrderForm()
    await loadOrders()
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '提交失败', 'error')
  }
}

const switchTab = async (k: TabKey) => {
  tab.value = k
  await refreshCurrent()
}

onMounted(() => {
  void refreshCurrent()
})
</script>

<template>
  <div class="topbar">
    <div class="topbar__inner container">
      <div class="title">管理端</div>
      <RouterLink class="tab" to="/user/index">用户端</RouterLink>
    </div>
    <div class="tabs container">
      <button class="tab" :class="{ 'tab--active': tab === 'goods' }" @click="switchTab('goods')">商品</button>
      <button class="tab" :class="{ 'tab--active': tab === 'tournaments' }" @click="switchTab('tournaments')">赛事</button>
      <button class="tab" :class="{ 'tab--active': tab === 'taskDefs' }" @click="switchTab('taskDefs')">任务</button>
      <button class="tab" :class="{ 'tab--active': tab === 'users' }" @click="switchTab('users')">用户</button>
      <button class="tab" :class="{ 'tab--active': tab === 'orders' }" @click="switchTab('orders')">订单</button>
      <button class="tab" :class="{ 'tab--active': tab === 'audit' }" @click="switchTab('audit')">审计</button>
    </div>
  </div>

  <div class="page">
    <div class="container grid">
      <div class="row">
        <div class="title">{{ title }}</div>
        <div class="spacer" />
        <button class="btn btn--ghost" :disabled="loading" @click="refreshCurrent">刷新</button>
      </div>

      <div v-if="tab === 'goods'" class="grid">
        <div class="card">
          <div class="title">新增/编辑</div>
          <div class="grid" style="margin-top: 10px">
            <input v-model="goodsForm.name" class="input" placeholder="商品名" />
            <div class="row">
              <button class="btn btn--ghost" :disabled="loading" @click="triggerPick(goodsCoverFileEl)">选择封面</button>
              <input
                ref="goodsCoverFileEl"
                style="display: none"
                type="file"
                accept="image/*"
                @change="onPickGoodsCover"
              />
            </div>
            <div v-if="goodsCoverFile" class="help">已选择：{{ goodsCoverFile.name }}</div>
            <img
              v-if="goodsCoverPreviewUrl || goodsForm.coverUrl"
              :src="goodsCoverPreviewUrl || goodsForm.coverUrl"
              alt="cover"
              style="width: 120px; height: 120px; object-fit: cover; border-radius: 10px"
            />
            <div class="row">
              <input v-model="goodsForm.pointsPrice" class="input" placeholder="积分价格" />
              <input v-model="goodsForm.stock" class="input" placeholder="库存" />
              <input v-model="goodsForm.status" class="input" placeholder="状态(1/0)" />
            </div>
            <div class="row">
              <button class="btn" :disabled="loading" @click="submitGoods">{{ goodsForm.id ? '更新' : '创建' }}</button>
              <button class="btn btn--ghost" :disabled="loading" @click="resetGoodsForm">清空</button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="title">列表</div>
          <div class="grid" style="margin-top: 10px">
            <div v-if="goods.length === 0" class="muted">暂无数据</div>
            <div v-for="g in goods" :key="safeStr(g.id)" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ safeStr(g.name) || `#${safeStr(g.id)}` }}</div>
                  <div class="muted" style="margin-top: 6px">
                    id={{ safeStr(g.id) }} · price={{ safeStr(g.pointsPrice) }} · stock={{ safeStr(g.stock) }}
                  </div>
                </div>
                <div class="spacer" />
                <button class="btn btn--ghost" :disabled="loading" @click="editGoods(asNumber(g.id))">编辑</button>
                <button class="btn btn--danger" :disabled="loading" @click="deleteGoods(asNumber(g.id))">删除</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="tab === 'tournaments'" class="grid">
        <div class="card">
          <div class="title">新增/编辑</div>
          <div class="grid" style="margin-top: 10px">
            <input v-model="tournamentForm.title" class="input" placeholder="标题" />
            <div class="row">
              <button class="btn btn--ghost" :disabled="loading" @click="triggerPick(tournamentCoverFileEl)">选择封面</button>
              <input
                ref="tournamentCoverFileEl"
                style="display: none"
                type="file"
                accept="image/*"
                @change="onPickTournamentCover"
              />
            </div>
            <div v-if="tournamentCoverFile" class="help">已选择：{{ tournamentCoverFile.name }}</div>
            <img
              v-if="tournamentCoverPreviewUrl || tournamentForm.coverUrl"
              :src="tournamentCoverPreviewUrl || tournamentForm.coverUrl"
              alt="cover"
              style="width: 120px; height: 120px; object-fit: cover; border-radius: 10px"
            />
            <div class="row">
              <input v-model="tournamentForm.startAt" class="input" placeholder="startAt(RFC3339)" />
              <input v-model="tournamentForm.endAt" class="input" placeholder="endAt(RFC3339)" />
            </div>
            <input v-model="tournamentForm.status" class="input" placeholder="status(DRAFT/PUBLISHED...)" />
            <input v-model="tournamentForm.createdByAdminId" class="input" placeholder="createdByAdminId" />
            <input v-model="tournamentForm.content" class="input" placeholder="内容" />
            <div class="row">
              <button class="btn" :disabled="loading" @click="submitTournament">
                {{ tournamentForm.id ? '更新' : '创建' }}
              </button>
              <button class="btn btn--ghost" :disabled="loading" @click="resetTournamentForm">清空</button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="title">列表</div>
          <div class="grid" style="margin-top: 10px">
            <div v-if="tournaments.length === 0" class="muted">暂无数据</div>
            <div v-for="t in tournaments" :key="safeStr(t.id)" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ safeStr(t.title) || `#${safeStr(t.id)}` }}</div>
                  <div class="muted" style="margin-top: 6px">id={{ safeStr(t.id) }} · status={{ safeStr(t.status) }}</div>
                </div>
                <div class="spacer" />
                <button class="btn btn--ghost" :disabled="loading" @click="editTournament(asNumber(t.id))">编辑</button>
                <button class="btn btn--danger" :disabled="loading" @click="deleteTournament(asNumber(t.id))">删除</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="tab === 'taskDefs'" class="grid">
        <div class="card">
          <div class="title">新增/编辑</div>
          <div class="grid" style="margin-top: 10px">
            <input v-model="taskDefForm.taskCode" class="input" placeholder="taskCode(新建必填)" />
            <input v-model="taskDefForm.name" class="input" placeholder="任务名" />
            <div class="row">
              <input v-model="taskDefForm.periodType" class="input" placeholder="periodType(DAILY...)" />
              <input v-model="taskDefForm.targetCount" class="input" placeholder="targetCount" />
              <input v-model="taskDefForm.rewardPoints" class="input" placeholder="rewardPoints" />
              <input v-model="taskDefForm.status" class="input" placeholder="status(1/0)" />
            </div>
            <div class="row">
              <button class="btn" :disabled="loading" @click="submitTaskDef">{{ taskDefForm.id ? '更新' : '创建' }}</button>
              <button class="btn btn--ghost" :disabled="loading" @click="resetTaskDefForm">清空</button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="title">列表</div>
          <div class="grid" style="margin-top: 10px">
            <div v-if="taskDefs.length === 0" class="muted">暂无数据</div>
            <div v-for="t in taskDefs" :key="safeStr(t.id)" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ safeStr(t.name) || `#${safeStr(t.id)}` }}</div>
                  <div class="muted" style="margin-top: 6px">
                    id={{ safeStr(t.id) }} · code={{ safeStr(t.taskCode) }} · period={{ safeStr(t.periodType) }}
                  </div>
                </div>
                <div class="spacer" />
                <button class="btn btn--ghost" :disabled="loading" @click="editTaskDef(asNumber(t.id))">编辑</button>
                <button class="btn btn--danger" :disabled="loading" @click="deleteTaskDef(asNumber(t.id))">删除</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="tab === 'users'" class="grid">
        <div class="card">
          <div class="title">编辑用户（先点列表里的编辑）</div>
          <div class="grid" style="margin-top: 10px">
            <div class="muted">当前用户 id：{{ userForm.id || '-' }}</div>
            <input v-model="userForm.nickname" class="input" placeholder="昵称" />
            <div class="row">
              <button class="btn btn--ghost" :disabled="loading" @click="triggerPick(userAvatarFileEl)">选择头像</button>
              <input
                ref="userAvatarFileEl"
                style="display: none"
                type="file"
                accept="image/*"
                @change="onPickUserAvatar"
              />
            </div>
            <div v-if="userAvatarFile" class="help">已选择：{{ userAvatarFile.name }}</div>
            <img
              v-if="userAvatarPreviewUrl || userForm.avatarUrl"
              :src="userAvatarPreviewUrl || userForm.avatarUrl"
              alt="avatar"
              style="width: 80px; height: 80px; object-fit: cover; border-radius: 999px"
            />
            <input v-model="userForm.status" class="input" placeholder="status(1/0)" />
            <div class="row">
              <button class="btn" :disabled="loading" @click="submitUser">保存</button>
              <button class="btn btn--ghost" :disabled="loading" @click="resetUserForm">清空</button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="title">列表</div>
          <div class="grid" style="margin-top: 10px">
            <div v-if="users.length === 0" class="muted">暂无数据</div>
            <div v-for="u in users" :key="safeStr(u.id)" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ safeStr(u.nickname) || `用户#${safeStr(u.id)}` }}</div>
                  <div class="muted" style="margin-top: 6px">id={{ safeStr(u.id) }} · status={{ safeStr(u.status) }}</div>
                </div>
                <div class="spacer" />
                <button class="btn btn--ghost" :disabled="loading" @click="editUser(asNumber(u.id))">编辑</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="tab === 'orders'" class="grid">
        <div class="card">
          <div class="title">创建订单</div>
          <div class="grid" style="margin-top: 10px">
            <div class="row">
              <input v-model="orderForm.goodsId" class="input" placeholder="goodsId" />
            </div>
            <div class="row">
              <input v-model="orderForm.quantity" class="input" placeholder="数量" />
              <input v-model="orderForm.pointsPrice" class="input" placeholder="pointsPrice" />
            </div>
            <div class="row">
              <button class="btn" :disabled="loading" @click="submitOrder">创建</button>
              <button class="btn btn--ghost" :disabled="loading" @click="resetOrderForm">清空</button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="title">订单列表</div>
          <div class="grid" style="margin-top: 10px">
            <div v-if="orders.length === 0" class="muted">暂无数据</div>
            <div v-for="o in orders" :key="safeStr(o.id)" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ safeStr(o.orderNo) || `订单#${safeStr(o.id)}` }}</div>
                  <div class="muted" style="margin-top: 6px">
                    id={{ safeStr(o.id) }} · status={{ safeStr(o.status) }}
                  </div>
                </div>
                <div class="spacer" />
                <button class="btn btn--ghost" :disabled="loading" @click="useOrder(asNumber(o.id))">核销</button>
                <button class="btn btn--danger" :disabled="loading" @click="cancelOrder(asNumber(o.id))">取消</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="grid">
        <div class="card">
          <div class="title">审计日志</div>
          <div class="grid" style="margin-top: 10px">
            <div v-if="auditLogs.length === 0" class="muted">暂无数据</div>
            <div v-for="l in auditLogs" :key="safeStr(l.id)" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ safeStr(l.action) || '操作' }}</div>
                  <div class="muted" style="margin-top: 6px">
                    admin={{ safeStr(l.adminId) }} · ip={{ safeStr(l.ip) }} · {{ safeStr(l.createdAt) }}
                  </div>
                  <div v-if="l.targetId" class="muted" style="margin-top: 4px">targetId={{ l.targetId }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
