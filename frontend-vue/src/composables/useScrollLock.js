import { ref, nextTick } from 'vue'

/**
 * useScrollLock Composable
 * 提供智能滚动锁定功能，用于消息列表的自动滚动管理
 */
export const useScrollLock = () => {
  // 消息容器的 ref
  const containerRef = ref(null)

  // 追踪用户是否在容器底部附近
  const isUserNearBottom = ref(true)

  // 距离底部的阈值（单位：像素）
  const SCROLL_THRESHOLD = 50

  /**
   * 检查用户是否接近容器底部
   * 计算逻辑：scrollHeight - scrollTop - clientHeight < SCROLL_THRESHOLD
   * @returns {boolean} 用户是否在底部
   */
  const checkIfNearBottom = () => {
    if (!containerRef.value) return false

    const { scrollTop, scrollHeight, clientHeight } = containerRef.value
    const distanceFromBottom = scrollHeight - scrollTop - clientHeight

    // 如果距离底部小于阈值，认为用户在底部
    isUserNearBottom.value = distanceFromBottom < SCROLL_THRESHOLD

    return isUserNearBottom.value
  }

  /**
   * 设置滚动事件监听器
   * 在组件挂载时调用此方法
   */
  const setupScrollListener = () => {
    if (!containerRef.value) return

    // 使用 passive 选项提升滚动性能
    containerRef.value.addEventListener('scroll', checkIfNearBottom, {
      passive: true,
    })
  }

  /**
   * 移除滚动事件监听器
   * 在组件卸载时调用此方法
   */
  const removeScrollListener = () => {
    if (!containerRef.value) return

    containerRef.value.removeEventListener('scroll', checkIfNearBottom)
  }

  /**
   * 智能滚动到底部
   * 仅当用户当前在底部附近时，才自动滚动
   * 如果用户正在往上翻看历史，则不强制跳转
   */
  const autoScrollToBottom = async () => {
    // 等待 DOM 更新完成
    await nextTick()

    // 只有当用户在底部时才自动滚动
    if (isUserNearBottom.value && containerRef.value) {
      // 使用平滑滚动（如果浏览器支持）
      containerRef.value.scrollTo({
        top: containerRef.value.scrollHeight,
        behavior: 'smooth',  // 平滑滚动效果
      })

      // 同步更新状态
      checkIfNearBottom()
    }
  }

  /**
   * 强制滚到底部
   * 用于组件初始化或特殊场景（如首次加载消息）
   */
  const scrollToBottom = () => {
    if (containerRef.value) {
      containerRef.value.scrollTop = containerRef.value.scrollHeight

      // 更新状态
      nextTick(() => {
        checkIfNearBottom()
      })
    }
  }

  /**
   * 滚到顶部
   * 用于加载历史消息时的定位
   */
  const scrollToTop = () => {
    if (containerRef.value) {
      containerRef.value.scrollTop = 0
      checkIfNearBottom()
    }
  }

  /**
   * 手动设置用户是否在底部的状态
   * 用于特殊场景（如切换任务时重置状态）
   */
  const setNearBottom = (value) => {
    isUserNearBottom.value = value
  }

  return {
    // refs
    containerRef,
    isUserNearBottom,

    // methods
    checkIfNearBottom,
    setupScrollListener,
    removeScrollListener,
    autoScrollToBottom,
    scrollToBottom,
    scrollToTop,
    setNearBottom,
  }
}
