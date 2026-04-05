<template>
  <div class="relative">
    <!-- Bell Button -->
    <button
      @click="toggle"
      class="relative p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
    >
      <span class="material-symbols-outlined text-xl">notifications</span>
      <span
        v-if="unreadCount > 0"
        class="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] flex items-center justify-center px-1 bg-red-500 text-white text-[10px] font-bold rounded-full"
      >{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>

    <!-- Dropdown -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 scale-95 -translate-y-1"
      leave-active-class="transition-all duration-150 ease-in"
      leave-to-class="opacity-0 scale-95 -translate-y-1"
    >
      <div
        v-if="isOpen"
        class="absolute right-0 top-full mt-2 w-96 max-h-[480px] bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 shadow-lg overflow-hidden z-50"
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">Notifications</h3>
          <button
            v-if="unreadCount > 0"
            @click="markAllAsRead"
            class="text-xs text-blue-600 dark:text-blue-400 hover:underline"
          >Mark all read</button>
        </div>

        <!-- List -->
        <div class="overflow-y-auto max-h-[400px] custom-scrollbar">
          <div v-if="notifications.length === 0" class="p-6 text-center text-sm text-gray-400">
            No notifications yet
          </div>

          <div
            v-for="n in notifications"
            :key="n.id"
            @click="handleClick(n)"
            :class="[
              'px-4 py-3 border-b border-gray-50 dark:border-gray-800 cursor-pointer transition-colors',
              n.is_read
                ? 'hover:bg-gray-50 dark:hover:bg-gray-800/50'
                : 'bg-blue-50/50 dark:bg-blue-900/10 hover:bg-blue-50 dark:hover:bg-blue-900/20'
            ]"
          >
            <div class="flex items-start gap-3">
              <span
                :class="[
                  'material-symbols-outlined text-lg mt-0.5 shrink-0',
                  n.decision === 'INTERESTING'
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-amber-600 dark:text-amber-400'
                ]"
              >{{ n.decision === 'INTERESTING' ? 'star' : 'bookmark' }}</span>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-gray-900 dark:text-white line-clamp-1">{{ n.title }}</p>
                <p v-if="n.summary" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">{{ n.summary }}</p>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-[10px] text-blue-600 dark:text-blue-400 font-medium">Innovation {{ n.innovation_score }}/10</span>
                  <span class="text-[10px] text-green-600 dark:text-green-400 font-medium">Depth {{ n.depth_score }}/10</span>
                  <span class="text-[10px] text-gray-400 ml-auto">{{ formatTime(n.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const { notifications, unreadCount, isOpen, loadNotifications, markAsRead, markAllAsRead, connectSSE, disconnectSSE, toggle } = useNotification()

const formatTime = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = Math.floor((now - d) / 60000)
  if (diff < 1) return 'just now'
  if (diff < 60) return `${diff}m`
  if (diff < 1440) return `${Math.floor(diff / 60)}h`
  return d.toLocaleDateString()
}

const handleClick = (n) => {
  if (!n.is_read) markAsRead(n.id)
  isOpen.value = false
  router.push(`/reader/${n.content_id}`)
}

// Close on outside click
const handleOutsideClick = (e) => {
  if (isOpen.value && !e.target.closest('.relative')) {
    isOpen.value = false
  }
}

onMounted(() => {
  loadNotifications()
  connectSSE()
  document.addEventListener('click', handleOutsideClick)
})

onUnmounted(() => {
  disconnectSSE()
  document.removeEventListener('click', handleOutsideClick)
})
</script>
