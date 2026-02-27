# Scripts 文件夹

本文件夹包含项目的所有启动和验证脚本。

## 📋 脚本说明

### 启动脚本

#### `start-all.bat` / `start-all.sh`
启动整个项目（包括 Go 后端、Mock 后端、Vue 前端）
```bash
./scripts/start-all.sh          # Linux/Mac
.\scripts\start-all.bat         # Windows
```

#### `start-phase3.bat` / `start-phase3.sh`
启动 Phase 3 特定的服务组合
```bash
./scripts/start-phase3.sh
.\scripts\start-phase3.bat
```

#### `start-go-backend.bat`
仅启动 Go 后端服务
```bash
.\scripts\start-go-backend.bat
```

### 验证脚本

#### `verify-day1.bat` / `verify-day1.sh`
验证 Day 1 基础设施（Docker、PostgreSQL、Redis）
```bash
./scripts/verify-day1.sh
.\scripts\verify-day1.bat
```

#### `smoke-test-check.bat` / `smoke-test-check.sh`
运行冒烟测试检查应用整体状态
```bash
./scripts/smoke-test-check.sh
.\scripts\smoke-test-check.bat
```

### 诊断脚本

#### `diagnose-sse.bat` / `diagnose-sse.sh`
诊断 SSE（Server-Sent Events）连接问题
```bash
./scripts/diagnose-sse.sh
.\scripts\diagnose-sse.bat
```

## 🚀 快速开始

最简单的方法是运行 `start-all` 脚本：

```bash
# Linux/Mac
cd scripts
chmod +x *.sh
./start-all.sh

# Windows
cd scripts
start-all.bat
```

## ✅ 验证安装

启动后，运行验证脚本确保一切正常：

```bash
# Linux/Mac
./scripts/verify-day1.sh

# Windows
.\scripts\verify-day1.bat
```

## 📝 脚本维护

- 所有脚本遵循统一的命名规范
- Windows 脚本为 `.bat` 格式，Unix 脚本为 `.sh` 格式
- 保持脚本的对称性（同一功能的脚本总是成对出现）

## 🔧 自定义脚本

如需修改脚本内容：

1. 编辑相应的 `.bat` 或 `.sh` 文件
2. 测试脚本是否正常工作
3. 保持两个版本同步

---

**最后更新**: 2026-02-27
