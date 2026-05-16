@echo off
chcp 65001 >nul 2>&1
title Suite E2E - Playwright (Chromium real)
setlocal enabledelayedexpansion

cd /d "%~dp0..\plataforma-frontend"

echo.
echo ================================================================
echo   SUITE E2E PLAYWRIGHT - sistemas_unificados
echo ================================================================
echo.

where node >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js no esta instalado o no esta en el PATH.
    echo         Descarga desde https://nodejs.org/
    pause
    exit /b 1
)

for /f "delims=" %%i in ('node --version') do set "NODEVER=%%i"
echo Node version: %NODEVER%
echo Directorio  : %CD%
echo.

if not exist "node_modules\@playwright" (
    echo Instalando dependencias npm ...
    call npm install
    if errorlevel 1 (
        echo [FAIL] npm install fallo.
        pause
        exit /b 1
    )
)

echo ----------------------------------------------------------------
echo  [1/4] Verificando navegadores de Playwright
echo ----------------------------------------------------------------
call npx playwright install chromium
if errorlevel 1 (
    echo [FAIL] No se pudo instalar Chromium para Playwright.
    pause
    exit /b 1
)
echo [OK] Chromium disponible.
echo.

echo ----------------------------------------------------------------
echo  [2/4] Verificando que backend este corriendo (puerto 8080)
echo ----------------------------------------------------------------
powershell -NoProfile -Command "$ProgressPreference='SilentlyContinue'; try { $r = Invoke-WebRequest -Uri 'http://localhost:8080/salud' -UseBasicParsing -TimeoutSec 3; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }"
if errorlevel 1 (
    echo [FAIL] El backend no esta corriendo en http://localhost:8080.
    echo        Abre OTRA ventana y arranca:
    echo            cd ..\plataforma-backend
    echo            .\bin\plataforma.exe
    pause
    exit /b 1
)
echo [OK] Backend OK.
echo.

echo ----------------------------------------------------------------
echo  [3/4] Verificando que frontend este corriendo (puerto 5173)
echo ----------------------------------------------------------------
powershell -NoProfile -Command "$ProgressPreference='SilentlyContinue'; try { $r = Invoke-WebRequest -Uri 'http://localhost:5173/' -UseBasicParsing -TimeoutSec 3; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }"
if errorlevel 1 (
    echo [FAIL] El frontend no esta corriendo en http://localhost:5173.
    echo        Abre OTRA ventana y arranca:
    echo            cd ..\plataforma-frontend
    echo            npm run dev
    pause
    exit /b 1
)
echo [OK] Frontend OK.
echo.

echo ----------------------------------------------------------------
echo  [4/4] Ejecutando tests E2E con navegador real
echo ----------------------------------------------------------------
echo.
call npx playwright test --reporter=list
set "TEST_EXIT=%errorlevel%"

echo.
echo ================================================================
if "%TEST_EXIT%"=="0" (
    echo                  ===  TODOS LOS TESTS E2E PASARON  ===
    echo.
    echo Para ver el informe HTML interactivo:
    echo     npx playwright show-report
) else (
    echo                  ===  HAY TESTS E2E FALLIDOS  ===
    echo.
    echo Para ver detalle:
    echo     npx playwright show-report
    echo Para correr con navegador visible:
    echo     npx playwright test --headed
)
echo ================================================================
echo.

pause
exit /b %TEST_EXIT%
