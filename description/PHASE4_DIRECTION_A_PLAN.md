# 🚀 Phase 4 方向 A - 快速功能迭代执行计划

**状态**: 冒烟测试完全通过 ✅
**当前阶段**: Phase 4 - 快速功能迭代
**目标**: 2-3 天内完成所有新功能开发，后期统一切换到真实后端

---

## 📅 时间表

```
第 1 天 (Day 1)：Config.vue 完整实现
  ├─ 上午：RSS 源管理表格 (2h)
  └─ 下午：AI 模型配置面板 (3h)

第 2 天 (Day 2)：消息功能扩展 + 任务管理
  ├─ 上午：消息搜索/筛选/导出 (3h)
  └─ 下午：任务执行管理 (2h)

第 3 天 (Day 3)：UX 优化 + 测试
  ├─ 上午：加载状态、空状态、错误提示 (2h)
  └─ 下午：测试和 Bug 修复 (2h)

第 4 天：切换真实后端
  └─ 1-2 小时配置切换
```

---

## 🎯 Task 1: Config.vue 完整实现

### 目标
实现完整的配置中心，包括 RSS 源管理和 AI 模型配置

### 当前状态
✅ Config.vue 已有基础框架
✅ 表格布局已完成
✅ Modal 占位符已有

### 需要完成的工作

#### 1.1 Mock 后端扩展（15 分钟）

**文件**: `D:\TrueSignal\backend-mock\server.js`

需要添加的 Mock API 端点：

```javascript
// 1. RSS 源管理 API（如果还没有）
GET /api/sources           // 获取所有源
POST /api/sources          // 创建源
PUT /api/sources/:id       // 更新源
DELETE /api/sources/:id    // 删除源

// 2. 同步日志 API
GET /api/sources/:id/logs  // 获取源的同步日志

// 3. AI 配置 API
GET /api/config/ai         // 获取 AI 配置
POST /api/config/ai        // 保存 AI 配置
```

Mock 数据示例：
```javascript
{
  id: 1,
  name: "Ars Technica",
  url: "https://feeds.arstechnica.com/arstechnica/index",
  frequency: "daily",
  status: "active",
  lastSyncTime: "2026-02-27T20:00:00Z",
  lastSyncStatus: "success",
  syncLogs: [
    {
      timestamp: "2026-02-27T20:00:00Z",
      status: "success",
      itemsCount: 15,
      message: "成功获取 15 条新文章"
    }
  ]
}
```

#### 1.2 useConfigStore 完整实现（45 分钟）

**文件**: `D:\TrueSignal\frontend-vue\src\stores\useConfigStore.js`

需要完成的功能：

```javascript
// 1. RSS 源管理状态
const sources = ref([...初始数据])           // RSS 源列表
const showAddRssModal = ref(false)            // 添加 RSS Modal
const newRssForm = ref({...})                 // RSS 表单数据
const expandedSourceIds = ref([])             // 展开的行

// 2. AI 模型管理状态
const aiModels = ref([])                      // AI 模型列表
const showAddModelModal = ref(false)          // 添加模型 Modal
const newModelForm = ref({...})               // 模型表单数据

// 3. 配置状态
const temperature = ref(0.7)                  // 温度参数
const topP = ref(0.9)                         // Top P 参数
const maxTokens = ref(2000)                   // 最大 token

// 4. 方法
const toggleRssModal = () => {}               // 切换 RSS Modal
const addSource = async (data) => {}          // 添加源
const deleteSource = async (id) => {}         // 删除源
const syncSource = async (id) => {}           // 同步源（模拟）
const toggleSourceExpanded = (id) => {}       // 展开/收起行

const toggleModelModal = () => {}             // 切换模型 Modal
const addModel = async (data) => {}           // 添加模型
const deleteModel = async (id) => {}          // 删除模型

const saveConfig = async () => {}             // 保存所有配置
```

**实现清单**:
- [ ] 初始化状态
- [ ] 实现方法（CRUD 操作）
- [ ] 添加验证逻辑
- [ ] 集成 useToast（成功/失败提示）
- [ ] 本地存储持久化

#### 1.3 Config.vue 组件完成（2 小时）

**需要完成的部分**:

```vue
<!-- 1. RSS 源表格部分 - 已有框架，需要：-->
- [ ] 连接 configStore.sources 数据
- [ ] 实现手动同步功能（handleSyncSource）
- [ ] 实现删除功能（handleDeleteSource）
- [ ] 实现展开行动画
- [ ] 添加空状态提示

<!-- 2. 添加 RSS 源 Modal - 需要实现：-->
- [ ] 表单验证（名称、URL）
- [ ] URL 格式检查
- [ ] 提交按钮（创建）
- [ ] 取消按钮
- [ ] 加载态显示

<!-- 3. AI 模型配置区 - 需要实现：-->
- [ ] 模型选择下拉菜单
- [ ] API 密钥输入
- [ ] Temperature 滑块 + 数值显示
- [ ] Top P 滑块 + 数值显示
- [ ] Max Tokens 输入框
- [ ] 参数说明文本

<!-- 4. 添加 AI 模型 Modal - 需要实现：-->
- [ ] 模型名称输入
- [ ] 服务商选择（OpenAI, DeepSeek, etc）
- [ ] API 密钥输入（password type）
- [ ] Base URL 输入（可选）
- [ ] 表单验证

<!-- 5. 保存按钮 - 需要：-->
- [ ] 保存所有配置
- [ ] 显示成功/失败提示
- [ ] 加载态
```

### 验收标准

- [ ] 能显示 3-4 个 RSS 源
- [ ] 能添加新的 RSS 源（Modal 工作正常）
- [ ] 能删除 RSS 源（有确认提示）
- [ ] 能手动同步源（显示加载动画）
- [ ] 能展开/收起同步日志
- [ ] 能配置 AI 模型（温度、Top P 等）
- [ ] 能保存配置（Toast 提示成功）
- [ ] 暗黑模式工作正常
- [ ] 所有交互流畅无卡顿

---

## 🎯 Task 2: 消息功能扩展

### 目标
增强消息管理功能，支持搜索、筛选、导出

### 需要完成的工作

#### 2.1 Mock 后端 API 扩展（20 分钟）

```javascript
// 新增端点
GET /api/messages/search?q=keyword      // 搜索消息
GET /api/messages/export?format=csv     // 导出消息
DELETE /api/messages/:id                // 删除单条消息
PUT /api/messages/:id                   // 编辑消息（可选）
```

#### 2.2 消息搜索功能（1 小时）

**文件**: `D:\TrueSignal\frontend-vue\src\components\TaskChat.vue`

新增功能：
```vue
<!-- 搜索栏 -->
<input
  v-model="searchQuery"
  placeholder="搜索消息..."
  @input="handleSearchMessages"
/>

<!-- 搜索结果高亮 -->
<div v-for="message in filteredMessages">
  <!-- 高亮显示匹配的关键词 -->
</div>
```

实现清单：
- [ ] 搜索输入框 UI
- [ ] 防抖搜索（避免频繁请求）
- [ ] 结果高亮显示
- [ ] 清空搜索功能
- [ ] 搜索结果计数

#### 2.3 消息筛选功能（45 分钟）

```vue
<!-- 筛选选项 -->
<select v-model="filterRole">
  <option value="">全部</option>
  <option value="user">用户消息</option>
  <option value="ai">AI 回复</option>
</select>

<select v-model="filterDate">
  <option value="">任意时间</option>
  <option value="today">今天</option>
  <option value="week">本周</option>
  <option value="month">本月</option>
</select>
```

实现清单：
- [ ] 按消息类型筛选（用户/AI）
- [ ] 按日期筛选（今天/本周/本月）
- [ ] 按状态筛选（已读/未读）
- [ ] 多条件组合筛选
- [ ] 筛选结果计数

#### 2.4 消息导出功能（45 分钟）

```vue
<!-- 导出按钮 -->
<button @click="exportMessages">
  <span class="material-icons">download</span>
  导出消息
</button>

<!-- 导出格式选择 Modal -->
<select v-model="exportFormat">
  <option value="csv">CSV 文件</option>
  <option value="json">JSON 文件</option>
  <option value="pdf">PDF 文件</option>
</select>
```

实现清单：
- [ ] 导出为 CSV
- [ ] 导出为 JSON
- [ ] 导出为 PDF（可选）
- [ ] 导出进度显示
- [ ] 下载提示

### 验收标准

- [ ] 能搜索消息内容
- [ ] 搜索结果实时高亮
- [ ] 能按类型筛选（用户/AI）
- [ ] 能按日期筛选
- [ ] 能导出为 CSV
- [ ] 导出文件包含所有字段
- [ ] 所有操作有加载态和反馈

---

## 🎯 Task 3: 任务执行管理

### 目标
实现手动执行任务、查看历史记录功能

### 需要完成的工作

#### 3.1 Mock 后端 API（15 分钟）

```javascript
// 新增端点
POST /api/sources/:id/execute        // 手动执行源
GET /api/sources/:id/execution-history  // 获取执行历史
```

#### 3.2 任务执行功能（1 小时）

**文件**: `D:\TrueSignal\frontend-vue\src/components/TaskSidebar.vue`

新增功能：
```vue
<!-- 右键菜单或按钮 -->
<button @click="executeTask(task)">
  <span class="material-icons">play_arrow</span>
  执行任务
</button>

<!-- 执行进度 Modal -->
<div v-if="isExecuting" class="modal">
  <div class="progress">
    <div class="progress-bar" :style="{ width: progress + '%' }"></div>
  </div>
  <p>正在执行... {{ progress }}%</p>
</div>
```

实现清单：
- [ ] 执行按钮 UI
- [ ] 执行进度显示
- [ ] 执行成功/失败提示
- [ ] 实时日志显示（可选）
- [ ] 取消执行功能（可选）

#### 3.2 执行历史功能（1 小时）

**新增组件**: `D:\TrueSignal\frontend-vue\src/components/ExecutionHistory.vue`

显示信息：
```
时间 | 状态 | 耗时 | 获取数量 | 详情

2026-02-27 20:00:00 | ✅ 成功 | 2.3s | 15 条 | 查看日志
2026-02-27 19:00:00 | ❌ 失败 | 5s   | 0 条  | 查看错误
```

实现清单：
- [ ] 历史记录表格
- [ ] 状态指示符（成功/失败）
- [ ] 执行时间显示
- [ ] 详情查看功能
- [ ] 分页支持

### 验收标准

- [ ] 能点击按钮手动执行任务
- [ ] 显示执行进度
- [ ] 显示执行结果
- [ ] 能查看执行历史
- [ ] 历史记录显示准确信息
- [ ] 所有操作有反馈提示

---

## 🎯 Task 4: UX 优化

### 目标
改进用户体验，添加加载状态、空状态、错误提示

### 4.1 加载状态（1 小时）

实现全局加载指示：
```vue
<!-- Skeleton 加载 -->
<div v-if="isLoading" class="skeleton-loader">
  <div class="skeleton-item"></div>
  <div class="skeleton-item"></div>
</div>

<!-- 加载动画 -->
<div v-show="isLoading" class="loading-spinner">
  <div class="spinner"></div>
  <p>加载中...</p>
</div>
```

实现清单：
- [ ] 任务列表加载态
- [ ] 消息列表加载态
- [ ] 配置加载态
- [ ] API 请求加载态
- [ ] Skeleton 动画

### 4.2 空状态提示（45 分钟）

```vue
<!-- 任务列表为空 -->
<div v-if="tasks.length === 0" class="empty-state">
  <span class="material-icons">inbox</span>
  <p>还没有任务</p>
  <button @click="openNewTaskModal">创建第一个任务</button>
</div>

<!-- 消息列表为空 -->
<div v-if="messages.length === 0" class="empty-state">
  <span class="material-icons">chat_bubble_outline</span>
  <p>选择任务开始聊天</p>
</div>
```

实现清单：
- [ ] 任务列表空状态
- [ ] 消息列表空状态
- [ ] 搜索无结果提示
- [ ] 筛选无结果提示

### 4.3 错误处理提升（1 小时）

```vue
<!-- 错误卡片 -->
<div v-if="error" class="error-card">
  <span class="material-icons">error_outline</span>
  <p>{{ error.message }}</p>
  <button @click="retry">重试</button>
</div>

<!-- 网络错误提示 -->
<div v-if="isOffline" class="offline-banner">
  <p>网络已断开，请检查连接</p>
</div>
```

实现清单：
- [ ] 网络错误提示
- [ ] API 错误信息显示
- [ ] 错误重试按钮
- [ ] 离线模式提示
- [ ] 超时提示

### 4.4 动画和过渡（45 分钟）

```css
/* Modal 打开/关闭动画 */
.modal-enter-active { animation: modalIn 0.3s ease-out; }
.modal-leave-active { animation: modalOut 0.2s ease-in; }

/* 列表项淡入 */
.list-enter-active { animation: fadeInUp 0.3s; }

/* 按钮交互 */
button:active { transform: scale(0.95); }
```

实现清单：
- [ ] Modal 淡入淡出
- [ ] 列表项淡入动画
- [ ] 按钮按下动画
- [ ] Skeleton 加载动画
- [ ] 所有过渡 300ms 以内

### 验收标准

- [ ] 所有操作都有加载状态
- [ ] 空状态显示提示和操作
- [ ] 错误显示清晰且可重试
- [ ] 动画流畅（60fps）
- [ ] 暗黑模式对比度足够

---

## 🔄 Task 5: 测试和 Bug 修复（第 3 天）

### 5.1 功能测试

```
[ ] Config.vue 所有功能
  [ ] RSS 源 CRUD
  [ ] AI 配置保存/加载
  [ ] 同步日志显示

[ ] 消息功能
  [ ] 搜索功能
  [ ] 筛选功能
  [ ] 导出功能

[ ] 任务执行
  [ ] 手动执行
  [ ] 历史记录

[ ] 暗黑模式
  [ ] 所有组件在暗黑模式下显示正常

[ ] 响应式设计
  [ ] 桌面版
  [ ] 平板版
  [ ] 移动版
```

### 5.2 边界情况

```
[ ] 特殊字符处理（emoji, 中文, 特殊符号）
[ ] 长文本处理（长消息、长 URL）
[ ] 快速连续操作
[ ] 网络延迟（关闭 Mock 服务器后恢复）
[ ] 浏览器刷新后数据持久化
```

### 5.3 性能优化

```
[ ] 消息列表虚拟滚动（如果超过 100 条）
[ ] 搜索防抖（300ms）
[ ] 图片懒加载（如果有图片）
[ ] 减少不必要的重渲染
```

---

## 📊 开发进度追踪

### Task 1: Config.vue 完整实现
- [ ] Mock API 扩展（15 分钟）✅
- [ ] useConfigStore 实现（45 分钟）⏳
- [ ] Config.vue 组件完成（2 小时）⏳

**总耗时**: 3 小时

### Task 2: 消息功能扩展
- [ ] Mock API 扩展（20 分钟）⏳
- [ ] 搜索功能（1 小时）⏳
- [ ] 筛选功能（45 分钟）⏳
- [ ] 导出功能（45 分钟）⏳

**总耗时**: 2.75 小时

### Task 3: 任务执行管理
- [ ] Mock API（15 分钟）⏳
- [ ] 执行功能（1 小时）⏳
- [ ] 历史记录（1 小时）⏳

**总耗时**: 2.25 小时

### Task 4: UX 优化
- [ ] 加载状态（1 小时）⏳
- [ ] 空状态（45 分钟）⏳
- [ ] 错误处理（1 小时）⏳
- [ ] 动画和过渡（45 分钟）⏳

**总耗时**: 3.75 小时

### Task 5: 测试和修复
- [ ] 功能测试（1.5 小时）⏳
- [ ] 边界情况（1 小时）⏳
- [ ] 性能优化（1.5 小时）⏳

**总耗时**: 4 小时

---

## 💾 Mock 数据完整示例

### sources.json
```json
[
  {
    "id": 1,
    "name": "Ars Technica",
    "url": "https://feeds.arstechnica.com/arstechnica/index",
    "frequency": "daily",
    "status": "active",
    "lastSyncTime": "2026-02-27T20:00:00Z",
    "lastSyncStatus": "success",
    "syncLogs": [
      {
        "timestamp": "2026-02-27T20:00:00Z",
        "status": "success",
        "itemsCount": 15,
        "message": "成功获取 15 条新文章"
      }
    ]
  }
]
```

### executionHistory.json
```json
[
  {
    "id": 1,
    "sourceId": 1,
    "timestamp": "2026-02-27T20:00:00Z",
    "status": "success",
    "duration": 2.3,
    "itemsCount": 15,
    "message": "成功执行"
  }
]
```

---

## 🎯 最后一步：切换真实后端（第 4 天）

完成所有前端功能后，切换到真实后端只需：

```javascript
// .env.local
VITE_API_URL=http://localhost:8080  // 改为你的真实后端地址

// 无需修改任何前端代码！
```

---

## ✅ 检查清单

在开始每个 Task 前，检查：

- [ ] 所有依赖已安装
- [ ] 开发服务器正在运行（npm run dev）
- [ ] Mock 后端正在运行（node server.js）
- [ ] 浏览器开发工具已打开
- [ ] 代码编辑器已准备好

---

## 📝 提交规范

每个 Task 完成后：

```bash
# 添加更改
git add .

# 提交（示例）
git commit -m "feat: 完成 Config.vue 完整实现

- 实现 RSS 源管理表格
- 实现 AI 模型配置面板
- 添加模态框和表单验证
- 集成 useConfigStore"
```

---

**准备开始了吗？让我们开始 Task 1！** 🚀
