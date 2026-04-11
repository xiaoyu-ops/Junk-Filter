import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useTimelineStore = defineStore('timeline', () => {
  const activeFilter = ref('All')
  const isDetailDrawerOpen = ref(false)
  const selectedCard = ref(null)
  const cards = ref([])
  const isLoading = ref(false)
  const error = ref(null)

  // ✨ RSS 抓取进度统计（新增）
  const stats = ref({
    pending: 0,
    processing: 0,
    evaluated: 0,
    discarded: 0,
    total: 0
  })
  const isLoadingStats = ref(false)

  const API_BASE_URL = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`

  /**
   * 决策值到显示文本和颜色的映射
   */
  const decisionMap = {
    'INTERESTING': { text: 'Interesting', color: 'green' },
    'BOOKMARK': { text: 'Bookmark', color: 'amber' },
    'SKIP': { text: 'Skip', color: 'red' }
  }

  /**
   * 将数据库中的内容记录转换为 UI 卡片格式
   */
  const transformContentToCard = (contentItem) => {
    // 处理不同的响应格式：可能是嵌套的 {ContentResponse, evaluation} 或直接的 ContentResponse
    const content = contentItem.title ? contentItem : contentItem
    const evaluation = contentItem.evaluation || {}

    const decision = evaluation.decision || 'BOOKMARK'
    const decisionInfo = decisionMap[decision] || { text: 'Unknown', color: 'gray' }

    // Use author_name from RSS item, fallback to source_name from sources table
    const sourceName = content.author_name || contentItem.source_name || 'Unknown Source'

    return {
      id: content.id,
      author: sourceName,
      authorTime: formatTimeAgo(content.published_at || new Date().toISOString()),
      title: content.title,
      content: content.clean_content || content.content || '',
      url: content.original_url || content.url || '',
      status: decisionInfo.text,
      statusColor: decisionInfo.color,
      faviconUrl: contentItem.favicon_url || null,
      // 评估数据
      innovationScore: evaluation.innovation_score || 0,
      depthScore: evaluation.depth_score || 0,
      tldr: evaluation.tldr || content.title,
      keyConepts: evaluation.key_concepts || [],
      decision: decision,
      reasoning: evaluation.reasoning || '',
      publishedAt: content.published_at,
      sourceId: content.source_id,
    }
  }

  /**
   * 格式化时间为相对时间（如 "2h ago"）
   */
  const formatTimeAgo = (dateString) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`

    return date.toLocaleDateString()
  }

  /**
   * 从 API 加载已评估的内容
   * GET /api/content?status=EVALUATED
   */
  const loadContent = async () => {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch(
        `${API_BASE_URL}/content?status=EVALUATED&limit=500`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        }
      )

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      const contentList = data.data || []

      // 转换为 UI 卡片格式
      cards.value = contentList.map(transformContentToCard)

      if (cards.value.length === 0) {
        console.info('[Timeline] No evaluated content available, displaying mock data')
      } else {
        console.info(`[Timeline] Loaded ${cards.value.length} evaluated content items`)
      }
    } catch (err) {
      error.value = err.message
      console.error('[Timeline] Failed to load content:', err)
      // 降级：使用空数组而不是 mock 数据
      cards.value = []
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 加载 RSS 抓取进度统计（新增）
   * GET /api/content/stats
   */
  const loadStats = async () => {
    isLoadingStats.value = true

    try {
      const response = await fetch(
        `${API_BASE_URL}/content/stats`,
        {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        }
      )

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      stats.value = {
        pending: data.pending || 0,
        processing: data.processing || 0,
        evaluated: data.evaluated || 0,
        discarded: data.discarded || 0,
        total: data.total || 0
      }

      console.info(`[Timeline] Stats loaded: ${stats.value.pending} pending, ${stats.value.processing} processing, ${stats.value.evaluated} evaluated`)
    } catch (err) {
      console.error('[Timeline] Failed to load stats:', err)
    } finally {
      isLoadingStats.value = false
    }
  }

  /**
   * 刷新内容列表
   */
  const refreshContent = async () => {
    await Promise.all([loadContent(), loadStats()])
  }

  /**
   * 终止评估：将所有 PENDING/PROCESSING 内容标记为 DISCARDED
   */
  const isStopping = ref(false)
  const isStopped = ref(false)
  const stopEvaluation = async () => {
    isStopping.value = true
    try {
      const response = await fetch(`${API_BASE_URL}/content/stop-evaluation`, { method: 'POST' })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const data = await response.json()
      console.info(`[Timeline] Evaluation stopped, ${data.affected} items discarded`)
      isStopped.value = true
      await loadStats()
      return data.affected
    } catch (err) {
      console.error('[Timeline] Failed to stop evaluation:', err)
      throw err
    } finally {
      isStopping.value = false
    }
  }

  /**
   * 重启评估：将 DISCARDED 内容恢复为 PENDING
   */
  const restartEvaluation = async () => {
    isStopping.value = true
    try {
      const response = await fetch(`${API_BASE_URL}/content/restart-evaluation`, { method: 'POST' })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const data = await response.json()
      console.info(`[Timeline] Evaluation restarted, ${data.affected} items reset to PENDING`)
      isStopped.value = false
      await loadStats()
      return data.affected
    } catch (err) {
      console.error('[Timeline] Failed to restart evaluation:', err)
      throw err
    } finally {
      isStopping.value = false
    }
  }

  /**
   * 设置过滤器
   * TODO: 在 API 支持后添加过滤参数
   */
  const setFilter = (filter) => {
    activeFilter.value = filter
    // 当前版本加载所有 EVALUATED 内容
    // 后续可扩展为：
    // - filter = 'Interesting' → decision=INTERESTING
    // - filter = 'Bookmark' → decision=BOOKMARK
    // - filter = 'Skip' → decision=SKIP
  }

  /**
   * 打开详情抽屉
   */
  const openDetailDrawer = (card) => {
    selectedCard.value = card
    isDetailDrawerOpen.value = true
  }

  /**
   * 关闭详情抽屉
   */
  const closeDetailDrawer = () => {
    isDetailDrawerOpen.value = false
    selectedCard.value = null
  }

  /**
   * 计算过滤后的卡片列表
   */
  const filteredCards = computed(() => {
    const list = activeFilter.value === 'All'
      ? cards.value
      : cards.value.filter(card => card.status === activeFilter.value)
    return [...list].sort((a, b) => {
      const timeA = a.publishedAt ? new Date(a.publishedAt).getTime() : 0
      const timeB = b.publishedAt ? new Date(b.publishedAt).getTime() : 0
      return timeB - timeA
    })
  })

  /**
   * 初始化：页面加载时自动获取内容和进度
   */
  const initialize = async () => {
    // 并行加载内容和统计
    await Promise.all([
      loadContent(),
      loadStats()
    ])

    // 根据统计数据判断是否处于终止状态
    if (stats.value.discarded > 0 && stats.value.pending === 0 && stats.value.processing === 0) {
      isStopped.value = true
    }

    // ✨ 定时刷新统计信息（每 3 秒）
    setInterval(loadStats, 3000)
  }

  return {
    activeFilter,
    isDetailDrawerOpen,
    selectedCard,
    cards,
    filteredCards,
    isLoading,
    error,
    stats,           // ✨ 导出统计信息
    isLoadingStats,  // ✨ 导出加载状态
    setFilter,
    openDetailDrawer,
    closeDetailDrawer,
    loadContent,
    loadStats,       // ✨ 导出加载统计方法
    refreshContent,
    stopEvaluation,
    restartEvaluation,
    isStopping,
    isStopped,
    initialize,
  }
})
