import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useReaderStore = defineStore('reader', () => {
  const articles = ref([])
  const selectedArticleId = ref(null)
  const selectedSource = ref('all')
  const selectedAuthor = ref(null)
  const expandedSources = ref({})
  const isLoading = ref(false)

  const API_BASE_URL = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`

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

    const sourceName = item.source_name || 'Unknown Source'
    const authorName = item.author_name || sourceName

    return {
      id: item.id,
      title: item.title,
      source: authorName,
      sourceName: sourceName,
      authorName: authorName,
      sourceInitials: sourceName.slice(0, 2).toUpperCase(),
      faviconUrl: item.favicon_url || null,
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
      imageUrls: item.image_urls || [],
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
   * Sources derived from loaded articles (two-level: source -> authors)
   */
  const sources = computed(() => {
    const map = {}
    articles.value.forEach(a => {
      const sn = a.sourceName
      if (!map[sn]) {
        map[sn] = { name: sn, initials: sn.slice(0, 2).toUpperCase(), faviconUrl: a.faviconUrl, count: 0, authors: {} }
      }
      if (!map[sn].faviconUrl && a.faviconUrl) {
        map[sn].faviconUrl = a.faviconUrl
      }
      map[sn].count++
      const an = a.authorName
      if (!map[sn].authors[an]) {
        map[sn].authors[an] = { name: an, count: 0 }
      }
      map[sn].authors[an].count++
    })
    return Object.values(map)
      .map(s => ({
        ...s,
        authors: Object.values(s.authors).sort((a, b) => b.count - a.count),
      }))
      .sort((a, b) => b.count - a.count)
  })

  const totalCount = computed(() => articles.value.length)

  /**
   * Filtered articles by selected source and author
   */
  const filteredArticles = computed(() => {
    let list = articles.value
    if (selectedSource.value !== 'all') {
      list = list.filter(a => a.sourceName === selectedSource.value)
      if (selectedAuthor.value) {
        list = list.filter(a => a.authorName === selectedAuthor.value)
      }
    }
    return list
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
    if (selectedSource.value === source && source !== 'all') {
      // Toggle expand/collapse
      expandedSources.value[source] = !expandedSources.value[source]
    } else {
      selectedSource.value = source
      selectedAuthor.value = null
      selectedArticleId.value = null
      if (source !== 'all') {
        expandedSources.value[source] = true
      }
    }
  }

  const selectAuthor = (sourceName, authorName) => {
    selectedSource.value = sourceName
    selectedAuthor.value = authorName
    selectedArticleId.value = null
    expandedSources.value[sourceName] = true
  }

  const toggleSourceExpand = (sourceName) => {
    expandedSources.value[sourceName] = !expandedSources.value[sourceName]
  }

  return {
    articles,
    selectedArticleId,
    selectedSource,
    selectedAuthor,
    expandedSources,
    isLoading,
    sources,
    totalCount,
    filteredArticles,
    selectedArticle,
    loadArticles,
    selectArticle,
    selectSource,
    selectAuthor,
    toggleSourceExpand,
  }
})
