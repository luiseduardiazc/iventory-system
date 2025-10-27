# Start inventory system with PostgreSQL
$env:DATABASE_DRIVER="postgres"
$env:POSTGRES_HOST="localhost"
$env:POSTGRES_PORT="5432"
$env:POSTGRES_USER="inventory"
$env:POSTGRES_PASSWORD="inventory123"
$env:POSTGRES_DB="inventory_db"
$env:MESSAGE_BROKER="redis"
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="6379"
$env:API_KEYS="admin-key-123"
$env:LOG_LEVEL="debug"

.\bin\inventory-api.exe
