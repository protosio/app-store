BEGIN;
DROP INDEX ix_installer_tsv;
ALTER TABLE installer DROP COLUMN tsv;
END;