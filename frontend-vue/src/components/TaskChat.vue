<template>
  <section class="flex-1 bg-white dark:bg-[#111827] rounded-xl border border-gray-200 dark:border-gray-700/50 shadow-sm flex flex-col overflow-hidden relative">
    <!-- 工具栏：搜索、过滤、导出 -->
    <div class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 p-4 space-y-3">
      <!-- 搜索栏 -->
      <div class="flex items-center gap-3">
        <div class="flex-1 relative">
          <input
            v-model="searchText"
            @input="(e) => handleSearchInput(e.target.value)"
            type="text"
            placeholder="搜索消息... (实时搜索)"
            class="w-full pl-10 pr-10 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent transition-all"
          />
          <span class="material-icons-outlined absolute left-3 top-2.5 text-gray-400 dark:text-gray-500 text-lg">search</span>
          <button
            v-if="searchText"
            @click="clearSearch"
            class="absolute right-3 top-2.5 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors"
          >
            <span class="material-icons-outlined text-lg">close</span>
          </button>
        </div>
        <button
          @click="handleExportMessages"
          :disabled="isExporting || filteredMessages.length === 0"
          class="px-3 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 active:scale-95 disabled:bg-gray-400 disabled:cursor-not-allowed disabled:opacity-60 text-white text-sm font-medium transition-colors flex items-center gap-2"
        >
          <span class="material-icons-outlined text-sm" :class="{ 'animate-spin': isExporting }">{{ isExporting ? 'hourglass_top' : 'download' }}</span>
          <span class="hidden sm:inline">{{ isExporting ? '导出中...' : '导出' }}</span>
        </button>
      </div>

      <!-- 过滤栏 -->
      <div class="flex items-center gap-3 flex-wrap">
        <span class="text-xs font-medium text-gray-600 dark:text-gray-400">过滤:</span>

        <!-- 日期范围过滤 -->
        <select
          v-model="filterDateRange"
          class="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-all"
        >
          <option value="all">全部时间</option>
          <option value="today">今天</option>
          <option value="week">本周</option>
          <option value="month">本月</option>
        </select>

        <!-- 消息状态过滤 -->
        <select
          v-model="filterStatus"
          class="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-all"
        >
          <option value="all">全部状态</option>
          <option value="unread">未读</option>
          <option value="read">已读</option>
        </select>

        <!-- 导出格式选择 -->
        <select
          v-model="exportFormat"
          class="px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-all"
        >
          <option value="markdown">Markdown</option>
          <option value="json">JSON</option>
          <option value="csv">CSV</option>
        </select>

        <!-- 结果计数 -->
        <span class="text-xs text-gray-600 dark:text-gray-400 ml-auto">
          {{ filteredMessages.length }} / {{ messages.length }} 条消息
        </span>
      </div>

      <!-- 搜索状态指示 -->
      <div v-if="showSearchResults && searchText" class="flex items-center gap-2 text-xs text-blue-600 dark:text-blue-400">
        <span class="material-icons-outlined text-sm">search_insights</span>
        <span>找到 {{ searchResults.length }} 条搜索结果</span>
      </div>
    </div>

    <!-- 消息列表加载状态 -->
    <div
      v-if="messagesLoading && !taskStore.selectedTaskId === false"
      ref="containerRef"
      class="flex-1 overflow-y-auto space-y-4 p-6"
    >
      <SkeletonLoader :count="5" height="80px" />
    </div>

    <!-- 消息加载错误 -->
    <div
      v-else-if="messagesError"
      class="flex-1 overflow-y-auto p-6 flex items-center"
    >
      <div class="w-full">
        <ErrorCard
          :error="messagesError"
          title="加载消息失败"
          :message="messagesError.message || '无法加载消息历史，请检查网络连接'"
          :dismissible="false"
          @retry="handleRetryLoadMessages"
        />
      </div>
    </div>

    <!-- 消息列表容器 -->
    <div
      v-else-if="filteredMessages.length > 0"
      ref="containerRef"
      class="flex-1 overflow-y-auto space-y-4 p-6"
    >
      <!-- 消息列表循环 -->
      <template v-for="msg in filteredMessages" :key="msg.id">
        <ChatMessage :message="msg" />
      </template>

      <!-- AI 流式加载状态 -->
      <div v-if="isLoading" class="flex gap-4 animate-slide-in">
        <div class="flex-shrink-0 w-8 h-8 rounded-full bg-indigo-600 dark:bg-indigo-500 flex items-center justify-center">
          <span class="material-icons-outlined text-sm text-white">smart_toy</span>
        </div>
        <div class="max-w-2xl bg-gray-100 dark:bg-gray-800 rounded-2xl rounded-tl-sm p-4 flex gap-1.5">
          <span class="w-2 h-2 bg-gray-500 dark:bg-gray-400 rounded-full animate-bounce"></span>
          <span class="w-2 h-2 bg-gray-500 dark:bg-gray-400 rounded-full animate-bounce delay-100"></span>
          <span class="w-2 h-2 bg-gray-500 dark:bg-gray-400 rounded-full animate-bounce delay-200"></span>
        </div>
      </div>
    </div>

    <!-- 空状态提示 -->
    <div
      v-else
      class="flex-1 p-8 overflow-y-auto flex items-center justify-center"
    >
      <div class="text-center">
        <div class="w-16 h-16 bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-700 dark:to-gray-800 rounded-full mx-auto mb-4 flex items-center justify-center">
          <span class="material-icons-outlined text-4xl text-gray-500 dark:text-gray-400">
            {{ !taskStore.selectedTaskId ? 'chat' : (searchText ? 'search_off' : 'inbox') }}
          </span>
        </div>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
          {{ !taskStore.selectedTaskId ? '选择一个任务开始' : (searchText ? '未找到匹配消息' : '暂无消息') }}
        </h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {{ !taskStore.selectedTaskId
            ? '从左侧任务列表中选择一个任务，查看对话历史和执行详情'
            : (searchText
              ? `没有找到包含"${searchText}"的消息，试试其他关键词`
              : '该任务暂无消息，开始对话以创建记录') }}
        </p>
        <!-- 清空搜索按钮 (仅搜索无结果时显示) -->
        <button
          v-if="searchText"
          @click="clearSearch"
          class="px-4 py-2 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-white text-sm font-medium rounded-lg transition-colors"
        >
          清空搜索
        </button>
      </div>
    </div>

    <!-- 底部消息输入框 -->
    <div class="p-4 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-[#111827]">
      <div class="relative flex items-center w-full">
        <textarea
          v-model="inputText"
          @keydown.enter.exact="handleSendMessage"
          @keydown.shift.enter="insertNewline"
          :disabled="isLoading"
          placeholder="输入消息... (Shift+Enter 换行)"
          class="w-full pl-5 pr-14 py-3 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-[#1F2937] text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-indigo-500/50 focus:border-gray-300 dark:focus:border-transparent transition-all shadow-sm resize-none disabled:opacity-50 disabled:cursor-not-allowed"
          rows="2"
        />
        <button
          @click="handleSendMessage"
          :disabled="isLoading || !inputText.trim()"
          class="absolute right-2 bottom-2 p-2 bg-gray-100 hover:bg-gray-200 active:scale-95 dark:bg-[#374151] dark:hover:bg-[#4B5563] rounded-lg text-gray-600 dark:text-white transition-colors flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span class="material-icons-outlined transform -rotate-45" style="font-size: 20px;">send</span>
        </button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import { useAPI } from '@/composables/useAPI'
import { useScrollLock } from '@/composables/useScrollLock'
import { useSSE } from '@/composables/useSSE'
import { useToast } from '@/composables/useToast'
import ChatMessage from './ChatMessage.vue'
import SkeletonLoader from './SkeletonLoader.vue'
import ErrorCard from './ErrorCard.vue'

const taskStore = useTaskStore()
const { messages: messagesAPI } = useAPI()
const { show: showToast } = useToast()

// 消息列表
const messages = ref([])

// 消息加载状态
const messagesLoading = ref(false)

// 消息加载错误
const messagesError = ref(null)

// 输入框文本
const inputText = ref('')

// 加载状态
const isLoading = ref(false)

// 当前 AI 消息 ID（用于更新流式文本）
const currentAiMessageId = ref(null)

// ==================== 搜索功能 ====================
const searchText = ref('')
const searchDebounceTimer = ref(null)
const isSearching = ref(false)
const searchResults = ref([])
const showSearchResults = ref(false)

/**
 * 执行消息搜索（带防抖）
 */
const performSearch = async (query) => {
  if (!query.trim()) {
    searchResults.value = []
    showSearchResults.value = false
    return
  }

  isSearching.value = true
  try {
    searchResults.value = await messagesAPI.search(query, taskStore.selectedTaskId)
    showSearchResults.value = true
  } catch (error) {
    console.error('搜索消息失败:', error)
  } finally {
    isSearching.value = false
  }
}

/**
 * 处理搜索输入（防抖 300ms）
 */
const handleSearchInput = (value) => {
  clearTimeout(searchDebounceTimer.value)
  searchDebounceTimer.value = setTimeout(() => {
    performSearch(value)
  }, 300)
}

/**
 * 清空搜索
 */
const clearSearch = () => {
  searchText.value = ''
  searchResults.value = []
  showSearchResults.value = false
}

// ==================== 过滤功能 ====================
const filterDateRange = ref('all') // all, today, week, month
const filterStatus = ref('all') // all, read, unread

/**
 * 获取过滤后的消息列表
 */
const filteredMessages = computed(() => {
  if (showSearchResults.value) {
    return highlightSearchResults(searchResults.value)
  }

  let filtered = [...messages.value]

  // 按日期范围过滤
  const now = new Date()
  if (filterDateRange.value !== 'all') {
    const messageDate = (msg) => new Date(msg.timestamp)

    if (filterDateRange.value === 'today') {
      filtered = filtered.filter(msg => {
        const msgDate = messageDate(msg)
        return msgDate.toDateString() === now.toDateString()
      })
    } else if (filterDateRange.value === 'week') {
      const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
      filtered = filtered.filter(msg => messageDate(msg) >= weekAgo)
    } else if (filterDateRange.value === 'month') {
      const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
      filtered = filtered.filter(msg => messageDate(msg) >= monthAgo)
    }
  }

  // 按状态过滤
  if (filterStatus.value !== 'all') {
    filtered = filtered.filter(msg => {
      const isRead = msg.read === true
      if (filterStatus.value === 'read') return isRead
      if (filterStatus.value === 'unread') return !isRead
      return true
    })
  }

  return filtered
})

/**
 * 为搜索结果中的关键词进行高亮
 */
const highlightSearchResults = (results) => {
  return results.map(msg => ({
    ...msg,
    highlightedContent: msg.content,
    searchQuery: searchText.value,
  }))
}

// ==================== 导出功能 ====================
const isExporting = ref(false)
const exportFormat = ref('markdown')

/**
 * 处理消息导出
 */
const handleExportMessages = async () => {
  isExporting.value = true
  try {
    const { blob, filename } = await messagesAPI.export(exportFormat.value, taskStore.selectedTaskId)

    // 创建下载链接
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    showToast(`已导出 ${filteredMessages.value.length} 条消息为 ${exportFormat.value}`, 'success', 2000)
  } catch (error) {
    console.error('导出失败:', error)
    showToast('导出失败，请重试', 'error', 2000)
  } finally {
    isExporting.value = false
  }
}

// 使用 useScrollLock Composable
const {
  containerRef,
  isUserNearBottom,
  setupScrollListener,
  removeScrollListener,
  autoScrollToBottom,
  scrollToBottom,
} = useScrollLock()

// 使用 useSSE Composable
const { streamingText, connectSSE, closeSSE } = useSSE()

/**
 * 加载消息历史
 */
const loadMessages = async (taskId) => {
  if (!taskId) {
    messages.value = []
    messagesError.value = null
    return
  }

  messagesLoading.value = true
  messagesError.value = null
  try {
    messages.value = await messagesAPI.list(taskId)
  } catch (error) {
    console.error('加载消息失败:', error)
    messagesError.value = error
    messages.value = []
  } finally {
    messagesLoading.value = false
  }
}

/**
 * 重试加载消息
 */
const handleRetryLoadMessages = async () => {
  messagesError.value = null
  await loadMessages(taskStore.selectedTaskId)
}

/**
 * 初始化（组件挂载时）
 */
onMounted(() => {
  // 设置滚动监听
  setupScrollListener()
  // 初始化时滚到底部
  nextTick(() => {
    scrollToBottom()
  })
  // 如果已有选中的任务，加载消息
  if (taskStore.selectedTaskId) {
    loadMessages(taskStore.selectedTaskId)
  }
})

/**
 * 清理（组件卸载时）
 */
onUnmounted(() => {
  removeScrollListener()
  closeSSE()  // 关闭任何打开的 SSE 连接
})

/**
 * 监听选中任务变化，加载对应的消息历史
 */
watch(() => taskStore.selectedTaskId, (taskId) => {
  if (taskId) {
    loadMessages(taskId)
    clearSearch()  // 切换任务时清空搜索
  }
})

/**
 * 监听消息列表变化，自动滚到底部（如果用户在底部）
 */
watch(filteredMessages, async () => {
  await autoScrollToBottom()
}, { deep: true })

/**
 * 处理发送消息
 */
const handleSendMessage = async (e) => {
  // 防止默认行为（如果是键盘事件）
  if (e && e.preventDefault) {
    e.preventDefault()
  }

  // 验证输入
  const trimmedText = inputText.value.trim()
  if (!trimmedText || isLoading.value) {
    return
  }

  // 创建用户消息
  const userMessage = {
    id: `msg-user-${Date.now()}`,
    role: 'user',
    type: 'text',
    content: trimmedText,
    timestamp: new Date().toISOString(),
    read: false,
  }

  // 添加到消息列表
  messages.value.push(userMessage)
  inputText.value = ''

  // 保存用户消息到后端
  try {
    await messagesAPI.save({
      task_id: taskStore.selectedTaskId,
      role: 'user',
      type: 'text',
      content: trimmedText,
    })
  } catch (error) {
    console.error('保存用户消息失败:', error)
    // 即使保存失败也继续处理 AI 回复
  }

  // 设置加载状态
  isLoading.value = true

  try {
    // 尝试真实的 SSE 连接
    await handleSSEResponse(trimmedText)
  } catch (sseError) {
    console.warn('[TaskChat] SSE 失败，降级到模拟响应:', sseError)
    // SSE 失败时降级到模拟响应
    await simulateAiResponse(trimmedText)
  } finally {
    isLoading.value = false
  }
}

/**
 * Handle real SSE streaming response - Fixed version
 */
const handleSSEResponse = async (userInput) => {
  // Build request URL (configure in environment variables)
  // Use Mock backend (3000) for SSE chat, NOT Go backend (8080)
  const mockUrl = import.meta.env.VITE_MOCK_URL || 'http://localhost:3000'

  // ⚠️ Key fix: Add required parameters taskId and message
  const streamEndpoint = `${mockUrl}/api/chat/stream?taskId=${taskStore.selectedTaskId}&message=${encodeURIComponent(userInput)}`

  // 创建 AI 消息占位符（延迟添加）
  const aiMessagePlaceholder = {
    id: `msg-ai-${Date.now()}`,
    role: 'ai',
    type: 'text',
    content: '',
    timestamp: new Date().toISOString(),
    read: false,
  }

  currentAiMessageId.value = aiMessagePlaceholder.id

  // ⚠️ 修复：跟踪消息是否已添加（防止重复或缺失）
  let aiMessageAdded = false

  try {
    // 使用 SSE 连接流式接收
    await connectSSE(streamEndpoint, {
      // ⚠️ 修复：首次接收数据时才添加消息
      onStreamingText: (text) => {
        if (!aiMessageAdded) {
          messages.value.push(aiMessagePlaceholder)
          aiMessageAdded = true
        }

        const messageIndex = messages.value.findIndex(m => m.id === aiMessagePlaceholder.id)
        if (messageIndex !== -1) {
          messages.value[messageIndex].content = text
          // 触发 watch 以更新滚动
          messages.value[messageIndex] = { ...messages.value[messageIndex] }
        }
      },

      // 处理执行卡片或其他特殊消息
      onMessage: (msgData) => {
        if (msgData.type === 'execution') {
          // 添加执行卡片消息
          messages.value.push({
            id: `msg-execution-${Date.now()}`,
            role: 'ai',
            type: 'execution',
            executionData: msgData.data,
            timestamp: new Date().toISOString(),
            read: false,
          })
        }
      },

      // 流完成时
      onComplete: (finalText, data) => {
        console.log('[SSE] 流式回复完成:', finalText)
        currentAiMessageId.value = null

        // ⚠️ 修复：确保消息已添加
        if (!aiMessageAdded) {
          messages.value.push(aiMessagePlaceholder)
          aiMessageAdded = true
        }

        // 保存 AI 消息到后端
        messagesAPI.save({
          task_id: taskStore.selectedTaskId,
          role: 'ai',
          type: 'text',
          content: finalText,
        }).catch(error => {
          console.error('保存 AI 消息失败:', error)
        })
      },

      // 错误处理 - 修复版
      onError: (err) => {
        console.error('[SSE] 流式回复错误:', err)

        // ⚠️ 修复：只有在完全没数据的情况下才显示错误卡片
        if (!aiMessageAdded) {
          messages.value.push({
            id: `msg-error-${Date.now()}`,
            role: 'ai',
            type: 'error',
            content: `流式回复失败: ${err}`,
            timestamp: new Date().toISOString(),
            read: false,
          })
        } else {
          // 已有部分数据，只在控制台记录错误
          console.warn('[SSE] 已接收部分数据，忽略错误卡片')
        }

        throw err
      },
    })
  } catch (error) {
    console.error('[SSE] 连接失败:', error)
    throw error
  }
}

/**
 * 模拟 AI 回复（降级用或演示用）
 * 这个函数会被保留，用于：
 * 1. SSE 连接失败时的降级方案
 * 2. 本地开发没有后端时的演示
 */
const simulateAiResponse = async (userInput) => {
  // 模拟网络延迟
  await new Promise(resolve => setTimeout(resolve, 800))

  // 简单的 AI 回复示例
  const responses = {
    '你好': '你好！👋 我是 TrueSignal AI 助手。有什么我可以帮助你的吗？',
    '帮助': '我可以帮助你：\n1. 分析 RSS 内容\n2. 生成摘要\n3. 评估信息质量',
    default: `我已收到你的消息："${userInput}"。我正在思考如何最好地帮助你...\n\n这是一条示例回复，用来演示 Markdown 渲染：\n\n### 功能示例\n- **加粗文本**\n- *斜体文本*\n- [链接示例](https://example.com)\n- \`代码示例\`\n\n\`\`\`javascript\nconst hello = () => {\n  console.log('Hello World')\n}\n\`\`\``,
  }

  const responseText = responses[userInput] || responses.default

  // 创建 AI 消息
  const aiMessage = {
    id: `msg-ai-${Date.now()}`,
    role: 'ai',
    type: 'text',
    content: responseText,
    timestamp: new Date().toISOString(),
    read: false,
  }

  messages.value.push(aiMessage)

  // 50% 概率添加执行卡片
  if (Math.random() > 0.5) {
    await new Promise(resolve => setTimeout(resolve, 500))

    const executionMessage = {
      id: `msg-execution-${Date.now()}`,
      role: 'ai',
      type: 'execution',
      executionData: {
        status: 'success',
        itemCount: Math.floor(Math.random() * 100) + 10,
        summary: '成功处理了任务，获取了相关信息。',
        timestamp: new Date().toISOString(),
      },
      timestamp: new Date().toISOString(),
      read: false,
    }

    messages.value.push(executionMessage)
  }
}

/**
 * 处理 Shift+Enter 换行
 */
const insertNewline = (e) => {
  e.preventDefault()
  const textarea = e.target
  const start = textarea.selectionStart
  const end = textarea.selectionEnd

  inputText.value = inputText.value.substring(0, start) + '\n' + inputText.value.substring(end)

  // 重新设置光标位置
  nextTick(() => {
    textarea.selectionStart = textarea.selectionEnd = start + 1
  })
}
</script>

<style scoped>
/* 加载动画延迟 */
.delay-100 {
  animation-delay: 0.1s;
}

.delay-200 {
  animation-delay: 0.2s;
}

/* 消息进入动画 */
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

/* Textarea 样式 */
textarea::placeholder {
  @apply text-gray-400 dark:text-gray-400;
}

textarea:focus {
  @apply outline-none;
}
</style>
