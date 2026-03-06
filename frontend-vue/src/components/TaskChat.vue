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
      <!-- 消息列表循环 -->
      <template v-for="msg in messages" :key="msg.id">
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
            {{ !taskStore.selectedTaskId ? 'chat' : 'chat_bubble_outline' }}
          </span>
        </div>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
          {{ !taskStore.selectedTaskId ? '选择一个任务开始' : '暂无消息' }}
        </h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {{ !taskStore.selectedTaskId
            ? '从左侧任务列表中选择一个任务，查看对话历史和执行详情'
            : '该任务暂无消息，开始对话以创建记录' }}
        </p>
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
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
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

  // 保存用户消息到后端
  try {
    await messagesAPI.save(taskStore.selectedTaskId, {
      role: 'user',
      type: 'text',
      content: trimmedText,
      thread_id: taskStore.selectedThreadId || undefined,
    })
  } catch (error) {
    console.error('保存用户消息失败:', error)
    // 即使保存失败也继续处理 AI 回复
  }

  // 设置加载状态
  isLoading.value = true

  try {
    // 使用真实的 SSE 连接
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
 * 处理真实 SSE 流式聊天响应（Agent 调优与咨询）
 * 从 Go 后端的 /api/tasks/{id}/chat 接收自然语言回复
 */
const handleSSEResponse = async (userInput) => {
  // 创建 AI 消息占位符
  const aiMessagePlaceholder = {
    id: `msg-ai-${Date.now()}`,
    role: 'ai',
    type: 'text',
    content: '',
    timestamp: new Date().toISOString(),
    processing: true,
    messageType: 'ai_reply',  // 标记为自然语言回复
  }

  currentAiMessageId.value = aiMessagePlaceholder.id
  let aiMessageAdded = false
  let aiResponseContent = ''
  let parameterUpdates = null
  let referencedCards = []

  try {
    // 获取当前 LLM 和评估配置
    // 优先从 localStorage 读取，其次从 store 读取
    const savedConfig = localStorage.getItem('config') ? JSON.parse(localStorage.getItem('config')) : {}

    const llmConfig = {
      modelName: savedConfig.modelName || configStore.modelName,
      apiKey: savedConfig.apiKey || configStore.apiKey,
      baseUrl: savedConfig.baseUrl || configStore.baseUrl,
    }
    const evalConfig = {
      temperature: savedConfig.temperature || configStore.temperature,
      topP: savedConfig.topP || configStore.topP,
      maxTokens: savedConfig.maxTokens || configStore.maxTokens,
    }

    console.log('[Task Chat] Using LLM Config:', llmConfig)
    console.log('[Task Chat] Using Eval Config:', evalConfig)

    // 启动新的任务聊天 API（传递配置参数）
    closeSseConnection.value = chatAPI.taskChat(
      taskStore.selectedTaskId,
      userInput,
      llmConfig,      // ← 传递 LLM 配置
      evalConfig,     // ← 传递评估配置
      taskStore.selectedThreadId,  // ← 传递子对话 ID
      (eventData) => {
        console.log('[Task Chat Event]', eventData)

        // 首次接收数据时添加占位符
        if (!aiMessageAdded) {
          messages.value.push(aiMessagePlaceholder)
          aiMessageAdded = true
        }

        // 处理不同的事件状态
        if (eventData.status === 'processing') {
          // 更新占位符显示处理中状态
          const msgIndex = messages.value.findIndex(m => m.id === aiMessagePlaceholder.id)
          if (msgIndex !== -1) {
            messages.value[msgIndex].processing = true
            messages.value[msgIndex] = { ...messages.value[msgIndex] }
          }
        } else if (eventData.status === 'completed') {
          // 处理完成，更新自然语言回复
          if (eventData.result) {
            const result = eventData.result
            aiResponseContent = result.reply  // 自然语言回复

            // 提取参数更新和卡片引用
            parameterUpdates = result.parameter_updates
            referencedCards = result.referenced_card_ids || []

            const msgIndex = messages.value.findIndex(m => m.id === aiMessagePlaceholder.id)
            if (msgIndex !== -1) {
              messages.value[msgIndex].content = aiResponseContent
              messages.value[msgIndex].processing = false
              messages.value[msgIndex].messageType = 'ai_reply'
              messages.value[msgIndex].referencedCards = referencedCards
              messages.value[msgIndex].parameterUpdates = parameterUpdates
              messages.value[msgIndex] = { ...messages.value[msgIndex] }

              // 如果有参数更新建议，在控制台输出
              if (parameterUpdates) {
                console.log('[Agent Suggestion] 参数更新建议:', parameterUpdates)
              }
            }
          }
        } else if (eventData.status === 'stream_end') {
          // 流完成标记，正常结束
          console.log('[Task Chat] 流正常完成')
        } else if (eventData.status === 'error') {
          // 错误处理
          const msgIndex = messages.value.findIndex(m => m.id === aiMessagePlaceholder.id)
          if (msgIndex !== -1) {
            messages.value[msgIndex].content = `⚠️ Agent 响应失败: ${eventData.error || '未知错误'}`
            messages.value[msgIndex].processing = false
            messages.value[msgIndex].type = 'error'
            messages.value[msgIndex] = { ...messages.value[msgIndex] }
          }
        }
      }
    )

    // 等待聊天完成（通过 Promise 或 timeout）
    await new Promise(resolve => {
      const checkCompletion = setInterval(() => {
        const msg = messages.value.find(m => m.id === aiMessagePlaceholder.id)
        if (msg && !msg.processing) {
          clearInterval(checkCompletion)
          resolve()
        }
      }, 100)

      // 30秒超时
      setTimeout(() => {
        clearInterval(checkCompletion)
        resolve()
      }, 30000)
    })

    // 保存 AI 消息到后端
    if (aiResponseContent) {
      try {
        await messagesAPI.save(taskStore.selectedTaskId, {
          role: 'ai',
          type: 'text',
          content: aiResponseContent,
          thread_id: taskStore.selectedThreadId || undefined,
          metadata: JSON.stringify({
            messageType: 'ai_reply',
            parameterUpdates: parameterUpdates,
            referencedCards: referencedCards,
          }),
        })
      } catch (error) {
        console.error('保存 AI 消息失败:', error)
      }
    }
  } catch (error) {
    console.error('[Task Chat] 连接失败:', error)

    // 如果消息已添加，标记为错误；否则抛出异常让调用方降级
    const msgIndex = messages.value.findIndex(m => m.id === aiMessagePlaceholder.id)
    if (msgIndex !== -1) {
      messages.value[msgIndex].content = `⚠️ 连接失败: ${error.message}`
      messages.value[msgIndex].processing = false
      messages.value[msgIndex].type = 'error'
    } else {
      throw error
    }
  } finally {
    currentAiMessageId.value = null
  }
}

/**
 * 模拟 AI 回复（降级用或演示用）
 */
const simulateAiResponse = async (userInput) => {
  // 模拟网络延迟
  await new Promise(resolve => setTimeout(resolve, 800))

  const responseText = `我已收到你的消息："${userInput}"。我正在思考如何最好地帮助你...\n\n这是一条示例回复，用来演示 Markdown 渲染：\n\n### 功能示例\n- **加粗文本**\n- *斜体文本*\n- \`代码示例\`\n\n\`\`\`javascript\nconst hello = () => {\n  console.log('Hello World')\n}\n\`\`\``

  // 创建 AI 消息
  const aiMessage = {
    id: `msg-ai-${Date.now()}`,
    role: 'ai',
    type: 'text',
    content: responseText,
    timestamp: new Date().toISOString(),
  }

  messages.value.push(aiMessage)
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
