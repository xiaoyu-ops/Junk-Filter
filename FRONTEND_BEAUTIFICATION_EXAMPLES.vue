# 美化实战：TrueSignal 增强示例

## 示例 1：增强的 StatCard 组件

**文件**: `src/components/common/StatCardEnhanced.vue`

```vue
<template>
  <div class="stat-card-enhanced" :class="`level-${level}`">
    <!-- 背景渐变动画 -->
    <div class="card-gradient"></div>

    <!-- 装饰元素 -->
    <div class="decoration decoration-top"></div>
    <div class="decoration decoration-bottom"></div>

    <!-- 主要内容 -->
    <div class="card-content">
      <div class="stat-icon">
        <span v-html="iconSvg"></span>
      </div>

      <div class="stat-info">
        <p class="stat-label">{{ label }}</p>
        <div class="stat-value-wrapper">
          <h2 class="stat-value">{{ formattedValue }}</h2>
          <span v-if="showChange" :class="['stat-change', changeClass]">
            {{ changeIndicator }}{{ Math.abs(changePercent) }}%
          </span>
        </div>
      </div>
    </div>

    <!-- 进度条 -->
    <div v-if="showProgress" class="stat-progress">
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
      </div>
      <span class="progress-text">{{ progressPercent }}% 完成</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  label: string
  value: number
  icon?: string
  level?: 'high' | 'medium' | 'low'
  changePercent?: number
  showChange?: boolean
  showProgress?: boolean
  progressMax?: number
  trend?: 'up' | 'down' | 'stable'
}

const props = withDefaults(defineProps<Props>(), {
  level: 'medium',
  changePercent: 0,
  showChange: true,
  showProgress: false,
  progressMax: 100,
  trend: 'stable'
})

// 格式化数值
const formattedValue = computed(() => {
  if (props.value >= 1000000) {
    return (props.value / 1000000).toFixed(1) + 'M'
  }
  if (props.value >= 1000) {
    return (props.value / 1000).toFixed(1) + 'K'
  }
  return props.value.toLocaleString()
})

// 变化指示符
const changeIndicator = computed(() => {
  if (props.changePercent > 0) return '↑ '
  if (props.changePercent < 0) return '↓ '
  return ''
})

// 变化样式类
const changeClass = computed(() => {
  if (props.changePercent > 0) return 'positive'
  if (props.changePercent < 0) return 'negative'
  return 'neutral'
})

// 进度条百分比
const progressPercent = computed(() => {
  return Math.round((props.value / props.progressMax) * 100)
})

// 简单 SVG 图标
const iconSvg = computed(() => {
  const icons: Record<string, string> = {
    chart: '<svg viewBox="0 0 24 24" fill="currentColor"><rect x="3" y="13" width="4" height="8"/><rect x="10" y="3" width="4" height="18"/><rect x="17" y="9" width="4" height="12"/></svg>',
    box: '<svg viewBox="0 0 24 24" fill="currentColor"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>',
    star: '<svg viewBox="0 0 24 24" fill="currentColor"><polygon points="12,2 15,10 24,10 17,16 20,24 12,19 4,24 7,16 0,10 9,10"/></svg>'
  }
  return icons[props.icon || 'box']
})
</script>

<style scoped>
.stat-card-enhanced {
  position: relative;
  overflow: hidden;
  border-radius: 16px;
  padding: 24px;
  background: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  min-height: 160px;
  display: flex;
  flex-direction: column;
}

.stat-card-enhanced:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 32px rgba(102, 126, 234, 0.25);
}

/* 背景渐变 */
.card-gradient {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05), rgba(118, 75, 162, 0.05));
  opacity: 0;
  transition: opacity 0.3s ease;
}

.stat-card-enhanced:hover .card-gradient {
  opacity: 1;
}

/* 装饰元素 */
.decoration {
  position: absolute;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  opacity: 0.08;
  transition: all 0.3s ease;
}

.decoration-top {
  top: -50px;
  left: -50px;
  background: linear-gradient(135deg, #667eea, #764ba2);
}

.decoration-bottom {
  bottom: -50px;
  right: -50px;
  background: linear-gradient(135deg, #764ba2, #667eea);
}

.stat-card-enhanced:hover .decoration {
  opacity: 0.15;
  transform: scale(1.1);
}

/* 内容 */
.card-content {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  flex: 1;
}

.stat-icon {
  width: 56px;
  height: 56px;
  min-width: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  font-size: 24px;
}

.stat-icon svg {
  width: 32px;
  height: 32px;
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-label {
  font-size: 12px;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0 0 4px 0;
  font-weight: 600;
}

.stat-value-wrapper {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #333;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stat-change {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.stat-change.positive {
  color: #10b981;
  background: #dcfce7;
}

.stat-change.negative {
  color: #ef4444;
  background: #fee2e2;
}

.stat-change.neutral {
  color: #6b7280;
  background: #f3f4f6;
}

/* 进度条 */
.stat-progress {
  position: relative;
  z-index: 2;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.progress-bar {
  width: 100%;
  height: 6px;
  background: #f0f0f0;
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 6px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea, #764ba2);
  transition: width 0.3s ease;
  border-radius: 3px;
}

.progress-text {
  font-size: 11px;
  color: #999;
}

/* 等级样式 */
.level-high {
  border-top: 3px solid #10b981;
}

.level-medium {
  border-top: 3px solid #f59e0b;
}

.level-low {
  border-top: 3px solid #ef4444;
}

/* 响应式 */
@media (max-width: 640px) {
  .stat-card-enhanced {
    padding: 16px;
    min-height: 140px;
  }

  .stat-value {
    font-size: 24px;
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    min-width: 48px;
  }
}
</style>
```

---

## 示例 2：渐变按钮组件

**文件**: `src/components/common/GradientButton.vue`

```vue
<template>
  <button
    class="gradient-button"
    :class="[`size-${size}`, `variant-${variant}`, { disabled }]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <span class="button-text">{{ text }}</span>
    <span v-if="showArrow" class="button-arrow">→</span>
  </button>
</template>

<script setup lang="ts">
interface Props {
  text: string
  size?: 'sm' | 'md' | 'lg'
  variant?: 'primary' | 'secondary' | 'success' | 'danger'
  disabled?: boolean
  showArrow?: boolean
}

withDefaults(defineProps<Props>(), {
  size: 'md',
  variant: 'primary',
  disabled: false,
  showArrow: false
})

defineEmits<{
  click: []
}>()
</script>

<style scoped>
.gradient-button {
  position: relative;
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: inherit;
}

/* 渐变背景 */
.variant-primary {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.variant-secondary {
  background: linear-gradient(135deg, #8b9cff, #9b6fc4);
  color: white;
  box-shadow: 0 4px 12px rgba(139, 156, 255, 0.3);
}

.variant-success {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.variant-danger {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

/* 尺寸 */
.size-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.size-md {
  padding: 10px 20px;
  font-size: 14px;
}

.size-lg {
  padding: 14px 28px;
  font-size: 16px;
}

/* 悬停效果 */
.gradient-button:not(.disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
}

.gradient-button:not(.disabled):active {
  transform: translateY(0);
}

/* 箭头动画 */
.button-arrow {
  transition: transform 0.3s ease;
}

.gradient-button:not(.disabled):hover .button-arrow {
  transform: translateX(4px);
}

/* 禁用状态 */
.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.disabled:hover {
  transform: none;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 加载状态动画 */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.loading {
  animation: pulse 1.5s ease-in-out infinite;
}
</style>
```

---

## 示例 3：增强的表格组件

**文件**: `src/components/common/EnhancedTable.vue`

```vue
<template>
  <div class="table-wrapper">
    <div class="table-header">
      <h3>{{ title }}</h3>
      <div class="table-actions">
        <slot name="actions"></slot>
      </div>
    </div>

    <a-table
      :columns="columns"
      :data-source="dataSource"
      :loading="loading"
      :pagination="pagination"
      class="enhanced-table"
      :scroll="{ x: 1200 }"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, text, record }">
        <slot name="cell" :column="column" :text="text" :record="record">
          {{ text }}
        </slot>
      </template>

      <template #emptyText>
        <div class="empty-state">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/>
          </svg>
          <p>暂无数据</p>
        </div>
      </template>
    </a-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  title?: string
  columns: any[]
  dataSource: any[]
  loading?: boolean
  total?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  total: 0
})

const emit = defineEmits<{
  change: [page: number, pageSize: number, sorter: any, filter: any]
}>()

const current = ref(1)
const pageSize = ref(20)

const pagination = computed(() => ({
  current: current.value,
  pageSize: pageSize.value,
  total: props.total || props.dataSource.length,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条记录`
}))

const handleTableChange = (page: any, filter: any, sorter: any) => {
  current.value = page.current
  pageSize.value = page.pageSize
  emit('change', page.current, page.pageSize, sorter, filter)
}
</script>

<style scoped>
.table-wrapper {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.table-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: #333;
}

.table-actions {
  display: flex;
  gap: 8px;
}

.enhanced-table :deep(.ant-table) {
  border-radius: 8px;
  overflow: hidden;
}

.enhanced-table :deep(.ant-table-thead > tr > th) {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05), rgba(118, 75, 162, 0.05));
  font-weight: 600;
  color: #333;
  border-bottom: 2px solid #e0e0e0;
}

.enhanced-table :deep(.ant-table-tbody > tr) {
  transition: all 0.2s ease;
}

.enhanced-table :deep(.ant-table-tbody > tr:hover) {
  background: rgba(102, 126, 234, 0.05);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #999;
}

.empty-state svg {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
  opacity: 0.3;
}

.empty-state p {
  margin: 0;
  font-size: 14px;
}

@media (max-width: 768px) {
  .table-wrapper {
    padding: 16px;
  }

  .table-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}
</style>
```

---

## 示例 4：使用美化组件

**修改**: `src/views/DashboardView.vue`

```vue
<template>
  <div class="dashboard-view">
    <!-- 统计卡片网格 -->
    <a-row :gutter="[20, 20]" class="stats-grid">
      <a-col
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
        v-for="stat in stats"
        :key="stat.id"
      >
        <StatCardEnhanced
          :label="stat.label"
          :value="stat.value"
          :icon="stat.icon"
          :level="stat.level"
          :change-percent="stat.changePercent"
          :show-progress="stat.showProgress"
          :progress-max="stat.progressMax"
        />
      </a-col>
    </a-row>

    <!-- 最新有趣内容 -->
    <div class="content-section mt-6">
      <EnhancedTable
        title="最新有趣内容"
        :columns="contentColumns"
        :data-source="recentContent"
        :loading="loading"
        :total="evaluationStore.evaluations.length"
      >
        <template #cell="{ column, text, record }">
          <template v-if="column.key === 'title'">
            <a href="#" class="content-title">{{ truncateText(text, 40) }}</a>
          </template>
          <template v-else-if="column.key === 'scores'">
            <a-tag color="green">{{ record.innovation_score }}/10</a-tag>
            <a-tag color="blue">{{ record.depth_score }}/10</a-tag>
          </template>
          <template v-else-if="column.key === 'date'">
            <span class="date-text">{{ formatDate(text) }}</span>
          </template>
        </template>

        <template #actions>
          <GradientButton text="刷新" variant="primary" @click="loadEvaluations" />
        </template>
      </EnhancedTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useEvaluationStore } from '@/stores'
import StatCardEnhanced from '@/components/common/StatCardEnhanced.vue'
import EnhancedTable from '@/components/common/EnhancedTable.vue'
import GradientButton from '@/components/common/GradientButton.vue'
import { formatDate, truncateText } from '@/utils/format'

const evaluationStore = useEvaluationStore()
const loading = ref(false)

const stats = ref([
  {
    id: 1,
    label: 'RSS 源',
    value: 12,
    icon: 'chart',
    level: 'high' as const,
    changePercent: 8,
    showProgress: false
  },
  {
    id: 2,
    label: '内容数',
    value: 1250,
    icon: 'box',
    level: 'medium' as const,
    changePercent: -3,
    showProgress: false
  },
  {
    id: 3,
    label: '已评估',
    value: 856,
    icon: 'chart',
    level: 'medium' as const,
    changePercent: 12,
    showProgress: true,
    progressMax: 1250
  },
  {
    id: 4,
    label: '有趣内容',
    value: 156,
    icon: 'star',
    level: 'high' as const,
    changePercent: 25,
    showProgress: false
  }
])

const contentColumns = [
  { title: '标题', dataIndex: 'title', key: 'title', width: '40%' },
  { title: '评分', dataIndex: 'scores', key: 'scores', width: '25%' },
  { title: '时间', dataIndex: 'created_at', key: 'date', width: '20%' },
  { title: '操作', key: 'action', width: '15%' }
]

const recentContent = computed(() => {
  return evaluationStore.evaluations
    .filter(e => e.decision === 'INTERESTING')
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 6)
})

const loadEvaluations = async () => {
  loading.value = true
  try {
    await evaluationStore.fetchEvaluations()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadEvaluations()
})
</script>

<style scoped>
.dashboard-view {
  animation: fadeIn 0.5s ease-out;
}

.stats-grid {
  margin-bottom: 40px;
}

.content-section {
  animation: slideInUp 0.5s ease-out;
}

.content-title {
  color: #667eea;
  text-decoration: none;
  transition: color 0.2s ease;
}

.content-title:hover {
  color: #764ba2;
  text-decoration: underline;
}

.date-text {
  font-size: 12px;
  color: #999;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.mt-6 {
  margin-top: 24px;
}
</style>
```

---

## 实施步骤

### 第一步：创建增强组件
```bash
# 在 frontend/src/components/common/ 目录下创建
# - StatCardEnhanced.vue
# - GradientButton.vue
# - EnhancedTable.vue
```

### 第二步：更新导出
**文件**: `src/components/index.ts`（如果有的话）或直接在视图中导入

### 第三步：修改现有视图
在各个视图中使用新的增强组件

### 第四步：验证效果
```bash
docker-compose up -d
# 访问 http://localhost:5173
# 刷新页面查看效果
```

---

## 下一步建议

✅ **立即可实施**：
- 统计卡片增强
- 渐变按钮
- 增强表格

⭐ **短期优化**（1-2 周）：
- 添加 AOS 动画库
- 实现暗黑模式
- 添加加载骨架屏

🎨 **中期美化**（2-4 周）：
- 集成 ECharts 数据可视化
- 自定义主题系统
- 高级悬停效果

