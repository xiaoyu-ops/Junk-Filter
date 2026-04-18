# 本地小模型部署方案

**背景**：当前项目依赖外部中转站 API，存在模型随时失效的问题（已踩坑：gpt-5.4 content 字段 null）。  
**目标**：本地部署 Qwen3.5-9b，彻底摆脱外部依赖。

---

## 一、项目对模型的实际需求

| 场景 | 输入规模 | 输出要求 | 难度 |
|------|---------|---------|------|
| 内容评估（content_evaluator） | 3000 字符文章 + ~430 行 system prompt | 6 字段 JSON | 低 |
| ReAct Agent（react_agent） | 用户对话 + 工具结果，最多 8 轮 | 工具调用 + 自然语言 | 中 |
| 任务聊天（task chat） | prompt + 10 张卡片 + 5 条历史 | 自然语言 + 参数建议 | 中 |

**不需要**：强推理、数学、代码生成、超长文档（无 RAG）。

---

## 二、为什么选 Qwen3.5-9b

### 硬件匹配（M5 Pro 48GB）

Apple Silicon 统一内存 = 内存即显存，无瓶颈。

| 模型 | 占用 | 速度（M5 Pro 估算） | 备注 |
|------|------|-------------------|------|
| Qwen3.5-9b | **6.6 GB** | ~100 tok/s | ✅ 首选 |
| Qwen3.5-27b | ~17 GB | ~35 tok/s | 备选，更稳 |
| Qwen3.5-35b | ~22 GB | ~25 tok/s | 备选，最优质量 |
| Qwen3.5-122b | ~75 GB | 跑不了 | ❌ 超出内存 |

9b 跑起来飞快。你的评估任务输出约 200-300 token，一篇文章 **2-3 秒**出结果，比调外部 API 还快。

### Qwen3.5 对比 Qwen3 的升级

| | Qwen3-14B | Qwen3.5-9b |
|--|--|--|
| 参数量 | 14B | 9B（更小更快）|
| 上下文窗口 | 128K | **256K** |
| 视觉支持 | ❌ | **✅**（以后可以看文章图片）|
| 发布时间 | 2025.04 | **2026.03**（更新）|
| 工具调用 | ✅ | ✅ |

256K 上下文对这个项目是绰绰有余（最重的场景也就 8K）。视觉支持是未来扩展的红利——如果以后想让 Agent 分析 RSS 文章里的图表，直接可用。

### 针对本项目的具体胜算分析

**内容评估** ✅ 9b 完全胜任  
输入只有 3000 字符，输出是固定格式 JSON，无复杂推理，9b 稳定输出没有问题。

**ReAct Agent** ✅ 大概率够用  
5 个工具，最复杂的 `query_articles` 只有 4 个参数，MAX_ITERATIONS=8 有充足容错空间。  
Qwen3.5 对工具调用有专项优化，9b 在这个复杂度下表现良好。

**风险场景**：多轮工具调用（比如先查文章，再根据结果更新偏好）可能偶尔出错。  
→ 如果出错率高（>20%），换 `qwen3.5:27b` 即可，一行配置搞定。

---

## 三、⚠️ 必须处理的坑：Thinking 模式

Qwen3/3.5 默认开启 thinking 模式，输出会在正文前插入 `<think>...</think>` 推理块：

```
<think>
用户要我评估这篇文章，我需要分析它的创新度和深度...
</think>
{"innovation_score": 7, "depth_score": 6, ...}
```

**问题**：content_evaluator 用正则 `\{.*\}` 提取 JSON，thinking 块不影响匹配，但如果模型把 JSON 也写进 `<think>` 里就会失败。更稳妥的做法是直接关掉 thinking。

**已在代码里修复**：`content_evaluator.py` 的 system prompt 开头加了 `/no_think` 指令：

```python
self.system_prompt = """/no_think
你是一个专业的内容评估专家。
...
```

`/no_think` 是 Qwen3/3.5 的官方指令，写在 system prompt 或 user prompt 开头都有效，强制关闭 thinking 模式，直接输出结果。

**ReAct Agent 不加 `/no_think`**：保留 thinking 有助于工具调用决策，且 react_agent.py 读的是流式 text_content，不解析 JSON，thinking 块不影响功能。

---

## 四、部署步骤

### 1. 安装 Ollama

```bash
brew install ollama
```

安装完后 Ollama 会作为系统服务自动启动，也可以手动：

```bash
ollama serve   # 前台运行，日志可见
# 或
brew services start ollama  # 后台自启
```

验证服务正常：
```bash
curl http://localhost:11434/v1/models
```

### 2. 拉取模型

```bash
ollama pull qwen3.5:9b     # 6.6GB，等待下载
```

下载完成后验证：
```bash
ollama list
# 应该能看到 qwen3.5:9b
```

快速测试模型是否正常：
```bash
ollama run qwen3.5:9b "用JSON格式说你好：{\"message\":\"...\"}"
```

### 3. 接入项目（前端 Config 页面）

打开 http://localhost:5173，进入 Config 页面，修改以下三个字段：

```
Base URL:  http://localhost:11434/v1
API Key:   ollama
Model:     qwen3.5:9b
```

保存后 Python Consumer 在 60 秒内自动热加载新配置，**无需重启任何服务**。

### 4. 验证评估效果

```bash
# 触发一次手动抓取
curl -X POST http://localhost:8080/api/sources/{source_id}/fetch

# 等待 1 分钟后查看评估结果
docker exec junkfilter-db psql -U junkfilter -c \
  "SELECT title, decision, innovation_score, depth_score, tldr FROM evaluation e JOIN content c ON e.content_id=c.id ORDER BY e.id DESC LIMIT 5;"
```

重点看：
- `decision` 是否合理（INTERESTING / BOOKMARK / SKIP）
- `tldr` 是否准确总结了文章
- `reasoning` 推理过程是否有逻辑

### 5. 验证 Agent 工具调用

在前端聊天界面测试几个典型问题：

```
"最近有哪些有趣的文章？"        → 应调用 query_articles
"帮我添加 https://... 这个 RSS 源"  → 应调用 add_source
"现在管道里有多少待评估的文章？"  → 应调用 get_pipeline_status
```

观察 Agent 是否正确识别意图、调用正确的工具、参数填写是否准确。

---

## 五、如果 9b 不够用

升级只需一行：

```bash
ollama pull qwen3.5:27b   # ~17GB，质量更稳定
```

然后前端 Config 页面把 model 改成 `qwen3.5:27b`，60 秒生效。

升级触发条件建议：
- Agent 工具调用错误率 > 20%（参数填错、调错工具、幻觉答复）
- 评估结果 decision 明显不合理（垃圾文章评 INTERESTING，好文章评 SKIP）

---

## 六、对比总结

| 维度 | 外部中转站（现状） | 本地 Qwen3.5-9b |
|------|-----------------|----------------|
| 稳定性 | 模型随时可能失效 | 完全自控 |
| 费用 | 按 token 计费 | 零费用 |
| 速度 | 网络延迟 + 排队 | ~100 tok/s，2-3s/篇 |
| 隐私 | 文章发送外部服务器 | 完全本地 |
| 代码改动 | — | 零（Config 页面改配置）|
| 内存占用 | — | 6.6GB 常驻 |
| 扩展性 | 受制于中转站 | 随时换更大模型 |
