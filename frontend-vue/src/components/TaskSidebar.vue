<template>
  <aside class="w-80 flex flex-col bg-sidebar-light dark:bg-[#1F2937] rounded-xl p-6 shadow-sm overflow-hidden relative border border-gray-100 dark:border-gray-700/30">
    <!-- 标题 -->
    <h2 class="text-xl font-bold mb-6 text-gray-900 dark:text-gray-100">任务列表</h2>

    <!-- 添加任务按钮 -->
    <button
      @click="taskStore.openModal"
      class="w-full py-3 mb-6 bg-white border border-gray-200 hover:bg-gray-50 dark:bg-gray-700 dark:border-transparent dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded-lg flex items-center justify-center gap-2 transition-colors font-medium shadow-sm group"
    >
      <span class="material-icons-outlined text-lg group-hover:text-primary dark:group-hover:text-white transition-colors">add_circle_outline</span>
      <span>添加任务</span>
    </button>

    <!-- 任务列表加载状态 -->
    <div v-if="taskStore.isLoading" class="flex-1">
      <SkeletonLoader :count="4" height="80px" />
    </div>

    <!-- 任务列表 -->
    <div v-else-if="taskStore.tasks.length > 0" class="flex flex-col gap-3 overflow-y-auto pr-2 custom-scrollbar flex-1">
      <template v-for="task in taskStore.tasks" :key="task.id">
        <div
          class="transition-all"
          :class="taskStore.selectedTaskId === task.id ? 'opacity-100' : 'opacity-75 hover:opacity-100'"
        >
          <!-- 任务主行 -->
          <div
            @click="taskStore.selectTask(task.id)"
            :class="[
              'p-4 rounded-lg cursor-pointer border-l-4 transition-all shadow-sm',
              taskStore.selectedTaskId === task.id
                ? 'bg-gray-200 dark:bg-gray-600 border-l-gray-900 dark:border-l-white'
                : 'bg-white dark:bg-gray-700/60 border-l-transparent hover:bg-gray-50 dark:hover:bg-gray-700 border border-gray-200 dark:border-transparent'
            ]"
          >
            <h3 class="font-medium text-gray-700 dark:text-gray-300">{{ task.name }}</h3>
            <p v-if="task.frequency" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ taskStore.getFrequencyLabel(task.frequency) }}
            </p>
          </div>

          <!-- 执行按钮栏（仅选中任务时显示） -->
          <div
            v-if="taskStore.selectedTaskId === task.id"
            class="mt-3 flex gap-2 px-4"
          >
            <!-- 执行按钮 -->
            <button
              @click.stop="handleExecuteTask(task.id)"
              :disabled="taskStore.isTaskExecuting(task.id)"
              class="flex-1 px-3 py-2 rounded-md bg-blue-600 hover:bg-blue-700 active:scale-95 disabled:bg-gray-400 disabled:cursor-not-allowed disabled:opacity-60 text-white text-xs font-medium transition-colors flex items-center justify-center gap-1"
            >
              <span class="material-icons-outlined text-sm" :class="taskStore.isTaskExecuting(task.id) ? 'animate-spin' : ''">
                {{ taskStore.isTaskExecuting(task.id) ? 'hourglass_top' : 'play_arrow' }}
              </span>
              <span>{{ taskStore.isTaskExecuting(task.id) ? '执行中' : '执行' }}</span>
            </button>

            <!-- 历史按钮 -->
            <button
              @click.stop="handleOpenHistory(task.id)"
              class="flex-1 px-3 py-2 rounded-md bg-gray-600 hover:bg-gray-700 dark:bg-gray-700 dark:hover:bg-gray-600 text-white text-xs font-medium transition-colors flex items-center justify-center gap-1"
            >
              <span class="material-icons-outlined text-sm">history</span>
              <span>历史</span>
            </button>
          </div>

          <!-- 执行进度条（仅执行中时显示） -->
          <div
            v-if="taskStore.isTaskExecuting(task.id)"
            class="mt-2 px-4"
          >
            <div class="w-full bg-gray-300 dark:bg-gray-600 rounded-full h-1.5 overflow-hidden">
              <div
                class="bg-blue-500 h-full transition-all duration-300"
                :style="{ width: taskStore.getExecutionProgress() + '%' }"
              ></div>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 text-center">
              {{ Math.round(taskStore.getExecutionProgress()) }}%
            </p>
          </div>

          <!-- 最近执行记录（可选） -->
          <div
            v-if="taskStore.executionHistory.length > 0"
            class="mt-2 px-4 text-xs"
          >
            <div
              class="py-1.5 px-2 rounded bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700"
            >
              <div class="flex items-center gap-1.5">
                <span
                  class="material-icons-outlined text-xs"
                  :class="taskStore.executionHistory[0].status === 'success'
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-red-600 dark:text-red-400'"
                >
                  {{ taskStore.executionHistory[0].status === 'success' ? 'check_circle' : 'error' }}
                </span>
                <span class="text-gray-600 dark:text-gray-400">
                  {{ taskStore.executionHistory[0].status === 'success' ? '✓ 成功' : '✗ 失败' }}
                  - {{ taskStore.executionHistory[0].duration.toFixed(1) }}s
                </span>
              </div>
              <p class="text-gray-500 dark:text-gray-500 mt-0.5">
                {{ new Date(taskStore.executionHistory[0].timestamp).toLocaleTimeString('zh-CN') }}
              </p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- 空状态 -->
    <EmptyState
      v-else
      icon="inbox"
      title="还没有任务"
      subtitle="从左边的菜单开始创建任务"
      action="创建任务"
      actionIcon="add_circle"
      @action="taskStore.openModal"
    />

    <!-- 执行历史 Modal -->
    <ExecutionHistoryModal
      v-if="taskStore.showExecutionHistoryModal"
      @close="taskStore.closeExecutionHistoryModal"
    />
  </aside>
</template>

<script setup>
import { useTaskStore } from '@/stores/useTaskStore'
import { useToast } from '@/composables/useToast'
import ExecutionHistoryModal from './ExecutionHistoryModal.vue'
import SkeletonLoader from './SkeletonLoader.vue'
import EmptyState from './EmptyState.vue'

const taskStore = useTaskStore()
const { show: showToast } = useToast()

/**
 * 处理任务执行
 */
const handleExecuteTask = async (taskId) => {
  try {
    await taskStore.executeTask(taskId)
  } catch (error) {
    console.error('执行任务出错:', error)
  }
}

/**
 * 打开执行历史
 */
const handleOpenHistory = async (taskId) => {
  taskStore.openExecutionHistoryModal()
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

