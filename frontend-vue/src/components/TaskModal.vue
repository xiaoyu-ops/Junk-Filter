<template>
  <!-- 模态框外层遮罩 + 背景模糊 -->
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
    leave-active-class="transition-all duration-300 ease-out"
  >
    <div
      v-if="taskStore.isModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-md p-4 transition-opacity"
    >
      <!-- 模态框容器：缩放动画 -->
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div v-if="taskStore.isModalOpen" class="bg-white dark:bg-gray-800 w-full max-w-2xl rounded-3xl shadow-2xl border border-gray-200 dark:border-gray-700 flex flex-col overflow-hidden">
          <!-- 标题区 -->
          <div class="px-8 py-6 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center bg-white dark:bg-gray-800">
            <div>
              <h2 class="text-2xl font-bold text-gray-900 dark:text-white">创建新分发任务</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">使用自然语言描述您的任务需求</p>
            </div>
            <button
              @click="taskStore.closeModal"
              class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors p-1"
            >
              <span class="material-icons-outlined">close</span>
            </button>
          </div>

          <!-- 表单内容 -->
          <form @submit.prevent="handleSubmit" class="p-8 space-y-6 overflow-y-auto">
            <!-- 任务名称 -->
            <div class="space-y-2">
              <label for="taskName" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                任务名称
              </label>
              <input
                id="taskName"
                v-model="taskStore.taskForm.name"
                type="text"
                placeholder="例如：Twitter AI 新闻早报"
                class="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-gray-900 dark:focus:ring-white focus:border-transparent transition-shadow placeholder-gray-400 dark:placeholder-gray-500"
              />
            </div>

            <!-- 任务指令 -->
            <div class="space-y-2">
              <label for="taskCommand" class="block text-sm font-medium text-gray-700 dark:text-gray-300 flex justify-between">
                <span>任务指令</span>
                <span class="text-xs text-gray-500 font-normal">支持自然语言</span>
              </label>
              <div class="relative">
                <textarea
                  id="taskCommand"
                  v-model="taskStore.taskForm.command"
                  placeholder="例如：每天早上9点总结Twitter上关于AI的新闻，并发送到我的邮箱。"
                  rows="5"
                  class="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-gray-900 dark:focus:ring-white focus:border-transparent transition-shadow resize-none placeholder-gray-400 dark:placeholder-gray-500 leading-relaxed"
                ></textarea>
                <!-- Sparkles 图标 -->
                <div class="absolute bottom-3 right-3 text-gray-400 dark:text-gray-500 pointer-events-none">
                  <span class="material-icons-outlined text-lg">auto_awesome</span>
                </div>
              </div>
            </div>

            <!-- 高级设置 (折叠菜单) -->
            <div class="pt-2">
              <details class="group">
                <summary class="flex items-center gap-2 cursor-pointer text-sm font-medium text-gray-600 dark:text-gray-300 hover:text-primary dark:hover:text-white transition-colors select-none">
                  <span class="material-icons-outlined transition-transform group-open:rotate-90 text-lg">chevron_right</span>
                  <span>高级设置 (调度与通知)</span>
                </summary>

                <div class="pl-7 pt-4 grid grid-cols-2 gap-4">
                  <!-- 执行频率 -->
                  <div>
                    <label for="frequency" class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
                      执行频率
                    </label>
                    <select
                      id="frequency"
                      v-model="taskStore.taskForm.frequency"
                      class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm py-2 px-3 focus:outline-none focus:ring-2 focus:ring-gray-900 dark:focus:ring-white"
                    >
                      <option value="daily">每日</option>
                      <option value="weekly">每周</option>
                      <option value="hourly">每小时</option>
                      <option value="once">仅一次</option>
                    </select>
                  </div>

                  <!-- 执行时间 -->
                  <div>
                    <label for="executionTime" class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
                      执行时间
                    </label>
                    <input
                      id="executionTime"
                      v-model="taskStore.taskForm.execution_time"
                      type="time"
                      class="w-full rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm py-2 px-3 focus:outline-none focus:ring-2 focus:ring-gray-900 dark:focus:ring-white"
                    />
                  </div>

                  <!-- 通知渠道 -->
                  <div class="col-span-2">
                    <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">
                      通知渠道
                    </label>
                    <div class="flex gap-3">
                      <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-700 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600">
                        <input
                          v-model="taskStore.taskForm.notification_channels"
                          type="checkbox"
                          value="email"
                          class="rounded text-gray-900 dark:text-white focus:ring-gray-900 dark:focus:ring-white border-gray-300 dark:border-gray-500 bg-transparent"
                        />
                        <span>Email</span>
                      </label>
                      <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-700 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600">
                        <input
                          v-model="taskStore.taskForm.notification_channels"
                          type="checkbox"
                          value="slack"
                          class="rounded text-gray-900 dark:text-white focus:ring-gray-900 dark:focus:ring-white border-gray-300 dark:border-gray-500 bg-transparent"
                        />
                        <span>Slack</span>
                      </label>
                      <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-700 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600">
                        <input
                          v-model="taskStore.taskForm.notification_channels"
                          type="checkbox"
                          value="telegram"
                          class="rounded text-gray-900 dark:text-white focus:ring-gray-900 dark:focus:ring-white border-gray-300 dark:border-gray-500 bg-transparent"
                        />
                        <span>Telegram</span>
                      </label>
                    </div>
                  </div>
                </div>
              </details>
            </div>
          </form>

          <!-- 按钮区域 -->
          <div class="bg-gray-50 dark:bg-gray-900/50 px-8 py-5 flex items-center justify-end gap-3 border-t border-gray-200 dark:border-gray-700">
            <button
              @click="taskStore.closeModal"
              type="button"
              class="px-5 py-2.5 rounded-lg text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
            >
              取消
            </button>
            <button
              @click="handleSubmit"
              :disabled="taskStore.isCreating"
              class="px-5 py-2.5 rounded-lg text-sm font-medium bg-gray-900 dark:bg-white text-white dark:text-gray-900 shadow-lg hover:bg-gray-800 dark:hover:bg-gray-200 disabled:opacity-60 disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              <span v-if="!taskStore.isCreating">立即创建</span>
              <span v-else>创建中...</span>
              <span v-if="!taskStore.isCreating" class="material-icons-outlined text-sm">arrow_forward</span>
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<script setup>
import { useTaskStore } from '@/stores/useTaskStore'
import { useToast } from '@/composables/useToast'

const taskStore = useTaskStore()
const { show: showToast } = useToast()

const handleSubmit = async () => {
  // 表单验证
  if (!taskStore.taskForm.name.trim()) {
    showToast('请输入任务名称', 'error', 2000)
    return
  }
  if (!taskStore.taskForm.command.trim()) {
    showToast('请输入任务指令', 'error', 2000)
    return
  }

  // 创建任务
  const newTask = await taskStore.createTask()
  if (newTask) {
    showToast(`任务 "${newTask.name}" 创建成功！`, 'success', 2000)
  }
}
</script>

<style scoped>
/* 自定义 checkbox 样式 */
input[type="checkbox"] {
  cursor: pointer;
}

/* 详情元素箭头旋转动画 */
details summary .material-icons-outlined {
  display: inline-block;
  transition: transform 0.3s ease-out;
}

details[open] summary .material-icons-outlined {
  transform: rotate(90deg);
}
</style>
