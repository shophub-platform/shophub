-- 0001_init.down.sql : poništava početnu ShopHub šemu.
DROP INDEX IF EXISTS idx_shops_owner_id;
DROP TABLE IF EXISTS shops;
DROP TABLE IF EXISTS users;
