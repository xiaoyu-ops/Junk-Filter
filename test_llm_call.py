#!/usr/bin/env python3
"""
直接测试 LLM API 调用，诊断问题
"""

from openai import OpenAI

# 使用你前端配置中的参数
API_KEY = "sk-DassY8jqfgrFcsgqzyDVpgOiBkqyaU0sHKsWmDoO6YVJsUIn"
BASE_URL = "https://elysiver.h-e.top/v1"
MODEL = "gemini-3-flash-preview"

print("=" * 60)
print("LLM API 直接测试")
print("=" * 60)
print(f"API Key: {API_KEY[:30]}...")
print(f"Base URL: {BASE_URL}")
print(f"Model: {MODEL}")
print("=" * 60)

try:
    print("\n1️⃣ 初始化 OpenAI 客户端...")
    client = OpenAI(
        api_key=API_KEY,
        base_url=BASE_URL,
        timeout=10
    )
    print("✅ 客户端初始化成功")
    print(f"   Base URL: {client.base_url}")
    print(f"   Timeout: {client.timeout}")

    print("\n2️⃣ 调用 API...")
    response = client.chat.completions.create(
        model=MODEL,
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "Hello, say something brief."}
        ],
        temperature=0.7,
        max_tokens=100,
    )
    print("✅ API 调用成功！")
    print(f"   Response: {response.choices[0].message.content[:100]}")

except Exception as e:
    print(f"\n❌ 错误发生：")
    print(f"   错误类型: {type(e).__name__}")
    print(f"   错误信息: {str(e)}")
    print(f"\n   完整错误信息:")
    import traceback
    traceback.print_exc()
