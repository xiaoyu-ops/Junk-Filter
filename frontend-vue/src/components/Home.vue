<template>
  <main class="w-full min-h-screen flex flex-col bg-surface-light dark:bg-[#0f0f11] overflow-y-auto">
    <!-- 初始态：搜索栏居中 -->
    <div
      v-if="!isSearching"
      class="w-full flex-grow flex flex-col items-center justify-center px-4 transition-all duration-500"
      :style="{ paddingTop: '1vh' }" 
    >
      <div class="w-full max-w-3xl flex flex-col items-center justify-center space-y-12">
        <!-- 标题 -->
        <div class="flex flex-col items-center space-y-8 text-center">
          <div class="w-24 h-24 bg-gray-50 dark:bg-gray-800 rounded-3xl flex items-center justify-center shadow-sm ring-1 ring-gray-100 dark:ring-gray-700">
            <span class="material-icons-outlined text-6xl text-gray-800 dark:text-gray-200">delete_outline</span>
          </div>
          <h2 class="text-4xl font-bold text-gray-900 dark:text-white tracking-tight">What do you want to filter?</h2>
        </div>

        <!-- 搜索框（初始态） -->
        <div class="w-full max-w-2xl relative">
          <SearchBar
            v-model="keyword"
            :is-searching="isSearching"
            :is-platform-menu-open="isPlatformMenuOpen"
            :selected-platform="selectedPlatform"
            :platforms="platforms"
            @search="handleSearch"
            @toggle-platform-menu="togglePlatformMenu"
            @select-platform="selectPlatform"
          />
        </div>

        <!-- 快捷标签 -->
        <div class="flex flex-wrap justify-center gap-3 w-full max-w-2xl">
          <button
            v-for="tag in quickTags"
            :key="tag"
            @click="handleQuickTag(tag)"
            class="px-4 py-2 rounded-full text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 transition-colors"
          >
            {{ tag }}
          </button>
        </div>
      </div>
    </div>

    <!-- 激活态：搜索栏吸顶 + 结果列表 -->
    <div v-if="isSearching" class="w-full flex flex-col min-h-screen overflow-y-auto">
      <!-- 吸顶搜索栏容器 -->
      <div class="sticky top-16 z-40 w-full bg-surface-light dark:bg-[#0f0f11] py-4 px-4 border-b border-gray-100 dark:border-gray-800 transition-all duration-300">
        <div class="max-w-2xl mx-auto relative">
          <SearchBar
            v-model="keyword"
            :is-searching="isSearching"
            compact
            :is-platform-menu-open="isPlatformMenuOpen"
            :selected-platform="selectedPlatform"
            :platforms="platforms"
            @search="handleSearch"
            @toggle-platform-menu="togglePlatformMenu"
            @select-platform="selectPlatform"
          />
        </div>
      </div>

      <!-- 结果数量提示 -->
      <div class="px-4 pt-6 pb-4 max-w-2xl mx-auto w-full">
        <div class="flex items-center justify-between">
          <p class="text-sm text-gray-500 dark:text-gray-400 font-medium">
            Found {{ searchResults.length }} results for "{{ keyword }}"
          </p>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-400 uppercase font-semibold tracking-wider">Sort by:</span>
            <button class="text-sm font-medium text-gray-700 dark:text-gray-300 flex items-center gap-1 hover:text-black dark:hover:text-white">
              Relevance <span class="material-icons-outlined text-base">arrow_drop_down</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 结果列表 -->
      <div class="flex-grow px-4 pb-20">
        <div class="max-w-2xl mx-auto w-full flex flex-col space-y-4">
          <Transition
            enter-active-class="transition-all duration-300 ease-out"
            enter-from-class="opacity-0 translate-y-2"
            enter-to-class="opacity-100 translate-y-0"
          >
            <div v-if="searchResults.length > 0" class="space-y-4">
              <div
                v-for="(result, index) in searchResults"
                :key="result.id"
                class="group flex items-center justify-between p-4 bg-white dark:bg-[#18181b] border border-gray-100 dark:border-gray-800 rounded-xl hover:shadow-lg hover:border-gray-200 dark:hover:border-gray-700 transition-all duration-200 cursor-pointer"
                :style="{ animationDelay: `${index * 50}ms` }"
              >
                <!-- 左侧：头像 + 信息 -->
                <div class="flex items-center gap-4 flex-1 min-w-0">
                  <div class="relative flex-shrink-0">
                    <img
                      :alt="result.name"
                      :src="result.avatar"
                      class="w-12 h-12 rounded-full object-cover border border-gray-100 dark:border-gray-700"
                    />
                    <div class="absolute -bottom-1 -right-1 bg-white dark:bg-[#18181b] rounded-full p-0.5">
                      <div
                        :class="[
                          'rounded-full p-1 text-white flex items-center justify-center w-5 h-5 flex-shrink-0',
                          result.statusColor === 'green' ? 'bg-blue-500' :
                          result.statusColor === 'red' ? 'bg-red-500' :
                          result.statusColor === 'sky' ? 'bg-sky-500' :
                          result.statusColor === 'black' ? 'bg-black dark:bg-white' :
                          'bg-gray-500'
                        ]"
                      >
                        <span class="material-symbols-outlined text-[10px]">{{ result.icon }}</span>
                      </div>
                    </div>
                  </div>

                  <div class="flex-1 min-w-0">
                    <h3 class="font-semibold text-gray-900 dark:text-white text-lg leading-tight group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors truncate">
                      {{ result.name }}
                    </h3>
                    <div class="flex items-center gap-3 mt-1 text-sm text-gray-500 dark:text-gray-400 flex-wrap">
                      <span class="flex items-center gap-1">{{ result.username }}</span>
                      <span class="w-1 h-1 rounded-full bg-gray-300 dark:bg-gray-600 flex-shrink-0"></span>
                      <span>{{ result.followers }} followers</span>
                      <span v-if="result.status" class="w-1 h-1 rounded-full bg-gray-300 dark:bg-gray-600 flex-shrink-0"></span>
                      <span v-if="result.status" :class="[
                        'text-xs font-medium px-1.5 py-0.5 rounded flex-shrink-0',
                        result.statusColor === 'green' ? 'text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/30' :
                        'text-gray-400'
                      ]">
                        {{ result.status }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- 右侧：订阅按钮 -->
                <button
                  @click.stop="toggleSubscribe(result.id)"
                  :class="[
                    'px-5 py-2 rounded-full text-sm font-semibold transition-all shadow-sm ml-4 flex-shrink-0 whitespace-nowrap',
                    result.isSubscribed
                      ? 'border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800'
                      : 'bg-gray-900 hover:bg-gray-800 dark:bg-white dark:hover:bg-gray-200 text-white dark:text-black hover:shadow active:scale-95'
                  ]"
                >
                  {{ result.isSubscribed ? 'Subscribed' : 'Subscribe' }}
                </button>
              </div>
            </div>
          </Transition>

          <!-- 无结果提示 -->
          <div v-if="searchResults.length === 0" class="py-12 text-center">
            <p class="text-gray-500 dark:text-gray-400">No results found for "{{ keyword }}"</p>
          </div>
        </div>
      </div>

      <!-- 返回按钮 -->
      <div class="fixed bottom-8 right-8 z-30">
        <button
          @click="closeSearch"
          class="p-4 bg-gray-900 hover:bg-gray-800 dark:bg-white dark:hover:bg-gray-200 text-white dark:text-black rounded-full shadow-lg transition-all hover:shadow-xl active:scale-95"
          title="返回搜索"
        >
          <span class="material-icons-outlined">arrow_upward</span>
        </button>
      </div>
    </div>
  </main>
</template>

<script setup>
import { ref } from 'vue'
import SearchBar from './SearchBar.vue'

// ============ 状态初始化 ============
// 搜索激活状态（仅在用户按 Enter 或点击搜索按钮时激活）
const isSearching = ref(false)

// 搜索关键词（**必须初始化为空字符串**，防止 trim() 错误）
const keyword = ref('')

// 平台菜单状态
const isPlatformMenuOpen = ref(false)

// 选中的平台
const selectedPlatform = ref('All Platforms')

// 可用平台列表
const platforms = ['All Platforms', 'Twitter', 'YouTube', 'RSS', 'Medium', 'Email']

// 搜索结果列表（模拟数据）
const searchResults = ref([
  {
    id: 1,
    name: 'Design Digest Daily',
    username: '@designdigest',
    followers: '245k',
    avatar: 'https://lh3.googleusercontent.com/aida-public/AB6AXuB_Aox6L1K_TzjDLlxjUQz40FcjNylIcXlGCbJl3QT0sDmQlXc6CjdF4sZ5rHhc8UyGWCyUFCfocLyueBbU-DnjxrKcpCIQCm_Pi4OvqIlbDbA2CksI14u3_1RdRl9eHxvYjXhGvJ2G_buvLQ6Zupn4DzizlVf6O9jPBC8YMzDUX2i9Ugg3L9HpYWMAQ3aOP5urYH4SoHDpFHQLehHbBHH-0fsSCIDq5XSv5eCzBXttQdI-1mpNbiNCg6SUhCoKdd2lNV6EfAtKRlM',
    status: 'Highly Active',
    statusColor: 'green',
    icon: 'public',
    isSubscribed: false,
  },
  {
    id: 2,
    name: 'Tech Innovations 2024',
    username: '@techinnovate',
    followers: '1.2M',
    avatar: 'https://lh3.googleusercontent.com/aida-public/AB6AXuA6n9vMAMtsBIvs9IT42o3oMrJZF2yofq1w74DH2HacfSdk-P_-DwGFZCT6ielz9U8BYjtMIPcf6IpksMfRoxvYMJLLu7ZQuIDuzy2pBr-ddH7YAwx_MX-nD7w-s-y8bNIOXXDNLh_mI3piPB8_aoOZT958nznW619iXEBOhpdKvopxj6Ntd_-b8OfbSv3nXP6kij-RncK2GwMvMU_mbzBwt000dncbIvgzuqodBMVnS9p9Y2mt95YY9w7uSWdpQapskkO_5Vs3yxk',
    status: null,
    statusColor: 'red',
    icon: 'play_arrow',
    isSubscribed: false,
  },
  {
    id: 3,
    name: 'Future AI Weekly',
    username: '@future_ai',
    followers: '89k',
    avatar: 'https://lh3.googleusercontent.com/aida-public/AB6AXuAo1yRDYyt65TQOlJETB2AgmZFPtoJ4s30sJuojyijUFuYKbwakyUK2yZGRiEDpY6uKHLT3C83Vzr0W5VF7-JM8rD-XiVeBldWh6XsaJtR9lmfB69IWlalfan32tGI7OrmG91Wb68bMM_0X2LO7u4NoAQzwCd61PWjVZHTlIkK0SkLdFtfcenAIYI1Ot1eXOtu2_EO5BBXeAUWH5PfhC1MXzyh1rxv2hvnDsa1xfgf4_7j06xmcKz1iPKAAnIa4QInz1GwaQbRXlvo',
    status: 'Newsletter',
    statusColor: 'sky',
    icon: 'send',
    isSubscribed: true,
  },
  {
    id: 4,
    name: "Sarah's AI Musings",
    username: '@sarahwrites',
    followers: '45k',
    avatar: 'https://lh3.googleusercontent.com/aida-public/AB6AXuCXbLrKQYVGJRxiTRHnB64S3oNgy8m3S86xfoQW9WWmTjQ1uunPGoqFsB4aYrtmc0TnHZzHd3U6cMQKQOWQdYdAyGlRWbkWOukvm24Q567eHEWWZBC5eQU1uGxbjTexvfvPuHXlwyexLcDjqAqVZojOsvWfwsGlFMtU6xhFUnAdlq4U-V7UOUnNWjNR5GguYrbAvIG-dvicJV9YhiiLN97x0BwYLjkNvsRIjV2R83k47HRbviJO9fWYpwJVOj8qbBxor6He4dEopUA',
    status: null,
    statusColor: 'black',
    icon: 'alternate_email',
    isSubscribed: false,
  },
  {
    id: 5,
    name: 'Legacy Tech News',
    username: '@legacytech',
    followers: '12k',
    avatar: 'https://lh3.googleusercontent.com/aida-public/AB6AXuBSYERAmujiyZZbebg1XEzIDxXMZfeE95zkMciBnBQlb1gkaD6QhDdk6o1mK0PxaYLjDSxtbCJGjp-HONBUhi5YtvcGkba-2xePU5hKKorPaTfcKjEr9YeWiAO050LJP6nyHMTqxoQYYZevw7-v8_d8ttPEe_rruCuFgbXYZlns_k6mzMh4l3u4T1tTlRj0MIuAiFzoGYFJKFV8GezwtZcz3_V3WS1vL2YdwrBiLhfre6SM0qu4I5ChJGsuWgpD1cXJUtsmOHJfbRk',
    status: 'Inactive',
    statusColor: 'gray',
    icon: 'rss_feed',
    isSubscribed: false,
  },
])

// 快捷标签
const quickTags = ['Recent News', 'Social Media Spam', 'Email Filtering', 'Ad Block Rules']

// ============ 事件处理 ============
/**
 * 执行搜索（当用户按 Enter 或点击箭头按钮时调用）
 * 关键：仅在这里设置 isSearching = true，导致搜索栏上移
 */
const handleSearch = async (query) => {
  // 防守：确保 keyword 是字符串且不为空
  if (!query || typeof query !== 'string' || query.trim().length === 0) {
    console.warn('⚠️ 搜索词为空或无效，取消搜索')
    return
  }

  console.log(`🔍 搜索启动: "${query}" (平台: ${selectedPlatform.value})`)

  // 激活搜索态（导致搜索栏上移）
  isSearching.value = true

  // 关闭平台菜单
  isPlatformMenuOpen.value = false

  // TODO: 后续对接 Go 后端 API
  // const results = await fetchSearchResults(query, selectedPlatform.value)
  // searchResults.value = results
}

/**
 * 关闭搜索（返回初始态）
 */
const closeSearch = () => {
  isSearching.value = false
  keyword.value = ''
  isPlatformMenuOpen.value = false
  console.log('🔙 返回初始态')
}

/**
 * 切换平台菜单
 */
const togglePlatformMenu = () => {
  isPlatformMenuOpen.value = !isPlatformMenuOpen.value
  console.log(`📌 平台菜单: ${isPlatformMenuOpen.value ? '打开' : '关闭'}`)
}

/**
 * 选择平台
 */
const selectPlatform = (platform) => {
  selectedPlatform.value = platform
  isPlatformMenuOpen.value = false
  console.log(`✓ 已选择平台: ${platform}`)
}

/**
 * 处理快捷标签点击
 */
const handleQuickTag = (tag) => {
  keyword.value = tag
  handleSearch(tag)
}

/**
 * 切换订阅状态
 */
const toggleSubscribe = (resultId) => {
  const result = searchResults.value.find(r => r.id === resultId)
  if (result) {
    result.isSubscribed = !result.isSubscribed
    console.log(`${result.isSubscribed ? '✓ 已订阅' : '✗ 已取消订阅'}: ${result.name}`)
  }
}
</script>

<style scoped>
/* 平滑动画过渡 */
.transition-all {
  transition: all 0.5s ease-in-out;
}

/* 防止文本溢出 */
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
