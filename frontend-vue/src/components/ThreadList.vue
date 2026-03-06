<template>
  <aside class="w-64 flex flex-col bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 shadow-sm overflow-hidden">
    <!-- 标题栏 -->
    <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
      <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">子对话</h3>
      <button
        @click="handleCreateThread"
        :disabled="!taskStore.selectedTaskId"
        class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        title="新建子对话"
      >
        <span class="material-icons-outlined text-lg">add</span>
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="taskStore.isLoadingThreads" class="flex-1 p-4">
      <div v-for="i in 3" :key="i" class="h-12 bg-gray-100 dark:bg-gray-800 rounded-lg mb-2 animate-pulse"></div>
    </div>

    <!-- 未选择任务 -->
    <div v-else-if="!taskStore.selectedTaskId" class="flex-1 flex items-center justify-center p-4">
      <p class="text-xs text-gray-400 dark:text-gray-500 text-center">请先选择一个任务</p>
    </div>

    <!-- 子对话列表 -->
    <div v-else-if="taskStore.threads.length > 0" class="flex-1 overflow-y-auto custom-scrollbar">
      <!-- 全部消息入口 -->
      <div
        @click="taskStore.selectThread(null)"
        :class="[
          'px-4 py-3 cursor-pointer border-b border-gray-100 dark:border-gray-700/50 transition-colors',
          taskStore.selectedThreadId === null
            ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
            : 'hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-400'
        ]"
      >
        <div class="flex items-center gap-2">
          <span class="material-icons-outlined text-sm">forum</span>
          <span class="text-sm font-medium">全部消息</span>
        </div>
      </div>

      <!-- 各子对话 -->
      <div
        v-for="thread in taskStore.threads"
        :key="thread.id"
        @click="taskStore.selectThread(thread.id)"
        :class="[
          'group px-4 py-3 cursor-pointer border-b border-gray-100 dark:border-gray-700/50 transition-colors',
          taskStore.selectedThreadId === thread.id
            ? 'bg-blue-50 dark:bg-blue-900/20'
            : 'hover:bg-gray-50 dark:hover:bg-gray-800'
        ]"
      >
        <div class="flex items-center justify-between">
          <span
            :class="[
              'text-sm truncate flex-1',
              taskStore.selectedThreadId === thread.id
                ? 'font-medium text-blue-700 dark:text-blue-300'
                : 'text-gray-700 dark:text-gray-300'
            ]"
          >
            {{ thread.title }}
          </span>
          <button
            @click.stop="handleDeleteThread(thread.id)"
            class="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-red-100 dark:hover:bg-red-900/30 text-gray-400 hover:text-red-500 dark:hover:text-red-400 transition-all"
            title="删除子对话"
          >
            <span class="material-icons-outlined text-sm">close</span>
          </button>
        </div>
        <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">
          {{ formatTime(thread.updated_at || thread.created_at) }}
        </p>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="flex-1 flex flex-col items-center justify-center p-4">
      <span class="material-icons-outlined text-3xl text-gray-300 dark:text-gray-600 mb-2">chat_bubble_outline</span>
      <p class="text-xs text-gray-400 dark:text-gray-500 text-center mb-3">暂无子对话</p>
      <button
        @click="handleCreateThread"
        class="px-3 py-1.5 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
      >
        新建对话
      </button>
    </div>
  </aside>
</template>

<script setup>
import { useTaskStore } from '@/stores/useTaskStore'
import { useToast } from '@/composables/useToast'

const taskStore = useTaskStore()
const { show: showToast } = useToast()

const handleCreateThread = async () => {
  const title = prompt('请输入子对话标题:')
  if (!title || !title.trim()) return

  try {
    await taskStore.createThread(taskStore.selectedTaskId, title.trim())
  } catch (error) {
    console.error('创建子对话失败:', error)
  }
}

const handleDeleteThread = async (threadId) => {
  if (!confirm('确定删除此子对话？相关消息也将被删除。')) return
  await taskStore.deleteThread(threadId)
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.4);
  border-radius: 20px;
}
.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.4);
}
</style>
