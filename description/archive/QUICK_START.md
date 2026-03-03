# 🚀 TrueSignal 快速启动指南

**日期**: 2026-02-28
**版本**: 1.0
**完成时间**: 2 分钟

---

## 📋 一句话总结

这是一个完整的 **后端一键启动脚本**，会自动按顺序启动：
1. Docker（PostgreSQL + Redis）
2. Go 后端（8080）
3. Python 评估服务（8081，可选）

然后在另一个终端启动前端即可。

---

## 🎯 快速开始 (2 分钟)

### Windows 用户

```bash
# 1. 启动所有后端服务（自动在新窗口中）
start-all-services.bat

# 2. 等待 Go 后端启动（看到 "Go 后端已启动" 提示）

# 3. 在新的终端窗口启动前端
start-frontend.bat

# 4. 打开浏览器
# http://localhost:5173
```

### Linux/Mac 用户

```bash
# 1. 启动所有后端服务
bash start-all-services.sh

# 2. 在新的终端窗口启动前端
bash start-frontend.sh

# 3. 打开浏览器
# http://localhost:5173
```

---

## 📊 启动流程图

```
┌─────────────────────────────────┐
│ 执行启动脚本                     │
│ (start-all-services.{sh|bat})   │
└────────────┬────────────────────┘
             │
             ↓
    ┌────────────────────┐
    │ Phase 1: Docker    │
    │ ✓ PostgreSQL       │
    │ ✓ Redis            │
    └────────┬───────────┘
             │
             ↓ (等待 3 秒)
    ┌────────────────────┐
    │ Phase 2: Go 后端   │
    │ ✓ 8080 端口        │
    │ ✓ API 就绪         │
    └────────┬───────────┘
             │
             ↓ (可选)
    ┌────────────────────┐
    │ Phase 3: Python    │
    │ ✓ LLM 评估服务     │
    │ ✓ 8081 端口        │
    └────────┬───────────┘
             │
             ↓
    ┌────────────────────────────┐
    │ 前端在单独终端启动         │
    │ bash start-frontend.sh      │
    │ 或 start-frontend.bat       │
    └────────┬───────────────────┘
             │
             ↓
    ┌────────────────────┐
    │ 打开浏览器         │
    │ http://localhost   │
    │ :5173              │
    └────────────────────┘
```

---

## 🔍 验证启动成功

### 检查清单

```bash
# 方法 1: 运行自动化测试
bash smoke_test.sh  # 或 smoke_test.bat

# 方法 2: 手动检查（都应该返回 200）
curl http://localhost:8080/health           # ✓ Go 后端
curl http://localhost:8080/api/sources      # ✓ API 可用
```

### 预期输出

```
========== 前置检查 ==========
✓ Go 后端 (8080) 正在运行

========== Test 1: 获取源列表 ==========
✓ 获取源列表

========== Test 2: 创建新源 ==========
✓ 源创建成功，ID: 4

...

========================================
测试总结:
  通过: 8
  失败: 0
========================================

[成功] 所有测试通过！
```

---

## 📍 服务地址

启动完成后，所有服务的访问地址：

| 服务 | 地址 | 用途 |
|------|------|------|
| **前端** | http://localhost:5173 | Web UI |
| **Go API** | http://localhost:8080 | REST API |
| **Go 健康检查** | http://localhost:8080/health | 状态检查 |
| **Python 后端** | http://localhost:8081 | LLM 评估（可选） |
| **PostgreSQL** | localhost:5432 | 数据库 |
| **Redis** | localhost:6379 | 消息队列 |

---

## 🐛 常见问题

### Q1: 启动脚本运行不了

**Windows**:
```bash
# 如果看到 "cannot be loaded" 错误，运行：
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Linux/Mac**:
```bash
# 添加执行权限
chmod +x start-all-services.sh
chmod +x stop-all-services.sh
chmod +x start-frontend.sh
chmod +x smoke_test.sh
```

### Q2: "Go 后端未运行或不可达"

**检查清单**:
1. Docker 是否运行？
   ```bash
   docker ps  # 应该看到 truesignal-db 和 truesignal-redis
   ```

2. Go 进程是否启动？
   ```bash
   # Windows
   tasklist | findstr go

   # Linux/Mac
   ps aux | grep "go run"
   ```

3. 8080 端口是否被占用？
   ```bash
   # Windows
   netstat -ano | findstr :8080

   # Linux/Mac
   lsof -i :8080
   ```

**解决**:
```bash
# 重新启动脚本
bash start-all-services.sh

# 或手动启动 Go
cd backend-go
go run main.go
```

### Q3: 前端无法连接到后端 API

**症状**: 浏览器 DevTools → Network 显示 CORS 错误或 404

**检查**:
1. Go 后端是否真的在运行？
   ```bash
   curl http://localhost:8080/health
   ```

2. API 端点是否正确？
   ```bash
   curl http://localhost:8080/api/sources
   ```

3. 前端 API_BASE_URL 是否正确？
   - 检查 `useConfigStore.js` 中的 `API_BASE_URL`
   - 应该是 `http://localhost:8080/api`

### Q4: Docker 容器无法启动

**检查 Docker 状态**:
```bash
# 查看容器日志
docker-compose logs

# 检查特定容器
docker-compose logs postgres
docker-compose logs redis

# 清理旧容器并重新启动
docker-compose down -v
docker-compose up -d
```

### Q5: 前端依赖安装失败

```bash
# 清空缓存后重新安装
cd frontend-vue
rm -rf node_modules package-lock.json
npm install
npm run dev
```

---

## 🛑 停止服务

### 优雅停止

```bash
# Windows
stop-all-services.bat

# Linux/Mac
bash stop-all-services.sh
```

### 强制停止

**Windows**:
- 关闭 Go 后端和 Python 后端的窗口
- 运行 `docker-compose down`

**Linux/Mac**:
- 按 Ctrl+C 停止脚本
- 运行 `bash stop-all-services.sh`

---

## 📊 服务启动时间

| 服务 | 启动时间 | 等待时间 |
|------|----------|---------|
| Docker | 2-5 秒 | 3 秒 |
| Go 后端 | 2-3 秒 | 10 秒（最多） |
| Python 后端 | 5-10 秒 | 15 秒（最多） |
| 前端 | 2-3 秒 | 3 秒 |
| **总计** | **~15 秒** | **~35 秒** |

---

## 🎯 验证完整链路

启动完成后，验证完整的数据流：

```bash
# 1. 添加 RSS 源
curl -X POST http://localhost:8080/api/sources \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://news.ycombinator.com/rss",
    "author_name": "Hacker News",
    "priority": 9,
    "enabled": true,
    "fetch_interval_seconds": 1800
  }' | jq '.id'
# 返回: 4 (假设)

# 2. 触发同步
curl -X POST http://localhost:8080/api/sources/4/fetch

# 3. 查看日志
curl http://localhost:8080/api/sources/4/sync-logs | jq

# 4. 访问前端
# http://localhost:5173 → Config 页面
# 应该看到新的源并能同步
```

---

## 📝 脚本文件说明

| 脚本 | 功能 | 何时使用 |
|------|------|---------|
| `start-all-services.sh/bat` | 启动所有后端服务 | 首次启动或服务停止后 |
| `start-frontend.sh/bat` | 启动前端 | 后端服务就绪后 |
| `stop-all-services.sh/bat` | 停止所有服务 | 需要关闭应用时 |
| `smoke_test.sh/bat` | 自动化测试 | 验证服务状态 |

---

## 🚀 高级用法

### 查看实时日志

```bash
# Linux/Mac
tail -f /tmp/go-backend.log
tail -f /tmp/python-backend.log

# Windows - 日志在各服务的终端窗口中
```

### 重启单个服务

```bash
# 重启 Docker
docker-compose restart

# 重启 Go 后端（需要手动）
# 关闭 Go 终端窗口，重新运行 start-all-services.sh
```

### 性能监控

```bash
# 查看 API 响应时间
curl -w "\nTotal time: %{time_total}s\n" \
  http://localhost:8080/api/sources

# 压力测试（如果安装了 Apache Bench）
ab -n 100 -c 10 http://localhost:8080/api/sources
```

---

## 📚 相关文档

- **API 完整参考**: `API_QUICK_REFERENCE.md`
- **集成指南**: `description/API_INTEGRATION_GUIDE.md`
- **实施清单**: `description/API_INTEGRATION_IMPLEMENTATION.md`
- **方案总结**: `description/FULL_INTEGRATION_SUMMARY.md`

---

## ✅ 完整检查清单

启动前：
- [ ] Docker 已安装
- [ ] Go 1.18+ 已安装
- [ ] Node.js 16+ 已安装
- [ ] 在项目根目录

启动中：
- [ ] 看到 "Docker 容器启动完成"
- [ ] 看到 "Go 后端已启动 (8080)"
- [ ] 看到 "所有后端服务已启动"

启动后：
- [ ] `http://localhost:5173` 可访问
- [ ] `http://localhost:8080/health` 返回 200
- [ ] `http://localhost:8080/api/sources` 返回数据
- [ ] 冒烟测试全部通过
- [ ] 浏览器无 CORS 错误

---

## 🎉 成功标志

看到以下内容说明启动成功：

```
✓ Go 后端已启动 (8080)
✓ 所有后端服务已启动！

🎉 访问地址:
  • Go API:     http://localhost:8080/api/sources
  • 前端:       http://localhost:5173

📝 快速命令:
  • 查看日志:   tail -f /tmp/go-backend.log
  • 停止服务:   bash stop-all-services.sh
  • 测试 API:   bash smoke_test.sh
```

---

## 🆘 获取帮助

如果遇到问题：

1. **查看日志**: 检查启动窗口或 `/tmp/*.log` 文件
2. **运行测试**: `bash smoke_test.sh` 获取详细错误信息
3. **检查配置**: 确保 `.env` 和 `docker-compose.yml` 正确
4. **查看文档**: 参考 `API_QUICK_REFERENCE.md` 和实施指南

---

**最后更新**: 2026-02-28
**版本**: 1.0 - 完全自动化启动

