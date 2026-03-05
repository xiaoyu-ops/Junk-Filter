<template>
  <div class="w-full h-[calc(100vh-64px)] flex overflow-hidden p-4 pt-20 gap-4">
    <!-- Left Sidebar: Sources -->
    <aside class="w-60 flex flex-col shrink-0 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div class="p-3 flex flex-col gap-1 overflow-y-auto custom-scrollbar h-full">
        <!-- All Feed -->
        <div class="mb-3">
          <button
            @click="readerStore.selectSource('all')"
            :class="[
              'w-full flex items-center justify-between px-3 py-2 rounded-xl font-medium transition-colors',
              readerStore.selectedSource === 'all'
                ? 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'
            ]"
          >
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-lg">inbox</span>
              <span class="text-sm">All Feed</span>
            </div>
            <span class="text-xs font-semibold bg-gray-200 dark:bg-gray-700 px-2 py-0.5 rounded-full">{{ readerStore.totalCount }}</span>
          </button>
        </div>

        <!-- Sources (two-level: source -> authors) -->
        <div>
          <h3 class="px-3 text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-2">Sources</h3>
          <div class="space-y-0.5">
            <div v-for="source in readerStore.sources" :key="source.name">
              <!-- Source row -->
              <button
                @click="readerStore.selectSource(source.name)"
                :class="[
                  'w-full flex items-center justify-between px-3 py-2 rounded-xl transition-colors',
                  readerStore.selectedSource === source.name && !readerStore.selectedAuthor
                    ? 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white'
                    : readerStore.selectedSource === source.name
                      ? 'bg-gray-50 dark:bg-gray-800/50 text-gray-700 dark:text-gray-300'
                      : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'
                ]"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <span
                    class="material-symbols-outlined text-[14px] text-gray-400 transition-transform duration-200"
                    :style="{ transform: readerStore.expandedSources[source.name] ? 'rotate(90deg)' : 'rotate(0deg)' }"
                  >chevron_right</span>
                  <div class="w-5 h-5 bg-gray-100 dark:bg-gray-700 rounded flex items-center justify-center text-[10px] font-bold shrink-0">{{ source.initials }}</div>
                  <span class="text-sm truncate">{{ source.name }}</span>
                </div>
                <span class="text-xs text-gray-400 shrink-0 ml-2">{{ source.count }}</span>
              </button>

              <!-- Authors (expanded) -->
              <Transition
                enter-active-class="transition-all duration-200 ease-out"
                enter-from-class="opacity-0 max-h-0"
                enter-to-class="opacity-100 max-h-96"
                leave-active-class="transition-all duration-150 ease-in"
                leave-from-class="opacity-100 max-h-96"
                leave-to-class="opacity-0 max-h-0"
              >
                <div v-if="readerStore.expandedSources[source.name]" class="overflow-hidden">
                  <button
                    v-for="author in source.authors"
                    :key="author.name"
                    @click="readerStore.selectAuthor(source.name, author.name)"
                    :class="[
                      'w-full flex items-center justify-between pl-12 pr-3 py-1.5 rounded-lg transition-colors text-xs',
                      readerStore.selectedAuthor === author.name && readerStore.selectedSource === source.name
                        ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 font-medium'
                        : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                    ]"
                  >
                    <span class="truncate">{{ author.name }}</span>
                    <span class="text-gray-400 shrink-0 ml-2">{{ author.count }}</span>
                  </button>
                </div>
              </Transition>
            </div>
          </div>
        </div>
      </div>
    </aside>

    <!-- Middle: Article List -->
    <section class="w-[340px] flex flex-col shrink-0 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div class="p-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between shrink-0">
        <h2 class="font-semibold text-sm text-gray-900 dark:text-white">
          {{ readerStore.selectedSource === 'all' ? 'All Articles' : readerStore.selectedAuthor || readerStore.selectedSource }}
        </h2>
        <span class="text-xs text-gray-400">{{ readerStore.filteredArticles.length }}</span>
      </div>
      <div class="flex-1 overflow-y-auto custom-scrollbar">
        <!-- Loading -->
        <div v-if="readerStore.isLoading" class="p-6 text-center text-sm text-gray-400">Loading...</div>

        <!-- Empty -->
        <div v-else-if="readerStore.filteredArticles.length === 0" class="p-6 text-center text-sm text-gray-400">No articles</div>

        <!-- Article Items -->
        <article
          v-for="article in readerStore.filteredArticles"
          :key="article.id"
          @click="readerStore.selectArticle(article.id)"
          :class="[
            'p-3 border-b border-gray-50 dark:border-gray-800 cursor-pointer transition-colors',
            readerStore.selectedArticle?.id === article.id
              ? 'bg-blue-50/50 dark:bg-blue-900/10'
              : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'
          ]"
        >
          <div class="flex justify-between items-start mb-1.5">
            <span class="text-[10px] font-bold text-gray-400 uppercase tracking-tight">{{ article.source }} &bull; {{ article.timeAgo }}</span>
            <span
              :class="[
                'px-1.5 py-0.5 rounded text-[10px] font-bold border shrink-0 ml-2',
                article.aiLabelColor === 'green' ? 'bg-green-100 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800' :
                article.aiLabelColor === 'amber' ? 'bg-amber-100 text-amber-700 border-amber-200 dark:bg-amber-900/30 dark:text-amber-400 dark:border-amber-800' :
                'bg-gray-100 text-gray-600 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
              ]"
            >{{ article.aiScore }} {{ article.aiLabel }}</span>
          </div>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white leading-snug mb-1 line-clamp-2">{{ article.title }}</h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 leading-relaxed">{{ article.summary || article.content }}</p>
        </article>
      </div>
    </section>

    <!-- Right: Reading Pane -->
    <section class="flex-1 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 overflow-y-auto custom-scrollbar">
      <!-- No article selected -->
      <div v-if="!readerStore.selectedArticle" class="h-full flex items-center justify-center">
        <div class="text-center text-gray-400">
          <span class="material-symbols-outlined text-5xl mb-3 block">article</span>
          <p class="text-sm">Select an article to read</p>
        </div>
      </div>

      <!-- Article Content -->
      <div v-else class="max-w-3xl mx-auto px-8 py-10">
        <!-- AI Summary Card -->
        <div v-if="readerStore.selectedArticle.summary" class="mb-8 p-5 bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm relative">
          <div class="absolute -top-3 left-5 px-2.5 py-0.5 bg-gray-900 dark:bg-white text-white dark:text-gray-900 text-[10px] font-bold rounded-full flex items-center gap-1.5">
            <span class="material-symbols-outlined text-[12px]">auto_awesome</span>
            AI SUMMARY
          </div>
          <p class="text-gray-700 dark:text-gray-300 text-sm leading-relaxed mt-2">{{ readerStore.selectedArticle.summary }}</p>

          <!-- Key Concepts -->
          <div v-if="readerStore.selectedArticle.keyConcepts.length > 0" class="mt-3 flex flex-wrap gap-1.5">
            <span
              v-for="(concept, idx) in readerStore.selectedArticle.keyConcepts"
              :key="idx"
              class="px-2 py-0.5 bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 text-[10px] font-medium rounded-full"
            >{{ concept }}</span>
          </div>
        </div>

        <!-- Article Header -->
        <div class="space-y-3 mb-8">
          <div class="flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
            <span class="font-medium text-gray-900 dark:text-white">{{ readerStore.selectedArticle.source }}</span>
            <span>&bull;</span>
            <span>{{ readerStore.selectedArticle.timeAgo }}</span>
            <span>&bull;</span>
            <a
              v-if="readerStore.selectedArticle.url"
              :href="readerStore.selectedArticle.url"
              target="_blank"
              rel="noopener"
              class="flex items-center gap-1 text-blue-600 dark:text-blue-400 hover:underline"
            >
              Original
              <span class="material-symbols-outlined text-[14px]">open_in_new</span>
            </a>
          </div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight leading-tight">{{ readerStore.selectedArticle.title }}</h1>

          <!-- Score badges -->
          <div class="flex items-center gap-3">
            <span class="text-xs text-gray-500 dark:text-gray-400">
              Innovation <span class="font-semibold text-blue-600 dark:text-blue-400">{{ readerStore.selectedArticle.innovationScore }}/10</span>
            </span>
            <span class="text-xs text-gray-500 dark:text-gray-400">
              Depth <span class="font-semibold text-green-600 dark:text-green-400">{{ readerStore.selectedArticle.depthScore }}/10</span>
            </span>
          </div>
        </div>

        <!-- Article Body -->
        <div class="prose prose-gray dark:prose-invert max-w-none text-base text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-line">{{ readerStore.selectedArticle.content }}</div>

        <!-- Reasoning -->
        <div v-if="readerStore.selectedArticle.reasoning" class="mt-10 pt-6 border-t border-gray-100 dark:border-gray-700">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">AI Reasoning</h4>
          <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{{ readerStore.selectedArticle.reasoning }}</p>
        </div>

        <!-- Bottom Actions -->
        <div class="mt-10 pt-6 border-t border-gray-100 dark:border-gray-700 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <button
              v-if="readerStore.selectedArticle.url"
              @click="openUrl(readerStore.selectedArticle.url)"
              class="px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 flex items-center gap-2 text-sm font-medium transition-colors"
            >
              <span class="material-symbols-outlined text-sm">open_in_new</span>
              View Original
            </button>
          </div>
          <button
            @click="$router.push('/timeline')"
            class="text-sm font-medium text-gray-500 hover:text-gray-900 dark:hover:text-white transition-colors"
          >
            Back to Timeline
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useReaderStore } from '@/stores/useReaderStore'

const route = useRoute()
const readerStore = useReaderStore()

const openUrl = (url) => {
  if (url) window.open(url, '_blank')
}

onMounted(async () => {
  await readerStore.loadArticles()

  // If navigated with an article ID, select it
  if (route.params.id) {
    readerStore.selectArticle(parseInt(route.params.id))
  }
})

// Watch for route param changes
watch(() => route.params.id, (newId) => {
  if (newId) {
    readerStore.selectArticle(parseInt(newId))
  }
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}
</style>
