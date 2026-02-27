<template>
  <div id="app" class="min-h-screen flex flex-col bg-surface-light dark:bg-[#0f0f11] transition-colors duration-200">
    <!-- 导航条 -->
    <RouterView name="navbar" />

    <!-- 页面内容 -->
    <main class="flex-1">
      <RouterView />
    </main>

    <!-- Toast容器 -->
    <div class="fixed top-4 right-4 z-50 space-y-2 pointer-events-none">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast pointer-events-auto', toast.type]"
      >
        {{ toast.message }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useToast } from '@/composables/useToast'
import { useThemeStore } from '@/stores'

const { toasts } = useToast()
const themeStore = useThemeStore()

// 初始化主题
onMounted(() => {
  themeStore.initTheme()
})
</script>

<style scoped>
</style>
