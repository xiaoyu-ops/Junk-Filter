# LLM_MODEL_ID 环境变量问题 - 快速参考

## 问题症状
```
❌ .env 中设置：LLM_MODEL_ID=gpt-5.2
❌ 实际读到的：LLM_MODEL_ID=gpt-4
❌ Agent 用错模型，无法调用中转站 API
```

## 根本原因
```
Conda 全局环境变量：LLM_MODEL_ID=gpt-4
           ↓
Python 启动时继承：os.environ['LLM_MODEL_ID'] = 'gpt-4'
           ↓
代码加载 .env 时：发现已存在，跳过覆盖 ❌
           ↓
最终结果：使用 gpt-4 而不是 gpt-5.2
```

## 解决方案
在 `backend-python/config.py` 中：

**从**：
```python
if key not in os.environ:
    os.environ[key] = value
```

**改为**：
```python
# 强制覆盖 LLM 相关变量（优先使用 .env）
if key in ["LLM_MODEL_ID", "OPENAI_API_KEY", "LLM_BASE_URL", "LLM_TEMPERATURE", "LLM_MAX_TOKENS"]:
    os.environ[key] = value  # 覆盖全局变量
elif key not in os.environ:
    os.environ[key] = value
```

## 验证方法
```bash
# 启动时查看日志
[DEBUG] Override: LLM_MODEL_ID: gpt-4 → gpt-5.2
[DEBUG] Final os.environ['LLM_MODEL_ID']: gpt-5.2

# 或直接在 Python 中检查
import os
print(os.getenv('LLM_MODEL_ID'))  # 应该显示 gpt-5.2
```

## 防止再发生
1. 检查 Conda 环境变量：`echo $LLM_MODEL_ID`
2. 如果有全局设置，删除它：`conda env config vars unset LLM_MODEL_ID -n junkfilter`
3. 只依赖 `.env` 文件配置

## 关键洞察
- **环境变量优先级**：全局变量 > Python 继承 > .env 加载 > 代码默认
- **调试技巧**：在代码最早时刻检查环境变量，不要等到 Pydantic 初始化
- **配置策略**：关键配置要显式覆盖全局变量，确保一致性

## 相关文件
- 完整诊断：`description/guides/RELAY_API_ENV_ISSUE_DIAGNOSIS.md`
- 配置指南：`description/guides/RELAY_API_SETUP.md`
