-- Migration: 添加 favicon_url 字段到 sources 表
ALTER TABLE sources ADD COLUMN IF NOT EXISTS favicon_url VARCHAR(500);

-- 根据 platform 预填 favicon URL（使用 Google Favicon API 作为备用）
UPDATE sources SET favicon_url = CASE
  WHEN platform = 'github' THEN 'https://github.githubassets.com/favicons/favicon.svg'
  WHEN platform = 'weibo' THEN 'https://weibo.com/favicon.ico'
  WHEN platform = 'zhihu' THEN 'https://static.zhihu.com/heifetz/favicon.ico'
  WHEN platform = 'bilibili' THEN 'https://www.bilibili.com/favicon.ico'
  WHEN platform = 'douyin' THEN 'https://www.douyin.com/favicon.ico'
  WHEN platform = 'douban' THEN 'https://www.douban.com/favicon.ico'
  WHEN platform = 'juejin' THEN 'https://lf3-cdn-tos.bytescm.com/obj/static/xitu_juejin_web/static/favicons/favicon-32x32.png'
  WHEN platform = 'csdn' THEN 'https://g.csdnimg.cn/static/logo/favicon32.ico'
  WHEN platform = 'toutiao' THEN 'https://www.toutiao.com/favicon.ico'
  WHEN platform = 'baidu' THEN 'https://www.baidu.com/favicon.ico'
  WHEN platform = '36kr' THEN 'https://36kr.com/favicon.ico'
  WHEN platform = 'hn' THEN 'https://news.ycombinator.com/favicon.ico'
  WHEN platform = 'smzdm' THEN 'https://www.smzdm.com/favicon.ico'
  WHEN platform = 'blog' AND url LIKE '%github%' THEN 'https://github.githubassets.com/favicons/favicon.svg'
  WHEN platform = 'blog' AND url LIKE '%medium%' THEN 'https://miro.medium.com/v2/1*m-R_BkNf1Qjr1YbyOIJY2w.png'
  ELSE NULL
END
WHERE favicon_url IS NULL;
