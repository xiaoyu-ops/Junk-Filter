/**
 * useAPI Composable
 *
 * 提供统一的 API 调用接口
 * 处理所有的 HTTP 请求，包括：
 * - 任务 CRUD 操作（支持 Go 后端适配）
 * - 消息保存和查询（可独立配置）
 * - SSE 流式聊天（保留 Mock）
 * - 错误处理和重试
 *
 * 环境变量配置:
 * - VITE_API_URL: 业务 API 地址（Go 后端，默认 http://localhost:8080）
 * - VITE_MOCK_URL: Mock 后端地址（消息和聊天，默认 http://localhost:3000）
 */

import { ref } from 'vue'
import { useToast } from './useToast'

export const useAPI = () => {
  const { show: showToast } = useToast()

  // API 基础 URL（业务 API - Go 后端）
  const apiUrl = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080'

  // Mock API URL（消息和 SSE - Mock 后端）
  const mockUrl = import.meta.env.VITE_MOCK_URL || 'http://localhost:3000'

  // 是否正在加载
  const isLoading = ref(false)

  // ==================== 数据适配层 ====================

  /**
   * 将 Go 后端的 Source 对象适配为前端期待的 Task 对象
   *
   * @param {Object} source - Go 后端返回的 Source 对象
   * @returns {Object} 前端使用的 Task 对象
   *
   * 字段映射:
   * - id (int64) → id (string: "source-{id}")
   * - name (string) → name (string)
   * - url (string) → command (string) [RSS URL 作为命令]
   * - priority (int) → frequency (string) [优先级映射为频率]
   * - enabled (bool) → status (string: "active" | "paused")
   * - last_fetch_time (timestamp) → last_execution (timestamp)
   * - created_at (timestamp) → created_at (timestamp)
   */
  const adaptSourceToTask = (source) => {
    const priorityToFrequency = {
      10: 'hourly',
      8: 'hourly',
      6: 'daily',
      5: 'daily',
      3: 'weekly',
      1: 'weekly',
    }

    return {
      id: `source-${source.id}`,
      name: source.author_name || source.url,  // ✅ 使用 author_name 而不是 name
      command: source.url,
      frequency: priorityToFrequency[source.priority] || 'daily',
      status: source.enabled ? 'active' : 'paused',
      last_execution: source.last_fetch_time || source.created_at,
      created_at: source.created_at,
      updated_at: source.updated_at,

      // 保留原始数据以防需要
      _source: source,
    }
  }

  /**
   * 反向适配：将前端的 Task 对象转换回 Go 后端的 Source 格式
   * （用于创建/更新操作）
   */
  const adaptTaskToSource = (task) => {
    const frequencyToPriority = {
      'hourly': 8,
      'daily': 6,
      'weekly': 3,
      'once': 5,
    }

    return {
      name: task.name,
      url: task.command || task.name,
      priority: frequencyToPriority[task.frequency] || 5,
      enabled: task.status === 'active',
    }
  }

  /**
   * 发送 HTTP 请求的通用方法
   * @param {string} path - API 路径 (如 /api/tasks)
   * @param {object} options - 请求选项
   * @param {string} options.baseUrl - 可选的自定义基础 URL (默认 apiUrl)
   * @returns {promise} 返回响应数据
   */
  const request = async (path, options = {}) => {
    const baseUrl = options.baseUrl || apiUrl
    const url = `${baseUrl}${path}`
    const timeout = options.timeout || 30000

    try {
      isLoading.value = true

      // 创建超时 Promise
      const timeoutPromise = new Promise((_, reject) =>
        setTimeout(() => reject(new Error('请求超时')), timeout)
      )

      // 执行请求
      const fetchPromise = fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      })

      const response = await Promise.race([fetchPromise, timeoutPromise])

      // 检查响应状态
      if (!response.ok) {
        let errorMessage = `请求失败 (${response.status})`

        // 尝试获取详细错误信息
        try {
          const errorData = await response.json()
          errorMessage = errorData.error || errorMessage
        } catch (e) {
          // 忽略解析错误
        }

        throw new Error(errorMessage)
      }

      // 解析响应
      const data = await response.json()

      // 返回数据（支持 { data: ... } 和 { success: true, data: ... } 格式）
      return data.data !== undefined ? data.data : data

    } catch (error) {
      console.error('[API] 请求失败:', error)

      // 显示错误提示
      if (!(error instanceof Error && error.message === '已中止')) {
        showToast(error.message || '请求失败，请重试', 'error')
      }

      throw error

    } finally {
      isLoading.value = false
    }
  }

  /**
   * 重试请求（带指数退避）
   */
  const retryRequest = async (path, options = {}, maxRetries = 3) => {
    let lastError
    const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms))

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await request(path, options)
      } catch (error) {
        lastError = error
        console.warn(`[API] 第 ${attempt} 次尝试失败，将在 ${2 ** attempt * 100}ms 后重试`)

        if (attempt < maxRetries) {
          await delay(2 ** attempt * 100)  // 指数退避
        }
      }
    }

    throw lastError
  }

  // ==================== 任务相关 API ====================
  // 注意: tasks API 已适配为与 Go 后端的 /api/sources 兼容

  const tasks = {
    /**
     * 获取任务列表
     * 调用 Go 后端的 /api/sources，转换为前端的 Task 格式
     */
    list: async () => {
      const sources = await request('/api/sources', { baseUrl: apiUrl })
      return Array.isArray(sources) ? sources.map(adaptSourceToTask) : []
    },

    /**
     * 获取单个任务详情
     * 注意: 需要从 id 中提取原始 source id
     */
    get: async (id) => {
      const sourceId = id.startsWith('source-') ? id.replace('source-', '') : id
      const source = await request(`/api/sources/${sourceId}`, { baseUrl: apiUrl })
      return adaptSourceToTask(source)
    },

    /**
     * 创建任务
     * 转换前端的 Task 格式为 Go 后端的 Source 格式
     */
    create: async (data) => {
      const sourceData = adaptTaskToSource(data)
      const source = await request('/api/sources', {
        baseUrl: apiUrl,
        method: 'POST',
        body: JSON.stringify(sourceData),
      })
      return adaptSourceToTask(source)
    },

    /**
     * 更新任务
     * 转换前端的 Task 格式为 Go 后端的 Source 格式
     */
    update: async (id, data) => {
      const sourceId = id.startsWith('source-') ? id.replace('source-', '') : id
      const sourceData = adaptTaskToSource(data)
      const source = await request(`/api/sources/${sourceId}`, {
        baseUrl: apiUrl,
        method: 'PUT',
        body: JSON.stringify(sourceData),
      })
      return adaptSourceToTask(source)
    },

    /**
     * 删除任务
     * 注意: 需要从 id 中提取原始 source id
     */
    delete: async (id) => {
      const sourceId = id.startsWith('source-') ? id.replace('source-', '') : id
      await request(`/api/sources/${sourceId}`, {
        baseUrl: apiUrl,
        method: 'DELETE',
      })
    },

    /**
     * 手动执行任务 (触发 RSS 抓取)
     * 注意: 当前 Go 后端无此端点，此处保留接口
     * 如需实现，可添加 POST /api/sources/{id}/fetch 端点
     */
    execute: async (id) => {
      console.warn('[API] 手动执行任务功能需要 Go 后端支持，暂未实现')
      // 可选：降级到 Mock 后端
      // return request(`/api/tasks/${id}/execute`, {
      //   baseUrl: mockUrl,
      //   method: 'POST',
      // })
    },

    /**
     * 手动执行任务（通过 Mock 后端）
     * 模拟 RSS 源同步
     */
    executeTask: async (taskId) => {
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId
      return request(`/api/tasks/${actualTaskId}/execute`, {
        baseUrl: mockUrl,
        method: 'POST',
      })
    },

    /**
     * 获取任务执行历史
     * 获取特定任务的执行记录
     */
    getExecutionHistory: async (taskId, { limit = 20, offset = 0 } = {}) => {
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId
      return request(
        `/api/tasks/${actualTaskId}/execution-history?limit=${limit}&offset=${offset}`,
        { baseUrl: mockUrl }
      )
    },
  }

  // ==================== 消息相关 API ====================
  // 已迁移到真实 Go 后端 - 使用 /api/tasks/{id}/messages

  const messages = {
    /**
     * 获取任务的消息历史
     * 调用 Go 后端的 /api/tasks/{id}/messages
     */
    list: async (taskId, { limit = 50, offset = 0 } = {}) => {
      // taskId 可能是 "source-{id}" 格式，需要转换为原始 id 用于查询
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId
      try {
        return await request(
          `/api/tasks/${actualTaskId}/messages?limit=${limit}&offset=${offset}`,
          { baseUrl: apiUrl }
        )
      } catch (error) {
        // 如果消息接口失败，返回空数组而不中断
        console.warn('[API] 获取消息历史失败:', error)
        return []
      }
    },

    /**
     * 保存消息（用户或 AI 消息）
     * 调用 Go 后端的 /api/tasks/{id}/messages (POST)
     */
    save: async (taskId, messageData) => {
      // taskId 可能是 "source-{id}" 格式，需要转换为原始 id
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId

      // 确保必需字段存在
      const payload = {
        role: messageData.role || 'user',
        type: messageData.type || 'text',
        content: messageData.content || '',
        metadata: messageData.metadata || null,
        ...messageData,
      }

      try {
        return await request(`/api/tasks/${actualTaskId}/messages`, {
          baseUrl: apiUrl,
          method: 'POST',
          body: JSON.stringify(payload),
        })
      } catch (error) {
        console.error('[API] 保存消息失败:', error)
        throw error
      }
    },

    /**
     * 搜索消息
     * 支持按关键词搜索，可选按任务ID过滤
     */
    search: async (query, taskId = null) => {
      let searchUrl = `/api/messages/search?q=${encodeURIComponent(query)}`

      // 如果提供了 taskId，也进行过滤
      if (taskId) {
        const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId
        searchUrl += `&taskId=${encodeURIComponent(actualTaskId)}`
      }

      return request(searchUrl, { baseUrl: mockUrl })
    },

    /**
     * 更新消息状态（已读/未读）
     */
    updateStatus: async (messageId, status) => {
      return request(`/api/messages/${messageId}`, {
        baseUrl: mockUrl,
        method: 'PUT',
        body: JSON.stringify({ read: status }),
      })
    },

    /**
     * 导出消息
     * 支持 markdown、json、csv 格式
     * 返回 Blob 用于下载
     */
    export: async (format = 'markdown', taskId = null) => {
      let exportUrl = `/api/messages/export?format=${format}`

      if (taskId) {
        const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId
        exportUrl += `&taskId=${encodeURIComponent(actualTaskId)}`
      }

      try {
        isLoading.value = true

        const baseUrl = mockUrl
        const url = `${baseUrl}${exportUrl}`

        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Accept': 'application/octet-stream',
          },
        })

        if (!response.ok) {
          throw new Error(`导出失败 (${response.status})`)
        }

        // 获取文件名
        const contentDisposition = response.headers.get('content-disposition')
        let filename = `export.${format}`
        if (contentDisposition) {
          const matches = contentDisposition.match(/filename="(.+?)"/)
          if (matches) filename = matches[1]
        }

        // 返回 Blob 和文件名，由调用方处理下载
        const blob = await response.blob()
        return { blob, filename }

      } catch (error) {
        console.error('[API] 导出失败:', error)
        showToast(error.message || '导出失败，请重试', 'error')
        throw error
      } finally {
        isLoading.value = false
      }
    },
  }

  // ==================== 聊天相关 API ====================
  // 使用 SSE (Server-Sent Events) 进行流式聊天

  const chat = {
    /**
     * 启动 SSE 流式聊天（旧版本 - 用于内容评估）
     * 调用 Go 后端的 /api/chat/stream
     *
     * @deprecated 使用 chat.taskChat() 替代
     * @param {number} taskId - 任务 ID
     * @param {string} message - 用户消息
     * @param {function} onEvent - 事件回调函数
     * @param {object} configParams - 配置参数 (temperature, topP, maxTokens 等)
     * @returns {function} 取消连接的函数
     */
    stream: (taskId, message, onEvent, configParams = {}) => {
      // taskId 可能是 "source-{id}" 格式，需要转换为原始 id
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId

      // 构建查询参数
      const params = new URLSearchParams({
        taskId: actualTaskId,
        message: message,
      })

      // 添加配置参数（如果提供）
      if (configParams.temperature !== undefined) {
        params.append('temperature', configParams.temperature)
      }
      if (configParams.topP !== undefined) {
        params.append('topP', configParams.topP)
      }
      if (configParams.maxTokens !== undefined) {
        params.append('maxTokens', configParams.maxTokens)
      }

      const url = `${apiUrl}/api/chat/stream?${params.toString()}`

      try {
        const eventSource = new EventSource(url)

        // 处理消息事件
        eventSource.onmessage = (event) => {
          try {
            // 解析 SSE 数据（可能是 JSON）
            const data = JSON.parse(event.data)
            onEvent(data)

            // 如果收到 stream_end，说明流正常完成，关闭连接
            if (data.status === 'stream_end') {
              eventSource.close()
            }
          } catch (e) {
            // 如果不是 JSON，当作纯文本处理
            onEvent({ text: event.data })
          }
        }

        // 处理错误事件
        eventSource.onerror = (error) => {
          console.error('[API SSE] 连接错误:', error)
          eventSource.close()
          onEvent({ status: 'error', error: '连接中断' })
        }

        // 返回取消函数
        return () => {
          eventSource.close()
        }
      } catch (error) {
        console.error('[API SSE] 初始化失败:', error)
        onEvent({ status: 'error', error: error.message })
        return () => {} // 返回空函数
      }
    },

    /**
     * 任务特定的聊天 - Agent 调优与咨询（新版本）
     * 调用 Go 后端的 POST /api/tasks/{task_id}/chat
     *
     * @param {string} taskId - 任务 ID (可以是 "source-123" 或 "123" 格式)
     * @param {string} message - 用户的问题或指令
     * @param {function} onEvent - SSE 事件回调
     * @returns {function} 取消连接的函数
     */
    taskChat: (taskId, message, onEvent) => {
      // taskId 可能是 "source-{id}" 格式，需要转换为原始 id
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId

      const url = `${apiUrl}/api/tasks/${actualTaskId}/chat`

      try {
        // 发送 POST 请求
        const requestBody = JSON.stringify({
          message: message,
        })

        const response = fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: requestBody,
        })

        // 处理流式响应
        response.then(res => {
          if (!res.ok) {
            throw new Error(`HTTP ${res.status}`)
          }

          const reader = res.body.getReader()
          const decoder = new TextDecoder()
          let buffer = ''

          const processStream = async () => {
            try {
              while (true) {
                const { done, value } = await reader.read()
                if (done) break

                buffer += decoder.decode(value, { stream: true })

                // 按行处理 SSE 格式
                const lines = buffer.split('\n')
                buffer = lines.pop() // 保留未完成的行

                for (const line of lines) {
                  if (line.startsWith('data: ')) {
                    const data = line.slice(6) // 移除 "data: " 前缀
                    try {
                      const event = JSON.parse(data)
                      onEvent(event)
                    } catch (e) {
                      onEvent({ text: data })
                    }
                  }
                }
              }
            } catch (error) {
              console.error('[API Task Chat] Stream error:', error)
              onEvent({ status: 'error', error: error.message })
            }
          }

          processStream()
        }).catch(error => {
          console.error('[API Task Chat] Request error:', error)
          onEvent({ status: 'error', error: error.message })
        })

        // 返回取消函数
        return () => {
          // 无法取消 fetch，但返回空函数以兼容接口
        }
      } catch (error) {
        console.error('[API Task Chat] Initialization error:', error)
        onEvent({ status: 'error', error: error.message })
        return () => {}
      }
    },

    /**
     * 另一种 stream 实现，基于 fetch + ReadableStream (备用)
     * 某些浏览器或网络配置下 EventSource 可能不可用
     */
    streamWithFetch: async (taskId, message, onChunk, configParams = {}) => {
      const actualTaskId = taskId.startsWith('source-') ? taskId.replace('source-', '') : taskId

      const params = new URLSearchParams({
        taskId: actualTaskId,
        message: message,
      })

      if (configParams.temperature !== undefined) {
        params.append('temperature', configParams.temperature)
      }
      if (configParams.topP !== undefined) {
        params.append('topP', configParams.topP)
      }
      if (configParams.maxTokens !== undefined) {
        params.append('maxTokens', configParams.maxTokens)
      }

      const url = `${apiUrl}/api/chat/stream?${params.toString()}`

      try {
        const response = await fetch(url)
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`)
        }

        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        // 读取流
        while (true) {
          const { done, value } = await reader.read()

          if (done) break

          buffer += decoder.decode(value, { stream: true })

          // 按行处理（SSE 格式）
          const lines = buffer.split('\n')
          buffer = lines.pop() // 保留未完成的行

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              const data = line.slice(6) // 移除 "data: " 前缀
              try {
                onChunk(JSON.parse(data))
              } catch (e) {
                onChunk({ text: data })
              }
            }
          }
        }
      } catch (error) {
        console.error('[API SSE Fetch] 失败:', error)
        onChunk({ status: 'error', error: error.message })
      }
    },
  }

  // ==================== 认证相关 API ====================

  const auth = {
    /**
     * 用户登录（可选）
     */
    login: (email, password) => request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

    /**
     * 用户注册（可选）
     */
    register: (email, password, name) => request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    }),
  }

  return {
    // 状态
    isLoading,
    apiUrl,
    mockUrl,

    // 方法
    request,
    retryRequest,

    // 适配器
    adaptSourceToTask,
    adaptTaskToSource,

    // API 分组
    tasks,
    messages,
    chat,
    auth,
  }
}

/**
 * useAPI 使用示例
 *
 * const { tasks, messages, adaptSourceToTask } = useAPI()
 *
 * ==================== 任务操作 ====================
 *
 * // 获取任务列表（从 Go 后端 /api/sources，自动适配为 Task）
 * const taskList = await tasks.list()
 * // 返回: [
 * //   { id: 'source-1', name: '新闻源', command: 'https://...', frequency: 'daily', status: 'active', ... }
 * // ]
 *
 * // 创建任务（转换为 Source 格式发送给 Go 后端）
 * const newTask = await tasks.create({
 *   name: '科技新闻',
 *   command: 'https://techcrunch.com/feed/',
 *   frequency: 'daily',
 *   status: 'active'
 * })
 *
 * // 更新任务
 * const updated = await tasks.update('source-1', {
 *   name: '更新的名称',
 *   frequency: 'hourly'
 * })
 *
 * // 删除任务
 * await tasks.delete('source-1')
 *
 * ==================== 消息操作 ====================
 *
 * // 获取任务的消息历史（从 Mock 后端）
 * const taskMessages = await messages.list('source-1', { limit: 50 })
 *
 * // 保存消息（到 Mock 后端）
 * await messages.save({
 *   task_id: 'source-1',
 *   role: 'user',
 *   type: 'text',
 *   content: '用户消息内容'
 * })
 *
 * ==================== 适配器使用 ====================
 *
 * // 手动适配 Source 对象为 Task 对象
 * const taskObj = adaptSourceToTask(sourceFromGoBackend)
 *
 * // 手动适配 Task 对象为 Source 对象
 * const sourceObj = adaptTaskToSource(taskFromFrontend)
 *
 * ==================== 环境变量配置 ====================
 *
 * .env 文件示例:
 * VITE_API_URL=http://localhost:8080          # Go 后端（业务 API）
 * VITE_MOCK_URL=http://localhost:3000         # Mock 后端（消息和 SSE）
 *
 * ==================== 关键注意事项 ====================
 *
 * 1. Task ID 格式: "source-{goSourceId}"
 *    - 前端使用 "source-1" 格式的 ID
 *    - 调用 Go 后端 API 时自动提取数字 ID
 *
 * 2. 消息 API 暂时使用 Mock 后端
 *    - 后续 Go 后端实现消息存储后可切换
 *    - 只需修改 messages.list() 和 messages.save() 中的 baseUrl
 *
 * 3. SSE 聊天端点
 *    - 仍由 Mock 后端提供 (/api/chat/stream)
 *    - useSSE.js 中的 connectSSE() 会自动使用 VITE_MOCK_URL
 *    - 无需修改 TaskChat.vue 组件
 *
 * 4. 字段映射
 *    Source (Go) ←→ Task (前端):
 *    - id (int) ↔ id ("source-{id}" string)
 *    - name ↔ name
 *    - url ↔ command
 *    - priority (1-10) ↔ frequency (hourly/daily/weekly)
 *    - enabled (bool) ↔ status ("active"/"paused")
 *    - last_fetch_time ↔ last_execution
 */

