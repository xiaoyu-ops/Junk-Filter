import { defineStore } from 'pinia'
import { ref, computed, onMounted } from 'vue'
import { useAPI } from '@/composables/useAPI'
import { useToast } from '@/composables/useToast'

export const useTaskStore = defineStore('task', () => {
  // ==================== Composables ====================

  const { tasks: tasksAPI } = useAPI()
  const { show: showToast } = useToast()

  // ==================== State ====================

  // 任务列表
  const tasks = ref([])

  // 加载状态
  const isLoading = ref(false)

  // 选中的任务 ID
  const selectedTaskId = ref(null)

  // 模态框状态
  const isModalOpen = ref(false)

  // AI 助手模态框状态
  const isAIAssistantModalOpen = ref(false)

  // 创建任务的表单状态
  const taskForm = ref({
    name: '',
    command: '',
    frequency: 'daily',
    execution_time: '09:00',
    notification_channels: ['email']
  })

  // 创建任务的加载状态
  const isCreating = ref(false)

  // ==================== 执行管理状态 ====================

  // 执行中的任务 ID
  const executingTaskId = ref(null)

  // 执行进度 (0-100)
  const executionProgress = ref(0)

  // 执行历史
  const executionHistory = ref([])

  // 执行历史加载状态
  const isLoadingExecutionHistory = ref(false)

  // 是否显示执行历史modal
  const showExecutionHistoryModal = ref(false)

  // 获取选中的任务
  const selectedTask = computed(() =>
    tasks.value.find(t => t.id === selectedTaskId.value) || null
  )

  // 检查任务列表是否为空
  const hasTask = computed(() => tasks.value.length > 0)

  // ==================== Actions ====================

  // 打开模态框
  const openModal = () => {
    isModalOpen.value = true
    resetForm()
  }

  // 关闭模态框
  const closeModal = () => {
    isModalOpen.value = false
    resetForm()
  }

  // 打开 AI 助手模态框
  const openAIAssistantModal = () => {
    isAIAssistantModalOpen.value = true
    resetForm()
  }

  // 关闭 AI 助手模态框
  const closeAIAssistantModal = () => {
    isAIAssistantModalOpen.value = false
    resetForm()
  }

  // 选中任务
  const selectTask = (taskId) => {
    selectedTaskId.value = taskId
    // 切换任务时重置子对话选择
    selectedThreadId.value = null
    threads.value = []
    if (taskId) {
      loadThreads(taskId)
    }
  }

  // ==================== 子对话管理 ====================

  const threads = ref([])
  const selectedThreadId = ref(null)
  const isLoadingThreads = ref(false)

  const selectedThread = computed(() =>
    threads.value.find(t => t.id === selectedThreadId.value) || null
  )

  const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
  const apiBase = API_BASE.endsWith('/api') ? API_BASE : `${API_BASE}/api`

  // 从前端 taskId (如 "source-6") 提取真实数字 ID
  const extractSourceId = (taskId) => {
    if (typeof taskId === 'string' && taskId.startsWith('source-')) {
      return taskId.replace('source-', '')
    }
    return taskId
  }

  const loadThreads = async (taskId) => {
    isLoadingThreads.value = true
    try {
      const realId = extractSourceId(taskId)
      const response = await fetch(`${apiBase}/tasks/${realId}/threads`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const data = await response.json()
      threads.value = data.data || []
    } catch (error) {
      console.error('Failed to load threads:', error)
      threads.value = []
    } finally {
      isLoadingThreads.value = false
    }
  }

  const createThread = async (taskId, title) => {
    try {
      const realId = extractSourceId(taskId)
      const response = await fetch(`${apiBase}/tasks/${realId}/threads`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title }),
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const thread = await response.json()
      threads.value.unshift(thread)
      selectedThreadId.value = thread.id
      return thread
    } catch (error) {
      console.error('Failed to create thread:', error)
      showToast('创建子对话失败', 'error')
      throw error
    }
  }

  const deleteThread = async (threadId) => {
    try {
      const response = await fetch(`${apiBase}/threads/${threadId}`, { method: 'DELETE' })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      threads.value = threads.value.filter(t => t.id !== threadId)
      if (selectedThreadId.value === threadId) {
        selectedThreadId.value = null
      }
    } catch (error) {
      console.error('Failed to delete thread:', error)
      showToast('删除子对话失败', 'error')
    }
  }

  const selectThread = (threadId) => {
    selectedThreadId.value = threadId
  }

  // 重置表单
  const resetForm = () => {
    taskForm.value = {
      name: '',
      command: '',
      frequency: 'daily',
      execution_time: '09:00',
      notification_channels: ['email']
    }
  }

  // 加载任务列表
  const loadTasks = async () => {
    isLoading.value = true
    try {
      tasks.value = await tasksAPI.list()
      // 自动选中第一个任务
      if (tasks.value.length > 0 && !selectedTaskId.value) {
        selectedTaskId.value = tasks.value[0].id
      }
    } catch (error) {
      console.error('加载任务列表失败:', error)
      showToast('加载任务失败，请重试', 'error')
    } finally {
      isLoading.value = false
    }
  }

  // 创建新任务
  const createTask = async () => {
    // 表单验证
    if (!taskForm.value.name.trim()) {
      showToast('任务名称不能为空', 'error')
      return
    }
    if (!taskForm.value.command.trim()) {
      showToast('任务指令不能为空', 'error')
      return
    }

    isCreating.value = true

    try {
      // 调用 API 创建任务
      const newTask = await tasksAPI.create({
        name: taskForm.value.name,
        command: taskForm.value.command,
        frequency: taskForm.value.frequency,
        execution_time: taskForm.value.execution_time,
        notification_channels: taskForm.value.notification_channels
      })

      // 添加到本地列表
      tasks.value.push(newTask)

      // 自动选中新任务
      selectedTaskId.value = newTask.id

      // 关闭模态框
      closeModal()

      // 显示成功提示
      showToast('任务创建成功', 'success')

      return newTask
    } catch (error) {
      console.error('创建任务失败:', error)
      showToast('创建任务失败，请重试', 'error')
      throw error
    } finally {
      isCreating.value = false
    }
  }

  // 删除任务
  const deleteTask = async (taskId) => {
    try {
      // 调用 API 删除任务
      await tasksAPI.delete(taskId)

      // 从列表中移除
      tasks.value = tasks.value.filter(t => t.id !== taskId)

      // 如果删除的是选中项，选择第一个
      if (selectedTaskId.value === taskId && tasks.value.length > 0) {
        selectedTaskId.value = tasks.value[0].id
      } else if (selectedTaskId.value === taskId) {
        selectedTaskId.value = null
      }

      // 显示成功提示
      showToast('任务已删除', 'success')
    } catch (error) {
      console.error('删除任务失败:', error)
      showToast('删除任务失败，请重试', 'error')
    }
  }

  // 更新任务状态
  const updateTaskStatus = (taskId, status) => {
    const task = tasks.value.find(t => t.id === taskId)
    if (task) {
      task.status = status
    }
  }

  // ==================== Helpers ====================

  // 将频率转换为 Cron 表达式
  const generateCronExpression = (frequency, time) => {
    const [hour, minute] = time.split(':')
    switch (frequency) {
      case 'daily':
        return `${minute} ${hour} * * *`
      case 'weekly':
        return `${minute} ${hour} * * 1` // 周一
      case 'hourly':
        return `${minute} * * * *`
      case 'once':
        return `${minute} ${hour} * * *`
      default:
        return `0 9 * * *`
    }
  }

  // 获取频率的显示文本
  const getFrequencyLabel = (frequency) => {
    const labels = {
      'daily': '每日',
      'weekly': '每周',
      'hourly': '每小时',
      'once': '仅一次'
    }
    return labels[frequency] || frequency
  }

  // ==================== 执行管理方法 ====================

  /**
   * 执行任务
   */
  const executeTask = async (taskId) => {
    if (executingTaskId.value) {
      showToast('有任务正在执行，请稍候', 'warning')
      return
    }

    const task = tasks.value.find(t => t.id === taskId)
    if (!task) {
      showToast('任务不存在', 'error')
      return
    }

    executingTaskId.value = taskId
    executionProgress.value = 0

    try {
      // 模拟进度更新
      const progressInterval = setInterval(() => {
        executionProgress.value = Math.min(executionProgress.value + Math.random() * 30, 90)
      }, 200)

      // 调用 API 执行任务
      const result = await tasksAPI.executeTask(taskId)

      // 停止进度更新
      clearInterval(progressInterval)
      executionProgress.value = 100

      // 添加到执行历史
      executionHistory.value.unshift({
        id: result.executionId,
        taskId: taskId,
        status: result.status,
        duration: result.duration,
        itemsCount: result.itemsCount,
        message: result.message,
        timestamp: result.timestamp,
      })

      // 显示结果
      const message = result.status === 'success'
        ? `✅ 任务执行成功，耗时 ${result.duration.toFixed(2)}s，获取 ${result.itemsCount} 条内容`
        : `❌ 任务执行失败: ${result.message}`

      showToast(message, result.status === 'success' ? 'success' : 'error', 3000)

      return result

    } catch (error) {
      console.error('执行任务失败:', error)
      showToast('任务执行失败，请重试', 'error')
      throw error

    } finally {
      executingTaskId.value = null
      executionProgress.value = 0
    }
  }

  /**
   * 加载执行历史
   */
  const loadExecutionHistory = async (taskId) => {
    isLoadingExecutionHistory.value = true
    try {
      executionHistory.value = await tasksAPI.getExecutionHistory(taskId)
    } catch (error) {
      console.error('加载执行历史失败:', error)
      showToast('加载执行历史失败', 'error')
    } finally {
      isLoadingExecutionHistory.value = false
    }
  }

  /**
   * 打开执行历史 modal
   */
  const openExecutionHistoryModal = async () => {
    if (selectedTaskId.value) {
      showExecutionHistoryModal.value = true
      await loadExecutionHistory(selectedTaskId.value)
    }
  }

  /**
   * 关闭执行历史 modal
   */
  const closeExecutionHistoryModal = () => {
    showExecutionHistoryModal.value = false
  }

  /**
   * 判断是否正在执行指定任务
   */
  const isTaskExecuting = (taskId) => {
    return executingTaskId.value === taskId
  }

  /**
   * 获取执行进度百分比
   */
  const getExecutionProgress = () => {
    return executingTaskId.value ? executionProgress.value : 0
  }

  return {
    // State
    tasks,
    selectedTaskId,
    isModalOpen,
    isAIAssistantModalOpen,
    taskForm,
    isCreating,
    isLoading,

    // Execution State
    executingTaskId,
    executionProgress,
    executionHistory,
    isLoadingExecutionHistory,
    showExecutionHistoryModal,

    // Computed
    selectedTask,
    hasTask,

    // Actions
    openModal,
    closeModal,
    openAIAssistantModal,
    closeAIAssistantModal,
    selectTask,
    resetForm,
    loadTasks,
    createTask,
    deleteTask,
    updateTaskStatus,

    // Execution Methods
    executeTask,
    loadExecutionHistory,
    openExecutionHistoryModal,
    closeExecutionHistoryModal,
    isTaskExecuting,
    getExecutionProgress,

    // Thread State
    threads,
    selectedThreadId,
    selectedThread,
    isLoadingThreads,

    // Thread Actions
    loadThreads,
    createThread,
    deleteThread,
    selectThread,

    // Helpers
    getFrequencyLabel
  }
})
