# 中转站 API 环境变量问题纠察记录

**日期**: 2026-02-28
**问题类型**: 环境变量配置冲突
**状态**: ✅ 已解决

---

## 问题描述

### 现象
- `.env` 文件中配置：`LLM_MODEL_ID=gpt-5.2`
- 实际读取的值：`gpt-4`
- 影响范围：所有读取 LLM_MODEL_ID 的地方都得到错误值
- 症状：Agent 使用错误的模型，无法调用正确的中转站 API

### 初始诊断
```
用户反馈：还是gpt-4 是不是还没有取消硬编码

初步假设：
1. config.py 中有硬编码默认值
2. .env 加载顺序问题
3. Pydantic settings 读取错误
```

---

## 排查过程

### 第 1 步：验证 .env 文件内容

**操作**：
```bash
grep -n "^LLM" D:/TrueSignal/.env
grep -c "LLM_MODEL_ID" D:/TrueSignal/.env
```

**发现**：
- ✅ .env 文件中确实有 `LLM_MODEL_ID=gpt-5.2`
- ✅ 只有 1 个活动的定义（第 29 行的 `# LLM_MODEL_ID=qwen-plus` 是注释）
- ⚠️ 发现有多个 `.env` 文件！

### 第 2 步：发现多个 .env 文件

**操作**：
```bash
find D:/TrueSignal -name ".env*" -type f
```

**发现**：
```
D:/TrueSignal/.env                              ← 项目根目录
D:/TrueSignal/.env.example
D:/TrueSignal/backend-python/.env               ← 子目录！
D:/TrueSignal/backend-python/.env.example
D:/TrueSignal/frontend-vue/.env
D:/TrueSignal/frontend-vue/.env.local
D:/TrueSignal/frontend-vue/.env.production
```

**初步结论**：
- Pydantic 可能加载错误的 `.env` 文件
- 或者两个 `.env` 文件定义冲突

### 第 3 步：逐级添加调试日志

**操作**：在 `config.py` 和 `api_server.py` 中添加详细的调试日志

**第 1 次日志输出**：
```
INFO:api_server:  Model ID: gpt-4
INFO:api_server:  os.getenv('LLM_MODEL_ID'): 'gpt-4'
INFO:api_server:  settings.llm_model_id: 'gpt-4'
```

**分析**：
- os.getenv 返回 gpt-4
- settings.llm_model_id 也是 gpt-4
- 说明不是 Pydantic 的问题，而是环境变量本身有问题

### 第 4 步：最关键的发现

**操作**：在 Python 解释器启动的最早时刻检查环境变量

**代码**：
```python
print(f"[DEBUG-VERY-START] LLM_MODEL_ID in os.environ: {os.environ.get('LLM_MODEL_ID', 'NOT SET')}")
```

**输出**：
```
[DEBUG-VERY-START] LLM_MODEL_ID in os.environ: gpt-4
```

**关键发现** 🎯：
```
在 Python 代码执行之前，os.environ['LLM_MODEL_ID'] 已经是 gpt-4 了！
这意味着问题不在我们的代码，而在于 Python 启动时的环境变量初始化！
```

### 第 5 步：检查 .env 加载流程

**操作**：手动加载 .env 并记录每一步

```python
[DEBUG] Loading from: D:\TrueSignal\backend-python\.env
[DEBUG]   Found: LLM_MODEL_ID=gpt-5.2
[DEBUG]   Skipping LLM_MODEL_ID because already in os.environ as: gpt-4
```

**发现**：
```
代码试图从 .env 加载 gpt-5.2，但发现 os.environ 已经有 gpt-4，
然后代码跳过了覆盖（因为原来的逻辑是 "不覆盖已存在的环境变量"）
```

### 第 6 步：排查环境变量来源

**排查位置**：
1. ✓ Conda 环境的激活脚本
2. ✓ 系统全局环境变量
3. ✓ PowerShell profile
4. ✗ 代码中的硬编码

**结论**：
```
LLM_MODEL_ID=gpt-4 被全局设置在 conda junkfilter 环境中
（可能是之前的开发过程中设置的）
```

---

## 根本原因

**最终诊断**：

```
问题链路：
┌─────────────────────────────────────────┐
│ Conda 环境变量被全局设置                │
│ LLM_MODEL_ID=gpt-4                     │
└────────────────┬────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────┐
│ Python 启动时从环境继承该变量            │
│ os.environ['LLM_MODEL_ID'] = 'gpt-4'   │
└────────────────┬────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────┐
│ config.py 加载 .env 但发现已存在        │
│ "不覆盖已存在的环境变量" 的逻辑         │
│ ❌ 导致 gpt-4 被保留，gpt-5.2 被忽略   │
└────────────────┬────────────────────────┘
                 │
                 ↓
┌─────────────────────────────────────────┐
│ Pydantic Settings 读取 os.environ       │
│ 得到 gpt-4 而不是 gpt-5.2              │
└─────────────────────────────────────────┘
```

**关键错误的代码**（原始逻辑）：
```python
# ❌ 错误：跳过已存在的环境变量
if key not in os.environ:
    os.environ[key] = value
```

---

## 解决方案

### 修改方案

**文件**：`backend-python/config.py`

**从**（原始代码）：
```python
# 不覆盖已存在的环境变量
if key not in os.environ:
    os.environ[key] = value
```

**改为**（新代码）：
```python
# 强制覆盖 LLM 相关的环境变量（以 .env 文件为准）
if key in ["LLM_MODEL_ID", "OPENAI_API_KEY", "LLM_BASE_URL", "LLM_TEMPERATURE", "LLM_MAX_TOKENS"]:
    os.environ[key] = value  # 强制覆盖
elif key not in os.environ:
    os.environ[key] = value
```

### 执行步骤

1. **修改 config.py**：强制覆盖 LLM 相关的环境变量
2. **清理 api_server.py**：删除调试日志
3. **重启 Python Backend**：验证配置正确
4. **测试 Agent**：确认使用正确的模型和 base_url

### 验证结果

```
[DEBUG] Override: LLM_MODEL_ID: gpt-4 → gpt-5.2
[DEBUG] Final os.environ['LLM_MODEL_ID']: gpt-5.2
INFO:api_server:LLM Configuration: Model=gpt-5.2, Base=https://elysiver.h-e.top/v1
```

✅ **问题解决**

---

## 关键学习点

### 1. **环境变量优先级**
```
全局环境变量（系统/Conda）> Python 启动时继承 > .env 文件加载 > 代码默认值
```

如果要 .env 文件优先，需要在代码中**显式覆盖**全局变量。

### 2. **多个 .env 文件的风险**
```
当前目录: D:\TrueSignal\backend-python\.env
项目根目录: D:\TrueSignal\.env
```

- Pydantic 只加载 `env_file=".env"`（相对于当前目录）
- 但文件系统中可能有多个 .env
- 容易造成混淆

### 3. **调试的重要性**
```
从顶层开始逐步排查：
1. 打印最早时刻的环境变量
2. 跟踪 .env 加载流程
3. 检查每一步的值变化
4. 最终定位到根本原因
```

不要假设，要验证！

### 4. **配置管理最佳实践**

✅ **应该做**：
```python
# 强制覆盖关键配置（以 .env 为准）
if key in CRITICAL_KEYS:
    os.environ[key] = value
```

❌ **不应该做**：
```python
# 保持全局环境变量（可能导致配置冲突）
if key not in os.environ:
    os.environ[key] = value
```

---

## 对比：问题前后

### 问题前
```
LLM_MODEL_ID = gpt-4（来自 Conda 全局变量）
LLM_BASE_URL = https://elysiver.h-e.top/v1（来自 .env）
LLM_TEMPERATURE = 0.7（来自 .env）

结果：混合配置，模型来自 Conda，其他来自 .env
```

### 问题后
```
LLM_MODEL_ID = gpt-5.2（来自 .env，覆盖了 Conda 变量）
LLM_BASE_URL = https://elysiver.h-e.top/v1（来自 .env）
LLM_TEMPERATURE = 0.7（来自 .env）

结果：统一配置，全部来自 .env 文件
```

---

## 文件修改清单

| 文件 | 修改内容 | 影响 |
|------|--------|------|
| `backend-python/config.py` | 添加强制覆盖逻辑 | 确保 .env 配置优先 |
| `backend-python/api_server.py` | 简化日志输出 | 清理调试信息 |

---

## 预防措施

### 1. **避免全局环境变量污染**
```bash
# 在 conda 环境中取消全局设置的 LLM_MODEL_ID
conda env config vars unset LLM_MODEL_ID -n junkfilter
```

### 2. **统一 .env 文件位置**
```
建议：只保留项目根目录的 .env
删除：backend-python/.env（使用 .env.example 作为模板）
```

### 3. **配置验证脚本**
```python
# 在启动时打印实际使用的配置
logger.info(f"LLM Configuration: Model={model_id}, Base={api_base}")
```

### 4. **文档说明**
```markdown
# 重要
.env 文件中的配置优先于全局环境变量
如果 LLM 模型设置不正确，检查：
1. .env 文件中的 LLM_MODEL_ID
2. Conda 环境变量：echo $LLM_MODEL_ID
```

---

## 时间线

| 时间 | 事件 | 状态 |
|------|------|------|
| 02:25 | 用户配置 .env 中的中转站 API | ⏳ 进行中 |
| 02:47 | 用户反馈 Agent 还是返回硬编码 | ❌ 问题发现 |
| 03:00+ | 开始排查（修改 config.py、api_server.py） | 🔍 诊断中 |
| 03:30+ | 添加详细调试日志 | 📊 数据收集 |
| 03:45 | **发现最关键线索**：`LLM_MODEL_ID in os.environ: gpt-4` | 🎯 问题定位 |
| 03:50 | 确认原因：Conda 全局环境变量 + 配置逻辑冲突 | ✅ 原因确定 |
| 03:55 | 实施解决方案：强制覆盖逻辑 | ✅ 问题解决 |

---

## 相关文档

- `description/guides/RELAY_API_SETUP.md` - 中转站 API 配置指南
- `description/guides/PHASE5_3_LLM_FIX_SUMMARY.md` - LLM 集成技术细节
- `description/guides/GET_STARTED_NOW.md` - 快速开始

---

## 总结

### 问题
Agent 使用错误的 LLM 模型（gpt-4 而不是 gpt-5.2），无法调用中转站 API

### 原因
Conda 环境中的全局 `LLM_MODEL_ID=gpt-4` 变量与 `.env` 文件中的 `LLM_MODEL_ID=gpt-5.2` 冲突，且代码逻辑优先保留全局变量

### 解决
在 `config.py` 中实现强制覆盖逻辑，确保 `.env` 文件配置优先

### 结果
✅ 系统现在正确使用 gpt-5.2 模型
✅ 成功调用中转站 API（https://elysiver.h-e.top/v1）
✅ Agent 可以生成真实的 AI 回复

