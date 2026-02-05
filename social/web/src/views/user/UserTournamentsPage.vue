<script setup lang="ts">
import { onMounted, ref } from 'vue'

import {
  cancelTournamentJoin,
  getTournament,
  getTournamentResults,
  joinTournament,
  listJoinedTournaments,
  listTournaments,
} from '../../api'
import { useToastStore } from '../../stores/toast'

type TournamentItem = {
  id?: number
  title?: string
  content?: string
  coverUrl?: string
  status?: string
  joined?: boolean
  startAt?: string
  endAt?: string
  [k: string]: unknown
}

type ResultItem = {
  userId?: number
  rankNo?: number
  score?: number
  nickname?: string
  avatarUrl?: string
  [k: string]: unknown
}

const toast = useToastStore()

const loading = ref(false)
const items = ref<TournamentItem[]>([])

const joinedLoading = ref(false)
const joinedMoreLoading = ref(false)
const joined = ref<TournamentItem[]>([])
const joinedHasMore = ref(false)

const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<TournamentItem | null>(null)

const resultsLoading = ref(false)
const resultsMoreLoading = ref(false)
const results = ref<ResultItem[]>([])
const myRank = ref<ResultItem | null>(null)
const resultsHasMore = ref(false)

const refresh = async () => {
  loading.value = true
  try {
    items.value = (await listTournaments(0, 50)) as TournamentItem[]
  } catch (e) {
    const err = e as { message?: unknown }
    items.value = []
    if (err.message === 'unauthorized') return
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const refreshJoined = async (reset: boolean) => {
  if (joinedLoading.value) return
  if (!reset && (!joinedHasMore.value || joinedMoreLoading.value)) return

  if (reset) joinedLoading.value = true
  else joinedMoreLoading.value = true

  try {
    const offset = reset ? 0 : joined.value.length
    const list = (await listJoinedTournaments(offset, 50)) as TournamentItem[]
    joined.value = reset ? list : joined.value.concat(list)
    joinedHasMore.value = list.length >= 50
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      joined.value = []
      joinedHasMore.value = false
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '加载参赛记录失败', 'error')
  } finally {
    if (reset) joinedLoading.value = false
    else joinedMoreLoading.value = false
  }
}

const refreshDetailIfOpen = async (id: number) => {
  if (!detailOpen.value) return
  if (!detail.value || detail.value.id !== id) return
  detailLoading.value = true
  try {
    detail.value = (await getTournament(id)) as TournamentItem
  } catch {
  } finally {
    detailLoading.value = false
  }
  if (canShowResults(detail.value)) void loadResults(true)
  else {
    results.value = []
    myRank.value = null
    resultsHasMore.value = false
  }
}

const join = async (id: number) => {
  const it = (detail.value && detail.value.id === id ? detail.value : null) || items.value.find(x => x.id === id) || null
  if (!canJoin(it)) {
    toast.show(`当前状态：${phaseLabel(it)}，不可报名`, 'error')
    return
  }
  try {
    await joinTournament(id)
    toast.show('报名成功', 'success')
    await refresh()
    await refreshJoined(true)
    await refreshDetailIfOpen(id)
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '报名失败', 'error')
  }
}

const cancelJoin = async (id: number) => {
  const it = (detail.value && detail.value.id === id ? detail.value : null) || items.value.find(x => x.id === id) || null
  if (!canCancel(it)) {
    toast.show(`当前状态：${phaseLabel(it)}，不可取消报名`, 'error')
    return
  }
  const okConfirm = window.confirm('确认取消报名？')
  if (!okConfirm) return
  try {
    await cancelTournamentJoin(id)
    toast.show('已取消', 'success')
    await refresh()
    await refreshJoined(true)
    await refreshDetailIfOpen(id)
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '取消失败', 'error')
  }
}

const openDetail = async (id: number) => {
  detailOpen.value = true
  detail.value = null
  results.value = []
  myRank.value = null
  resultsHasMore.value = false

  detailLoading.value = true
  try {
    detail.value = (await getTournament(id)) as TournamentItem
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') return
    toast.show((typeof err.message === 'string' && err.message) || '加载详情失败', 'error')
  } finally {
    detailLoading.value = false
  }

  if (canShowResults(detail.value)) void loadResults(true)
}

const closeDetail = () => {
  detailOpen.value = false
  detail.value = null
  results.value = []
  myRank.value = null
  resultsHasMore.value = false
}

const loadResults = async (reset: boolean) => {
  const id = detail.value?.id
  if (!id) return
  if (resultsLoading.value) return
  if (!reset && (!resultsHasMore.value || resultsMoreLoading.value)) return

  if (reset) resultsLoading.value = true
  else resultsMoreLoading.value = true

  try {
    const offset = reset ? 0 : results.value.length
    const { items: list, my } = await getTournamentResults(id, offset, 50)
    const next = (list as ResultItem[]) || []
    results.value = reset ? next : results.value.concat(next)
    myRank.value = (my && typeof my === 'object' ? (my as ResultItem) : null) || null
    resultsHasMore.value = next.length >= 50
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '加载排名失败', 'error')
  } finally {
    if (reset) resultsLoading.value = false
    else resultsMoreLoading.value = false
  }
}

const statusUpper = (s?: unknown) => String(s || '').trim().toUpperCase()

type Phase = 'DRAFT' | 'SIGNUP' | 'RUNNING' | 'FINISHED' | 'CANCELED'

const parseMs = (s?: unknown) => {
  if (typeof s !== 'string') return NaN
  const t = Date.parse(s)
  return Number.isFinite(t) ? t : NaN
}

const phaseOf = (it?: { status?: unknown; startAt?: unknown; endAt?: unknown } | null): Phase => {
  const st = statusUpper(it?.status)
  if (st === 'CANCELED') return 'CANCELED'
  if (st === 'ENDED' || st === 'FINISHED') return 'FINISHED'
  if (st === 'DRAFT') return 'DRAFT'
  if (st === 'PUBLISHED') return 'SIGNUP'
  if (st === 'RUNNING' || st === 'IN_PROGRESS' || st === 'ONGOING' || st === 'STARTED') return 'RUNNING'

  const startAtMs = parseMs(it?.startAt)
  const endAtMs = parseMs(it?.endAt)
  const hasWindow = Number.isFinite(startAtMs) && Number.isFinite(endAtMs)

  if (hasWindow) {
    const now = Date.now()
    if (now < startAtMs) return 'SIGNUP'
    if (now >= startAtMs && now < endAtMs) return 'RUNNING'
    return 'FINISHED'
  }

  return 'DRAFT'
}

const phaseLabel = (it?: { status?: unknown; startAt?: unknown; endAt?: unknown } | null) => {
  const map: Record<Phase, string> = {
    DRAFT: '筹备中',
    SIGNUP: '报名中',
    RUNNING: '进行中',
    FINISHED: '已结束',
    CANCELED: '已取消',
  }
  return map[phaseOf(it)]
}

const canJoin = (it?: { status?: unknown } | null) => statusUpper(it?.status) === 'PUBLISHED'
const canCancel = (it?: { status?: unknown } | null) => statusUpper(it?.status) === 'PUBLISHED'
const canShowResults = (it?: { status?: unknown; startAt?: unknown; endAt?: unknown } | null) => phaseOf(it) === 'FINISHED'

const formatDate = (s?: string) => {
  if (!s) return '-'
  return s.replace('T', ' ').slice(0, 16)
}

onMounted(() => {
  void refresh()
  void refreshJoined(true)
})
</script>

<template>
  <div class="grid">
    <div class="row">
      <div class="title">赛事</div>
      <div class="spacer" />
      <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
    </div>

    <div class="card">
      <div class="row">
        <div class="title">参赛记录</div>
        <div class="spacer" />
        <button class="btn btn--ghost" :disabled="joinedLoading" @click="refreshJoined(true)">刷新</button>
      </div>
      <div v-if="joinedLoading" class="muted" style="margin-top: 10px">加载中…</div>
      <div v-else class="grid" style="margin-top: 10px">
        <div v-if="joined.length === 0" class="muted">暂无参赛记录</div>
        <div v-for="it in joined" :key="String(it.id) + '-joined'" class="card card--flat">
          <div class="row">
            <div>
              <div class="title">{{ it.title || `赛事#${it.id}` }}</div>
              <div class="muted" style="margin-top: 6px">
                {{ formatDate(it.startAt) }} ~ {{ formatDate(it.endAt) }} · 状态：{{ phaseLabel(it) }}
              </div>
            </div>
            <div class="spacer" />
            <button class="btn btn--ghost" :disabled="!it.id" @click="it.id && openDetail(it.id)">详情</button>
          </div>
        </div>
        <button
          v-if="joinedHasMore"
          class="btn btn--ghost"
          :disabled="joinedMoreLoading"
          @click="refreshJoined(false)"
        >
          {{ joinedMoreLoading ? '加载中…' : '加载更多' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="card muted">加载中…</div>
    <div v-else class="grid">
      <div v-if="items.length === 0" class="card muted">暂无赛事</div>
      <div v-for="it in items" :key="it.id || it.title" class="card card--flat">
        <div class="row">
          <div>
            <div class="row" style="gap: 8px">
              <div class="title">{{ it.title || `赛事#${it.id}` }}</div>
              <span v-if="it.joined" class="badge badge--success">已参加</span>
            </div>
            <div class="muted" style="margin-top: 6px">
              {{ formatDate(it.startAt) }} ~ {{ formatDate(it.endAt) }} · 状态：{{ phaseLabel(it) }}
            </div>
            <div v-if="it.content" class="muted" style="margin-top: 8px">{{ it.content }}</div>
          </div>
          <div class="spacer" />
          <button class="btn btn--ghost" :disabled="!it.id" @click="it.id && openDetail(it.id)">详情</button>
          <div v-if="it.joined" class="grid" style="gap: 8px; justify-items: end">
            <button class="btn btn--ghost" disabled>已参加</button>
            <button v-if="canCancel(it)" class="btn btn--danger" :disabled="!it.id" @click="it.id && cancelJoin(it.id)">
              取消报名
            </button>
          </div>
          <button v-else class="btn" :disabled="!it.id || !canJoin(it)" @click="it.id && join(it.id)">
            {{ canJoin(it) ? '报名' : phaseLabel(it) }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <div v-if="detailOpen" class="modal" @click.self="closeDetail">
    <div class="modal__panel card">
      <div class="row">
        <div class="title">赛事详情</div>
        <div class="spacer" />
        <button class="btn btn--ghost" @click="closeDetail">关闭</button>
      </div>

      <div v-if="detailLoading" class="muted" style="margin-top: 10px">加载中…</div>
      <template v-else>
        <div v-if="!detail" class="muted" style="margin-top: 10px">暂无详情</div>
        <template v-else>
          <div class="grid" style="margin-top: 10px">
            <img v-if="detail.coverUrl" class="t-cover" :src="detail.coverUrl" alt="cover" />
            <div class="row" style="gap: 8px; flex-wrap: wrap">
              <div class="title" style="min-width: 0">{{ detail.title || `赛事#${detail.id}` }}</div>
              <span class="badge badge--muted">{{ phaseLabel(detail) }}</span>
              <span v-if="detail.joined" class="badge badge--success">已参加</span>
            </div>
            <div class="muted">
              {{ formatDate(detail.startAt) }} ~ {{ formatDate(detail.endAt) }}
            </div>
            <div v-if="detail.content" class="help" style="white-space: pre-wrap">{{ detail.content }}</div>

            <div class="row" style="gap: 10px; flex-wrap: wrap">
              <button
                v-if="!detail.joined"
                class="btn"
                :disabled="!detail.id || !canJoin(detail)"
                @click="detail.id && join(detail.id)"
              >
                {{ canJoin(detail) ? '报名参加' : phaseLabel(detail) }}
              </button>
              <button
                v-else-if="canCancel(detail)"
                class="btn btn--danger"
                :disabled="!detail.id"
                @click="detail.id && cancelJoin(detail.id)"
              >
                取消报名
              </button>
              <button v-else class="btn btn--ghost" disabled>已参加</button>
              <button class="btn btn--ghost" :disabled="!detail.id" @click="detail.id && openDetail(detail.id)">刷新详情</button>
            </div>

            <div class="card card--flat">
              <div class="row">
                <div class="title">排名</div>
                <div class="spacer" />
                <button
                  class="btn btn--ghost"
                  :disabled="resultsLoading || !detail.id || !canShowResults(detail)"
                  @click="loadResults(true)"
                >
                  刷新
                </button>
              </div>

              <div v-if="!canShowResults(detail)" class="muted" style="margin-top: 10px">
                当前阶段为 {{ phaseLabel(detail) }}，暂无排名。
              </div>
              <div v-else-if="resultsLoading" class="muted" style="margin-top: 10px">加载中…</div>
              <template v-else>
                <div v-if="myRank" class="row" style="gap: 10px; margin-top: 10px; flex-wrap: wrap">
                  <span class="badge badge--primary">我的名次：{{ myRank.rankNo ?? '-' }}</span>
                  <span class="badge badge--muted">分数：{{ myRank.score ?? 0 }}</span>
                </div>

                <div class="grid" style="margin-top: 10px">
                  <div v-if="results.length === 0" class="muted">暂无排名</div>
                  <div v-for="r in results" :key="String(r.userId) + '-' + String(r.rankNo)" class="card card--flat">
                    <div class="row" style="align-items: center; gap: 10px">
                      <img v-if="r.avatarUrl" class="r-avatar" :src="r.avatarUrl" alt="avatar" />
                      <div class="grid" style="gap: 4px; min-width: 0">
                        <div class="row" style="gap: 8px; flex-wrap: wrap">
                          <span class="badge badge--primary">#{{ r.rankNo ?? '-' }}</span>
                          <div class="title" style="font-size: 14px; min-width: 0">
                            {{ r.nickname || `用户#${r.userId}` }}
                          </div>
                        </div>
                        <div class="muted">分数：{{ r.score ?? 0 }}</div>
                      </div>
                      <div class="spacer" />
                    </div>
                  </div>

                  <button
                    v-if="resultsHasMore"
                    class="btn btn--ghost"
                    :disabled="resultsMoreLoading"
                    @click="loadResults(false)"
                  >
                    {{ resultsMoreLoading ? '加载中…' : '加载更多' }}
                  </button>
                </div>
              </template>
            </div>
          </div>
        </template>
      </template>
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
  width: min(860px, calc(100vw - 36px));
  max-height: min(86vh, 860px);
  overflow: auto;
}

.t-cover {
  width: 100%;
  aspect-ratio: 16 / 9;
  object-fit: cover;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.04);
}

.r-avatar {
  width: 36px;
  height: 36px;
  border-radius: 12px;
  object-fit: cover;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.04);
}
</style>
