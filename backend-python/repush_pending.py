"""
将 DB 中 PENDING 状态的内容重新推入 Redis Stream
"""
import asyncio
import json
import asyncpg
import redis.asyncio as aioredis
from config import settings

async def repush():
    db = await asyncpg.connect(
        host=settings.db_host,
        port=settings.db_port,
        user=settings.db_user,
        password=settings.db_password,
        database=settings.db_name,
    )
    redis = aioredis.from_url(settings.redis_url, decode_responses=True)

    rows = await db.fetch(
        """SELECT id, task_id, title, original_url, clean_content,
                  published_at, platform, author_name, content_hash
           FROM content WHERE status = 'PENDING'"""
    )

    count = 0
    for row in rows:
        message = {
            "content_id": row["id"],
            "task_id": str(row["task_id"]),
            "title": row["title"] or "",
            "url": row["original_url"] or "",
            "content": row["clean_content"] or "",
            "published_at": row["published_at"].isoformat() if row["published_at"] else "",
            "platform": row["platform"] or "blog",
            "author_name": row["author_name"] or "",
            "content_hash": row["content_hash"] or "",
        }
        await redis.xadd("ingestion_queue", {"data": json.dumps(message, ensure_ascii=False)})
        count += 1
        print(f"  Pushed: [{row['id']}] {row['title'][:40]}")

    print(f"\n✓ Pushed {count} messages to ingestion_queue")
    await db.close()
    await redis.aclose()

asyncio.run(repush())
