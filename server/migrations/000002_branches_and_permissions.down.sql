DROP INDEX IF EXISTS idx_store_staff_branch;
DROP INDEX IF EXISTS idx_orders_branch;
DROP INDEX IF EXISTS idx_modifiers_branch;
DROP INDEX IF EXISTS idx_modifier_groups_branch;
DROP INDEX IF EXISTS idx_items_branch;
DROP INDEX IF EXISTS idx_categories_branch;
DROP INDEX IF EXISTS idx_branches_store;

ALTER TABLE store_staff
    DROP CONSTRAINT IF EXISTS chk_store_staff_role;
ALTER TABLE store_staff
    DROP CONSTRAINT IF EXISTS fk_store_staff_branch;
ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS fk_orders_branch;
ALTER TABLE modifiers
    DROP CONSTRAINT IF EXISTS fk_modifiers_branch;
ALTER TABLE modifier_groups
    DROP CONSTRAINT IF EXISTS fk_modifier_groups_branch;
ALTER TABLE items
    DROP CONSTRAINT IF EXISTS fk_items_branch;
ALTER TABLE categories
    DROP CONSTRAINT IF EXISTS fk_categories_branch;

ALTER TABLE store_staff
    DROP COLUMN IF EXISTS branch_id;
ALTER TABLE orders
    DROP COLUMN IF EXISTS branch_id;
ALTER TABLE modifiers
    DROP COLUMN IF EXISTS branch_id;
ALTER TABLE modifier_groups
    DROP COLUMN IF EXISTS branch_id;
ALTER TABLE items
    DROP COLUMN IF EXISTS branch_id;
ALTER TABLE categories
    DROP COLUMN IF EXISTS branch_id;

ALTER TABLE stores
    DROP COLUMN IF EXISTS category;

DROP TABLE IF EXISTS branches;
