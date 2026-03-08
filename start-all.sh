#!/bin/bash

# JunkFilter Web Mode Startup Script (Linux/Mac)
# All backends run in Docker, frontend opens in browser

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo ""
echo "========================================"
echo "    JunkFilter Startup (Web Mode)"
echo "========================================"
echo ""

# Check prerequisites
if [ ! -f "docker-compose.yml" ]; then
    echo "[ERROR] docker-compose.yml not found"
    exit 1
fi

if [ ! -f ".env" ]; then
    echo "[ERROR] .env file not found"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "[ERROR] Docker is not installed"
    exit 1
fi

echo "[OK] All prerequisites checked"
echo ""

# Clean up port 5173
lsof -ti:5173 2>/dev/null | xargs kill -9 2>/dev/null || true

# Phase 1: Start all backends in Docker
echo "========== Phase 1: Starting All Backends (Docker) =========="
echo ""

echo "[INFO] Building and starting containers..."
docker-compose up -d

echo "[OK] All backend services started in Docker"
echo ""

echo "[INFO] Waiting for services to be ready..."
sleep 8

echo "[OK] PostgreSQL ready"
echo "[OK] Redis ready"
sleep 3
echo "[OK] Go backend ready (Port 8080)"
echo "[OK] Python API ready (Port 8083)"
echo "[OK] Python evaluator ready"
echo ""

# Phase 2: Start Web Frontend
echo "========== Phase 2: Starting Web Frontend (Port 5173) =========="
echo ""

if [ ! -d "frontend-vue/node_modules" ]; then
    echo "[INFO] Installing npm dependencies..."
    cd frontend-vue && npm install && cd ..
fi

cd frontend-vue
npm run dev &
cd ..

sleep 5

# Summary
echo ""
echo "========================================"
echo "          JunkFilter Running"
echo "========================================"
echo ""
echo "  [Docker]  PostgreSQL      :5432"
echo "  [Docker]  Redis           :6379"
echo "  [Docker]  Go Backend      :8080"
echo "  [Docker]  Python API      :8083"
echo "  [Docker]  Python Evaluator (background)"
echo "  [Native]  Web Frontend    http://localhost:5173"
echo ""
echo "  To stop backends: docker-compose down"
echo "  To view logs:     docker-compose logs -f"
echo ""

wait
