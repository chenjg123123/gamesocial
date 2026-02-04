<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const active = computed(() => {
  const p = route.path || ''
  if (p.startsWith('/user/index')) return 'index'
  if (p.startsWith('/user/tasks')) return 'tasks'
  if (p.startsWith('/user/shop')) return 'shop'
  if (p.startsWith('/user/tournaments')) return 'tournaments'
  if (p.startsWith('/user/me')) return 'me'
  return 'index'
})
</script>

<template>
  <div class="topbar">
    <div class="topbar__inner container">
      <div class="title">GameSocial</div>
    </div>
  </div>

  <div class="page">
    <div class="container">
      <RouterView v-slot="{ Component }">
        <component :is="Component" :key="route.fullPath" />
      </RouterView>
    </div>
  </div>

  <div class="bottombar">
    <div class="bottombar__inner">
      <RouterLink class="navitem" :class="{ 'navitem--active': active === 'index' }" to="/user/index">
        <div class="navitem__icon">🏠</div>
        <div class="navitem__label">首页</div>
      </RouterLink>
      <RouterLink class="navitem" :class="{ 'navitem--active': active === 'tasks' }" to="/user/tasks">
        <div class="navitem__icon">✅</div>
        <div class="navitem__label">任务</div>
      </RouterLink>
      <RouterLink class="navitem" :class="{ 'navitem--active': active === 'shop' }" to="/user/shop">
        <div class="navitem__icon">🛍️</div>
        <div class="navitem__label">商城</div>
      </RouterLink>
      <RouterLink class="navitem" :class="{ 'navitem--active': active === 'tournaments' }" to="/user/tournaments">
        <div class="navitem__icon">🏆</div>
        <div class="navitem__label">赛事</div>
      </RouterLink>
      <RouterLink class="navitem" :class="{ 'navitem--active': active === 'me' }" to="/user/me">
        <div class="navitem__icon">👤</div>
        <div class="navitem__label">我的</div>
      </RouterLink>
    </div>
  </div>
</template>
