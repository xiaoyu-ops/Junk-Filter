# VS Code 本地开发指南

**适用系统**: macOS  
**方案**: VS Code Tasks + 独立 Terminal 标签

---

## 方案说明

每个服务在 VS Code 里独占一个 Terminal 标签，日志实时显示在对应标签中。

**优点**：
- 不需要额外工具，VS Code 内置支持
- 每个服务日志在独立标签，一目了然
- 开发时 VS Code 本来就一直开着

**缺点**：
- 关闭 VS Code 后所有服务进程同时停止
- Mac 休眠唤醒后可能需要重启部分服务（见下方说明）

---

## 启动方式

### 方式一：一键启动全部服务

```
Cmd + Shift + P  →  输入 "Tasks: Run Task"  →  选择 "🚀 Start All Services"
```

五个服务会依次启动，每个服务在独立 Terminal 标签中显示日志。

### 方式二：单独启动某个服务

```
Cmd + Shift + P  →  "Tasks: Run Task"  →  选择对应服务
```

可选服务：

| Task 名称 | 说明 | 端口 |
|-----------|------|------|
| 🐘 DB + Redis (Docker) | 启动 PostgreSQL 和 Redis 容器 | 5432 / 6379 |
| 🐹 Go Backend :8080 | Go API 网关 + RSS 抓取服务 | 8080 |
| 🐍 Python API :8083 | FastAPI + ReAct Agent 对话 | 8083 |
| 🐍 Python Consumer (evaluator) | Redis Stream 消费者，自动 LLM 评估 | 无端口 |
| 🌐 Frontend :5173 | Vue 3 前端开发服务器 | 5173 |
| 🚀 Start All Services | 按序启动以上全部 | — |
| 🛑 Stop All Services | 停止全部服务 | — |

### 快捷键（可选配置）

默认 Build Task 快捷键：
```
Cmd + Shift + B  →  直接触发 "🚀 Start All Services"
```

---

## 查看各服务日志

启动后，VS Code 底部 Terminal 面板会出现多个标签：

```
Terminal 标签栏：
┌──────────────┬────────────┬──────────────────┬──────────────────┬──────────────┐
│ DB + Redis   │ Go :8080   │ Python API :8083  │ Python Consumer  │ Frontend :5173│
└──────────────┴────────────┴──────────────────┴──────────────────┴──────────────┘
```

点击对应标签即可看到该服务的实时日志。

### 各服务日志关键信息

**Go :8080 日志**（正常启动标志）：
```
✓ Database connected
✓ Redis connected
✓ Server starting on 0.0.0.0:8080
✓ RSS Service started
```
请求日志格式：
```
[GIN] 2026/04/11 - 21:22:42 | 200 | 1.5ms | ::1 | GET "/api/sources"
```

**Python API :8083 日志**（正常启动标志）：
```
✓ Database connected: localhost:5432/junkfilter
✓ Redis connected: redis://localhost:6379/0
✓ Python API Server initialized
Application startup complete.
```

**Python Consumer 日志**（正常启动标志）：
```
✓ Database connected
✓ Redis connected
Created consumer group: evaluators
Stream consumer running. Press Ctrl+C to stop.
Starting consumer: evaluator-1
```
有文章待评估时会显示：
```
[Evaluator] Processing item: <title>
[Evaluator] Evaluation complete: decision=INTERESTING score=8/7
```

**Frontend :5173 日志**（正常启动标志）：
```
VITE v5.x.x  ready in 384 ms
➜  Local:   http://localhost:5173/
```

---

## 健康检查

启动后可以快速验证各服务是否正常：

```bash
# 在 VS Code 新建 Terminal（Ctrl+` 然后点 +）
curl http://localhost:8080/health   # {"status":"ok"}
curl http://localhost:8083/health   # {"status":"healthy",...}
open http://localhost:5173          # 打开浏览器
```

---

## Mac 休眠唤醒后的处理

休眠期间 TCP 连接会断开，唤醒后以下服务可能需要重启：

| 服务 | 症状 | 处理 |
|------|------|------|
| Python Consumer | 日志卡住不动，无新评估 | 在对应标签 `Ctrl+C` 后重新运行 Task |
| Go 后端 | 首个请求报 DB 连接错误 | 通常会自动重连，若持续报错则重启 |
| Python API | 同 Go 后端 | 同上 |

DB + Redis 在 Docker 里不受影响，无需重启。

快速重启某个服务：在对应 Terminal 标签按 `Ctrl+C` 停掉，然后：
```
Cmd + Shift + P  →  "Tasks: Rerun Last Task"
```
或直接在该 Terminal 标签里按 `↑ Enter` 重新执行上一条命令。

---

## 停止所有服务

```
Cmd + Shift + P  →  "Tasks: Run Task"  →  "🛑 Stop All Services"
```

或者直接关掉 VS Code 窗口，所有服务会随之停止（DB + Redis 容器除外，它们跑在 Docker 里）。

单独停 Docker 容器：
```bash
docker-compose stop postgres redis
# 或完全删除容器（数据保留在 volume 里）
docker-compose down
```

---

## 服务访问地址汇总

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:5173 |
| Go API | http://localhost:8080 |
| Python API | http://localhost:8083 |
| PostgreSQL | localhost:5432 (user: junkfilter, pwd: junkfilter123, db: junkfilter) |
| Redis | localhost:6379 |
