-- Assign admin role to alice and bob.
INSERT INTO account_roles (account_id, role_id)
SELECT a.id, r.id
FROM accounts a, roles r
WHERE a.email IN ('alice@ohm-dev.local', 'bob@ohm-dev.local')
  AND r.name = 'admin';
