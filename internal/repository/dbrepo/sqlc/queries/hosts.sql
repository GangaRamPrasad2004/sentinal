-- name: InsertHost :one
INSERT INTO hosts (host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;

-- name: ListServiceIDs :many
SELECT id FROM services;

-- name: InsertHostServiceDefault :exec
INSERT INTO host_services 
    (host_id, service_id, active, schedule_number, schedule_unit, status, created_at, updated_at)
VALUES ($1, $2, 0, 3, 'm', 'pending', $3, $4);

-- name: GetHostByID :one
SELECT id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at
FROM hosts
WHERE id = $1;

-- name: GetHostServicesForHost :many
SELECT 
    hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit, 
    hs.last_check, hs.status, hs.created_at, hs.updated_at,
    COALESCE(s.id, 0) as service_id_ref, 
    COALESCE(s.service_name, '') as service_name, 
    COALESCE(s.active, 0) as service_active, 
    COALESCE(s.icon, '') as icon, 
    COALESCE(s.created_at, hs.created_at) as service_created_at, 
    COALESCE(s.updated_at, hs.updated_at) as service_updated_at, 
    hs.last_message
FROM host_services hs
LEFT JOIN services s ON (s.id = hs.service_id)
WHERE hs.host_id = $1
ORDER BY s.service_name;

-- name: UpdateHost :exec
UPDATE hosts 
SET host_name = $1, canonical_name = $2, url = $3, ip = $4, ipv6 = $5, os = $6,
    active = $7, location = $8, updated_at = $9 
WHERE id = $10;

-- name: GetAllServiceStatusCounts :one
SELECT 
    (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'pending') as pending,
    (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'healthy') as healthy,
    (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'warning') as warning,
    (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'problem') as problem;

-- name: AllHosts :many
SELECT id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at 
FROM hosts 
ORDER BY host_name;

-- name: UpdateHostServiceStatus :exec
UPDATE host_services 
SET active = $1 
WHERE host_id = $2 AND service_id = $3;

-- name: UpdateHostService :exec
UPDATE host_services 
SET host_id = $1, service_id = $2, active = $3,
    schedule_number = $4, schedule_unit = $5, 
    last_check = $6, status = $7, updated_at = $8, last_message = $9
WHERE id = $10;

-- name: GetServicesByStatus :many
SELECT
    hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
    hs.last_check, hs.status, hs.created_at, hs.updated_at,
    COALESCE(h.host_name, '') as host_name, 
    COALESCE(s.service_name, '') as service_name, 
    hs.last_message
FROM host_services hs
LEFT JOIN hosts h ON (hs.host_id = h.id)
LEFT JOIN services s ON (hs.service_id = s.id)
WHERE hs.status = $1 AND hs.active = 1
ORDER BY host_name, service_name;

-- name: GetHostServiceByID :one
SELECT 
    hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number,
    hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at,
    COALESCE(s.id, 0) as service_id_ref, 
    COALESCE(s.service_name, '') as service_name, 
    COALESCE(s.active, 0) as service_active, 
    COALESCE(s.icon, '') as icon, 
    COALESCE(s.created_at, hs.created_at) as service_created_at, 
    COALESCE(s.updated_at, hs.updated_at) as service_updated_at, 
    COALESCE(h.host_name, '') as host_name, 
    hs.last_message
FROM host_services hs
LEFT JOIN services s ON (hs.service_id = s.id)
LEFT JOIN hosts h ON (hs.host_id = h.id)
WHERE hs.id = $1;

-- name: GetServicesToMonitor :many
SELECT 
    hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number,
    hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at,
    COALESCE(s.id, 0) as service_id_ref, 
    COALESCE(s.service_name, '') as service_name, 
    COALESCE(s.active, 0) as service_active, 
    COALESCE(s.icon, '') as icon, 
    COALESCE(s.created_at, hs.created_at) as service_created_at, 
    COALESCE(s.updated_at, hs.updated_at) as service_updated_at,
    COALESCE(h.host_name, '') as host_name, 
    hs.last_message
FROM host_services hs
LEFT JOIN services s ON (hs.service_id = s.id)
LEFT JOIN hosts h ON (h.id = hs.host_id)
WHERE h.active = 1 AND hs.active = 1;

-- name: GetHostServiceByHostIDServiceID :one
SELECT 
    hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number,
    hs.schedule_unit, hs.last_check, hs.status, hs.created_at, hs.updated_at,
    COALESCE(s.id, 0) as service_id_ref, 
    COALESCE(s.service_name, '') as service_name, 
    COALESCE(s.active, 0) as service_active, 
    COALESCE(s.icon, '') as icon, 
    COALESCE(s.created_at, hs.created_at) as service_created_at, 
    COALESCE(s.updated_at, hs.updated_at) as service_updated_at, 
    COALESCE(h.host_name, '') as host_name, 
    hs.last_message
FROM host_services hs
LEFT JOIN services s ON (hs.service_id = s.id)
LEFT JOIN hosts h ON (hs.host_id = h.id)
WHERE hs.host_id = $1 AND hs.service_id = $2;

-- name: InsertEvent :exec
INSERT INTO events (host_service_id, event_type, host_id, service_name, host_name, message, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetAllEvents :many
SELECT id, event_type, host_service_id, host_id, service_name, host_name, message, created_at, updated_at 
FROM events 
ORDER BY created_at;
