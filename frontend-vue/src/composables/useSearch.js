import { ref } from 'vue'

export function useSearch() {
  // ============ 状态管理 ============
  /** @type {import('vue').Ref<string>} */
  const selectedPlatform = ref('Blog')

  /** @type {import('vue').Ref<string>} */
  const keyword = ref('')

  /** @type {import('vue').Ref<boolean>} 控制下拉菜单显隐 */
  const isDropdownOpen = ref(false)

  /** @type {import('vue').Ref<HTMLElement|null>} 下拉菜单容器 DOM 引用 */
  const dropdownRef = ref(null)

  // ============ 平台列表数据 ============
  const platforms = [
    { name: 'Blog', icon: 'rss_feed' },
    { name: 'Twitter', icon: 'language' },
    { name: 'Medium', icon: 'article' },
    { name: 'Email', icon: 'mail' },
    { name: 'YouTube', icon: 'play_circle' },
  ]

  // ============ 下拉菜单控制方法 ============
  /**
   * 切换下拉菜单的开关状态
   * @returns {void}
   */
  const toggleDropdown = () => {
    isDropdownOpen.value = !isDropdownOpen.value
    console.log(`📌 切换下拉菜单: ${isDropdownOpen.value ? '打开' : '关闭'}`)
  }

  /**
   * 打开下拉菜单
   * @returns {void}
   */
  const openDropdown = () => {
    isDropdownOpen.value = true
    console.log(`📌 打开下拉菜单`)
  }

  /**
   * 关闭下拉菜单
   * @returns {void}
   */
  const closeDropdown = () => {
    isDropdownOpen.value = false
    console.log(`📌 关闭下拉菜单`)
  }

  /**
   * 选择平台并关闭菜单
   * @param {string} platformName - 平台名称
   * @returns {void}
   */
  const selectPlatform = (platformName) => {
    selectedPlatform.value = platformName
    // 立刻关闭菜单，防止任何延迟
    isDropdownOpen.value = false
    console.log(`✓ 已选择平台: ${platformName}`)
  }

  // ============ 搜索功能 ============
  /**
   * 处理搜索逻辑
   * @returns {boolean} 是否成功执行搜索
   */
  const handleSearch = () => {
    if (keyword.value && keyword.value.toString().trim()) {
      console.log(`🔍 搜索: ${keyword.value} (平台: ${selectedPlatform.value})`)
      // 保存搜索词用于其他用途（如日志、分析）
      const searchQuery = keyword.value
      // 清空输入框
      keyword.value = ''
      return true
    }
    return false
  }

  // ============ 导出接口 ============
  return {
    // 状态
    selectedPlatform,
    keyword,
    isDropdownOpen,
    dropdownRef,

    // 常量
    platforms,

    // 方法
    toggleDropdown,
    openDropdown,
    closeDropdown,
    selectPlatform,
    handleSearch,
  }
}
