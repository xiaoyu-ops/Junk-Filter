# 🚀 TrueSignal 一键启动脚本说明

## 快速开始（3 步）

### 第 1 步：选择你的操作系统

**Windows 用户：**
```cmd
start-all.bat
```

**Linux/Mac 用户：**
```bash
chmod +x start-all.sh
./start-all.sh
```

### 第 2 步：等待启动完成
- 脚本会自动启动 Docker、Go、Python、Vue 等所有服务
- 完整启动通常需要 20-30 秒
- 每个服务在独立的窗口中运行，你能看到实时日志

### 第 3 步：访问应用
```
🌐 前端:        http://localhost:5173
🔌 Go API:      http://localhost:8080/health
🔌 Python API:  http://localhost:8081/health
```

---

## 📋 启动脚本包含的操作

1. ✅ 检查 Docker 状态，必要时启动 PostgreSQL 和 Redis
2. ✅ 启动 Go 后端 (REST API，端口 8080)
3. ✅ 启动 Python 后端 (LLM 服务，端口 8081，自动激活 conda junkfilter)
4. ✅ 启动 Vue 前端 (用户界面，端口 5173)

---

## 🧪 测试聊天功能

打开 http://localhost:5173 后：
1. 进入任务详情页面
2. 在右侧聊天面板输入问题
3. 尝试："现在的执行进度如何？" 或 "为什么这张卡片被标记为 SKIP？"
4. 观察 Agent 的自然语言回复

---

## 📚 更多帮助

- **完整启动指南**: 查看 `description/guides/STARTUP-GUIDE.md`
- **故障排查**: 同上文件中的"常见问题"部分
- **项目规范**: 查看 `CLAUDE.md`

---

## ⚠️ 注意

- 首次启动可能需要更长时间，因为 Docker 要初始化数据库
- 如果端口被占用，脚本会提示错误，此时需要手动释放端口或修改环境变量中的端口号
- 所有服务日志实时显示在各自的窗口中，这对调试很有帮助

---

**就这么简单！祝你使用愉快！** 🎉
