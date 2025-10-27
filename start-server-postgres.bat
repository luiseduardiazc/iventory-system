@echo off
set DATABASE_DRIVER=postgres
set POSTGRES_HOST=localhost
set POSTGRES_PORT=5432
set POSTGRES_USER=postgres
set POSTGRES_PASSWORD=postgres
set POSTGRES_DB=inventory
set MESSAGE_BROKER=none
set API_KEYS=dev-key-store-001
set SERVER_PORT=8080
start "Inventory API Server" bin\inventory-api.exe
