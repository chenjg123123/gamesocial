import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'info'

export const useToastStore = defineStore('toast', () => {
  const current = ref<{ message: string; type: ToastType } | null>(null)
  let timer: number | undefined

  const show = (message: string, type: ToastType = 'info') => {
    current.value = { message, type }
    if (timer) window.clearTimeout(timer)
    timer = window.setTimeout(() => {
      current.value = null
    }, 2200)
  }

  const clear = () => {
    if (timer) window.clearTimeout(timer)
    current.value = null
  }

  return { current, show, clear }
})

