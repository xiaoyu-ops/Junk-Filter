<template>
  <header class="w-full px-8 py-4 flex items-center justify-between sticky top-0 z-50 bg-white dark:bg-[#0f0f11] backdrop-blur-sm border-b border-gray-100 dark:border-gray-800 transition-colors duration-200">
    <!-- Logo -->
    <div class="flex items-center gap-3">
      <span class="material-icons-outlined text-3xl text-gray-900 dark:text-gray-100">delete_outline</span>
      <h1 class="text-xl font-bold tracking-tight text-gray-900 dark:text-gray-100">Junk Filter</h1>
    </div>

    <!-- Navigation Links -->
    <nav class="flex items-center space-x-6">
      <RouterLink
        to="/"
        :class="['nav-link', { active: currentRoute === '/' }]"
      >
        主页
      </RouterLink>
      <RouterLink
        to="/timeline"
        :class="['nav-link', { active: currentRoute === '/timeline' }]"
      >
        时间轴
      </RouterLink>
      <RouterLink
        to="/config"
        :class="['nav-link', { active: currentRoute === '/config' }]"
      >
        配置中心
      </RouterLink>
      <RouterLink
        to="/task"
        :class="['nav-link', { active: currentRoute === '/task' }]"
      >
        分发任务
      </RouterLink>
    </nav>

    <!-- Theme Toggle -->
    <button
      @click="themeStore.toggleTheme"
      class="p-2 rounded-full text-gray-500 hover:bg-gray-100 dark:hover:bg-[#27272a] dark:text-gray-400 transition-colors"
      :title="`切换到${themeStore.isDark ? '亮色' : '深色'}模式`"
    >
      <span v-if="!themeStore.isDark" class="material-icons-outlined">dark_mode</span>
      <span v-else class="material-icons-outlined">light_mode</span>
    </button>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from '@/stores'

const router = useRouter()
const themeStore = useThemeStore()

const currentRoute = computed(() => router.currentRoute.value.path)
</script>
