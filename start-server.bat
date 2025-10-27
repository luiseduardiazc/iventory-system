@echo off
set DATABASE_DRIVER=sqlite
set SQLITE_PATH=test.db
set MESSAGE_BROKER=none
set API_KEYS=dev-key-store-001
set SERVER_PORT=8080
start /B bin\inventory-api.exe
