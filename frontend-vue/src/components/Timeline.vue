<template>
  <div class="w-full flex flex-col pt-20">
    <!-- 过滤按钮 -->
    <div class="container mx-auto px-4 sm:px-6 lg:px-8 mt-8 mb-6">
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="filter in filters"
          :key="filter"
          @click="timelineStore.setFilter(filter)"
          :class="[
            'px-4 py-1.5 rounded-full text-sm font-medium transition-colors',
            timelineStore.activeFilter === filter
              ? 'bg-black text-white dark:bg-white dark:text-gray-900'
              : 'text-gray-600 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-800'
          ]"
        >
          {{ filter }}
        </button>
        <button class="px-4 py-1.5 rounded-full text-sm font-medium text-gray-600 hover:bg-gray-200 dark:text-gray-300 dark:hover:bg-gray-800 transition-colors flex items-center gap-1 ml-auto sm:ml-0">
          <span class="material-symbols-outlined text-[18px]">filter_list</span> Filter
        </button>
      </div>
    </div>

    <!-- 卡片网格 -->
    <main class="flex-grow container mx-auto px-4 sm:px-6 lg:px-8 pb-12">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <article
          v-for="card in timelineStore.cards"
          :key="card.id"
          @click="timelineStore.openDetailDrawer(card)"
          class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-[#374151] shadow-soft dark:shadow-none overflow-hidden flex flex-col h-full hover:shadow-md hover:scale-[1.02] transition-all duration-200 cursor-pointer"
        >
          <div class="p-6 flex gap-5">
            <!-- 作者头像和信息 -->
            <div class="flex flex-col items-center gap-2 min-w-[72px]">
              <div class="w-14 h-14 rounded-full bg-gray-50 dark:bg-gray-800 flex items-center justify-center overflow-hidden border border-gray-100 dark:border-gray-700">
                <span class="material-symbols-outlined text-3xl text-gray-300 dark:text-gray-500">person</span>
              </div>
              <h3 class="font-semibold text-sm text-center text-gray-900 dark:text-white">{{ card.author }}</h3>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ card.authorTime }}</span>
            </div>

            <!-- 文章内容 -->
            <div class="flex-1 space-y-2">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white leading-tight">{{ card.title }}</h2>
              <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{{ card.content }}</p>
            </div>
          </div>

          <!-- 底部状态 -->
          <div class="mt-auto border-t border-gray-100 dark:border-[#374151]">
            <div class="px-6 py-3 flex items-center justify-between bg-gray-50 dark:bg-[#1F2937]/50">
              <div class="flex items-center gap-2">
                <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Status:</span>
                <div :class="['flex items-center gap-1.5', `text-${card.statusColor}-600 dark:text-${card.statusColor}-400`]">
                  <span class="material-symbols-outlined text-[18px]">
                    {{ card.status === 'Approved' ? 'check_circle' : card.status === 'Rejected' ? 'cancel' : 'help' }}
                  </span>
                  <span class="text-xs font-semibold">{{ card.status }}</span>
                </div>
              </div>
              <button class="text-xs font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white transition-colors">
                View Details
              </button>
            </div>
          </div>
        </article>
      </div>
    </main>

    <!-- 侧滑抽屉 -->
    <Transition
      enter-active-class="transition-all duration-400"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-all duration-300"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="timelineStore.isDetailDrawerOpen"
        @click="timelineStore.closeDetailDrawer"
        class="fixed inset-0 bg-black/30 z-30"
      ></div>
    </Transition>

    <Transition
      enter-active-class="transition-all duration-400 ease-out"
      enter-from-class="translate-x-full opacity-0"
      enter-to-class="translate-x-0 opacity-100"
      leave-active-class="transition-all duration-300 ease-out"
      leave-from-class="translate-x-0 opacity-100"
      leave-to-class="translate-x-full opacity-0"
    >
      <div
        v-if="timelineStore.isDetailDrawerOpen && timelineStore.selectedCard"
        class="fixed right-0 top-0 h-full w-96 bg-white dark:bg-[#1F2937] shadow-2xl z-40 flex flex-col overflow-hidden"
      >
        <!-- 抽屉头部 -->
        <div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">详情</h2>
          <button
            @click="timelineStore.closeDetailDrawer"
            class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
          >
            <span class="material-icons-outlined">close</span>
          </button>
        </div>

        <!-- 抽屉内容 -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6">
          <!-- 作者信息 -->
          <div class="flex items-center gap-4">
            <div class="w-16 h-16 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
              <span class="material-icons-outlined text-3xl text-gray-400 dark:text-gray-500">person</span>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ timelineStore.selectedCard.author }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ timelineStore.selectedCard.authorTime }}</p>
            </div>
          </div>

          <!-- 文章标题 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">标题</h4>
            <p class="text-base text-gray-700 dark:text-gray-300 leading-relaxed">{{ timelineStore.selectedCard.title }}</p>
          </div>

          <!-- 文章内容 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">摘要</h4>
            <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{{ timelineStore.selectedCard.content }}</p>
          </div>

          <!-- 评分 -->
          <div>
            <h4 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">评分</h4>
            <div class="space-y-3">
              <div>
                <div class="flex justify-between text-xs text-gray-600 dark:text-gray-400 mb-1">
                  <span>创新度</span>
                  <span>8/10</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div class="bg-blue-500 h-2 rounded-full" style="width: 80%"></div>
                </div>
              </div>
              <div>
                <div class="flex justify-between text-xs text-gray-600 dark:text-gray-400 mb-1">
                  <span>深度</span>
                  <span>7/10</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div class="bg-green-500 h-2 rounded-full" style="width: 70%"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 抽屉底部按钮 -->
        <div class="p-6 border-t border-gray-200 dark:border-gray-700 flex gap-3">
          <button class="flex-1 px-4 py-2.5 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-900 dark:text-white rounded-lg font-medium transition-colors">
            订阅作者
          </button>
          <button class="flex-1 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors">
            查看原文
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useTimelineStore } from '@/stores'

const timelineStore = useTimelineStore()
const filters = ['All', 'Worth Watching', 'Latest']

onMounted(() => {
  // 初始化时间轴
})
</script>
