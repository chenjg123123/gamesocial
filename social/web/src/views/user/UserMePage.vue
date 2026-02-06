<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getMe, getPointsLedgers, updateMe } from '../../api'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'

type LedgerItem = {
  id?: number
  type?: string
  delta?: number
  balanceAfter?: number
  createdAt?: string
  remark?: string
  [k: string]: unknown
}

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()

const loading = ref(false)
const saving = ref(false)

const nickname = ref('')
const avatarUrl = ref('')
const avatarFileEl = ref<HTMLInputElement | null>(null)
const pendingAvatarFile = ref<File | null>(null)
const pendingAvatarPreviewUrl = ref('')

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

const triggerPick = () => {
  avatarFileEl.value?.click()
}

const profileOpen = ref(false)
const ledgerOpen = ref(false)
const ledgerLoading = ref(false)
const ledgerMoreLoading = ref(false)
const ledger = ref<LedgerItem[]>([])
const ledgerHasMore = ref(false)

const logout = () => {
  if (!confirm('确定要退出登录吗？')) return
  auth.clear()
  nickname.value = ''
  avatarUrl.value = ''
  ledger.value = []
  ledgerHasMore.value = false
  ledgerOpen.value = false
  toast.show('已退出', 'success')
  void router.replace('/login')
}

const applyProfile = (profile: Record<string, unknown> | null) => {
  const nn = profile && typeof profile.nickname === 'string' ? profile.nickname : ''
  const av = profile && typeof profile.avatarUrl === 'string' ? profile.avatarUrl : ''
  nickname.value = nn
  avatarUrl.value = av
}

const refresh = async () => {
  loading.value = true
  try {
    const profile = await getMe()
    auth.setUser(profile)
    applyProfile(profile)
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      void router.replace({ path: '/login', query: { redirect: '/user/me' } })
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const openProfileEdit = () => {
  pendingAvatarFile.value = null
  if (pendingAvatarPreviewUrl.value) URL.revokeObjectURL(pendingAvatarPreviewUrl.value)
  pendingAvatarPreviewUrl.value = ''
  profileOpen.value = true
}

const closeProfileEdit = () => {
  profileOpen.value = false
  pendingAvatarFile.value = null
  if (pendingAvatarPreviewUrl.value) URL.revokeObjectURL(pendingAvatarPreviewUrl.value)
  pendingAvatarPreviewUrl.value = ''
}

const pickAvatar = (e: Event) => {
  const file = pickFileFromChange(e)
  if (!file) return
  const msg = validateImageFile(file)
  if (msg) {
    toast.show(msg, 'error')
    return
  }
  pendingAvatarFile.value = file
  if (pendingAvatarPreviewUrl.value) URL.revokeObjectURL(pendingAvatarPreviewUrl.value)
  pendingAvatarPreviewUrl.value = URL.createObjectURL(file)
  toast.show('已选择头像，保存时上传', 'success')
}

const saveProfile = async () => {
  const nn = nickname.value.trim()
  const av = avatarUrl.value.trim()
  saving.value = true
  try {
    const profile = await updateMe({ nickname: nn, avatarUrl: av, file: pendingAvatarFile.value })
    auth.setUser(profile)
    applyProfile(profile)
    toast.show('已保存', 'success')
    closeProfileEdit()
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '保存失败', 'error')
  } finally {
    saving.value = false
  }
}

const openLedger = async () => {
  if (ledgerOpen.value) return
  ledgerOpen.value = true
  ledgerLoading.value = true
  try {
    const items = (await getPointsLedgers(0, 20)) as LedgerItem[]
    ledger.value = items
    ledgerHasMore.value = items.length >= 20
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      ledgerOpen.value = false
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  } finally {
    ledgerLoading.value = false
  }
}

const closeLedger = () => {
  ledgerOpen.value = false
}

const loadMoreLedger = async () => {
  if (!ledgerHasMore.value || ledgerMoreLoading.value) return
  ledgerMoreLoading.value = true
  try {
    const offset = ledger.value.length
    const items = (await getPointsLedgers(offset, 20)) as LedgerItem[]
    ledger.value = ledger.value.concat(items)
    ledgerHasMore.value = items.length >= 20
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  } finally {
    ledgerMoreLoading.value = false
  }
}

const goOrders = async () => {
  await router.push({ path: '/user/shop', query: { tab: 'orders' } })
}

onMounted(() => {
  applyProfile((auth.user as Record<string, unknown> | null) || null)
  void refresh()
})
</script>

<template>
  <div class="grid">
    <div class="row">
      <div class="title">个人中心</div>
      <div class="spacer" />
      <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
    </div>

    <!-- 用户基本信息 -->
    <div class="card profile-card">
      <div class="row">
        <img
          v-if="avatarUrl"
          :src="avatarUrl"
          alt="avatar"
          class="avatar avatar--lg"
        />
        <div v-else class="avatar avatar--lg placeholder-avatar">👤</div>
        
        <div class="profile-info">
          <div class="nickname">{{ nickname || '未设置昵称' }}</div>
          <div class="uid">UID: {{ auth.user?.id || '-' }}</div>
        </div>
      </div>
    </div>

    <!-- 菜单列表 -->
    <div class="menu-list card">
      <div class="menu-item" @click="openProfileEdit">
        <div class="menu-icon">📝</div>
        <div class="menu-label">编辑资料</div>
        <div class="menu-arrow">›</div>
      </div>
      <div class="menu-divider" />
      
      <div class="menu-item" @click="openLedger">
        <div class="menu-icon">💰</div>
        <div class="menu-label">积分流水</div>
        <div class="menu-arrow">›</div>
      </div>
      <div class="menu-divider" />
      
      <div class="menu-item" @click="goOrders">
        <div class="menu-icon">📦</div>
        <div class="menu-label">兑换记录</div>
        <div class="menu-arrow">›</div>
      </div>
    </div>

    <div class="card" style="margin-top: 12px">
       <div class="menu-item" style="color: var(--danger)" @click="logout">
        <div class="menu-icon">🚪</div>
        <div class="menu-label">退出登录</div>
      </div>
    </div>

    <!-- 编辑资料弹窗 -->
    <div v-if="profileOpen" class="modal" @click.self="closeProfileEdit">
      <div class="modal__panel card">
        <div class="title">编辑资料</div>
        <div class="grid" style="margin-top: 16px; gap: 12px">
          <div class="form-item">
            <label class="label">头像</label>
            <div class="row">
              <button class="btn btn--ghost" :disabled="saving" @click="triggerPick">选择头像</button>
              <input
                ref="avatarFileEl"
                style="display: none"
                type="file"
                accept="image/*"
                @change="pickAvatar"
              />
              <input v-model="avatarUrl" class="input" placeholder="可粘贴头像 URL（或选择本地图片）" />
            </div>
            <div v-if="pendingAvatarFile" class="help" style="margin-top: 6px">已选择：{{ pendingAvatarFile.name }}</div>
            <img
              v-if="pendingAvatarPreviewUrl || avatarUrl"
              :src="pendingAvatarPreviewUrl || avatarUrl"
              alt="avatar"
              style="width: 80px; height: 80px; object-fit: cover; border-radius: 999px; margin-top: 10px"
            />
          </div>
          <div class="form-item">
            <label class="label">昵称</label>
            <input v-model="nickname" class="input" placeholder="请输入昵称" />
          </div>
          <div class="row" style="margin-top: 8px">
            <button class="btn" :disabled="saving" @click="saveProfile">保存</button>
            <button class="btn btn--ghost" @click="closeProfileEdit">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 积分流水弹窗 -->
    <div v-if="ledgerOpen" class="modal" @click.self="closeLedger">
      <div class="modal__panel card">
        <div class="row">
          <div class="title">积分流水</div>
          <div class="spacer" />
          <button class="btn btn--ghost" @click="closeLedger">关闭</button>
        </div>
        <div class="grid" style="margin-top: 12px">
          <div v-if="ledgerLoading" class="muted">加载中…</div>
          <template v-else>
            <div v-if="ledger.length === 0" class="muted">暂无流水</div>
            <div v-for="it in ledger" :key="it.id || it.createdAt" class="card card--flat">
              <div class="row">
                <div>
                  <div class="title">{{ it.type || '记录' }}</div>
                  <div class="muted" style="margin-top: 6px">变化：{{ it.delta ?? '-' }} · 余额：{{ it.balanceAfter ?? '-' }}</div>
                  <div v-if="it.remark" class="muted" style="margin-top: 6px">{{ it.remark }}</div>
                  <div v-if="it.createdAt" class="muted" style="margin-top: 6px">{{ it.createdAt }}</div>
                </div>
                <div class="spacer" />
              </div>
            </div>
            <button v-if="ledgerHasMore" class="btn btn--ghost" :disabled="ledgerMoreLoading" @click="loadMoreLedger">
              {{ ledgerMoreLoading ? '加载中…' : '加载更多' }}
            </button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.avatar--lg { width: 64px; height: 64px; border-radius: 50%; object-fit: cover; }
.placeholder-avatar { display: grid; place-items: center; background: #eee; font-size: 24px; color: #aaa; }
.profile-info { margin-left: 16px; display: flex; flex-direction: column; justify-content: center; }
.nickname { font-size: 18px; font-weight: bold; }
.uid { font-size: 13px; color: #888; margin-top: 4px; }

.menu-list { padding: 0; overflow: hidden; }
.menu-item { display: flex; align-items: center; padding: 16px; cursor: pointer; transition: background 0.2s; }
.menu-item:active { background: #f5f5f5; }
.menu-icon { font-size: 20px; width: 32px; text-align: center; margin-right: 12px; }
.menu-label { flex: 1; font-size: 16px; }
.menu-arrow { color: #ccc; font-size: 18px; }
.menu-divider { height: 1px; background: #eee; margin: 0 16px; }

.modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  padding: 18px;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(8px);
}

.modal__panel {
  width: min(720px, calc(100vw - 36px));
  max-height: min(82vh, 720px);
  overflow: auto;
}

.label { font-size: 14px; font-weight: bold; margin-bottom: 4px; display: block; }
</style>
