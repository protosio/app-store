BEGIN;
UPDATE installer SET tsv =
    setweight(to_tsvector(name), 'A') ||
    setweight(to_tsvector('English', version_metadata::text), 'B');
DROP INDEX ix_installer_tsv;
CREATE INDEX ix_installer_tsv ON installer USING GIN(tsv);
END;