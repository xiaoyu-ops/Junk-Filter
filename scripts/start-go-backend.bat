@echo off
REM Start Go Backend with JunkFilter Environment Variables
REM This script sets the correct database credentials and starts the Go backend

setlocal enabledelayedexpansion

REM Set JunkFilter database credentials
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=junkfilter
set DB_PASSWORD=junkfilter123
set DB_NAME=junkfilter

REM Navigate to backend-go directory
cd /d "%~dp0backend-go"

REM Start Go backend
echo.
echo ========== Starting JunkFilter Go Backend ==========
echo Database: %DB_HOST%:%DB_PORT%/%DB_NAME%
echo User: %DB_USER%
echo.
go run main.go

pause
