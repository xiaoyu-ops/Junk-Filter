<template>
  <div class="space-y-3">
    <div
      v-for="n in count"
      :key="n"
      class="rounded-lg bg-gray-200 dark:bg-gray-700 animate-pulse"
      :style="{ height: heightClass }"
    ></div>
  </div>
</template>

<script setup>
/**
 * SkeletonLoader Component
 *
 * Reusable skeleton loader component for displaying loading state
 * while data is being fetched.
 *
 * @component
 * @example
 * // Show 4 skeleton items with 64px height
 * <SkeletonLoader :count="4" height="64px" />
 *
 * @example
 * // Show 5 items with 80px height and rounded corners
 * <SkeletonLoader :count="5" height="80px" rounded />
 */

const props = defineProps({
  /**
   * Number of skeleton items to display
   * @type {Number}
   */
  count: {
    type: Number,
    default: 3
  },

  /**
   * Height of each skeleton item (CSS height value)
   * Can be: '64px', '4rem', '100%', etc.
   * @type {String}
   */
  height: {
    type: String,
    default: '64px'
  },

  /**
   * Additional CSS classes for customization
   * @type {String}
   */
  class: {
    type: String,
    default: ''
  }
})

const heightClass = props.height
</script>

<style scoped>
/* Skeleton pulse animation uses Tailwind's animate-pulse
   which creates a smooth 2s ease-in-out opacity loop */

/* Optional: Custom spacing between skeleton items */
:deep(.space-y-3) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0.75rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0.75rem * var(--tw-space-y-reverse));
}

:deep(.space-y-3 > :not([hidden]) ~ :not([hidden])) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0.75rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0.75rem * var(--tw-space-y-reverse));
}
</style>
