# 🎯 快速参考：隐藏默认源 & 添加删除按钮

**实现状态**: ✅ 完成
**日期**: 2026-03-02
**提交**: `5f5114d`, `47861c6`, `9786944`

---

## 📋 做了什么？

### 1️⃣ 隐藏默认源
- **改动**: `sql/02_schema.sql` - 3 个默认源 `enabled` 改为 `FALSE`
- **效果**: 前端初始化时任务列表为空（但源仍在数据库中）
- **验证**: ✅ 数据库已确认 3 个源 enabled = false

### 2️⃣ 添加删除按钮
- **改动**: `frontend-vue/src/components/TaskSidebar.vue` - 添加红色删除按钮
- **功能**: 选中任务时显示"执行 | 历史 | 删除"三个按钮
- **流程**: 点击删除 → 确认对话框 → 删除成功/失败 Toast

---

## ✅ 验证状态

| 项目 | 状态 | 证据 |
|------|------|------|
| 数据库源禁用 | ✅ | `SELECT enabled FROM sources` = 3 x FALSE |
| 前端删除按钮 | ✅ | TaskSidebar.vue 第 78-85 行 |
| 删除处理函数 | ✅ | TaskSidebar.vue 第 189-211 行 |
| Docker 容器 | ✅ | 5 个容器运行（DB, Redis, Python x3） |

---

## 🚀 快速启动

### 验证数据库改动（30 秒）
```bash
# Windows
.\verify-sources.bat

# Linux/Mac
chmod +x verify-sources.sh
./verify-sources.sh
```

### 启动完整系统（3 分钟）
```bash
# Windows
.\start-all.bat

# Linux/Mac
./start-all.sh
```

### 手动启动（3 个终端）
```bash
# 终端 1
docker-compose down -v && docker-compose up -d

# 终端 2
cd backend-go && go run main.go

# 终端 3
cd frontend-vue && npm run dev

# 浏览器访问: http://localhost:5173
```

---

## 👀 手动验证清单

**前端初始状态**:
- [ ] 访问 http://localhost:5173
- [ ] 右侧侧边栏显示"还没有任务"（空状态）
- [ ] 不显示 Ars Technica、Hacker News 等默认源

**创建和删除任务**:
- [ ] 点击"添加任务"按钮
- [ ] 创建新任务（名称、URL 等）
- [ ] 点击任务卡片选中
- [ ] 验证显示三个按钮：执行、历史、**删除**
- [ ] 点击红色"删除"按钮
- [ ] 弹出确认对话框
- [ ] 点击确认后删除
- [ ] 显示绿色 Toast: "任务已删除"
- [ ] 任务从列表移除

**取消删除**:
- [ ] 重复上述步骤，但在确认对话框点击"取消"
- [ ] 验证任务仍然存在

---

## 📚 详细文档

| 文档 | 说明 |
|------|------|
| `description/guides/HIDE_DEFAULT_SOURCES_VERIFICATION.md` | 详细验证步骤（6 个场景） |
| `description/guides/IMPLEMENTATION_SUMMARY.md` | 实现总结和验证结果 |
| `verify-sources.bat/sh` | 快速验证脚本 |

---

## 🔍 代码改动一览

**文件 1: `sql/02_schema.sql` (行 111-115)**
```diff
- ('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, TRUE),
- ('https://news.ycombinator.com/rss', 'Hacker News', 9, TRUE),
- ('https://feeds.medium.com/tag/technology/latest', 'Medium Tech', 7, TRUE)

+ ('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, FALSE),
+ ('https://news.ycombinator.com/rss', 'Hacker News', 9, FALSE),
+ ('https://feeds.medium.com/tag/technology/latest', 'Medium Tech', 7, FALSE)
```

**文件 2: `frontend-vue/src/components/TaskSidebar.vue`**
```diff
+ <!-- 删除按钮 (第 78-85 行) -->
+ <button @click.stop="handleDeleteTask(task.id)" ...>
+   <span class="material-icons-outlined">delete_outline</span>
+   <span>删除</span>
+ </button>

+ <!-- 删除处理函数 (第 189-211 行) -->
+ const handleDeleteTask = async (taskId) => {
+   if (confirm('确定要删除这个任务吗？此操作不可撤销。')) {
+     try {
+       await taskStore.deleteTask(taskId)
+       showToast({ message: '任务已删除', type: 'success' })
+     } catch (error) {
+       showToast({ message: '删除失败，请重试', type: 'error' })
+     }
+   }
+ }
```

---

## 💡 核心实现细节

### 为什么禁用而不删除源？
- ✅ 数据保留（便于恢复或分析）
- ✅ 后端 RSS 抓取服务可选择使用
- ✅ 用户体验：清空起始状态，保留选项

### 删除按钮集成
- ✅ 复用 `taskStore.deleteTask()` 方法（既有逻辑）
- ✅ 复用 `showToast()` 方法（既有 Toast 系统）
- ✅ 包含确认对话框（防止误操作）
- ✅ 成功/失败提示（用户反馈）

### 数据流向
```
PostgreSQL (enabled=FALSE)
    ↓
Go Backend /api/sources (过滤enabled=TRUE)
    ↓
Frontend (不显示禁用源)
    ↓
用户创建新源 → 创建时 enabled=TRUE → 显示在列表
用户删除源 → DELETE /api/sources/{id} → 从列表移除
```

---

## ⚠️ 常见问题

### Q: 删除按钮不显示？
**A**: 需要选中任务卡片，按钮才会出现

### Q: 数据库改动不生效？
**A**: 使用 `docker-compose down -v` 清除旧数据后重新启动

### Q: 删除后能恢复吗？
**A**: 删除是永久的（物理删除）。未来可考虑逻辑删除或回收站

### Q: API 还是返回默认源？
**A**: 重启 Go 后端，确保使用最新代码

---

## 🎉 成果

✅ **功能完整** - 隐藏默认源 + 删除按钮
✅ **验证通过** - 数据库改动已确认
✅ **代码质量** - 复用既有逻辑，错误处理完整
✅ **文档完善** - 详细验证指南 + 快速参考
✅ **脚本工具** - 验证脚本可快速检查状态

---

**立即验证**: `.\verify-sources.bat` (Windows) 或 `./verify-sources.sh` (Linux/Mac)

**完整启动**: `.\start-all.bat` (Windows) 或 `./start-all.sh` (Linux/Mac)
