/**
 * task-page.js - 任务管理交互逻辑
 * 包含：打字机效果、任务气泡、列表切换、换行&发送、异常处理
 */

document.addEventListener('DOMContentLoaded', () => {
  initTaskList();
  initInputBox();
  initSendButton();
  ensureThemeToggle();
  injectTaskAnimations();
});

// ============================================================================
// 1. 任务列表切换
// ============================================================================

function initTaskList() {
  const taskItems = document.querySelectorAll('aside .p-4');
  const indicator = document.querySelector('aside .border-l-4');

  let activeItem = null;

  taskItems.forEach((item, index) => {
    item.addEventListener('click', () => {
      // 移除旧项的激活状态
      taskItems.forEach(i => {
        i.classList.remove('bg-[#E5E7EB]', 'dark:bg-[#4B5563]', 'border-l-4', 'border-gray-800', 'dark:border-gray-200', 'shadow-sm');
        i.classList.add('bg-white', 'dark:bg-[#374151]', 'hover:bg-gray-50', 'dark:hover:bg-[#4B5563]/80', 'border', 'border-gray-200', 'dark:border-transparent', 'hover:border-gray-300', 'dark:hover:border-gray-500');
      });

      // 激活当前项
      item.classList.remove('bg-white', 'dark:bg-[#374151]', 'hover:bg-gray-50', 'dark:hover:bg-[#4B5563]/80', 'border', 'border-gray-200', 'dark:border-transparent', 'hover:border-gray-300', 'dark:hover:border-gray-500');
      item.classList.add('bg-[#E5E7EB]', 'dark:bg-[#4B5563]', 'border-l-4', 'border-gray-800', 'dark:border-gray-200', 'shadow-sm');

      // 侧滑效果：更新左边框位置
      item.style.transition = 'all 250ms cubic-bezier(0.4, 0, 0.2, 1)';

      // 清空对话框并加载新内容
      const chatContent = document.querySelector('section.flex-1 .flex-1.p-8');
      if (chatContent) {
        // 淡出旧内容
        chatContent.style.transition = 'opacity 200ms ease-out';
        chatContent.style.opacity = '0';

        setTimeout(() => {
          // 更新内容为模拟数据
          const taskName = item.querySelector('h3')?.textContent || '任务';
          chatContent.innerHTML = `
            <div class="flex gap-4">
              <div class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center border border-gray-200 dark:border-transparent">
                <span class="material-icons-outlined text-sm text-gray-600 dark:text-gray-300">person</span>
              </div>
              <div class="flex flex-col space-y-1 max-w-3xl">
                <span class="font-bold text-gray-900 dark:text-gray-300 text-sm">User:</span>
                <p class="text-[#111827] dark:text-gray-100 leading-relaxed text-lg">
                  加载任务: ${taskName}
                </p>
              </div>
            </div>
          `;

          // 淡入新内容
          chatContent.style.opacity = '1';
        }, 200);
      }

      ToastManager.show(`已选择: ${item.querySelector('h3')?.textContent}`, 'info', 1500);
    });
  });
}

// ============================================================================
// 2. 输入框交互：Shift+Enter 换行，单独 Enter 发送
// ============================================================================

function initInputBox() {
  const inputBox = document.querySelector('input[placeholder="输入消息..."]');

  if (!inputBox) return;

  // 初始化自适应高度
  InputAutoResize.init(inputBox, 150);

  inputBox.addEventListener('keydown', (e) => {
    const isShiftEnter = e.shiftKey && e.key === 'Enter';
    const isEnter = !e.shiftKey && e.key === 'Enter';

    if (isShiftEnter) {
      // Shift+Enter：插入换行符
      e.preventDefault();

      const start = inputBox.selectionStart;
      const end = inputBox.selectionEnd;
      const text = inputBox.value;

      inputBox.value = text.substring(0, start) + '\n' + text.substring(end);
      inputBox.selectionStart = inputBox.selectionEnd = start + 1;

      // 触发 input 事件以调整高度
      inputBox.dispatchEvent(new Event('input', { bubbles: true }));
    } else if (isEnter) {
      // 单独 Enter：发送消息
      e.preventDefault();
      sendMessage();
    }
  });
}

// ============================================================================
// 3. 发送消息与 AI 回复
// ============================================================================

function initSendButton() {
  const sendBtn = document.querySelector('button[style*="absolute"][style*="right"]');

  if (sendBtn) {
    sendBtn.addEventListener('click', sendMessage);
  }
}

function sendMessage() {
  const inputBox = document.querySelector('input[placeholder="输入消息..."]');
  const chatContent = document.querySelector('section.flex-1 .flex-1.p-8');

  if (!inputBox || !chatContent) return;

  const message = inputBox.value.trim();

  if (!message) {
    ToastManager.show('请输入消息', 'error', 2000);
    return;
  }

  // 添加用户消息
  const userMessageDiv = document.createElement('div');
  userMessageDiv.className = 'flex gap-4 animate-[slideUp_0.3s_ease-out]';
  userMessageDiv.innerHTML = `
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center border border-gray-200 dark:border-transparent">
      <span class="material-icons-outlined text-sm text-gray-600 dark:text-gray-300">person</span>
    </div>
    <div class="flex flex-col space-y-1 max-w-3xl">
      <span class="font-bold text-gray-900 dark:text-gray-300 text-sm">User:</span>
      <p class="text-[#111827] dark:text-gray-100 leading-relaxed text-lg whitespace-pre-wrap">${escapeHtml(message)}</p>
    </div>
  `;

  chatContent.appendChild(userMessageDiv);

  // 清空输入框
  inputBox.value = '';
  inputBox.style.height = 'auto';

  // 滚动到底部
  chatContent.scrollTop = chatContent.scrollHeight;

  // 模拟 AI 回复
  simulateAIResponse(chatContent);
}

// ============================================================================
// 4. 模拟 AI 回复（包含异常处理）
// ============================================================================

function simulateAIResponse(chatContent) {
  // 添加 AI 消息容器（显示正在输入）
  const aiMessageDiv = document.createElement('div');
  aiMessageDiv.className = 'flex gap-4 animate-[slideUp_0.3s_ease-out]';
  aiMessageDiv.innerHTML = `
    <div class="flex-shrink-0 w-8 h-8 rounded-full bg-black dark:bg-indigo-600 flex items-center justify-center">
      <span class="material-icons-outlined text-sm text-white dark:text-white">smart_toy</span>
    </div>
    <div class="flex flex-col space-y-1 max-w-3xl">
      <span class="font-bold text-gray-900 dark:text-gray-300 text-sm">AI:</span>
      <div class="typing-indicator">
        <span class="typing-dot"></span>
        <span class="typing-dot"></span>
        <span class="typing-dot"></span>
      </div>
    </div>
  `;

  chatContent.appendChild(aiMessageDiv);
  chatContent.scrollTop = chatContent.scrollHeight;

  // 模拟 API 延迟
  setTimeout(() => {
    // 随机决定成功或失败（30% 失败）
    const isSuccess = Math.random() > 0.3;

    if (isSuccess) {
      // 成功：显示打字机效果
      const responseText = generateAIResponse();
      displayTypingEffect(aiMessageDiv, responseText);
    } else {
      // 失败：显示错误消息
      displayAIError(chatContent, aiMessageDiv);
    }
  }, 800);
}

// ============================================================================
// 5. 打字机效果
// ============================================================================

function displayTypingEffect(messageDiv, fullText) {
  const typingIndicator = messageDiv.querySelector('.typing-indicator');

  // 清空 typing dots
  typingIndicator?.remove();

  // 创建文本容器
  const textContainer = document.createElement('p');
  textContainer.className = 'text-[#111827] dark:text-gray-100 leading-relaxed text-lg';
  textContainer.style.minHeight = '1.5em';

  messageDiv.querySelector('div.flex-col').appendChild(textContainer);

  // 逐字显示
  let index = 0;
  const typingSpeed = 30; // 毫秒

  const typeNextCharacter = () => {
    if (index < fullText.length) {
      textContainer.textContent += fullText[index];
      index++;

      // 滚动到底部
      const chatContent = messageDiv.closest('.flex-1');
      if (chatContent) {
        chatContent.scrollTop = chatContent.scrollHeight;
      }

      setTimeout(typeNextCharacter, typingSpeed);
    }
  };

  typeNextCharacter();
}

// ============================================================================
// 6. AI 错误显示与重试
// ============================================================================

function displayAIError(chatContent, aiMessageDiv) {
  const typingIndicator = aiMessageDiv.querySelector('.typing-indicator');
  typingIndicator?.remove();

  const errorContainer = document.createElement('div');
  errorContainer.className = 'flex flex-col space-y-3 max-w-3xl';
  errorContainer.innerHTML = `
    <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 flex items-start gap-3">
      <span class="material-icons-outlined text-red-600 dark:text-red-400 flex-shrink-0">error</span>
      <div class="flex-1">
        <p class="text-sm font-medium text-red-900 dark:text-red-100">抱歉，AI 服务暂时不可用</p>
        <p class="text-xs text-red-700 dark:text-red-300 mt-1">请检查网络连接或稍后重试</p>
      </div>
    </div>
    <button class="retry-btn px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors self-start">
      重试
    </button>
  `;

  aiMessageDiv.querySelector('div.flex-col').appendChild(errorContainer);

  // 重试按钮
  errorContainer.querySelector('.retry-btn').addEventListener('click', () => {
    // 移除错误消息和 AI 消息容器
    aiMessageDiv.remove();
    // 重新触发 AI 回复
    simulateAIResponse(chatContent);
  });

  chatContent.scrollTop = chatContent.scrollHeight;
}

// ============================================================================
// 7. 生成 AI 响应文本
// ============================================================================

function generateAIResponse() {
  const responses = [
    '好的，我已经理解了您的需求。根据最新的数据分析，这个话题确实很有热度。我可以帮助您从多个角度深入分析...',
    '非常有趣的观点！这涉及到了几个关键的行业趋势。让我为您总结一下核心要点：\n1. 市场动向\n2. 技术发展\n3. 未来展望',
    '我同意你的看法。这个话题在最近的讨论中频繁出现，反映了行业的深层变化。让我提供一些相关的数据和见解...',
    '这是个很好的问题！根据我的分析，这个方向的发展潜力很大。主要有以下几个方面值得关注...',
    '有意思的提问！我们可以从历史背景、现状分析和未来预测这三个维度来看待这个问题...',
  ];

  return responses[Math.floor(Math.random() * responses.length)];
}

// ============================================================================
// 8. HTML 转义（防止 XSS）
// ============================================================================

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// ============================================================================
// 9. 主题切换
// ============================================================================

function ensureThemeToggle() {
  const themeToggleBtn = document.querySelector('button[onclick*="classList.toggle"]');

  if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', (e) => {
      e.preventDefault();
      ThemeManager.toggle();
    });
  }
}

// ============================================================================
// 10. 注入任务页面特有的动画
// ============================================================================

function injectTaskAnimations() {
  const style = document.createElement('style');
  style.textContent = `
    @keyframes slideUp {
      from {
        opacity: 0;
        transform: translateY(10px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    .typing-indicator {
      display: flex;
      gap: 4px;
      height: 1.75rem;
      align-items: center;
    }

    .typing-dot {
      width: 6px;
      height: 6px;
      background-color: currentColor;
      border-radius: 50%;
      display: inline-block;
      animation: bounce 1.4s infinite ease-in-out both;
    }

    .typing-dot:nth-child(1) {
      animation-delay: -0.32s;
    }

    .typing-dot:nth-child(2) {
      animation-delay: -0.16s;
    }

    @keyframes bounce {
      0%, 80%, 100% {
        transform: scale(0);
        opacity: 0.5;
      }
      40% {
        transform: scale(1);
        opacity: 1;
      }
    }
  `;
  document.head.appendChild(style);
}
