CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(500) DEFAULT '',
    lat DECIMAL(10,8),
    lng DECIMAL(11,8),
    banner_image_url VARCHAR(500) DEFAULT '',
    telegram_group_chat_id BIGINT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (store_id, name)
);

CREATE INDEX idx_branches_store ON branches(store_id);

ALTER TABLE stores
    ADD COLUMN category VARCHAR(50);

ALTER TABLE categories
    ADD COLUMN branch_id UUID;
ALTER TABLE items
    ADD COLUMN branch_id UUID;
ALTER TABLE modifier_groups
    ADD COLUMN branch_id UUID;
ALTER TABLE modifiers
    ADD COLUMN branch_id UUID;
ALTER TABLE orders
    ADD COLUMN branch_id UUID;
ALTER TABLE store_staff
    ADD COLUMN branch_id UUID;

UPDATE stores
SET category = COALESCE(category, 'bar')
WHERE name = 'Demo Bar' AND (category IS NULL OR category = '');

INSERT INTO branches (store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active)
SELECT
    s.id,
    s.name || ' - Main',
    COALESCE(s.address, ''),
    41.29950000,
    69.24010000,
    '',
    s.telegram_group_chat_id,
    true
FROM stores s
ON CONFLICT (store_id, name) DO UPDATE
SET address = EXCLUDED.address,
    lat = EXCLUDED.lat,
    lng = EXCLUDED.lng,
    banner_image_url = EXCLUDED.banner_image_url,
    telegram_group_chat_id = EXCLUDED.telegram_group_chat_id,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

UPDATE categories c
SET branch_id = b.id
FROM branches b
WHERE c.store_id = b.store_id;

UPDATE items i
SET branch_id = b.id
FROM branches b
WHERE i.store_id = b.store_id;

UPDATE modifier_groups mg
SET branch_id = b.id
FROM branches b
WHERE mg.store_id = b.store_id;

UPDATE modifiers m
SET branch_id = b.id
FROM branches b
WHERE m.store_id = b.store_id;

UPDATE orders o
SET branch_id = b.id
FROM branches b
WHERE o.store_id = b.store_id;

UPDATE store_staff ss
SET
    role = CASE WHEN ss.role = 'owner' THEN 'director' ELSE ss.role END,
    branch_id = CASE WHEN ss.role = 'owner' THEN NULL ELSE b.id END
FROM branches b
WHERE ss.store_id = b.store_id;

ALTER TABLE categories
    ALTER COLUMN branch_id SET NOT NULL;
ALTER TABLE items
    ALTER COLUMN branch_id SET NOT NULL;
ALTER TABLE modifier_groups
    ALTER COLUMN branch_id SET NOT NULL;
ALTER TABLE modifiers
    ALTER COLUMN branch_id SET NOT NULL;
ALTER TABLE orders
    ALTER COLUMN branch_id SET NOT NULL;

ALTER TABLE categories
    ADD CONSTRAINT fk_categories_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;
ALTER TABLE items
    ADD CONSTRAINT fk_items_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;
ALTER TABLE modifier_groups
    ADD CONSTRAINT fk_modifier_groups_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;
ALTER TABLE modifiers
    ADD CONSTRAINT fk_modifiers_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;
ALTER TABLE orders
    ADD CONSTRAINT fk_orders_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE RESTRICT;
ALTER TABLE store_staff
    ADD CONSTRAINT fk_store_staff_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE SET NULL;

ALTER TABLE store_staff
    ADD CONSTRAINT chk_store_staff_role CHECK (role IN ('director', 'manager', 'barista'));

CREATE INDEX idx_categories_branch ON categories(branch_id);
CREATE INDEX idx_items_branch ON items(branch_id);
CREATE INDEX idx_modifier_groups_branch ON modifier_groups(branch_id);
CREATE INDEX idx_modifiers_branch ON modifiers(branch_id);
CREATE INDEX idx_orders_branch ON orders(branch_id);
CREATE INDEX idx_store_staff_branch ON store_staff(branch_id);
