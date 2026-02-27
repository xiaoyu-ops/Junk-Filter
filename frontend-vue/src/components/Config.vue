<template>
  <main class="flex-grow px-8 py-8 w-full max-w-7xl mx-auto pt-20">
    <div class="space-y-8">
      <!-- RSS订阅源管理 -->
      <section>
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-bold text-[#111827] dark:text-white">订阅源管理 (RSS)</h2>
          <button @click="configStore.toggleRssModal" class="btn-primary">
            <span class="material-icons-outlined text-sm">add</span>
            <span>添加订阅源</span>
          </button>
        </div>

        <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
          <!-- 加载状态 -->
          <div v-if="isConfigLoading" class="p-6">
            <SkeletonLoader :count="4" height="64px" />
          </div>

          <!-- 表格 -->
          <div v-else-if="configStore.sources && configStore.sources.length > 0" class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="border-b border-gray-100 dark:border-gray-700">
                  <th class="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-[#6B7280]" style="width: 5%;"></th>
                  <th class="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-[#6B7280]">源名称</th>
                  <th class="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-[#6B7280]">URL</th>
                  <th class="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-[#6B7280]">更新频率</th>
                  <th class="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-[#6B7280]">状态</th>
                  <th class="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-[#6B7280] text-right">操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                <template v-for="source in configStore.sources" :key="`row-${source.id}`">
                  <!-- 主行 -->
                  <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                    <td class="py-4 px-6 text-center">
                      <button
                        @click="configStore.toggleSourceExpanded(source.id)"
                        class="p-1 text-[#6B7280] hover:text-[#111827] dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors"
                        :title="isSourceExpanded(source.id) ? '收起日志' : '展开日志'"
                      >
                        <span class="material-icons-outlined text-lg" :style="{ transform: isSourceExpanded(source.id) ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.3s' }">expand_more</span>
                      </button>
                    </td>
                    <td class="py-4 px-6 font-medium text-[#111827] dark:text-white">{{ source.name }}</td>
                    <td class="py-4 px-6 text-[#6B7280] text-sm">{{ source.url }}</td>
                    <td class="py-4 px-6 text-[#6B7280] text-sm">{{ formatFrequency(source.frequency) }}</td>
                    <td class="py-4 px-6">
                      <span
                        :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', source.statusClass]"
                      >
                        {{ formatStatus(source.status) }}
                      </span>
                    </td>
                    <td class="py-4 px-6 text-right">
                      <div class="flex items-center justify-end gap-2">
                        <button
                          @click="handleSyncSource(source)"
                          :disabled="source.lastSyncStatus === 'pending'"
                          class="p-1.5 text-[#6B7280] hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                          :title="source.lastSyncStatus === 'pending' ? '同步中...' : '手动同步'"
                        >
                          <span class="material-icons-outlined text-lg" :class="{ 'animate-spin': source.lastSyncStatus === 'pending' }">sync</span>
                        </button>
                        <button
                          @click="handleDeleteSource(source)"
                          class="p-1.5 text-[#6B7280] hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                        >
                          <span class="material-icons-outlined text-lg">delete</span>
                        </button>
                      </div>
                    </td>
                  </tr>

                  <!-- 展开行：同步日志 -->
                  <tr
                    v-if="isSourceExpanded(source.id)"
                    :key="`expanded-${source.id}`"
                    class="bg-gray-50 dark:bg-gray-800/30 border-t border-gray-100 dark:border-gray-700 expanded-row"
                  >
                    <td colspan="6" class="py-0 px-6 overflow-hidden">
                      <div class="max-h-96 transition-all duration-300 ease-in-out" style="max-height: 500px;">
                        <div class="py-4 space-y-3">
                          <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">同步日志</h4>
                          <div v-if="source.syncLogs.length > 0" class="space-y-2 max-h-48 overflow-y-auto">
                            <div
                              v-for="(log, idx) in source.syncLogs"
                              :key="idx"
                              class="text-xs text-gray-600 dark:text-gray-400 p-3 bg-white dark:bg-gray-900/50 rounded border border-gray-200 dark:border-gray-700"
                            >
                              <div class="flex items-start gap-3">
                                <span :class="['material-icons-outlined text-sm', log.status === 'success' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400']">
                                  {{ log.status === 'success' ? 'check_circle' : 'error' }}
                                </span>
                                <div class="flex-1">
                                  <div class="font-medium text-gray-700 dark:text-gray-300">{{ formatDateTime(log.timestamp) }}</div>
                                  <div class="text-gray-600 dark:text-gray-400">{{ log.message }}</div>
                                  <div v-if="log.itemsCount > 0" class="text-gray-500 dark:text-gray-500 mt-1">
                                    获取项目数: {{ log.itemsCount }}
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                          <div v-else class="text-xs text-gray-500 dark:text-gray-400 p-3 bg-white dark:bg-gray-900/50 rounded">
                            暂无同步日志
                          </div>
                        </div>
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>

          <!-- 空状态 -->
          <EmptyState
            v-else
            icon="rss_feed"
            title="还没有 RSS 源"
            subtitle="添加第一个源以开始获取内容"
            action="添加订阅源"
            actionIcon="add_circle"
            @action="configStore.toggleRssModal"
          />
        </div>
      </section>

      <!-- AI模型配置 -->
      <section>
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-bold text-[#111827] dark:text-white">AI 模型配置</h2>
        </div>

        <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm p-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            <!-- 左列：模型名称、API Key 和 Base URL -->
            <div class="space-y-6">
              <!-- 模型名称 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">模型名称</label>
                <input
                  v-model="configStore.modelName"
                  type="text"
                  placeholder="例如: gpt-4-turbo, deepseek-chat"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2.5 px-4 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700 focus:border-gray-400 dark:focus:border-gray-500 transition-shadow"
                />
                <p class="mt-1.5 text-xs text-[#6B7280]">输入模型标识符，例如：gpt-4-turbo、claude-3-sonnet、deepseek-chat</p>
              </div>

              <!-- API 密钥 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">API 密钥</label>
                <div class="relative">
                  <input
                    v-model="configStore.apiKey"
                    :type="showApiKey ? 'text' : 'password'"
                    class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2.5 px-4 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700 focus:border-gray-400 dark:focus:border-gray-500 transition-shadow"
                  />
                  <button
                    @click="showApiKey = !showApiKey"
                    class="absolute right-12 top-2.5 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors"
                  >
                    <span class="material-icons-outlined text-lg">{{ showApiKey ? 'visibility' : 'visibility_off' }}</span>
                  </button>
                  <button
                    @click="copyApiKey"
                    class="absolute right-3 top-2.5 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors"
                    :title="apiKeyCopied ? '已复制' : '复制API Key'"
                  >
                    <span class="material-icons-outlined text-lg">{{ apiKeyCopied ? 'check' : 'content_copy' }}</span>
                  </button>
                </div>
              </div>

              <!-- Base URL -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">
                  Base URL
                  <span class="material-icons-outlined text-base align-text-bottom" style="display: inline; font-size: 16px; margin-left: 4px;">link</span>
                </label>
                <input
                  v-model="configStore.baseUrl"
                  type="text"
                  placeholder="https://api.openai.com/v1"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2.5 px-4 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700 focus:border-gray-400 dark:focus:border-gray-500 transition-shadow"
                />
                <p class="mt-1.5 text-xs text-[#6B7280]">自定义 API 终端地址，留空则使用默认地址</p>
              </div>
            </div>

            <!-- 右列：温度、Top P 和Token -->
            <div class="space-y-6">
              <!-- 温度滑块 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">温度 (Temperature)</label>
                <div class="flex items-center gap-4">
                  <input
                    v-model.number="configStore.temperature"
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-[#111827] dark:accent-gray-300"
                  />
                  <span class="text-sm font-medium text-[#111827] dark:text-white min-w-[3rem]">{{ configStore.temperature.toFixed(1) }}</span>
                </div>
                <p class="mt-1.5 text-xs text-[#6B7280]">较高的值会使输出更加随机，而较低的值会使其更加集中和确定。</p>
              </div>

              <!-- Top P 滑块 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">Top P (核采样)</label>
                <div class="flex items-center gap-4">
                  <input
                    v-model.number="configStore.topP"
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-[#111827] dark:accent-gray-300"
                  />
                  <span class="text-sm font-medium text-[#111827] dark:text-white min-w-[3rem]">{{ configStore.topP.toFixed(2) }}</span>
                </div>
                <p class="mt-1.5 text-xs text-[#6B7280]">较高的值保留更多的低概率词，允许更有创意的生成。</p>
              </div>

              <!-- 最大Token数 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">最大 Token 数</label>
                <input
                  v-model.number="configStore.maxTokens"
                  type="number"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2.5 px-4 focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700 focus:border-gray-400 dark:focus:border-gray-500 transition-shadow"
                />
              </div>
            </div>
          </div>

          <!-- 按钮区域 -->
          <div class="mt-8 flex justify-end gap-3 pt-6 border-t border-gray-200 dark:border-gray-700">
            <button
              @click="exportConfig"
              class="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 active:scale-95 text-white rounded-full text-sm font-medium transition-colors shadow-sm flex items-center gap-2"
            >
              <span class="material-icons-outlined text-sm">file_download</span>
              <span>导出配置</span>
            </button>
            <button
              @click="configStore.saveConfig"
              :disabled="configStore.isSaving"
              :class="['btn-primary', { 'opacity-60 cursor-not-allowed': configStore.isSaving }]"
            >
              <span v-if="configStore.isSaving" class="material-icons-outlined animate-spin">sync</span>
              <span v-else class="material-icons-outlined">check</span>
              <span>{{ configStore.isSaving ? '保存中...' : '保存配置' }}</span>
            </button>
          </div>

          <!-- 保存状态提示 -->
          <div v-if="configStore.saveStatus === 'success'" class="mt-4 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg flex items-center gap-3">
            <span class="material-icons-outlined text-green-600 dark:text-green-400">check_circle</span>
            <span class="text-sm font-medium text-green-800 dark:text-green-200">配置已保存</span>
          </div>
          <div v-if="configStore.saveStatus === 'error'" class="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-center gap-3">
            <span class="material-icons-outlined text-red-600 dark:text-red-400">error</span>
            <span class="text-sm font-medium text-red-800 dark:text-red-200">保存失败，请重试</span>
          </div>
        </div>
      </section>
    </div>

    <!-- 模态框：添加RSS源 -->
    <Transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
      leave-active-class="transition-all duration-300 ease-out"
    >
      <div v-if="configStore.showAddRssModal" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center">
        <Transition
          enter-active-class="transition-all duration-300 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition-all duration-200 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
        >
          <div class="bg-white dark:bg-[#1F2937] rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
            <div class="flex items-center justify-between mb-6">
              <h3 class="text-lg font-bold text-[#111827] dark:text-white">添加订阅源</h3>
              <button
                @click="configStore.toggleRssModal"
                class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              >
                <span class="material-icons-outlined">close</span>
              </button>
            </div>

            <form @submit.prevent="handleAddRss" class="space-y-4">
              <!-- 源名称 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">源名称</label>
                <input
                  v-model="configStore.newRssForm.name"
                  type="text"
                  placeholder="例如：TechCrunch"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                />
                <p v-if="rssErrors.name" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ rssErrors.name }}</p>
              </div>

              <!-- URL -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">RSS Feed URL</label>
                <input
                  v-model="configStore.newRssForm.url"
                  type="text"
                  placeholder="https://example.com/rss"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                />
                <p v-if="rssErrors.url" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ rssErrors.url }}</p>
              </div>

              <!-- 更新频率 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">更新频率</label>
                <div class="relative">
                  <select
                    v-model="configStore.newRssForm.frequency"
                    class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 pr-10 appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                  >
                    <option value="hourly">每小时</option>
                    <option value="30min">每30分钟</option>
                    <option value="2hours">每2小时</option>
                    <option value="daily">每天</option>
                  </select>
                  <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-gray-500">
                    <span class="material-icons-outlined text-sm">expand_more</span>
                  </div>
                </div>
                <p v-if="rssErrors.frequency" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ rssErrors.frequency }}</p>
              </div>

              <!-- 过滤规则（可选） -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">过滤规则（可选）</label>
                <textarea
                  v-model="configStore.newRssForm.filterRules"
                  placeholder="例如：优先级 >= 7"
                  rows="3"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow resize-none"
                ></textarea>
              </div>

              <!-- 按钮 -->
              <div class="flex gap-3 pt-4">
                <button
                  type="button"
                  @click="configStore.toggleRssModal"
                  class="flex-1 px-4 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-white rounded-lg text-sm font-medium transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  :disabled="isRssSubmitting"
                  class="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-lg text-sm font-medium transition-colors disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  <span v-if="!isRssSubmitting" class="material-icons-outlined text-sm">add</span>
                  <span>{{ isRssSubmitting ? '添加中...' : '添加' }}</span>
                </button>
              </div>
            </form>
          </div>
        </Transition>
      </div>
    </Transition>

    <!-- 模态框：添加AI模型 -->
    <Transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
      leave-active-class="transition-all duration-300 ease-out"
    >
      <div v-if="configStore.showAddModelModal" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center">
        <Transition
          enter-active-class="transition-all duration-300 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition-all duration-200 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
        >
          <div class="bg-white dark:bg-[#1F2937] rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
            <div class="flex items-center justify-between mb-6">
              <h3 class="text-lg font-bold text-[#111827] dark:text-white">添加AI模型</h3>
              <button
                @click="configStore.toggleModelModal"
                class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              >
                <span class="material-icons-outlined">close</span>
              </button>
            </div>

            <form @submit.prevent="handleAddModel" class="space-y-4">
              <!-- 模型名称 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">模型名称</label>
                <input
                  v-model="configStore.newModelForm.name"
                  type="text"
                  placeholder="例如：GPT-4"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                />
                <p v-if="modelErrors.name" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ modelErrors.name }}</p>
              </div>

              <!-- 服务商 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">服务商</label>
                <div class="relative">
                  <select
                    v-model="configStore.newModelForm.provider"
                    class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 pr-10 appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                  >
                    <option>OpenAI</option>
                    <option>Anthropic</option>
                    <option>Meta</option>
                    <option>Google</option>
                  </select>
                  <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-gray-500">
                    <span class="material-icons-outlined text-sm">expand_more</span>
                  </div>
                </div>
                <p v-if="modelErrors.provider" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ modelErrors.provider }}</p>
              </div>

              <!-- API 密钥 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">API 密钥</label>
                <div class="relative">
                  <input
                    v-model="configStore.newModelForm.apiKey"
                    :type="showNewModelApiKey ? 'text' : 'password'"
                    placeholder="输入API密钥"
                    class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 pr-10 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                  />
                  <button
                    type="button"
                    @click="showNewModelApiKey = !showNewModelApiKey"
                    class="absolute right-3 top-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                  >
                    <span class="material-icons-outlined text-lg">{{ showNewModelApiKey ? 'visibility' : 'visibility_off' }}</span>
                  </button>
                </div>
                <p v-if="modelErrors.apiKey" class="mt-1 text-xs text-red-600 dark:text-red-400">{{ modelErrors.apiKey }}</p>
              </div>

              <!-- 基础URL（可选） -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-1">基础 URL（可选）</label>
                <input
                  v-model="configStore.newModelForm.baseUrl"
                  type="text"
                  placeholder="https://api.example.com"
                  class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-2 px-3 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 transition-shadow"
                />
              </div>

              <!-- 按钮 -->
              <div class="flex gap-3 pt-4">
                <button
                  type="button"
                  @click="configStore.toggleModelModal"
                  class="flex-1 px-4 py-2 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-white rounded-lg text-sm font-medium transition-colors"
                >
                  取消
                </button>
                <button
                  type="submit"
                  :disabled="isModelSubmitting"
                  class="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white rounded-lg text-sm font-medium transition-colors disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  <span v-if="!isModelSubmitting" class="material-icons-outlined text-sm">add</span>
                  <span>{{ isModelSubmitting ? '添加中...' : '添加' }}</span>
                </button>
              </div>
            </form>
          </div>
        </Transition>
      </div>
    </Transition>
  </main>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useConfigStore } from '@/stores'
import { useToast } from '@/composables/useToast'
import { useFormValidation } from '@/composables/useFormValidation'
import SkeletonLoader from './SkeletonLoader.vue'
import EmptyState from './EmptyState.vue'
import ErrorCard from './ErrorCard.vue'

const configStore = useConfigStore()
const { show: showToast } = useToast()
const { validateRssForm, validateModelForm, errors: validationErrors, clearErrors: clearValidationErrors } = useFormValidation()

const showApiKey = ref(false)
const showNewModelApiKey = ref(false)
const apiKeyCopied = ref(false)
const isRssSubmitting = ref(false)
const isModelSubmitting = ref(false)

const rssErrors = ref({})
const modelErrors = ref({})

// 配置加载状态
const isConfigLoading = ref(false)
const configError = ref(null)

// 格式化频率文本
const formatFrequency = (frequency) => {
  const map = {
    'hourly': '每小时',
    '30min': '每30分钟',
    '2hours': '每2小时',
    'daily': '每天',
  }
  return map[frequency] || frequency
}

// 格式化状态文本
const formatStatus = (status) => {
  const map = {
    'active': '活跃',
    'paused': '暂停',
    'error': '错误',
  }
  return map[status] || status
}

// 格式化日期时间
const formatDateTime = (isoString) => {
  const date = new Date(isoString)
  return date.toLocaleString('zh-CN')
}

// 判断源是否展开
const isSourceExpanded = (id) => {
  return configStore.expandedSourceIds.includes(id)
}

// 处理添加RSS源
const handleAddRss = async () => {
  clearValidationErrors()
  rssErrors.value = {}

  if (!validateRssForm(configStore.newRssForm)) {
    rssErrors.value = validationErrors.value
    return
  }

  isRssSubmitting.value = true

  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 800))

    configStore.addSource()
    showToast(`已添加订阅源: ${configStore.newRssForm.name}`, 'success', 2000)
    configStore.toggleRssModal()
  } catch (error) {
    showToast('添加订阅源失败，请重试', 'error', 2000)
  } finally {
    isRssSubmitting.value = false
  }
}

// 处理添加AI模型
const handleAddModel = async () => {
  clearValidationErrors()
  modelErrors.value = {}

  if (!validateModelForm(configStore.newModelForm)) {
    modelErrors.value = validationErrors.value
    return
  }

  isModelSubmitting.value = true

  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 800))

    // 实现addModel方法（简化，仅显示toast）
    showToast(`已添加模型: ${configStore.newModelForm.name}`, 'success', 2000)
    configStore.toggleModelModal()
  } catch (error) {
    showToast('添加模型失败，请重试', 'error', 2000)
  } finally {
    isModelSubmitting.value = false
  }
}

// 处理删除RSS源
const handleDeleteSource = (source) => {
  if (confirm(`确定要删除订阅源 "${source.name}" 吗？`)) {
    configStore.deleteSource(source.id)
    showToast(`已删除订阅源: ${source.name}`, 'success', 2000)
  }
}

// 处理同步RSS源
const handleSyncSource = async (source) => {
  try {
    await configStore.syncSource(source.id)
    const success = source.lastSyncStatus === 'success'
    const message = success ? `${source.name} 同步成功` : `${source.name} 同步失败`
    showToast(message, success ? 'success' : 'error', 2000)
  } catch (error) {
    showToast('同步出错，请重试', 'error', 2000)
  }
}

// API Key复制
const copyApiKey = async () => {
  try {
    await navigator.clipboard.writeText(configStore.apiKey)
    apiKeyCopied.value = true
    showToast('API Key 已复制', 'success', 2000)
    setTimeout(() => {
      apiKeyCopied.value = false
    }, 1500)
  } catch (err) {
    showToast('复制失败，请手动选择', 'error', 2000)
  }
}

// 导出配置
const exportConfig = async () => {
  const configJson = JSON.stringify({
    modelName: configStore.modelName,
    baseUrl: configStore.baseUrl,
    temperature: configStore.temperature,
    topP: configStore.topP,
    maxTokens: configStore.maxTokens,
  }, null, 2)

  try {
    await navigator.clipboard.writeText(configJson)
    showToast('配置已复制', 'success', 2000)
  } catch (err) {
    showToast('复制失败', 'error', 2000)
  }
}

onMounted(() => {
  configStore.loadConfig()
})
</script>

<style scoped>
.expanded-row {
  overflow: hidden;
  transition: max-height 0.4s ease-in-out;
}
</style>
