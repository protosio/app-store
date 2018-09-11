BEGIN;
ALTER TABLE installer ADD COLUMN tsv tsvector;
UPDATE installer SET tsv =
    setweight(to_tsvector(name), 'A') ||
    setweight(to_tsvector(description), 'B');
CREATE INDEX ix_installer_tsv ON installer USING GIN(tsv);
END;