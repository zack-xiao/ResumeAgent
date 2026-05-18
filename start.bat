@echo off
echo ================================
echo   Portfolio AI 启动脚本
echo ================================

REM 检查 Go 环境
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go
    pause
    exit /b 1
)

REM 检查 API Key
if "%DEEPSEEK_API_KEY%"=="" (
    echo [警告] 未设置 DEEPSEEK_API_KEY 环境变量
    echo 请先在 .env 文件中设置，或运行:
    echo   set DEEPSEEK_API_KEY=sk-your-key
)

cd /d "%~dp0backend"

echo.
echo [1/3] 安装依赖...
go mod tidy

echo.
echo [2/3] 编译并启动服务...
echo 访问地址: http://localhost:8080
echo.

go run main.go

pause
