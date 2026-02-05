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
  <div class="login-page">
    <div class="login-card card">
      <div class="brand">GameSocial</div>
      <div class="title">欢迎回来</div>
      <div class="subtitle">请输入您的账号凭证以登录</div>
      
      <div class="form-group">
        <input 
          v-model="openIdDraft" 
          class="input" 
          placeholder="请输入 OpenID / 账号凭证" 
          @keyup.enter="login"
        />
      </div>
      
      <button class="btn btn--primary btn--block" :disabled="loading" @click="login">
        {{ loading ? '登录中…' : '立即登录' }}
      </button>
      
      <div class="footer-tip">
        测试阶段直接输入 OpenID 即可登录
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  background: #f3f4f6;
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 32px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.brand {
  font-size: 24px;
  font-weight: 800;
  color: var(--primary);
  text-align: center;
  margin-bottom: 24px;
}

.title {
  font-size: 20px;
  font-weight: bold;
  text-align: center;
  color: #111827;
}

.subtitle {
  text-align: center;
  color: #6b7280;
  font-size: 14px;
  margin-top: 8px;
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.input {
  width: 100%;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  outline: none;
  transition: border-color 0.2s;
}

.input:focus {
  border-color: var(--primary);
}

.btn--block {
  width: 100%;
  padding: 12px;
  font-size: 16px;
}

.footer-tip {
  margin-top: 24px;
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
}
</style>
