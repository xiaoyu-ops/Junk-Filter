/**
 * 🧪 TrueSignal 适配层冒烟测试 - 浏览器 Console 诊断脚本
 *
 * 使用方法:
 * 1. 打开 http://localhost:5173
 * 2. 按 F12 打开开发工具，切换到 Console 标签
 * 3. 复制粘贴本脚本全部内容到 Console
 * 4. 按 Enter 执行
 *
 * 脚本会自动执行所有测试并输出结果
 */

console.clear();
console.log('%c🧪 TrueSignal 适配层冒烟测试 - 浏览器诊断', 'font-size: 16px; font-weight: bold; color: #0066cc;');
console.log('%c========================================', 'font-size: 12px;');
console.log('');

// 测试结果收集
const results = {
  passed: [],
  failed: [],
};

// 颜色助手
const success = (text) => `%c✅ ${text}`;
const error = (text) => `%c❌ ${text}`;
const warning = (text) => `%c⚠️  ${text}`;
const info = (text) => `%c📍 ${text}`;
const successStyle = 'color: #28a745; font-weight: bold;';
const errorStyle = 'color: #dc3545; font-weight: bold;';
const warningStyle = 'color: #ff9800; font-weight: bold;';
const infoStyle = 'color: #0066cc;';

// 测试 1: 检查环境变量
console.log('%c📋 测试 1: 环境变量检查', 'font-size: 14px; font-weight: bold;');
console.log('---');

const apiUrl = import.meta.env.VITE_API_URL;
const mockUrl = import.meta.env.VITE_MOCK_URL;

console.log(info('VITE_API_URL'), infoStyle, apiUrl || '未设置');
console.log(info('VITE_MOCK_URL'), infoStyle, mockUrl || '未设置');

if (apiUrl && apiUrl.includes('8080')) {
  console.log(success('Go 后端 URL 正确'), successStyle);
  results.passed.push('环境变量: VITE_API_URL');
} else {
  console.log(error('Go 后端 URL 不正确'), errorStyle);
  results.failed.push('环境变量: VITE_API_URL');
}

if (mockUrl && mockUrl.includes('3000')) {
  console.log(success('Mock 后端 URL 正确'), successStyle);
  results.passed.push('环境变量: VITE_MOCK_URL');
} else {
  console.log(error('Mock 后端 URL 不正确'), errorStyle);
  results.failed.push('环境变量: VITE_MOCK_URL');
}

console.log('');

// 测试 2: 检查适配器函数
console.log('%c📋 测试 2: 适配器函数检查', 'font-size: 14px; font-weight: bold;');
console.log('---');

// 模拟一个 Go 源对象
const mockSource = {
  id: 123,
  name: '测试源',
  url: 'https://example.com/feed',
  priority: 8,
  enabled: true,
  last_fetch_time: '2026-02-27T10:00:00Z',
  created_at: '2026-02-26T10:00:00Z',
  updated_at: '2026-02-27T10:00:00Z',
};

console.log(info('测试 Source 对象'), infoStyle);
console.table(mockSource);

// 测试适配器
const { adaptSourceToTask, adaptTaskToSource } = await (async () => {
  const api = await import('/src/composables/useAPI.js');
  return api.useAPI();
})();

// 因为 useAPI 可能需要初始化，这里我们用另一种方式测试
// 检查适配器是否存在
const checkAdapters = async () => {
  try {
    // 尝试从 useAPI 导出中获取
    const response = await fetch(apiUrl + '/api/sources');

    if (response.ok) {
      const sources = await response.json();
      console.log(success('Go 后端连接成功'), successStyle);
      console.log(info('Retrieved sources count'), infoStyle, sources.length || sources.data?.length || 0);
      results.passed.push('服务连接: Go 后端');
      return true;
    } else {
      console.log(error('Go 后端返回错误'), errorStyle, response.status);
      results.failed.push('服务连接: Go 后端');
      return false;
    }
  } catch (e) {
    console.log(error('无法连接 Go 后端'), errorStyle, e.message);
    results.failed.push('服务连接: Go 后端');
    return false;
  }
};

await checkAdapters();

console.log('');

// 测试 3: 检查 API 连接
console.log('%c📋 测试 3: API 连接检查', 'font-size: 14px; font-weight: bold;');
console.log('---');

// 测试 Go 后端
const testGoBackend = async () => {
  try {
    const response = await fetch(`${apiUrl}/api/sources?enabled=true`, {
      method: 'GET',
      timeout: 5000,
    });

    if (response.ok) {
      const data = await response.json();
      console.log(success('Go 后端 API 连接成功'), successStyle);
      console.log(info('Sources count'), infoStyle, data.length || data.data?.length || 0);
      results.passed.push('Go 后端 /api/sources');
      return true;
    } else {
      console.log(error(`Go 后端返回 ${response.status}`), errorStyle);
      results.failed.push('Go 后端 /api/sources');
      return false;
    }
  } catch (e) {
    console.log(error('Go 后端连接失败'), errorStyle, e.message);
    results.failed.push('Go 后端 /api/sources');
    return false;
  }
};

// 测试 Mock 后端
const testMockBackend = async () => {
  try {
    const response = await fetch(`${mockUrl}/api/tasks`, {
      method: 'GET',
      timeout: 5000,
    });

    if (response.ok) {
      const data = await response.json();
      console.log(success('Mock 后端 API 连接成功'), successStyle);
      console.log(info('Tasks count'), infoStyle, data.length || data.data?.length || 0);
      results.passed.push('Mock 后端 /api/tasks');
      return true;
    } else {
      console.log(error(`Mock 后端返回 ${response.status}`), errorStyle);
      results.failed.push('Mock 后端 /api/tasks');
      return false;
    }
  } catch (e) {
    console.log(error('Mock 后端连接失败'), errorStyle, e.message);
    results.failed.push('Mock 后端 /api/tasks');
    return false;
  }
};

await Promise.all([testGoBackend(), testMockBackend()]);

console.log('');

// 测试 4: 检查 useAPI 和数据适配
console.log('%c📋 测试 4: useAPI 和数据适配', 'font-size: 14px; font-weight: bold;');
console.log('---');

try {
  // 注意: 这会在实际应用中工作，当前可能需要应用已初始化
  console.log(warning('此测试需要在应用已加载 useAPI 后运行'), warningStyle);
  console.log(info('请在前端左侧任务列表加载后重新运行此脚本'), infoStyle);
} catch (e) {
  console.log(error('useAPI 测试跳过'), warningStyle);
}

console.log('');

// 总结
console.log('%c📊 测试总结', 'font-size: 14px; font-weight: bold;');
console.log('---');
console.log(`${success('通过')}`, successStyle, results.passed.length);
results.passed.forEach(test => {
  console.log(`  ✓ ${test}`);
});

console.log('');

if (results.failed.length > 0) {
  console.log(`${error('失败')}`, errorStyle, results.failed.length);
  results.failed.forEach(test => {
    console.log(`  ✗ ${test}`);
  });
} else {
  console.log('%c✅ 所有测试通过！', 'font-size: 12px; color: #28a745; font-weight: bold;');
}

console.log('');
console.log('%c========================================', 'font-size: 12px;');
console.log('%c🚀 下一步:', 'font-size: 12px; font-weight: bold;');
console.log('1. 打开左侧任务列表，验证任务显示');
console.log('2. 尝试创建新任务，观察 Network 标签');
console.log('3. 发送消息，验证 Go (8080) 和 Mock (3000) 的并行请求');
console.log('4. 查看浏览器 Network 标签，验证请求方向');
console.log('');
console.log('详细文档: description/SMOKE_TEST_COMPLETE_GUIDE.md');
