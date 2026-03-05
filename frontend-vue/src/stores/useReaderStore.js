import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useReaderStore = defineStore('reader', () => {
  const articles = ref([])
  const selectedArticleId = ref(null)
  const selectedSource = ref('all')
  const isLoading = ref(false)

  const API_BASE_URL = 'http://localhost:8080/api'

  /**
   * AI score label mapping based on average score
   */
  const getScoreLabel = (avgScore) => {
    if (avgScore >= 8) return { text: 'CLEAN', color: 'green' }
    if (avgScore >= 6) return { text: 'RELEVANT', color: 'gray' }
    return { text: 'NOISY', color: 'amber' }
  }

  const formatTimeAgo = (dateString) => {
    if (!dateString) return ''
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

  const transformArticle = (item) => {
    const evaluation = item.evaluation || {}
    const innovationScore = evaluation.innovation_score || 0
    const depthScore = evaluation.depth_score || 0
    const avgScore = ((innovationScore + depthScore) / 2).toFixed(1)
    const label = getScoreLabel(parseFloat(avgScore))

    return {
      id: item.id,
      title: item.title,
      source: item.author_name || item.source_name || 'Unknown',
      sourceInitials: (item.author_name || item.source_name || 'U').slice(0, 2).toUpperCase(),
      timeAgo: formatTimeAgo(item.published_at),
      publishedAt: item.published_at,
      summary: evaluation.tldr || '',
      content: item.clean_content || item.content || '',
      url: item.original_url || item.url || '',
      aiScore: parseFloat(avgScore),
      aiLabel: label.text,
      aiLabelColor: label.color,
      innovationScore,
      depthScore,
      decision: evaluation.decision || 'BOOKMARK',
      keyConcepts: evaluation.key_concepts || [],
      reasoning: evaluation.reasoning || '',
    }
  }

  /**
   * Load all evaluated articles from API
   */
  const loadArticles = async () => {
    isLoading.value = true
    try {
      const response = await fetch(`${API_BASE_URL}/content?status=EVALUATED&limit=500`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const data = await response.json()
      articles.value = (data.data || []).map(transformArticle)
    } catch (err) {
      console.error('[Reader] Failed to load articles:', err)
      articles.value = []
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Sources derived from loaded articles
   */
  const sources = computed(() => {
    const map = {}
    articles.value.forEach(a => {
      if (!map[a.source]) {
        map[a.source] = { name: a.source, initials: a.sourceInitials, count: 0 }
      }
      map[a.source].count++
    })
    return Object.values(map).sort((a, b) => b.count - a.count)
  })

  const totalCount = computed(() => articles.value.length)

  /**
   * Filtered articles by selected source
   */
  const filteredArticles = computed(() => {
    if (selectedSource.value === 'all') return articles.value
    return articles.value.filter(a => a.source === selectedSource.value)
  })

  /**
   * Currently selected article object
   */
  const selectedArticle = computed(() => {
    if (!selectedArticleId.value) return filteredArticles.value[0] || null
    return articles.value.find(a => a.id === selectedArticleId.value) || null
  })

  const selectArticle = (id) => {
    selectedArticleId.value = id
  }

  const selectSource = (source) => {
    selectedSource.value = source
    selectedArticleId.value = null
  }

  return {
    articles,
    selectedArticleId,
    selectedSource,
    isLoading,
    sources,
    totalCount,
    filteredArticles,
    selectedArticle,
    loadArticles,
    selectArticle,
    selectSource,
  }
})
