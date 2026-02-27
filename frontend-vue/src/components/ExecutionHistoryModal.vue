<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
    leave-active-class="transition-all duration-300 ease-out"
  >
    <div v-if="true" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center p-4">
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div class="bg-white dark:bg-[#1F2937] rounded-xl shadow-lg p-6 w-full max-w-2xl mx-4 max-h-[80vh] overflow-y-auto">
          <!-- 标题 -->
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-lg font-bold text-[#111827] dark:text-white flex items-center gap-2">
              <span class="material-icons-outlined">history</span>
              执行历史
            </h3>
            <button
              @click="handleClose"
              class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            >
              <span class="material-icons-outlined">close</span>
            </button>
          </div>

          <!-- 加载状态 -->
          <div v-if="taskStore.isLoadingExecutionHistory">
            <SkeletonLoader :count="5" height="44px" />
          </div>

          <!-- 历史记录列表 -->
          <div v-else-if="taskStore.executionHistory.length > 0" class="space-y-3">
            <!-- 统计信息 -->
            <div class="grid grid-cols-3 gap-3 mb-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div class="text-center">
                <p class="text-sm text-gray-600 dark:text-gray-400">总执行次数</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ taskStore.executionHistory.length }}</p>
              </div>
              <div class="text-center">
                <p class="text-sm text-gray-600 dark:text-gray-400">成功次数</p>
                <p class="text-2xl font-bold text-green-600 dark:text-green-400">
                  {{ successCount }}
                </p>
              </div>
              <div class="text-center">
                <p class="text-sm text-gray-600 dark:text-gray-400">失败次数</p>
                <p class="text-2xl font-bold text-red-600 dark:text-red-400">
                  {{ failureCount }}
                </p>
              </div>
            </div>

            <!-- 执行记录表格 -->
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-200 dark:border-gray-700">
                    <th class="text-left py-3 px-4 font-semibold text-gray-700 dark:text-gray-300">时间</th>
                    <th class="text-left py-3 px-4 font-semibold text-gray-700 dark:text-gray-300">状态</th>
                    <th class="text-left py-3 px-4 font-semibold text-gray-700 dark:text-gray-300">耗时</th>
                    <th class="text-left py-3 px-4 font-semibold text-gray-700 dark:text-gray-300">获取数</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="record in taskStore.executionHistory"
                    :key="record.id"
                    class="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
                  >
                    <!-- 时间 -->
                    <td class="py-3 px-4 text-gray-700 dark:text-gray-300">
                      {{ formatDateTime(record.timestamp) }}
                    </td>

                    <!-- 状态 -->
                    <td class="py-3 px-4">
                      <span
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium"
                        :class="record.status === 'success'
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                          : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'"
                      >
                        <span class="material-icons-outlined text-xs">
                          {{ record.status === 'success' ? 'check_circle' : 'error' }}
                        </span>
                        {{ record.status === 'success' ? '成功' : '失败' }}
                      </span>
                    </td>

                    <!-- 耗时 -->
                    <td class="py-3 px-4 text-gray-700 dark:text-gray-300">
                      {{ record.duration.toFixed(2) }}s
                    </td>

                    <!-- 获取数 -->
                    <td class="py-3 px-4 text-gray-700 dark:text-gray-300">
                      {{ record.itemsCount }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- 提示信息 -->
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-4 p-3 bg-gray-50 dark:bg-gray-800 rounded">
              <span class="material-icons-outlined text-sm align-middle">info</span>
              最多显示最近 20 条执行记录，更早的记录自动归档。
            </div>

          </div>

          <!-- 空状态 -->
          <div v-else class="text-center py-8">
            <span class="material-icons-outlined text-4xl text-gray-300 dark:text-gray-700 mx-auto block mb-3">
              history
            </span>
            <p class="text-gray-500 dark:text-gray-400">暂无执行历史</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">点击"执行"按钮开始执行任务</p>
          </div>

          <!-- 关闭按钮 -->
          <div class="mt-6 flex gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
            <button
              @click="handleClose"
              class="flex-1 px-4 py-2.5 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-white rounded-lg text-sm font-medium transition-colors"
            >
              关闭
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import SkeletonLoader from './SkeletonLoader.vue'

const taskStore = useTaskStore()

/**
 * 计算成功次数
 */
const successCount = computed(() => {
  return taskStore.executionHistory.filter(h => h.status === 'success').length
})

/**
 * 计算失败次数
 */
const failureCount = computed(() => {
  return taskStore.executionHistory.filter(h => h.status === 'error').length
})

/**
 * 格式化日期时间
 */
const formatDateTime = (isoString) => {
  const date = new Date(isoString)
  return date.toLocaleString('zh-CN')
}

/**
 * 处理关闭
 */
const handleClose = () => {
  taskStore.closeExecutionHistoryModal()
}
</script>

<style scoped>
/* 表格行间距 */
table tbody tr {
  height: 44px;
}

/* 平滑滚动 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 20px;
}

.dark ::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.5);
}
</style>
