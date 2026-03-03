# ⚡ 快速参考 - 搜索结果页面

## 📂 新增文件

```
src/components/
├── SearchBar.vue          ← 新增（搜索栏子组件）
└── Home.vue               ← 重写（双态搜索 + 结果列表）
```

## 🎬 交互状态图

```
初始态 ← → 激活态
  ↓         ↓
中央搜索栏   吸顶搜索栏 + 结果列表
max-w-2xl   max-w-xl
rounded-full rounded-lg
```

## 🔑 关键代码片段

### 状态管理
```javascript
const isSearching = ref(false)  // 激活状态
const keyword = ref('')         // 搜索词（必须初始化！）
const searchResults = ref([...]) // 模拟数据
```

### 事件处理
```javascript
onSearchFocus()   // 点击输入框 → 激活
onSearchBlur()    // 失焦 + 内容为空 → 恢复初始态
onSearchClose()   // 点击关闭 → 恢复初始态
onSearch(query)   // 执行搜索 → 后端调用接口
```

## 🎨 样式关键点

| 属性 | 初始态 | 激活态 |
|-----|-------|--------|
| 宽度 | max-w-2xl | max-w-xl |
| 圆角 | rounded-full | rounded-lg |
| 阴影 | shadow-xl | shadow-sm |
| 位置 | 中央 flex items-center justify-center | sticky top-16 |
| 转换 | duration-500 | duration-500 |

## 📋 数据结构

```javascript
{
  id: 1,              // 唯一标识
  name: 'Design Digest Daily',
  username: '@designdigest',
  followers: '245k',
  avatar: 'https://...',
  status: 'Highly Active',      // 可为 null
  statusColor: 'green',          // green|red|sky|black|gray
  icon: 'public',                // Material Icon name
  isSubscribed: false,
}
```

## 🚀 快速上线

### 1️⃣ 测试现有功能
```bash
npm run dev
# 访问 http://localhost:5173
```

### 2️⃣ 连接后端 API
在 Home.vue 中修改 `onSearch()` 函数：
```javascript
const onSearch = async (query) => {
  const results = await fetch(`/api/search?q=${query}`)
  searchResults.value = await results.json()
}
```

### 3️⃣ 处理导航栏高度
如果 AppNavbar 不是 64px 高，在 Home.vue 第 51 行修改：
```vue
<!-- 原本 -->
<div class="sticky top-16 z-40 ...">

<!-- 如果导航栏是 80px，改为 top-20 -->
<div class="sticky top-20 z-40 ...">
```

## ✅ 验证清单

运行此检查清单确保一切正常：

```bash
□ 页面初始加载显示大搜索栏（中央）
□ 点击搜索框，搜索栏上移到顶部
□ 结果列表显示 5 个卡片
□ 卡片显示头像、名称、粉丝、订阅按钮
□ 点击订阅按钮，按钮文字翻转
□ 清空搜索框后失焦，回到初始态
□ 点击搜索框左侧关闭按钮，回到初始态
□ 深色模式正常显示
□ 控制台无错误信息
```

## 🔧 常见修改

### 改动画速度
```vue
<!-- 快速（300ms） -->
class="transition-all duration-300"

<!-- 慢速（700ms） -->
class="transition-all duration-700"
```

### 改颜色
```vue
<!-- 搜索按钮从深灰改为蓝色 -->
class="bg-blue-600 hover:bg-blue-700"
```

### 改圆角大小
```vue
<!-- 从 rounded-full 改为 rounded-2xl -->
class="rounded-2xl"

<!-- 从 rounded-lg 改为 rounded-md -->
class="rounded-md"
```

## 🎯 模拟数据替换

### 现在（本地模拟）
```javascript
const searchResults = ref([...5个硬编码卡片...])
```

### 后续（API 数据）
```javascript
import { fetchSearchResults } from '@/services/searchService'

const onSearch = async (query) => {
  searchResults.value = await fetchSearchResults(query)
}
```

## 📞 问题快速排查

| 问题 | 原因 | 解决 |
|-----|------|------|
| 搜索栏与导航栏重叠 | top 值不对 | 调整 top-16 |
| 搜索框无法输入 | disabled 属性问题 | 检查 keyword 初始化 |
| 结果不显示 | v-if 条件错误 | 检查 isSearching 状态 |
| 动画卡顿 | 性能问题 | 用 v-if 代替 v-show |

---

**快速导航**：
- 📖 [完整实现文档](./SEARCH_RESULTS_IMPLEMENTATION.md)
- 🎨 [设计稿对应代码](./src/components/Home.vue)
- 🔧 [SearchBar 子组件](./src/components/SearchBar.vue)

