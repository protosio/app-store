BEGIN;
ALTER TABLE installer ADD COLUMN description TEXT NOT NULL DEFAULT 'n/a';
END;