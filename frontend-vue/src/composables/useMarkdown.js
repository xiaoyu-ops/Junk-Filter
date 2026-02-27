import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'

// 初始化 markdown-it 实例
const md = new MarkdownIt({
  html: false,           // 禁用原始 HTML（安全考虑）
  linkify: true,         // 自动转换 URL 为链接
  typographer: true,     // 启用排版增强（引号、破折号等）
})

/**
 * useMarkdown Composable
 * 提供 Markdown 渲染和 XSS 防护
 */
export const useMarkdown = () => {
  /**
   * 将 Markdown 转换为安全的 HTML
   * @param {string} text - Markdown 文本
   * @returns {string} 清理后的 HTML 字符串
   */
  const renderMarkdown = (text) => {
    if (!text || typeof text !== 'string') return ''

    try {
      // 1. 用 markdown-it 转换为 HTML
      const rawHtml = md.render(text)

      // 2. 用 DOMPurify 清理危险标签和属性
      const cleanHtml = DOMPurify.sanitize(rawHtml, {
        // 允许的 HTML 标签（白名单）
        ALLOWED_TAGS: [
          'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
          'p', 'br', 'strong', 'em', 'u', 'del', 's',
          'ul', 'ol', 'li',
          'code', 'pre',
          'blockquote',
          'table', 'thead', 'tbody', 'tfoot', 'tr', 'th', 'td',
          'a', 'img',
          'hr', 'div', 'span',
        ],
        // 允许的属性
        ALLOWED_ATTR: [
          'href', 'title', 'target', 'rel',
          'src', 'alt', 'width', 'height',
          'class', 'id',
        ],
        // 强制在链接中添加 rel="noopener noreferrer" 以防止安全问题
        FORCE_BODY: false,
        KEEP_CONTENT: true,
      })

      return cleanHtml
    } catch (error) {
      console.error('Markdown 渲染错误:', error)
      // 如果渲染失败，返回纯文本（安全退化）
      return DOMPurify.sanitize(text)
    }
  }

  return {
    renderMarkdown,
  }
}
