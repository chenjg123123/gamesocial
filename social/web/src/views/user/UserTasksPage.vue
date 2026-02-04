<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { getTasks, taskCheckin, taskClaim } from '../../api'
import { useToastStore } from '../../stores/toast'

type TaskItem = {
  taskCode?: string
  name?: string
  progress?: number
  targetCount?: number
  claimed?: boolean
  rewardPoints?: number
  [k: string]: unknown
}

const toast = useToastStore()

const loading = ref(false)
const checkinLoading = ref(false)
const items = ref<TaskItem[]>([])

const refresh = async () => {
  loading.value = true
  try {
    items.value = (await getTasks()) as TaskItem[]
  } catch (e) {
    const err = e as { message?: unknown }
    items.value = []
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '加载失败', 'error')
  } finally {
    loading.value = false
  }
}

const doCheckin = async () => {
  checkinLoading.value = true
  try {
    await taskCheckin()
    toast.show('打卡成功', 'success')
    await refresh()
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '打卡失败', 'error')
  } finally {
    checkinLoading.value = false
  }
}

const claim = async (taskCode: string) => {
  try {
    await taskClaim(taskCode)
    toast.show('领取成功', 'success')
    await refresh()
  } catch (e) {
    const err = e as { message?: unknown }
    if (err.message === 'unauthorized') {
      toast.show('请先登录', 'error')
      return
    }
    toast.show((typeof err.message === 'string' && err.message) || '领取失败', 'error')
  }
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="grid">
    <div class="row">
      <div class="title">任务中心</div>
      <div class="spacer" />
      <button class="btn btn--ghost" :disabled="loading" @click="refresh">刷新</button>
      <button class="btn" :disabled="checkinLoading" @click="doCheckin">打卡</button>
    </div>

    <div v-if="loading" class="card muted">加载中…</div>

    <div v-else class="grid">
      <div v-if="items.length === 0" class="card muted">暂无任务</div>
      <div v-for="it in items" :key="it.taskCode || it.name" class="card card--flat">
        <div class="row">
          <div>
            <div class="title">{{ it.name || it.taskCode || '任务' }}</div>
            <div class="muted" style="margin-top: 6px">
              进度：{{ it.progress ?? 0 }}/{{ it.targetCount ?? 0 }} · 奖励：{{ it.rewardPoints ?? '-' }}
            </div>
          </div>
          <div class="spacer" />
          <button
            class="btn"
            :disabled="!it.taskCode || it.claimed || (it.targetCount ?? 0) > (it.progress ?? 0)"
            @click="it.taskCode && claim(it.taskCode)"
          >
            {{ it.claimed ? '已领取' : '领取' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
