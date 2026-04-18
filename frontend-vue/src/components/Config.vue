<template>
  <main class="flex-grow px-8 py-8 w-full max-w-7xl mx-auto pt-20">
    <div class="space-y-8">
      <!-- RSS订阅源管理 -->
      <section>
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-bold text-[#111827] dark:text-white">订阅源管理 (RSS)</h2>
          <div class="flex items-center gap-2">
            <button @click="showPresetModal = true" class="flex items-center gap-1.5 px-3 py-2 text-sm font-medium rounded-lg border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
              <span class="material-icons-outlined text-sm">auto_awesome</span>
              <span>推荐源</span>
            </button>
            <button @click="configStore.toggleRssModal" class="btn-primary">
              <span class="material-icons-outlined text-sm">add</span>
              <span>添加订阅源</span>
            </button>
          </div>
        </div>

        <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
          <!-- 加载状态 -->
          <div v-if="configStore.isLoadingSources" class="p-6">
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

      <!-- RSS 代理配置 -->
      <section>
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-bold text-[#111827] dark:text-white">RSS 代理设置</h2>
        </div>

        <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm p-6">
          <p class="text-sm text-[#6B7280] mb-4">
            用于抓取无法直接访问的 RSS 源（如 feedburner.com），支持 HTTP/HTTPS/SOCKS5 代理。修改后即时生效，无需重启服务。
          </p>
          <div class="flex items-end gap-4">
            <div class="flex-1">
              <label class="block text-sm font-medium text-[#374151] dark:text-gray-300 mb-1.5">代理地址</label>
              <input
                v-model="configStore.rssProxyUrl"
                type="text"
                placeholder="例如: http://127.0.0.1:7890 或 socks5://127.0.0.1:1080"
                class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg text-sm text-[#111827] dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              />
            </div>
            <button
              @click="handleSaveProxy"
              :disabled="configStore.isSavingProxy"
              class="btn-primary whitespace-nowrap"
            >
              <span v-if="configStore.isSavingProxy" class="material-icons-outlined text-sm animate-spin">sync</span>
              <span v-else class="material-icons-outlined text-sm">save</span>
              <span>{{ configStore.isSavingProxy ? '保存中...' : '保存' }}</span>
            </button>
          </div>
          <!-- 保存状态提示 -->
          <div v-if="configStore.proxySaveStatus === 'success'" class="mt-3 flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400">
            <span class="material-icons-outlined text-sm">check_circle</span>
            <span>代理配置已更新，即时生效</span>
          </div>
          <div v-else-if="configStore.proxySaveStatus === 'error'" class="mt-3 flex items-center gap-1.5 text-sm text-red-600 dark:text-red-400">
            <span class="material-icons-outlined text-sm">error</span>
            <span>保存失败，请检查后端是否运行</span>
          </div>
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

      <!-- 通知设置 -->
      <section>
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-bold text-[#111827] dark:text-white">通知设置</h2>
        </div>

        <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm p-6">
          <!-- 加载状态 -->
          <div v-if="notifLoading" class="flex items-center gap-2 text-gray-500">
            <span class="material-icons-outlined animate-spin text-sm">sync</span>
            <span class="text-sm">加载中...</span>
          </div>

          <div v-else class="space-y-6">
            <!-- 总开关 -->
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-sm font-medium text-[#111827] dark:text-white">启用通知</h3>
                <p class="text-xs text-[#6B7280] mt-0.5">当高价值内容被评估完成时推送通知</p>
              </div>
              <button
                @click="notifSettings.enabled = !notifSettings.enabled"
                :class="[
                  'relative w-11 h-6 rounded-full transition-colors',
                  notifSettings.enabled ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'
                ]"
              >
                <span
                  :class="[
                    'absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform',
                    notifSettings.enabled ? 'translate-x-5' : 'translate-x-0'
                  ]"
                />
              </button>
            </div>

            <div v-if="notifSettings.enabled" class="space-y-6">
              <!-- INTERESTING 决策通知 -->
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-sm font-medium text-[#111827] dark:text-white">INTERESTING 决策自动通知</h3>
                  <p class="text-xs text-[#6B7280] mt-0.5">无论分数高低，只要决策为 INTERESTING 就发送通知</p>
                </div>
                <button
                  @click="notifSettings.notify_on_interesting = !notifSettings.notify_on_interesting"
                  :class="[
                    'relative w-11 h-6 rounded-full transition-colors',
                    notifSettings.notify_on_interesting ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'
                  ]"
                >
                  <span
                    :class="[
                      'absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform',
                      notifSettings.notify_on_interesting ? 'translate-x-5' : 'translate-x-0'
                    ]"
                  />
                </button>
              </div>

              <!-- 分数阈值 -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">
                    最低创新分 (Innovation Score)
                  </label>
                  <div class="flex items-center gap-4">
                    <input
                      v-model.number="notifSettings.min_innovation_score"
                      type="range" min="1" max="10" step="1"
                      class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-[#111827] dark:accent-gray-300"
                    />
                    <span class="text-sm font-bold text-[#111827] dark:text-white min-w-[2rem] text-center">{{ notifSettings.min_innovation_score }}</span>
                  </div>
                  <p class="mt-1 text-xs text-[#6B7280]">创新分达到此值以上时触发通知</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">
                    最低深度分 (Depth Score)
                  </label>
                  <div class="flex items-center gap-4">
                    <input
                      v-model.number="notifSettings.min_depth_score"
                      type="range" min="1" max="10" step="1"
                      class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-[#111827] dark:accent-gray-300"
                    />
                    <span class="text-sm font-bold text-[#111827] dark:text-white min-w-[2rem] text-center">{{ notifSettings.min_depth_score }}</span>
                  </div>
                  <p class="mt-1 text-xs text-[#6B7280]">深度分达到此值以上时触发通知</p>
                </div>
              </div>

              <!-- 关注的 RSS 源 -->
              <div>
                <label class="block text-sm font-medium text-[#111827] dark:text-white mb-2">
                  关注的 RSS 源
                  <span class="text-xs font-normal text-[#6B7280] ml-1">（留空则监控所有源）</span>
                </label>
                <div v-if="configStore.sources && configStore.sources.length > 0" class="space-y-2 max-h-48 overflow-y-auto border border-gray-200 dark:border-gray-700 rounded-lg p-3">
                  <label
                    v-for="source in configStore.sources"
                    :key="source.id"
                    class="flex items-center gap-3 py-1.5 px-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer transition-colors"
                  >
                    <input
                      type="checkbox"
                      :value="source.id"
                      v-model="notifSettings.watched_source_ids"
                      class="w-4 h-4 rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                    />
                    <span class="text-sm text-[#111827] dark:text-white">{{ source.name }}</span>
                    <span class="text-xs text-[#6B7280] ml-auto truncate max-w-[200px]">{{ source.url }}</span>
                  </label>
                </div>
                <p v-else class="text-xs text-[#6B7280] p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  暂无 RSS 源，请先在上方添加订阅源
                </p>
              </div>

            <!-- 推送渠道配置 -->
              <div>
                <div class="flex items-center justify-between mb-2">
                  <label class="block text-sm font-medium text-[#111827] dark:text-white">
                    推送渠道
                    <span class="text-xs font-normal text-[#6B7280] ml-1">（将通知推送到手机）</span>
                  </label>
                  <button
                    @click="addPushChannel"
                    class="text-xs px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-1"
                  >
                    <span class="material-icons-outlined text-sm">add</span>
                    添加渠道
                  </button>
                </div>

                <div v-if="notifSettings.push_channels && notifSettings.push_channels.length > 0" class="space-y-3">
                  <div
                    v-for="(ch, idx) in notifSettings.push_channels"
                    :key="idx"
                    class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 space-y-3"
                  >
                    <div class="flex items-center justify-between">
                      <div class="flex items-center gap-3">
                        <!-- 渠道类型 -->
                        <span class="text-sm font-medium text-[#111827] dark:text-white">Telegram</span>
                        <input type="hidden" v-model="ch.type" />
                        <!-- 启用开关 -->
                        <button
                          @click="ch.enabled = !ch.enabled"
                          :class="[
                            'relative w-9 h-5 rounded-full transition-colors',
                            ch.enabled ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'
                          ]"
                        >
                          <span
                            :class="[
                              'absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full shadow transition-transform',
                              ch.enabled ? 'translate-x-4' : 'translate-x-0'
                            ]"
                          />
                        </button>
                      </div>
                      <div class="flex items-center gap-2">
                        <button
                          @click="testPushChannel(ch, idx)"
                          :disabled="pushTestingIdx === idx"
                          class="text-xs px-2.5 py-1 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-lg transition-colors disabled:opacity-50"
                        >
                          {{ pushTestingIdx === idx ? '测试中...' : '测试' }}
                        </button>
                        <button
                          @click="removePushChannel(idx)"
                          class="p-1 text-gray-400 hover:text-red-500 transition-colors"
                        >
                          <span class="material-icons-outlined text-lg">delete</span>
                        </button>
                      </div>
                    </div>

                    <!-- Telegram: bot_token + chat_id -->
                    <div class="space-y-2">
                      <div>
                        <label class="block text-xs text-[#6B7280] mb-1">Bot Token</label>
                        <input v-model="ch.bot_token" type="text" placeholder="1234567890:ABCDEFGxxxxxxxx"
                          class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-1.5 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700" />
                        <p class="text-xs text-[#6B7280] mt-1">从 @BotFather 获取</p>
                      </div>
                      <div>
                        <label class="block text-xs text-[#6B7280] mb-1">Chat ID</label>
                        <input v-model="ch.chat_id" type="text" placeholder="123456789"
                          class="w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 text-[#111827] dark:text-white rounded-lg py-1.5 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-gray-200 dark:focus:ring-gray-700" />
                        <p class="text-xs text-[#6B7280] mt-1">给 @userinfobot 发消息获取你的 Chat ID</p>
                      </div>
                    </div>

                    <!-- 测试结果 -->
                    <div v-if="pushTestResults[idx]" :class="[
                      'text-xs p-2 rounded',
                      pushTestResults[idx].success ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300' : 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300'
                    ]">
                      {{ pushTestResults[idx].message }}
                    </div>
                  </div>
                </div>
                <p v-else class="text-xs text-[#6B7280] p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  暂未配置推送渠道，点击"添加渠道"开始配置
                </p>
              </div>
            </div>

            <!-- 保存按钮 -->
            <div class="flex justify-end pt-4 border-t border-gray-200 dark:border-gray-700">
              <button
                @click="saveNotifSettings"
                :disabled="notifSaving"
                :class="['btn-primary', { 'opacity-60 cursor-not-allowed': notifSaving }]"
              >
                <span v-if="notifSaving" class="material-icons-outlined animate-spin">sync</span>
                <span v-else class="material-icons-outlined">check</span>
                <span>{{ notifSaving ? '保存中...' : '保存通知设置' }}</span>
              </button>
            </div>

            <!-- 保存状态提示 -->
            <div v-if="notifSaveStatus === 'success'" class="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg flex items-center gap-3">
              <span class="material-icons-outlined text-green-600 dark:text-green-400">check_circle</span>
              <span class="text-sm font-medium text-green-800 dark:text-green-200">通知设置已保存</span>
            </div>
          </div>
        </div>
      </section>

      <!-- 系统维护 -->
      <section>
        <h2 class="text-2xl font-bold text-[#111827] dark:text-white mb-6">系统维护</h2>
        <div class="bg-white dark:bg-[#1F2937] rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm p-6">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-sm font-medium text-[#111827] dark:text-white">清理消息队列</h3>
              <p class="text-xs text-[#6B7280] mt-1">清空 Redis Stream 中的所有待处理消息并重置消费者组。适用于数据库重置后清理残留消息。</p>
            </div>
            <button
              @click="handlePurgeStream"
              :disabled="purging"
              :class="[
                'px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 whitespace-nowrap',
                purging
                  ? 'bg-gray-300 dark:bg-gray-600 text-gray-500 dark:text-gray-400 cursor-not-allowed'
                  : 'bg-red-50 hover:bg-red-100 dark:bg-red-900/20 dark:hover:bg-red-900/40 text-red-600 dark:text-red-400 border border-red-200 dark:border-red-800'
              ]"
            >
              <span v-if="purging" class="material-icons-outlined text-sm animate-spin">sync</span>
              <span v-else class="material-icons-outlined text-sm">delete_sweep</span>
              <span>{{ purging ? '清理中...' : '清理队列' }}</span>
            </button>
          </div>
        </div>
      </section>
    </div>

    <!-- 模态框：推荐RSS源 -->
    <Transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
      leave-active-class="transition-all duration-300 ease-out"
    >
      <div v-if="showPresetModal" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center" @click.self="showPresetModal = false">
        <Transition
          enter-active-class="transition-all duration-300 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition-all duration-200 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
        >
          <div class="bg-white dark:bg-[#1F2937] rounded-xl shadow-lg w-full max-w-2xl mx-4 flex flex-col max-h-[80vh]">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-700">
              <div>
                <h3 class="text-lg font-bold text-[#111827] dark:text-white">推荐订阅源</h3>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">以下均为提供全文内容的 RSS 源，适合 LLM 深度评估</p>
              </div>
              <button @click="showPresetModal = false" class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                <span class="material-icons-outlined">close</span>
              </button>
            </div>
            <div class="overflow-y-auto flex-1 px-6 py-4 space-y-6">
              <div v-for="category in PRESET_SOURCES" :key="category.name">
                <div class="flex items-center gap-2 mb-3">
                  <span class="text-base">{{ category.icon }}</span>
                  <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ category.name }}</h4>
                  <div class="flex-1 h-px bg-gray-100 dark:bg-gray-700"></div>
                </div>
                <div class="space-y-2">
                  <div
                    v-for="source in category.sources"
                    :key="source.url"
                    class="flex items-start justify-between gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                  >
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-gray-900 dark:text-white mb-0.5">{{ source.name }}</p>
                      <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">{{ source.desc }}</p>
                      <span class="text-xs text-gray-400 dark:text-gray-500 font-mono truncate block">{{ source.url }}</span>
                    </div>
                    <button
                      @click="addPresetSource(source)"
                      :disabled="isPresetAdded(source.url) || presetAdding[source.url]"
                      class="flex-shrink-0 flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
                      :class="isPresetAdded(source.url)
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 cursor-default'
                        : presetAdding[source.url]
                          ? 'bg-gray-100 dark:bg-gray-700 text-gray-400 cursor-not-allowed'
                          : 'bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/20 dark:hover:bg-blue-900/40 text-blue-700 dark:text-blue-400'"
                    >
                      <span class="material-icons-outlined text-sm">
                        {{ isPresetAdded(source.url) ? 'check' : presetAdding[source.url] ? 'hourglass_empty' : 'add' }}
                      </span>
                      {{ isPresetAdded(source.url) ? '已添加' : presetAdding[source.url] ? '添加中' : '添加' }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div class="px-6 py-3 border-t border-gray-100 dark:border-gray-700 flex justify-end">
              <button @click="showPresetModal = false" class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors">
                关闭
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>

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

// 推荐RSS源
const showPresetModal = ref(false)
const presetAdding = ref({})

const PRESET_SOURCES = [
  {
    name: 'AI 与研究',
    icon: '🤖',
    sources: [
      { name: 'OpenAI Blog', url: 'https://openai.com/news/rss.xml', desc: 'OpenAI 官方博客，产品发布与研究进展' },
      { name: 'Google DeepMind', url: 'https://deepmind.google/blog/rss.xml', desc: 'DeepMind 研究博客，前沿 AI 突破' },
      { name: 'Hugging Face Blog', url: 'https://huggingface.co/blog/feed.xml', desc: 'AI 模型与工具生态，开源 ML 社区动态' },
      { name: 'The Gradient', url: 'https://thegradient.pub/rss/', desc: '学术级 AI 研究科普，深度分析前沿论文与趋势' },
      { name: 'Simon Willison\'s Blog', url: 'https://simonwillison.net/atom/everything/', desc: '前 Django 作者，AI 工具与 LLM 应用的精准观察' },
      { name: 'Lilian Weng\'s Blog', url: 'https://lilianweng.github.io/index.xml', desc: 'OpenAI 研究员，强化学习与 LLM 原理深度长文' },
      { name: 'fast.ai Blog', url: 'https://www.fast.ai/index.xml', desc: '深度学习实践，Jeremy Howard 团队的研究与教程' },
    ],
  },
  {
    name: '技术社区',
    icon: '💻',
    sources: [
      { name: 'Hacker News 精华', url: 'https://hnrss.org/best', desc: '硅谷技术社区最高评分帖，长期价值内容' },
      { name: 'Hacker News 首页', url: 'https://hnrss.org/frontpage', desc: '硅谷技术社区实时热帖' },
      { name: 'HN AI 讨论', url: 'https://hnrss.org/newest?q=AI', desc: 'Hacker News 上 AI 相关的最新讨论' },
      { name: 'HN LLM 讨论', url: 'https://hnrss.org/newest?q=LLM', desc: 'Hacker News 上 LLM 相关的最新讨论' },
      { name: 'Ars Technica', url: 'https://feeds.arstechnica.com/arstechnica/index', desc: '深度技术报道，科学、计算机与文化' },
      { name: 'Dev.to', url: 'https://dev.to/feed', desc: '开发者社区文章，实用教程与经验分享' },
    ],
  },
  {
    name: '科技媒体',
    icon: '📰',
    sources: [
      { name: 'MIT Technology Review', url: 'https://www.technologyreview.com/feed/', desc: 'MIT 技术评论，科技趋势的深度分析' },
      { name: 'Wired', url: 'https://www.wired.com/feed/rss', desc: '连线杂志，科技与文化的前沿报道' },
      { name: 'TechCrunch', url: 'https://techcrunch.com/feed/', desc: '科技创业与投资动态' },
      { name: 'The Verge', url: 'https://www.theverge.com/rss/index.xml', desc: '科技产品与行业新闻，叙事性强' },
      { name: 'Nature', url: 'https://www.nature.com/nature.rss', desc: '顶级学术期刊，科学研究前沿动态' },
    ],
  },
  {
    name: '工程博客',
    icon: '🛠️',
    sources: [
      { name: 'GitHub Blog', url: 'https://github.blog/feed/', desc: 'GitHub 官方工程博客，开发者工具与产品更新' },
      { name: 'Cloudflare Blog', url: 'https://blog.cloudflare.com/rss/', desc: '网络、安全与边缘计算的深度工程实践' },
      { name: 'Netflix Tech Blog', url: 'https://netflixtechblog.com/feed', desc: 'Netflix 工程团队，大规模系统与 ML 实践' },
      { name: 'AWS Blog', url: 'https://aws.amazon.com/blogs/aws/feed/', desc: 'AWS 官方博客，云服务与架构实践' },
      { name: 'Meta Engineering', url: 'https://engineering.fb.com/feed/', desc: 'Meta 工程博客，大规模系统与开源项目' },
      { name: 'Mozilla Hacks', url: 'https://hacks.mozilla.org/feed/', desc: 'Mozilla 工程师，Web 标准与浏览器技术' },
      { name: 'Stripe Blog', url: 'https://stripe.com/blog/feed.rss', desc: 'Stripe 工程与产品思考，支付与 API 设计' },
      { name: 'Supabase Blog', url: 'https://supabase.com/blog/rss.xml', desc: 'Supabase 技术博客，开源 BaaS 实践' },
      { name: 'ByteByteGo', url: 'https://blog.bytebytego.com/feed', desc: '系统设计深度 Newsletter，图解大型系统架构' },
      { name: 'Tailscale Blog', url: 'https://tailscale.com/blog/index.xml', desc: '网络与安全深度好文，写作质量极高' },
    ],
  },
  {
    name: '编程语言',
    icon: '⚙️',
    sources: [
      { name: 'Go Blog', url: 'https://go.dev/blog/feed.atom', desc: 'Go 语言官方博客，语言特性与最佳实践' },
      { name: 'Rust Blog', url: 'https://blog.rust-lang.org/feed.xml', desc: 'Rust 官方博客，版本发布与语言进展' },
      { name: 'This Week in Rust', url: 'https://this-week-in-rust.org/atom.xml', desc: 'Rust 社区周报，精选文章与项目' },
      { name: 'TypeScript Blog', url: 'https://devblogs.microsoft.com/typescript/feed/', desc: 'TypeScript 官方博客，版本特性详解' },
      { name: 'React Blog', url: 'https://react.dev/rss.xml', desc: 'React 官方博客，框架更新与最佳实践' },
      { name: 'Vue Blog', url: 'https://blog.vuejs.org/feed.rss', desc: 'Vue.js 官方博客，生态系统动态' },
      { name: 'Python Blog', url: 'https://blog.python.org/feeds/posts/default', desc: 'Python 官方博客，语言发展与社区动态' },
      { name: 'Swift Blog', url: 'https://www.swift.org/atom.xml', desc: 'Swift 官方博客，Apple 生态开发进展' },
      { name: 'Kotlin Blog', url: 'https://blog.jetbrains.com/kotlin/feed/', desc: 'Kotlin 官方博客，JVM 与多平台开发' },
    ],
  },
  {
    name: '安全',
    icon: '🔒',
    sources: [
      { name: 'Krebs on Security', url: 'https://krebsonsecurity.com/feed/', desc: '知名安全记者 Brian Krebs，深度安全事件报道' },
      { name: 'Schneier on Security', url: 'https://www.schneier.com/feed/', desc: '安全专家 Bruce Schneier，安全与隐私思考' },
      { name: 'The Hacker News', url: 'https://feeds.feedburner.com/TheHackersNews', desc: '网络安全新闻，漏洞与威胁情报' },
      { name: 'FreeBuf', url: 'https://www.freebuf.com/feed', desc: '国内安全社区，漏洞分析与安全研究' },
    ],
  },
  {
    name: '中文技术',
    icon: '🇨🇳',
    sources: [
      { name: '阮一峰的网络日志', url: 'http://www.ruanyifeng.com/blog/atom.xml', desc: '每周科技文章与工具推荐，中文技术界标杆' },
      { name: 'LinuxDo 热门', url: 'https://linux.do/top.rss', desc: 'Linux Do 社区高热度帖，技术与开发者话题' },
      { name: 'V2EX 技术', url: 'https://www.v2ex.com/feed/tab/tech.xml', desc: 'V2EX 技术板块，国内开发者讨论' },
      { name: 'IT之家', url: 'https://www.ithome.com/rss/', desc: '国内科技资讯，产品与行业动态' },
      { name: 'SSPAI', url: 'https://sspai.com/feed', desc: '少数派，效率工具与数字生活方式' },
    ],
  },
  {
    name: '产品与设计',
    icon: '🎨',
    sources: [
      { name: 'Product Hunt', url: 'https://www.producthunt.com/feed', desc: '每日新产品发现，科技创业产品聚合' },
      { name: 'Smashing Magazine', url: 'https://www.smashingmagazine.com/feed/', desc: 'Web 设计与前端开发深度教程' },
      { name: 'Tailwind CSS Blog', url: 'https://tailwindcss.com/feeds/feed.xml', desc: 'Tailwind CSS 官方博客，框架更新与设计理念' },
    ],
  },
]

const isPresetAdded = (url) => {
  return configStore.sources?.some(s => s.url === url)
}

const addPresetSource = async (source) => {
  if (isPresetAdded(source.url) || presetAdding.value[source.url]) return
  presetAdding.value[source.url] = true
  try {
    const res = await fetch(`${API_BASE_URL}/sources`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        url: source.url,
        author_name: source.name,
        platform: 'blog',
        priority: 7,
        enabled: true,
        fetch_interval_seconds: 3600,
      }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || `HTTP ${res.status}`)
    }
    const newSource = await res.json()
    configStore.sources.push({
      id: newSource.id,
      name: newSource.author_name || source.name,
      url: newSource.url,
      frequency: 'hourly',
      status: 'active',
      statusClass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      lastSyncTime: null,
      lastSyncStatus: 'success',
      syncLogs: [],
    })
    showToast(`已添加：${source.name}`, 'success', 2000)
  } catch (err) {
    showToast(`添加失败：${err.message}`, 'error', 2000)
  } finally {
    presetAdding.value[source.url] = false
  }
}

// Notification settings
const API_BASE_URL = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`
const notifLoading = ref(false)
const notifSaving = ref(false)
const notifSaveStatus = ref(null)
const notifSettings = ref({
  min_innovation_score: 8,
  min_depth_score: 7,
  notify_on_interesting: true,
  watched_source_ids: [],
  enabled: true,
  push_channels: [],
})
const pushTestingIdx = ref(null)
const pushTestResults = ref({})

// 系统维护：清理消息队列
const purging = ref(false)
const handlePurgeStream = async () => {
  if (!confirm('确定要清空消息队列吗？这将删除所有待处理的评估任务。')) return
  purging.value = true
  try {
    const res = await fetch(`${API_BASE_URL}/admin/purge-stream`, { method: 'POST' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    showToast('消息队列已清理', 'success', 2000)
  } catch (err) {
    console.error('[Config] Failed to purge stream:', err)
    showToast('清理失败: ' + err.message, 'error', 2000)
  } finally {
    purging.value = false
  }
}

const loadNotifSettings = async () => {
  notifLoading.value = true
  try {
    const res = await fetch(`${API_BASE_URL}/notifications/settings`)
    if (res.ok) {
      const data = await res.json()
      notifSettings.value = {
        min_innovation_score: data.min_innovation_score ?? 8,
        min_depth_score: data.min_depth_score ?? 7,
        notify_on_interesting: data.notify_on_interesting ?? true,
        watched_source_ids: data.watched_source_ids ?? [],
        enabled: data.enabled ?? true,
        push_channels: data.push_channels ?? [],
      }
    }
  } catch (err) {
    console.error('[Config] Failed to load notification settings:', err)
  } finally {
    notifLoading.value = false
  }
}

const saveNotifSettings = async () => {
  notifSaving.value = true
  notifSaveStatus.value = null
  try {
    const res = await fetch(`${API_BASE_URL}/notifications/settings`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(notifSettings.value),
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    notifSaveStatus.value = 'success'
    showToast('通知设置已保存', 'success', 2000)
    setTimeout(() => { notifSaveStatus.value = null }, 3000)
  } catch (err) {
    console.error('[Config] Failed to save notification settings:', err)
    showToast('保存通知设置失败', 'error', 2000)
  } finally {
    notifSaving.value = false
  }
}

const addPushChannel = () => {
  if (!notifSettings.value.push_channels) {
    notifSettings.value.push_channels = []
  }
  notifSettings.value.push_channels.push({ type: 'telegram', bot_token: '', chat_id: '', enabled: true })
}

const removePushChannel = (idx) => {
  notifSettings.value.push_channels.splice(idx, 1)
  pushTestResults.value = {}
}

const testPushChannel = async (channel, idx) => {
  pushTestingIdx.value = idx
  pushTestResults.value[idx] = null
  try {
    const res = await fetch(`${API_BASE_URL}/notifications/test-push`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ channel }),
    })
    const data = await res.json()
    if (res.ok) {
      pushTestResults.value[idx] = { success: true, message: data.message || '测试消息已发送' }
    } else {
      pushTestResults.value[idx] = { success: false, message: data.detail || data.error || '测试失败' }
    }
  } catch (err) {
    pushTestResults.value[idx] = { success: false, message: '网络错误: ' + err.message }
  } finally {
    pushTestingIdx.value = null
  }
}

const rssErrors = ref({})
const modelErrors = ref({})

// 配置加载状态（使用 store 的状态）
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

// RSS 代理保存
const handleSaveProxy = async () => {
  const success = await configStore.saveRssProxy()
  showToast(
    success ? '代理配置已更新' : '保存失败',
    success ? 'success' : 'error',
    2000
  )
  if (success) {
    setTimeout(() => { configStore.proxySaveStatus = null }, 3000)
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

onMounted(async () => {
  await Promise.all([configStore.loadConfig(), loadNotifSettings()])
  console.log('[Config] Loaded config:', {
    apiKey: configStore.apiKey ? configStore.apiKey.substring(0, 20) + '***' : 'empty',
    modelName: configStore.modelName,
    baseUrl: configStore.baseUrl,
    temperature: configStore.temperature,
    topP: configStore.topP,
    maxTokens: configStore.maxTokens,
  })
})
</script>

<style scoped>
.expanded-row {
  overflow: hidden;
  transition: max-height 0.4s ease-in-out;
}
</style>
