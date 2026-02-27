# 🚨 白屏问题诊断与修复方案

**诊断时间**：2026-02-26
**问题等级**：🔴 Critical - 应用可能出现白屏或部分加载失败

---

## 📋 诊断结果

### ✅ 已验证正常的部分

1. **HTML 结构** ✅
   - `index.html` 有正确的 `<div id="app"></div>`
   - 脚本正确加载：`<script type="module" src="/src/main.js"></script>`
   - 字体和图标资源正确引入

2. **App.vue** ✅
   - 根元素 `<div id="app">` 正确定义
   - `<RouterView name="navbar" />` 正确
   - `<RouterView />` 正确（默认视图）
   - 所有 HTML 标签都已闭合

3. **路由配置** ✅
   - 4 个路由都正确定义
   - named view 正确使用（navbar 和 default）
   - 所有组件都正确导入

4. **样式系统** ✅
   - `globals.css` 正确导入 Tailwind
   - 所有 @layer 配置正确
   - 无语法错误

5. **Home.vue 结构** ✅
   - 所有 HTML 标签正确闭合
   - v-if 逻辑正确：`v-if="search.isDropdownOpen"`
   - 数据绑定正确：`{{ search.selectedPlatform }}`（显示字符串，不会是 [object Object]）

---

## 🔴 可能导致白屏的问题

### 问题 1: @vueuse/core 未安装
**症状**：
```
[plugin:vue] Failed to resolve '@vueuse/core'
ReferenceError: onClickOutside is not defined
```

**影响范围**：Home.vue 会因为导入错误而加载失败

**解决方案**：✅ 运行 npm install

---

### 问题 2: 路由渲染问题
**可能的原因**：
- 组件在初始化时可能有循环依赖
- Router 在 main.js 中的初始化顺序问题

**检查点**：所有都正确

---

### 问题 3: Pinia Store 初始化
**可能的原因**：
- Store 初始化时发生错误

**检查**：stores/index.js 需要验证

---

## 🛠️ 完整的修复步骤

### 步骤 1️⃣：确保依赖完整安装

```bash
# 清除缓存
npm cache clean --force

# 删除旧的 node_modules
rm -rf node_modules package-lock.json

# 完整重新安装
npm install

# 验证 @vueuse/core 是否安装
npm list @vueuse/core
```

**预期输出**：
```
junk-filter-vue@1.0.0
└── @vueuse/core@10.8.1
```

---

### 步骤 2️⃣：验证 Tailwind CSS

确保 `tailwind.config.js` 中 @tailwindcss/forms 被正确注释：

```javascript
plugins: [
  // @tailwindcss/forms 已被移除
  // 若要使用，请运行: npm install -D @tailwindcss/forms
  // 然后取消下行注释:
  // require('@tailwindcss/forms'),
],
```

**验证**：✅ 已正确处理

---

### 步骤 3️⃣：清除浏览器缓存

```bash
# 完全清理 Vite 缓存
rm -rf .vite

# 在浏览器中：
# 按 F12 → 右键点击刷新按钮 → 清空所有缓存并硬刷新
# 或按 Ctrl+Shift+Delete 打开浏览器缓存清理
```

---

### 步骤 4️⃣：验证所有 Store 初始化

检查 `src/stores/index.js`：

```javascript
export { useThemeStore } from './useThemeStore.js'
export { useConfigStore } from './useConfigStore.js'
export { useTaskStore } from './useTaskStore.js'
export { useTimelineStore } from './useTimelineStore.js'
```

**验证**：需要检查该文件

---

### 步骤 5️⃣：重启开发服务器

```bash
# 停止当前服务（Ctrl+C）

# 完整重启
npm run dev

# 应该看到：
# VITE v5.0.11 ready in xxx ms
# ➜  Local:   http://localhost:5173/
```

---

## 🔍 排查白屏问题的步骤

### 如果页面是完全白屏

**1. 打开浏览器开发者工具（F12）**

**2. 查看 Console 标签，检查错误**

**常见错误及解决方案**：

#### 错误 1: "Cannot find module '@vueuse/core'"
```bash
npm install @vueuse/core
```

#### 错误 2: "Failed to resolve '@/composables/useSearch'"
检查文件路径是否正确，运行：
```bash
ls src/composables/useSearch.js
```

#### 错误 3: "router is not defined"
检查 App.vue 是否正确导入：
```javascript
import router from '@/router'
```

#### 错误 4: Tailwind CSS 未加载（页面显示但无样式）
```bash
# 重建样式
npm run dev

# 或检查 globals.css
cat src/styles/globals.css | head -5
```

---

### 如果页面有内容但显示不完整

**检查点**：

1. **导航栏不显示**
   - App.vue 中 `<RouterView name="navbar" />` 是否正确
   - AppNavbar.vue 是否能正确加载

2. **主要内容不显示**
   - Home.vue 是否有 v-if 初始值为 false 的根元素
   - 检查：`<main>` 是否条件渲染

3. **样式丢失**
   - 刷新浏览器（Ctrl+Shift+R）
   - 清除浏览器缓存（F12 → Storage → Clear All）

---

## ✅ 最终验证清单

运行以下命令后，逐项检查：

```bash
cd /d/TrueSignal/frontend-vue
npm cache clean --force
rm -rf node_modules package-lock.json
npm install
npm run dev
```

打开 http://localhost:5173

- [ ] **页面加载完成**（不是空白）
- [ ] **导航栏显示**（top-0）
- [ ] **主页标题显示**："What do you want to filter?"
- [ ] **搜索框显示**（平台按钮 + 输入框 + 搜索按钮）
- [ ] **平台按钮显示为 "Blog"**（不是 [object Object]）
- [ ] **点击平台按钮**，下拉菜单展开
- [ ] **快捷标签显示**（Recent News 等）
- [ ] **浏览器 Console**（F12）无红色错误
- [ ] **样式正常**（Tailwind 类有效果）

---

## 🚀 快速恢复脚本

创建 `fix-whitespace.sh`（或 `fix-whitespace.bat`）：

### Linux/Mac
```bash
#!/bin/bash
cd /d/TrueSignal/frontend-vue
echo "🧹 清理缓存..."
npm cache clean --force
rm -rf node_modules package-lock.json .vite
echo "📦 重新安装依赖..."
npm install
echo "✅ 完成！启动开发服务器..."
npm run dev
```

### Windows (PowerShell)
```powershell
cd D:\TrueSignal\frontend-vue
Write-Host "🧹 清理缓存..." -ForegroundColor Green
npm cache clean --force
Remove-Item -Recurse -Force node_modules -ErrorAction SilentlyContinue
Remove-Item -Force package-lock.json -ErrorAction SilentlyContinue
Remove-Item -Recurse -Force .vite -ErrorAction SilentlyContinue
Write-Host "📦 重新安装依赖..." -ForegroundColor Green
npm install
Write-Host "✅ 完成！启动开发服务器..." -ForegroundColor Green
npm run dev
```

---

## 📊 完整的依赖检查

运行以下命令验证所有关键依赖：

```bash
# 检查核心依赖
npm list vue vue-router pinia @vueuse/core

# 预期输出示例：
# junk-filter-vue@1.0.0
# ├── @vueuse/core@10.8.1
# ├── pinia@2.1.7
# ├── vue@3.4.21
# └── vue-router@4.3.2
```

---

## 📝 调试建议

如果问题仍未解决，在浏览器 Console 中执行：

```javascript
// 检查 Vue
console.log('Vue app:', window.__VUE_DEVTOOLS_GLOBAL_HOOK__)

// 检查 Router
console.log('Router defined:', typeof router)

// 检查 Pinia
console.log('Pinia defined:', typeof pinia)

// 查看完整错误
window.addEventListener('error', (e) => console.error('Error:', e))
window.addEventListener('unhandledrejection', (e) => console.error('Promise rejection:', e))
```

---

## 📞 常见问题解答

### Q: npm install 总是失败？
```bash
# 使用国内镜像
npm install --registry https://registry.npmmirror.com
```

### Q: 仍然看不到页面？
1. 确保 http://localhost:5173 不是 http://localhost:5174 或其他端口
2. 查看终端输出，找到正确的 URL
3. 清除浏览器地址栏的缓存建议

### Q: 样式正确但功能不工作？
检查浏览器 Console 中是否有 JavaScript 错误，然后：
```bash
npm run dev
# 查看构建输出
```

---

## 🎯 预期结果

所有步骤完成后，你应该看到：

✅ **完整页面**
- 导航栏在顶部
- 标题 "What do you want to filter?" 在中央
- 搜索框（平台选择 + 输入框 + 搜索按钮）
- 快捷标签
- Tailwind 样式完整应用

✅ **功能正常**
- 点击平台按钮，菜单展开/收起
- 选择平台后菜单关闭
- 搜索框可以输入
- 无 JavaScript 错误

---

**修复完成后，所有功能应该恢复正常！**

如有问题，请提供浏览器 Console 中的错误信息，我会进一步诊断。
