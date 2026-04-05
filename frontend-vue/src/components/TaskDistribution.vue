<template>
  <!-- 主容器：三栏布局 - 任务列表 | 子对话列表 | 对话区 -->
  <main class="flex-grow px-8 py-6 h-[calc(100vh-80px)] relative overflow-hidden">
    <div class="flex h-full gap-4">
      <!-- 左栏：任务列表 (固定宽度 280px) -->
      <TaskSidebar />

      <!-- 中栏：子对话列表 (可折叠) -->
      <div class="relative flex-shrink-0 thread-panel" :class="threadCollapsed ? 'w-0' : 'w-64'">
        <div v-show="!threadCollapsed" class="h-full">
          <ThreadList />
        </div>
        <!-- 折叠/展开按钮 -->
        <button
          @click="threadCollapsed = !threadCollapsed"
          class="absolute -right-3 top-1/2 -translate-y-1/2 z-10 w-6 h-12 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-full shadow-sm flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
          :title="threadCollapsed ? '展开子对话' : '收起子对话'"
        >
          <span class="material-icons-outlined text-sm text-gray-500 dark:text-gray-400">
            {{ threadCollapsed ? 'chevron_right' : 'chevron_left' }}
          </span>
        </button>
      </div>

      <!-- 右栏：对话区域 (自适应宽度) -->
      <TaskChat />
    </div>

    <!-- 创建任务模态框 -->
    <TaskModal />

    <!-- AI 助手模态框 -->
    <AIAssistantModal />
  </main>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import TaskSidebar from './TaskSidebar.vue'
import ThreadList from './ThreadList.vue'
import TaskChat from './TaskChat.vue'
import TaskModal from './TaskModal.vue'
import AIAssistantModal from './AIAssistantModal.vue'

const taskStore = useTaskStore()
const threadCollapsed = ref(false)

// 组件挂载时加载任务列表
onMounted(() => {
  taskStore.loadTasks()
})
</script>

<style scoped>
.thread-panel {
  transition: width 0.2s ease;
}
</style>
