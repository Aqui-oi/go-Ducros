@echo off
REM ========================================
REM  Ducros Network - xmrig Miner (Windows)
REM ========================================
REM
REM Serveur: 92.222.10.107:3333
REM Wallet: 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
REM ========================================

echo.
echo ========================================
echo   Ducros Network - RandomX Mining
echo ========================================
echo.
echo Serveur Stratum: 92.222.10.107:3333
echo Wallet: 0x25fFA18Fb7E35E0a3272020305f4BEa0B770A7F2
echo.
echo Demarrage de xmrig...
echo.

REM Lancer xmrig avec la configuration
xmrig.exe --config=xmrig-windows-config.json

pause
