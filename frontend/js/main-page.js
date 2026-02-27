/**
 * main-page.js - 主页交互逻辑
 * 包含：搜索框聚焦、平台选择、快捷标签悬停
 */

document.addEventListener('DOMContentLoaded', () => {
  initSearchBoxFocus();
  initPlatformDropdown();
  initQuickTagsHover();
});

// ============================================================================
// 1. 搜索框聚焦效果
// ============================================================================

function initSearchBoxFocus() {
  const searchInput = document.querySelector('input[placeholder="Search for keywords..."]');
  const searchContainer = document.querySelector('.rounded-full[style*="flex"]') ||
                         searchInput?.closest('.flex.items-center[style*="w-full"]');

  if (!searchInput || !searchContainer) return;

  const parentContainer = searchContainer.closest('.relative.group.max-w-2xl');

  if (!parentContainer) return;

  // 聚焦事件
  searchInput.addEventListener('focus', () => {
    const flexContainer = parentContainer.querySelector('.flex.items-center.w-full');
    if (flexContainer) {
      flexContainer.style.transition = 'all 300ms ease-out';
      flexContainer.style.boxShadow = 'var(--tw-shadow-lg)';
      flexContainer.style.borderColor = '#9ca3af'; // 深灰色
      flexContainer.classList.add('ring-2', 'ring-gray-200', 'dark:ring-gray-700');
    }
  });

  // 失焦事件
  searchInput.addEventListener('blur', () => {
    const flexContainer = parentContainer.querySelector('.flex.items-center.w-full');
    if (flexContainer) {
      flexContainer.style.boxShadow = 'var(--tw-shadow-xl)';
      flexContainer.style.borderColor = '';
      flexContainer.classList.remove('ring-2', 'ring-gray-200', 'dark:ring-gray-700');
    }
  });

  // 发送按钮点击事件
  const sendButton = parentContainer.querySelector('button[style*="p-3"]');
  if (sendButton) {
    sendButton.addEventListener('click', () => {
      const keyword = searchInput.value.trim();
      if (keyword) {
        ToastManager.show(`搜索关键词: "${keyword}"`, 'info', 2000);
        searchInput.value = '';
      } else {
        ToastManager.show('请输入搜索关键词', 'error', 2000);
      }
    });

    // Enter 键发送
    searchInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        sendButton.click();
      }
    });
  }
}

// ============================================================================
// 2. 平台选择下拉菜单
// ============================================================================

function initPlatformDropdown() {
  const platformButton = document.querySelector('button span:first-child')?.closest('button');

  // 更精确地获取 Select Platform 按钮
  let selectBtn = null;
  document.querySelectorAll('button').forEach(btn => {
    if (btn.textContent.includes('Select Platform')) {
      selectBtn = btn;
    }
  });

  if (!selectBtn) return;

  // 创建下拉菜单容器
  const dropdownContainer = document.createElement('div');
  dropdownContainer.className = 'absolute top-full left-0 mt-2 w-56 bg-white dark:bg-[#2D3748] rounded-lg border border-gray-200 dark:border-gray-600 shadow-lg z-10 overflow-hidden opacity-0 max-h-0 transition-all duration-300 ease-out pointer-events-none';
  dropdownContainer.style.transformOrigin = 'top left';

  const platforms = [
    { name: 'Blog', icon: 'rss_feed' },
    { name: 'Twitter', icon: 'language' },
    { name: 'Medium', icon: 'article' },
    { name: 'Email', icon: 'mail' },
    { name: 'YouTube', icon: 'play_circle' },
  ];

  dropdownContainer.innerHTML = platforms.map(p => `
    <div class="px-4 py-3 hover:bg-gray-50 dark:hover:bg-[#3D4A5C] cursor-pointer transition-colors border-b border-gray-100 dark:border-gray-700 last:border-b-0 flex items-center gap-3">
      <span class="material-icons-outlined text-gray-600 dark:text-gray-300">${p.icon}</span>
      <span class="text-sm font-medium text-gray-900 dark:text-gray-100">${p.name}</span>
    </div>
  `).join('');

  // 获取 Select Platform 按钮的父容器，插入下拉菜单
  const platformButtonParent = selectBtn.closest('.relative');
  if (platformButtonParent) {
    platformButtonParent.appendChild(dropdownContainer);
  }

  let isOpen = false;

  // 打开/关闭下拉菜单
  selectBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    isOpen = !isOpen;

    if (isOpen) {
      dropdownContainer.style.opacity = '1';
      dropdownContainer.style.maxHeight = '300px';
      dropdownContainer.style.pointerEvents = 'auto';

      // 旋转箭头
      const arrow = selectBtn.querySelector('.material-icons-outlined');
      if (arrow) {
        arrow.style.transform = 'rotate(180deg)';
        arrow.style.transition = 'transform 300ms ease-out';
      }
    } else {
      dropdownContainer.style.opacity = '0';
      dropdownContainer.style.maxHeight = '0';
      dropdownContainer.style.pointerEvents = 'none';

      const arrow = selectBtn.querySelector('.material-icons-outlined');
      if (arrow) {
        arrow.style.transform = 'rotate(0deg)';
      }
    }
  });

  // 菜单项点击事件
  dropdownContainer.querySelectorAll('div:not(.pointer-events-none)').forEach((item, index) => {
    item.addEventListener('click', () => {
      const platformName = platforms[index].name;
      selectBtn.textContent = platformName;

      // 重新添加箭头
      const arrow = document.createElement('span');
      arrow.className = 'material-icons-outlined text-base';
      arrow.textContent = 'expand_more';
      arrow.style.transform = '0deg';
      arrow.style.transition = 'transform 300ms ease-out';
      selectBtn.appendChild(arrow);

      // 关闭菜单
      isOpen = false;
      dropdownContainer.style.opacity = '0';
      dropdownContainer.style.maxHeight = '0';
      dropdownContainer.style.pointerEvents = 'none';

      ToastManager.show(`已选择平台: ${platformName}`, 'info', 2000);
    });
  });

  // 点击页面其他地方关闭菜单
  document.addEventListener('click', () => {
    if (isOpen) {
      isOpen = false;
      dropdownContainer.style.opacity = '0';
      dropdownContainer.style.maxHeight = '0';
      dropdownContainer.style.pointerEvents = 'none';

      const arrow = selectBtn.querySelector('.material-icons-outlined');
      if (arrow) {
        arrow.style.transform = 'rotate(0deg)';
      }
    }
  });
}

// ============================================================================
// 3. 快捷标签悬停效果
// ============================================================================

function initQuickTagsHover() {
  const quickTags = document.querySelectorAll('button.px-5.py-2\\.5.rounded-full.text-sm.font-medium');

  quickTags.forEach(tag => {
    // 跳过导航链接按钮（如果有的话）
    if (tag.closest('nav')) return;

    tag.addEventListener('mouseenter', () => {
      tag.style.transition = 'all 300ms ease-out';
      tag.style.transform = 'translateY(-2px)';
      tag.style.backgroundColor = '#f3f4f6'; // 更深的背景色
      tag.style.boxShadow = '0 4px 6px rgba(0, 0, 0, 0.07)';
    });

    tag.addEventListener('mouseleave', () => {
      tag.style.transform = 'translateY(0)';
      tag.style.backgroundColor = '';
      tag.style.boxShadow = '';
    });

    // 点击事件
    tag.addEventListener('click', () => {
      const tagName = tag.textContent.trim();
      ToastManager.show(`正在筛选: ${tagName}`, 'info', 2000);
    });
  });
}
