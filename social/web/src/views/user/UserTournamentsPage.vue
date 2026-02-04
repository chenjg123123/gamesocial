<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { cancelTournamentJoin, joinTournament, listTournaments } from '../../api'
import { useToastStore } from '../../stores/toast'

type TournamentItem = {
  id?: number
  title?: string
  content?: string
  status?: string
  joined?: boolean
  startAt?: string
  endAt?: string
  [k: string]: unknown
}

const toast = useToastStore()

const loading = ref(false)
const items = ref<TournamentItem[]>([])

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

const join = async (id: number) => {
  try {
    await joinTournament(id)
    toast.show('报名成功', 'success')
    await refresh()
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
  try {
    await cancelTournamentJoin(id)
    toast.show('已取消', 'success')
    await refresh()
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '取消失败', 'error')
  }
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="grid">
    <div class="row">
      <div class="title">赛事</div>
      <div class="spacer" />
      <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
    </div>

    <div v-if="loading" class="card muted">加载中…</div>
    <div v-else class="grid">
      <div v-if="items.length === 0" class="card muted">暂无赛事</div>
      <div v-for="it in items" :key="it.id || it.title" class="card card--flat">
        <div class="row">
          <div>
            <div class="title">{{ it.title || `赛事#${it.id}` }}</div>
            <div class="muted" style="margin-top: 6px">
              {{ it.startAt || '-' }} ~ {{ it.endAt || '-' }} · 状态：{{ it.status || '-' }}
            </div>
            <div v-if="it.content" class="muted" style="margin-top: 8px">{{ it.content }}</div>
          </div>
          <div class="spacer" />
          <button v-if="it.joined" class="btn btn--danger" :disabled="!it.id" @click="it.id && cancelJoin(it.id)">
            取消报名
          </button>
          <button v-else class="btn" :disabled="!it.id" @click="it.id && join(it.id)">报名</button>
        </div>
      </div>
    </div>
  </div>
</template>
