import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/components/Home.vue'
import Timeline from '@/components/Timeline.vue'
import Config from '@/components/Config.vue'
import TaskDistribution from '@/components/TaskDistribution.vue'
import AppNavbar from '@/components/AppNavbar.vue'

const routes = [
  {
    path: '/',
    components: {
      default: Home,
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - 主页' },
  },
  {
    path: '/timeline',
    components: {
      default: Timeline,
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - Timeline' },
  },
  {
    path: '/config',
    components: {
      default: Config,
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - 配置中心' },
  },
  {
    path: '/task',
    components: {
      default: TaskDistribution,
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - 分发任务' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 页面title更新
router.afterEach((to) => {
  document.title = to.meta.title || 'Junk Filter'
})

export default router
