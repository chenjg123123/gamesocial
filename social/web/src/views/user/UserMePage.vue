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
const ledgerOpen = ref(false)
const ledgerLoading = ref(false)
const ledgerMoreLoading = ref(false)
const ledger = ref<LedgerItem[]>([])
const ledgerHasMore = ref(false)

const logout = () => {
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

const saveProfile = async () => {
  const nn = nickname.value.trim()
  const av = avatarUrl.value.trim()
  saving.value = true
  try {
    const profile = await updateMe({ nickname: nn, avatarUrl: av })
    auth.setUser(profile)
    applyProfile(profile)
    toast.show('已保存', 'success')
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
      <div class="title">我的</div>
      <div class="spacer" />
      <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
    </div>

    <div class="card">
      <div class="row">
        <div class="title">账号</div>
        <div class="spacer" />
        <button class="btn btn--ghost" @click="logout">退出登录</button>
      </div>
      <div class="help" style="margin-top: 6px">资料、积分流水、兑换记录都从这里进入。</div>
      <div class="row" style="margin-top: 10px">
        <button class="btn btn--ghost" @click="openLedger">积分流水</button>
        <button class="btn btn--ghost" @click="goOrders">兑换记录</button>
      </div>
    </div>

    <div class="card">
      <div class="title">资料</div>
      <div class="grid" style="margin-top: 10px">
        <div class="row" style="align-items: flex-start">
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            alt="avatar"
            class="avatar"
          />
          <div class="spacer" />
        </div>
        <input v-model="nickname" class="input" placeholder="昵称" />
        <input v-model="avatarUrl" class="input" placeholder="头像 URL" />
        <button class="btn" :disabled="saving" @click="saveProfile">保存资料</button>
      </div>
    </div>

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
</style>
