<template>
  <div
    :class="[
      'relative group w-full transition-all duration-500',
      compact ? 'max-w-xl' : 'max-w-2xl'
    ]"
  >
    <div
      :class="[
        'flex items-center w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-[#1F2937] transition-all duration-500',
        'focus-within:ring-2 focus-within:ring-gray-100 dark:focus-within:ring-gray-800 focus-within:border-gray-300 dark:focus-within:border-gray-600',
        compact
          ? 'rounded-lg shadow-sm p-1'
          : 'rounded-full shadow-xl shadow-gray-200/50 dark:shadow-none p-1.5'
      ]"
    >
      <!-- 平台选择按钮 -->
      <button
        @click="emit('togglePlatformMenu')"
        :class="[
          'flex items-center gap-2 px-5 py-3 text-sm font-medium text-gray-700 dark:text-gray-200',
          'bg-gray-50 hover:bg-gray-100 dark:bg-gray-800 dark:hover:bg-gray-700',
          'rounded-full transition-colors focus:outline-none whitespace-nowrap border border-transparent hover:border-gray-200 dark:hover:border-gray-600 h-full'
        ]"
        aria-label="选择数据源平台"
      >
        <span>{{ selectedPlatform }}</span>
        <span class="material-icons-outlined text-base transition-transform duration-300" :class="{ 'rotate-180': isPlatformMenuOpen }">
          expand_more
        </span>
      </button>

      <!-- 分割线 -->
      <div class="h-6 w-px bg-gray-200 dark:bg-gray-600 mx-2"></div>

      <!-- 搜索输入框 -->
      <div class="flex-1 flex items-center relative">
        <!-- 搜索图标 -->
        <div class="pl-2 flex items-center pointer-events-none">
          <span class="material-icons-outlined text-gray-400 dark:text-gray-500 text-lg">search</span>
        </div>

        <!-- 输入框 -->
        <input
          :value="modelValue"
          @input="$emit('update:modelValue', $event.target.value)"
          @keydown.enter="$emit('search', modelValue)"
          :class="[
            'w-full pl-3 pr-4 py-3 bg-transparent border-none text-gray-900 dark:text-white',
            'placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-0',
            compact ? 'text-sm' : 'text-base'
          ]"
          :placeholder="compact ? 'Search results...' : 'Search for keywords...'"
          aria-label="搜索输入框"
        />

        <!-- 清空按钮（仅在有内容且激活态显示） -->
        <button
          v-if="compact && modelValue"
          @click="$emit('update:modelValue', '')"
          class="mr-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors flex-shrink-0"
          aria-label="清空搜索"
        >
          <span class="material-icons-outlined text-lg">close</span>
        </button>
      </div>

      <!-- 搜索按钮（箭头） -->
      <button
        @click="$emit('search', modelValue)"
        :disabled="!modelValue || modelValue.toString().trim().length === 0"
        :class="[
          'p-3 text-white rounded-full transition-colors flex items-center justify-center shadow-md flex-shrink-0',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          'bg-gray-900 hover:bg-gray-800 dark:bg-gray-700 dark:hover:bg-gray-600'
        ]"
        aria-label="执行搜索"
      >
        <span class="material-icons-outlined text-xl">arrow_forward</span>
      </button>
    </div>

    <!-- 平台下拉菜单 -->
    <Transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0 scale-95 -translate-y-2"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition-all duration-300 ease-out"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 -translate-y-2"
    >
      <div
        v-if="isPlatformMenuOpen"
        class="absolute top-full left-0 mt-2 w-56 bg-white dark:bg-[#2D3748] rounded-lg border border-gray-200 dark:border-gray-600 shadow-lg z-50 overflow-hidden"
        role="listbox"
      >
        <button
          v-for="platform in platforms"
          :key="platform"
          @click="$emit('selectPlatform', platform)"
          :class="[
            'w-full px-4 py-3 hover:bg-gray-50 dark:hover:bg-[#3D4A5C] cursor-pointer transition-colors',
            'border-b border-gray-100 dark:border-gray-700 last:border-b-0',
            'text-left text-gray-900 dark:text-gray-100 hover:text-gray-900 dark:hover:text-white',
            'text-sm font-medium'
          ]"
          role="option"
          :aria-selected="selectedPlatform === platform"
        >
          {{ platform }}
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup>
defineProps({
  modelValue: {
    type: String,
    default: '',
    validator: (val) => typeof val === 'string'
  },
  isSearching: {
    type: Boolean,
    default: false
  },
  compact: {
    type: Boolean,
    default: false
  },
  isPlatformMenuOpen: {
    type: Boolean,
    default: false
  },
  selectedPlatform: {
    type: String,
    default: 'All Platforms'
  },
  platforms: {
    type: Array,
    default: () => ['All Platforms', 'Twitter', 'YouTube', 'RSS', 'Medium']
  }
})

const emit = defineEmits([
  'update:modelValue',
  'search',
  'togglePlatformMenu',
  'selectPlatform'
])
</script>

<style scoped>
.transition-all {
  transition: all 0.5s ease-in-out;
}
</style>

