# LLM 配置指南

JunkFilter 的 Agent 聊天功能支持真实 LLM 集成，支持官方 OpenAI API 和兼容 OpenAI 格式的中转站 API。

## 快速配置

在 `.env` 文件中设置以下参数：

```env
# OpenAI API 密钥（官方或中转站提供）
OPENAI_API_KEY=sk-your-api-key-here

# 中转站地址（可选，不填则使用官方 OpenAI）
LLM_BASE_URL=https://your-relay.example.com/v1

# 模型 ID
LLM_MODEL_ID=gpt-4

# 可选配置
LLM_TEMPERATURE=0.7      # 创意程度（0.0=严谨, 1.0=随意）
LLM_MAX_TOKENS=2000      # 最大回复长度
LLM_TIMEOUT=30           # API 调用超时（秒）
```

## 中转站 vs 官方 API

| 方式 | LLM_BASE_URL 设置 | 说明 |
|------|------------------|------|
| 官方 OpenAI | 不填 / 删除此行 | 使用 api.openai.com |
| 中转站 | 填写中转站地址 | 支持任何兼容 OpenAI 格式的中转站 |

**中转站优势**：通常更便宜、延迟低、支持更多模型。

## 环境变量优先级问题（重要）

如果你使用 Conda，全局环境变量可能覆盖 `.env` 设置。代码中已有强制覆盖机制：

```python
# backend-python/config.py 中的强制覆盖逻辑
CRITICAL_KEYS = ["LLM_MODEL_ID", "OPENAI_API_KEY", "LLM_BASE_URL", "LLM_TEMPERATURE", "LLM_MAX_TOKENS"]
# 这些 key 强制使用 .env 文件的值，覆盖 Conda 全局环境变量
```

如遇 `.env` 配置未生效，检查：
```bash
# 查看 Conda 全局变量
echo $LLM_MODEL_ID

# 删除 Conda 全局设置
conda env config vars unset LLM_MODEL_ID -n junkfilter
```

## 降级行为

LLM 不可用时（API Key 未设置 / 网络错误 / 中转站故障），系统自动降级为规则匹配回复，聊天功能不中断。

## 故障排查

| 问题 | 解决方案 |
|------|---------|
| 超时 | 增加 `LLM_TIMEOUT=60` |
| Invalid model | 确认模型名称是中转站支持的 |
| Invalid API key | 确认 Key 格式正确（`sk-` 开头），无多余空格 |
| 仍返回模板回复 | 重启 Python Backend，检查日志是否有 `LLM API call failed` |
| .env 配置不生效 | 检查 Conda 全局变量是否覆盖（见上方说明） |

## 验证配置

重启 Python Backend 后，查看日志：
```
INFO: Using custom LLM base_url: https://your-relay.example.com/v1
```

或在前端向 Agent 提问，确认回复是 AI 生成而非硬编码模板。
