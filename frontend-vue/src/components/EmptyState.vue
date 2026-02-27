<template>
  <div class="flex flex-col items-center justify-center py-12 px-4">
    <!-- Icon -->
    <span
      v-if="icon"
      class="material-icons-outlined text-6xl text-gray-300 dark:text-gray-700 mb-4"
    >
      {{ icon }}
    </span>

    <!-- Custom Icon Slot -->
    <slot v-else name="icon"></slot>

    <!-- Title -->
    <h3
      v-if="title"
      class="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2"
    >
      {{ title }}
    </h3>

    <!-- Subtitle -->
    <p
      v-if="subtitle"
      class="text-sm text-gray-500 dark:text-gray-400 text-center mb-6 max-w-sm"
    >
      {{ subtitle }}
    </p>

    <!-- Action Slot (for custom actions) -->
    <slot name="actions"></slot>

    <!-- Default CTA Button -->
    <button
      v-if="action && !$slots.actions"
      @click="emit('action')"
      class="flex items-center gap-2 px-4 py-2.5 bg-primary hover:bg-primary-dark dark:bg-blue-600 dark:hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm"
    >
      <span
        v-if="actionIcon"
        class="material-icons-outlined text-lg"
      >
        {{ actionIcon }}
      </span>
      <span>{{ action }}</span>
    </button>
  </div>
</template>

<script setup>
/**
 * EmptyState Component
 *
 * Reusable component for displaying empty state messages
 * with optional icon, title, subtitle, and call-to-action.
 *
 * @component
 * @example
 * // Show empty task list with create button
 * <EmptyState
 *   icon="inbox"
 *   title="还没有任务"
 *   subtitle="从左边的菜单开始创建任务"
 *   action="创建任务"
 *   actionIcon="add_circle"
 *   @action="openTaskModal"
 * />
 *
 * @example
 * // Show search results empty state
 * <EmptyState
 *   icon="search_off"
 *   title="未找到匹配消息"
 *   :subtitle="`没有消息包含 \"${searchQuery}\"`"
 * >
 *   <template #actions>
 *     <button @click="clearSearch" class="...">清空搜索</button>
 *   </template>
 * </EmptyState>
 */

const props = defineProps({
  /**
   * Material icon name to display
   * @type {String}
   */
  icon: {
    type: String,
    default: null
  },

  /**
   * Title text
   * @type {String}
   */
  title: {
    type: String,
    default: null
  },

  /**
   * Subtitle/description text
   * @type {String}
   */
  subtitle: {
    type: String,
    default: null
  },

  /**
   * CTA button text (if no actions slot provided)
   * @type {String}
   */
  action: {
    type: String,
    default: null
  },

  /**
   * Icon for CTA button
   * @type {String}
   */
  actionIcon: {
    type: String,
    default: null
  }
})

const emit = defineEmits(['action'])
</script>

<style scoped>
/* Empty state uses standard Tailwind classes with dark mode variants */
</style>
