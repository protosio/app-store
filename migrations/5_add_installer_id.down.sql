BEGIN;
DROP INDEX id_idx;
ALTER TABLE installer DROP COLUMN id;
END;