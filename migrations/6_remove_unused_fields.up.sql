BEGIN;
ALTER TABLE installer DROP COLUMN description,
                      DROP COLUMN provides,
                      DROP COLUMN versions;
END;