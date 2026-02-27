import { defineStore } from 'pinia'
import { ref, computed, watch, onMounted } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // 状态
  const isDark = ref(false)

  // 计算属性
  const theme = computed(() => isDark.value ? 'dark' : 'light')

  // 初始化主题
  const initTheme = () => {
    const saved = localStorage.getItem('theme')
    if (saved) {
      isDark.value = saved === 'dark'
    } else {
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    updateDOM()
  }

  // 更新DOM和localStorage
  const updateDOM = () => {
    const html = document.documentElement
    if (isDark.value) {
      html.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      html.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }

  // 切换主题
  const toggleTheme = () => {
    isDark.value = !isDark.value
  }

  // 监听主题变化
  watch(isDark, () => {
    updateDOM()
  })

  return {
    isDark,
    theme,
    initTheme,
    toggleTheme,
  }
})
