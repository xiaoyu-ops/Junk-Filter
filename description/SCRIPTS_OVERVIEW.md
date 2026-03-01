# TrueSignal 启动脚本总览

**最后更新**: 2026-02-28
**所有脚本已就绪**，可立即使用

---

## 📋 脚本清单

### 一键启动脚本（**首选**）

| 脚本 | 系统 | 功能 | 启动时间 |
|------|------|------|---------|
| **start-all-services.sh** | Linux/Mac | 一键启动所有后端 | ~15 秒 |
| **start-all-services.bat** | Windows | 一键启动所有后端 | ~15 秒 |

### 单独启动脚本

| 脚本 | 系统 | 功能 | 何时使用 |
|------|------|------|---------|
| **start-frontend.sh** | Linux/Mac | 仅启动前端 | 后端已运行 |
| **start-frontend.bat** | Windows | 仅启动前端 | 后端已运行 |

### 停止脚本

| 脚本 | 系统 | 功能 |
|------|------|------|
| **stop-all-services.sh** | Linux/Mac | 优雅停止所有服务 |
| **stop-all-services.bat** | Windows | 优雅停止所有服务 |

### 测试脚本

| 脚本 | 系统 | 功能 |
|------|------|------|
| **smoke_test.sh** | Linux/Mac | 自动化冒烟测试（8 个 test case） |
| **smoke_test.bat** | Windows | 自动化冒烟测试 |

---

## 🚀 推荐使用流程

### 第一次启动（新安装）

```bash
# Windows
start-all-services.bat
# 然后在新终端
start-frontend.bat

# Linux/Mac
bash start-all-services.sh
# 然后在新终端
bash start-frontend.sh
```

### 验证启动

```bash
# Windows
smoke_test.bat

# Linux/Mac
bash smoke_test.sh
```

### 停止服务

```bash
# Windows
stop-all-services.bat

# Linux/Mac
bash stop-all-services.sh
```

---

## 📊 启动顺序和时间线

```
时间  服务                     状态
────────────────────────────────────
0s   start-all-services
     ↓
2s   ✓ Docker 启动
     ↓
5s   ✓ PostgreSQL 就绪
     ✓ Redis 就绪
     ↓
7s   ✓ Go 后端启动
     ↓
10s  ✓ Go 后端就绪 (8080)
     ✓ API 可用
     ↓
12s  ✓ Python 后端启动（可选）
     ↓
15s  ✓ 所有后端服务就绪
     (终端显示: "所有后端服务已启动！")
     ↓
~30s 前端启动
     ↓
~40s ✓ 前端就绪 (5173)
     (浏览器可访问)
```

---

## 🔍 如何确认启动成功

### 方法 1: 运行冒烟测试（**最简单**）

```bash
# Windows
smoke_test.bat

# Linux/Mac
bash smoke_test.sh

# 应该看到:
# ✓ CORS 配置正确
# ✓ Test 1: 获取源列表
# ✓ Test 2: 创建新源
# ... (总共 8 个测试)
# 测试总结: 通过 8, 失败 0
```

### 方法 2: 手动验证

```bash
# 1. 检查 Docker
docker ps
# 应该看到: truesignal-db, truesignal-redis

# 2. 检查 Go 后端
curl http://localhost:8080/health
# 应该返回: {"status":"ok","time":"..."}

# 3. 检查 API
curl http://localhost:8080/api/sources | jq
# 应该返回: [{"id":1,"url":"..."},...]

# 4. 检查前端
# 在浏览器打开: http://localhost:5173
# 应该看到: TrueSignal 应用界面
```

---

## 🛠️ 脚本功能详解

### start-all-services (启动所有服务)

**包括的步骤**:
1. ✓ 清理旧的 Docker 容器
2. ✓ 启动 PostgreSQL 和 Redis
3. ✓ 等待数据库就绪
4. ✓ 验证初始数据（3 个 RSS 源）
5. ✓ 启动 Go 后端（8080）
6. ✓ 启动 Python 后端（8081，可选）
7. ✓ 验证所有服务状态
8. ✓ 显示访问地址和快速命令

**输出示例**:
```
✓ Go 后端已启动 (PID: 12345)
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

### start-frontend (仅启动前端)

**前提条件**:
- Go 后端已在 8080 运行
- 前端依赖已安装（首次会自动安装）

**包括的步骤**:
1. ✓ 检查前端目录
2. ✓ 安装 npm 依赖（如需）
3. ✓ 启动 Vite 开发服务器
4. ✓ 显示访问地址

**输出示例**:
```
🚀 启动前端开发服务器...

前端访问地址: http://localhost:5173
后端 API: http://localhost:8080/api

按 Ctrl+C 停止前端服务
```

---

### stop-all-services (停止所有服务)

**包括的步骤**:
1. ✓ 停止 Go 后端进程
2. ✓ 停止 Python 后端进程
3. ✓ 停止所有 Docker 容器
4. ✓ 清理临时文件

**输出示例**:
```
✓ Go 后端已停止 (PID: 12345)
✓ Python 后端已停止 (PID: 12346)
✓ Docker 容器已停止

✓ 所有 TrueSignal 服务已停止
```

---

### smoke_test (自动化冒烟测试)

**包括的测试**:
1. ✓ CORS 配置检查
2. ✓ 获取源列表
3. ✓ 创建新源
4. ✓ 获取单个源
5. ✓ 更新源
6. ✓ 触发同步
7. ✓ 获取同步日志
8. ✓ 删除源
9. ✓ 检查前端

**输出示例**:
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

## 🐛 故障排查

### 问题：脚本无法运行

**Windows**:
```powershell
# 解决权限问题
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Linux/Mac**:
```bash
# 添加执行权限
chmod +x *.sh
```

---

### 问题：Docker 容器无法启动

```bash
# 检查 Docker 状态
docker ps
docker-compose logs

# 重新启动
docker-compose down
docker-compose up -d
```

---

### 问题：后端启动失败

```bash
# 查看日志
tail -f /tmp/go-backend.log  # Linux/Mac

# 手动启动 Go 以查看错误
cd backend-go
go run main.go
```

---

### 问题：前端依赖安装失败

```bash
cd frontend-vue
rm -rf node_modules
npm install
npm run dev
```

---

## 📚 关键文件位置

```
D:\TrueSignal\
├── start-all-services.sh         ← 后端一键启动（推荐）
├── start-all-services.bat
├── start-frontend.sh             ← 前端启动
├── start-frontend.bat
├── stop-all-services.sh          ← 停止所有服务
├── stop-all-services.bat
├── smoke_test.sh                 ← 自动化测试
├── smoke_test.bat
├── QUICK_START.md                ← 快速入门指南
├── API_QUICK_REFERENCE.md        ← API 参考
└── description/
    ├── API_INTEGRATION_GUIDE.md
    ├── API_INTEGRATION_IMPLEMENTATION.md
    ├── FULL_INTEGRATION_SUMMARY.md
    └── ...
```

---

## 🎯 完整工作流

### 第一次使用（5 分钟）

```bash
# 1. 一键启动所有后端
bash start-all-services.sh
# 等待看到: "所有后端服务已启动！"

# 2. 在新终端启动前端
bash start-frontend.sh
# 等待看到: "ready in xxx ms"

# 3. 打开浏览器
# http://localhost:5173

# 4. 进入 Config 页面，验证真实数据加载
```

### 日常使用

```bash
# 启动
bash start-all-services.sh
bash start-frontend.sh

# 工作...

# 停止
bash stop-all-services.sh
```

### 验证一切正常

```bash
bash smoke_test.sh
# 应该看到 "所有测试通过！"
```

---

## 📊 资源使用

| 资源 | 用量 | 备注 |
|------|------|------|
| CPU | ~5-10% | 待机时 |
| 内存 | ~1-2 GB | Docker + Go + Node.js |
| 磁盘 | ~2-3 GB | node_modules + 容器镜像 |
| 网络 | ~1 MB/min | 轮询同步日志 |

---

## ✅ 启动检查清单

启动前：
- [ ] Docker 已安装且运行
- [ ] Go 1.18+ 已安装
- [ ] Node.js 16+ 已安装
- [ ] 在项目根目录

启动中：
- [ ] 脚本无权限错误
- [ ] Docker 容器正常启动
- [ ] Go 后端正常启动

启动后：
- [ ] curl 能连接到 API
- [ ] 前端加载无错误
- [ ] 冒烟测试全部通过

---

## 🚀 快速命令参考

```bash
# 启动所有
start-all-services

# 启动前端（后端已运行）
start-frontend

# 停止所有
stop-all-services

# 测试
smoke_test

# 查看日志
tail -f /tmp/go-backend.log

# 手动测试 API
curl http://localhost:8080/api/sources

# 访问前端
# http://localhost:5173
```

---

## 📞 支持信息

如需帮助：

1. **查看快速开始**: `QUICK_START.md`
2. **查看 API 参考**: `API_QUICK_REFERENCE.md`
3. **查看详细实施**: `description/API_INTEGRATION_IMPLEMENTATION.md`
4. **运行冒烟测试**: 获取详细诊断信息

---

**版本**: 1.0
**最后更新**: 2026-02-28
**状态**: ✅ 生产就绪

