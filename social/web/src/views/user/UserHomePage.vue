<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getMe, getPointsBalance, getVipStatus, listTournaments, getTasks, listGoods } from '../../api'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'

const auth = useAuthStore()
const toast = useToastStore()
const router = useRouter()

const loading = ref(false)
const vipLabel = ref('普通用户')
const vipExpireText = ref('')
const pointsBalance = ref<number>(0)

const activeTournament = ref<Record<string, unknown> | null>(null)
const dailyTask = ref<Record<string, unknown> | null>(null)
const featuredGoods = ref<Array<Record<string, unknown>>>([])

const avatarSrc = computed(() => {
  const v = auth.user?.avatarUrl
  const s = v === undefined || v === null ? '' : String(v)
  return s.trim().replace(/`/g, '')
})

const levelText = computed(() => {
  const raw = auth.user && typeof auth.user === 'object' ? (auth.user as { level?: unknown }).level : undefined
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? String(n) : '1'
})

const expText = computed(() => {
  const raw = auth.user && typeof auth.user === 'object' ? (auth.user as { exp?: unknown }).exp : undefined
  const n = Number(raw)
  return Number.isFinite(n) && n >= 0 ? String(n) : '0'
})

const formatDate = (s?: string) => {
  if (!s) return ''
  return s.replace('T', ' ').slice(0, 16)
}

const safeStr = (v: unknown) => (v === undefined || v === null ? '' : String(v))

const refresh = async () => {
  loading.value = true
  try {
    const [profile, vip, points, tournaments, tasks, goods] = await Promise.all([
      getMe().catch(() => null),
      getVipStatus().catch(() => null),
      getPointsBalance().catch(() => null),
      listTournaments(0, 5).catch(() => []),
      getTasks().catch(() => []),
      listGoods(0, 4).catch(() => []),
    ])

    if (profile) auth.setUser(profile)

    let isVip = false
    let expireAt = ''
    if (vip && typeof vip === 'object') {
      const v = vip as { isVip?: unknown; active?: unknown; expireAt?: unknown }
      isVip = v.isVip === true || v.active === true
      expireAt = typeof v.expireAt === 'string' ? v.expireAt : ''
    }
    vipLabel.value = isVip ? 'VIP' : '普通用户'
    vipExpireText.value = isVip && expireAt ? formatDate(expireAt) : ''

    const bal = points && typeof points.balance === 'number' ? points.balance : 0
    pointsBalance.value = bal

    // 筛选进行中/报名中的赛事
    const tList = (tournaments as Array<Record<string, unknown>>) || []
    activeTournament.value = tList.find(t => {
      const s = String(t.status || '').toUpperCase()
      return s === 'PUBLISHED' || s === 'RUNNING'
    }) || null

    // 筛选每日任务
    const tkList = (tasks as Array<Record<string, unknown>>) || []
    dailyTask.value = tkList.find(t => String(t.periodType || '').toUpperCase() === 'DAILY') || null

    // 推荐商品
    featuredGoods.value = (goods as Array<Record<string, unknown>>) || []

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
    <!-- 用户信息卡片 -->
    <div class="card bg-gradient">
      <div class="hero">
        <img v-if="avatarSrc" class="avatar avatar--lg" :src="avatarSrc" alt="avatar" />
        <div v-else class="avatar avatar--lg placeholder-avatar">👤</div>
        
        <div class="hero-content">
          <div class="title text-white">Hi, {{ auth.user?.nickname || '未登录用户' }}</div>
          <div class="chips">
            <div class="chip chip--glass">Lv.{{ levelText }}</div>
            <div v-if="expText !== '0'" class="chip chip--glass">Exp {{ expText }}</div>
            <div class="chip chip--glass">{{ vipLabel }}</div>
          </div>
          <div class="points-display">
            <span class="points-val">{{ pointsBalance }}</span>
            <span class="points-label">积分</span>
          </div>
        </div>
        
        <div class="spacer" />
        <button class="btn btn--icon" @click="refresh">↻</button>
      </div>
      
      <div v-if="vipExpireText" class="vip-expire">
        会员有效期至 {{ vipExpireText }}
      </div>
    </div>

    <!-- 快捷入口 -->
    <div class="quick-actions">
      <RouterLink to="/user/tasks" class="quick-action">
        <div class="qa-icon" style="background: #e0f2fe; color: #0284c7">✅</div>
        <span>签到任务</span>
      </RouterLink>
      <RouterLink to="/user/tournaments" class="quick-action">
        <div class="qa-icon" style="background: #fef3c7; color: #d97706">🏆</div>
        <span>热门赛事</span>
      </RouterLink>
      <RouterLink to="/user/shop" class="quick-action">
        <div class="qa-icon" style="background: #fce7f3; color: #db2777">🛍️</div>
        <span>积分商城</span>
      </RouterLink>
      <RouterLink to="/user/me" class="quick-action">
        <div class="qa-icon" style="background: #f3f4f6; color: #4b5563">⚙️</div>
        <span>个人设置</span>
      </RouterLink>
    </div>

    <!-- 推荐赛事 -->
    <div v-if="activeTournament" class="section">
      <div class="section-header">
        <div class="section-title">热门赛事</div>
        <RouterLink to="/user/tournaments" class="link-more">更多</RouterLink>
      </div>
      <div class="card card--flat tournament-card" @click="router.push('/user/tournaments')">
        <img v-if="activeTournament.coverUrl" :src="String(activeTournament.coverUrl)" class="t-cover" />
        <div class="t-info">
          <div class="title">{{ activeTournament.title }}</div>
          <div class="muted">{{ formatDate(String(activeTournament.startAt)) }} 开赛</div>
          <span class="badge badge--primary" style="align-self: flex-start; margin-top: 4px">报名中</span>
        </div>
      </div>
    </div>

    <!-- 每日任务 -->
    <div v-if="dailyTask" class="section">
      <div class="section-header">
        <div class="section-title">每日任务</div>
        <RouterLink to="/user/tasks" class="link-more">全部</RouterLink>
      </div>
      <div class="card card--flat task-card">
        <div class="row">
          <div>
            <div class="title">{{ dailyTask.name }}</div>
            <div class="muted">进度：{{ dailyTask.progress || 0 }}/{{ dailyTask.targetCount || 1 }}</div>
          </div>
          <div class="spacer" />
          <RouterLink to="/user/tasks" class="btn btn--sm">去完成</RouterLink>
        </div>
      </div>
    </div>

    <!-- 推荐商品 -->
    <div v-if="featuredGoods.length > 0" class="section">
      <div class="section-header">
        <div class="section-title">精选兑换</div>
        <RouterLink to="/user/shop" class="link-more">更多</RouterLink>
      </div>
      <div class="goods-grid">
        <div v-for="g in featuredGoods" :key="safeStr(g.id)" class="card card--flat goods-item" @click="router.push('/user/shop')">
          <div class="goods-img-box">
             <img v-if="g.coverUrl" :src="String(g.coverUrl)" class="goods-img" />
             <div v-else class="goods-placeholder">🎁</div>
          </div>
          <div class="goods-name">{{ g.name }}</div>
          <div class="goods-price">{{ g.pointsPrice }} 积分</div>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
.bg-gradient {
  background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
  color: white;
}
.text-white { color: white; }
.hero { display: flex; align-items: center; gap: 16px; }
.hero-content { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.avatar--lg { width: 64px; height: 64px; border-radius: 50%; border: 2px solid rgba(255,255,255,0.2); }
.placeholder-avatar { display: grid; place-items: center; background: rgba(255,255,255,0.1); font-size: 24px; }
.chip--glass { background: rgba(255,255,255,0.2); border: none; color: white; }
.points-display { display: flex; align-items: baseline; gap: 4px; margin-top: 8px; }
.points-val { font-size: 24px; font-weight: bold; }
.points-label { font-size: 12px; opacity: 0.8; }
.vip-expire { margin-top: 12px; font-size: 12px; opacity: 0.7; text-align: right; }
.btn--icon { background: rgba(255,255,255,0.1); color: white; width: 32px; height: 32px; padding: 0; border-radius: 50%; }

.quick-actions { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-top: 8px; }
.quick-action { display: flex; flex-direction: column; align-items: center; gap: 8px; text-decoration: none; color: inherit; font-size: 12px; }
.qa-icon { width: 48px; height: 48px; border-radius: 16px; display: grid; place-items: center; font-size: 20px; }

.section { margin-top: 24px; display: flex; flex-direction: column; gap: 12px; }
.section-header { display: flex; align-items: center; justify-content: space-between; }
.section-title { font-size: 18px; font-weight: bold; }
.link-more { color: var(--primary); text-decoration: none; font-size: 14px; }

.tournament-card { display: flex; gap: 12px; padding: 12px; cursor: pointer; }
.t-cover { width: 80px; height: 60px; border-radius: 8px; object-fit: cover; background: #eee; }
.t-info { display: flex; flex-direction: column; justify-content: center; }

.goods-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.goods-item { padding: 8px; cursor: pointer; }
.goods-img-box { width: 100%; aspect-ratio: 1; border-radius: 8px; overflow: hidden; background: #f9fafb; display: grid; place-items: center; }
.goods-img { width: 100%; height: 100%; object-fit: cover; }
.goods-placeholder { font-size: 32px; }
.goods-name { font-size: 14px; margin-top: 8px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.goods-price { font-size: 14px; color: var(--primary); font-weight: bold; margin-top: 4px; }
</style>
