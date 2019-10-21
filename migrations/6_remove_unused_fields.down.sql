BEGIN;
ALTER TABLE installer ADD COLUMN description TEXT NOT NULL DEFAULT 'n/a',
                      ADD COLUMN provides TEXT[] NOT NULL,
                      ADD COLUMN VERSIONS TEXT[];
END;