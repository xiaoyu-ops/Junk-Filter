# LLM Integration Setup Guide

## Overview

TrueSignal's Agent chat feature now supports real LLM integration. When you ask the Agent questions about task progress, content rules, or evaluation decisions, it will use OpenAI's GPT models (or other LLM providers) to generate intelligent, context-aware responses instead of simple rule-based answers.

## What Changed

### Before
The Agent used basic keyword matching to respond to user queries:
- Ask "现在的执行进度如何？" → Get hardcoded statistics
- Ask anything else → Get generic help message with same 4 options

### Now
The Agent uses real LLM:
- Sends the user's question + full task context to OpenAI API
- Receives natural language response based on your actual task data
- Falls back to rule-based responses if LLM is unavailable

## Setup Instructions

### Step 1: Get OpenAI API Key

1. Visit [OpenAI API Keys](https://platform.openai.com/api-keys)
2. Sign up or log in to your OpenAI account
3. Click "Create new secret key"
4. Copy the key (looks like: `sk-proj-...`)

### Step 2: Update .env File

Open `D:\TrueSignal\.env` and replace the placeholder:

**Before:**
```env
OPENAI_API_KEY=sk-proj-replace-with-your-api-key
```

**After:**
```env
OPENAI_API_KEY=sk-proj-YOUR_ACTUAL_KEY_HERE
```

### Step 3: Restart Python Backend

Stop and restart the Python Backend to load the new API key:

```bash
# Kill the existing Python Backend process in the terminal window
# Then restart with:
uvicorn api_server:app --host 0.0.0.0 --port 8081
```

Or use the one-click startup script:
```bash
start-all.bat
```

### Step 4: Test

Open the frontend and ask the Agent a question. You should now get intelligent responses based on your actual task context!

## Configuration Options

All LLM settings are in `.env`:

```env
# Required: Your OpenAI API key
OPENAI_API_KEY=sk-proj-...

# Optional: Adjust the model (default: gpt-3.5-turbo)
LLM_MODEL_ID=gpt-4

# Optional: Response quality/creativity (0.0=deterministic, 1.0=creative)
LLM_TEMPERATURE=0.7

# Optional: Maximum response length (tokens)
LLM_MAX_TOKENS=2000

# Optional: Timeout for API calls (seconds)
LLM_TIMEOUT=30
```

## Fallback Behavior

If LLM is unavailable (API key not set, API error, network issue), the Agent will:
1. Log a warning message
2. Automatically fall back to rule-based responses
3. Continue functioning normally

This means the chat feature will **always** work, even without LLM setup.

## Cost Considerations

- **OpenAI API**: Usage-based pricing (~$0.0005 per 1K input tokens for gpt-3.5-turbo)
- **Estimated cost**: ~1-2¢ per chat interaction
- Set `LLM_MAX_TOKENS` lower to reduce costs

## Troubleshooting

### "LangChain/OpenAI not available"
**Solution**: Install dependencies
```bash
cd backend-python
pip install -r requirements.txt
```

### "API key not found"
**Solution**: Check `.env` file
```bash
# On Windows
type D:\TrueSignal\.env | find "OPENAI_API_KEY"

# On Linux/Mac
grep OPENAI_API_KEY D:\TrueSignal\.env
```

### "Invalid API key"
**Solution**: Verify key format
- Should start with `sk-proj-`
- Must be your active key (not revoked)
- Check [OpenAI dashboard](https://platform.openai.com/account/api-keys)

### Still getting rule-based responses
**Steps to debug**:
1. Check Python Backend logs for `"LLM API call failed"`
2. Verify API key in `.env` is correct
3. Check API key has available credits
4. Try with `LLM_TEMPERATURE=0.5` for more consistent results

## Supported Models

| Model | Cost | Speed | Quality |
|-------|------|-------|---------|
| gpt-3.5-turbo | $0.50/M input | ⚡⚡⚡ | ⭐⭐⭐ |
| gpt-4 | $3/M input | ⚡⚡ | ⭐⭐⭐⭐ |
| gpt-4-turbo | $0.01/1K input | ⚡⚡ | ⭐⭐⭐⭐ |

Default is `gpt-3.5-turbo` - good balance of cost and quality.

## Advanced: Use Other LLM Providers

The code supports plugging in other LLM providers. To implement:

1. Edit `backend-python/api_server.py` line ~410 (`_call_llm` function)
2. Add logic to detect provider and call appropriate API
3. Implement similar pattern for Claude, Anthropic, etc.

Example for future Claude integration:
```python
if settings.llm_provider == "anthropic":
    from langchain_anthropic import ChatAnthropic
    llm = ChatAnthropic(api_key=settings.anthropic_api_key)
```

## More Info

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [LangChain Documentation](https://docs.langchain.com/)
- Main API endpoint: `POST /api/task/{task_id}/chat`

