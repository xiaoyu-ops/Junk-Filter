<template>
  <div class="w-full flex flex-col pt-20">
    <!-- ✨ RSS 抓取进度条（新增） -->
    <div class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border-b border-blue-200 dark:border-blue-800 px-4 sm:px-6 lg:px-8 py-4">
      <div class="max-w-7xl mx-auto">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">RSS 抓取进度</h3>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <!-- Pending -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-3 shadow-sm">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="material-symbols-outlined text-[20px] text-amber-500">schedule</span>
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400">待处理</span>
              </div>
              <span class="text-lg font-bold text-amber-600 dark:text-amber-400">{{ timelineStore.stats.pending }}</span>
            </div>
            <div class="mt-2 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
              <div
                class="bg-amber-500 h-1.5 rounded-full transition-all duration-300"
                :style="{ width: totalWidth > 0 ? (timelineStore.stats.pending / timelineStore.stats.total * 100) + '%' : '0%' }">
              </div>
            </div>
          </div>

          <!-- Processing -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-3 shadow-sm">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="material-symbols-outlined text-[20px] text-blue-500 animate-spin">sync</span>
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400">评估中</span>
              </div>
              <span class="text-lg font-bold text-blue-600 dark:text-blue-400">{{ timelineStore.stats.processing }}</span>
            </div>
            <div class="mt-2 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
              <div
                class="bg-blue-500 h-1.5 rounded-full transition-all duration-300"
                :style="{ width: totalWidth > 0 ? (timelineStore.stats.processing / timelineStore.stats.total * 100) + '%' : '0%' }">
              </div>
            </div>
          </div>

          <!-- Evaluated -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-3 shadow-sm">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="material-symbols-outlined text-[20px] text-green-500">check_circle</span>
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400">已评估</span>
              </div>
              <span class="text-lg font-bold text-green-600 dark:text-green-400">{{ timelineStore.stats.evaluated }}</span>
            </div>
            <div class="mt-2 w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
              <div
                class="bg-green-500 h-1.5 rounded-full transition-all duration-300"
                :style="{ width: totalWidth > 0 ? (timelineStore.stats.evaluated / timelineStore.stats.total * 100) + '%' : '0%' }">
              </div>
            </div>
          </div>

          <!-- Total -->
          <div class="bg-white dark:bg-gray-800 rounded-lg p-3 shadow-sm">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="material-symbols-outlined text-[20px] text-gray-500">article</span>
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400">总数</span>
              </div>
              <span class="text-lg font-bold text-gray-600 dark:text-gray-400">{{ timelineStore.stats.total }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 过滤按钮 + 加载状态 -->
    <div class="container mx-auto px-4 sm:px-6 lg:px-8 mt-8 mb-6">
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="filter in filters"
          :key="filter"
          @click="timelineStore.setFilter(filter)"
          :class="[
            'px-4 py-1.5 rounded-full text-sm font-medium transition-colors',
            timelineStore.activeFilter === filter
              ? 'bg-black text-white dark:bg-white dark:text-gray-900'
              : 'text-gray-600 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-800'
          ]"
        >
          {{ filter }}
        </button>
        <button
          @click="timelineStore.refreshContent"
          :disabled="timelineStore.isLoading"
          class="px-4 py-1.5 rounded-full text-sm font-medium text-gray-600 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-800 transition-colors flex items-center gap-1 ml-auto sm:ml-0 disabled:opacity-50"
        >
          <span class="material-symbols-outlined text-[18px]">refresh</span>
          刷新
        </button>
        <button
          v-if="!timelineStore.isStopped"
          @click="handleStopEvaluation"
          :disabled="timelineStore.isStopping || (timelineStore.stats.pending === 0 && timelineStore.stats.processing === 0)"
          class="px-4 py-1.5 rounded-full text-sm font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20 transition-colors flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span class="material-symbols-outlined text-[18px]">{{ timelineStore.isStopping ? 'sync' : 'stop_circle' }}</span>
          {{ timelineStore.isStopping ? '终止中...' : '终止评估' }}
        </button>
        <button
          v-else
          @click="handleRestartEvaluation"
          :disabled="timelineStore.isStopping"
          class="px-4 py-1.5 rounded-full text-sm font-medium text-green-600 hover:bg-green-50 dark:text-green-400 dark:hover:bg-green-900/20 transition-colors flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span class="material-symbols-outlined text-[18px]">{{ timelineStore.isStopping ? 'sync' : 'play_circle' }}</span>
          {{ timelineStore.isStopping ? '重启中...' : '重启评估' }}
        </button>
      </div>
      <!-- 加载状态指示 -->
      <div v-if="timelineStore.isLoading" class="mt-3 text-sm text-gray-500 dark:text-gray-400">
        <span class="inline-flex items-center gap-1">
          <span class="animate-spin material-symbols-outlined text-[16px]">sync</span>
          加载已评估内容中...
        </span>
      </div>
      <!-- 错误提示 -->
      <div v-if="timelineStore.error" class="mt-3 text-sm text-red-600 dark:text-red-400">
        错误: {{ timelineStore.error }}
      </div>
    </div>

    <!-- 卡片网格 -->
    <main class="flex-grow container mx-auto px-4 sm:px-6 lg:px-8 pb-12">
      <div v-if="timelineStore.filteredCards.length === 0" class="text-center py-12">
        <p class="text-gray-500 dark:text-gray-400">No evaluated content available</p>
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <article
          v-for="card in timelineStore.filteredCards"
          :key="card.id"
          @click="timelineStore.openDetailDrawer(card)"
          class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-[#374151] shadow-soft dark:shadow-none overflow-hidden flex flex-col h-full hover:shadow-md hover:scale-[1.02] transition-all duration-200 cursor-pointer"
        >
          <div class="p-6 flex gap-5">
            <!-- 作者头像和信息 -->
            <div class="flex flex-col items-center gap-2 min-w-[72px]">
              <div class="w-14 h-14 rounded-full bg-gray-50 dark:bg-gray-800 flex items-center justify-center overflow-hidden border border-gray-100 dark:border-gray-700">
                <img v-if="card.faviconUrl" :src="card.faviconUrl" :alt="card.author" class="w-8 h-8 object-contain" @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''" />
                <span v-if="card.faviconUrl" class="material-symbols-outlined text-3xl text-gray-300 dark:text-gray-500" style="display:none">person</span>
                <span v-if="!card.faviconUrl" class="material-symbols-outlined text-3xl text-gray-300 dark:text-gray-500">person</span>
              </div>
              <h3 class="font-semibold text-sm text-center text-gray-900 dark:text-white line-clamp-2">{{ card.author }}</h3>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ card.authorTime }}</span>
            </div>

            <!-- 文章内容 -->
            <div class="flex-1 space-y-2">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white leading-tight line-clamp-2">{{ card.title }}</h2>
              <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed line-clamp-3">{{ card.content || card.tldr }}</p>

              <!-- 评估分数显示 -->
              <div class="flex gap-3 mt-3 pt-2 border-t border-gray-100 dark:border-gray-700">
                <div class="text-xs">
                  <span class="text-gray-500 dark:text-gray-400">Innovation:</span>
                  <span class="font-semibold text-blue-600 dark:text-blue-400 ml-1">{{ card.innovationScore }}/10</span>
                </div>
                <div class="text-xs">
                  <span class="text-gray-500 dark:text-gray-400">Depth:</span>
                  <span class="font-semibold text-green-600 dark:text-green-400 ml-1">{{ card.depthScore }}/10</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部状态 -->
          <div class="mt-auto border-t border-gray-100 dark:border-[#374151]">
            <div class="px-6 py-3 flex items-center justify-between bg-gray-50 dark:bg-[#1F2937]/50">
              <div class="flex items-center gap-2">
                <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Decision:</span>
                <div :class="[
                  'flex items-center gap-1.5',
                  card.statusColor === 'green' ? 'text-green-600 dark:text-green-400' :
                  card.statusColor === 'red' ? 'text-red-600 dark:text-red-400' :
                  'text-amber-600 dark:text-amber-400'
                ]">
                  <span class="material-symbols-outlined text-[18px]">
                    {{ card.statusColor === 'green' ? 'check_circle' : card.statusColor === 'red' ? 'cancel' : 'bookmark' }}
                  </span>
                  <span class="text-xs font-semibold">{{ card.status }}</span>
                </div>
              </div>
              <button class="text-xs font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white transition-colors">
                Details
              </button>
            </div>
          </div>
        </article>
      </div>
    </main>

    <!-- 侧滑抽屉 -->
    <Transition
      enter-active-class="transition-all duration-400"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-all duration-300"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="timelineStore.isDetailDrawerOpen"
        @click="timelineStore.closeDetailDrawer"
        class="fixed inset-0 bg-black/30 z-30"
      ></div>
    </Transition>

    <Transition
      enter-active-class="transition-all duration-400 ease-out"
      enter-from-class="translate-x-full opacity-0"
      enter-to-class="translate-x-0 opacity-100"
      leave-active-class="transition-all duration-300 ease-out"
      leave-from-class="translate-x-0 opacity-100"
      leave-to-class="translate-x-full opacity-0"
    >
      <div
        v-if="timelineStore.isDetailDrawerOpen && timelineStore.selectedCard"
        class="fixed right-0 top-0 h-full w-96 bg-white dark:bg-[#1F2937] shadow-2xl z-40 flex flex-col overflow-hidden"
      >
        <!-- 抽屉头部 -->
        <div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">Details</h2>
          <button
            @click="timelineStore.closeDetailDrawer"
            class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
          >
            <span class="material-icons-outlined">close</span>
          </button>
        </div>

        <!-- 抽屉内容 -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6">
          <!-- 作者信息 -->
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center flex-shrink-0 overflow-hidden">
              <img v-if="timelineStore.selectedCard.faviconUrl" :src="timelineStore.selectedCard.faviconUrl" :alt="timelineStore.selectedCard.author" class="w-10 h-10 object-contain" @error="$event.target.style.display='none'; $event.target.nextElementSibling.style.display=''" />
              <span v-if="timelineStore.selectedCard.faviconUrl" class="material-icons-outlined text-3xl text-gray-400 dark:text-gray-500" style="display:none">person</span>
              <span v-if="!timelineStore.selectedCard.faviconUrl" class="material-icons-outlined text-3xl text-gray-400 dark:text-gray-500">person</span>
            </div>
            <div class="min-w-0 flex-1">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white truncate">{{ timelineStore.selectedCard.author }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ timelineStore.selectedCard.authorTime }}</p>
            </div>
          </div>

          <!-- 文章标题 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">Title</h4>
            <p class="text-base text-gray-700 dark:text-gray-300 leading-relaxed">{{ timelineStore.selectedCard.title }}</p>
          </div>

          <!-- AI摘要 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">摘要</h4>
            <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{{ timelineStore.selectedCard.tldr }}</p>
          </div>

          <!-- 文章内容 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">Content</h4>
            <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed max-h-32 overflow-y-auto">{{ timelineStore.selectedCard.content }}</p>
          </div>

          <!-- 评分详情 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">AI Evaluation Scores</h4>
            <div class="space-y-4">
              <!-- 创新度 -->
              <div>
                <div class="flex justify-between text-xs text-gray-600 dark:text-gray-400 mb-1">
                  <span>Innovation Score</span>
                  <span class="font-semibold">{{ timelineStore.selectedCard.innovationScore }}/10</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div class="bg-blue-500 h-2 rounded-full" :style="{ width: (timelineStore.selectedCard.innovationScore * 10) + '%' }"></div>
                </div>
              </div>

              <!-- 深度 -->
              <div>
                <div class="flex justify-between text-xs text-gray-600 dark:text-gray-400 mb-1">
                  <span>Depth Score</span>
                  <span class="font-semibold">{{ timelineStore.selectedCard.depthScore }}/10</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div class="bg-green-500 h-2 rounded-full" :style="{ width: (timelineStore.selectedCard.depthScore * 10) + '%' }"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- 决策说明 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">Decision</h4>
            <div :class="[
              'inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium',
              timelineStore.selectedCard.statusColor === 'green' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' :
              timelineStore.selectedCard.statusColor === 'red' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
              'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
            ]">
              <span class="material-symbols-outlined text-[18px]">{{ timelineStore.selectedCard.statusColor === 'green' ? 'check_circle' : timelineStore.selectedCard.statusColor === 'red' ? 'cancel' : 'bookmark' }}</span>
              {{ timelineStore.selectedCard.status }}
            </div>
          </div>

          <!-- 推理过程 -->
          <div v-if="timelineStore.selectedCard.reasoning">
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">Reasoning</h4>
            <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{{ timelineStore.selectedCard.reasoning }}</p>
          </div>

          <!-- 关键概念 -->
          <div v-if="timelineStore.selectedCard.keyConepts && timelineStore.selectedCard.keyConepts.length > 0">
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">Key Concepts</h4>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="(concept, idx) in timelineStore.selectedCard.keyConepts"
                :key="idx"
                class="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 text-xs rounded-full"
              >
                {{ concept }}
              </span>
            </div>
          </div>
        </div>

        <!-- 抽屉底部按钮 -->
        <div class="p-6 border-t border-gray-200 dark:border-gray-700 flex gap-3">
          <button
            @click="goToReader(timelineStore.selectedCard.id)"
            class="flex-1 px-4 py-2.5 bg-gray-900 hover:bg-gray-800 dark:bg-white dark:hover:bg-gray-200 text-white dark:text-gray-900 rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <span class="material-symbols-outlined text-[18px]">menu_book</span>
            Read
          </button>
          <button
            @click="openUrl(timelineStore.selectedCard.url)"
            class="flex-1 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <span class="material-symbols-outlined text-[18px]">open_in_new</span>
            View Original
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useTimelineStore } from '@/stores'

const router = useRouter()
const timelineStore = useTimelineStore()
const filters = ['All', 'Interesting', 'Bookmark', 'Skip']

/**
 * 计算统计数据总和（用于进度条计算）
 */
const totalWidth = computed(() => {
  return timelineStore.stats.total || 0
})

/**
 * 在新标签页打开 URL
 */
const openUrl = async (url) => {
  if (!url) return
  try {
    const { open } = await import('@tauri-apps/plugin-shell')
    await open(url)
  } catch {
    window.open(url, '_blank')
  }
}

const goToReader = (articleId) => {
  timelineStore.closeDetailDrawer()
  router.push(`/reader/${articleId}`)
}

const handleStopEvaluation = async () => {
  if (!confirm('确定要终止所有待处理和评估中的内容吗？此操作可通过重启评估恢复。')) return
  try {
    const affected = await timelineStore.stopEvaluation()
    alert(`已终止评估，${affected} 条内容被暂停`)
  } catch {
    alert('终止评估失败，请重试')
  }
}

const handleRestartEvaluation = async () => {
  try {
    const affected = await timelineStore.restartEvaluation()
    alert(`已重启评估，${affected} 条内容恢复待处理`)
  } catch {
    alert('重启评估失败，请重试')
  }
}

/**
 * 组件挂载时初始化时间轴
 */
onMounted(() => {
  timelineStore.initialize()
})
</script>
