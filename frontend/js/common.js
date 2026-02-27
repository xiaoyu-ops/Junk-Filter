/**
 * common.js - 通用工具库
 * 包含：主题切换、导航、Toast 提示、剪贴板、Modal、防抖节流等
 */

// ============================================================================
// 1. 主题管理
// ============================================================================

const ThemeManager = {
  /**
   * 初始化主题（页面加载时调用）
   */
  init() {
    const html = document.documentElement;
    const savedTheme = localStorage.getItem('theme');

    if (savedTheme === 'dark') {
      html.classList.add('dark');
    } else if (savedTheme === 'light') {
      html.classList.remove('dark');
    } else {
      // 使用系统偏好
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        html.classList.add('dark');
      }
    }
  },

  /**
   * 切换主题
   */
  toggle() {
    const html = document.documentElement;
    if (html.classList.contains('dark')) {
      html.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    } else {
      html.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    }
  },

  /**
   * 获取当前主题
   */
  getCurrent() {
    return document.documentElement.classList.contains('dark') ? 'dark' : 'light';
  },
};

// ============================================================================
// 2. Toast 提示系统
// ============================================================================

const ToastManager = {
  /**
   * 初始化 Toast 容器
   */
  init() {
    if (!document.getElementById('toast-container')) {
      const container = document.createElement('div');
      container.id = 'toast-container';
      container.className = 'fixed top-4 right-4 z-50 space-y-2 pointer-events-none';
      document.body.appendChild(container);
    }
  },

  /**
   * 显示 Toast
   * @param {string} message - 消息文本
   * @param {string} type - 类型：'success' | 'error' | 'info'
   * @param {number} duration - 显示时长（毫秒），0 表示不自动关闭
   */
  show(message, type = 'success', duration = 3000) {
    const container = document.getElementById('toast-container') || this.init();

    const bgColor = {
      'success': 'bg-green-500',
      'error': 'bg-red-500',
      'info': 'bg-blue-500',
    }[type] || 'bg-green-500';

    const icon = {
      'success': 'check_circle',
      'error': 'error',
      'info': 'info',
    }[type] || 'check_circle';

    const toast = document.createElement('div');
    toast.className = `
      ${bgColor} text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3
      animate-[slideIn_0.3s_ease-out] cursor-pointer pointer-events-auto
      dark:${bgColor}
    `;
    toast.innerHTML = `
      <span class="material-icons-outlined text-lg">${icon}</span>
      <span class="text-sm font-medium">${message}</span>
      <button class="ml-auto p-1 hover:bg-white/20 rounded-lg transition-colors">
        <span class="material-icons-outlined text-base">close</span>
      </button>
    `;

    const container_elem = document.getElementById('toast-container');
    container_elem.appendChild(toast);

    // 关闭按钮
    toast.querySelector('button').addEventListener('click', () => {
      this.dismiss(toast);
    });

    // 自动关闭
    if (duration > 0) {
      setTimeout(() => {
        this.dismiss(toast);
      }, duration);
    }

    return toast;
  },

  /**
   * 关闭 Toast
   */
  dismiss(toastElement) {
    toastElement.style.animation = 'slideOut 0.3s ease-out forwards';
    setTimeout(() => {
      toastElement.remove();
    }, 300);
  },
};

// ============================================================================
// 3. 剪贴板管理
// ============================================================================

const ClipboardManager = {
  /**
   * 复制文本到剪贴板
   * @param {string} text - 要复制的文本
   * @param {HTMLElement} triggerButton - 触发元素（用于显示反馈）
   */
  async copy(text, triggerButton) {
    try {
      await navigator.clipboard.writeText(text);
      this._showCopyFeedback(triggerButton);
      return true;
    } catch (err) {
      console.error('复制失败:', err);
      // 降级方案：使用 execCommand
      return this._fallbackCopy(text, triggerButton);
    }
  },

  /**
   * 显示复制成功反馈
   */
  _showCopyFeedback(button) {
    if (!button) return;

    const originalHTML = button.innerHTML;
    const originalText = button.textContent;

    // 改变按钮状态
    button.innerHTML = '<span class="material-icons-outlined text-lg">check</span>';
    button.style.color = '#10b981'; // 绿色
    button.disabled = true;

    // 显示 Toast
    ToastManager.show('已复制到剪贴板', 'success', 2000);

    // 1.5 秒后恢复
    setTimeout(() => {
      button.innerHTML = originalHTML;
      button.style.color = '';
      button.disabled = false;
    }, 1500);
  },

  /**
   * 降级方案：使用 execCommand（旧浏览器）
   */
  _fallbackCopy(text, button) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);

    try {
      textarea.select();
      document.execCommand('copy');
      this._showCopyFeedback(button);
      return true;
    } catch (err) {
      ToastManager.show('复制失败，请手动选择', 'error', 3000);
      return false;
    } finally {
      document.body.removeChild(textarea);
    }
  },
};

// ============================================================================
// 4. 防抖与节流
// ============================================================================

const Throttle = {
  /**
   * 防抖函数
   * @param {Function} func - 要执行的函数
   * @param {number} delay - 延迟时间（毫秒）
   */
  debounce(func, delay = 300) {
    let timeoutId;
    return function (...args) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func.apply(this, args), delay);
    };
  },

  /**
   * 节流函数
   * @param {Function} func - 要执行的函数
   * @param {number} limit - 时间间隔（毫秒）
   */
  throttle(func, limit = 300) {
    let inThrottle;
    return function (...args) {
      if (!inThrottle) {
        func.apply(this, args);
        inThrottle = true;
        setTimeout(() => (inThrottle = false), limit);
      }
    };
  },
};

// ============================================================================
// 5. 动画工具
// ============================================================================

const AnimationUtils = {
  /**
   * 通用元素动画
   * @param {HTMLElement} element - 要动画的元素
   * @param {Object} options - 动画选项
   *   - {string} keyframes - CSS 关键帧名称
   *   - {number} duration - 持续时间（毫秒）
   *   - {string} easing - 缓动函数
   */
  animate(element, { keyframes = '', duration = 300, easing = 'ease-out' } = {}) {
    return new Promise((resolve) => {
      element.style.animation = `${keyframes} ${duration}ms ${easing} forwards`;
      setTimeout(resolve, duration);
    });
  },

  /**
   * 淡入动画
   */
  fadeIn(element, duration = 300) {
    element.style.opacity = '0';
    element.style.transition = `opacity ${duration}ms ease-out`;
    element.offsetHeight; // 强制重排
    element.style.opacity = '1';
    return new Promise(resolve => setTimeout(resolve, duration));
  },

  /**
   * 淡出动画
   */
  fadeOut(element, duration = 300) {
    element.style.opacity = '1';
    element.style.transition = `opacity ${duration}ms ease-out`;
    element.offsetHeight; // 强制重排
    element.style.opacity = '0';
    return new Promise(resolve => setTimeout(resolve, duration));
  },

  /**
   * 滑入动画（从上方）
   */
  slideDown(element, duration = 300) {
    element.style.maxHeight = '0';
    element.style.overflow = 'hidden';
    element.style.transition = `max-height ${duration}ms ease-out`;
    element.offsetHeight; // 强制重排
    element.style.maxHeight = element.scrollHeight + 'px';
    return new Promise(resolve => setTimeout(resolve, duration));
  },

  /**
   * 滑出动画（向上）
   */
  slideUp(element, duration = 300) {
    element.style.maxHeight = element.scrollHeight + 'px';
    element.style.overflow = 'hidden';
    element.style.transition = `max-height ${duration}ms ease-out`;
    element.offsetHeight; // 强制重排
    element.style.maxHeight = '0';
    return new Promise(resolve => setTimeout(resolve, duration));
  },

  /**
   * 缩放动画
   */
  scale(element, targetScale = 1.02, duration = 300) {
    element.style.transition = `transform ${duration}ms ease-out`;
    element.style.transform = `scale(${targetScale})`;
  },

  /**
   * 取消缩放
   */
  scaleReset(element, duration = 300) {
    element.style.transition = `transform ${duration}ms ease-out`;
    element.style.transform = 'scale(1)';
  },
};

// ============================================================================
// 6. 导航管理
// ============================================================================

const NavigationManager = {
  /**
   * 获取当前页面名称
   */
  getCurrentPage() {
    const pathname = window.location.pathname;
    if (pathname.includes('main.html') || pathname.endsWith('/frontend/')) {
      return 'main';
    } else if (pathname.includes('timeline.html')) {
      return 'timeline';
    } else if (pathname.includes('config.html')) {
      return 'config';
    } else if (pathname.includes('task.html')) {
      return 'task';
    }
    return '';
  },

  /**
   * 初始化导航激活状态
   */
  initNavigation() {
    const currentPage = this.getCurrentPage();
    const navLinks = document.querySelectorAll('nav a');

    navLinks.forEach(link => {
      // 移除所有 font-semibold 和 dark:text-gray-100
      link.classList.remove('font-semibold');
      link.classList.add('text-gray-500', 'dark:text-gray-400');
      link.classList.remove('text-gray-900', 'dark:text-gray-100');

      // 检查链接是否对应当前页面
      if (
        (currentPage === 'main' && link.textContent.includes('主页')) ||
        (currentPage === 'timeline' && link.textContent.includes('时间轴')) ||
        (currentPage === 'config' && link.textContent.includes('配置中心')) ||
        (currentPage === 'task' && link.textContent.includes('分发任务'))
      ) {
        link.classList.add('font-semibold', 'text-gray-900', 'dark:text-gray-100');
        link.classList.remove('text-gray-500', 'dark:text-gray-400');
      }
    });
  },

  /**
   * 绑定导航链接
   */
  bindNavigation() {
    const navMap = {
      '主页': 'main.html',
      '时间轴': 'timeline.html',
      '配置中心': 'config.html',
      '分发任务': 'task.html',
    };

    document.querySelectorAll('nav a').forEach(link => {
      const text = link.textContent.trim();
      if (navMap[text]) {
        link.href = navMap[text];
        link.addEventListener('click', (e) => {
          // 允许正常导航
        });
      }
    });
  },
};

// ============================================================================
// 7. Modal 对话框
// ============================================================================

const ModalManager = {
  /**
   * 创建并显示确认对话框
   */
  confirm(message, onConfirm, onCancel) {
    const modal = document.createElement('div');
    modal.className = 'fixed inset-0 bg-black/50 flex items-center justify-center z-50 animate-[fadeIn_0.2s_ease-out]';
    modal.innerHTML = `
      <div class="bg-white dark:bg-[#1F2937] rounded-lg shadow-xl p-6 max-w-sm mx-4 animate-[slideUp_0.3s_ease-out]">
        <p class="text-gray-900 dark:text-gray-100 mb-6">${message}</p>
        <div class="flex gap-3 justify-end">
          <button class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors cancel-btn">
            取消
          </button>
          <button class="px-4 py-2 text-sm font-medium bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors confirm-btn">
            确认删除
          </button>
        </div>
      </div>
    `;

    document.body.appendChild(modal);

    const confirmBtn = modal.querySelector('.confirm-btn');
    const cancelBtn = modal.querySelector('.cancel-btn');

    confirmBtn.addEventListener('click', () => {
      this._closeModal(modal);
      onConfirm?.();
    });

    cancelBtn.addEventListener('click', () => {
      this._closeModal(modal);
      onCancel?.();
    });

    modal.addEventListener('click', (e) => {
      if (e.target === modal) {
        this._closeModal(modal);
        onCancel?.();
      }
    });
  },

  /**
   * 关闭模态框
   */
  _closeModal(modal) {
    modal.style.animation = 'fadeOut 0.2s ease-out forwards';
    setTimeout(() => modal.remove(), 200);
  },
};

// ============================================================================
// 8. 输入框高度自适应
// ============================================================================

const InputAutoResize = {
  /**
   * 初始化输入框自适应
   */
  init(inputElement, maxHeight = 150) {
    inputElement.style.resize = 'none';
    inputElement.style.overflowY = 'hidden';

    const resize = () => {
      inputElement.style.height = 'auto';
      const newHeight = Math.min(inputElement.scrollHeight, maxHeight);
      inputElement.style.height = newHeight + 'px';
      inputElement.style.overflowY = newHeight >= maxHeight ? 'auto' : 'hidden';
    };

    inputElement.addEventListener('input', resize);
    inputElement.addEventListener('keydown', resize);

    return resize;
  },
};

// ============================================================================
// 9. 页面加载时初始化
// ============================================================================

document.addEventListener('DOMContentLoaded', () => {
  ThemeManager.init();
  ToastManager.init();
  NavigationManager.initNavigation();
  NavigationManager.bindNavigation();

  // 绑定所有主题切换按钮
  document.querySelectorAll('button[onclick*="classList.toggle(\'dark\')"]').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.preventDefault();
      ThemeManager.toggle();
    });
  });

  // 也支持直接调用的主题切换
  const themeToggleBtn = document.getElementById('theme-toggle');
  if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', () => {
      ThemeManager.toggle();
    });
  }
});

// ============================================================================
// 10. CSS 动画定义（需要在 style 标签中添加）
// ============================================================================

const injectAnimationStyles = () => {
  const style = document.createElement('style');
  style.textContent = `
    @keyframes slideIn {
      from {
        opacity: 0;
        transform: translateY(-10px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    @keyframes slideOut {
      from {
        opacity: 1;
        transform: translateY(0);
      }
      to {
        opacity: 0;
        transform: translateY(-10px);
      }
    }

    @keyframes slideUp {
      from {
        opacity: 0;
        transform: translateY(20px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    @keyframes fadeIn {
      from {
        opacity: 0;
      }
      to {
        opacity: 1;
      }
    }

    @keyframes fadeOut {
      from {
        opacity: 1;
      }
      to {
        opacity: 0;
      }
    }

    @keyframes bounce {
      0%, 100% {
        transform: translateY(0);
      }
      50% {
        transform: translateY(-10px);
      }
    }
  `;
  document.head.appendChild(style);
};

injectAnimationStyles();
