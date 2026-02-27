<template>
  <main class="flex-grow px-8 py-6 h-[calc(100vh-80px)] relative pt-20">
    <div class="flex h-full gap-6">
      <!-- 左侧任务列表 -->
      <aside class="w-1/3 min-w-[300px] flex flex-col bg-[#F8F9FA] dark:bg-[#1F2937] rounded-xl p-6 shadow-sm overflow-hidden relative border border-gray-100 dark:border-gray-700/30">
        <h2 class="text-xl font-bold mb-6 text-gray-900 dark:text-gray-100">任务列表</h2>

        <!-- 添加任务按钮 -->
        <button class="w-full py-3 mb-6 bg-white border border-gray-200 hover:bg-gray-50 dark:bg-gray-700 dark:border-transparent dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded-lg flex items-center justify-center gap-2 transition-colors font-medium shadow-sm">
          <span class="material-icons-outlined text-lg">add_circle_outline</span>
          <span>添加任务</span>
        </button>

        <!-- 任务列表 -->
        <div class="flex flex-col gap-3 overflow-y-auto pr-2 custom-scrollbar">
          <button
            v-for="task in taskStore.tasks"
            :key="task.id"
            @click="taskStore.switchTask(task.id)"
            :class="[
              'p-4 rounded-lg cursor-pointer border-l-4 transition-all shadow-sm text-left',
              taskStore.activeTaskId === task.id
                ? 'bg-[#E5E7EB] dark:bg-[#4B5563] border-gray-800 dark:border-gray-200'
                : 'bg-white dark:bg-[#374151] hover:bg-gray-50 dark:hover:bg-[#4B5563]/80 border-gray-200 dark:border-transparent hover:border-gray-300 dark:hover:border-gray-500'
            ]"
          >
            <h3 :class="['font-medium', taskStore.activeTaskId === task.id ? 'text-gray-900 dark:text-white' : 'text-gray-700 dark:text-gray-300']">
              {{ task.name }}
            </h3>
          </button>
        </div>
      </aside>

      <!-- 右侧对话框 -->
      <section class="flex-1 bg-white dark:bg-[#111827] rounded-xl border border-gray-200 dark:border-gray-700/50 shadow-sm flex flex-col overflow-hidden relative">
        <!-- 消息容器 -->
        <div class="flex-1 p-8 overflow-y-auto space-y-8" ref="messagesContainer">
          <div
            v-for="msg in taskStore.messages"
            :key="msg.id"
            :class="['flex gap-4 animate-slide-in', msg.role === 'user' ? 'justify-end' : 'justify-start']"
          >
            <!-- AI 头像 -->
            <div v-if="msg.role === 'ai'" class="flex-shrink-0 w-8 h-8 rounded-full bg-black dark:bg-indigo-600 flex items-center justify-center">
              <span class="material-icons-outlined text-sm text-white">smart_toy</span>
            </div>

            <!-- 消息内容 -->
            <div v-if="msg.type === 'error'" class="flex flex-col space-y-3 max-w-xl">
              <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 flex items-start gap-3">
                <span class="material-icons-outlined text-red-600 dark:text-red-400 flex-shrink-0">error</span>
                <div class="flex-1">
                  <p class="text-sm font-medium text-red-900 dark:text-red-100">{{ msg.content }}</p>
                </div>
              </div>
              <button
                @click="taskStore.sendMessage(lastUserMessage)"
                class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors self-start"
              >
                重试
              </button>
            </div>
            <div v-else class="flex flex-col space-y-1 max-w-3xl">
              <span class="font-bold text-gray-900 dark:text-gray-300 text-sm">{{ msg.role === 'user' ? 'User:' : 'AI:' }}</span>
              <p v-if="msg.role === 'user'" class="text-[#111827] dark:text-gray-100 leading-relaxed text-lg whitespace-pre-wrap">
                {{ msg.content }}
              </p>
              <p v-else class="text-[#111827] dark:text-gray-100 leading-relaxed text-lg whitespace-pre-wrap">
                {{ msg.content }}
              </p>
            </div>

            <!-- 用户头像 -->
            <div v-if="msg.role === 'user'" class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center border border-gray-200 dark:border-transparent">
              <span class="material-icons-outlined text-sm text-gray-600 dark:text-gray-300">person</span>
            </div>
          </div>

          <!-- 正在加载的AI消息 -->
          <div v-if="taskStore.isLoading" class="flex gap-4 animate-slide-in">
            <div class="flex-shrink-0 w-8 h-8 rounded-full bg-black dark:bg-indigo-600 flex items-center justify-center">
              <span class="material-icons-outlined text-sm text-white">smart_toy</span>
            </div>
            <div class="flex flex-col space-y-1 max-w-3xl">
              <span class="font-bold text-gray-900 dark:text-gray-300 text-sm">AI:</span>
              <div class="text-gray-500 dark:text-gray-400 text-lg flex items-center h-7 space-x-1">
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
              </div>
            </div>
          </div>
        </div>

        <!-- 输入框 -->
        <div class="p-4 border-t border-gray-100 dark:border-gray-800 bg-white dark:bg-[#111827]">
          <div class="relative flex items-center w-full">
            <textarea
              v-model="inputText"
              @keydown.enter.exact="handleSendMessage"
              @keydown.shift.enter="insertNewline"
              class="w-full pl-5 pr-14 py-4 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-[#1F2937] text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-indigo-500/50 focus:border-gray-300 dark:focus:border-transparent transition-all shadow-sm resize-none"
              placeholder="输入消息... (Shift+Enter 换行)"
              rows="1"
            ></textarea>
            <button
              @click="handleSendMessage"
              :disabled="taskStore.isLoading || !inputText || !inputText.toString().trim()"
              class="absolute right-2 p-2 bg-gray-100 hover:bg-gray-200 dark:bg-[#374151] dark:hover:bg-[#4B5563] rounded-lg text-gray-600 dark:text-white transition-colors flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span class="material-icons-outlined transform -rotate-45" style="font-size: 20px;">send</span>
            </button>
          </div>
        </div>
      </section>
    </div>
  </main>
</template>

<script setup>
import { ref, nextTick, onMounted } from 'vue'
import { useTaskStore } from '@/stores'

const taskStore = useTaskStore()
const inputText = ref('')
const messagesContainer = ref(null)
const lastUserMessage = ref('')

const handleSendMessage = async (e) => {
  if (e) {
    e.preventDefault()
  }

  if (!inputText.value || !inputText.value.toString().trim() || taskStore.isLoading) {
    return
  }

  lastUserMessage.value = inputText.value
  await taskStore.sendMessage(inputText.value)
  inputText.value = ''

  // 滚动到底部
  await nextTick()
  scrollToBottom()
}

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

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

onMounted(() => {
  // 初始化任务
  taskStore.messages = []
  scrollToBottom()
})
</script>

<style scoped>
.typing-dot {
  width: 6px;
  height: 6px;
  background-color: currentColor;
  border-radius: 50%;
  display: inline-block;
  animation: bounce 1.4s infinite ease-in-out both;
}

.typing-dot:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-dot:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}
</style>
