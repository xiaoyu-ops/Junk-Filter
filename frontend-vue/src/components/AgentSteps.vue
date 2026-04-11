<template>
  <div class="mb-3 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700/50 bg-gray-50 dark:bg-[#1a1f2e] text-xs">

    <!-- 标题栏（点击折叠/展开） -->
    <button
      @click="expanded = !expanded"
      class="w-full flex items-center gap-2 px-3 py-2.5 hover:bg-gray-100/80 dark:hover:bg-white/5 transition-colors text-left"
    >
      <!-- 执行中 spinner / 完成 checkmark -->
      <span
        v-if="processing"
        class="w-3.5 h-3.5 rounded-full border-2 border-indigo-400 border-t-transparent animate-spin flex-shrink-0"
      ></span>
      <span
        v-else
        class="material-icons-outlined flex-shrink-0 text-indigo-500 dark:text-indigo-400"
        style="font-size: 14px;"
      >check_circle</span>

      <!-- 标题文字 -->
      <span class="flex-1 font-medium text-gray-600 dark:text-gray-300">
        {{ processing ? '执行中...' : `已执行 ${toolCalls.length} 步` }}
      </span>

      <!-- 折叠箭头 -->
      <span
        class="material-icons-outlined text-gray-400 dark:text-gray-500 transition-transform duration-200"
        :class="expanded ? 'rotate-180' : 'rotate-0'"
        style="font-size: 14px;"
      >expand_more</span>
    </button>

    <!-- 步骤列表（可折叠） -->
    <transition name="steps">
      <div v-show="expanded" class="border-t border-gray-200 dark:border-gray-700/50 px-3 py-2 space-y-2">
        <div
          v-for="(tc, i) in toolCalls"
          :key="i"
          class="flex items-start gap-2"
        >
          <!-- 步骤序号 -->
          <span class="mt-0.5 w-4 h-4 rounded-full bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400 flex items-center justify-center flex-shrink-0 font-bold" style="font-size: 10px;">
            {{ i + 1 }}
          </span>

          <!-- 工具名 + 结果摘要 -->
          <div class="flex-1 min-w-0 leading-relaxed">
            <span class="font-mono text-indigo-600 dark:text-indigo-400">{{ tc.tool }}</span>
            <span class="text-gray-400 dark:text-gray-500 mx-1.5">→</span>
            <span class="text-gray-600 dark:text-gray-300">{{ summarize(tc) }}</span>
          </div>
        </div>

        <!-- 当前仍在执行时的等待行 -->
        <div v-if="processing" class="flex items-center gap-2 text-gray-400 dark:text-gray-500">
          <span class="w-4 h-4 flex-shrink-0"></span>
          <span class="animate-pulse">思考中...</span>
        </div>
      </div>
    </transition>

  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  toolCalls: { type: Array, default: () => [] },
  processing: { type: Boolean, default: false },
})

// 默认展开；完成后 1.5s 自动折叠
const expanded = ref(true)

watch(() => props.processing, (isProcessing) => {
  if (!isProcessing && props.toolCalls.length > 0) {
    setTimeout(() => { expanded.value = false }, 1500)
  }
})

const summarize = (tc) => {
  const r = tc.result || {}
  if (r.error)              return `失败: ${r.error}`
  if (r.message)            return r.message
  if (r.count !== undefined) return `返回 ${r.count} 篇文章`
  if (r.pending !== undefined) return `待评估 ${r.pending} | 已评估 ${r.evaluated}`
  if (r.success === true)   return '成功'
  return JSON.stringify(r).slice(0, 80)
}
</script>

<style scoped>
.steps-enter-active,
.steps-leave-active {
  transition: opacity 0.2s ease, max-height 0.25s ease;
  overflow: hidden;
  max-height: 300px;
}
.steps-enter-from,
.steps-leave-to {
  opacity: 0;
  max-height: 0;
}
</style>
