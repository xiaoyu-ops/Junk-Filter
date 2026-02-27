/**
 * config-page.js - 配置中心交互逻辑
 * 包含：滑块联动、表格操作、保存反馈、一键复制功能
 */

document.addEventListener('DOMContentLoaded', () => {
  initTemperatureSlider();
  initTableRowInteractions();
  initSaveButton();
  initCopyButtons();
});

// ============================================================================
// 1. Temperature 滑块实时联动
// ============================================================================

function initTemperatureSlider() {
  const slider = document.querySelector('input[type="range"]');
  const valueDisplay = document.querySelector('span.text-sm.font-medium.text-\\[\\#111827\\].min-w-\\[3rem\\]');

  if (!slider || !valueDisplay) return;

  // 实时同步值
  const updateValue = () => {
    const value = parseFloat(slider.value).toFixed(1);
    valueDisplay.textContent = value;

    // 更新滑块的背景渐变（进度条效果）
    const percentage = (slider.value - slider.min) / (slider.max - slider.min) * 100;
    slider.style.background = `linear-gradient(to right, #111827 0%, #111827 ${percentage}%, #e5e7eb ${percentage}%, #e5e7eb 100%)`;
  };

  slider.addEventListener('input', updateValue);

  // 初始化
  updateValue();

  // 视觉反馈：拖动时增加阴影
  slider.addEventListener('mousedown', () => {
    slider.style.boxShadow = '0 0 0 6px rgba(17, 24, 39, 0.1)';
    slider.style.transition = 'none';
  });

  slider.addEventListener('mouseup', () => {
    slider.style.boxShadow = '';
    slider.style.transition = 'box-shadow 200ms ease-out';
  });
}

// ============================================================================
// 2. 表格行操作反馈
// ============================================================================

function initTableRowInteractions() {
  const tableRows = document.querySelectorAll('tbody tr');

  tableRows.forEach(row => {
    const deleteBtn = row.querySelector('button.p-1.5.text-\\[\\#6B7280\\].hover\\:text-red-600');

    if (!deleteBtn) return;

    // 行悬停效果
    row.addEventListener('mouseenter', () => {
      row.style.backgroundColor = '#f9fafb';
      row.style.transition = 'background-color 200ms ease-out';
    });

    row.addEventListener('mouseleave', () => {
      row.style.backgroundColor = '';
    });

    // 删除按钮点击
    deleteBtn.addEventListener('click', (e) => {
      e.stopPropagation();

      const sourceName = row.querySelector('td:first-child')?.textContent.trim();

      ModalManager.confirm(
        `确定要删除订阅源 "${sourceName}" 吗？`,
        () => {
          // 删除动画：向左滑出
          row.style.transition = 'all 300ms ease-out';
          row.style.transform = 'translateX(-100%)';
          row.style.opacity = '0';

          setTimeout(() => {
            row.style.display = 'none';
            ToastManager.show(`已删除订阅源: ${sourceName}`, 'success', 2000);
          }, 300);
        }
      );
    });
  });

  // 编辑按钮（演示）
  document.querySelectorAll('tbody tr button.p-1\\.5.text-\\[\\#6B7280\\].hover\\:text-\\[\\#111827\\]').forEach(editBtn => {
    editBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      ToastManager.show('编辑功能即将推出', 'info', 2000);
    });
  });
}

// ============================================================================
// 3. 保存配置按钮
// ============================================================================

function initSaveButton() {
  const saveBtn = document.querySelector('button:has(.material-icons-outlined[style*="font-size"])')
    || Array.from(document.querySelectorAll('button')).find(btn => btn.textContent.includes('保存配置'));

  if (!saveBtn) return;

  saveBtn.addEventListener('click', async () => {
    const originalText = saveBtn.textContent;
    const originalHTML = saveBtn.innerHTML;

    // 禁用按钮
    saveBtn.disabled = true;
    saveBtn.style.opacity = '0.6';

    // 显示加载状态
    saveBtn.innerHTML = '<span class="material-icons-outlined animate-spin">sync</span><span class="ml-2">保存中...</span>';
    saveBtn.style.pointerEvents = 'none';

    // 模拟 API 请求（随机失败 20%）
    await new Promise(resolve => setTimeout(resolve, 1000));

    const isSuccess = Math.random() > 0.2;

    if (isSuccess) {
      // 成功状态
      saveBtn.style.backgroundColor = '#10b981'; // 绿色
      saveBtn.innerHTML = '<span class="material-icons-outlined">check</span><span class="ml-2">保存成功</span>';

      ToastManager.show('配置已保存', 'success', 3000);

      // 2 秒后恢复
      setTimeout(() => {
        saveBtn.innerHTML = originalHTML;
        saveBtn.style.backgroundColor = '';
        saveBtn.disabled = false;
        saveBtn.style.opacity = '1';
        saveBtn.style.pointerEvents = 'auto';
      }, 2000);
    } else {
      // 失败状态
      saveBtn.style.backgroundColor = '#ef4444'; // 红色
      saveBtn.innerHTML = '<span class="material-icons-outlined">error</span><span class="ml-2">保存失败</span>';

      ToastManager.show('配置保存失败，请检查网络连接', 'error', 3000);

      // 恢复为重试状态
      setTimeout(() => {
        saveBtn.disabled = false;
        saveBtn.style.opacity = '1';
        saveBtn.style.pointerEvents = 'auto';
        saveBtn.style.backgroundColor = '#ef4444';
        saveBtn.innerHTML = '<span class="material-icons-outlined">refresh</span><span class="ml-2">重试</span>';
      }, 2000);
    }
  });
}

// ============================================================================
// 4. 一键复制功能
// ============================================================================

function initCopyButtons() {
  // 为 API 密钥框添加复制功能
  const apiKeyInput = document.querySelector('input[type="password"]');
  const apiKeyContainer = apiKeyInput?.closest('.relative');

  if (apiKeyInput && apiKeyContainer) {
    // 创建复制按钮
    const copyBtn = document.createElement('button');
    copyBtn.type = 'button';
    copyBtn.className = 'absolute right-12 top-2.5 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors p-1';
    copyBtn.innerHTML = '<span class="material-icons-outlined text-lg">content_copy</span>';
    copyBtn.title = '复制 API Key';

    copyBtn.addEventListener('click', async (e) => {
      e.preventDefault();
      e.stopPropagation();

      const apiKey = apiKeyInput.value;
      if (apiKey) {
        await ClipboardManager.copy(apiKey, copyBtn);
      } else {
        ToastManager.show('API Key 为空', 'error', 2000);
      }
    });

    apiKeyContainer.appendChild(copyBtn);
  }

  // 为可见性按钮添加功能
  const visibilityBtn = apiKeyContainer?.querySelector('button');
  if (visibilityBtn && apiKeyInput) {
    visibilityBtn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();

      if (apiKeyInput.type === 'password') {
        apiKeyInput.type = 'text';
        visibilityBtn.innerHTML = '<span class="material-icons-outlined text-lg">visibility</span>';
      } else {
        apiKeyInput.type = 'password';
        visibilityBtn.innerHTML = '<span class="material-icons-outlined text-lg">visibility_off</span>';
      }
    });
  }

  // 添加"导出配置"按钮
  addConfigExportButton();
}

// ============================================================================
// 5. 配置导出（一键复制配置代码）
// ============================================================================

function addConfigExportButton() {
  const modelSelect = document.querySelector('select');
  const temperatureSlider = document.querySelector('input[type="range"]');
  const tokenInput = document.querySelector('input[type="number"]');

  if (!modelSelect || !temperatureSlider || !tokenInput) return;

  // 在"保存配置"按钮上方添加"导出配置"按钮
  const configSection = document.querySelector('section:last-of-type');
  const saveButtonContainer = configSection?.querySelector('.flex.justify-end');

  if (saveButtonContainer) {
    const exportBtn = document.createElement('button');
    exportBtn.type = 'button';
    exportBtn.className = 'bg-blue-500 hover:bg-blue-600 text-white px-5 py-2.5 rounded-full text-sm font-medium transition-colors shadow-sm mr-3';
    exportBtn.innerHTML = '<span class="material-icons-outlined text-sm">file_download</span><span class="ml-2">导出配置</span>';

    exportBtn.addEventListener('click', async (e) => {
      e.preventDefault();

      const configCode = {
        model: modelSelect.value,
        temperature: parseFloat(temperatureSlider.value),
        maxTokens: parseInt(tokenInput.value),
      };

      const configJson = JSON.stringify(configCode, null, 2);

      // 复制到剪贴板
      await ClipboardManager.copy(configJson, exportBtn);

      // 显示导出内容
      console.log('导出的配置:', configJson);
    });

    saveButtonContainer.insertBefore(exportBtn, saveButtonContainer.firstChild);
  }
}
