# 🚀 JunkFilter 完整启动指南（Windows 环境）

## 问题分析

你的 PowerShell 中运行了 `.bat` 脚本，但遇到编码问题。解决方案有两种：

## 方案 A：使用 PowerShell 启动（推荐）

打开 PowerShell，进入项目根目录后，逐个执行以下命令：

### 步骤 1：启动 Docker 容器

```powershell
docker-compose up -d
```

**预期输出**：
```
[+] Running 5/5
 ✔ Container junkfilter-db      Started
 ✔ Container junkfilter-redis    Started
 ✔ Container junkfilter-python-1 Started
 ✔ Container junkfilter-python-2 Started
 ✔ Container junkfilter-python-3 Started
```

**验证**：
```powershell
docker-compose ps
```

应该显示 5 个容器都在运行。

### 步骤 2：启动 Go 后端（新的 PowerShell 窗口）

```powershell
# 进入 Go 目录
cd backend-go

# 编译
go build -o junkfilter-go.exe main.go

# 运行
.\junkfilter-go.exe
```

**预期输出**：
```
✓ Configuration loaded
✓ Database connected
✓ Redis connected
========== JunkFilter Backend ==========
Database: localhost:5432/truesignal
Redis: localhost:6379
Server: listening on :8080
========================================
```

### 步骤 3：启动 Python 后端（新的 PowerShell 窗口）

```powershell
# 进入 Python 目录
cd backend-python

# 激活 conda 环境
conda activate junkfilter

# 运行
python main.py
```

**预期输出**：
```
✓ Python API Server initialized
INFO:     Uvicorn running on http://127.0.0.1:8081 (Press CTRL+C to quit)
```

### 步骤 4：启动前端（新的 PowerShell 窗口）

```powershell
# 进入前端目录
cd frontend-vue

# 第一次需要安装依赖
npm install

# 启动开发服务器
npm run dev
```

**预期输出**：
```
  ➜  Local:   http://localhost:5173/
  ➜  press h + enter to show help
```

---

## 方案 B：使用命令提示符 (cmd.exe)

如果 PowerShell 有问题，直接用 cmd.exe：

### 打开 4 个命令提示符窗口

#### 窗口 1：Docker
```cmd
docker-compose up -d
```

#### 窗口 2：Go 后端
```cmd
cd backend-go
go build -o junkfilter-go.exe main.go
junkfilter-go.exe
```

#### 窗口 3：Python 后端
```cmd
cd backend-python
conda activate junkfilter
python main.py
```

#### 窗口 4：前端
```cmd
cd frontend-vue
npm install
npm run dev
```

---

## 步骤 5：验证所有服务

打开第 5 个 PowerShell 或 cmd 窗口，运行验证命令：

### 5.1 验证 Docker 容器

```powershell
docker-compose ps
```

✅ 应该看到 5 个容器都是 "Up" 状态

### 5.2 验证 Go 后端

```powershell
curl http://localhost:8080/health
```

✅ 应该返回：
```json
{"status":"ok","time":"2026-03-01T..."}
```

### 5.3 验证 Python 后端

```powershell
curl http://localhost:8081/health
```

✅ 应该返回：
```json
{"status":"healthy","database":"connected","redis":"connected","llm":"configured"}
```

### 5.4 验证数据库

```powershell
docker exec junkfilter-db psql -U truesignal -d truesignal -c "SELECT version();"
```

✅ 应该返回 PostgreSQL 版本信息

### 5.5 验证 Redis

```powershell
docker exec junkfilter-redis redis-cli ping
```

✅ 应该返回：`PONG`

---

## 步骤 6：打开前端应用

打开浏览器访问：

```
http://localhost:5173
```

✅ 应该看到 JunkFilter 应用的主界面

---

## 完整启动顺序总结

| 顺序 | 窗口 | 命令 | 端口 |
|------|------|------|------|
| 1 | PowerShell 1 | `docker-compose up -d` | 5432, 6379 |
| 2 | PowerShell 2 | `cd backend-go && go build -o junkfilter-go.exe main.go && .\junkfilter-go.exe` | 8080 |
| 3 | PowerShell 3 | `cd backend-python && conda activate junkfilter && python main.py` | 8081 |
| 4 | PowerShell 4 | `cd frontend-vue && npm install && npm run dev` | 5173 |
| 5 | 浏览器 | 访问 `http://localhost:5173` | - |

---

## 如果 `npm install` 失败

```powershell
# 清理 npm 缓存
npm cache clean --force

# 删除 node_modules 和 lock 文件
rm -r frontend-vue/node_modules
rm frontend-vue/package-lock.json

# 重新安装
cd frontend-vue
npm install
```

---

## 如果 `conda activate junkfilter` 失败

创建环境：

```powershell
# 创建新的 conda 环境
conda create -n junkfilter python=3.10

# 激活环境
conda activate junkfilter

# 进入项目目录
cd backend-python

# 安装依赖
pip install -r requirements.txt
```

---

## 快速检查清单

在继续测试之前，请确保：

- [ ] Docker Desktop 已安装并运行
- [ ] Go 1.18+ 已安装：`go version`
- [ ] Python 3.10+ 已安装：`python --version`
- [ ] Node.js 16+ 已安装：`node -v`
- [ ] npm 已安装：`npm -v`
- [ ] conda 已安装（如果使用 Python 环境）：`conda --version`

---

## 停止所有服务

```powershell
# 停止 Docker 容器
docker-compose down

# 停止 Go 后端：在 Go 终端按 Ctrl+C
# 停止 Python 后端：在 Python 终端按 Ctrl+C
# 停止前端：在前端终端按 Ctrl+C
```

---

## 完全重置

如果出现无法解决的问题：

```powershell
# 1. 停止所有容器和删除数据
docker-compose down -v

# 2. 清理 Go 编译
cd backend-go
go clean
del junkfilter-go.exe
cd ..

# 3. 清理 npm
cd frontend-vue
rm -r node_modules
rm package-lock.json
cd ..

# 4. 重新启动（按照上面的步骤 1-4）
```

---

现在请按照上面的步骤启动系统！如果遇到任何错误，请复制完整的错误消息告诉我。
