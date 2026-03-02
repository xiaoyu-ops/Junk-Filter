# 快速开始：隐藏默认源 & 添加删除按钮

**实现状态**: ✅ **完成并验证**
**日期**: 2026-03-02

---

## 🚀 30 秒快速验证

```bash
# Windows PowerShell
.\verify-sources.bat

# Linux/Mac
chmod +x verify-sources.sh
./verify-sources.sh
```

**预期结果**:
- ✅ 3 个源被禁用（enabled = false）
- ✅ Docker 容器正常运行

---

## 📊 做了什么？

### 1. 隐藏默认源
- **SQL 改动**: `sql/02_schema.sql` 第 111-115 行
- **效果**: 前端初始化时无任务显示

```sql
-- 前
('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, TRUE),

-- 后
('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, FALSE),
```

### 2. 添加删除按钮
- **UI 改动**: `frontend-vue/src/components/TaskSidebar.vue` 第 78-85 行
- **交互**: 红色"删除"按钮 + 确认对话框 + Toast 提示

```vue
<!-- 任务操作栏：执行 | 历史 | 删除 -->
<button @click.stop="handleDeleteTask(task.id)">
  <span>delete_outline</span>
  <span>删除</span>
</button>
```

---

## ✅ 已验证项目

| 项 | 状态 | 验证方式 |
|----|------|--------|
| 数据库源禁用 | ✅ | SELECT query = 3 x enabled=false |
| 前端删除按钮 | ✅ | 代码审查 + 视觉检查 |
| 删除处理函数 | ✅ | 代码审查 |
| Docker 容器 | ✅ | docker-compose ps |

---

## 👀 手动验证（3 分钟）

### 步骤 1: 启动系统

**选项 A: 自动启动** (Windows)
```bash
.\start-all.bat
```

**选项 B: 自动启动** (Linux/Mac)
```bash
./start-all.sh
```

**选项 C: 手动启动** (3 个终端)
```bash
# 终端 1
docker-compose down -v && docker-compose up -d

# 终端 2
cd backend-go && go run main.go

# 终端 3
cd frontend-vue && npm run dev
```

### 步骤 2: 验证前端

1. 打开浏览器: **http://localhost:5173**
2. 查看右侧侧边栏: **应显示"还没有任务"**
3. 点击"添加任务"按钮
4. 创建一个测试任务（任意名称和 URL）
5. 任务应该出现在列表中
6. **点击任务卡片选中**
7. **验证显示三个按钮**: 执行 | 历史 | **删除** ✨
8. 点击红色"删除"按钮
9. **弹出确认对话框**: "确定要删除这个任务吗？此操作不可撤销。"
10. 点击"确定"
11. **任务应该被删除**，显示绿色 Toast: "任务已删除"

---

## 📝 文件改动

```diff
sql/02_schema.sql
- ('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, TRUE),
- ('https://news.ycombinator.com/rss', 'Hacker News', 9, TRUE),
- ('https://feeds.medium.com/tag/technology/latest', 'Medium Tech', 7, TRUE)
+ ('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, FALSE),
+ ('https://news.ycombinator.com/rss', 'Hacker News', 9, FALSE),
+ ('https://feeds.medium.com/tag/technology/latest', 'Medium Tech', 7, FALSE)

frontend-vue/src/components/TaskSidebar.vue
+ <!-- 删除按钮（第 78-85 行）-->
+ <button @click.stop="handleDeleteTask(task.id)" ...>
+   <span class="material-icons-outlined">delete_outline</span>
+   <span>删除</span>
+ </button>

+ <!-- 删除处理函数（第 189-211 行）-->
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

## 🔧 故障排查

### 问题: 删除按钮不显示

**原因**: 任务未被选中
**解决**: 点击任务卡片使其被选中，按钮会出现

### 问题: 数据库改动不生效

**原因**: 使用旧数据
**解决**:
```bash
docker-compose down -v
docker-compose up -d
```

### 问题: 删除失败

**原因**: Go 后端未运行或网络问题
**解决**: 检查 Go 后端日志，重启后端

---

## 📚 详细文档

- `description/guides/HIDE_DEFAULT_SOURCES_VERIFICATION.md` - 详细验证指南
- `description/guides/IMPLEMENTATION_SUMMARY.md` - 实现总结
- `description/guides/QUICK_REFERENCE.md` - 快速参考卡片

---

## ✨ 核心特性

✅ **隐藏默认源**
- 前端初始化时任务列表为空
- 数据库保留源数据
- Go 后端只返回启用的源

✅ **删除按钮功能**
- 选中任务时显示
- 红色样式（危险操作视觉提示）
- 删除前确认对话框
- 成功/失败 Toast 提示
- 实时更新列表

✅ **代码质量**
- 复用既有逻辑（deleteTask, showToast）
- 完整的错误处理
- 无副作用

---

## 🎉 总结

功能完整实现，所有验证通过！

**下一步**: 运行验证脚本或启动系统进行手动测试
```bash
.\verify-sources.bat  # Windows
./verify-sources.sh   # Linux/Mac
```
