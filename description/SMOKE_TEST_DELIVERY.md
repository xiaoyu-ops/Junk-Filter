# ✅ 冒烟测试完整包交付清单

**日期**: 2026-02-27
**状态**: ✅ 所有测试文件已准备完毕
**预计测试时间**: 20-30 分钟

---

## 📦 交付内容

### 1️⃣ 完整的冒烟测试文档

| 文件 | 用途 | 时间 |
|------|------|------|
| **SMOKE_TEST_QUICK_START.md** | ⚡ 快速启动指南（推荐先读） | 5 分钟 |
| **SMOKE_TEST_COMPLETE_GUIDE.md** | 📖 详细测试步骤和验收标准 | 20 分钟 |

### 2️⃣ 自动化检查工具

| 文件 | 平台 | 用途 |
|------|------|------|
| **smoke-test-check.sh** | Linux/Mac | 自动检查环境配置 |
| **smoke-test-check.bat** | Windows | 自动检查环境配置 |
| **browser-console-test.js** | 浏览器 Console | 自动诊断 API 连接 |

### 3️⃣ 已完成的代码

| 文件 | 修改内容 | 状态 |
|------|----------|------|
| **useAPI.js** | ✅ 数据适配层实现（80+ 行） | 完成 |
| **.env.local** | ✅ 环境变量配置 | 完成 |

---

## 🚀 立即启动测试

### 方案 A: 最快（2 分钟）

```bash
# 1. Windows
smoke-test-check.bat

# 或 Linux/Mac
./smoke-test-check.sh

# 2. 如果检查通过，打开浏览器
# http://localhost:5173

# 3. 按 F12，按照 SMOKE_TEST_QUICK_START.md 进行测试
```

### 方案 B: 完整（5 分钟）

```bash
# 如果还未启动后端，先启动所有服务

# 终端 1: Docker 基础设施
docker-compose up -d

# 终端 2: Go 后端
cd backend-go && go run main.go

# 终端 3: Mock 后端
cd backend-mock && node server.js

# 终端 4: 前端
cd frontend-vue && npm run dev

# 然后按方案 A 的步骤继续
```

---

## 📊 测试覆盖范围

### ✅ 第 1 步：任务发现测试 (3 min)
- 验证 Go 后端数据正确加载
- 验证数据适配正确（Source → Task）
- 验证 ID 格式转换正确（1 → "source-1"）

**成功指标**:
```
左侧任务列表显示 3-4 个任务
任务名称、频率、状态映射正确
无网络错误
```

### ✅ 第 2 步：写操作验证 (5 min)
- 验证前端创建任务流程
- 验证数据格式转换（Task → Source）
- 验证后端正确保存

**成功指标**:
```
POST 请求发送到 http://localhost:8080/api/sources
Request 格式: {"name":"...", "url":"...", "priority":6, "enabled":true}
Response 201/200
新任务立即显示在列表中
刷新页面后仍存在
```

### ✅ 第 3 步：混合链路测试 (5 min)
- 验证 Go (8080) 和 Mock (3000) 并行工作
- 验证消息保存到 Mock
- 验证 SSE 流式聊天工作正常

**成功指标**:
```
消息发送到 http://localhost:3000/api/messages
SSE 连接到 http://localhost:3000/api/chat/stream
两个请求互不干扰
流式回复正常显示
```

### ✅ 第 4 步：异常边界测试 (3 min)
- 验证 Go 故障时的错误处理
- 验证 Mock 仍可用
- 验证恢复能力

**成功指标**:
```
关闭 Go 后端时提示清晰错误
Mock 后端仍可继续使用消息和 SSE
重启 Go 后自动恢复
```

---

## 🎯 快速验收清单

完成 4 个测试步骤后，逐个检查：

### ✅ 数据适配层
- [ ] Source ID 自动转换为 "source-{id}" 格式
- [ ] Priority (1-10) 自动转换为 frequency (hourly/daily/weekly)
- [ ] Enabled (bool) 自动转换为 status ("active"/"paused")
- [ ] 反向转换正确（Task → Source）

### ✅ 多后端协作
- [ ] 业务 API 请求到 8080 (Go)
- [ ] 消息/SSE 请求到 3000 (Mock)
- [ ] 两者请求互不干扰
- [ ] 可同时维持多个连接

### ✅ 用户体验
- [ ] UI 流畅，无卡顿
- [ ] 数据加载快速（<1 秒）
- [ ] 流式回复平滑自然
- [ ] 错误提示清晰明了

### ✅ 整体稳定性
- [ ] 没有红色错误（Console）
- [ ] 没有 JavaScript 异常
- [ ] 没有未处理的 Promise 拒绝
- [ ] 页面性能良好

---

## 📖 文档导航

```
冒烟测试 (Smoke Test)
├── 📄 SMOKE_TEST_QUICK_START.md        ⭐ 从这里开始（5 分钟）
├── 📄 SMOKE_TEST_COMPLETE_GUIDE.md     详细步骤和验收标准（20 分钟）
│
├── 🔧 自动化工具
│   ├── smoke-test-check.bat            Windows 环境检查
│   ├── smoke-test-check.sh             Linux/Mac 环境检查
│   └── browser-console-test.js         浏览器 Console 诊断
│
└── 📊 相关文档
    ├── ADAPTER_LAYER_IMPLEMENTATION.md 适配层实现细节
    ├── BACKEND_ANALYSIS_REPORT.md      后端架构分析
    └── FRONTEND_BACKEND_COMPATIBILITY.md 兼容性评估
```

---

## 🚨 常见问题快速解答

### Q1: 我应该按什么顺序启动后端？

**A**: 顺序无关，可以同时启动。但建议：
1. Docker 容器（PostgreSQL + Redis）
2. Go 后端（业务 API）
3. Mock 后端（消息和 SSE）
4. 前端

### Q2: 我可以跳过某些步骤吗？

**A**: 不建议。完整的 4 步测试可以验证：
- ✅ 第 1 步：前端显示后端数据
- ✅ 第 2 步：前端创建数据到后端
- ✅ 第 3 步：两个后端协作
- ✅ 第 4 步：故障恢复能力

### Q3: 测试失败了怎么办？

**A**: 按照故障排查清单：
1. 验证环境变量配置（VITE_API_URL, VITE_MOCK_URL）
2. 检查各个后端是否启动
3. 查看 Network 标签的请求方向
4. 查看 Console 的错误信息
5. 重启对应的服务

### Q4: 冒烟测试通过后可以做什么？

**A**: 可以进入下一阶段：
- [ ] 性能测试（创建 100 个任务）
- [ ] 并发测试（多任务同时发送消息）
- [ ] 集成 Python 评估引擎
- [ ] 添加用户认证
- [ ] 部署到生产

---

## 📞 核心概念速查

### 数据适配层工作原理

```
前端 (Task 格式)              useAPI.js 适配层              Go 后端 (Source 格式)
┌──────────────────┐        ┌─────────────────┐        ┌──────────────────┐
│ id: "source-1"   │ ──────>│ adaptSourceToTask│─────>│ id: 1            │
│ frequency: daily │        │ adaptTaskToSource│        │ priority: 6      │
│ status: active   │        │                 │        │ enabled: true    │
└──────────────────┘        └─────────────────┘        └──────────────────┘
                                    ↕
                            双向自动转换
                            无需组件改动
```

### 混合后端请求流向

```
前端应用
  │
  ├──── 业务请求 ────> Go 后端 (8080)
  │     GET /api/sources
  │     POST /api/sources
  │
  └──── 消息/SSE ────> Mock 后端 (3000)
        GET /api/messages
        POST /api/messages
        GET /api/chat/stream (SSE)
```

---

## ✨ 特色亮点

### 🎯 完全透明的适配
- UI 组件无需改动
- 数据自动转换
- 对开发者透明

### 🔄 无缝迁移能力
- 消息 API 从 Mock 迁移到 Go 只需改 1 行
- 无需修改前端代码
- 可随时切换

### 💪 完善的错误处理
- 网络故障有清晰提示
- 部分故障不影响整体
- 自动恢复机制

### 📊 详细的测试覆盖
- 4 个完整测试步骤
- 20+ 个验收标准
- 自动化检查工具

---

## 🎉 预期结果

成功完成冒烟测试意味着：

```
✅ 适配层工作完美
   - Source/Task 转换无缝
   - ID 格式自动处理
   - 所有字段映射正确

✅ 混合架构可靠
   - Go 后端处理业务
   - Mock 后端处理消息/SSE
   - 两者协作无缝

✅ 系统稳定可用
   - 无 UI 崩溃或卡顿
   - 错误提示清晰
   - 数据完整正确

✅ 可进入生产阶段
   - 基础功能验证完毕
   - 可开始集成其他服务
   - 可部署到测试环境
```

---

## 📅 后续时间表

| 时间 | 任务 |
|------|------|
| 现在 | ✅ 冒烟测试（20-30 分钟） |
| 明天 | 🔜 性能和压力测试 |
| 3 天后 | 🔜 集成 Python 评估引擎 |
| 1 周后 | 🔜 添加用户认证和授权 |
| 2 周后 | 🔜 部署到生产环境 |

---

## 🎯 开始测试

### 最快的方式（推荐）

1. **打开本页面所列的快速启动文档**
   ```
   description/SMOKE_TEST_QUICK_START.md
   ```

2. **按照 4 个步骤进行**
   - 任务发现（3 min）
   - 写操作验证（5 min）
   - 混合链路（5 min）
   - 异常边界（3 min）

3. **验收清单打勾**
   - 完成所有要求的验收标准

4. **享受成功的喜悦**
   - 适配层完美工作 ✅
   - 系统稳定可用 ✅
   - 可进入下一阶段 ✅

---

**准备好了吗？开始测试吧！** 🚀

