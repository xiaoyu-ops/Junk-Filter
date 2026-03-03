# 真实 RSS 源和搜索功能实现方案

**日期**: 2026-03-01
**状态**: 实现规划和代码准备

---

## 需求确认

### 搜索功能需求

**场景**：用户想在主页搜索 Bilibili（或其他平台）的某个博主

**流程**：
```
用户输入: "某博主名称" + 选择平台 "Bilibili"
    ↓
后端搜索 RSS 储备库
    ↓
如果找到对应博主的 RSS 源 → 显示"可订阅" + 订阅按钮
如果未找到 → 显示"目前不能订阅该博主，请检查 RSS 源库"
```

**实现思路**：
1. 前端发送搜索请求：`GET /api/sources/search?query={keyword}&platform={platform}`
2. 后端在 sources 表中搜索
3. 返回匹配的源列表
4. 前端根据结果决定显示"可订阅"或"不能订阅"

---

## 第一步：获取您的真实 RSS 源

### 需要的信息格式

请按照以下 JSON 格式提供您的真实 RSS 源清单：

```json
[
  {
    "name": "作者/源名称",
    "url": "https://example.com/rss.xml",
    "platform": "blog|bilibili|twitter|medium|youtube|email",
    "priority": 1-10,
    "fetch_interval_seconds": 3600,
    "description": "源的描述"
  },
  {
    "name": "某 Bilibili 博主",
    "url": "https://api.bilibili.com/x/space/acc/feeds?mid=123456&ps=30&pn=1",
    "platform": "bilibili",
    "priority": 8,
    "fetch_interval_seconds": 1800,
    "description": "Bilibili UP主的视频更新"
  }
]
```

### 字段说明

| 字段 | 类型 | 是否必须 | 说明 |
|------|------|--------|------|
| `name` | string | ✅ | 博主/源的名称，用于显示 |
| `url` | string | ✅ | RSS Feed URL 或 API 端点 |
| `platform` | string | ✅ | 平台类型（blog, bilibili, twitter, medium, youtube, email） |
| `priority` | number | ✅ | 优先级 1-10，数字越大优先级越高 |
| `fetch_interval_seconds` | number | ✅ | 抓取间隔（秒），Bilibili 通常 1800（30分钟） |
| `description` | string | ❌ | 源的描述（可选） |

### 示例数据

```json
[
  {
    "name": "技术博客 - Python 深度学习",
    "url": "https://example-blog.com/feed.xml",
    "platform": "blog",
    "priority": 8,
    "fetch_interval_seconds": 3600,
    "description": "深度学习和 Python 开发教程"
  },
  {
    "name": "某 Bilibili UP主 - 编程教学",
    "url": "https://api.bilibili.com/x/space/acc/feeds?mid=456789",
    "platform": "bilibili",
    "priority": 9,
    "fetch_interval_seconds": 1800,
    "description": "编程、算法、数据结构教学"
  },
  {
    "name": "Medium - 技术分享",
    "url": "https://medium.com/feed/@author",
    "platform": "medium",
    "priority": 6,
    "fetch_interval_seconds": 3600,
    "description": "技术文章分享"
  },
  {
    "name": "Twitter - 技术新闻",
    "url": "https://twitter.com/user/feed",
    "platform": "twitter",
    "priority": 5,
    "fetch_interval_seconds": 1800,
    "description": "最新技术资讯"
  }
]
```

---

## 第二步：实现搜索 API（Go 后端）

### 添加搜索处理器

**文件**: `backend-go/handlers/source_handler.go`

在文件末尾添加以下代码：

```go
// SearchSources 搜索 RSS 源
// GET /api/sources/search?query={keyword}&platform={platform}
func (sh *SourceHandler) SearchSources(c *gin.Context) {
	query := c.Query("query")
	platform := c.Query("platform")  // 可选

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 调用仓储层进行搜索
	sources, err := sh.sourceRepo.Search(ctx, query, platform)
	if err != nil {
		log.Printf("Error searching sources: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search sources"})
		return
	}

	// 转换为响应格式
	var responses []map[string]interface{}
	for _, source := range sources {
		responses = append(responses, map[string]interface{}{
			"id":                     source.ID,
			"name":                   source.AuthorName,
			"url":                    source.URL,
			"platform":               source.Platform,
			"priority":               source.Priority,
			"fetch_interval_seconds": source.FetchIntervalSeconds,
			"enabled":                source.Enabled,
			"last_fetch_time":        source.LastFetchTime,
			"created_at":             source.CreatedAt,
			"updated_at":             source.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, responses)
}
```

### 添加搜索仓储方法

**文件**: `backend-go/repositories/source_repo.go`

在文件末尾添加以下代码：

```go
// Search 搜索 RSS 源
// 根据 query 搜索源名称和 URL，可选按 platform 筛选
func (sr *SourceRepository) Search(ctx context.Context, query, platform string) ([]models.Source, error) {
	var sources []models.Source

	// 构建查询
	sqlQuery := `
		SELECT id, platform, url, author_name, priority,
		       fetch_interval_seconds, enabled, last_fetch_time,
		       created_at, updated_at
		FROM sources
		WHERE (author_name ILIKE $1 OR url ILIKE $1)
	`
	args := []interface{}{"%" + query + "%"}

	// 如果指定了 platform，添加过滤条件
	if platform != "" {
		sqlQuery += " AND platform = $2"
		args = append(args, platform)
	}

	sqlQuery += " ORDER BY priority DESC, created_at DESC"

	rows, err := sr.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var source models.Source
		err := rows.Scan(
			&source.ID,
			&source.Platform,
			&source.URL,
			&source.AuthorName,
			&source.Priority,
			&source.FetchIntervalSeconds,
			&source.Enabled,
			&source.LastFetchTime,
			&source.CreatedAt,
			&source.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return sources, nil
}
```

### 注册搜索路由

**文件**: `backend-go/handlers/routes.go`

修改 `RegisterSourceRoutes` 函数，添加搜索路由：

```go
func RegisterSourceRoutes(router *gin.Engine, handler *SourceHandler) {
	sources := router.Group("/api/sources")
	{
		sources.GET("", handler.ListSources)              // 列出所有源
		sources.GET("/search", handler.SearchSources)     // ✅ 添加搜索路由（必须在 /:id 之前）
		sources.GET("/:id", handler.GetSource)            // 获取单个源
		sources.POST("", handler.CreateSource)            // 创建源
		sources.PUT("/:id", handler.UpdateSource)         // 更新源
		sources.DELETE("/:id", handler.DeleteSource)      // 删除源
	}
}
```

**⚠️ 重要**：搜索路由必须注册在 `/:id` 之前，否则 `/search` 会被当作 ID 处理。

### 修复 FetchSourceNow 路由（顺便）

**文件**: `backend-go/handlers/routes.go`

添加 `/fetch` 路由：

```go
func RegisterSourceRoutes(router *gin.Engine, handler *SourceHandler) {
	sources := router.Group("/api/sources")
	{
		sources.GET("", handler.ListSources)
		sources.GET("/search", handler.SearchSources)      // 搜索
		sources.POST("/:id/fetch", handler.FetchSourceNow) // ✅ 添加这行
		sources.GET("/:id", handler.GetSource)
		sources.POST("", handler.CreateSource)
		sources.PUT("/:id", handler.UpdateSource)
		sources.DELETE("/:id", handler.DeleteSource)
	}
}
```

---

## 第三步：前端搜索功能实现

### 修改 Home.vue 搜索处理

**文件**: `frontend-vue/src/components/Home.vue`

修改 `handleSearch` 函数（第 263 行）：

```javascript
const handleSearch = async (query) => {
  // 验证输入
  if (!query || typeof query !== 'string' || query.trim().length === 0) {
    console.warn('⚠️ 搜索词为空或无效，取消搜索')
    return
  }

  console.log(`🔍 搜索启动: "${query}" (平台: ${selectedPlatform.value})`)

  // 激活搜索态
  isSearching.value = true
  isPlatformMenuOpen.value = false

  try {
    // ✅ 调用真实搜索 API
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

    // 构建搜索参数
    const params = new URLSearchParams()
    params.append('query', query)

    // 如果不是"All Platforms"，添加 platform 参数
    if (selectedPlatform.value !== 'All Platforms') {
      const platformMap = {
        'Twitter': 'twitter',
        'YouTube': 'youtube',
        'Bilibili': 'bilibili',
        'RSS': 'blog',
        'Medium': 'medium',
        'Email': 'email',
      }
      const platformParam = platformMap[selectedPlatform.value]
      if (platformParam) {
        params.append('platform', platformParam)
      }
    }

    const response = await fetch(
      `${apiUrl}/sources/search?${params.toString()}`,
      {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      }
    )

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const results = await response.json()

    // ✅ 转换 Go 后端的 Source 格式为前端期望的格式
    searchResults.value = results.map(source => ({
      id: source.id,
      name: source.name,                    // author_name
      username: `@${source.platform}`,      // 平台@用户
      followers: `优先级 ${source.priority}`, // 用优先级替代粉丝数
      avatar: getAvatarByPlatform(source.platform),  // 根据平台获取头像
      status: source.enabled ? 'Active' : 'Inactive',
      statusColor: source.enabled ? 'green' : 'gray',
      icon: getIconByPlatform(source.platform),
      isSubscribed: false,                  // 暂未实现订阅状态
      _originalData: source,                // 保留原始数据用于订阅
    }))

    console.log(`✅ 搜索完成，找到 ${searchResults.value.length} 个结果`)
  } catch (error) {
    console.error('❌ 搜索失败:', error)
    searchResults.value = []

    // 显示错误提示
    const { show: showToast } = useToast()
    showToast(`搜索失败: ${error.message}`, 'error', 3000)
  }
}

// ✅ 根据平台获取头像
const getAvatarByPlatform = (platform) => {
  const avatars = {
    'bilibili': 'https://static.bilibili.com/upload/web_platform/logo.png',
    'twitter': 'https://abs.twimg.com/sticky/twitter_logo_blue.png',
    'youtube': 'https://www.youtube.com/favicon.ico',
    'medium': 'https://miro.medium.com/v2/resize:fit:96/1*eKHqpKVow6GM1zGyAd4mzg.png',
    'blog': 'https://www.google.com/favicon.ico',
    'email': 'https://fonts.gstatic.com/s/i/productlogos/mail_2020q4/v8/web-48dp/logo_gmail_colorful_circles_ios.png',
  }
  return avatars[platform] || 'https://www.google.com/favicon.ico'
}

// ✅ 根据平台获取图标
const getIconByPlatform = (platform) => {
  const icons = {
    'bilibili': 'ondemand_video',
    'twitter': 'language',
    'youtube': 'play_circle',
    'medium': 'article',
    'blog': 'rss_feed',
    'email': 'mail',
  }
  return icons[platform] || 'language'
}
```

### 添加 useToast 导入

确保在 Home.vue 的 script 部分添加导入：

```javascript
import { useToast } from '@/composables/useToast'
```

### 修改订阅按钮逻辑

**文件**: `frontend-vue/src/components/Home.vue`

修改 `toggleSubscribe` 函数，实现实际的订阅功能：

```javascript
const toggleSubscribe = async (resultId) => {
  const result = searchResults.value.find(r => r.id === resultId)
  if (!result) return

  // 如果原本没有订阅，执行订阅操作
  if (!result.isSubscribed) {
    try {
      const { show: showToast } = useToast()

      // 调用 configStore 的订阅方法（后续实现）
      // 或直接调用 API 创建源
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

      const response = await fetch(
        `${apiUrl}/sources`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            url: result._originalData.url,
            author_name: result.name,
            platform: result._originalData.platform,
            priority: result._originalData.priority,
            fetch_interval_seconds: result._originalData.fetch_interval_seconds,
          })
        }
      )

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      result.isSubscribed = true
      showToast(`✓ 已订阅: ${result.name}`, 'success', 2000)
      console.log(`✓ 已订阅: ${result.name}`)
    } catch (error) {
      const { show: showToast } = useToast()
      showToast(`订阅失败: ${error.message}`, 'error', 2000)
      console.error('Subscribe error:', error)
    }
  } else {
    // 取消订阅
    try {
      const { show: showToast } = useToast()
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

      const response = await fetch(
        `${apiUrl}/sources/${result.id}`,
        { method: 'DELETE' }
      )

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      result.isSubscribed = false
      showToast(`✗ 已取消订阅: ${result.name}`, 'success', 2000)
      console.log(`✗ 已取消订阅: ${result.name}`)
    } catch (error) {
      const { show: showToast } = useToast()
      showToast(`取消订阅失败: ${error.message}`, 'error', 2000)
      console.error('Unsubscribe error:', error)
    }
  }
}
```

---

## 第四步：初始化真实 RSS 源到数据库

### 方式 1：使用 SQL 脚本（推荐）

创建文件：`backend-go/sql/init_sources.sql`

```sql
-- 插入真实 RSS 源
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled, created_at, updated_at) VALUES
  ('blog', 'https://example1.com/feed.xml', '技术博客 - Python 深度学习', 8, 3600, true, NOW(), NOW()),
  ('bilibili', 'https://api.bilibili.com/x/space/acc/feeds?mid=456789', 'Bilibili UP主 - 编程教学', 9, 1800, true, NOW(), NOW()),
  ('medium', 'https://medium.com/feed/@author', 'Medium - 技术分享', 6, 3600, true, NOW(), NOW()),
  ('twitter', 'https://twitter.com/user/feed', 'Twitter - 技术新闻', 5, 1800, true, NOW(), NOW());
```

### 方式 2：使用 Go 初始化脚本

**文件**: `backend-go/main.go`

在 `main()` 函数中添加初始化逻辑（仅在首次运行时执行）：

```go
// 在 initDatabase() 之后添加
if err := initDefaultSources(ctx, sourceRepo); err != nil {
    log.Printf("Warning: Failed to initialize default sources: %v", err)
}

// 添加新函数
func initDefaultSources(ctx context.Context, sourceRepo *repositories.SourceRepository) error {
    // 检查是否已初始化（count > 0）
    count, err := sourceRepo.CountAll(ctx)
    if err != nil || count > 0 {
        return nil  // 已经有源了，不重复初始化
    }

    defaultSources := []models.CreateSourceRequest{
        {
            Platform:             "blog",
            URL:                  "https://example1.com/feed.xml",
            AuthorName:           "技术博客 - Python 深度学习",
            Priority:             8,
            FetchIntervalSeconds: 3600,
        },
        {
            Platform:             "bilibili",
            URL:                  "https://api.bilibili.com/x/space/acc/feeds?mid=456789",
            AuthorName:           "Bilibili UP主 - 编程教学",
            Priority:             9,
            FetchIntervalSeconds: 1800,
        },
        // ... 更多源
    }

    for _, req := range defaultSources {
        _, err := sourceRepo.Create(ctx, &req)
        if err != nil {
            log.Printf("Error creating default source %s: %v", req.AuthorName, err)
        }
    }

    return nil
}
```

### 方式 3：使用 API 插入（最灵活）

启动后端后，使用 curl 或 Postman：

```bash
curl -X POST http://localhost:8080/api/sources \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/feed.xml",
    "author_name": "技术博客",
    "platform": "blog",
    "priority": 8,
    "fetch_interval_seconds": 3600
  }'
```

---

## 第五步：测试完整流程

### 测试清单

- [ ] 1. 启动 Go 后端：`cd backend-go && go run main.go`
- [ ] 2. 验证数据库中有真实 RSS 源：`SELECT count(*) FROM sources;`
- [ ] 3. 启动前端：`cd frontend-vue && npm run dev`
- [ ] 4. 在首页搜索框输入关键词（如 "编程"、"技术"）
- [ ] 5. 验证搜索结果正确显示真实源
- [ ] 6. 点击"Subscribe"按钮
- [ ] 7. 验证源被添加到配置页面的 RSS 源表中
- [ ] 8. 查看后端日志，验证源的优先级和抓取间隔正确
- [ ] 9. 在配置页面手动同步某个源，验证日志显示

### 预期行为

**场景 1：搜索存在的博主**
```
用户输入: "编程教学"
→ 后端查询 sources 表
→ 找到 "Bilibili UP主 - 编程教学"
→ 前端显示结果，包含 "Subscribe" 按钮
→ 用户点击 "Subscribe"
→ 源被添加到 sources 表
✅ 完成
```

**场景 2：搜索不存在的博主**
```
用户输入: "不存在的博主"
→ 后端查询返回空数组 []
→ 前端显示 "No results found for '不存在的博主'"
✅ 完成
```

**场景 3：按平台筛选搜索**
```
用户输入: "教学" + 选择 "Bilibili"
→ 后端查询: WHERE author_name ILIKE '%教学%' AND platform='bilibili'
→ 只返回 Bilibili 平台的源
✅ 完成
```

---

## 总结

### 需要您做的

1. **提供真实 RSS 源清单**
   - 格式：JSON 数组
   - 包含：name, url, platform, priority, fetch_interval_seconds
   - 最好 5-10 个不同平台的源

2. **确认搜索范围**
   - 搜索的是源的 `author_name` 和 `url` 吗？
   - 是否支持模糊搜索（ILIKE）？

3. **确认订阅流程**
   - 用户订阅后，是直接在 sources 表创建新记录吗？
   - 是否需要检查重复订阅？

### 我们要做的

1. ✅ 实现 Go 后端搜索 API
2. ✅ 前端集成搜索功能
3. ✅ 初始化真实 RSS 源
4. ✅ 测试完整流程

---

## 代码文件清单

### Go 后端修改
- `backend-go/handlers/source_handler.go` - 添加 SearchSources 处理器
- `backend-go/repositories/source_repo.go` - 添加 Search 方法
- `backend-go/handlers/routes.go` - 添加搜索路由和 fetch 路由
- `backend-go/sql/init_sources.sql` - 初始化脚本（可选）

### 前端修改
- `frontend-vue/src/components/Home.vue` - 实现真实搜索和订阅逻辑

### 文档
- 本文件：`description/SEARCH_IMPLEMENTATION_GUIDE.md`
