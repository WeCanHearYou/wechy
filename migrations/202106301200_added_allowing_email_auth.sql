ALTER TABLE tenants ADD COLUMN is_allowing_email_auth BOOLEAN;
UPDATE tenants SET is_allowing_email_auth = TRUE;
ALTER TABLE tenants ALTER COLUMN is_allowing_email_auth SET NOT NULL;