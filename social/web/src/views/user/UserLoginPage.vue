<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getToken } from '../../lib/storage'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'

const auth = useAuthStore()
const toast = useToastStore()
const route = useRoute()
const router = useRouter()

const openIdDraft = ref('')
const loading = ref(false)

const redirectTo = computed(() => {
  const q = route.query.redirect
  const raw = Array.isArray(q) ? q[0] : q
  const next = typeof raw === 'string' ? raw : ''
  return next || '/user/index'
})

onMounted(() => {
  const token = getToken()
  if (token) void router.replace(redirectTo.value)
})

const login = async () => {
  loading.value = true
  try {
    const fn = (auth as unknown as { loginWithOpenId?: unknown }).loginWithOpenId
    const alt = (auth as unknown as { loginWithWechatCode?: unknown }).loginWithWechatCode
    const loginFn = (typeof fn === 'function' ? fn : null) || (typeof alt === 'function' ? alt : null)
    if (!loginFn) throw new Error('登录方法缺失')
    await loginFn(openIdDraft.value)
    toast.show('登录成功', 'success')
    await router.replace(redirectTo.value)
  } catch (e) {
    const err = e as { message?: unknown }
    toast.show((typeof err.message === 'string' && err.message) || '登录失败', 'error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <div class="container">
      <div class="grid" style="max-width: 520px; margin: 0 auto; padding-top: 18vh">
        <div class="card">
          <div class="title">登录</div>
          <div class="help" style="margin-top: 8px">当前临时方案：输入 openid，走 /api/auth/wechat/login 换取 token。</div>
          <div class="grid" style="margin-top: 12px">
            <input v-model="openIdDraft" class="input" placeholder="请输入 openid" />
            <button class="btn" :disabled="loading" @click="login">{{ loading ? '登录中…' : '登录' }}</button>
          </div>
        </div>
        <div class="card card--flat muted" style="box-shadow: none">
          登录后会自动跳回你刚才访问的页面。
        </div>
      </div>
    </div>
  </div>
</template>
