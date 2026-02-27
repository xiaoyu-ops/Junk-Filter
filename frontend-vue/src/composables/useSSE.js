/**
 * useSSE Composable
 * 提供 Server-Sent Events (SSE) 流式消息处理
 * 用于实时接收 AI 流式回复
 */

import { ref, computed } from 'vue'

export const useSSE = () => {
  // EventSource 连接实例
  let eventSource = null

  // 是否正在连接
  const isConnecting = ref(false)

  // 当前流式文本（逐字积累）
  const streamingText = ref('')

  // 错误信息
  const error = ref(null)

  // 连接状态
  const connectionState = ref('disconnected') // disconnected | connecting | connected | error

  /**
   * 开始 SSE 连接并流式接收消息
   * @param {string} url - SSE 端点 URL
   * @param {object} options - 配置选项
   * @param {object} options.headers - 请求头（如 Authorization）
   * @param {function} options.onMessage - 每次收到消息时的回调
   * @param {function} options.onStreamingText - 逐字累积时的回调
   * @param {function} options.onComplete - 流完成时的回调
   * @param {function} options.onError - 错误时的回调
   * @returns {promise} 返回完整的响应文本
   */
  const connectSSE = (url, options = {}) => {
    return new Promise((resolve, reject) => {
      try {
        // 重置状态
        streamingText.value = ''
        error.value = null
        isConnecting.value = true
        connectionState.value = 'connecting'

        // ⚠️ 新增：标记流是否已结束（防止重复处理事件）
        let isStreamEnded = false

        // 验证 URL
        if (!url) {
          throw new Error('SSE 端点 URL 不能为空')
        }

        // 创建 EventSource 连接
        eventSource = new EventSource(url)

        /**
         * 连接打开事件
         */
        eventSource.onopen = () => {
          console.log('[SSE] 连接已建立')
          connectionState.value = 'connected'
          isConnecting.value = false
        }

        /**
         * 接收消息事件
         * 后端可以发送不同类型的事件：
         * - 'message' - 纯文本消息
         * - 'delta' - Markdown 增量（推荐用于流式文本）
         * - 'execution' - 执行卡片数据
         * - 'done' - 流完成
         * - 'error' - 错误信息
         */
        eventSource.addEventListener('message', (event) => {
          try {
            const data = JSON.parse(event.data)

            // 处理 Markdown 增量文本（逐字显示）
            if (data.type === 'delta' && data.content) {
              streamingText.value += data.content

              // 触发逐字更新回调
              if (options.onStreamingText) {
                options.onStreamingText(streamingText.value)
              }
            }

            // 处理完整消息
            if (data.type === 'message') {
              streamingText.value = data.content

              if (options.onStreamingText) {
                options.onStreamingText(streamingText.value)
              }
            }

            // 处理执行卡片数据
            if (data.type === 'execution' && options.onMessage) {
              options.onMessage({
                type: 'execution',
                data: data,
              })
            }

            // 处理流完成事件
            if (data.type === 'done') {
              closeSSE()
              connectionState.value = 'connected'

              if (options.onComplete) {
                options.onComplete(streamingText.value)
              }

              resolve(streamingText.value)
            }
          } catch (parseError) {
            console.error('[SSE] 消息解析失败:', parseError)
          }
        })

        /**
         * 错误事件 - 关键修复
         */
        eventSource.onerror = (event) => {
          console.error('[SSE] 连接错误:', event)

          // ⚠️ 修复：检查是否已有部分数据
          const hadData = streamingText.value.length > 0

          connectionState.value = 'error'
          error.value = '连接错误：无法接收流式数据'
          isStreamEnded = true
          closeSSE()

          // ⚠️ 修复：有数据时当做成功，无数据时才报错
          if (!hadData) {
            // 完全没数据，这是真正的错误
            if (options.onError) {
              options.onError(error.value)
            }
            reject(new Error(error.value))
          } else {
            // 有部分数据，视为成功（网络中断但已传输部分消息）
            console.warn('[SSE] 网络中断但已接收部分数据，视为成功')
            if (options.onComplete) {
              options.onComplete(streamingText.value)
            }
            resolve(streamingText.value)
          }
        }

        /**
         * 自定义 delta 事件监听（用于逐字显示）
         */
        eventSource.addEventListener('delta', (event) => {
          // ⚠️ 修复：检查流是否已结束
          if (isStreamEnded) {
            console.warn('[SSE] 流已结束，忽略后续 delta 事件')
            return
          }

          try {
            const data = JSON.parse(event.data)

            if (data.content) {
              streamingText.value += data.content

              if (options.onStreamingText) {
                options.onStreamingText(streamingText.value)
              }
            }
          } catch (parseError) {
            console.error('[SSE] Delta 事件解析失败:', parseError)
          }
        })

        /**
         * 自定义 execution 事件监听
         */
        eventSource.addEventListener('execution', (event) => {
          // ⚠️ 修复：检查流是否已结束
          if (isStreamEnded) {
            console.warn('[SSE] 流已结束，忽略后续 execution 事件')
            return
          }

          try {
            const data = JSON.parse(event.data)

            if (options.onMessage) {
              options.onMessage({
                type: 'execution',
                data: data,
              })
            }
          } catch (parseError) {
            console.error('[SSE] Execution 事件解析失败:', parseError)
          }
        })

        /**
         * 自定义 done 事件监听
         */
        eventSource.addEventListener('done', (event) => {
          // ⚠️ 修复：防止重复处理 done 事件
          if (isStreamEnded) {
            console.warn('[SSE] 流已结束，忽略重复的 done 事件')
            return
          }

          try {
            isStreamEnded = true
            const data = JSON.parse(event.data)

            closeSSE()
            connectionState.value = 'connected'

            if (options.onComplete) {
              options.onComplete(streamingText.value, data)
            }

            resolve(streamingText.value)
          } catch (parseError) {
            console.error('[SSE] Done 事件解析失败:', parseError)
            resolve(streamingText.value)
          }
        })

        /**
         * 自定义 error 事件监听
         */
        eventSource.addEventListener('error', (event) => {
          // ⚠️ 修复：检查流是否已结束
          if (isStreamEnded) {
            console.warn('[SSE] 流已结束，忽略后续 error 事件')
            return
          }

          try {
            isStreamEnded = true
            const data = JSON.parse(event.data)
            error.value = data.message || '未知错误'

            closeSSE()
            connectionState.value = 'error'

            if (options.onError) {
              options.onError(error.value)
            }

            reject(new Error(error.value))
          } catch (parseError) {
            console.error('[SSE] Error 事件解析失败:', parseError)
            reject(parseError)
          }
        })
      } catch (err) {
        console.error('[SSE] 连接失败:', err)
        connectionState.value = 'error'
        error.value = err.message
        reject(err)
      }
    })
  }

  /**
   * 关闭 SSE 连接
   */
  const closeSSE = () => {
    if (eventSource) {
      eventSource.close()
      eventSource = null
      connectionState.value = 'disconnected'
      console.log('[SSE] 连接已关闭')
    }
  }

  /**
   * 重试连接（用于错误恢复）
   * @param {string} url - SSE 端点 URL
   * @param {number} maxRetries - 最大重试次数
   * @param {number} retryDelay - 重试延迟（毫秒）
   */
  const retryConnectSSE = async (url, options = {}) => {
    const maxRetries = options.maxRetries || 3
    const retryDelay = options.retryDelay || 1000
    let retries = 0

    while (retries < maxRetries) {
      try {
        return await connectSSE(url, options)
      } catch (err) {
        retries++
        console.warn(`[SSE] 重试 ${retries}/${maxRetries}，延迟 ${retryDelay}ms...`)

        if (retries >= maxRetries) {
          throw new Error(`[SSE] 连接失败，已尝试 ${maxRetries} 次`)
        }

        // 等待后重试
        await new Promise(resolve => setTimeout(resolve, retryDelay))
      }
    }
  }

  /**
   * 获取当前连接状态文本
   */
  const statusText = computed(() => {
    const statusMap = {
      disconnected: '未连接',
      connecting: '连接中...',
      connected: '已连接',
      error: '连接错误',
    }
    return statusMap[connectionState.value] || '未知'
  })

  /**
   * 是否已连接
   */
  const isConnected = computed(() => connectionState.value === 'connected')

  return {
    // 状态
    streamingText,
    error,
    connectionState,
    isConnecting,
    statusText,
    isConnected,

    // 方法
    connectSSE,
    closeSSE,
    retryConnectSSE,
  }
}

/**
 * useSSE 使用示例：
 *
 * const { streamingText, connectSSE, closeSSE } = useSSE()
 *
 * // 开始流式接收
 * await connectSSE('/api/chat/stream', {
 *   onStreamingText: (text) => {
 *     console.log('逐字累积:', text)
 *   },
 *   onComplete: (finalText) => {
 *     console.log('完成:', finalText)
 *   },
 *   onError: (err) => {
 *     console.error('错误:', err)
 *   },
 * })
 *
 * // 后端返回数据格式：
 * // 事件: delta
 * // data: { "type": "delta", "content": "你好" }
 * // data: { "type": "delta", "content": "，我是" }
 * // data: { "type": "delta", "content": "AI助手" }
 * //
 * // 事件: execution
 * // data: { "type": "execution", "status": "success", "itemCount": 24 }
 * //
 * // 事件: done
 * // data: { "type": "done" }
 */
