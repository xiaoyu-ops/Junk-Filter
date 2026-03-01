import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// API 基础 URL
const API_BASE_URL = 'http://localhost:8080/api'

// 频率转换函数
const frequencyToSeconds = (frequency) => {
  const map = {
    'hourly': 3600,
    '30min': 1800,
    '2hours': 7200,
    'daily': 86400,
  }
  return map[frequency] || 3600
}

const secondsToFrequency = (seconds) => {
  const map = {
    3600: 'hourly',
    1800: '30min',
    7200: '2hours',
    86400: 'daily',
  }
  return map[seconds] || 'hourly'
}

export const useConfigStore = defineStore('config', () => {
  // 状态：从环境变量初始化
  const apiKey = ref(import.meta.env.VITE_API_KEY || '')
  const apiUrl = ref(import.meta.env.VITE_API_URL || 'http://localhost:8000/api')
  const modelName = ref(import.meta.env.VITE_API_MODEL || 'gpt-4o')
  const baseUrl = ref(import.meta.env.VITE_BASE_URL || '')
  const temperature = ref(0.7)
  const topP = ref(0.9)
  const maxTokens = ref(2048)
  const isSaving = ref(false)
  const saveStatus = ref(null) // 'success' | 'error' | null

  // RSS源管理状态
  const sources = ref([])

  // 加载状态
  const isLoadingSources = ref(false)
  const sourceLoadError = ref(null)

  // 模态框状态
  const showAddRssModal = ref(false)
  const showAddModelModal = ref(false)

  // 表单数据
  const newRssForm = ref({
    name: '',
    url: '',
    frequency: 'hourly',
    filterRules: '',
  })

  const newModelForm = ref({
    name: '',
    provider: 'OpenAI',
    apiKey: '',
    baseUrl: '',
  })

  // 展开行追踪
  const expandedSourceIds = ref([])

  // 计算属性
  const isConfigValid = computed(() => apiKey.value.length > 0)

  // ============ API 调用方法 ============

  /**
   * 从 Go 后端加载所有 RSS 源
   */
  const loadSources = async () => {
    isLoadingSources.value = true
    sourceLoadError.value = null

    try {
      const response = await fetch(`${API_BASE_URL}/sources`)

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      const sourceList = Array.isArray(data) ? data : []

      // 转换 Go API 响应为前端格式
      sources.value = sourceList.map(source => ({
        id: source.id,
        name: source.author_name || '未命名',
        url: source.url,
        frequency: secondsToFrequency(source.fetch_interval_seconds),
        status: source.enabled ? 'active' : 'paused',
        statusClass: source.enabled
          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
          : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
        lastSyncTime: source.last_fetch_time || null,
        lastSyncStatus: 'success',
        syncLogs: [],
        filterRules: '',
      }))

      console.log('[Config Store] Loaded sources:', sources.value)
    } catch (error) {
      sourceLoadError.value = error.message
      console.error('[Config Store] Failed to load sources:', error)
      sources.value = []
    } finally {
      isLoadingSources.value = false
    }
  }

  /**
   * 初始化配置（从localStorage恢复）
   */
  const loadConfig = async () => {
    const saved = localStorage.getItem('config')
    if (saved) {
      try {
        const config = JSON.parse(saved)
        modelName.value = config.modelName || modelName.value
        baseUrl.value = config.baseUrl || baseUrl.value
        if (config.apiModel && !config.modelName) {
          modelName.value = config.apiModel
        }
        temperature.value = config.temperature || temperature.value
        topP.value = config.topP || topP.value
        maxTokens.value = config.maxTokens || maxTokens.value
      } catch (error) {
        console.error('Failed to load config:', error)
      }
    }

    // 同时加载 RSS 源
    await loadSources()
  }

  /**
   * 保存配置到 localStorage
   */
  const saveConfig = async () => {
    isSaving.value = true
    saveStatus.value = null

    try {
      // 模拟API请求延迟
      await new Promise(resolve => setTimeout(resolve, 500))

      // 保存到localStorage
      localStorage.setItem('config', JSON.stringify({
        modelName: modelName.value,
        baseUrl: baseUrl.value,
        temperature: temperature.value,
        topP: topP.value,
        maxTokens: maxTokens.value,
      }))
      saveStatus.value = 'success'
      return true
    } catch (error) {
      saveStatus.value = 'error'
      console.error('Config save error:', error)
      return false
    } finally {
      isSaving.value = false
    }
  }

  /**
   * 添加新 RSS 源到 Go 后端
   */
  const addSource = async () => {
    const requestBody = {
      url: newRssForm.value.url,
      author_name: newRssForm.value.name,
      platform: 'blog',
      priority: 5,
      enabled: true,
      fetch_interval_seconds: frequencyToSeconds(newRssForm.value.frequency),
    }

    try {
      const response = await fetch(`${API_BASE_URL}/sources`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      const newSource = await response.json()

      // 转换为前端格式并添加到列表
      sources.value.push({
        id: newSource.id,
        name: newSource.author_name || '未命名',
        url: newSource.url,
        frequency: secondsToFrequency(newSource.fetch_interval_seconds),
        status: newSource.enabled ? 'active' : 'paused',
        statusClass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
        lastSyncTime: newSource.last_fetch_time || null,
        lastSyncStatus: 'success',
        syncLogs: [
          {
            timestamp: new Date().toISOString(),
            status: 'success',
            itemsCount: 0,
            message: '新增订阅源',
          },
        ],
        filterRules: '',
      })

      console.log('[Config Store] Added source:', newSource)
      return true
    } catch (error) {
      console.error('[Config Store] Failed to add source:', error)
      throw error
    }
  }

  /**
   * 删除 RSS 源
   */
  const deleteSource = async (id) => {
    try {
      const response = await fetch(`${API_BASE_URL}/sources/${id}`, {
        method: 'DELETE',
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      // 从列表中删除
      sources.value = sources.value.filter(s => s.id !== id)
      // 移除展开状态
      expandedSourceIds.value = expandedSourceIds.value.filter(sid => sid !== id)

      console.log('[Config Store] Deleted source:', id)
      return true
    } catch (error) {
      console.error('[Config Store] Failed to delete source:', error)
      throw error
    }
  }

  /**
   * 手动同步 RSS 源
   * 触发 Go 后端立即抓取，并更新 lastSyncStatus 为 'pending'
   */
  const syncSource = async (id) => {
    const source = sources.value.find(s => s.id === id)
    if (!source) {
      console.error('[Config Store] Source not found:', id)
      return
    }

    source.lastSyncStatus = 'pending'

    try {
      const response = await fetch(`${API_BASE_URL}/sources/${id}/fetch`, {
        method: 'POST',
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || `HTTP ${response.status}`)
      }

      // Go 后端已触发同步，现在监听日志更新
      // 这里可以选择：
      // 1. 轮询获取日志 (polling)
      // 2. 使用 SSE 流式获取日志 (streaming)

      // 暂时使用轮询方式，每 2 秒检查一次
      let pollCount = 0
      const maxPolls = 30 // 最多 60 秒

      const pollLogs = async () => {
        if (pollCount >= maxPolls) {
          source.lastSyncStatus = 'success'
          return
        }

        pollCount++

        try {
          const logsResponse = await fetch(`${API_BASE_URL}/sources/${id}/sync-logs?limit=5`)
          if (logsResponse.ok) {
            const logsData = await logsResponse.json()
            if (logsData.logs && logsData.logs.length > 0) {
              // 合并新日志
              const existingIds = new Set(source.syncLogs.map(log => log.timestamp))
              const newLogs = logsData.logs.filter(log => !existingIds.has(log.timestamp))
              source.syncLogs.unshift(...newLogs)

              // 检查是否完成
              const latestLog = logsData.logs[0]
              if (latestLog && latestLog.status === 'success') {
                source.lastSyncStatus = 'success'
                source.lastSyncTime = latestLog.timestamp
                return
              }
            }
          }

          // 继续轮询
          setTimeout(pollLogs, 2000)
        } catch (err) {
          console.warn('[Config Store] Failed to poll logs:', err)
          setTimeout(pollLogs, 2000)
        }
      }

      pollLogs()
    } catch (error) {
      source.lastSyncStatus = 'error'
      source.syncLogs.unshift({
        timestamp: new Date().toISOString(),
        status: 'error',
        itemsCount: 0,
        message: error.message,
      })
      console.error('[Config Store] Failed to sync source:', error)
      throw error
    }
  }

  /**
   * 切换展开行
   */
  const toggleSourceExpanded = (id) => {
    const index = expandedSourceIds.value.indexOf(id)
    if (index > -1) {
      expandedSourceIds.value.splice(index, 1)
    } else {
      expandedSourceIds.value.push(id)
    }
  }

  /**
   * 切换 RSS 添加模态框
   */
  const toggleRssModal = () => {
    showAddRssModal.value = !showAddRssModal.value
    if (!showAddRssModal.value) {
      resetNewRssForm()
    }
  }

  /**
   * 切换模型添加模态框
   */
  const toggleModelModal = () => {
    showAddModelModal.value = !showAddModelModal.value
    if (!showAddModelModal.value) {
      resetNewModelForm()
    }
  }

  /**
   * 重置 RSS 表单
   */
  const resetNewRssForm = () => {
    newRssForm.value = {
      name: '',
      url: '',
      frequency: 'hourly',
      filterRules: '',
    }
  }

  /**
   * 重置模型表单
   */
  const resetNewModelForm = () => {
    newModelForm.value = {
      name: '',
      provider: 'OpenAI',
      apiKey: '',
      baseUrl: '',
    }
  }

  // ============ 其他方法 ============

  // 更新API Key
  const updateApiKey = (key) => {
    apiKey.value = key
  }

  // 更新模型名称
  const updateModelName = (name) => {
    modelName.value = name
  }

  // 更新 Base URL
  const updateBaseUrl = (url) => {
    baseUrl.value = url
  }

  // 更新温度
  const updateTemperature = (temp) => {
    temperature.value = parseFloat(temp)
  }

  // 更新Top P
  const updateTopP = (p) => {
    topP.value = parseFloat(p)
  }

  // 更新Token数
  const updateMaxTokens = (tokens) => {
    maxTokens.value = parseInt(tokens)
  }

  // 获取当前配置参数（用于 API 调用）
  const getEvaluationConfig = () => {
    return {
      temperature: temperature.value,
      topP: topP.value,
      maxTokens: maxTokens.value,
    }
  }

  // 获取当前 LLM 配置（模型 + API Key + Base URL）
  const getLLMConfig = () => {
    return {
      modelName: modelName.value,
      apiKey: apiKey.value,
      baseUrl: baseUrl.value,
    }
  }

  return {
    // 状态
    apiKey,
    apiUrl,
    modelName,
    baseUrl,
    temperature,
    topP,
    maxTokens,
    isSaving,
    saveStatus,

    // RSS源状态
    sources,
    isLoadingSources,
    sourceLoadError,
    showAddRssModal,
    showAddModelModal,
    newRssForm,
    newModelForm,
    expandedSourceIds,

    // 计算属性
    isConfigValid,

    // 方法
    loadConfig,
    loadSources,
    saveConfig,
    addSource,
    deleteSource,
    syncSource,
    updateApiKey,
    updateModelName,
    updateBaseUrl,
    updateTemperature,
    updateTopP,
    updateMaxTokens,
    toggleRssModal,
    toggleModelModal,
    toggleSourceExpanded,
    resetNewRssForm,
    resetNewModelForm,
    getEvaluationConfig,
    getLLMConfig,
  }
})
