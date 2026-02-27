import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useTimelineStore = defineStore('timeline', () => {
  const activeFilter = ref('All')
  const isDetailDrawerOpen = ref(false)
  const selectedCard = ref(null)

  const cards = ref([
    {
      id: 1,
      author: 'TechDaily',
      authorTime: '2h ago',
      title: 'AI Model Breakdown',
      content: 'A comprehensive look into the new architecture changes proposed in the latest research paper. It highlights significant improvements in processing efficiency and reduced latency for real-time applications.',
      status: 'Approved',
      statusColor: 'green',
    },
    {
      id: 2,
      author: 'DesignPro',
      authorTime: '4h ago',
      title: 'UI Trends 2024',
      content: 'Discussing the shift towards neo-brutalism and high contrast interfaces. While visually striking, concerns about accessibility and long-term user retention remain a topic of heated debate.',
      status: 'Rejected',
      statusColor: 'red',
    },
    {
      id: 3,
      author: 'CodeMaster',
      authorTime: '5h ago',
      title: 'Rust vs Go',
      content: 'Performance benchmarks comparing Rust and Go in high-concurrency scenarios. The results show Rust edging out in memory safety, while Go maintains superiority in development speed.',
      status: 'Approved',
      statusColor: 'green',
    },
    {
      id: 4,
      author: 'StartupLife',
      authorTime: '6h ago',
      title: 'Funding Winter?',
      content: 'Analyzing the current venture capital landscape. Despite the gloom, certain sectors like Generative AI and Climate Tech are seeing record investments.',
      status: 'Review',
      statusColor: 'amber',
    },
  ])

  // 设置过滤器
  const setFilter = (filter) => {
    activeFilter.value = filter
  }

  // 打开详情抽屉
  const openDetailDrawer = (card) => {
    selectedCard.value = card
    isDetailDrawerOpen.value = true
  }

  // 关闭详情抽屉
  const closeDetailDrawer = () => {
    isDetailDrawerOpen.value = false
    selectedCard.value = null
  }

  return {
    activeFilter,
    isDetailDrawerOpen,
    selectedCard,
    cards,
    setFilter,
    openDetailDrawer,
    closeDetailDrawer,
  }
})
