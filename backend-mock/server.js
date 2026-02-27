#!/usr/bin/env node

/**
 * TrueSignal Mock 后端服务器
 *
 * 用途：
 * 1. 为前端提供完整的 REST API
 * 2. 实现 SSE 流式聊天
 * 3. 使用 JSON 文件模拟数据库
 * 4. 为真实后端开发提供参考
 *
 * 使用方法：
 * node server.js
 * 服务器运行在 http://localhost:3000
 */

const http = require('http')
const fs = require('fs')
const path = require('path')
const url = require('url')

const PORT = 3000
const DATA_DIR = path.join(__dirname, 'data')

// 确保 data 目录存在
if (!fs.existsSync(DATA_DIR)) {
  fs.mkdirSync(DATA_DIR, { recursive: true })
}

// 数据文件路径
const TASKS_FILE = path.join(DATA_DIR, 'tasks.json')
const MESSAGES_FILE = path.join(DATA_DIR, 'messages.json')
const EXECUTION_HISTORY_FILE = path.join(DATA_DIR, 'execution-history.json')

// 初始化数据文件
function initializeDataFiles() {
  if (!fs.existsSync(TASKS_FILE)) {
    const initialTasks = [
      {
        id: 'task-1',
        name: 'Twitter AI 新闻早报',
        command: '每天早上9点总结Twitter上关于AI的新闻，并发送到邮箱',
        frequency: 'daily',
        execution_time: '09:00',
        notification_channels: ['email'],
        status: 'active',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      {
        id: 'task-2',
        name: '技术文章周报',
        command: '每周一上午10点汇总过去一周的技术文章',
        frequency: 'weekly',
        execution_time: '10:00',
        notification_channels: ['email', 'slack'],
        status: 'active',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ]
    fs.writeFileSync(TASKS_FILE, JSON.stringify(initialTasks, null, 2))
  }

  if (!fs.existsSync(MESSAGES_FILE)) {
    fs.writeFileSync(MESSAGES_FILE, JSON.stringify([], null, 2))
  }

  if (!fs.existsSync(EXECUTION_HISTORY_FILE)) {
    fs.writeFileSync(EXECUTION_HISTORY_FILE, JSON.stringify([], null, 2))
  }
}

// 读取任务数据
function readTasks() {
  try {
    const data = fs.readFileSync(TASKS_FILE, 'utf-8')
    return JSON.parse(data)
  } catch (error) {
    console.error('读取任务文件失败:', error)
    return []
  }
}

// 保存任务数据
function writeTasks(tasks) {
  fs.writeFileSync(TASKS_FILE, JSON.stringify(tasks, null, 2))
}

// 读取消息数据
function readMessages() {
  try {
    const data = fs.readFileSync(MESSAGES_FILE, 'utf-8')
    return JSON.parse(data)
  } catch (error) {
    console.error('读取消息文件失败:', error)
    return []
  }
}

// 保存消息数据
function writeMessages(messages) {
  fs.writeFileSync(MESSAGES_FILE, JSON.stringify(messages, null, 2))
}

// 读取执行历史数据
function readExecutionHistory() {
  try {
    const data = fs.readFileSync(EXECUTION_HISTORY_FILE, 'utf-8')
    return JSON.parse(data)
  } catch (error) {
    console.error('读取执行历史文件失败:', error)
    return []
  }
}

// 保存执行历史数据
function writeExecutionHistory(history) {
  fs.writeFileSync(EXECUTION_HISTORY_FILE, JSON.stringify(history, null, 2))
}

// 生成唯一 ID
function generateId() {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

// 睡眠函数
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

// 发送 JSON 响应
function sendJson(res, statusCode, data) {
  res.writeHead(statusCode, {
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type',
  })
  res.end(JSON.stringify({ data, success: statusCode < 400 }))
}

// 发送错误响应
function sendError(res, statusCode, message) {
  res.writeHead(statusCode, {
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
  })
  res.end(JSON.stringify({ error: message, success: false }))
}

// 创建服务器
const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true)
  const pathname = parsedUrl.pathname
  const query = parsedUrl.query
  const method = req.method

  console.log(`[${method}] ${pathname}`)

  // CORS preflight
  if (method === 'OPTIONS') {
    res.writeHead(200, {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    })
    res.end()
    return
  }

  // 路由处理
  try {
    if (pathname === '/api/tasks' && method === 'GET') {
      handleGetTasks(res)
    } else if (pathname === '/api/tasks' && method === 'POST') {
      handleCreateTask(req, res)
    } else if (pathname.match(/^\/api\/tasks\/[^/]+$/) && method === 'GET') {
      const taskId = pathname.split('/')[3]
      handleGetTask(res, taskId)
    } else if (pathname.match(/^\/api\/tasks\/[^/]+$/) && method === 'PUT') {
      const taskId = pathname.split('/')[3]
      handleUpdateTask(req, res, taskId)
    } else if (pathname.match(/^\/api\/tasks\/[^/]+$/) && method === 'DELETE') {
      const taskId = pathname.split('/')[3]
      handleDeleteTask(res, taskId)
    } else if (pathname.match(/^\/api\/tasks\/[^/]+\/messages$/) && method === 'GET') {
      const taskId = pathname.split('/')[3]
      handleGetMessages(res, taskId, query)
    } else if (pathname.match(/^\/api\/tasks\/[^/]+\/execute$/) && method === 'POST') {
      const taskId = pathname.split('/')[3]
      handleExecuteTask(req, res, taskId)
    } else if (pathname.match(/^\/api\/tasks\/[^/]+\/execution-history$/) && method === 'GET') {
      const taskId = pathname.split('/')[3]
      handleGetExecutionHistory(res, taskId, query)
    } else if (pathname === '/api/messages' && method === 'POST') {
      handleSaveMessage(req, res)
    } else if (pathname === '/api/messages/search' && method === 'GET') {
      handleSearchMessages(res, query)
    } else if (pathname === '/api/messages/export' && method === 'GET') {
      handleExportMessages(res, query)
    } else if (pathname.match(/^\/api\/messages\/[^/]+$/) && method === 'PUT') {
      const messageId = pathname.split('/')[3]
      handleUpdateMessage(req, res, messageId)
    } else if (pathname === '/api/chat/stream' && method === 'GET') {
      handleChatStream(req, res, query)
    } else {
      sendError(res, 404, '端点不存在')
    }
  } catch (error) {
    console.error('处理请求出错:', error)
    sendError(res, 500, '服务器错误')
  }
})

// 获取任务列表
function handleGetTasks(res) {
  const tasks = readTasks()
  sendJson(res, 200, tasks)
}

// 获取单个任务
function handleGetTask(res, taskId) {
  const tasks = readTasks()
  const task = tasks.find(t => t.id === taskId)
  if (task) {
    sendJson(res, 200, task)
  } else {
    sendError(res, 404, '任务不存在')
  }
}

// 创建任务
function handleCreateTask(req, res) {
  let body = ''

  req.on('data', chunk => {
    body += chunk
  })

  req.on('end', () => {
    try {
      const taskData = JSON.parse(body)
      const tasks = readTasks()

      const newTask = {
        id: generateId(),
        name: taskData.name,
        command: taskData.command,
        frequency: taskData.frequency || 'daily',
        execution_time: taskData.execution_time || '00:00',
        notification_channels: taskData.notification_channels || [],
        status: 'active',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }

      tasks.push(newTask)
      writeTasks(tasks)

      console.log(`✅ 创建任务: ${newTask.id}`)
      sendJson(res, 201, newTask)
    } catch (error) {
      console.error('创建任务失败:', error)
      sendError(res, 400, '无效的请求数据')
    }
  })
}

// 更新任务
function handleUpdateTask(req, res, taskId) {
  let body = ''

  req.on('data', chunk => {
    body += chunk
  })

  req.on('end', () => {
    try {
      const taskData = JSON.parse(body)
      const tasks = readTasks()
      const taskIndex = tasks.findIndex(t => t.id === taskId)

      if (taskIndex === -1) {
        sendError(res, 404, '任务不存在')
        return
      }

      tasks[taskIndex] = {
        ...tasks[taskIndex],
        ...taskData,
        id: taskId,  // 防止修改 ID
        created_at: tasks[taskIndex].created_at,  // 防止修改创建时间
        updated_at: new Date().toISOString(),
      }

      writeTasks(tasks)

      console.log(`✅ 更新任务: ${taskId}`)
      sendJson(res, 200, tasks[taskIndex])
    } catch (error) {
      console.error('更新任务失败:', error)
      sendError(res, 400, '无效的请求数据')
    }
  })
}

// 删除任务
function handleDeleteTask(res, taskId) {
  const tasks = readTasks()
  const taskIndex = tasks.findIndex(t => t.id === taskId)

  if (taskIndex === -1) {
    sendError(res, 404, '任务不存在')
    return
  }

  tasks.splice(taskIndex, 1)
  writeTasks(tasks)

  // 同时删除该任务的所有消息
  let messages = readMessages()
  messages = messages.filter(m => m.task_id !== taskId)
  writeMessages(messages)

  console.log(`✅ 删除任务: ${taskId}`)
  sendJson(res, 200, { success: true })
}

// 获取任务消息历史
function handleGetMessages(res, taskId, query) {
  const limit = parseInt(query.limit) || 50
  const offset = parseInt(query.offset) || 0

  const messages = readMessages()
  const taskMessages = messages.filter(m => m.task_id === taskId)

  // 分页
  const paged = taskMessages.slice(offset, offset + limit)

  console.log(`✅ 获取消息: task_id=${taskId}, 返回 ${paged.length} 条`)
  sendJson(res, 200, paged)
}

// 保存消息
function handleSaveMessage(req, res) {
  let body = ''

  req.on('data', chunk => {
    body += chunk
  })

  req.on('end', () => {
    try {
      const messageData = JSON.parse(body)
      const messages = readMessages()

      const newMessage = {
        id: generateId(),
        task_id: messageData.task_id,
        role: messageData.role,
        type: messageData.type || 'text',
        content: messageData.content,
        timestamp: new Date().toISOString(),
        read: false,
      }

      messages.push(newMessage)
      writeMessages(messages)

      console.log(`✅ 保存消息: ${newMessage.id} (${messageData.role})`)
      sendJson(res, 201, newMessage)
    } catch (error) {
      console.error('保存消息失败:', error)
      sendError(res, 400, '无效的请求数据')
    }
  })
}

// SSE 流式聊天处理 - 修复版
async function handleChatStream(req, res, query) {
  const taskId = query.taskId
  const userMessage = query.message || '你好'

  if (!taskId) {
    sendError(res, 400, '缺少 taskId 参数')
    return
  }

  console.log(`📡 SSE 连接: taskId=${taskId}, message="${userMessage}"`)

  // SSE 响应头
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'Access-Control-Allow-Origin': '*',
  })

  // ⚠️ 修复：跟踪客户端连接状态
  let isClientConnected = true

  // ⚠️ 修复：监听客户端断开事件
  res.on('error', () => {
    console.warn(`📡 客户端错误，连接断开: taskId=${taskId}`)
    isClientConnected = false
  })

  res.on('close', () => {
    console.warn(`📡 客户端已关闭连接: taskId=${taskId}`)
    isClientConnected = false
  })

  try {
    // 生成 AI 回复（根据不同的消息返回不同的内容）
    let aiResponse = getAiResponse(userMessage)

    // 流式发送 delta 事件（逐字发送）
    console.log(`📤 发送流式数据...`)
    for (let i = 0; i < aiResponse.length; i++) {
      // ⚠️ 修复：检查客户端是否仍然连接
      if (!isClientConnected) {
        console.warn(`📡 客户端已断开，停止发送数据 (已发送 ${i}/${aiResponse.length} 字符)`)
        res.end()
        return
      }

      const char = aiResponse[i]

      res.write(`event: delta\n`)
      res.write(`data: ${JSON.stringify({
        type: 'delta',
        content: char,
      })}\n\n`)

      // 模拟网络延迟（30-70ms）
      await sleep(30 + Math.random() * 40)
    }

    // 50% 概率发送执行卡片
    if (Math.random() > 0.5) {
      // ⚠️ 修复：检查客户端连接
      if (!isClientConnected) {
        console.warn(`📡 客户端已断开，跳过执行卡片`)
        res.end()
        return
      }

      console.log(`📤 发送执行卡片...`)
      await sleep(500)

      res.write(`event: execution\n`)
      res.write(`data: ${JSON.stringify({
        type: 'execution',
        status: 'success',
        itemCount: Math.floor(Math.random() * 100) + 10,
        summary: `成功处理了关于"${userMessage}"的请求，获取了相关信息。`,
        timestamp: new Date().toISOString(),
      })}\n\n`)
    }

    // 发送完成事件
    // ⚠️ 修复：检查客户端连接
    if (isClientConnected) {
      console.log(`✅ SSE 流式完成`)
      res.write(`event: done\n`)
      res.write(`data: ${JSON.stringify({ type: 'done' })}\n\n`)
    }

    res.end()

    // 保存 AI 消息到数据库（异步，不阻塞响应）
    const messages = readMessages()
    messages.push({
      id: generateId(),
      task_id: taskId,
      role: 'ai',
      type: 'text',
      content: aiResponse,
      timestamp: new Date().toISOString(),
      read: false,
    })
    writeMessages(messages)
  } catch (error) {
    console.error('SSE 流式处理错误:', error)

    // ⚠️ 修复：只有在客户端仍连接时才发送错误
    if (isClientConnected) {
      res.write(`event: error\n`)
      res.write(`data: ${JSON.stringify({
        type: 'error',
        message: error.message,
      })}\n\n`)
    }

    res.end()
  }
}

// 搜索消息
function handleSearchMessages(res, query) {
  const searchQuery = query.q || ''
  const taskId = query.taskId

  const messages = readMessages()
  let filtered = messages

  // 按任务ID过滤（可选）
  if (taskId) {
    filtered = filtered.filter(m => m.task_id === taskId)
  }

  // 搜索关键词
  if (searchQuery) {
    const lowerQuery = searchQuery.toLowerCase()
    filtered = filtered.filter(m =>
      m.content.toLowerCase().includes(lowerQuery)
    )
  }

  console.log(`🔍 搜索消息: q="${searchQuery}", 找到 ${filtered.length} 条`)
  sendJson(res, 200, filtered)
}

// 更新消息状态（已读/未读）
function handleUpdateMessage(req, res, messageId) {
  let body = ''

  req.on('data', chunk => {
    body += chunk
  })

  req.on('end', () => {
    try {
      const updateData = JSON.parse(body)
      const messages = readMessages()
      const messageIndex = messages.findIndex(m => m.id === messageId)

      if (messageIndex === -1) {
        sendError(res, 404, '消息不存在')
        return
      }

      // 只允许更新 read 状态
      if (updateData.hasOwnProperty('read')) {
        messages[messageIndex].read = updateData.read
      }

      writeMessages(messages)

      console.log(`✅ 更新消息: ${messageId}, read=${messages[messageIndex].read}`)
      sendJson(res, 200, messages[messageIndex])
    } catch (error) {
      console.error('更新消息失败:', error)
      sendError(res, 400, '无效的请求数据')
    }
  })
}

// 导出消息
function handleExportMessages(res, query) {
  const format = query.format || 'markdown'
  const taskId = query.taskId

  const messages = readMessages()
  let filtered = messages

  // 按任务ID过滤
  if (taskId) {
    filtered = filtered.filter(m => m.task_id === taskId)
  }

  let contentType = 'application/json'
  let content = ''
  let filename = `export.${format}`

  try {
    if (format === 'markdown') {
      // Obsidian 友好的 Markdown 格式，带 YAML frontmatter
      const now = new Date().toISOString()
      content = `---
title: 聊天记录导出
source: TrueSignal
date: ${now}
tags: [chat, export]
---

# 聊天记录\n\n`

      filtered.forEach((msg, index) => {
        const timestamp = new Date(msg.timestamp).toLocaleString('zh-CN')
        const roleLabel = msg.role === 'user' ? '👤 用户' : '🤖 AI'
        content += `## ${index + 1}. ${roleLabel}\n\n`
        content += `**时间**: ${timestamp}\n\n`
        content += `${msg.content}\n\n`
        content += `---\n\n`
      })

      contentType = 'text/markdown; charset=utf-8'
      filename = 'chat-export.md'

    } else if (format === 'json') {
      content = JSON.stringify(filtered, null, 2)
      contentType = 'application/json; charset=utf-8'
      filename = 'chat-export.json'

    } else if (format === 'csv') {
      // CSV 格式
      content = '角色,时间,内容\n'
      filtered.forEach(msg => {
        const timestamp = new Date(msg.timestamp).toLocaleString('zh-CN')
        const role = msg.role === 'user' ? 'User' : 'AI'
        // 转义 CSV 中的特殊字符
        const escapedContent = `"${msg.content.replace(/"/g, '""')}"`
        content += `${role},"${timestamp}",${escapedContent}\n`
      })

      contentType = 'text/csv; charset=utf-8'
      filename = 'chat-export.csv'
    }

    console.log(`📥 导出消息: format=${format}, 导出 ${filtered.length} 条消息`)

    res.writeHead(200, {
      'Content-Type': contentType,
      'Content-Disposition': `attachment; filename="${filename}"`,
      'Access-Control-Allow-Origin': '*',
    })
    res.end(content)

  } catch (error) {
    console.error('导出消息失败:', error)
    sendError(res, 500, '导出失败')
  }
}

// 处理消息更新
function handleUpdateMessage(req, res, messageId) {
  // ... (existing code remains)
}

// 执行任务 (模拟 RSS 源同步)
async function handleExecuteTask(req, res, taskId) {
  const tasks = readTasks()
  const task = tasks.find(t => t.id === taskId)

  if (!task) {
    sendError(res, 404, '任务不存在')
    return
  }

  console.log(`🚀 开始执行任务: ${task.id} (${task.name})`)

  const executionId = generateId()
  const startTime = Date.now()

  try {
    // 模拟执行过程：随机生成成功/失败结果
    const success = Math.random() > 0.2  // 80% 成功率
    const duration = Math.random() * 3 + 1  // 1-4 秒

    // 模拟网络延迟
    await sleep(duration * 1000)

    const itemsCount = success ? Math.floor(Math.random() * 30) + 5 : 0  // 5-35 条或 0 条
    const endTime = Date.now()
    const actualDuration = (endTime - startTime) / 1000  // 转换为秒

    // 记录执行历史
    const executionRecord = {
      id: executionId,
      taskId: taskId,
      taskName: task.name,
      status: success ? 'success' : 'error',
      duration: Math.round(actualDuration * 100) / 100,  // 保留两位小数
      itemsCount: itemsCount,
      message: success
        ? `成功执行，获取了 ${itemsCount} 条新内容`
        : '执行失败，请检查 RSS 源状态',
      timestamp: new Date().toISOString(),
    }

    const history = readExecutionHistory()
    history.unshift(executionRecord)  // 最新的在前面
    writeExecutionHistory(history)

    console.log(`✅ 任务执行完成: ${taskId}, 状态=${success ? '成功' : '失败'}, 耗时=${actualDuration}s`)

    // 返回执行结果
    sendJson(res, 200, {
      executionId,
      taskId,
      status: success ? 'success' : 'error',
      duration: actualDuration,
      itemsCount: itemsCount,
      message: executionRecord.message,
      timestamp: new Date().toISOString(),
    })

  } catch (error) {
    console.error('任务执行出错:', error)
    sendError(res, 500, '任务执行失败')
  }
}

// 获取任务执行历史
function handleGetExecutionHistory(res, taskId, query) {
  const limit = parseInt(query.limit) || 20
  const offset = parseInt(query.offset) || 0

  const history = readExecutionHistory()
  const taskHistory = history.filter(h => h.taskId === taskId)

  // 分页
  const paged = taskHistory.slice(offset, offset + limit)

  console.log(`📋 获取执行历史: taskId=${taskId}, 返回 ${paged.length} 条`)
  sendJson(res, 200, paged)
}

// 根据不同消息返回不同的 AI 回复
function getAiResponse(message) {
  const responses = {
    '你好': '你好！👋 我是 TrueSignal AI 助手。很高兴认识你。我可以帮助你分析信息、生成总结或评估内容质量。有什么我可以帮助你的吗？',

    '帮助': '我可以为你提供以下帮助：\n\n**1. 信息分析** - 分析 RSS 源中的内容，识别关键信息和趋势。\n\n**2. 内容评估** - 根据创新度和深度评估文章的质量。\n\n**3. 自动总结** - 为长文章生成简洁的摘要，节省你的时间。\n\n**4. 多源聚合** - 从多个 RSS 源获取相关内容，避免信息重复。',

    default: `我已收到你的消息："${message}"。\n\n这是一条通过 SSE (Server-Sent Events) 实时流式传输的演示回复。我正在处理你的请求...\n\n### 功能演示\n\n- ✅ **实时流式** - 消息正在逐字发送\n- ✅ **Markdown 支持** - 完整的格式化文本\n- ✅ **代码示例** - 支持代码块\n- ✅ **表格显示** - 结构化数据\n\n### 代码示例\n\n\`\`\`javascript\n// 这是一个 JavaScript 代码示例\nconst message = "Hello TrueSignal"\nconsole.log(message)\n\`\`\`\n\n### 工作流程\n\n1. 你发送消息\n2. 服务器处理请求\n3. 流式发送响应\n4. 逐字显示文本\n5. 最后完成交互\n\n---\n\n这就是完整的 Phase 3 演示！所有数据都已持久化到本地 JSON 文件。`,
  }

  return responses[message] || responses.default
}

// 启动服务器
initializeDataFiles()

server.listen(PORT, () => {
  console.log(`\n${'='.repeat(70)}`)
  console.log(`🚀 TrueSignal Mock 后端服务器已启动`)
  console.log(`${'='.repeat(70)}`)
  console.log(`\n📍 服务器地址: http://localhost:${PORT}`)
  console.log(`📍 数据目录: ${DATA_DIR}`)
  console.log(`\n可用的 API 端点:`)
  console.log(`  GET    /api/tasks                    - 获取任务列表`)
  console.log(`  POST   /api/tasks                    - 创建任务`)
  console.log(`  GET    /api/tasks/:id                - 获取任务详情`)
  console.log(`  PUT    /api/tasks/:id                - 更新任务`)
  console.log(`  DELETE /api/tasks/:id                - 删除任务`)
  console.log(`  POST   /api/tasks/:id/execute        - 手动执行任务`)
  console.log(`  GET    /api/tasks/:id/execution-history - 获取执行历史`)
  console.log(`  GET    /api/tasks/:id/messages       - 获取消息历史`)
  console.log(`  POST   /api/messages                 - 保存消息`)
  console.log(`  GET    /api/messages/search          - 搜索消息`)
  console.log(`  GET    /api/messages/export          - 导出消息 (markdown|json|csv)`)
  console.log(`  PUT    /api/messages/:id             - 更新消息状态`)
  console.log(`  GET    /api/chat/stream              - SSE 流式聊天`)
  console.log(`\n前端配置:`)
  console.log(`  VITE_API_URL=http://localhost:${PORT}`)
  console.log(`\n运行前端:`)
  console.log(`  npm run dev`)
  console.log(`\n${'='.repeat(70)}\n`)
})

// 优雅关闭
process.on('SIGINT', () => {
  console.log('\n\n👋 Mock 服务器已关闭')
  process.exit(0)
})
