<template>
  <!-- AI 助手模态框 -->
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
    leave-active-class="transition-all duration-300 ease-out"
  >
    <div
      v-if="taskStore.isAIAssistantModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-md p-4 transition-opacity"
    >
      <!-- 模态框容器 -->
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div v-if="taskStore.isAIAssistantModalOpen" class="bg-white dark:bg-gray-800 w-full max-w-4xl h-[80vh] rounded-3xl shadow-2xl border border-gray-200 dark:border-gray-700 flex flex-col overflow-hidden">
          <!-- 标题区 -->
          <div class="px-8 py-6 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-white dark:bg-gray-800">
            <div>
              <h2 class="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <span class="material-icons-outlined text-blue-600 dark:text-blue-400">auto_awesome</span>
                AI 任务创建助手
              </h2>
              <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">用自然语言描述你的需求，AI 会帮你创建任务</p>
            </div>
            <button
              @click="taskStore.closeAIAssistantModal"
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors p-1"
            >
              <span class="material-icons-outlined">close</span>
            </button>
          </div>

          <!-- 对话区域 -->
          <div class="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-900/30 p-8 custom-scrollbar">
            <div class="max-w-3xl mx-auto space-y-4">
              <!-- 欢迎消息 -->
              <div v-if="messages.length === 0" class="text-center py-12">
                <span class="text-6xl mb-4 block">🤖</span>
                <h3 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                  你好！我是任务创建助手
                </h3>
                <p class="text-gray-600 dark:text-gray-400 max-w-sm mx-auto">
                  请描述你想要监控的内容或需求，我会帮你分析并创建任务。比如："我想监控 GitHub 上关于 AI 的项目"
                </p>
              </div>

              <!-- 消息列表 -->
              <template v-for="(msg, idx) in messages" :key="idx">
                <!-- 用户消息 -->
                <div v-if="msg.role === 'user'" class="flex justify-end">
                  <div class="bg-blue-600 text-white rounded-2xl px-6 py-3 max-w-2xl shadow-sm">
                    <p class="text-sm leading-relaxed">{{ msg.content }}</p>
                  </div>
                </div>

                <!-- AI 消息 -->
                <div v-else class="flex justify-start">
                  <div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl px-6 py-3 max-w-2xl shadow-sm">
                    <p class="text-sm leading-relaxed text-gray-900 dark:text-gray-100">{{ msg.content }}</p>
                  </div>
                </div>
              </template>

              <!-- 加载指示 -->
              <div v-if="isWaitingForAI" class="flex justify-start">
                <div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl px-6 py-3 shadow-sm">
                  <div class="flex gap-2">
                    <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0s"></div>
                    <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.1s"></div>
                    <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
                  </div>
                </div>
              </div>

              <!-- 确认区域 -->
              <div v-if="showConfirmation" class="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700/50 rounded-xl">
                <h4 class="font-medium text-gray-900 dark:text-white mb-3">即将创建以下任务：</h4>
                <div class="space-y-2 text-sm text-gray-700 dark:text-gray-300">
                  <p><strong>任务名称：</strong> {{ pendingTask.name }}</p>
                  <p><strong>RSS 源：</strong> {{ pendingTask.sourceName }}</p>
                  <p><strong>执行频率：</strong> {{ pendingTask.frequency }}</p>
                </div>
                <div class="flex gap-3 mt-4">
                  <button
                    @click="handleConfirmTask"
                    class="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium text-sm transition-colors"
                  >
                    确认创建
                  </button>
                  <button
                    @click="handleRejectTask"
                    class="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded-lg font-medium text-sm transition-colors"
                  >
                    重新调整
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 输入区域 -->
          <div class="border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-6">
            <div class="max-w-3xl mx-auto flex gap-3">
              <input
                v-model="userInput"
                @keyup.enter="handleSendMessage"
                :disabled="isWaitingForAI || showConfirmation"
                type="text"
                placeholder="输入你的需求..."
                class="flex-1 px-4 py-3 rounded-xl border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-600 focus:border-transparent outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              />
              <button
                @click="handleSendMessage"
                :disabled="!userInput.trim() || isWaitingForAI || showConfirmation"
                class="px-6 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white rounded-xl font-medium transition-colors flex items-center gap-2"
              >
                <span class="material-icons-outlined text-lg">send</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<script setup>
import { ref } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import { useAPI } from '@/composables/useAPI'
import { useToast } from '@/composables/useToast'

const taskStore = useTaskStore()
const { chat: chatAPI } = useAPI()
const { show: showToast } = useToast()

// 对话消息列表
const messages = ref([])

// 用户输入
const userInput = ref('')

// 是否等待 AI 响应
const isWaitingForAI = ref(false)

// 是否显示确认对话框
const showConfirmation = ref(false)

// 待创建的任务信息
const pendingTask = ref({
  name: '',
  sourceName: '',
  frequency: 'daily'
})

/**
 * 发送消息给 AI
 */
const handleSendMessage = async () => {
  if (!userInput.value.trim()) return

  // 添加用户消息
  messages.value.push({
    role: 'user',
    content: userInput.value
  })

  const userMessage = userInput.value
  userInput.value = ''
  isWaitingForAI.value = true

  try {
    // 调用 AI 接口处理任务创建对话
    // 这里需要实现一个新的 API 端点
    const response = await chatAPI.createTaskWithAI(
      userMessage,
      messages.value
    )

    // 添加 AI 回复
    messages.value.push({
      role: 'ai',
      content: response.reply
    })

    // 如果 AI 返回了任务确认信息，显示确认框
    if (response.pendingTask) {
      pendingTask.value = response.pendingTask
      showConfirmation.value = true
    }
  } catch (error) {
    console.error('[AI Assistant] 错误:', error)
    messages.value.push({
      role: 'ai',
      content: '抱歉，出现了错误。请重新描述你的需求。'
    })
  } finally {
    isWaitingForAI.value = false
  }
}

/**
 * 确认创建任务
 */
const handleConfirmTask = async () => {
  try {
    // 创建任务
    await taskStore.createTask({
      ...pendingTask.value,
      command: `监控 ${pendingTask.value.sourceName} 源`
    })

    showToast(`任务 "${pendingTask.value.name}" 创建成功！`, 'success', 2000)
    taskStore.closeAIAssistantModal()
  } catch (error) {
    console.error('[AI Assistant] 创建任务失败:', error)
    showToast('创建任务失败，请重试', 'error', 2000)
  }
}

/**
 * 拒绝任务，继续对话
 */
const handleRejectTask = () => {
  showConfirmation.value = false
  messages.value.push({
    role: 'ai',
    content: '没问题，请告诉我你需要调整什么。'
  })
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 20px;
}

.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.5);
}
</style>
