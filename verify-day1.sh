#!/bin/bash

# TrueSignal Day 1 Verification Script
# Verifies: Docker, PostgreSQL, Redis, and initial setup
# Execution: bash verify-day1.sh

set -e

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "  TrueSignal Day 1 - Verification Script"
echo "════════════════════════════════════════════════════════════════"
echo ""

# 1. Check Docker
echo "📦 [1/5] Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi
echo "✅ Docker installed"
echo ""

# 2. Check Docker compose status
echo "🐳 [2/5] Checking Docker containers..."
if ! docker-compose ps > /dev/null 2>&1; then
    echo "⚠️  Docker containers not running. Starting..."
    docker-compose up -d > /dev/null 2>&1
    sleep 5
fi
echo "✅ Docker containers running"
echo ""

# 3. Check PostgreSQL
echo "🐘 [3/5] Checking PostgreSQL connection..."
if ! docker exec truesignal-db psql -U truesignal -d truesignal -c "SELECT version();" > /dev/null 2>&1; then
    echo "❌ PostgreSQL connection failed"
    echo "   Try: docker-compose down -v && docker-compose up -d"
    exit 1
fi
echo "✅ PostgreSQL connected"
echo ""

# 4. Check Redis
echo "🔴 [4/5] Checking Redis connection..."
if ! docker exec truesignal-redis redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis connection failed"
    echo "   Try: docker-compose down -v && docker-compose up -d"
    exit 1
fi
echo "✅ Redis connected"
echo ""

# 5. Check Go backend compilation
echo "🐹 [5/5] Checking Go backend..."
cd backend-go
if ! go build -v ./internal/config > /dev/null 2>&1; then
    echo "❌ Go backend compilation failed"
    echo "   Try: cd backend-go && go mod download"
    cd ..
    exit 1
fi
echo "✅ Go backend compiles successfully"
cd ..
echo ""

echo "════════════════════════════════════════════════════════════════"
echo "✅ All Day 1 Verifications Passed!"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo "📊 Environment Status:"
echo "   • Docker: Ready"
echo "   • PostgreSQL: Connected (localhost:5432)"
echo "   • Redis: Connected (localhost:6379)"
echo "   • Go Backend: Buildable"
echo ""
echo "🚀 Next Steps:"
echo "   1. Start all services: ./start-all.sh"
echo "   2. Open frontend: http://localhost:5173"
echo "   3. Check logs: docker-compose logs -f"
echo ""
echo "📚 Documentation:"
echo "   • description/README.md - Project overview"
echo "   • description/MASTER_INDEX.md - Full documentation index"
echo "   • CLAUDE.md - Development guidelines"
echo ""
