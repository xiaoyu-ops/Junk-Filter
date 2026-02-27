<template>
  <!-- 用户消息 -->
  <div v-if="message.role === 'user'" class="flex gap-4 justify-end animate-slide-in">
    <!-- 用户消息气泡 -->
    <div class="max-w-2xl bg-gray-900 dark:bg-indigo-600 text-white rounded-2xl rounded-tr-sm p-4 shadow-sm">
      <div class="prose prose-sm prose-invert max-w-none text-white break-words">
        {{ message.content }}
      </div>
      <p v-if="message.timestamp" class="text-xs text-gray-300 mt-2 opacity-70">
        {{ formatTime(message.timestamp) }}
      </p>
    </div>

    <!-- 用户头像 -->
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center border border-gray-300 dark:border-gray-600">
      <span class="material-icons-outlined text-sm text-gray-600 dark:text-gray-300">person</span>
    </div>
  </div>

  <!-- AI 文本消息 -->
  <div v-else-if="message.type === 'text'" class="flex gap-4 justify-start animate-slide-in">
    <!-- AI 头像 -->
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-600 dark:bg-indigo-500 flex items-center justify-center">
      <span class="material-icons-outlined text-sm text-white">smart_toy</span>
    </div>

    <!-- AI 消息气泡 -->
    <div class="max-w-2xl bg-gray-100 dark:bg-gray-800 rounded-2xl rounded-tl-sm p-4 shadow-sm">
      <!-- Markdown 渲染的内容 -->
      <div
        class="prose prose-sm dark:prose-invert max-w-none"
        v-html="renderMarkdown(message.content)"
      />
      <p v-if="message.timestamp" class="text-xs text-gray-500 dark:text-gray-400 mt-2 opacity-70">
        {{ formatTime(message.timestamp) }}
      </p>
    </div>
  </div>

  <!-- AI 错误消息 -->
  <div v-else-if="message.type === 'error'" class="flex gap-4 justify-start animate-slide-in">
    <!-- AI 头像 -->
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-red-600 flex items-center justify-center">
      <span class="material-icons-outlined text-sm text-white">error</span>
    </div>

    <!-- 错误消息框 -->
    <div class="max-w-2xl">
      <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-2xl rounded-tl-sm p-4">
        <div class="flex gap-3 items-start">
          <span class="material-icons-outlined text-red-600 dark:text-red-400 flex-shrink-0">warning</span>
          <div class="flex-1">
            <p class="text-sm font-medium text-red-900 dark:text-red-100">发生错误</p>
            <p class="text-sm text-red-700 dark:text-red-200 mt-1">{{ message.content }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- AI 执行卡片消息 -->
  <div v-else-if="message.type === 'execution'" class="flex gap-4 justify-start animate-slide-in">
    <!-- AI 头像 -->
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-600 dark:bg-indigo-500 flex items-center justify-center">
      <span class="material-icons-outlined text-sm text-white">auto_awesome</span>
    </div>

    <!-- ExecutionCard 组件 -->
    <div class="max-w-2xl w-full">
      <ExecutionCard :execution-data="message.executionData" />
    </div>
  </div>

  <!-- AI 加载状态 -->
  <div v-else-if="message.type === 'loading'" class="flex gap-4 justify-start animate-slide-in">
    <!-- AI 头像 -->
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-600 dark:bg-indigo-500 flex items-center justify-center">
      <span class="material-icons-outlined text-sm text-white">smart_toy</span>
    </div>

    <!-- 加载动画 -->
    <div class="max-w-2xl bg-gray-100 dark:bg-gray-800 rounded-2xl rounded-tl-sm p-4 flex gap-1.5">
      <span class="w-2 h-2 bg-gray-500 dark:bg-gray-400 rounded-full animate-bounce"></span>
      <span class="w-2 h-2 bg-gray-500 dark:bg-gray-400 rounded-full animate-bounce delay-100"></span>
      <span class="w-2 h-2 bg-gray-500 dark:bg-gray-400 rounded-full animate-bounce delay-200"></span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useMarkdown } from '@/composables/useMarkdown'
import ExecutionCard from './ExecutionCard.vue'

const props = defineProps({
  message: {
    type: Object,
    required: true,
    // {
    //   id: string,
    //   role: 'user' | 'ai',
    //   type: 'text' | 'error' | 'execution' | 'loading',
    //   content: string,
    //   timestamp: ISO8601 string,
    //   executionData: object (仅当 type='execution' 时)
    // }
  },
})

const { renderMarkdown } = useMarkdown()

/**
 * 格式化时间戳为可读的时间字符串
 * @param {string} timestamp - ISO8601 时间戳
 * @returns {string} 格式化的时间（如 "14:30" 或 "昨天 14:30"）
 */
const formatTime = (timestamp) => {
  if (!timestamp) return ''

  try {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    // 不到 1 分钟：显示 "刚刚"
    if (diffMins < 1) {
      return '刚刚'
    }
    // 1 小时内：显示 "N 分钟前"
    if (diffHours < 1) {
      return `${diffMins} 分钟前`
    }
    // 同一天：显示 "14:30"
    if (diffDays < 1 && date.toDateString() === now.toDateString()) {
      return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    }
    // 昨天：显示 "昨天 14:30"
    if (diffDays < 2) {
      return `昨天 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
    }
    // 其他：显示 "2月26日 14:30"
    return date.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
  } catch (error) {
    console.error('时间格式化失败:', error)
    return ''
  }
}
</script>

<style scoped>
/* 消息气泡进入动画 */
@keyframes slide-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-slide-in {
  animation: slide-in 0.3s ease-out;
}

/* 加载动画延迟 */
.delay-100 {
  animation-delay: 0.1s;
}

.delay-200 {
  animation-delay: 0.2s;
}

/* Markdown 样式（深色模式友好） */
:deep(h1),
:deep(h2),
:deep(h3),
:deep(h4),
:deep(h5),
:deep(h6) {
  @apply font-semibold mt-3 mb-2;
}

:deep(h1) {
  @apply text-xl;
}

:deep(h2) {
  @apply text-lg;
}

:deep(h3),
:deep(h4),
:deep(h5),
:deep(h6) {
  @apply text-base;
}

:deep(p) {
  @apply mb-2 leading-relaxed;
}

:deep(ul),
:deep(ol) {
  @apply ml-4 mb-2;
}

:deep(li) {
  @apply mb-1;
}

:deep(code) {
  @apply bg-gray-200 dark:bg-gray-700 px-2 py-0.5 rounded text-sm font-mono;
}

:deep(pre) {
  @apply bg-gray-900 dark:bg-black text-gray-100 p-4 rounded-lg overflow-x-auto mb-2;
}

:deep(pre code) {
  @apply bg-transparent px-0 py-0;
}

:deep(blockquote) {
  @apply border-l-4 border-blue-500 pl-4 text-gray-600 dark:text-gray-400 italic my-2;
}

:deep(a) {
  @apply text-blue-600 dark:text-blue-400 hover:underline;
}

:deep(table) {
  @apply border-collapse border border-gray-300 dark:border-gray-600 w-full my-2;
}

:deep(th),
:deep(td) {
  @apply border border-gray-300 dark:border-gray-600 px-3 py-2 text-left;
}

:deep(th) {
  @apply bg-gray-200 dark:bg-gray-700 font-semibold;
}

:deep(hr) {
  @apply my-4 border-gray-300 dark:border-gray-600;
}

:deep(img) {
  @apply max-w-full h-auto rounded-lg;
}
</style>
