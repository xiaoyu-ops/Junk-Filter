/**
 * timeline-page.js - 时间轴交互逻辑
 * 包含：卡片悬浮反馈、侧滑抽屉、过滤切换、主题切换
 */

document.addEventListener('DOMContentLoaded', () => {
  initCardHoverEffect();
  initFilterToggle();
  initSideDrawer();
  ensureThemeToggle();
});

// ============================================================================
// 1. 卡片悬浮反馈
// ============================================================================

function initCardHoverEffect() {
  const cards = document.querySelectorAll('article.bg-white');

  cards.forEach(card => {
    card.addEventListener('mouseenter', () => {
      card.style.transition = 'all 200ms cubic-bezier(0.4, 0, 0.2, 1)';
      card.style.transform = 'scale(1.02)';

      // 移除旧的阴影类，添加新的
      card.classList.remove('shadow-soft');
      card.classList.add('shadow-lg');
      card.style.boxShadow = '0 10px 25px -5px rgba(0, 0, 0, 0.1)';
    });

    card.addEventListener('mouseleave', () => {
      card.style.transition = 'all 200ms cubic-bezier(0.4, 0, 0.2, 1)';
      card.style.transform = 'scale(1)';

      card.classList.add('shadow-soft');
      card.classList.remove('shadow-lg');
      card.style.boxShadow = '';
    });

    // 点击卡片打开侧滑抽屉
    card.addEventListener('click', () => {
      const authorName = card.querySelector('h3')?.textContent || '作者';
      openDetailDrawer(authorName, card);
    });
  });
}

// ============================================================================
// 2. 侧滑抽屉交互
// ============================================================================

function openDetailDrawer(authorName, cardElement) {
  // 创建遮罩
  const backdrop = document.createElement('div');
  backdrop.className = 'fixed inset-0 bg-black/30 z-30 transition-opacity duration-300';
  backdrop.style.animation = 'fadeIn 300ms ease-out forwards';
  document.body.appendChild(backdrop);

  // 创建侧滑抽屉
  const drawer = document.createElement('div');
  drawer.className = 'fixed right-0 top-0 h-full w-96 bg-white dark:bg-[#1F2937] shadow-2xl z-40 flex flex-col overflow-hidden';
  drawer.style.animation = 'slideInRight 400ms cubic-bezier(0.4, 0, 0.2, 1) forwards';

  // 获取卡片信息
  const title = cardElement.querySelector('h2')?.textContent || '标题';
  const content = cardElement.querySelector('p')?.textContent || '内容';
  const timeText = cardElement.querySelector('span:nth-of-type(2)')?.textContent || '时间';

  drawer.innerHTML = `
    <div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
      <h2 class="text-xl font-bold text-gray-900 dark:text-gray-100">详情</h2>
      <button class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors close-drawer-btn">
        <span class="material-icons-outlined">close</span>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-6 space-y-6">
      <!-- 作者信息 -->
      <div class="flex items-center gap-4">
        <div class="w-16 h-16 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
          <span class="material-icons-outlined text-3xl text-gray-400 dark:text-gray-500">person</span>
        </div>
        <div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">${authorName}</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400">${timeText}</p>
        </div>
      </div>

      <!-- 文章信息 -->
      <div>
        <h4 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">标题</h4>
        <p class="text-base text-gray-700 dark:text-gray-300 leading-relaxed">${title}</p>
      </div>

      <div>
        <h4 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">摘要</h4>
        <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">${content}</p>
      </div>

      <!-- 评分 -->
      <div>
        <h4 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">评分</h4>
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

    <div class="p-6 border-t border-gray-200 dark:border-gray-700 flex gap-3">
      <button class="flex-1 px-4 py-2.5 bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-lg font-medium transition-colors">
        订阅作者
      </button>
      <button class="flex-1 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors">
        查看原文
      </button>
    </div>
  `;

  document.body.appendChild(drawer);

  // 关闭按钮
  drawer.querySelector('.close-drawer-btn').addEventListener('click', () => {
    closeDetailDrawer(drawer, backdrop);
  });

  // 点击背景关闭
  backdrop.addEventListener('click', () => {
    closeDetailDrawer(drawer, backdrop);
  });

  // 防止抽屉内点击触发背景关闭
  drawer.addEventListener('click', (e) => {
    e.stopPropagation();
  });

  // 添加动画样式
  addSlideInRightAnimation();
}

function closeDetailDrawer(drawer, backdrop) {
  drawer.style.animation = 'slideOutRight 300ms cubic-bezier(0.4, 0, 0.2, 1) forwards';
  backdrop.style.animation = 'fadeOut 300ms ease-out forwards';

  setTimeout(() => {
    drawer.remove();
    backdrop.remove();
  }, 300);
}

// ============================================================================
// 3. 过滤切换动效
// ============================================================================

function initFilterToggle() {
  const filterButtons = document.querySelectorAll('button.px-4.py-1\\.5.rounded-full');

  filterButtons.forEach(button => {
    // 跳过导航中的按钮
    if (button.closest('nav')) return;

    button.addEventListener('click', () => {
      const isActive = button.classList.contains('bg-black') || button.classList.contains('dark:bg-white');

      if (!isActive) {
        // 移除其他按钮的激活状态
        filterButtons.forEach(btn => {
          if (btn.closest('nav')) return;

          btn.classList.remove('bg-black', 'text-white', 'dark:bg-white', 'dark:text-gray-900');
          btn.classList.add('text-gray-600', 'hover:bg-gray-200', 'dark:text-gray-300', 'dark:hover:bg-gray-800');
        });

        // 激活当前按钮
        button.classList.add('bg-black', 'text-white', 'dark:bg-white', 'dark:text-gray-900');
        button.classList.remove('text-gray-600', 'hover:bg-gray-200', 'dark:text-gray-300', 'dark:hover:bg-gray-800');

        // 卡片重排动画（淡隐淡现）
        const cards = document.querySelectorAll('article.bg-white');
        const filterName = button.textContent.trim();

        // 淡出
        cards.forEach(card => {
          card.style.transition = 'opacity 200ms ease-out';
          card.style.opacity = '0';
        });

        // 延迟后淡入（模拟内容更新）
        setTimeout(() => {
          cards.forEach((card, index) => {
            card.style.transition = 'opacity 300ms ease-out';
            card.style.opacity = '1';
          });

          ToastManager.show(`已切换到: ${filterName}`, 'info', 2000);
        }, 200);
      }
    });
  });
}

// ============================================================================
// 4. 主题切换（确保正常工作）
// ============================================================================

function ensureThemeToggle() {
  const themeToggleBtn = document.getElementById('theme-toggle');

  if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', (e) => {
      e.preventDefault();
      ThemeManager.toggle();
    });
  }
}

// ============================================================================
// 5. 动画定义
// ============================================================================

function addSlideInRightAnimation() {
  // 检查是否已添加
  if (document.getElementById('timeline-animations')) return;

  const style = document.createElement('style');
  style.id = 'timeline-animations';
  style.textContent = `
    @keyframes slideInRight {
      from {
        opacity: 0;
        transform: translateX(100%);
      }
      to {
        opacity: 1;
        transform: translateX(0);
      }
    }

    @keyframes slideOutRight {
      from {
        opacity: 1;
        transform: translateX(0);
      }
      to {
        opacity: 0;
        transform: translateX(100%);
      }
    }
  `;
  document.head.appendChild(style);
}
