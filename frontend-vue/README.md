# 🚀 Junk Filter Vue 3 迁移项目

**基于 Vue 3 + Vite + Pinia 的现代化前端工程**

项目地址：`frontend-vue/`

---

## 📦 项目结构

```
frontend-vue/
├── .env                    # 🔐 环境变量（开发，Git忽略）
├── .env.example            # 环境变量模板
├── .env.production         # 生产环境配置
├── index.html              # 单一入口
├── vite.config.js          # Vite配置
├── tailwind.config.js      # Tailwind配置
├── postcss.config.js       # PostCSS配置
├── package.json            # 依赖管理
└── src/
    ├── main.js             # 应用入口
    ├── App.vue             # 根组件
    ├── components/         # Vue组件
    │   ├── AppNavbar.vue   # 导航条
    │   ├── Home.vue        # 主页
    │   ├── Config.vue      # 配置中心
    │   ├── Timeline.vue    # 时间轴
    │   └── Task.vue        # 分发任务
    ├── stores/             # Pinia状态管理
    │   ├── useThemeStore.js
    │   ├── useConfigStore.js
    │   ├── useTaskStore.js
    │   ├── useTimelineStore.js
    │   └── index.js
    ├── composables/        # Composition API
    │   ├── useSearch.js
    │   ├── useDetailDrawer.js
    │   └── useToast.js
    ├── router/             # 路由
    │   └── index.js
    └── styles/             # 样式
        └── globals.css
```

---

## 🎯 核心特性

### ✅ 环境变量安全管理
- API Key 从环境变量读取，**绝不硬编码**
- 开发、生产环境配置分离
- .env 文件自动 Git 忽略

**示例：**
```bash
# .env（开发）
VITE_API_KEY=sk-proj-dev-key
VITE_API_URL=http://localhost:8000/api

# .env.production（生产）
VITE_API_KEY=${VITE_API_KEY}  # 通过 CI/CD 注入
VITE_API_URL=https://api.prod.example.com
```

在组件中使用：
```javascript
const apiKey = import.meta.env.VITE_API_KEY
```

### ✅ Pinia 全局状态管理

**4个核心Store：**

1. **useThemeStore** - 主题管理
   - 亮色/深色模式切换
   - localStorage 持久化

2. **useConfigStore** - 配置中心（★ 核心）
   - API Key、Model、Temperature、maxTokens
   - 保存配置（含 20% 失败模拟）
   - 跨页面数据同步

3. **useTaskStore** - 任务管理
   - 消息历史
   - AI 响应生成（含 30% 失败模拟）
   - 任务切换

4. **useTimelineStore** - 时间轴
   - 卡片数据
   - 过滤器管理
   - 侧滑抽屉状态

**跨页面数据同步示例：**
```javascript
// Config页面修改 API Key
configStore.updateApiKey('new-key')

// Task页面自动读取
const taskStore = useTaskStore()
const configStore = useConfigStore()
await taskStore.sendMessage(text)  // 自动使用最新的 API Key
```

### ✅ 路由管理 (vue-router)

```javascript
/ → Home（主页）
/timeline → Timeline（时间轴）
/config → Config（配置中心）
/task → Task（分发任务）
```

导航自动高亮，页面标题自动更新。

### ✅ 样式保留

- ✓ Tailwind CSS 所有原生类保留
- ✓ 自定义颜色配置保留（#111827、#f8a8e8 等）
- ✓ 深色模式完全兼容
- ✓ 响应式设计保留
- ✓ 所有动画保留

### ✅ 交互功能完整

- ✓ 搜索框聚焦效果
- ✓ 平台选择下拉菜单
- ✓ 快捷标签悬停
- ✓ 配置中心完整功能（API Key 复制、配置导出、保存反馈）
- ✓ 时间轴卡片悬浮、侧滑抽屉、过滤切换
- ✓ 任务管理（消息发送、AI 打字机、异常处理、重试）

---

## 🚀 快速开始

### 1. 安装依赖
```bash
cd frontend-vue
npm install
```

### 2. 配置环境变量
```bash
# 复制示例文件
cp .env.example .env

# 编辑 .env，填入你的 API Key
VITE_API_KEY=your-actual-key-here
```

### 3. 启动开发服务器
```bash
npm run dev
```

浏览器自动打开 `http://localhost:5173`

HMR（热更新）自动启用，修改代码实时生效。

### 4. 生产构建
```bash
npm run build
```

输出到 `dist/` 目录。

---

## 🔐 安全性检查清单

- [x] API Key 不硬编码，从环境变量读取
- [x] .env 文件 Git 忽略（`.gitignore` 配置）
- [x] .env.example 提供结构示例
- [x] 生产环境通过 CI/CD 注入敏感信息
- [x] localStorage 只存储非敏感配置

---

## 📊 与原生HTML对比

| 指标 | 原生HTML | Vue 3 |
|-----|---------|--------|
| 文件数 | 4个独立HTML | 1个SPA应用 |
| 状态管理 | 手动localStorage同步 | ✅ Pinia自动同步 |
| API Key管理 | 无安全机制 | ✅ 环境变量安全 |
| 页面切换 | 全页刷新 | ✅ 无缝SPA |
| HMR支持 | ❌ 无 | ✅ 完整支持 |
| 构建优化 | ❌ 无 | ✅ 自动分割+压缩 |
| 开发工具 | 无 | ✅ Vue DevTools |
| 部署 | 4个文件复制 | ✅ dist目录一键部署 |

---

## 🛠️ 常见开发任务

### 添加新的Store状态
```javascript
// src/stores/useMyStore.js
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useMyStore = defineStore('mystore', () => {
  const count = ref(0)
  const increment = () => count.value++

  return { count, increment }
})
```

### 在组件中使用Store
```vue
<script setup>
import { useMyStore } from '@/stores'

const myStore = useMyStore()
</script>

<template>
  <div>
    <p>{{ myStore.count }}</p>
    <button @click="myStore.increment">+1</button>
  </div>
</template>
```

### 创建新的Composable
```javascript
// src/composables/useMyComposable.js
import { ref } from 'vue'

export function useMyComposable() {
  const data = ref(null)

  const fetchData = async () => {
    // 获取数据
  }

  return { data, fetchData }
}
```

---

## 🐛 调试

### Vue DevTools
浏览器安装 Vue DevTools 扩展，可以：
- 实时查看组件树
- 调试Store状态
- 时间旅行调试

### Console日志
```javascript
// 在任何地方打印Store状态
const configStore = useConfigStore()
console.log('Current API Key:', configStore.apiKey)
```

### 网络调试
F12 → Network 标签，查看所有请求（当接入真实API后）

---

## 📝 环境变量说明

### 开发环境 (.env)
```bash
VITE_API_KEY=sk-proj-dev-key              # 开发API Key
VITE_API_URL=http://localhost:8000/api    # 本地后端地址
VITE_APP_ENV=development                  # 应用环境
VITE_LOG_LEVEL=debug                      # 日志级别
VITE_ENABLE_MOCK_DATA=true                # 使用Mock数据
```

### 生产环境 (.env.production)
```bash
VITE_API_KEY=${VITE_API_KEY}              # 通过CI/CD注入
VITE_API_URL=https://api.prod.example.com # 生产后端
VITE_APP_ENV=production                   # 应用环境
VITE_LOG_LEVEL=error                      # 只记录错误
VITE_ENABLE_MOCK_DATA=false               # 使用真实API
```

---

## 🌐 部署指南

### Vercel (推荐)
1. 关联GitHub仓库
2. 设置环境变量（VITE_API_KEY等）
3. 自动构建和部署

### Docker
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 传统服务器
```bash
npm run build
# 将 dist 目录上传到服务器
# 配置 web 服务器指向 dist/index.html
```

---

## 📚 学习资源

- [Vue 3 文档](https://vuejs.org/)
- [Pinia 文档](https://pinia.vuejs.org/)
- [Vue Router 文档](https://router.vuejs.org/)
- [Vite 文档](https://vitejs.dev/)
- [Tailwind CSS 文档](https://tailwindcss.com/)

---

## ✅ 验收检查清单

- [x] 所有4个页面完整迁移
- [x] 导航自动高亮当前路由
- [x] 环境变量安全管理
- [x] Pinia跨页面数据同步
- [x] 所有原生Tailwind类保留
- [x] 深色模式完全兼容
- [x] 所有交互功能保留
- [x] HMR热更新正常
- [x] 构建体积优化
- [x] 无console错误

---

**项目迁移完成！** 🎉

现在你拥有一个现代化、安全、易维护的Vue 3前端项目！

