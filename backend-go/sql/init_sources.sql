-- 初始化真实 RSS 源数据（更新版本）
-- 来源：https://github.com/JackyST0/awesome-rsshub-routes
-- 使用 RSShub 公开服务 (https://rsshub.app)

-- 社交媒体与社区
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled, created_at, updated_at) VALUES
('weibo', 'https://rsshub.app/weibo/search/hot', 'Weibo - 实时热搜', 9, 1800, true, NOW(), NOW()),
('weibo', 'https://rsshub.app/weibo/user/2803301701', 'Weibo - 官方账号', 8, 1800, true, NOW(), NOW()),
('zhihu', 'https://rsshub.app/zhihu/hot', 'Zhihu - 热门问题', 8, 3600, true, NOW(), NOW()),
('zhihu', 'https://rsshub.app/zhihu/daily', 'Zhihu - 每日精选', 7, 86400, true, NOW(), NOW()),
('douyin', 'https://rsshub.app/douyin/search/hot', 'Douyin - 热搜榜', 8, 1800, true, NOW(), NOW()),

-- 技术社区与开源
('github', 'https://rsshub.app/github/trending/daily', 'GitHub - 日趋势项目', 9, 3600, true, NOW(), NOW()),
('github', 'https://rsshub.app/github/trending/weekly', 'GitHub - 周趋势项目', 8, 86400, true, NOW(), NOW()),
('github', 'https://rsshub.app/github/trending/daily/python', 'GitHub - Python 趋势', 8, 3600, true, NOW(), NOW()),
('github', 'https://rsshub.app/github/trending/daily/javascript', 'GitHub - JavaScript 趋势', 8, 3600, true, NOW(), NOW()),
('juejin', 'https://rsshub.app/juejin/trending/all/weekly', 'Juejin - 全部热门', 8, 3600, true, NOW(), NOW()),
('juejin', 'https://rsshub.app/juejin/trending/frontend/weekly', 'Juejin - 前端热门', 8, 3600, true, NOW(), NOW()),
('juejin', 'https://rsshub.app/juejin/trending/backend/weekly', 'Juejin - 后端热门', 8, 3600, true, NOW(), NOW()),
('csdn', 'https://rsshub.app/csdn/hotrank', 'CSDN - 技术热文', 7, 3600, true, NOW(), NOW()),

-- 热榜与新闻
('toutiao', 'https://rsshub.app/toutiao/hot', 'Toutiao - 头条热搜', 8, 1800, true, NOW(), NOW()),
('baidu', 'https://rsshub.app/baidu/hot', 'Baidu - 百度热搜', 7, 1800, true, NOW(), NOW()),
('36kr', 'https://rsshub.app/36kr/newsflash', '36Kr - 快讯', 8, 3600, true, NOW(), NOW()),
('hn', 'https://rsshub.app/hackernews', 'Hacker News - 热门', 8, 3600, true, NOW(), NOW()),

-- 视频与影视
('bilibili', 'https://rsshub.app/bilibili/ranking/0/3/1', 'Bilibili - 全站日排行', 9, 3600, true, NOW(), NOW()),
('bilibili', 'https://rsshub.app/bilibili/search/keyword/编程教程', 'Bilibili - 编程教程搜索', 8, 3600, true, NOW(), NOW()),
('douban', 'https://rsshub.app/douban/movie/playing', 'Douban - 正在上映', 5, 86400, true, NOW(), NOW()),
('douban', 'https://rsshub.app/douban/movie/later', 'Douban - 即将上映', 5, 86400, true, NOW(), NOW()),

-- 购物与优惠
('smzdm', 'https://rsshub.app/smzdm/ranking/pinlei/11', 'SMZDM - 数码产品', 6, 3600, true, NOW(), NOW()),
('smzdm', 'https://rsshub.app/smzdm/ranking/pinlei/12', 'SMZDM - 电脑配件', 6, 3600, true, NOW(), NOW()),

-- 博客与个人网站
('blog', 'https://rsshub.app/github/blog', 'GitHub Blog', 6, 86400, true, NOW(), NOW()),
('blog', 'https://rsshub.app/medium/tag/programming', 'Medium - 编程', 6, 86400, true, NOW(), NOW());
