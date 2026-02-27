<template>
  <!-- 主容器：左侧列表 + 右侧对话区 -->
  <main class="flex-grow px-8 py-6 h-[calc(100vh-80px)] relative overflow-hidden">
    <div class="flex h-full gap-6">
      <!-- 左侧：任务列表 (固定宽度 320px) -->
      <TaskSidebar />

      <!-- 右侧：对话区域 (自适应宽度) -->
      <TaskChat />
    </div>

    <!-- 创建任务模态框 -->
    <TaskModal />
  </main>
</template>

<script setup>
import { onMounted } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import TaskSidebar from './TaskSidebar.vue'
import TaskChat from './TaskChat.vue'
import TaskModal from './TaskModal.vue'

const taskStore = useTaskStore()

// 组件挂载时加载任务列表
onMounted(() => {
  taskStore.loadTasks()
})
</script>

<style scoped>
/* 添加平滑的过渡效果 */
:deep(*) {
  transition-colors: all 200ms;
}
</style>
