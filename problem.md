1. 目前的问题有就是每个rss来源的文章我们都是用的一套相同的评估prompt，这样就导致了评估结果不够准确。我们需要针对不同的rss来源设计不同的评估prompt，以提高评估的准确性和针对性。
2. 目前的评估结果只包含了创新度和深度两个维度，我们可以考虑增加更多的评估维度，比如实用性、可读性等，以提供更全面的评估结果。
3. 此外我认为在每个任务下面还能创建子对话比较合理因为读者不一定想对一个任务的所有内容进行讨论，可能只想针对某个评估维度或者某个具体的评估结果进行讨论，所以在每个任务下面创建子对话可以让讨论更加有针对性和高效。
4. 我们应该丰富后端的agent的tools，增加一些针对不同评估维度的工具，这样在评估过程中agent就可以根据需要调用不同的工具来获取更准确的评估结果。
5. 还有一个小问题是目前的卡片都是没有图片的导致我们预留的头像都是默认的头像，这样看起来比较单调，我们可以考虑在评估结果中增加一个图片字段，这样在前端展示的时候就可以显示不同的图片来增加视觉效果。
6. 以及默认源是否有加入目前未知道。
7. 【未来计划】实时监测 LLM API 可用性。目前 API 中转站挂掉时系统无任何感知，评估任务静默失败，Bot 报错。计划：后台定期（如每 5 分钟）向配置的 base_url 发一个轻量探测请求，检测可用性；不可用时通过 Telegram Bot 主动推送告警通知用户；前端 Config 页面显示 API 当前状态（可用/不可用/延迟）。

---

## 待优化：终止/重启评估按钮的语义缺陷

**现状：**
- "终止评估"只执行一条 SQL：`UPDATE content SET status='DISCARDED' WHERE status IN ('PENDING','PROCESSING')`，不会真正暂停 Consumer 进程。
- "重启评估"只执行：`UPDATE content SET status='PENDING' WHERE status='DISCARDED'`，同样不通知 Consumer。

**两个已知问题：**

**问题 1：终止可能被 Stream 残留消息覆盖**
- 点击终止后，Redis Stream 里仍有这些文章的待处理消息
- Consumer 下一轮取到消息时，会把状态从 DISCARDED 改回 PROCESSING，再改成 EVALUATED
- 相当于"终止"被悄悄绕过了

**问题 2：重启后文章不会自动重新入队**
- `_requeue_pending_content`（把 PENDING 文章推进 Stream）只在 Consumer **启动时**调一次，主循环里不定期调用
- 所以点完重启，文章状态变成 PENDING，但 Consumer 感知不到，不会主动去处理
- 需要手动重启 Consumer 进程，或等 Go 有新文章进来触发下一批（也不会捞旧的 PENDING）

**建议修复方向：**
1. 终止时额外调用 `XDEL` 或 `purge-stream` 清空 Redis Stream 中对应消息，或在 Consumer 评估前检查状态是否仍为 PENDING（若已 DISCARDED 则跳过）
2. 主循环里加一个定期（如每 N 轮）调用 `_requeue_pending_content` 的逻辑，或在重启评估接口里额外向 Stream 推一条触发信号

---

## Bug：Telegram Bot 偶发性静默无响应

**现象：** 向 Bot 发消息后无任何回复，Bot 进程仍在运行。

**已排查的三个根因：**

**根因 A（最可能）：限流器静默丢弃**
- `_auth` 装饰器限流 1 秒/条，触发时 `return` 无任何反馈给用户
- 用户若在 1 秒内发送多条消息（包括连续点击/重发），后续消息全部静默丢弃
- 修复：触发限流时回复 "⏳ 请稍等 1 秒再发"，而非静默丢弃

**根因 B：`asyncio.get_event_loop()` 在 PTB 上下文中的行为不确定**
- `_auth` 装饰器用 `asyncio.get_event_loop().time()` 计时
- Python 3.10+ 推荐在协程内用 `asyncio.get_running_loop()`；`get_event_loop()` 在某些情况下可能返回错误的 loop，导致时间戳计算异常（极端情况：`now - last` 为负数或极大值，限流永远触发或永远不触发）
- 修复：改用 `asyncio.get_running_loop().time()`

**根因 C：Worker 异常退出后队列消息堆积**
- `_queue_worker` 中若 `_process_message` 抛未捕获异常，task 进入 done 状态
- 下条消息到来时 `workers[chat_id].done()` 为 True，重建 worker，但**队列中已堆积的旧消息**仍在，新 worker 会立即消费，可能触发重复处理
- 当前 `_queue_worker` 已有 try/except，此场景概率较低，但极端情况下可能静默失败

**建议修复优先级：** A > B > C

**根因 D（当前最主要）：macOS App Nap / 进程休眠**
- 电脑息屏或长时间空闲后，macOS 挂起终端进程，Telegram 长轮询连接断开，PTB 重连失败后静默卡死
- 进程仍在运行（VS Code 终端不报错），但实际不响应任何消息
- **临时处理：** 死了手动重启 Bot 进程即可
- **根本解决：** 部署到服务器后用 systemd 管理进程，不存在此问题（已列入下一步规划）