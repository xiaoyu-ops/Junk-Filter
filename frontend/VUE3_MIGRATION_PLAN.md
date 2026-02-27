# 🚀 Vue 3 + Vite + Pinia 完整迁移计划

**更新时间**：2026-02-26
**计划版本**：2.0（含环境变量 + Pinia 全局状态管理）
**预期交付**：完整的工程化Vue 3项目

---

## 📋 **目录**
1. [第一阶段：工程化初始化](#第一阶段工程化初始化)
2. [第二阶段：环境变量安全控制](#第二阶段环境变量安全控制)
3. [第三阶段：Pinia全局状态管理](#第三阶段pinia全局状态管理)
4. [第四阶段：组件提取与抽象](#第四阶段组件提取与抽象)
5. [第五阶段：路由配置](#第五阶段路由配置)
6. [第六阶段：样式迁移](#第六阶段样式迁移)
7. [第七阶段：逻辑整合](#第七阶段逻辑整合)
8. [检查清单](#检查清单)

---

## 第一阶段：工程化初始化

### 1.1 项目结构
```
frontend-vue/
├── .env                            # 🔐 环境变量（本地开发）
├── .env.example                    # 环境变量模板（示例）
├── .env.production                 # 生产环境配置
├── .gitignore                      # Git忽略规则（包含.env）
├── index.html                      # 单一入口
├── vite.config.js                  # Vite配置
├── tailwind.config.js              # Tailwind CSS配置
├── postcss.config.js               # PostCSS配置
├── package.json                    # 依赖管理
├── src/
│   ├── main.js                     # Vue应用入口
│   ├── App.vue                     # 根组件
│   ├── components/
│   │   ├── AppNavbar.vue           # 统一导航条
│   │   ├── Home.vue                # 主页
│   │   ├── Config.vue              # 配置中心
│   │   ├── Timeline.vue            # 时间轴
│   │   └── Task.vue                # 分发任务
│   ├── router/
│   │   └── index.js                # vue-router配置
│   ├── stores/                     # 🔑 Pinia状态管理
│   │   ├── index.js                # 导出所有store
│   │   ├── useThemeStore.js        # 主题状态
│   │   ├── useConfigStore.js       # 配置中心状态（API Key等）
│   │   ├── useTaskStore.js         # 任务管理状态
│   │   └── useTimelineStore.js     # 时间轴状态
│   ├── composables/                # Composition API逻辑
│   │   ├── useSearch.js            # 搜索框交互
│   │   ├── useDetailDrawer.js      # 侧滑抽屉
│   │   └── useToast.js             # Toast提示
│   ├── utils/
│   │   ├── api.js                  # API请求封装
│   │   └── constants.js            # 常量定义
│   ├── assets/                     # 静态资源
│   └── styles/
│       └── globals.css             # 全局样式
└── public/                         # 公共文件
```

### 1.2 核心依赖

```json
{
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.3.0",
    "pinia": "^2.1.0"
  },
  "devDependencies": {
    "vite": "^5.0.0",
    "tailwindcss": "^3.4.0",
    "postcss": "^8.4.0",
    "autoprefixer": "^10.4.0"
  }
}
```

---

## 第二阶段：环境变量安全控制

### 2.1 为什么需要环境变量？

❌ **反面示例（绝对不能这样）：**
```javascript
// ❌ 危险：API Key硬编码在代码里
const API_KEY = 'sk-proj-xxxxxxxxxxxxxxxxxxxxxx'
const API_URL = 'https://api.example.com'
```

✅ **正确做法：使用环境变量**
```javascript
// ✅ 安全：从环境中读取
const API_KEY = import.meta.env.VITE_API_KEY
const API_URL = import.meta.env.VITE_API_URL
```

### 2.2 .env 文件配置

**`.env`（本地开发，Git忽略）**
```bash
# API配置
VITE_API_KEY=sk-proj-your-actual-key-here
VITE_API_URL=http://localhost:8000/api
VITE_API_MODEL=GPT-4o

# 应用配置
VITE_APP_NAME=Junk Filter
VITE_APP_ENV=development
VITE_LOG_LEVEL=debug

# 特性开关
VITE_ENABLE_MOCK_DATA=true
```

**`.env.example`（版本控制，展示结构）**
```bash
# 复制此文件为 .env 并填入实际值

# API配置
VITE_API_KEY=your-api-key-here
VITE_API_URL=http://localhost:8000/api
VITE_API_MODEL=GPT-4o

# 应用配置
VITE_APP_NAME=Junk Filter
VITE_APP_ENV=development
VITE_LOG_LEVEL=debug

# 特性开关
VITE_ENABLE_MOCK_DATA=true
```

**`.env.production`（生产环境）**
```bash
# 生产环境配置（不含真实API Key，通过CI/CD注入）
VITE_API_KEY=${CI_API_KEY}
VITE_API_URL=https://api.production.example.com
VITE_API_MODEL=GPT-4o

# 应用配置
VITE_APP_NAME=Junk Filter
VITE_APP_ENV=production
VITE_LOG_LEVEL=error

# 特性开关
VITE_ENABLE_MOCK_DATA=false
```

### 2.3 环境变量使用示例

**在Vue组件中：**
```javascript
// src/components/Config.vue
const apiKey = ref(import.meta.env.VITE_API_KEY)
const apiUrl = import.meta.env.VITE_API_URL
const appEnv = import.meta.env.VITE_APP_ENV

console.log(`运行环境: ${appEnv}`) // development / production
```

**在Store中：**
```javascript
// src/stores/useConfigStore.js
import { defineStore } from 'pinia'

export const useConfigStore = defineStore('config', {
  state: () => ({
    apiKey: import.meta.env.VITE_API_KEY || '',
    apiUrl: import.meta.env.VITE_API_URL || 'http://localhost:8000/api',
    apiModel: import.meta.env.VITE_API_MODEL || 'GPT-4o',
  }),
})
```

### 2.4 .gitignore 配置

```bash
# 环境变量（绝对不提交真实数据）
.env
.env.local
.env.*.local

# 构建文件
dist/
.dist/

# 依赖
node_modules/

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db
```

---

## 第三阶段：Pinia全局状态管理

### 3.1 为什么必须用Pinia？

**场景：用户在"配置中心"修改API Key → 跳转到"分发任务"**

❌ **没有Pinia（数据丢失）：**
```
配置中心 → 用户修改API Key → localStorage存储
        ↓
跳转到任务页面 → 从localStorage读取？可能还要等待？
        ↓
最坏情况：两个页面的API Key不一致！
```

✅ **使用Pinia（数据共享）：**
```
配置中心 → 用户修改API Key → Pinia store更新（内存）+ localStorage持久化
        ↓
跳转到任务页面 → 直接从Pinia读取（毫秒级，无延迟）
        ↓
两个页面的API Key始终同步！
```

### 3.2 核心Store设计

#### 3.2.1 useThemeStore.js（主题管理）
```javascript
// src/stores/useThemeStore.js
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // 状态
  const isDark = ref(
    localStorage.getItem('theme') === 'dark' ||
    (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches)
  )

  // 计算属性
  const theme = computed(() => isDark.value ? 'dark' : 'light')

  // 方法
  const toggleTheme = () => {
    isDark.value = !isDark.value
    updateDOM()
  }

  const updateDOM = () => {
    const html = document.documentElement
    if (isDark.value) {
      html.classList.add('dark')
      localStorage.setItem('theme', 'dark')
    } else {
      html.classList.remove('dark')
      localStorage.setItem('theme', 'light')
    }
  }

  // 初始化
  onMounted(() => {
    updateDOM()
  })

  // 监听主题变化
  watch(isDark, () => {
    updateDOM()
  })

  return { isDark, theme, toggleTheme }
})
```

#### 3.2.2 useConfigStore.js（配置管理 - 核心）
```javascript
// src/stores/useConfigStore.js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useConfigStore = defineStore('config', () => {
  // 状态：直接从环境变量初始化
  const apiKey = ref(import.meta.env.VITE_API_KEY || '')
  const apiUrl = ref(import.meta.env.VITE_API_URL || 'http://localhost:8000/api')
  const apiModel = ref(import.meta.env.VITE_API_MODEL || 'GPT-4o')
  const temperature = ref(0.7)
  const maxTokens = ref(2048)
  const isSaving = ref(false)
  const saveStatus = ref(null) // 'success' | 'error' | null

  // 计算属性
  const isConfigValid = computed(() => apiKey.value.length > 0)

  // 方法：保存配置
  const saveConfig = async () => {
    isSaving.value = true
    saveStatus.value = null

    try {
      // 模拟API请求（20%失败率）
      await new Promise(resolve => setTimeout(resolve, 1000))

      const isSuccess = Math.random() > 0.2

      if (isSuccess) {
        // 保存到localStorage
        localStorage.setItem('config', JSON.stringify({
          apiModel: apiModel.value,
          temperature: temperature.value,
          maxTokens: maxTokens.value,
          // 注意：API Key不存储到localStorage，始终从环境变量读取
        }))
        saveStatus.value = 'success'
        return true
      } else {
        saveStatus.value = 'error'
        throw new Error('保存配置失败')
      }
    } catch (error) {
      saveStatus.value = 'error'
      console.error('Config save error:', error)
      return false
    } finally {
      isSaving.value = false
    }
  }

  // 方法：更新API Key（来自环境或用户输入）
  const updateApiKey = (key) => {
    apiKey.value = key
  }

  // 方法：更新温度
  const updateTemperature = (temp) => {
    temperature.value = parseFloat(temp)
  }

  // 方法：初始化配置（从localStorage恢复）
  const loadConfig = () => {
    const saved = localStorage.getItem('config')
    if (saved) {
      const config = JSON.parse(saved)
      apiModel.value = config.apiModel || apiModel.value
      temperature.value = config.temperature || temperature.value
      maxTokens.value = config.maxTokens || maxTokens.value
    }
  }

  return {
    // 状态
    apiKey,
    apiUrl,
    apiModel,
    temperature,
    maxTokens,
    isSaving,
    saveStatus,

    // 计算属性
    isConfigValid,

    // 方法
    saveConfig,
    updateApiKey,
    updateTemperature,
    loadConfig,
  }
}, {
  persist: {
    // 可选：使用 pinia-plugin-persistedstate 自动持久化
    enabled: true,
    strategies: [
      {
        key: 'config',
        storage: localStorage,
        paths: ['apiModel', 'temperature', 'maxTokens'],
        // 注意：故意不持久化 apiKey，始终从环境变量读取
      }
    ]
  }
})
```

**关键点：**
- ✅ API Key 从环境变量初始化，不从localStorage读取
- ✅ 其他配置（温度、Model、Token）可以持久化
- ✅ 提供 `saveConfig()` 方法，包含失败处理
- ✅ 提供 `loadConfig()` 方法，用于页面初始化

#### 3.2.3 useTaskStore.js（任务管理）
```javascript
// src/stores/useTaskStore.js
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useTaskStore = defineStore('task', () => {
  const messages = ref([])
  const isLoading = ref(false)
  const activeTaskId = ref(null)
  const tasks = ref([
    { id: 1, name: '每日新闻摘要 - 09:00 AM', status: 'active' },
    { id: 2, name: '每周数据报告 - 周一 10:00 AM', status: 'inactive' },
    { id: 3, name: '社交媒体监控 - 每小时', status: 'inactive' },
  ])

  // 方法：添加消息
  const addMessage = (role, content, type = 'text') => {
    messages.value.push({
      id: Date.now(),
      role, // 'user' | 'ai'
      content,
      type, // 'text' | 'error'
      timestamp: new Date(),
    })
  }

  // 方法：发送消息（集成useConfigStore的API Key）
  const sendMessage = async (inputText) => {
    if (!inputText.trim()) return

    addMessage('user', inputText)

    isLoading.value = true
    await new Promise(resolve => setTimeout(resolve, 800))

    // 这里可以访问其他store的数据
    const configStore = useConfigStore()
    console.log(`使用API Key: ${configStore.apiKey}`)

    const isSuccess = Math.random() > 0.3
    if (isSuccess) {
      const response = generateAIResponse()
      addMessage('ai', response, 'text')
    } else {
      addMessage('ai', '抱歉，AI 服务暂时不可用', 'error')
    }

    isLoading.value = false
  }

  // 方法：切换任务
  const switchTask = (taskId) => {
    activeTaskId.value = taskId
    messages.value = [] // 清空消息历史
  }

  return {
    messages,
    isLoading,
    activeTaskId,
    tasks,
    addMessage,
    sendMessage,
    switchTask,
  }
})
```

**关键点：**
- ✅ 在 `sendMessage()` 中可以直接访问 `useConfigStore()` 的 API Key
- ✅ 所有消息状态集中管理
- ✅ 任务切换时自动清空消息历史

#### 3.2.4 useTimelineStore.js（时间轴）
```javascript
// src/stores/useTimelineStore.js
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
      title: 'AI Model Breakdown',
      content: 'A comprehensive look...',
      status: 'Approved',
    },
    // ... 更多卡片
  ])

  // 方法
  const setFilter = (filter) => {
    activeFilter.value = filter
  }

  const openDetailDrawer = (card) => {
    selectedCard.value = card
    isDetailDrawerOpen.value = true
  }

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
```

### 3.3 Pinia in main.js

```javascript
// src/main.js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)

// 创建Pinia实例
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.mount('#app')
```

### 3.4 在组件中使用Store

```vue
<!-- src/components/Config.vue -->
<template>
  <div class="config-page">
    <!-- API Key显示（但不能直接编辑，这是从环境变量来的） -->
    <div class="api-key-display">
      <input
        v-model="configStore.apiKey"
        type="password"
        readonly
      />
      <button @click="copyApiKey">复制</button>
    </div>

    <!-- Temperature滑块（响应式绑定） -->
    <input
      type="range"
      v-model.number="configStore.temperature"
      min="0"
      max="1"
      step="0.1"
    />
    <span>{{ configStore.temperature }}</span>

    <!-- 保存按钮 -->
    <button
      @click="configStore.saveConfig"
      :disabled="configStore.isSaving"
    >
      {{ configStore.isSaving ? '保存中...' : '保存配置' }}
    </button>

    <!-- 状态提示 -->
    <div v-if="configStore.saveStatus === 'success'" class="toast success">
      配置已保存
    </div>
    <div v-if="configStore.saveStatus === 'error'" class="toast error">
      保存失败，请重试
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useConfigStore } from '@/stores/useConfigStore'

const configStore = useConfigStore()

// 初始化时加载配置
onMounted(() => {
  configStore.loadConfig()
})

// 复制API Key
const copyApiKey = async () => {
  try {
    await navigator.clipboard.writeText(configStore.apiKey)
    // 显示Toast提示
  } catch (err) {
    console.error('复制失败', err)
  }
}
</script>
```

```vue
<!-- src/components/Task.vue -->
<template>
  <div class="task-page">
    <!-- 消息列表（直接使用taskStore） -->
    <div class="messages">
      <div
        v-for="msg in taskStore.messages"
        :key="msg.id"
        :class="['message', msg.role, { error: msg.type === 'error' }]"
      >
        {{ msg.content }}
      </div>
    </div>

    <!-- 输入框 -->
    <input
      v-model="inputText"
      @keydown.enter="sendMessage"
      placeholder="输入消息..."
    />
    <button
      @click="sendMessage"
      :disabled="taskStore.isLoading"
    >
      {{ taskStore.isLoading ? '加载中...' : '发送' }}
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useTaskStore } from '@/stores/useTaskStore'
import { useConfigStore } from '@/stores/useConfigStore'

const taskStore = useTaskStore()
const configStore = useConfigStore()
const inputText = ref('')

const sendMessage = async () => {
  if (!inputText.value.trim()) return

  // taskStore会自动使用configStore的API Key
  await taskStore.sendMessage(inputText.value)
  inputText.value = ''
}
</script>
```

---

## 第四阶段：组件提取与抽象

### 4.1 AppNavbar.vue（统一导航）

```vue
<!-- src/components/AppNavbar.vue -->
<template>
  <header class="w-full px-8 py-4 flex items-center justify-between sticky top-0 z-50 bg-white dark:bg-[#0f0f11] backdrop-blur-sm border-b border-gray-100 dark:border-gray-800 transition-colors duration-200">
    <!-- Logo -->
    <div class="flex items-center gap-3">
      <span class="material-icons-outlined text-3xl text-gray-900 dark:text-gray-100">delete_outline</span>
      <h1 class="text-xl font-bold tracking-tight text-gray-900 dark:text-gray-100">Junk Filter</h1>
    </div>

    <!-- Navigation Links -->
    <nav class="flex items-center space-x-6">
      <RouterLink
        to="/"
        :class="['nav-link', { active: currentRoute === '/' }]"
      >
        主页
      </RouterLink>
      <RouterLink
        to="/timeline"
        :class="['nav-link', { active: currentRoute === '/timeline' }]"
      >
        时间轴
      </RouterLink>
      <RouterLink
        to="/config"
        :class="['nav-link', { active: currentRoute === '/config' }]"
      >
        配置中心
      </RouterLink>
      <RouterLink
        to="/task"
        :class="['nav-link', { active: currentRoute === '/task' }]"
      >
        分发任务
      </RouterLink>
    </nav>

    <!-- Theme Toggle -->
    <button
      @click="themeStore.toggleTheme"
      class="p-2 rounded-full text-gray-500 hover:bg-gray-100 dark:hover:bg-[#27272a] dark:text-gray-400 transition-colors"
    >
      <span v-if="!themeStore.isDark" class="material-icons-outlined">dark_mode</span>
      <span v-else class="material-icons-outlined">light_mode</span>
    </button>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useThemeStore } from '@/stores/useThemeStore'

const router = useRouter()
const themeStore = useThemeStore()

const currentRoute = computed(() => router.currentRoute.value.path)
</script>

<style scoped>
.nav-link {
  @apply text-sm font-medium text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 transition-colors;
}

.nav-link.active {
  @apply text-gray-900 dark:text-gray-100 font-semibold;
}
</style>
```

---

## 第五阶段：路由配置

```javascript
// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import AppNavbar from '@/components/AppNavbar.vue'

const routes = [
  {
    path: '/',
    components: {
      default: () => import('@/components/Home.vue'),
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - 主页' }
  },
  {
    path: '/timeline',
    components: {
      default: () => import('@/components/Timeline.vue'),
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - Timeline' }
  },
  {
    path: '/config',
    components: {
      default: () => import('@/components/Config.vue'),
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - 配置中心' }
  },
  {
    path: '/task',
    components: {
      default: () => import('@/components/Task.vue'),
      navbar: AppNavbar,
    },
    meta: { title: 'Junk Filter - 分发任务' }
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
```

---

## 第六阶段：样式迁移

### 6.1 Tailwind配置

```javascript
// tailwind.config.js
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: '#18181b',
        'background-light': '#f8f9fa',
        'background-dark': '#121212',
        'surface-light': '#ffffff',
        'surface-dark': '#1e1e1e',
        'sidebar-light': '#f8f9fa',
        'sidebar-dark': '#1F2937',
        'chat-bg-dark': '#111827',
        'active-light': '#e5e7eb',
        'active-dark': '#4B5563',
        'accent-dark': '#111827',
      },
      fontFamily: {
        display: ['Inter', 'sans-serif'],
        sans: ['Inter', 'sans-serif'],
      },
      borderRadius: {
        DEFAULT: '0.5rem',
      },
      boxShadow: {
        soft: '0 1px 3px 0 rgb(0 0 0 / 0.05), 0 1px 2px -1px rgb(0 0 0 / 0.05)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
```

### 6.2 全局样式

```css
/* src/styles/globals.css */
@import 'tailwindcss/base';
@import 'tailwindcss/components';
@import 'tailwindcss/utilities';

@layer base {
  body {
    @apply font-sans transition-colors duration-200;
  }

  html {
    @apply scroll-smooth;
  }
}

@layer components {
  .btn-primary {
    @apply px-5 py-2.5 bg-gray-900 hover:bg-gray-800 dark:bg-gray-700 dark:hover:bg-gray-600 text-white rounded-full text-sm font-medium transition-colors shadow-sm;
  }

  .btn-secondary {
    @apply px-5 py-2.5 bg-white hover:bg-gray-50 border border-gray-200 dark:bg-gray-800/50 dark:border-gray-700 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-300 rounded-full text-sm font-medium transition-all hover:shadow-sm;
  }
}
```

---

## 第七阶段：逻辑整合

### 7.1 useSearch.js（从main-page.js迁移）

```javascript
// src/composables/useSearch.js
import { ref } from 'vue'

export function useSearch() {
  const selectedPlatform = ref('Blog')
  const keyword = ref('')
  const isDropdownOpen = ref(false)
  const platforms = [
    { name: 'Blog', icon: 'rss_feed' },
    { name: 'Twitter', icon: 'language' },
    { name: 'Medium', icon: 'article' },
    { name: 'Email', icon: 'mail' },
    { name: 'YouTube', icon: 'play_circle' },
  ]

  const toggleDropdown = () => {
    isDropdownOpen.value = !isDropdownOpen.value
  }

  const selectPlatform = (platform) => {
    selectedPlatform.value = platform
    isDropdownOpen.value = false
  }

  const handleSearch = () => {
    if (keyword.value.trim()) {
      console.log(`搜索: ${keyword.value} (平台: ${selectedPlatform.value})`)
      keyword.value = ''
    }
  }

  return {
    selectedPlatform,
    keyword,
    isDropdownOpen,
    platforms,
    toggleDropdown,
    selectPlatform,
    handleSearch,
  }
}
```

### 7.2 useDetailDrawer.js（从timeline-page.js迁移）

```javascript
// src/composables/useDetailDrawer.js
import { ref } from 'vue'

export function useDetailDrawer() {
  const isOpen = ref(false)
  const selectedCard = ref(null)

  const openDrawer = (card) => {
    selectedCard.value = card
    isOpen.value = true
  }

  const closeDrawer = () => {
    isOpen.value = false
    selectedCard.value = null
  }

  return {
    isOpen,
    selectedCard,
    openDrawer,
    closeDrawer,
  }
}
```

### 7.3 useToast.js（通用提示）

```javascript
// src/composables/useToast.js
import { ref } from 'vue'

const toasts = ref([])

export function useToast() {
  const show = (message, type = 'success', duration = 3000) => {
    const id = Date.now()
    const toast = { id, message, type }

    toasts.value.push(toast)

    if (duration > 0) {
      setTimeout(() => {
        toasts.value = toasts.value.filter(t => t.id !== id)
      }, duration)
    }

    return id
  }

  const dismiss = (id) => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return {
    toasts: readonly(toasts),
    show,
    dismiss,
  }
}
```

---

## 第八阶段：App.vue 和 main.js

```vue
<!-- src/App.vue -->
<template>
  <div id="app" class="min-h-screen flex flex-col">
    <!-- 导航条 -->
    <AppNavbar />

    <!-- 页面内容 -->
    <main class="flex-1">
      <RouterView />
    </main>

    <!-- Toast容器 -->
    <div class="fixed top-4 right-4 z-50 space-y-2">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast', toast.type]"
      >
        {{ toast.message }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { useToast } from '@/composables/useToast'
import AppNavbar from '@/components/AppNavbar.vue'

const { toasts } = useToast()
</script>

<style scoped>
.toast {
  @apply px-4 py-3 rounded-lg text-white text-sm font-medium animate-[slideIn_0.3s_ease-out];
}

.toast.success {
  @apply bg-green-500;
}

.toast.error {
  @apply bg-red-500;
}

.toast.info {
  @apply bg-blue-500;
}
</style>
```

```javascript
// src/main.js
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/globals.css'

// 字体加载
import 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap'
import 'https://fonts.googleapis.com/icon?family=Material+Icons+Outlined'
import 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght,FILL@100..700,0..1&display=swap'

const app = createApp(App)

// 创建Pinia实例
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.mount('#app')
```

---

## 检查清单

### ✅ 工程化要求

- [x] **环境变量安全**
  - [ ] .env 文件创建（开发环境）
  - [ ] .env.example 文件创建（模板）
  - [ ] .env.production 文件创建（生产环境）
  - [ ] API Key从环境变量读取，不硬编码
  - [ ] .gitignore包含.env

- [x] **Pinia全局状态管理**
  - [ ] useThemeStore 创建
  - [ ] useConfigStore 创建（含API Key、Temperature等）
  - [ ] useTaskStore 创建（含消息历史）
  - [ ] useTimelineStore 创建（含卡片状态）
  - [ ] Store间数据同步正常

- [x] **组件提取**
  - [ ] AppNavbar.vue 创建（导航条统一）
  - [ ] Home.vue 创建（主页）
  - [ ] Config.vue 创建（配置中心，集成useConfigStore）
  - [ ] Timeline.vue 创建（时间轴，集成useTimelineStore）
  - [ ] Task.vue 创建（分发任务，集成useTaskStore）

- [x] **路由配置**
  - [ ] vue-router 配置完成
  - [ ] 路由映射：/ /timeline /config /task
  - [ ] 当前路由自动高亮

- [x] **样式保留**
  - [ ] Tailwind配置迁移（所有颜色保留）
  - [ ] 深色模式保留（dark: 类保留）
  - [ ] 响应式设计保留
  - [ ] 所有CSS动画保留

- [x] **逻辑整合**
  - [ ] 所有交互逻辑迁移到composables
  - [ ] onMounted钩子正确使用
  - [ ] 跨页面数据通过Pinia共享

### 🎯 验收标准

1. **功能完整性**
   - [ ] 所有4个页面都能正常访问
   - [ ] 导航链接跳转正常
   - [ ] 主题切换生效

2. **数据一致性**
   - [ ] 配置中心修改API Key → 任务页面能使用同一个Key
   - [ ] 配置中心的Temperature → 任务页面能读取
   - [ ] localStorage持久化正常

3. **性能指标**
   - [ ] HMR(热更新)正常工作
   - [ ] 页面切换无闪烁
   - [ ] 动画平滑流畅
   - [ ] 构建时间<5s

4. **安全性**
   - [ ] API Key不在代码中硬编码
   - [ ] .env不提交到Git
   - [ ] .env.example展示正确的结构

5. **开发体验**
   - [ ] Vite启动快速(<1s)
   - [ ] 修改代码自动刷新
   - [ ] 控制台无报错
   - [ ] 开发和生产环境变量正确切换

---

## 📊 技术栈对比

| 指标 | 原生HTML | Vue 3 + Pinia |
|-----|---------|--------------|
| 文件数 | 4个独立HTML | 1个入口 + 5个组件 |
| 状态管理 | localStorage手动同步 | ✅ Pinia自动同步 |
| API Key管理 | 硬编码或localStorage | ✅ 环境变量安全 |
| 页面切换 | 全页刷新 | ✅ SPA无缝切换 |
| HMR支持 | ❌ 无 | ✅ 完整支持 |
| 构建优化 | 无 | ✅ 代码分割+压缩 |
| 开发工具 | 无 | ✅ Vue DevTools |
| 部署 | 4个文件 | ✅ 单个dist目录 |

---

## 🚀 后续部署指南

### 开发环境
```bash
npm run dev
# Vite会自动:
# 1. 加载.env文件
# 2. 初始化Pinia
# 3. 启用HMR
```

### 生产构建
```bash
npm run build
# 构建结果:
# - dist/index.html
# - dist/assets/main.xxxxx.js
# - dist/assets/style.xxxxx.css
# 构建时使用.env.production中的变量
```

### CI/CD集成
```yaml
# .github/workflows/deploy.yml示例
env:
  VITE_API_KEY: ${{ secrets.PROD_API_KEY }}
  VITE_API_URL: ${{ secrets.PROD_API_URL }}
```

---

## 📝 总结

这个迁移计划的核心价值：

1. **安全性** ✅
   - API Key从环境变量读取，不硬编码
   - .env文件Git忽略，不泄露敏感信息

2. **可维护性** ✅
   - Pinia集中管理状态，数据流清晰
   - Composition API逻辑复用，代码整洁
   - 单一入口，部署简单

3. **开发体验** ✅
   - HMR热更新，修改即刻反馈
   - Vue DevTools调试便捷
   - TypeScript支持（可选）

4. **性能** ✅
   - Vite极速构建
   - 代码自动分割
   - Tree shaking移除未用代码

5. **团队协作** ✅
   - 标准化项目结构
   - 环境变量示例清晰
   - 状态管理规范统一

**现在可以开始编码了！** 🎉
