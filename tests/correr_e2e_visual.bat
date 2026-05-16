@echo off
chcp 65001 >nul 2>&1
title Suite E2E VISUAL - Playwright (Chromium con UI)
setlocal enabledelayedexpansion

cd /d "%~dp0..\plataforma-frontend"

echo.
echo ================================================================
echo   SUITE E2E PLAYWRIGHT - MODO VISUAL (ves el navegador)
echo ================================================================
echo.
echo  - El navegador se abre y vas viendo cada test en vivo.
echo  - Cada accion se ralentiza 400ms para que puedas seguirla.
echo  - Si quieres detenerte y hacer paso a paso, usa:
echo        npx playwright test --debug
echo  - Para el modo interactivo con timeline:
echo        npx playwright test --ui
echo.

where node >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js no esta instalado.
    pause
    exit /b 1
)

if not exist "node_modules\@playwright" (
    echo Instalando dependencias npm ...
    call npm install
    if errorlevel 1 ( pause & exit /b 1 )
)

echo Verificando navegadores ...
call npx playwright install chromium >nul 2>&1

echo Verificando backend en :8080 ...
powershell -NoProfile -Command "$ProgressPreference='SilentlyContinue'; try { (Invoke-WebRequest -Uri 'http://localhost:8080/salud' -UseBasicParsing -TimeoutSec 3).StatusCode | Out-Null; exit 0 } catch { exit 1 }"
if errorlevel 1 (
    echo [FAIL] Backend no esta corriendo. Arranca: cd plataforma-backend ^&^& .\bin\plataforma.exe
    pause
    exit /b 1
)

echo Verificando frontend en :5173 ...
powershell -NoProfile -Command "$ProgressPreference='SilentlyContinue'; try { (Invoke-WebRequest -Uri 'http://localhost:5173/' -UseBasicParsing -TimeoutSec 3).StatusCode | Out-Null; exit 0 } catch { exit 1 }"
if errorlevel 1 (
    echo [FAIL] Frontend no esta corriendo. Arranca: cd plataforma-frontend ^&^& npm run dev
    pause
    exit /b 1
)

echo.
echo Ejecutando con navegador VISIBLE y ralentizado ...
echo.

set "PLAYWRIGHT_SLOWMO=150"
call npx playwright test --headed --workers=1 --reporter=list
set "TEST_EXIT=%errorlevel%"

echo.
echo ================================================================
if "%TEST_EXIT%"=="0" (
    echo                  ===  TODOS LOS TESTS E2E PASARON  ===
) else (
    echo                  ===  HAY TESTS E2E FALLIDOS  ===
)
echo ================================================================
echo.

pause
exit /b %TEST_EXIT%
