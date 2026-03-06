-- 预置 RSS 源候选池（enabled=false，不会出现在任务列表）
-- AI 助手创建任务时可查询这些源作为推荐
-- 来源：https://github.com/JackyST0/awesome-rsshub-routes
-- 使用 RSShub 公开服务 (https://rsshub.app)

-- 社交媒体与社区
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled) VALUES
('weibo', 'https://rsshub.app/weibo/search/hot', 'Weibo - 实时热搜', 9, 1800, false),
('weibo', 'https://rsshub.app/weibo/user/2803301701', 'Weibo - 官方账号', 8, 1800, false),
('zhihu', 'https://rsshub.app/zhihu/hot', 'Zhihu - 热门问题', 8, 3600, false),
('zhihu', 'https://rsshub.app/zhihu/daily', 'Zhihu - 每日精选', 7, 86400, false),
('douyin', 'https://rsshub.app/douyin/search/hot', 'Douyin - 热搜榜', 8, 1800, false)
ON CONFLICT (url) DO NOTHING;

-- 技术社区与开源
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled) VALUES
('github', 'https://rsshub.app/github/trending/daily', 'GitHub - 日趋势项目', 9, 3600, false),
('github', 'https://rsshub.app/github/trending/weekly', 'GitHub - 周趋势项目', 8, 86400, false),
('github', 'https://rsshub.app/github/trending/daily/python', 'GitHub - Python 趋势', 8, 3600, false),
('github', 'https://rsshub.app/github/trending/daily/javascript', 'GitHub - JavaScript 趋势', 8, 3600, false),
('juejin', 'https://rsshub.app/juejin/trending/all/weekly', 'Juejin - 全部热门', 8, 3600, false),
('juejin', 'https://rsshub.app/juejin/trending/frontend/weekly', 'Juejin - 前端热门', 8, 3600, false),
('juejin', 'https://rsshub.app/juejin/trending/backend/weekly', 'Juejin - 后端热门', 8, 3600, false),
('csdn', 'https://rsshub.app/csdn/hotrank', 'CSDN - 技术热文', 7, 3600, false)
ON CONFLICT (url) DO NOTHING;

-- 热榜与新闻
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled) VALUES
('toutiao', 'https://rsshub.app/toutiao/hot', 'Toutiao - 头条热搜', 8, 1800, false),
('baidu', 'https://rsshub.app/baidu/hot', 'Baidu - 百度热搜', 7, 1800, false),
('36kr', 'https://rsshub.app/36kr/newsflash', '36Kr - 快讯', 8, 3600, false),
('hn', 'https://rsshub.app/hackernews', 'Hacker News - 热门', 8, 3600, false)
ON CONFLICT (url) DO NOTHING;

-- 视频与影视
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled) VALUES
('bilibili', 'https://rsshub.app/bilibili/ranking/0/3/1', 'Bilibili - 全站日排行', 9, 3600, false),
('bilibili', 'https://rsshub.app/bilibili/search/keyword/编程教程', 'Bilibili - 编程教程搜索', 8, 3600, false),
('douban', 'https://rsshub.app/douban/movie/playing', 'Douban - 正在上映', 5, 86400, false),
('douban', 'https://rsshub.app/douban/movie/later', 'Douban - 即将上映', 5, 86400, false)
ON CONFLICT (url) DO NOTHING;

-- 购物与优惠
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled) VALUES
('smzdm', 'https://rsshub.app/smzdm/ranking/pinlei/11', 'SMZDM - 数码产品', 6, 3600, false),
('smzdm', 'https://rsshub.app/smzdm/ranking/pinlei/12', 'SMZDM - 电脑配件', 6, 3600, false)
ON CONFLICT (url) DO NOTHING;

-- 博客与个人网站
INSERT INTO sources (platform, url, author_name, priority, fetch_interval_seconds, enabled) VALUES
('blog', 'https://rsshub.app/github/blog', 'GitHub Blog', 6, 86400, false),
('blog', 'https://rsshub.app/medium/tag/programming', 'Medium - 编程', 6, 86400, false)
ON CONFLICT (url) DO NOTHING;
