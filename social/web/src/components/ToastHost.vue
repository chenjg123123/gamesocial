<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useToastStore } from '../stores/toast'

const toast = useToastStore()
const { current } = storeToRefs(toast)

const klass = computed(() => {
  const type = current.value?.type || 'info'
  return ['toast', `toast--${type}`]
})
</script>

<template>
  <div v-if="current" :class="klass" @click="toast.clear()">
    {{ current.message }}
  </div>
</template>

<style scoped>
.toast {
  position: fixed;
  left: 50%;
  bottom: 84px;
  transform: translateX(-50%);
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(17, 24, 39, 0.92);
  color: #fff;
  font-size: 14px;
  max-width: calc(100vw - 32px);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  z-index: 9999;
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.25);
}

.toast--success {
  background: rgba(5, 150, 105, 0.95);
}

.toast--error {
  background: rgba(220, 38, 38, 0.95);
}
</style>

