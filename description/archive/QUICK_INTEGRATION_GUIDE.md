# 5-Minute Integration Guide

**Goal**: Get the dual-engine evaluation system running

---

## Step 1: Verify Dependencies (Python)

```bash
cd backend-python

# Check if langchain and langgraph are installed
pip list | grep -E "langchain|langgraph"

# If missing, install:
pip install langchain langchain-openai langgraph
```

## Step 2: Set Environment Variables

```bash
# For LLM evaluation (optional, falls back to rule-based if missing)
export OPENAI_API_KEY=sk-your-key-here

# Or for Claude:
export ANTHROPIC_API_KEY=sk-ant-your-key-here

# Or use .env file
echo "OPENAI_API_KEY=sk-..." > .env
source .env
```

## Step 3: Restart Services

```bash
# Terminal 1: Start Docker containers
docker-compose up -d

# Terminal 2: Start Go backend
cd backend-go && go run main.go

# Terminal 3: Start Python backend
cd backend-python && python main.py

# Terminal 4: Start Vue frontend
cd frontend-vue && npm run dev
```

## Step 4: Generate Test Data

```bash
# In a new terminal, run smoke test to create test articles
cd D:\TrueSignal
./smoke_test.bat  # Windows
# or
bash smoke_test.sh  # Linux/Mac
```

## Step 5: Open Timeline

```bash
# Open browser and navigate to
http://localhost:5173

# Click "Timeline" in navigation
# Should see loading spinner
# Then content cards with evaluation scores
```

## Step 6: Verify Evaluation

```bash
# Check database
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT id, title, status FROM content LIMIT 5;"

# Check evaluations
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT id, content_id, decision, innovation_score, depth_score FROM evaluation LIMIT 5;"
```

---

## Fallback Verification

### To test rule-based fallback:

```bash
# Method 1: Remove API key
unset OPENAI_API_KEY
python main.py
# Python will auto-use RuleBasedEvaluator

# Method 2: Use invalid API key
export OPENAI_API_KEY=invalid-key-test
python main.py
# Will trigger fallback after failed LLM call

# Method 3: Rate limit simulation
# ContentEvaluationAgent detects 429 and triggers fallback automatically
```

---

## Expected Behavior

### Timeline Page
- [ ] Shows loading spinner initially
- [ ] Displays cards with title, source, scores
- [ ] Click card → Detail drawer opens
- [ ] Scores: Innovation (0-10), Depth (0-10)
- [ ] Decision badge: Green (Interesting), Amber (Bookmark), Red (Skip)
- [ ] Refresh button works

### Detail Drawer
- [ ] Shows AI Summary (TLDR)
- [ ] Shows innovation_score and depth_score with progress bars
- [ ] Shows decision badge with reasoning
- [ ] Shows key concepts as tags
- [ ] Click "View Original" → Opens article in new tab

### API Response
```bash
curl http://localhost:8080/api/content?status=EVALUATED&limit=5

# Should return:
{
  "data": [
    {
      "id": 1,
      "title": "...",
      "evaluation": {
        "innovation_score": 8,
        "depth_score": 7,
        "decision": "INTERESTING",
        "tldr": "...",
        "evaluator_version": "dual-engine-llm"
      }
    }
  ],
  "count": 1
}
```

---

## Troubleshooting Quick Fixes

| Problem | Fix |
|---------|-----|
| Timeline shows empty | Run smoke_test to create articles |
| No evaluation data | Check Python backend is running |
| Scores are 0/0 | Wait for evaluator to process pending articles |
| API returns 500 | Check PostgreSQL: `docker-compose logs postgres` |
| Fallback not triggered | Check error logs: `python main.py` output |

---

## Next Steps

1. **Monitor Evaluation Quality**: Check rule-based vs LLM consistency
2. **Optimize Parameters**: Adjust temperature, top_p in config
3. **Add More RSS Sources**: Bulk test the system
4. **Review Decision Distribution**: Are scores reasonable?
5. **Production Deployment**: Configure proper monitoring and logging

---

## Key Files Modified

```
✅ backend-python/agents/content_evaluator.py
   - Dual-engine logic, fallback detection

✅ backend-go/handlers/content_handler.go
   - ListContent() now includes evaluation data

✅ frontend-vue/src/stores/useTimelineStore.js
   - Real API integration, no more mock data

✅ frontend-vue/src/components/Timeline.vue
   - Display scores, detail drawer enhancements
```

---

**Time to Complete**: ~5 minutes
**Next Action**: Open http://localhost:5173 and see Timeline in action!
