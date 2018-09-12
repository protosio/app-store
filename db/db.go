package db

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/protosio/app-store/util"
	// pq is required for sqlx to work even though it's not used directly
	"github.com/lib/pq"
)

var log = util.GetLogger()
var config = util.GetConfig()

func stripNilValues(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range in {
		if v != nil {
			out[k] = v
		}
	}
	return out
}

// Installer represents an installer as saved by the database
type Installer struct {
	Name        string
	Description string
	Thumbnail   string
	Provides    []string
	Versions    []string
}

// PGArrayToArray transforms a postgres string array to a Go string slice
func PGArrayToArray(pgarray string) []string {
	pgarray = strings.Replace(pgarray, "{", "", -1)
	pgarray = strings.Replace(pgarray, "}", "", -1)
	pgarray = strings.Replace(pgarray, "\"", "", -1)
	return strings.Split(pgarray, ",")
}

func dbConnectionString() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", config.DBHost, config.DBPort, config.DBName, config.DBUser, config.DBPass)
}

// Insert takes a db Installer and persists it to the database
func Insert(installer Installer) error {
	db, err := sqlx.Connect("postgres", dbConnectionString())
	if err != nil {
		return err
	}
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.
		Insert("installer").Columns("name", "description", "thumbnail", "provides", "versions").
		Values(installer.Name, installer.Description, installer.Thumbnail, pq.Array(installer.Provides), pq.Array(installer.Versions)).ToSql()
	if err != nil {
		return err
	}
	db.MustExec(sql, args...)
	return nil
}

// Update takes an Installer (db) and updates all the fields for that db entry
func Update(installer Installer) error {
	db, err := sqlx.Connect("postgres", dbConnectionString())
	if err != nil {
		return err
	}
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.Update("installer").SetMap(stripNilValues(map[string]interface{}{
		"description": installer.Description,
		"thumbnail":   installer.Thumbnail,
		"provides":    pq.Array(installer.Provides),
		"versions":    pq.Array(installer.Versions),
	})).Where("name = ?", installer.Name).ToSql()
	if err != nil {
		return err
	}
	log.Debugf("Performing update query: {%s} using arguments {%v}", sql, args)
	db.MustExec(sql, args...)
	return nil
}

// Get returns an Installer based on the provided name
func Get(name string) (Installer, bool) {
	db, err := sqlx.Connect("postgres", dbConnectionString())
	if err != nil {
		log.Fatalln(err)
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("*").From("installer").Where(sq.Eq{"name": name}).Limit(1).ToSql()
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Performing get query: {%s} using arguments {%v}", sql, args)
	installers := []Installer{}
	rows, err := db.Queryx(sql, args...)
	for rows.Next() {
		var installer Installer
		err := rows.Scan(&installer.Name, &installer.Description, &installer.Thumbnail, pq.Array(&installer.Provides), pq.Array(&installer.Versions))
		if err != nil {
			log.Error(err.Error())
			return Installer{}, false
		}
		installers = append(installers, installer)
	}

	if len(installers) < 1 {
		return Installer{}, false
	}
	return installers[0], true
}

// SearchProvider searches installers based on the provides field
func SearchProvider(providerType string) ([]Installer, error) {
	db, err := sqlx.Connect("postgres", dbConnectionString())
	if err != nil {
		return nil, err
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("name", "description", "thumbnail", "provides", "versions").From("installer").Where("array_length(provides, 1) > 0 AND (?) = ANY(provides)", providerType).ToSql()
	if err != nil {
		return nil, err
	}

	log.Debugf("Performing search query: {%s} using arguments {%v}", sql, args)
	installers := []Installer{}
	rows, err := db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var installer Installer
		err := rows.Scan(&installer.Name, &installer.Description, &installer.Thumbnail, pq.Array(&installer.Provides), pq.Array(&installer.Versions))
		if err != nil {
			return nil, err
		}
		installers = append(installers, installer)
	}

	return installers, nil
}

// Search searches installers using a full text search on name, description and provides field
func Search(searchTerm string) ([]Installer, error) {
	db, err := sqlx.Connect("postgres", dbConnectionString())
	if err != nil {
		return nil, err
	}

	sql := `SELECT name, description, thumbnail, provides, versions
FROM ( SELECT
        name, description, thumbnail, provides, versions,
        to_tsvector(name) ||
        to_tsvector(description) ||
        to_tsvector(array_to_string(provides, ' '))
       AS document
       FROM installer
       GROUP BY name) installer_search
WHERE installer_search.document @@ to_tsquery($1)
ORDER BY ts_rank(installer_search.document, to_tsquery($1)) DESC;`

	args := []interface{}{searchTerm}

	log.Debugf("Performing search query: {%s} using arguments {%v}", sql, args)
	installers := []Installer{}
	rows, err := db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var installer Installer
		err := rows.Scan(&installer.Name, &installer.Description, &installer.Thumbnail, pq.Array(&installer.Provides), pq.Array(&installer.Versions))
		if err != nil {
			return nil, err
		}
		installers = append(installers, installer)
	}

	return installers, nil
}
