CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Stores (tenant entity)
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    address VARCHAR(500) DEFAULT '',
    phone VARCHAR(50) DEFAULT '',
    logo_url VARCHAR(500) DEFAULT '',
    telegram_group_chat_id BIGINT,
    subscription_tier VARCHAR(20) DEFAULT 'free',
    subscription_expires_at TIMESTAMPTZ,
    commission_rate DECIMAL(5,2) DEFAULT 5.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Store staff
CREATE TABLE store_staff (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    staff_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'barista',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(store_id, staff_code)
);

-- Telegram users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id BIGINT NOT NULL UNIQUE,
    phone VARCHAR(20) DEFAULT '',
    first_name VARCHAR(255) DEFAULT '',
    last_name VARCHAR(255) DEFAULT '',
    username VARCHAR(255) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Menu categories
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_categories_store ON categories(store_id);

-- Menu items
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    base_price BIGINT NOT NULL,
    image_url VARCHAR(500) DEFAULT '',
    is_available BOOLEAN DEFAULT true,
    sort_order INT DEFAULT 0
);

CREATE INDEX idx_items_category ON items(category_id);
CREATE INDEX idx_items_store ON items(store_id);

-- Modifier groups
CREATE TABLE modifier_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    selection_type VARCHAR(20) NOT NULL DEFAULT 'single',
    is_required BOOLEAN DEFAULT false,
    min_selections INT DEFAULT 0,
    max_selections INT DEFAULT 1,
    sort_order INT DEFAULT 0
);

CREATE INDEX idx_modifier_groups_item ON modifier_groups(item_id);

-- Modifiers
CREATE TABLE modifiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    modifier_group_id UUID NOT NULL REFERENCES modifier_groups(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price_adjustment BIGINT DEFAULT 0,
    is_available BOOLEAN DEFAULT true,
    sort_order INT DEFAULT 0
);

CREATE INDEX idx_modifiers_group ON modifiers(modifier_group_id);

-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_number SERIAL,
    user_id UUID NOT NULL REFERENCES users(id),
    store_id UUID NOT NULL REFERENCES stores(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_price BIGINT NOT NULL,
    payment_method VARCHAR(20) DEFAULT 'pay_at_pickup',
    payment_status VARCHAR(20) DEFAULT 'pending',
    eta_minutes INT DEFAULT 15,
    rejection_reason TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_orders_store ON orders(store_id);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(store_id, status);

-- Order items (with price snapshots)
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id UUID REFERENCES items(id) ON DELETE SET NULL,
    item_name VARCHAR(255) NOT NULL,
    item_price BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

-- Order item modifiers (with snapshots)
CREATE TABLE order_item_modifiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    modifier_id UUID REFERENCES modifiers(id) ON DELETE SET NULL,
    modifier_name VARCHAR(255) NOT NULL,
    price_adjustment BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_order_item_modifiers_item ON order_item_modifiers(order_item_id);

-- Commission transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id),
    store_id UUID NOT NULL REFERENCES stores(id),
    order_total BIGINT NOT NULL,
    commission_rate DECIMAL(5,2) NOT NULL,
    commission_amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transactions_store ON transactions(store_id);
