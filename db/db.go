package db

import (
	"log"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/protosio/app-store/installer"
)

var createDB = `
DROP DATABASE IF EXISTS installers;
CREATE DATABASE installers;
`
var createTable = `
CREATE TABLE installer (
	name        varchar(60) NOT NULL PRIMARY KEY,
	description text NOT NULL,
	thumbnail	text NOT NULL,
	provides	text[]
);
`

type dbInstaller struct {
	Name        string
	Description string
	Thumbnail   string
	Provides    string
}

func setupDB() {
	log.Println("Initializing database")
	db, err := sqlx.Connect("postgres", "host=cockroachdb port=26257 user=root sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(createDB)

	db, err = sqlx.Connect("postgres", "host=cockroachdb port=26257 dbname=installers  user=root sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(createTable)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.
		Insert("installer").Columns("name", "description", "thumbnail", "provides").
		Values("namecheap-dns", "DNS resource provider using the Namecheap API", "https://apps.protos.io/static/apps/23424234234.png", pq.Array([]string{"dns"})).
		Values("letsencrypt-cert", "Certificate resource provider using the Letsencrypt API", "https://apps.protos.io/static/apps/23424234234.png", pq.Array([]string{"certificate"})).ToSql()
	if err != nil {
		log.Fatalln(err)
	}
	tx := db.MustBegin()
	tx.MustExec(sql, args...)
	tx.Commit()

}

func SearchDB(providerType string) []installer.Installer {
	db, err := sqlx.Connect("postgres", "host=cockroachdb port=26257 dbname=installers  user=root sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("*").From("installer").Where("(?::string) = ANY(provides)", providerType).ToSql()
	if err != nil {
		log.Fatal(err)
	}

	dbinstallers := []dbInstaller{}
	err = db.Select(&dbinstallers, sql, args...)
	if err != nil {
		log.Fatal(err)
	}

	installers := []installer.Installer{}
	for _, dbinstaller := range dbinstallers {
		installer := installer.Installer{}
		installer.Name = dbinstaller.Name
		installer.Description = dbinstaller.Description
		installer.Thumbnail = dbinstaller.Thumbnail
		dbinstaller.Provides = strings.Replace(dbinstaller.Provides, "{", "", -1)
		dbinstaller.Provides = strings.Replace(dbinstaller.Provides, "}", "", -1)
		dbinstaller.Provides = strings.Replace(dbinstaller.Provides, "\"", "", -1)
		installer.Provides = strings.Split(dbinstaller.Provides, ",")
		installers = append(installers, installer)
	}
	return installers
}
