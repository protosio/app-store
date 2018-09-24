BEGIN;
ALTER TABLE installer ADD COLUMN version_metadata jsonb;
END;