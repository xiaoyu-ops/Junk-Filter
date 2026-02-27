@echo off
REM ================================================================
REM  TrueSignal 前端 - 白屏问题快速修复脚本
REM  一键清理 + 重装 + 启动
REM ================================================================

setlocal enabledelayedexpansion

echo.
echo ================================================================
echo   🔧 TrueSignal 白屏问题修复脚本
echo ================================================================
echo.

REM 进入项目目录
cd /d "D:\TrueSignal\frontend-vue" || (
    echo ❌ 项目目录不存在
    pause
    exit /b 1
)

echo 📍 当前目录：%cd%
echo.

REM 步骤 1: 清理缓存
echo [1/4] 🧹 清理 npm 缓存...
call npm cache clean --force >nul 2>&1
if !errorlevel! equ 0 (
    echo ✅ npm 缓存已清理
) else (
    echo ⚠️  npm 缓存清理出现问题，继续...
)
echo.

REM 步骤 2: 删除旧文件
echo [2/4] 🗑️  删除旧的依赖和缓存...
if exist "node_modules" (
    echo   删除 node_modules...
    rmdir /s /q node_modules >nul 2>&1
)
if exist "package-lock.json" (
    echo   删除 package-lock.json...
    del /q package-lock.json >nul 2>&1
)
if exist ".vite" (
    echo   删除 .vite 缓存...
    rmdir /s /q .vite >nul 2>&1
)
echo ✅ 旧文件已清理
echo.

REM 步骤 3: 重新安装
echo [3/4] 📦 重新安装依赖...
call npm install
if !errorlevel! equ 0 (
    echo ✅ 依赖安装完成
) else (
    echo ❌ 依赖安装失败
    pause
    exit /b 1
)
echo.

REM 步骤 4: 验证关键依赖
echo [4/4] ✓ 验证关键依赖...
call npm list @vueuse/core pinia vue vue-router >nul 2>&1
if !errorlevel! equ 0 (
    echo ✅ 所有依赖验证通过
) else (
    echo ⚠️  某些依赖验证失败，但继续启动...
)
echo.

echo ================================================================
echo   ✅ 修复完成！现在启动开发服务器...
echo ================================================================
echo.
echo   浏览器将自动打开 http://localhost:5173
echo   按 Ctrl+C 停止服务器
echo.
echo ================================================================
echo.

REM 启动开发服务器
call npm run dev

pause
