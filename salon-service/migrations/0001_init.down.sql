DROP INDEX IF EXISTS idx_staff_status;
DROP INDEX IF EXISTS idx_services_status;
DROP INDEX IF EXISTS idx_branches_geo_location_gin;
DROP INDEX IF EXISTS idx_salons_geo_location_gin;

DROP TABLE IF EXISTS staff_services;
DROP TABLE IF EXISTS staff;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS salons;
