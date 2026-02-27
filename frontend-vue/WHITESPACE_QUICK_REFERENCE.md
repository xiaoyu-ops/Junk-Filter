# 🚨 白屏问题 - 快速参考卡

**问题**：修改代码后页面白屏或加载失败
**解决**：3 步 5 分钟

---

## 🔥 立即尝试（最可能有效）

```bash
cd /d/TrueSignal/frontend-vue
npm install
npm run dev
```

页面应该在 3-5 分钟内恢复。

---

## 📋 若上述方案无效

### 方案 A: 完整清理
```bash
npm cache clean --force
rm -rf node_modules package-lock.json .vite
npm install
npm run dev
```

### 方案 B: 浏览器缓存
```
按 F12 打开开发者工具
右键点击地址栏刷新按钮
选择 "清空缓存并硬刷新"
```

### 方案 C: 使用修复脚本
```
双击 fix-whitespace.bat（自动执行所有修复）
```

---

## ✅ 验证成功

看到以下表现 = ✅ 修复成功

- [ ] 页面显示内容（不是空白）
- [ ] 导航栏可见
- [ ] 搜索框完整
- [ ] 下拉菜单可以打开/关闭
- [ ] Console 无红色错误

---

## 📍 可能的问题及快速解决

| 症状 | 原因 | 解决 |
|-----|-----|------|
| 页面完全空白 | @vueuse/core 缺失 | `npm install` |
| 页面有内容但无样式 | Tailwind 未加载 | Ctrl+Shift+R |
| 功能不工作 | 缓存问题 | 删除 `.vite` |
| Console 有错误 | 代码问题 | 查看错误信息 |

---

## 📞 需要帮助？

查看完整诊断文档：
- `WHITESPACE_COMPLETE_DIAGNOSIS.md`（完整排查指南）
- `WHITESPACE_FIX_GUIDE.md`（详细步骤）

---

**99% 的情况下，运行 `npm install` 就能解决问题！**
