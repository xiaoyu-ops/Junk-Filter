<template>
  <div :class="['execution-card', `status-${executionData.status}`]">
    <!-- 顶部：状态指示器 -->
    <div class="flex items-center gap-3 mb-4">
      <!-- 状态图标（动态） -->
      <div :class="['status-icon', `icon-${executionData.status}`]">
        <span v-if="executionData.status === 'thinking'" class="animate-pulse text-xl">🧠</span>
        <span v-else-if="executionData.status === 'success'" class="text-xl">✅</span>
        <span v-else-if="executionData.status === 'failed'" class="text-xl">❌</span>
        <span v-else class="text-xl">⏳</span>
      </div>

      <!-- 状态文本 -->
      <div class="flex-1">
        <h4 :class="['font-semibold', `text-${statusColor}`]">
          {{ statusLabel }}
        </h4>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {{ formatTime(executionData.timestamp) }}
        </p>
      </div>
    </div>

    <!-- 中间：进度信息（思考中状态） -->
    <div v-if="executionData.status === 'thinking'" class="thinking-content mb-4">
      <!-- 进度条 -->
      <div class="thinking-bar h-1.5 bg-gray-300 dark:bg-gray-700 rounded-full overflow-hidden mb-2">
        <div class="thinking-progress h-full bg-gradient-to-r from-blue-400 to-blue-600 rounded-full animate-pulse"></div>
      </div>
      <!-- 状态文本 -->
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ executionData.message || '正在处理数据...' }}
      </p>
    </div>

    <!-- 中间：成功/失败结果显示 -->
    <div v-else class="result-content space-y-3">
      <!-- 统计信息 -->
      <div class="grid grid-cols-2 gap-2">
        <div class="bg-gray-50 dark:bg-gray-700/50 rounded p-3">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">获取文章数</p>
          <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ executionData.itemCount || 0 }}</p>
        </div>
        <div class="bg-gray-50 dark:bg-gray-700/50 rounded p-3">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">状态</p>
          <p :class="['text-lg font-semibold', `text-${statusColor}`]">{{ statusLabel }}</p>
        </div>
      </div>

      <!-- 摘要信息 -->
      <div v-if="executionData.summary" class="bg-blue-50 dark:bg-blue-900/20 rounded p-3 border border-blue-200 dark:border-blue-800">
        <p class="text-sm text-blue-900 dark:text-blue-100">{{ executionData.summary }}</p>
      </div>

      <!-- 错误信息 -->
      <div v-if="executionData.errorMessage" class="bg-red-50 dark:bg-red-900/20 rounded p-3 border border-red-200 dark:border-red-800">
        <div class="flex gap-2 items-start">
          <span class="material-icons-outlined text-sm text-red-600 dark:text-red-400 flex-shrink-0">error</span>
          <p class="text-sm text-red-700 dark:text-red-200">{{ executionData.errorMessage }}</p>
        </div>
      </div>

      <!-- 详细信息（可选） -->
      <div v-if="executionData.details" class="bg-gray-50 dark:bg-gray-700/30 rounded p-3 text-xs text-gray-600 dark:text-gray-400 max-h-32 overflow-y-auto">
        <p class="font-semibold mb-2">详细信息:</p>
        <pre class="whitespace-pre-wrap break-words">{{ executionData.details }}</pre>
      </div>
    </div>

    <!-- 底部：操作按钮 -->
    <div class="flex gap-2 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
      <button
        v-if="executionData.status === 'success' || executionData.status === 'failed'"
        @click="$emit('retry')"
        class="px-3 py-1.5 text-xs rounded-lg bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-900 dark:text-white transition-colors font-medium"
      >
        <span class="material-icons-outlined text-xs align-text-bottom inline mr-1">refresh</span>
        重新执行
      </button>
      <button
        @click="$emit('view-details')"
        class="px-3 py-1.5 text-xs rounded-lg bg-blue-100 hover:bg-blue-200 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 text-blue-700 dark:text-blue-300 transition-colors font-medium ml-auto"
      >
        <span class="material-icons-outlined text-xs align-text-bottom inline mr-1">info</span>
        查看详情
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  executionData: {
    type: Object,
    required: true,
    // {
    //   status: 'pending' | 'thinking' | 'success' | 'failed',
    //   itemCount: number,
    //   summary: string,
    //   errorMessage: string,
    //   message: string (仅 thinking 状态),
    //   timestamp: ISO8601,
    //   details: string (可选)
    // }
  },
})

const emit = defineEmits(['retry', 'view-details'])

/**
 * 计算状态标签文本
 */
const statusLabel = computed(() => {
  const labels = {
    pending: '待处理',
    thinking: '思考中...',
    success: '执行成功',
    failed: '执行失败',
  }
  return labels[props.executionData.status] || '未知状态'
})

/**
 * 计算状态对应的颜色
 */
const statusColor = computed(() => {
  const colors = {
    pending: 'gray-600 dark:text-gray-300',
    thinking: 'blue-600 dark:text-blue-400',
    success: 'green-600 dark:text-green-400',
    failed: 'red-600 dark:text-red-400',
  }
  return colors[props.executionData.status] || 'gray-600'
})

/**
 * 格式化时间戳
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

    if (diffMins < 1) return '刚刚'
    if (diffHours < 1) return `${diffMins} 分钟前`
    if (diffDays < 1 && date.toDateString() === now.toDateString()) {
      return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
    }
    if (diffDays < 2) {
      return `昨天 ${date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}`
    }
    return date.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
  } catch (error) {
    console.error('时间格式化失败:', error)
    return ''
  }
}
</script>

<style scoped>
.execution-card {
  @apply p-4 rounded-2xl rounded-tl-sm border transition-all duration-300;
}

.status-pending {
  @apply bg-gray-50 dark:bg-gray-900/30 border-gray-200 dark:border-gray-700;
}

.status-thinking {
  @apply bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800;
}

.status-success {
  @apply bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800;
}

.status-failed {
  @apply bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800;
}

.status-icon {
  @apply w-8 h-8 flex items-center justify-center flex-shrink-0;
}

/* 思考中的进度条动画 */
.thinking-bar {
  @apply relative;
}

.thinking-progress {
  animation: progress-pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes progress-pulse {
  0% {
    width: 0%;
    opacity: 1;
  }
  50% {
    width: 100%;
    opacity: 0.7;
  }
  100% {
    width: 100%;
    opacity: 0.3;
  }
}

.thinking-content {
  /* 思考状态的内容容器 */
}

.result-content {
  /* 结果显示的内容容器 */
}

/* 按钮样式 */
button {
  @apply flex items-center justify-center;
}

button:active {
  @apply scale-95;
}
</style>
