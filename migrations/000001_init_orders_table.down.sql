DROP TRIGGER IF EXISTS update_order_items_updated_at ON order_items;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_order_items_order_id;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;

DROP EXTENSION IF EXISTS "uuid-ossp";