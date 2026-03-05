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
      <!-- Search Input -->
      <div class="flex-1 flex items-center relative">
        <div class="pl-4 flex items-center pointer-events-none">
          <span class="material-icons-outlined text-gray-400 dark:text-gray-500 text-lg">search</span>
        </div>

        <input
          :value="modelValue"
          @input="$emit('update:modelValue', $event.target.value)"
          @keydown.enter="$emit('search', modelValue)"
          :class="[
            'w-full pl-3 pr-4 py-3 bg-transparent border-none text-gray-900 dark:text-white',
            'placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-0',
            compact ? 'text-sm' : 'text-base'
          ]"
          :placeholder="compact ? 'Search articles...' : 'Search articles by title, content, or author...'"
          aria-label="Search articles"
        />

        <!-- Clear button -->
        <button
          v-if="compact && modelValue"
          @click="$emit('update:modelValue', '')"
          class="mr-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors flex-shrink-0"
          aria-label="Clear search"
        >
          <span class="material-icons-outlined text-lg">close</span>
        </button>
      </div>

      <!-- Search Button -->
      <button
        @click="$emit('search', modelValue)"
        :disabled="!modelValue || modelValue.toString().trim().length === 0"
        :class="[
          'p-3 text-white rounded-full transition-colors flex items-center justify-center shadow-md flex-shrink-0',
          'disabled:opacity-50 disabled:cursor-not-allowed',
          'bg-gray-900 hover:bg-gray-800 dark:bg-gray-700 dark:hover:bg-gray-600'
        ]"
        aria-label="Search"
      >
        <span class="material-icons-outlined text-xl">arrow_forward</span>
      </button>
    </div>
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
})

const emit = defineEmits([
  'update:modelValue',
  'search',
])
</script>
