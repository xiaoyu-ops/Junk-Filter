<template>
  <div class="flex flex-col px-8 py-6 gap-4 h-[calc(100vh-80px)]">

    <!-- 第一行：4 个统计卡片 -->
    <div class="flex-shrink-0 grid grid-cols-4 gap-4">

      <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-yellow-100 dark:bg-yellow-900/30 flex items-center justify-center flex-shrink-0">
          <span class="material-icons-outlined text-yellow-600 dark:text-yellow-400">hourglass_empty</span>
        </div>
        <div>
          <p class="text-2xl font-bold text-gray-900 dark:text-white leading-none">{{ stats.pending }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">待评估</p>
        </div>
      </div>

      <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center flex-shrink-0">
          <span class="material-icons-outlined text-green-600 dark:text-green-400">check_circle</span>
        </div>
        <div>
          <p class="text-2xl font-bold text-gray-900 dark:text-white leading-none">{{ stats.evaluated }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">已评估</p>
        </div>
      </div>

      <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
          <span class="material-icons-outlined text-red-600 dark:text-red-400">delete_sweep</span>
        </div>
        <div>
          <p class="text-2xl font-bold text-gray-900 dark:text-white leading-none">{{ stats.discarded }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">已丢弃</p>
        </div>
      </div>

      <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
          <span class="material-icons-outlined text-blue-600 dark:text-blue-400">rss_feed</span>
        </div>
        <div>
          <p class="text-2xl font-bold text-gray-900 dark:text-white leading-none">{{ stats.sources }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">RSS 源</p>
        </div>
      </div>

    </div>

    <!-- 第二行：左侧折线图 + 右侧执行状态 -->
    <div class="flex-shrink-0 grid grid-cols-5 gap-4" style="height: 200px;">

      <!-- 左侧：评估趋势折线图（占 3/5） -->
      <div class="col-span-3 bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 flex flex-col">
        <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">近 7 天评估趋势</p>
        <div class="flex-1 min-h-0">
          <Line v-if="chartReady" :data="chartData" :options="chartOptions" />
        </div>
      </div>

      <!-- 右侧：执行状态面板（占 2/5） -->
      <div class="col-span-2 bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700/50 p-4 flex flex-col gap-2 overflow-hidden">
        <p class="text-sm font-medium text-gray-700 dark:text-gray-300 flex-shrink-0">执行状态</p>

        <!-- 队列 & 评估中 -->
        <div class="flex gap-3 flex-shrink-0">
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-yellow-400 flex-shrink-0"></span>
            队列 {{ stats.pending }} 篇
          </div>
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-blue-400 flex-shrink-0 animate-pulse"></span>
            评估中 {{ stats.processing }} 篇
          </div>
        </div>

        <!-- 最近抓取的源（可点击跳转到配置页） -->
        <div class="flex-1 overflow-y-auto space-y-1 min-h-0">
          <p class="text-xs text-gray-400 dark:text-gray-500 mb-1">最近抓取</p>
          <button
            v-for="src in recentSources"
            :key="src.id"
            @click="goToConfig"
            class="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors text-left group"
          >
            <img
              v-if="src.favicon_url"
              :src="src.favicon_url"
              class="w-4 h-4 rounded-sm flex-shrink-0 object-contain"
              @error="e => e.target.style.display='none'"
            />
            <span v-else class="material-icons-outlined text-gray-400 dark:text-gray-500 flex-shrink-0" style="font-size:14px">rss_feed</span>
            <span class="text-xs text-gray-700 dark:text-gray-300 truncate flex-1 group-hover:text-indigo-600 dark:group-hover:text-indigo-400">
              {{ src.name }}
            </span>
            <span class="text-xs text-gray-400 dark:text-gray-500 flex-shrink-0">{{ src.timeAgo }}</span>
          </button>
          <p v-if="recentSources.length === 0" class="text-xs text-gray-400 dark:text-gray-500 px-2">暂无抓取记录</p>
        </div>
      </div>

    </div>

    <!-- 第三行：Agent 对话区（占满剩余高度） -->
    <div class="flex-1 min-h-0">
      <TaskChat />
    </div>

  </div>
</template>

<script>
export default { name: 'TaskDistribution' }
</script>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useTaskStore } from '@/stores/useTaskStore'
import TaskChat from './TaskChat.vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale, LinearScale, PointElement, LineElement,
  Title, Tooltip, Filler,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Filler)

const router = useRouter()
const taskStore = useTaskStore()
const API_BASE = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080'

const stats = ref({ pending: 0, processing: 0, evaluated: 0, discarded: 0, sources: 0 })
const recentSources = ref([])
const timelineData = ref([])
const chartReady = ref(false)

// ── 图表配置 ──────────────────────────────────────
const chartData = computed(() => ({
  labels: timelineData.value.map(d => d.date.slice(5)),  // MM-DD
  datasets: [{
    label: '已评估',
    data: timelineData.value.map(d => d.count),
    borderColor: '#6366f1',
    backgroundColor: 'rgba(99,102,241,0.12)',
    borderWidth: 2,
    pointRadius: 3,
    pointBackgroundColor: '#6366f1',
    fill: true,
    tension: 0.35,
  }],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false }, title: { display: false } },
  scales: {
    x: {
      grid: { display: false },
      ticks: { color: '#9ca3af', font: { size: 11 } },
    },
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(156,163,175,0.15)' },
      ticks: { color: '#9ca3af', font: { size: 11 }, precision: 0 },
    },
  },
}

// ── 时间格式化 ─────────────────────────────────────
const timeAgo = (isoStr) => {
  if (!isoStr) return ''
  const diff = Date.now() - new Date(isoStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins}分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}小时前`
  return `${Math.floor(hours / 24)}天前`
}

// ── 跳转配置页 ─────────────────────────────────────
const goToConfig = () => router.push('/config')

// ── 数据加载 ───────────────────────────────────────
const loadStats = async () => {
  try {
    const [contentRes, sourcesRes, timelineRes] = await Promise.all([
      fetch(`${API_BASE}/api/content/stats`),
      fetch(`${API_BASE}/api/sources`),
      fetch(`${API_BASE}/api/content/stats/timeline?days=7`),
    ])

    if (contentRes.ok) {
      const data = await contentRes.json()
      stats.value.pending    = data.pending    || 0
      stats.value.processing = data.processing || 0
      stats.value.evaluated  = data.evaluated  || 0
      stats.value.discarded  = data.discarded  || 0
    }

    if (sourcesRes.ok) {
      const data = await sourcesRes.json()
      const list = Array.isArray(data.data) ? data.data : (Array.isArray(data) ? data : [])
      stats.value.sources = list.length

      // 取最近抓取的前5个源（有 last_fetch_time 的）
      recentSources.value = list
        .filter(s => s.last_fetch_time)
        .sort((a, b) => new Date(b.last_fetch_time) - new Date(a.last_fetch_time))
        .slice(0, 5)
        .map(s => ({
          id: s.id,
          name: s.author_name || s.url,
          favicon_url: s.favicon_url || null,
          timeAgo: timeAgo(s.last_fetch_time),
        }))
    }

    if (timelineRes.ok) {
      const data = await timelineRes.json()
      timelineData.value = data.timeline || []
      chartReady.value = true
    }
  } catch (err) {
    console.warn('[Dashboard] Failed to load stats:', err)
  }
}

onMounted(async () => {
  await taskStore.loadTasks()
  loadStats()
})
</script>
