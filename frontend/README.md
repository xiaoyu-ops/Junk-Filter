## 前端 JavaScript 交互系统 - 使用指南

### 📁 文件结构

```
frontend/
├── main.html              # 主页 - 搜索中心
├── config.html            # 配置中心 - RSS 源管理、模型配置
├── timeline.html          # 时间轴 - 内容流展示
├── task.html              # 分发任务 - AI 对话助手
└── js/
    ├── common.js          # 通用工具库（主题、Toast、复制、防抖等）
    ├── main-page.js       # 主页交互
    ├── config-page.js     # 配置中心交互
    ├── timeline-page.js   # 时间轴交互
    └── task-page.js       # 任务管理交互
```

---

## ✨ 功能清单

### 1️⃣ 主页 (main.html)

#### 搜索框聚焦效果
- **效果**：点击搜索框时，边框平滑过渡为深灰色，投影增强，营造"浮起"感
- **过渡时间**：300ms
- **技术**：CSS transition + box-shadow

#### 平台选择下拉菜单
- **功能**：点击"Select Platform"按钮展开下拉菜单
- **选项**：Blog、Twitter、Medium、Email、YouTube
- **动画**：向下滑入 + 淡入（300ms）
- **交互**：选择后更新按钮文本，菜单自动关闭

#### 快捷标签悬停
- **效果**：鼠标悬停时，标签上移 2px，背景加深，投影增强
- **过渡时间**：300ms
- **点击反馈**：显示 Info 类型 Toast

---

### 2️⃣ 时间轴 (timeline.html)

#### 卡片悬浮反馈
- **效果**：卡片缩放 1.02 倍，投影从 shadow-sm 升级为 shadow-lg
- **过渡时间**：200ms
- **缓动**：cubic-bezier(0.4, 0, 0.2, 1)

#### 侧滑抽屉（右侧详情面板）
- **触发**：点击卡片
- **动画**：从右侧屏幕外平滑滑入（400ms）
- **功能**：显示作者信息、文章内容、评分等
- **关闭**：点击关闭按钮或背景遮罩
- **遮罩**：半透明背景（rgba(0,0,0,0.3)）

#### 过滤切换
- **按钮**：All、Worth Watching、Latest
- **效果**：激活状态切换，卡片淡隐淡现（200ms + 200ms）
- **动画类型**：Fade（淡入淡出）

#### 主题切换
- **按钮**：右上角主题按钮
- **功能**：切换亮色/深色模式
- **持久化**：保存到 localStorage

---

### 3️⃣ 配置中心 (config.html)

#### Temperature 滑块实时联动
- **效果**：拖动时，右侧数值实时变化
- **进度条**：滑块背景显示进度渐变
- **视觉反馈**：拖动时出现光晕效果

#### 表格行操作
- **悬停**：行背景变浅色
- **删除**：
  - 点击删除按钮 → 确认对话框
  - 确认后行向左滑出（300ms）
  - 下方行自动上移

#### 保存配置
- **流程**：
  1. 点击"保存配置"按钮
  2. 按钮显示"保存中..." + 加载图标（禁用）
  3. 1 秒后模拟 API 请求
  4. **成功**：按钮变绿 + ✓ 图标 + Green Toast（3 秒自动消失）
  5. **失败**：按钮变红 + ✗ 图标 + Red Toast + 保留"重试"按钮
- **成功率**：80%（20% 模拟失败）

#### 一键复制功能

**API Key 复制：**
- **按钮**：在密钥框右侧添加复制图标
- **功能**：点击复制当前 API Key 到剪贴板
- **反馈**：
  - 按钮图标变为绿色 ✓
  - 显示 Green Toast："已复制到剪贴板"
  - 1.5 秒后恢复原样

**可见性切换：**
- **按钮**：密钥框右侧眼睛图标
- **功能**：切换密钥显示/隐藏

**配置导出：**
- **按钮**："导出配置"（蓝色按钮，在保存按钮左侧）
- **功能**：将当前配置（模型、温度、Token）导出为 JSON 并复制
- **反馈**：同 API Key 复制反馈

---

### 4️⃣ 分发任务 (task.html)

#### 任务列表切换
- **左侧列表**：显示已有任务
- **效果**：点击任务 → 激活状态切换（左边框高度过渡 250ms）
- **对话更新**：右侧对话框内容淡出淡入（200ms + 200ms）

#### 智能输入与发送

**Shift + Enter 换行：**
- 按下 Shift + Enter → 在输入框中插入换行符
- 输入框高度自动调整（最大 150px）
- 支持多行输入

**单独 Enter 发送：**
- 按下 Enter（无 Shift）→ 立即发送消息
- 清空输入框，恢复高度

**输入框高度自适应：**
- 初始：py-4（固定）
- 输入多行时：自动扩展（max-height: 150px）
- 超过 max-height 时：显示滚动条

#### 用户消息气泡
- **样式**：左侧头像 + 灰色气泡
- **动画**：从下往上淡入 + 上移（300ms）
- **内容**：支持多行显示（换行符保留）

#### AI 打字机效果
- **流程**：
  1. 用户发送消息
  2. AI 消息气泡出现，显示"typing dots"（3 个小球）
  3. 500-800ms 后判断成功/失败
  4. **成功**：隐藏 typing dots，开始逐字显示 AI 回复
  5. **失败**：显示红色错误气泡 + "重试"按钮

- **打字速度**：每个字符 30ms
- **响应示例**：
  - "好的，我已经理解了..."
  - "非常有趣的观点！这涉及到了..."
  - 等共 5 种预设回复

#### 错误处理与重试
- **失败率**：30%（随机）
- **错误样式**：
  - 红色背景（bg-red-50）
  - 红色边框和图标
  - 错误文本："抱歉，AI 服务暂时不可用"
  - "重试"按钮

- **重试机制**：
  - 点击"重试"按钮 → 移除错误消息
  - 重新触发 AI 回复流程
  - 可进入成功或失败分支

#### 页面滚动
- **自动滚动**：每次新消息出现后，自动滚动到对话框底部
- **滚动方式**：平滑滚动（smooth behavior）

---

## 🛠️ 通用工具库 (common.js)

### 主题管理 (ThemeManager)

```javascript
ThemeManager.init()          // 初始化（页面加载时自动调用）
ThemeManager.toggle()        // 切换亮色/深色模式
ThemeManager.getCurrent()    // 获取当前主题 ('dark' | 'light')
```

### Toast 提示 (ToastManager)

```javascript
ToastManager.show(message, type, duration)
// type: 'success' | 'error' | 'info'
// duration: 毫秒，0 表示不自动关闭

// 示例：
ToastManager.show('操作成功', 'success', 3000)
ToastManager.show('网络错误', 'error', 3000)
ToastManager.show('提示信息', 'info', 2000)
```

### 剪贴板管理 (ClipboardManager)

```javascript
await ClipboardManager.copy(text, triggerButton)
// 自动显示复制反馈（按钮图标变绿 + Toast）
// 支持降级方案（旧浏览器）
```

### 防抖与节流 (Throttle)

```javascript
const debouncedFunc = Throttle.debounce(func, 300)
const throttledFunc = Throttle.throttle(func, 300)
```

### 动画工具 (AnimationUtils)

```javascript
AnimationUtils.fadeIn(element, 300)      // 淡入
AnimationUtils.fadeOut(element, 300)     // 淡出
AnimationUtils.slideDown(element, 300)   // 向下滑入
AnimationUtils.slideUp(element, 300)     // 向上滑出
AnimationUtils.scale(element, 1.02, 300) // 缩放
```

### 输入框自适应 (InputAutoResize)

```javascript
InputAutoResize.init(inputElement, maxHeight = 150)
// 使输入框高度自动调整，支持多行输入
```

### Modal 对话框 (ModalManager)

```javascript
ModalManager.confirm(message, onConfirm, onCancel)
// 显示确认对话框，用户可点击确认或取消
```

---

## 🎨 CSS 动画定义

以下动画已在 common.js 中注入：

| 动画名称 | 效果 |
|---------|------|
| slideIn | 向下淡入（从上方）|
| slideOut | 向上淡出（往上方）|
| slideUp | 向上淡入（从下方）|
| fadeIn | 淡入 |
| fadeOut | 淡出 |
| bounce | 弹跳 |
| slideInRight | 从右侧滑入 |
| slideOutRight | 向右侧滑出 |

---

## 📊 性能优化

- **事件委托**：使用委托方式绑定事件，减少内存占用
- **防抖/节流**：用于频繁触发的事件（如 input、scroll）
- **CSS 动画**：优先使用 CSS transition，仅在必要时用 JavaScript
- **requestAnimationFrame**：复杂动画使用 RAF 替代 setInterval

---

## 🐛 已知情况

### 浏览器兼容性

- **现代浏览器**（Chrome、Firefox、Safari、Edge）：完全支持
- **IE 11**：不支持（CSS Grid、CSS Variables、async/await）
- **剪贴板 API**：不支持的浏览器会自动降级为 execCommand

### 响应式设计

- **桌面**：完全支持所有功能
- **平板**：侧滑抽屉宽度可能需要调整
- **手机**：侧滑抽屉建议全屏展开（可在 CSS 中添加媒体查询）

---

## 🚀 开发和测试

### 本地测试

1. 在浏览器中直接打开 HTML 文件
2. 打开浏览器开发者工具（F12）查看 console 日志
3. 测试各功能：点击、输入、拖动等

### 调试技巧

- **Theme Toggle**：在 console 中运行 `ThemeManager.toggle()` 切换主题
- **Show Toast**：`ToastManager.show('测试', 'success')`
- **Copy Text**：`ClipboardManager.copy('test', document.querySelector('button'))`

### 性能监测

- 打开 DevTools → Performance 标签
- 录制用户交互过程
- 检查帧率（应保持 60 FPS）

---

## 📝 后续扩展建议

1. **连接真实 API**：
   - 配置中心保存请求
   - 任务管理 WebSocket 实时对话
   - RSS 源添加/删除 API

2. **数据持久化**：
   - localStorage 保存配置偏好
   - IndexedDB 存储本地对话记录

3. **更多动画效果**：
   - 页面过渡动画
   - 加载骨架屏
   - 无限滚动加载

4. **无障碍优化**：
   - ARIA 标签
   - 键盘导航
   - 高对比度模式

---

## 📞 问题排查

| 问题 | 解决方案 |
|-----|--------|
| JS 文件 404 | 检查 HTML 中的 `src` 路径是否正确 |
| 主题切换不生效 | 确保 localStorage 未被禁用 |
| 复制功能不工作 | 检查浏览器是否支持 Clipboard API，或检查安全上下文（HTTPS） |
| Toast 不显示 | 检查是否有 CSS 冲突，或在 console 查看错误信息 |
| 动画卡顿 | 检查浏览器硬件加速是否开启，或减少同时运行的动画数量 |

---

**祝你使用愉快！** 🎉
