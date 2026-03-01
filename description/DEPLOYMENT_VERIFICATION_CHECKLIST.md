# ✅ Deployment Verification Checklist

## Pre-Deployment (Code Review)

- [x] **content_evaluator.py**
  - [x] Dual-engine logic implemented
  - [x] Fallback detection for 429, 500, auth errors
  - [x] Error handling and logging
  - [x] LangGraph state management

- [x] **content_handler.go**
  - [x] ListContent() returns evaluation data
  - [x] Response structure: {data: [...], count: N}
  - [x] Status filtering works

- [x] **useTimelineStore.js**
  - [x] Removed all Mock data
  - [x] Real API integration
  - [x] Data transformation logic
  - [x] Loading/error states
  - [x] Filter functionality

- [x] **Timeline.vue**
  - [x] Loading spinner
  - [x] Score display (innovation, depth)
  - [x] Decision badges (colored)
  - [x] Detail drawer with evaluation data
  - [x] Refresh button
  - [x] Empty state handling

---

## Deployment Sequence (In Order)

### Phase 1: Infrastructure (5 min)
```bash
# 1. Start Docker containers
cd D:\TrueSignal
docker-compose up -d

# 2. Verify containers running
docker-compose ps
# Expected:
#   truesignal-db    PostgreSQL ✓
#   truesignal-redis Redis ✓

# 3. Test database
docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT 1"
# Expected: Output: 1

# 4. Test Redis
docker exec truesignal-redis redis-cli ping
# Expected: PONG
```

### Phase 2: Backend Services (10 min)

#### Terminal 2 - Go Backend
```bash
cd backend-go
go run main.go

# Expected output:
# ✓ Configuration loaded
# ✓ Database connected
# ✓ Redis connected
# Listening on port 8080
```

#### Terminal 3 - Python Backend
```bash
cd backend-python

# Check API key (optional)
echo $OPENAI_API_KEY

# If not set, that's OK - will use rule-based evaluator

# Start Python backend
python main.py

# Expected output:
# [ContentEvaluationAgent] LLM initialized: gpt-4
# OR
# [ContentEvaluationAgent] No API key provided, using rule-based evaluator
```

#### Terminal 4 - Vue Frontend
```bash
cd frontend-vue
npm run dev

# Expected output:
# VITE v... ready in ... ms
# ➜ Local: http://localhost:5173/
# ➜ press h to show help
```

### Phase 3: Test Data Generation (5 min)

```bash
# Terminal 5 - Run smoke test
cd D:\TrueSignal
bash smoke_test.sh  # or smoke_test.bat on Windows

# Expected output:
# ========== Test 1: List all sources ==========
# [SUCCESS] Get sources working
# ... (8 tests, all should pass)
# [SUCCESS] All tests passed!
```

---

## Post-Deployment Verification

### Check 1: Database Content

```bash
# Check sources were created
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT id, name, url FROM sources LIMIT 3;"

# Expected: 3 rows with name and url

# Check content was created
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT id, title, status FROM content LIMIT 3;"

# Expected: Rows with status='PENDING' or 'EVALUATED'

# Check evaluation results exist
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT id, content_id, decision, innovation_score FROM evaluation LIMIT 3;"

# Expected: innovation_score between 0-10, decision in (INTERESTING, BOOKMARK, SKIP)
```

### Check 2: API Endpoints

```bash
# Test sources endpoint
curl -s http://localhost:8080/api/sources | jq '.[] | {id, name}' | head -10

# Test content endpoint with filter
curl -s "http://localhost:8080/api/content?status=EVALUATED&limit=5" | jq '.data[0]'

# Expected structure:
# {
#   "id": 1,
#   "title": "...",
#   "evaluation": {
#     "innovation_score": 8,
#     "depth_score": 7,
#     "decision": "INTERESTING",
#     "tldr": "...",
#     "evaluator_version": "dual-engine-llm"
#   }
# }
```

### Check 3: Frontend Functionality

```bash
# Open browser
http://localhost:5173

# Navigate to Timeline
# Click on "Timeline" in navigation menu

# Verify:
- [ ] Page shows loading spinner
- [ ] Cards appear with evaluation scores
- [ ] Click card → Detail drawer opens
- [ ] Drawer shows TLDR and reasoning
- [ ] Color-coded decision badge (green/amber/red)
- [ ] "View Original" button works
- [ ] Refresh button fetches new data
- [ ] Empty state shown if no articles evaluated yet
```

### Check 4: Dual-Engine Operation

#### Test LLM Engine (if API key set)
```bash
# Watch logs while processing
cd backend-python && python main.py

# Look for:
# [ContentEvaluationAgent] LLM initialized: gpt-4
# [ContentEvaluator] LLM evaluation successful - INTERESTING
# evaluator_version: dual-engine-llm
```

#### Test Fallback Engine
```bash
# Option A: Remove API key and restart
unset OPENAI_API_KEY
cd backend-python && python main.py

# Option B: Use invalid key
export OPENAI_API_KEY=invalid-test-key
cd backend-python && python main.py

# Look for:
# [ContentEvaluator] Fallback triggered - using rule-based evaluator
# [ContentEvaluator] Rule-based evaluation successful
# evaluator_version: dual-engine-rule_based
```

---

## Troubleshooting Guide

### Issue: Timeline shows "No evaluated content available"

**Causes & Solutions**:
1. No articles processed yet
   - Run smoke_test.sh to create articles
   - Wait 5-10 seconds for Python evaluator to process

2. Python backend not running
   - Check Terminal 3: `cd backend-python && python main.py`
   - Check for errors in output

3. Database not connected
   - Verify PostgreSQL running: `docker-compose ps`
   - Check database created: `docker exec truesignal-db psql -U truesignal -l`

### Issue: Scores are 0/0

**Causes & Solutions**:
1. Evaluation not created
   - Check Python logs for errors
   - Verify evaluation table has data: `SELECT COUNT(*) FROM evaluation;`

2. API not returning evaluation data
   - Test API directly: `curl http://localhost:8080/api/content?status=EVALUATED`
   - Check content_handler.go modification was applied

### Issue: Frontend shows error message

**Causes & Solutions**:
1. API unreachable
   - Verify Go backend running on port 8080
   - Check firewall rules

2. CORS error
   - Go backend should already have CORS enabled
   - Check main.go for CORS middleware

3. Invalid response format
   - Verify API returns: `{data: [...], count: N}`
   - Check content_handler.go ListContent() implementation

### Issue: Fallback not triggered when API key is invalid

**Causes & Solutions**:
1. Error message not being detected
   - Check exact error text from API
   - May need to add error string to fallback detection

2. LLM initialization succeeds but later fails
   - This is expected - fallback still triggers
   - Check for specific error codes (429, 500)

---

## Performance Benchmarks

| Operation | Expected | Actual |
|-----------|----------|--------|
| Database query (100 articles) | <300ms | ___ |
| API response with evaluation | <500ms | ___ |
| Frontend load (Timeline) | <2s | ___ |
| LLM evaluation | 5-20s | ___ |
| Rule-based evaluation | <100ms | ___ |

---

## Monitoring Commands

```bash
# Watch Python evaluator logs in real-time
cd backend-python && tail -f output.log

# Watch Go backend logs
docker-compose logs -f go

# Check evaluation progress
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT COUNT(*) as pending FROM content WHERE status='PENDING';"

# Check evaluator performance
docker exec -it truesignal-db psql -U truesignal -d truesignal -c \
  "SELECT evaluator_version, COUNT(*) as count
   FROM evaluation
   GROUP BY evaluator_version;"
```

---

## Rollback Steps (If Issues Found)

```bash
# Stop all services
docker-compose down
Ctrl+C (Go backend)
Ctrl+C (Python backend)
Ctrl+C (Vue frontend)

# Restore database
docker-compose up -d
docker exec -it truesignal-db psql -U truesignal -d truesignal < sql/schema.sql

# Revert code changes
git checkout backend-python/agents/content_evaluator.py
git checkout backend-go/handlers/content_handler.go
git checkout frontend-vue/src/stores/useTimelineStore.js
git checkout frontend-vue/src/components/Timeline.vue

# Restart services
docker-compose up -d
cd backend-go && go run main.go
cd backend-python && python main.py
cd frontend-vue && npm run dev
```

---

## Sign-Off Checklist

- [ ] All 4 services running without errors (Docker, Go, Python, Vue)
- [ ] Smoke test passes (8/8 tests)
- [ ] Database contains articles and evaluations
- [ ] API endpoint returns data with evaluation scores
- [ ] Timeline page loads and displays cards
- [ ] Cards show innovation_score and depth_score
- [ ] Detail drawer displays all evaluation data
- [ ] Fallback works (tested with/without API key)
- [ ] No CORS errors
- [ ] Dark mode renders correctly
- [ ] Performance acceptable (<2s load time)

---

## Final Status

**Ready for Production**: Yes / No ___

**Verified By**: _______________

**Date**: _______________

**Notes**:
```
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
```

---

**Next Steps After Deployment**:
1. Monitor evaluation quality for 24 hours
2. Review rule-based vs LLM consistency
3. Collect user feedback on score accuracy
4. Adjust keyword weights in rule_evaluator.py if needed
5. Plan for scaling (more RSS sources, higher evaluation volume)
