<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="opacity-0 translate-y-2"
    leave-to-class="opacity-0 translate-y-2"
    leave-active-class="transition-all duration-200 ease-in"
  >
    <div
      v-if="isVisible"
      class="rounded-lg border-l-4 p-4 bg-red-50 dark:bg-red-900/20 border-red-500 dark:border-red-400"
    >
      <!-- Error Header -->
      <div class="flex items-start justify-between gap-3">
        <div class="flex items-start gap-3 flex-1">
          <!-- Error Icon -->
          <span
            class="material-icons-outlined text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5 text-lg"
          >
            {{ errorIcon }}
          </span>

          <!-- Error Content -->
          <div class="flex-1 min-w-0">
            <!-- Error Title -->
            <h3 class="font-medium text-red-900 dark:text-red-300">
              {{ errorTitle }}
            </h3>

            <!-- Error Message -->
            <p class="text-sm text-red-800 dark:text-red-400 mt-1">
              {{ errorMessage }}
            </p>

            <!-- Error Details (if provided) -->
            <details v-if="errorDetails" class="mt-2">
              <summary class="cursor-pointer text-xs text-red-700 dark:text-red-500 hover:underline">
                详细信息
              </summary>
              <pre class="mt-2 text-xs overflow-auto max-h-48 p-2 bg-white dark:bg-gray-800 rounded border border-red-200 dark:border-red-800 text-red-900 dark:text-red-300">{{ errorDetails }}</pre>
            </details>
          </div>
        </div>

        <!-- Close Button -->
        <button
          v-if="dismissible"
          @click="handleDismiss"
          class="text-red-400 hover:text-red-600 dark:hover:text-red-300 transition-colors flex-shrink-0"
        >
          <span class="material-icons-outlined text-lg">close</span>
        </button>
      </div>

      <!-- Error Actions -->
      <div class="flex gap-2 mt-3 flex-wrap">
        <!-- Retry Button -->
        <button
          v-if="showRetry"
          @click="handleRetry"
          :disabled="isRetrying"
          class="flex items-center gap-1.5 px-3 py-1.5 bg-red-600 hover:bg-red-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white text-sm font-medium rounded transition-colors"
        >
          <span
            class="material-icons-outlined text-sm"
            :class="{ 'animate-spin': isRetrying }"
          >
            {{ isRetrying ? 'hourglass_top' : 'refresh' }}
          </span>
          <span>{{ isRetrying ? '重试中...' : '重试' }}</span>
        </button>

        <!-- Report Button (optional) -->
        <button
          v-if="showReport"
          @click="handleReport"
          class="flex items-center gap-1.5 px-3 py-1.5 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 text-sm font-medium rounded transition-colors"
        >
          <span class="material-icons-outlined text-sm">bug_report</span>
          <span>报告</span>
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, computed } from 'vue'

/**
 * ErrorCard Component
 *
 * Reusable error display component for showing API failures,
 * network errors, validation errors, etc.
 *
 * @component
 * @example
 * // Show network error with retry
 * <ErrorCard
 *   :error="{ type: 'network', message: '网络连接失败' }"
 *   @retry="handleRetry"
 * />
 *
 * @example
 * // Show API error with details
 * <ErrorCard
 *   type="api"
 *   title="请求失败"
 *   message="服务器返回错误"
 *   :errorDetails="errorResponse"
 *   @retry="refetchData"
 *   dismissible
 * />
 */

const props = defineProps({
  /**
   * Error object or type string
   * Types: 'network', 'api', 'timeout', 'validation', 'unknown'
   * @type {Object|String}
   */
  error: {
    type: [Object, String],
    required: true
  },

  /**
   * Custom error title (overrides default based on error type)
   * @type {String}
   */
  title: {
    type: String,
    default: null
  },

  /**
   * Custom error message (overrides default)
   * @type {String}
   */
  message: {
    type: String,
    default: null
  },

  /**
   * Detailed error information (shown in expandable section)
   * @type {String|Object}
   */
  errorDetails: {
    type: [String, Object],
    default: null
  },

  /**
   * Show retry button
   * @type {Boolean}
   */
  showRetry: {
    type: Boolean,
    default: true
  },

  /**
   * Show report button
   * @type {Boolean}
   */
  showReport: {
    type: Boolean,
    default: false
  },

  /**
   * Allow dismissing the error card
   * @type {Boolean}
   */
  dismissible: {
    type: Boolean,
    default: true
  },

  /**
   * Auto-dismiss after N milliseconds (0 = no auto-dismiss)
   * @type {Number}
   */
  autoDismissMs: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['retry', 'report', 'dismiss'])

const isVisible = ref(true)
const isRetrying = ref(false)

/**
 * Determine error type and icon
 */
const errorType = computed(() => {
  if (typeof props.error === 'string') {
    return props.error
  }
  return props.error?.type || props.error?.code || 'unknown'
})

/**
 * Get error icon based on type
 */
const errorIcon = computed(() => {
  const icons = {
    network: 'cloud_off',
    api: 'error_outline',
    timeout: 'schedule',
    validation: 'warning',
    unknown: 'info'
  }
  return icons[errorType.value] || 'error_outline'
})

/**
 * Get error title based on type
 */
const errorTitle = computed(() => {
  if (props.title) return props.title

  const titles = {
    network: '网络连接失败',
    api: '请求失败',
    timeout: '请求超时',
    validation: '验证失败',
    unknown: '出错了'
  }
  return titles[errorType.value] || '出错了'
})

/**
 * Get error message
 */
const errorMessage = computed(() => {
  if (props.message) return props.message

  if (typeof props.error === 'string') {
    return props.error
  }

  return props.error?.message || '请稍后重试'
})

/**
 * Handle retry button click
 */
const handleRetry = async () => {
  isRetrying.value = true
  try {
    emit('retry')
  } finally {
    isRetrying.value = false
  }
}

/**
 * Handle report button click
 */
const handleReport = () => {
  emit('report', {
    type: errorType.value,
    title: errorTitle.value,
    message: errorMessage.value,
    details: props.errorDetails
  })
}

/**
 * Handle dismiss button click
 */
const handleDismiss = () => {
  isVisible.value = false
  emit('dismiss')
}

/**
 * Auto-dismiss after specified time
 */
if (props.autoDismissMs > 0) {
  setTimeout(() => {
    handleDismiss()
  }, props.autoDismissMs)
}
</script>

<style scoped>
/* Error card uses Transition for smooth slide-down animation */
details {
  user-select: none;
}

details[open] {
  user-select: auto;
}

pre {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
