import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/**
 * Timeline Store - 完全使用真实 API 数据
 *
 * 不再依赖 Mock 数据，从后端 API 获取评估后的内容
 */
export const useTimelineStore = defineStore('timeline', () => {
  // ============ 状态 ============
  const cards = ref([])
  const loading = ref(false)
  const error = ref(null)
  const filter = ref('All')  // 'All', 'INTERESTING', 'BOOKMARK', 'SKIP'

  // 分页
  const pagination = ref({
    limit: 50,
    offset: 0,
    total: 0,
    hasMore: true
  })

  // ============ 计算属性 ============
  const filteredCards = computed(() => {
    if (filter.value === 'All') return cards.value

    return cards.value.filter(card => {
      const cardDecision = card.evaluation?.decision || 'SKIP'
      return cardDecision === filter.value
    })
  })

  const displayStats = computed(() => {
    return {
      total: cards.value.length,
      interesting: cards.value.filter(c => c.evaluation?.decision === 'INTERESTING').length,
      bookmark: cards.value.filter(c => c.evaluation?.decision === 'BOOKMARK').length,
      skip: cards.value.filter(c => c.evaluation?.decision === 'SKIP').length
    }
  })

  // ============ 方法 ============

  /**
   * 从 API 加载评估后的内容
   *
   * 流程：
   * 1. 获取 evaluated 状态的内容
   * 2. 对每个内容，获取对应的评估结果
   * 3. 转换为前端卡片格式
   */
  const loadCards = async (reset = false) => {
    if (reset) {
      pagination.value.offset = 0
      cards.value = []
    }

    loading.value = true
    error.value = null

    try {
      // API 端点：获取已评估的内容
      const url = new URL('http://127.0.0.1:8080/api/content')
      url.searchParams.append('status', 'EVALUATED')
      url.searchParams.append('limit', pagination.value.limit)
      url.searchParams.append('offset', pagination.value.offset)

      const response = await fetch(url.toString())
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      const contentList = data?.data || []

      // 转换每条内容为卡片格式
      // 注意：评估数据已经在 content 对象中（通过 join 获取）
      const newCards = contentList.map(content => ({
        id: content.id,
        title: content.title,
        description: (content.content || '').substring(0, 200) + '...',
        content: content.content,
        source: content.source?.name || 'Unknown Source',
        url: content.url,
        timestamp: formatDate(new Date(content.published_at)),

        // 评估数据
        evaluation: content.evaluation ? {
          innovation_score: content.evaluation.innovation_score || 0,
          depth_score: content.evaluation.depth_score || 0,
          decision: content.evaluation.decision || 'SKIP',
          tldr: content.evaluation.tldr || '',
          reasoning: content.evaluation.reasoning || ''
        } : {
          innovation_score: 0,
          depth_score: 0,
          decision: 'SKIP',
          tldr: 'Not evaluated yet',
          reasoning: ''
        },

        // 视觉属性
        statusColor: getStatusColor(content.evaluation?.decision)
      }))

      // 追加新卡片（分页）
      cards.value = reset ? newCards : [...cards.value, ...newCards]

      // 更新分页信息
      pagination.value.offset += contentList.length
      pagination.value.hasMore = contentList.length === pagination.value.limit

      logger.info(`Loaded ${newCards.length} cards, total: ${cards.value.length}`)

    } catch (err) {
      error.value = `Failed to load cards: ${err.message}`
      logger.error('Timeline load error:', err)
      cards.value = []
    } finally {
      loading.value = false
    }
  }

  /**
   * 刷新卡片（从头加载）
   */
  const refresh = async () => {
    await loadCards(true)
  }

  /**
   * 加载更多（分页）
   */
  const loadMore = async () => {
    if (!pagination.value.hasMore) return
    await loadCards(false)
  }

  /**
   * 设置过滤器
   */
  const setFilter = (newFilter) => {
    filter.value = newFilter
  }

  // ============ 辅助函数 ============

  /**
   * 格式化日期
   */
  function formatDate(date) {
    const now = new Date()
    const diff = now - date

    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) return 'just now'
    if (minutes < 60) return `${minutes}m ago`
    if (hours < 24) return `${hours}h ago`
    if (days < 7) return `${days}d ago`

    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
    })
  }

  /**
   * 获取状态颜色
   */
  function getStatusColor(decision) {
    switch (decision) {
      case 'INTERESTING':
        return '#22c55e'  // green-500
      case 'BOOKMARK':
        return '#f59e0b'  // amber-500
      case 'SKIP':
        return '#ef4444'  // red-500
      default:
        return '#6b7280'  // gray-500
    }
  }

  /**
   * 简单的日志函数
   */
  const logger = {
    info: (msg, data = '') => console.log(`[Timeline] ${msg}`, data),
    error: (msg, err = '') => console.error(`[Timeline Error] ${msg}`, err)
  }

  // 初始化时加载数据
  loadCards()

  // ============ 暴露接口 ============
  return {
    // 状态
    cards: filteredCards,
    allCards: cards,
    loading,
    error,
    filter,
    pagination,

    // 计算属性
    displayStats,

    // 方法
    loadCards,
    refresh,
    loadMore,
    setFilter,

    // 调试用
    rawCards: cards
  }
})
