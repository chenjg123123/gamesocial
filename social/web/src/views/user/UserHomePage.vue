<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { getMe, getPointsBalance, getVipStatus } from '../../api'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'

const auth = useAuthStore()
const toast = useToastStore()

const loading = ref(false)
const vipLabel = ref('普通用户')
const pointsBalance = ref<number>(0)

const refresh = async () => {
  loading.value = true
  try {
    const [profile, vip, points] = await Promise.all([
      getMe().catch(() => null),
      getVipStatus().catch(() => null),
      getPointsBalance().catch(() => null),
    ])

    if (profile) auth.setUser(profile)

    let label = '普通用户'
    if (vip && vip.active) {
      const plan = typeof vip.plan === 'string' ? vip.plan : ''
      label = (`VIP ${plan}`).trim()
    }
    vipLabel.value = label

    const bal = points && typeof points.balance === 'number' ? points.balance : 0
    pointsBalance.value = bal
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="grid">
    <div class="card">
      <div class="hero">
        <img v-if="auth.user?.avatarUrl" class="avatar" :src="String(auth.user.avatarUrl)" alt="avatar" />
        <div style="min-width: 0">
          <div class="title">你好，{{ auth.user?.nickname || '游客' }}</div>
          <div class="muted" style="margin-top: 6px">欢迎来到 GameSocial</div>
          <div class="chips">
            <div class="chip">身份：{{ vipLabel }}</div>
            <div class="chip">积分：{{ pointsBalance }}</div>
          </div>
        </div>
        <div class="spacer" />
        <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
      </div>
    </div>
  </div>
</template>
