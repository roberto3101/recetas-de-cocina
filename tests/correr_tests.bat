@echo off
chcp 65001 >nul 2>&1
title Suite de pruebas - Gestor de Credenciales
setlocal enabledelayedexpansion

cd /d "%~dp0..\plataforma-backend"

echo.
echo ================================================================
echo   SUITE DE PRUEBAS - sistemas_unificados (gestor de credenciales)
echo ================================================================
echo.

where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go no esta instalado o no esta en el PATH.
    echo         Descarga desde https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

for /f "delims=" %%i in ('go version') do set "GOVERSION=%%i"
echo Go version  : %GOVERSION%
echo Directorio  : %CD%

if exist ".env" (
    for /f "usebackq tokens=1,* delims==" %%a in (`findstr /b "BASE_DATOS_URL=" ".env"`) do set "BASE_DATOS_URL=%%b"
)
if not defined BASE_DATOS_URL set "BASE_DATOS_URL=postgresql://root@localhost:26258/sistemas_unificados?sslmode=disable"
echo Base datos  : %BASE_DATOS_URL%
echo.

echo ----------------------------------------------------------------
echo  [1/3] Compilando todo el codigo (go build)
echo ----------------------------------------------------------------
go build ./...
if errorlevel 1 (
    echo.
    echo [FAIL] La compilacion fallo. Corrige los errores antes de continuar.
    echo.
    pause
    exit /b 1
)
echo [OK] Compilacion correcta.
echo.

echo ----------------------------------------------------------------
echo  [2/3] Analisis estatico (go vet)
echo ----------------------------------------------------------------
go vet ./...
if errorlevel 1 (
    echo.
    echo [FAIL] go vet detecto advertencias. Corrigelas antes de continuar.
    echo.
    pause
    exit /b 1
)
echo [OK] Sin advertencias.
echo.

echo ----------------------------------------------------------------
echo  [3/3] Ejecutando tests (modo verbose, sin cache)
echo ----------------------------------------------------------------
echo.
go test -v -count=1 -timeout 120s ./...
set "TEST_EXIT=%errorlevel%"

echo.
echo ================================================================
if "%TEST_EXIT%"=="0" (
    echo                  ===  TODOS LOS TESTS PASARON  ===
) else (
    echo                  ===  HAY TESTS FALLIDOS  ===
    echo.
    echo Revisa la salida arriba: cada test fallido aparece marcado
    echo como  --- FAIL: NombreDelTest  con el detalle del error debajo.
)
echo ================================================================
echo.

pause
exit /b %TEST_EXIT%
