-- =========================================
-- Schema para Sistema de Inventario Distribuido
-- Compatible con SQLite y PostgreSQL
-- =========================================

-- Tabla de productos (catálogo global)
CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    sku TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT,
    price REAL NOT NULL CHECK (price >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para products
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- Tabla de stock (multi-tenant por store_id)
CREATE TABLE IF NOT EXISTS stock (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    store_id TEXT NOT NULL,              -- Identificador de la tienda
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved INTEGER NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    version INTEGER NOT NULL DEFAULT 1,  -- Optimistic locking
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    UNIQUE(product_id, store_id),        -- Un stock por producto-tienda
    CHECK (reserved <= quantity)
);

-- Índices para stock
CREATE INDEX IF NOT EXISTS idx_stock_product_store ON stock(product_id, store_id);
CREATE INDEX IF NOT EXISTS idx_stock_store ON stock(store_id);
CREATE INDEX IF NOT EXISTS idx_stock_product ON stock(product_id);

-- Tabla de reservas
CREATE TABLE IF NOT EXISTS reservations (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    store_id TEXT NOT NULL,              -- Tienda donde se reserva
    customer_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'EXPIRED')),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Índices para reservations
CREATE INDEX IF NOT EXISTS idx_reservations_store ON reservations(store_id);
CREATE INDEX IF NOT EXISTS idx_reservations_customer ON reservations(customer_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires ON reservations(expires_at) WHERE status = 'PENDING';
CREATE INDEX IF NOT EXISTS idx_reservations_status_expires ON reservations(status, expires_at);

-- Tabla de eventos (Event Sourcing)
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    store_id TEXT NOT NULL,              -- Origen del evento
    aggregate_id TEXT NOT NULL,          -- ID del producto/reserva afectado
    payload TEXT NOT NULL,               -- JSON con datos del evento
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed INTEGER DEFAULT 0          -- SQLite usa INTEGER para boolean
);

-- Índices para events
CREATE INDEX IF NOT EXISTS idx_events_type_timestamp ON events(event_type, timestamp);
CREATE INDEX IF NOT EXISTS idx_events_store ON events(store_id);
CREATE INDEX IF NOT EXISTS idx_events_aggregate ON events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_events_unprocessed ON events(processed) WHERE processed = 0;

-- Tabla de tiendas (metadata, opcional pero útil)
CREATE TABLE IF NOT EXISTS stores (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT,
    city TEXT,
    country TEXT,
    phone TEXT,
    email TEXT,
    active INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =========================================
-- Datos de ejemplo para testing
-- =========================================

-- Tiendas de ejemplo
INSERT OR IGNORE INTO stores (id, name, city, country, active) VALUES
    ('MAD-001', 'Madrid Centro', 'Madrid', 'España', 1),
    ('BCN-001', 'Barcelona Plaza Catalunya', 'Barcelona', 'España', 1),
    ('VAL-001', 'Valencia Norte', 'Valencia', 'España', 1),
    ('SEV-001', 'Sevilla Centro', 'Sevilla', 'España', 1);

-- Productos de ejemplo
INSERT OR IGNORE INTO products (id, sku, name, description, category, price) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'PROD-001', 'Laptop HP Pavilion 15', '15.6" FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD', 'electronics', 599.99),
    ('550e8400-e29b-41d4-a716-446655440001', 'PROD-002', 'Mouse Logitech MX Master 3', 'Wireless, ergonómico, 7 botones programables', 'accessories', 99.99),
    ('550e8400-e29b-41d4-a716-446655440002', 'PROD-003', 'Teclado Mecánico RGB', 'Switches Cherry MX Blue, iluminación RGB', 'accessories', 89.99),
    ('550e8400-e29b-41d4-a716-446655440003', 'PROD-004', 'Monitor LG 27" 4K', '27" IPS 4K UHD, 60Hz, HDR10', 'electronics', 349.99),
    ('550e8400-e29b-41d4-a716-446655440004', 'PROD-005', 'Webcam Logitech C920', 'Full HD 1080p, micrófono dual', 'accessories', 79.99);

-- Stock inicial (multi-tienda)
INSERT OR IGNORE INTO stock (id, product_id, store_id, quantity, reserved, version) VALUES
    -- Madrid
    ('stock-mad-001', '550e8400-e29b-41d4-a716-446655440000', 'MAD-001', 10, 0, 1),
    ('stock-mad-002', '550e8400-e29b-41d4-a716-446655440001', 'MAD-001', 50, 5, 1),
    ('stock-mad-003', '550e8400-e29b-41d4-a716-446655440002', 'MAD-001', 20, 2, 1),
    ('stock-mad-004', '550e8400-e29b-41d4-a716-446655440003', 'MAD-001', 5, 1, 1),
    ('stock-mad-005', '550e8400-e29b-41d4-a716-446655440004', 'MAD-001', 15, 0, 1),
    
    -- Barcelona
    ('stock-bcn-001', '550e8400-e29b-41d4-a716-446655440000', 'BCN-001', 15, 2, 1),
    ('stock-bcn-002', '550e8400-e29b-41d4-a716-446655440001', 'BCN-001', 30, 3, 1),
    ('stock-bcn-003', '550e8400-e29b-41d4-a716-446655440002', 'BCN-001', 25, 0, 1),
    ('stock-bcn-004', '550e8400-e29b-41d4-a716-446655440003', 'BCN-001', 8, 0, 1),
    ('stock-bcn-005', '550e8400-e29b-41d4-a716-446655440004', 'BCN-001', 20, 1, 1),
    
    -- Valencia
    ('stock-val-001', '550e8400-e29b-41d4-a716-446655440000', 'VAL-001', 5, 1, 1),
    ('stock-val-002', '550e8400-e29b-41d4-a716-446655440001', 'VAL-001', 40, 5, 1),
    ('stock-val-003', '550e8400-e29b-41d4-a716-446655440002', 'VAL-001', 0, 0, 1),
    ('stock-val-004', '550e8400-e29b-41d4-a716-446655440003', 'VAL-001', 3, 0, 1),
    ('stock-val-005', '550e8400-e29b-41d4-a716-446655440004', 'VAL-001', 12, 2, 1),
    
    -- Sevilla
    ('stock-sev-001', '550e8400-e29b-41d4-a716-446655440000', 'SEV-001', 20, 3, 1),
    ('stock-sev-002', '550e8400-e29b-41d4-a716-446655440001', 'SEV-001', 35, 2, 1),
    ('stock-sev-003', '550e8400-e29b-41d4-a716-446655440002', 'SEV-001', 18, 1, 1),
    ('stock-sev-004', '550e8400-e29b-41d4-a716-446655440003', 'SEV-001', 6, 0, 1),
    ('stock-sev-005', '550e8400-e29b-41d4-a716-446655440004', 'SEV-001', 10, 0, 1);
