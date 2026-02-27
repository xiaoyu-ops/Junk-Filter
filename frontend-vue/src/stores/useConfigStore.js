import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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
  const sources = ref([
    {
      id: 1,
      name: 'TechCrunch',
      url: 'https://techcrunch.com/feed/',
      frequency: 'hourly',
      status: 'active',
      statusClass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      lastSyncTime: '2026-02-26T10:30:00Z',
      lastSyncStatus: 'success',
      syncLogs: [
        {
          timestamp: '2026-02-26T10:30:00Z',
          status: 'success',
          itemsCount: 12,
          message: '成功获取 12 条新文章',
        },
        {
          timestamp: '2026-02-26T09:30:00Z',
          status: 'success',
          itemsCount: 8,
          message: '成功获取 8 条新文章',
        },
      ],
      filterRules: '优先级 >= 7',
    },
    {
      id: 2,
      name: 'Hacker News',
      url: 'https://news.ycombinator.com/rss',
      frequency: '30min',
      status: 'active',
      statusClass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      lastSyncTime: '2026-02-26T10:15:00Z',
      lastSyncStatus: 'success',
      syncLogs: [
        {
          timestamp: '2026-02-26T10:15:00Z',
          status: 'success',
          itemsCount: 5,
          message: '成功获取 5 条新文章',
        },
      ],
      filterRules: '',
    },
    {
      id: 3,
      name: 'The Verge',
      url: 'https://www.theverge.com/rss/index.xml',
      frequency: '2hours',
      status: 'paused',
      statusClass: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
      lastSyncTime: '2026-02-26T08:00:00Z',
      lastSyncStatus: 'pending',
      syncLogs: [],
      filterRules: 'category: tech',
    },
  ])

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

  // 初始化配置（从localStorage恢复）
  const loadConfig = () => {
    const saved = localStorage.getItem('config')
    if (saved) {
      try {
        const config = JSON.parse(saved)
        // 新字段：modelName 和 baseUrl
        modelName.value = config.modelName || modelName.value
        baseUrl.value = config.baseUrl || baseUrl.value
        // 向后兼容：如果存在旧的 apiModel 字段
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
  }

  // 保存配置
  const saveConfig = async () => {
    isSaving.value = true
    saveStatus.value = null

    try {
      // 模拟API请求（20%失败率）
      await new Promise(resolve => setTimeout(resolve, 1000))

      const isSuccess = Math.random() > 0.2

      if (isSuccess) {
        // 保存到localStorage（不包括apiKey，它来自环境变量）
        localStorage.setItem('config', JSON.stringify({
          modelName: modelName.value,
          baseUrl: baseUrl.value,
          temperature: temperature.value,
          topP: topP.value,
          maxTokens: maxTokens.value,
        }))
        saveStatus.value = 'success'
        return true
      } else {
        saveStatus.value = 'error'
        throw new Error('保存配置失败')
      }
    } catch (error) {
      saveStatus.value = 'error'
      console.error('Config save error:', error)
      return false
    } finally {
      isSaving.value = false
    }
  }

  // RSS源管理方法
  const toggleRssModal = () => {
    showAddRssModal.value = !showAddRssModal.value
    if (!showAddRssModal.value) {
      resetNewRssForm()
    }
  }

  const toggleModelModal = () => {
    showAddModelModal.value = !showAddModelModal.value
    if (!showAddModelModal.value) {
      resetNewModelForm()
    }
  }

  const addSource = () => {
    const newSource = {
      id: Math.max(...sources.value.map(s => s.id), 0) + 1,
      name: newRssForm.value.name,
      url: newRssForm.value.url,
      frequency: newRssForm.value.frequency,
      status: 'active',
      statusClass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      lastSyncTime: new Date().toISOString(),
      lastSyncStatus: 'success',
      syncLogs: [
        {
          timestamp: new Date().toISOString(),
          status: 'success',
          itemsCount: 0,
          message: '新增订阅源',
        },
      ],
      filterRules: newRssForm.value.filterRules,
    }
    sources.value.push(newSource)
  }

  const deleteSource = (id) => {
    sources.value = sources.value.filter(s => s.id !== id)
    // 移除展开状态
    expandedSourceIds.value = expandedSourceIds.value.filter(sid => sid !== id)
  }

  const syncSource = async (id) => {
    const source = sources.value.find(s => s.id === id)
    if (!source) return

    source.lastSyncStatus = 'pending'
    // 模拟同步（2秒）
    await new Promise(resolve => setTimeout(resolve, 2000))

    const success = Math.random() > 0.15
    source.lastSyncStatus = success ? 'success' : 'error'
    source.lastSyncTime = new Date().toISOString()

    if (success) {
      const itemCount = Math.floor(Math.random() * 20) + 1
      source.syncLogs.unshift({
        timestamp: new Date().toISOString(),
        status: 'success',
        itemsCount: itemCount,
        message: `成功获取 ${itemCount} 条新文章`,
      })
    } else {
      source.syncLogs.unshift({
        timestamp: new Date().toISOString(),
        status: 'error',
        itemsCount: 0,
        message: '连接超时，请检查URL',
      })
    }

    // 只保留最近10条日志
    source.syncLogs = source.syncLogs.slice(0, 10)
  }

  const toggleSourceExpanded = (id) => {
    const index = expandedSourceIds.value.indexOf(id)
    if (index > -1) {
      expandedSourceIds.value.splice(index, 1)
    } else {
      expandedSourceIds.value.push(id)
    }
  }

  const resetNewRssForm = () => {
    newRssForm.value = {
      name: '',
      url: '',
      frequency: 'hourly',
      filterRules: '',
    }
  }

  const resetNewModelForm = () => {
    newModelForm.value = {
      name: '',
      provider: 'OpenAI',
      apiKey: '',
      baseUrl: '',
    }
  }

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
    showAddRssModal,
    showAddModelModal,
    newRssForm,
    newModelForm,
    expandedSourceIds,

    // 计算属性
    isConfigValid,

    // 方法
    loadConfig,
    saveConfig,
    updateApiKey,
    updateModelName,
    updateBaseUrl,
    updateTemperature,
    updateTopP,
    updateMaxTokens,
    toggleRssModal,
    toggleModelModal,
    addSource,
    deleteSource,
    syncSource,
    toggleSourceExpanded,
    resetNewRssForm,
    resetNewModelForm,
    getEvaluationConfig,
    getLLMConfig,
  }
})
