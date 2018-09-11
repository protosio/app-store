BEGIN;
CREATE TABLE installer (
	name        varchar(60) NOT NULL PRIMARY KEY,
	description text NOT NULL,
	thumbnail	text NOT NULL,
	provides	text[],
	versions	text[] NOT NULL
);
END;