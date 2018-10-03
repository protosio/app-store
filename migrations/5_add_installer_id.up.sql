BEGIN;
ALTER TABLE installer ADD COLUMN id VARCHAR;
CREATE INDEX id_idx ON installer (id);
END;