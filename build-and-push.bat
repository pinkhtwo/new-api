@echo off
REM Docker 镜像构建与推送脚本
REM 用法: build-and-push.bat [版本号]
REM 示例: build-and-push.bat v1.0.1

setlocal enabledelayedexpansion

REM Docker Hub 用户名和仓库名
set DOCKER_USER=demokt
set REPO_NAME=new-api

REM 检查是否提供了版本号参数
if "%1"=="" (
    set VERSION=latest
    echo 未指定版本号，将使用 latest 标签
) else (
    set VERSION=%1
    echo 使用版本号: %VERSION%
)

echo.
echo ==========================================
echo  构建 Docker 镜像: %DOCKER_USER%/%REPO_NAME%
echo ==========================================
echo.

REM 检查 Docker 是否运行
docker info >nul 2>&1
if errorlevel 1 (
    echo [错误] Docker 未运行，请先启动 Docker Desktop
    pause
    exit /b 1
)

REM 检查是否已登录 Docker Hub
docker info 2>nul | findstr "Username" >nul
if errorlevel 1 (
    echo [提示] 请先登录 Docker Hub
    docker login
    if errorlevel 1 (
        echo [错误] Docker Hub 登录失败
        pause
        exit /b 1
    )
)

echo.
echo [1/3] 构建 Docker 镜像...
echo 使用国内镜像源 (Dockerfile.cn) 加速构建
echo.

docker build -f Dockerfile.cn -t %DOCKER_USER%/%REPO_NAME%:latest -t %DOCKER_USER%/%REPO_NAME%:%VERSION% .

if errorlevel 1 (
    echo.
    echo [错误] Docker 镜像构建失败
    pause
    exit /b 1
)

echo.
echo [2/3] 推送 latest 标签...
docker push %DOCKER_USER%/%REPO_NAME%:latest

if errorlevel 1 (
    echo.
    echo [错误] 推送 latest 标签失败
    pause
    exit /b 1
)

echo.
echo [3/3] 推送 %VERSION% 标签...
docker push %DOCKER_USER%/%REPO_NAME%:%VERSION%

if errorlevel 1 (
    echo.
    echo [错误] 推送 %VERSION% 标签失败
    pause
    exit /b 1
)

echo.
echo ==========================================
echo  构建并推送成功！
echo ==========================================
echo.
echo 镜像地址:
echo   - %DOCKER_USER%/%REPO_NAME%:latest
echo   - %DOCKER_USER%/%REPO_NAME%:%VERSION%
echo.
echo 查看镜像: https://hub.docker.com/r/%DOCKER_USER%/%REPO_NAME%/tags
echo.

pause