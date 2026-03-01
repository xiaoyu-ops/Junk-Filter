#!/bin/bash

# TrueSignal Complete Startup Script (Linux/Mac)
# Start in correct order: Docker - Go Backend - Python Backend - Frontend

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo ""
echo "========================================"
echo "    TrueSignal All Services Startup"
echo "    System: Linux/Mac"
echo "========================================"
echo ""

# Check if in project root directory
if [ ! -f "docker-compose.yml" ]; then
    echo "[ERROR] Please run this script from TrueSignal project root directory"
    exit 1
fi

# ========== Phase 1: Start Docker ==========
echo ""
echo "========== Phase 1: Starting Docker Containers =========="
echo ""

echo "[INFO] Checking Docker status..."
if ! command -v docker &> /dev/null; then
    echo "[ERROR] Docker is not installed or not in PATH"
    exit 1
fi

echo "[INFO] Stopping old containers..."
docker-compose down 2>/dev/null || true

echo "[INFO] Starting Docker containers..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo "[ERROR] Docker startup failed"
    exit 1
fi

echo "[SUCCESS] Docker containers started"
sleep 3

# ========== Phase 2: Start Go Backend ==========
echo ""
echo "========== Phase 2: Starting Go Backend (Port 8080) =========="
echo ""

if [ ! -f "backend-go/main.go" ]; then
    echo "[ERROR] backend-go/main.go not found"
    exit 1
fi

echo "[INFO] Starting Go backend..."
cd backend-go

# Start Go in background
go run main.go &
GO_PID=$!

cd ..

echo "[INFO] Waiting for Go backend to start..."
attempt=0
while [ $attempt -lt 30 ]; do
    sleep 1
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        break
    fi
    attempt=$((attempt + 1))
done

if [ $attempt -eq 30 ]; then
    echo "[ERROR] Go backend startup timeout"
    kill $GO_PID 2>/dev/null || true
    exit 1
fi

echo "[SUCCESS] Go backend started (Port 8080)"

# ========== Phase 3: Optional Python Backend ==========
if [ -f "backend-python/main.py" ]; then
    echo ""
    echo "========== Phase 3: Starting Python Backend (Port 8081) =========="
    echo ""

    echo "[INFO] Starting Python backend with conda junkfilter environment..."
    cd backend-python

    # Activate conda junkfilter environment and run Python backend in background
    (
        # Need to initialize conda first (for non-interactive shells)
        eval "$(conda shell.bash hook)"
        conda activate junkfilter
        python main.py
    ) &

    cd ..

    echo "[INFO] Python backend started in background"
else
    echo ""
    echo "[INFO] backend-python/main.py not found, skipping Python backend"
fi

# ========== Phase 4: Verify Services ==========
echo ""
echo "========== Verifying Services =========="
echo ""

echo "[CHECK] PostgreSQL..."
if docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT 1" > /dev/null 2>&1; then
    echo "[SUCCESS] PostgreSQL is running"
else
    echo "[WARNING] PostgreSQL connection failed"
fi

echo "[CHECK] Redis..."
if docker exec truesignal-redis redis-cli ping > /dev/null 2>&1; then
    echo "[SUCCESS] Redis is running"
else
    echo "[WARNING] Redis connection failed"
fi

echo "[CHECK] Go Backend..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "[SUCCESS] Go backend is running (Port 8080)"
else
    echo "[WARNING] Go backend is not responding"
fi

# ========== Phase 5: Start Frontend ==========
echo ""
echo "========== Phase 5: Starting Vue Frontend (Port 5173) =========="
echo ""

if [ ! -f "frontend-vue/package.json" ]; then
    echo "[ERROR] frontend-vue/package.json not found"
    exit 1
fi

echo "[INFO] Checking frontend dependencies..."
cd frontend-vue

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "[INFO] Installing npm dependencies..."
    npm install
    if [ $? -ne 0 ]; then
        echo "[ERROR] npm install failed"
        exit 1
    fi
fi

# Start frontend in background
echo "[INFO] Starting Vue development server..."
npm run dev &

cd ..

echo "[INFO] Waiting for frontend to start (10 seconds)..."
sleep 10

# ========== Summary ==========
echo ""
echo "========================================"
echo "            Startup Summary"
echo "========================================"
echo ""

echo "[COMPLETE] All services have started!"
echo ""
echo "Services Running:"
echo "  * Docker:        PostgreSQL + Redis (Port 5432, 6379)"
echo "  * Go Backend:    http://localhost:8080/api (Port 8080)"
echo "  * Python:        http://localhost:8081 (Port 8081, optional)"
echo "  * Vue Frontend:  http://localhost:5173 (Port 5173)"
echo ""
echo "Next Steps:"
echo "  1. Open browser: http://localhost:5173"
echo "  2. You should see the TrueSignal application"
echo "  3. Test API: Run bash smoke_test.sh in new terminal"
echo ""
echo "Quick Commands:"
echo "  * Test API:     bash smoke_test.sh"
echo "  * View logs:    docker-compose logs -f"
echo "  * Stop all:     docker-compose down"
echo ""
echo ""
echo "Verification:"
echo "  * Run tests (in new terminal): bash smoke_test.sh"
echo "  * Expected result: 8/8 tests pass"
echo ""

# Keep script running to maintain background services
wait
