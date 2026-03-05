<template>
  <main class="w-full min-h-screen flex flex-col bg-surface-light dark:bg-[#0f0f11] overflow-y-auto">
    <!-- Initial: centered search -->
    <div
      v-if="!isSearching"
      class="w-full flex-grow flex flex-col items-center justify-center px-4 transition-all duration-500"
      :style="{ paddingTop: '1vh' }"
    >
      <div class="w-full max-w-3xl flex flex-col items-center justify-center space-y-12">
        <div class="flex flex-col items-center space-y-8 text-center">
          <div class="w-24 h-24 bg-gray-50 dark:bg-gray-800 rounded-3xl flex items-center justify-center shadow-sm ring-1 ring-gray-100 dark:ring-gray-700">
            <span class="material-icons-outlined text-6xl text-gray-800 dark:text-gray-200">delete_outline</span>
          </div>
          <h2 class="text-4xl font-bold text-gray-900 dark:text-white tracking-tight">What are you looking for?</h2>
          <p class="text-gray-500 dark:text-gray-400 text-sm">Search through AI-evaluated articles by title, content, or author</p>
        </div>

        <div class="w-full max-w-2xl relative">
          <SearchBar
            v-model="keyword"
            :is-searching="isSearching"
            @search="handleSearch"
          />
        </div>
      </div>
    </div>

    <!-- Active: sticky search + results -->
    <div v-if="isSearching" class="w-full flex flex-col min-h-screen overflow-y-auto">
      <!-- Sticky search bar -->
      <div class="sticky top-16 z-40 w-full bg-surface-light dark:bg-[#0f0f11] py-4 px-4 border-b border-gray-100 dark:border-gray-800 transition-all duration-300">
        <div class="max-w-2xl mx-auto relative">
          <SearchBar
            v-model="keyword"
            :is-searching="isSearching"
            compact
            @search="handleSearch"
          />
        </div>
      </div>

      <!-- Result count -->
      <div class="px-4 pt-6 pb-4 max-w-2xl mx-auto w-full">
        <div class="flex items-center justify-between">
          <p class="text-sm text-gray-500 dark:text-gray-400 font-medium">
            <span v-if="isLoading">Searching...</span>
            <span v-else>Found {{ searchResults.length }} articles for "{{ keyword }}"</span>
          </p>
        </div>
      </div>

      <!-- Results -->
      <div class="flex-grow px-4 pb-20">
        <div class="max-w-2xl mx-auto w-full flex flex-col space-y-3">
          <div v-if="searchResults.length > 0" class="space-y-3">
            <article
              v-for="result in searchResults"
              :key="result.id"
              @click="goToReader(result.id)"
              class="p-4 bg-white dark:bg-[#18181b] border border-gray-100 dark:border-gray-800 rounded-xl hover:shadow-lg hover:border-gray-200 dark:hover:border-gray-700 transition-all duration-200 cursor-pointer"
            >
              <div class="flex justify-between items-start mb-2">
                <span class="text-[10px] font-bold text-gray-400 uppercase tracking-tight">
                  {{ result.author_name || result.source_name }} &bull; {{ formatTime(result.published_at) }}
                </span>
                <span
                  :class="[
                    'px-1.5 py-0.5 rounded text-[10px] font-bold border shrink-0 ml-2',
                    getScoreStyle(result)
                  ]"
                >{{ getAvgScore(result) }} {{ getScoreLabel(result) }}</span>
              </div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white leading-snug mb-1.5">{{ result.title }}</h3>
              <p v-if="result.tldr" class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2 leading-relaxed">{{ result.tldr }}</p>
              <p v-else class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2 leading-relaxed">{{ result.content }}</p>

              <!-- Scores -->
              <div class="flex items-center gap-4 mt-2">
                <span class="text-xs text-gray-400">
                  Innovation <span class="font-semibold text-blue-600 dark:text-blue-400">{{ result.innovation_score }}/10</span>
                </span>
                <span class="text-xs text-gray-400">
                  Depth <span class="font-semibold text-green-600 dark:text-green-400">{{ result.depth_score }}/10</span>
                </span>
                <span :class="[
                  'text-xs font-medium px-1.5 py-0.5 rounded',
                  result.decision === 'INTERESTING' ? 'text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20' :
                  result.decision === 'SKIP' ? 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20' :
                  'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/20'
                ]">{{ result.decision }}</span>
              </div>
            </article>
          </div>

          <!-- No results -->
          <div v-if="!isLoading && searchResults.length === 0" class="py-12 text-center">
            <span class="material-symbols-outlined text-4xl text-gray-300 dark:text-gray-600 block mb-3">search_off</span>
            <p class="text-gray-500 dark:text-gray-400">No articles found for "{{ keyword }}"</p>
          </div>
        </div>
      </div>

      <!-- Back button -->
      <div class="fixed bottom-8 right-8 z-30">
        <button
          @click="closeSearch"
          class="p-4 bg-gray-900 hover:bg-gray-800 dark:bg-white dark:hover:bg-gray-200 text-white dark:text-black rounded-full shadow-lg transition-all hover:shadow-xl active:scale-95"
          title="Back to search"
        >
          <span class="material-icons-outlined">arrow_upward</span>
        </button>
      </div>
    </div>
  </main>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import SearchBar from './SearchBar.vue'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const isSearching = ref(false)
const keyword = ref('')
const searchResults = ref([])
const isLoading = ref(false)

const handleSearch = async (query) => {
  if (!query || typeof query !== 'string' || query.trim().length === 0) return

  isSearching.value = true
  isLoading.value = true

  const { show: showToast } = useToast()

  try {
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const apiUrl = baseUrl.endsWith('/api') ? baseUrl : `${baseUrl}/api`
    const params = new URLSearchParams({ q: query, status: 'EVALUATED', limit: '50' })

    const response = await fetch(`${apiUrl}/search?${params.toString()}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)

    const data = await response.json()
    searchResults.value = data.data || []

    if (searchResults.value.length === 0) {
      showToast(`No articles found for "${query}"`, 'info', 2000)
    }
  } catch (error) {
    console.error('Search failed:', error)
    searchResults.value = []
    showToast(`Search failed: ${error.message}`, 'error', 3000)
  } finally {
    isLoading.value = false
  }
}

const closeSearch = () => {
  isSearching.value = false
  keyword.value = ''
}

const goToReader = (articleId) => {
  router.push(`/reader/${articleId}`)
}

const formatTime = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now - date
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)
  if (diffHours < 1) return 'just now'
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

const getAvgScore = (r) => {
  return ((r.innovation_score + r.depth_score) / 2).toFixed(1)
}

const getScoreLabel = (r) => {
  const avg = (r.innovation_score + r.depth_score) / 2
  if (avg >= 8) return 'CLEAN'
  if (avg >= 6) return 'RELEVANT'
  return 'NOISY'
}

const getScoreStyle = (r) => {
  const avg = (r.innovation_score + r.depth_score) / 2
  if (avg >= 8) return 'bg-green-100 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800'
  if (avg >= 6) return 'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
  return 'bg-amber-100 text-amber-700 border-amber-200 dark:bg-amber-900/30 dark:text-amber-400 dark:border-amber-800'
}
</script>
