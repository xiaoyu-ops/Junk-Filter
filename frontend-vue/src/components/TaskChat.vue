<template>
  <section class="flex-1 bg-white dark:bg-[#111827] rounded-xl border border-gray-200 dark:border-gray-700/50 shadow-sm flex flex-col overflow-hidden relative">
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
      v-else-if="messages.length > 0"
      ref="containerRef"
      class="flex-1 overflow-y-auto space-y-4 p-6"
    >
      <!-- 消息列表循环（跳过正在处理中的空消息） -->
      <template v-for="msg in messages" :key="msg.id">
        <ChatMessage v-if="!msg.processing || msg.content || (msg.toolCalls && msg.toolCalls.length > 0)" :message="msg" />
      </template>

      <!-- AI 流式加载状态：仅在没有任何 AI 消息正在显示内容时才展示 -->
      <div v-if="isLoading && !hasAiContent" class="flex gap-4 animate-slide-in">
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
        <div class="w-16 h-16 bg-gradient-to-br from-indigo-100 to-purple-100 dark:from-indigo-900/40 dark:to-purple-900/40 rounded-full mx-auto mb-4 flex items-center justify-center">
          <span class="material-icons-outlined text-4xl text-indigo-500 dark:text-indigo-400">smart_toy</span>
        </div>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">Agent 就绪</h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          你可以问我任何关于 RSS 内容、评估结果或系统状态的问题
        </p>
      </div>
    </div>

    <!-- 底部消息输入框 -->
    <div class="p-4 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-[#111827]">
      <div class="relative flex items-center w-full">
        <textarea
          v-model="inputText"
          @keydown.enter.exact.prevent="handleSendMessage"
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
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import { useConfigStore } from '@/stores/useConfigStore'
import { useAPI } from '@/composables/useAPI'
import { useScrollLock } from '@/composables/useScrollLock'
import { useToast } from '@/composables/useToast'
import ChatMessage from './ChatMessage.vue'
import SkeletonLoader from './SkeletonLoader.vue'
import ErrorCard from './ErrorCard.vue'

const taskStore = useTaskStore()
const configStore = useConfigStore()
const { messages: messagesAPI, chat: chatAPI } = useAPI()
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

// 当前 AI 占位消息是否已有内容（用于判断是否隐藏 loading 动画）
const hasAiContent = computed(() => {
  if (!currentAiMessageId.value) return false
  const msg = messages.value.find(m => m.id === currentAiMessageId.value)
  return msg && msg.content && msg.content.length > 0
})

// SSE 关闭函数
const closeSseConnection = ref(null)

// 使用 useScrollLock Composable
const {
  containerRef,
  isUserNearBottom,
  setupScrollListener,
  removeScrollListener,
  autoScrollToBottom,
  scrollToBottom,
} = useScrollLock()

/**
 * 加载消息历史
 */
const loadMessages = async (taskId, threadId = undefined) => {
  if (!taskId) {
    messages.value = []
    messagesError.value = null
    return
  }

  messagesLoading.value = true
  messagesError.value = null
  try {
    const messageList = await messagesAPI.list(taskId, { threadId })
    // 规范化消息格式（适配不同后端返回的字段）
    messages.value = (messageList || []).map(msg => ({
      id: msg.id || `msg-${Date.now()}`,
      role: msg.role || 'user',
      type: msg.type || 'text',
      content: msg.content || '',
      timestamp: msg.created_at || msg.timestamp || new Date().toISOString(),
      created_at: msg.created_at,
      updated_at: msg.updated_at,
      thread_id: msg.thread_id,
    }))
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
  await loadMessages(taskStore.selectedTaskId, taskStore.selectedThreadId)
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
    loadMessages(taskStore.selectedTaskId, taskStore.selectedThreadId)
  }
})

/**
 * 清理（组件卸载时）
 */
onUnmounted(() => {
  removeScrollListener()
  // 关闭任何打开的 SSE 连接
  if (closeSseConnection.value) {
    closeSseConnection.value()
  }
})

/**
 * 监听选中任务变化，加载对应的消息历史
 */
watch(() => taskStore.selectedTaskId, (taskId) => {
  if (taskId) {
    loadMessages(taskId, taskStore.selectedThreadId)
  }
})

/**
 * 监听选中子对话变化，重新加载消息
 */
watch(() => taskStore.selectedThreadId, (threadId) => {
  if (taskStore.selectedTaskId) {
    loadMessages(taskStore.selectedTaskId, threadId)
  }
})

/**
 * 监听消息列表变化，自动滚到底部（如果用户在底部）
 */
watch(messages, async () => {
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
  }

  // 添加到消息列表
  messages.value.push(userMessage)
  inputText.value = ''

  // 持久化用户消息到 DB
  if (taskStore.selectedTaskId) {
    messagesAPI.save(taskStore.selectedTaskId, userMessage).catch(() => {})
  }

  // 设置加载状态
  isLoading.value = true

  try {
    await handleSSEResponse(trimmedText)
  } catch (sseError) {
    console.error('[Agent Chat] SSE 失败:', sseError)
  } finally {
    isLoading.value = false
  }
}

/**
 * 处理 Agent SSE 流式响应
 * 调用 Python ReAct harness /api/agent/chat
 * 事件类型：tool_call / text / done / error
 */
const handleSSEResponse = async (userInput) => {
  const aiPlaceholder = {
    id: `msg-ai-${Date.now()}`,
    role: 'ai',
    type: 'text',
    content: '',
    toolCalls: [],        // 工具调用记录（内联显示）
    timestamp: new Date().toISOString(),
    processing: true,
  }

  currentAiMessageId.value = aiPlaceholder.id
  let placeholderAdded = false

  // 读取 LLM 配置
  const savedConfig = localStorage.getItem('config') ? JSON.parse(localStorage.getItem('config')) : {}
  const llmConfig = {
    modelName: savedConfig.modelName || configStore.modelName,
    apiKey:    savedConfig.apiKey    || configStore.apiKey,
    baseUrl:   savedConfig.baseUrl   || configStore.baseUrl,
  }

  // 构建历史（去除占位符）
  const history = messages.value.filter(m => m.content && !m.processing)

  try {
    await new Promise((resolve) => {
      const timeout = setTimeout(resolve, 60000)

      closeSseConnection.value = chatAPI.agentChat(
        userInput,
        history,
        llmConfig,
        (event) => {
          // 首次有事件时插入占位符
          if (!placeholderAdded) {
            messages.value.push(aiPlaceholder)
            placeholderAdded = true
          }

          const idx = messages.value.findIndex(m => m.id === aiPlaceholder.id)
          if (idx === -1) return

          if (event.type === 'tool_call') {
            // 追加工具调用记录
            messages.value[idx].toolCalls = [
              ...messages.value[idx].toolCalls,
              { tool: event.tool, args: event.args, result: event.result },
            ]
            messages.value[idx] = { ...messages.value[idx] }

          } else if (event.type === 'chunk') {
            // 流式文字片段 → 逐字追加
            messages.value[idx].content = (messages.value[idx].content || '') + (event.content || '')
            messages.value[idx] = { ...messages.value[idx] }

          } else if (event.type === 'text') {
            messages.value[idx].content = event.content || ''
            messages.value[idx] = { ...messages.value[idx] }

          } else if (event.type === 'done') {
            messages.value[idx].processing = false
            messages.value[idx] = { ...messages.value[idx] }
            isLoading.value = false
            // 持久化 AI 回复到 DB
            const aiMsg = messages.value[idx]
            if (taskStore.selectedTaskId && aiMsg.content) {
              messagesAPI.save(taskStore.selectedTaskId, {
                role: 'ai', type: 'text', content: aiMsg.content,
                metadata: JSON.stringify({ message_type: 'ai_reply' }),
              }).catch(() => {})
            }
            clearTimeout(timeout)
            resolve()

          } else if (event.type === 'error') {
            messages.value[idx].content = `⚠️ ${event.error || '未知错误'}`
            messages.value[idx].processing = false
            messages.value[idx].type = 'error'
            messages.value[idx] = { ...messages.value[idx] }
            isLoading.value = false
            clearTimeout(timeout)
            resolve()
          }
        }
      )
    })
  } catch (error) {
    console.error('[Agent Chat] 连接失败:', error)
    const idx = messages.value.findIndex(m => m.id === aiPlaceholder.id)
    if (idx !== -1) {
      messages.value[idx].content = `⚠️ 连接失败: ${error.message}`
      messages.value[idx].processing = false
      messages.value[idx].type = 'error'
    } else {
      throw error
    }
  } finally {
    currentAiMessageId.value = null
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
