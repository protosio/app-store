package db

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	sqlxTypes "github.com/jmoiron/sqlx/types"
	"github.com/protosio/app-store/util"

	// pq is required for sqlx to work even though it's not used directly
	_ "github.com/lib/pq"
)

var log = util.GetLogger()
var config = util.GetConfig()
var db *sqlx.DB

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
	ID              string
	Name            string
	Thumbnail       string
	VersionMetadata sqlxTypes.JSONText
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

// Connect connects to the databse at program start
func Connect() error {
	var err error
	log.Debugf("Connecting to the db using host: %s port: %d", config.DBHost, config.DBPort)
	db, err = sqlx.Connect("postgres", dbConnectionString())
	if err != nil {
		return err
	}
	return nil
}

func dbQuery(sql string, args []interface{}) ([]Installer, error) {
	log.Debugf("Performing search query: {%s} using arguments {%v}", sql, args)
	installers := []Installer{}
	rows, err := db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var installer Installer
		err := rows.Scan(&installer.ID, &installer.Name, &installer.Thumbnail, &installer.VersionMetadata)
		if err != nil {
			return nil, err
		}
		installers = append(installers, installer)
	}
	log.Debugf("Query returned: %v", installers)
	return installers, nil
}

// Insert takes a db Installer and persists it to the database
func Insert(installer Installer) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.
		Insert("installer").Columns("name", "thumbnail", "version_metadata").
		Values(installer.Name, installer.Thumbnail, installer.VersionMetadata).ToSql()
	if err != nil {
		return err
	}
	db.MustExec(sql, args...)
	return nil
}

// Update takes an Installer (db) and updates all the fields for that db entry
func Update(installer Installer) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.Update("installer").SetMap(stripNilValues(map[string]interface{}{
		"id":               installer.ID,
		"thumbnail":        installer.Thumbnail,
		"version_metadata": installer.VersionMetadata,
	})).Where("name = ?", installer.Name).ToSql()
	if err != nil {
		return err
	}
	log.Debugf("Performing update query: {%s} using arguments {%v}", sql, args)
	db.MustExec(sql, args...)
	return nil
}

// Get returns an Installer based on the provided filter
func Get(filter map[string]interface{}) (Installer, bool, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "name", "thumbnail", "version_metadata").From("installer").Where(filter).Limit(1).ToSql()
	if err != nil {
		log.Fatal(err)
	}

	installers, err := dbQuery(sql, args)
	if err != nil {
		log.Errorf("Error while performing get query: %s", err.Error())
		return Installer{}, false, err
	}

	if len(installers) < 1 {
		return Installer{}, false, nil
	}
	return installers[0], true, nil
}

// GetAll retrieves all installers from the database
func GetAll() ([]Installer, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "name", "thumbnail", "version_metadata").From("installer").ToSql()
	if err != nil {
		return nil, err
	}

	installers, err := dbQuery(sql, args)
	if err != nil {
		return nil, err
	}

	return installers, nil
}

// SearchProvider searches installers based on the provides field
func SearchProvider(providerType string) ([]Installer, error) {

	sql := `
SELECT
	installer.id,
    installer.name,
    installer.thumbnail,
    jsonb_object_agg(installer.key, installer.value) AS version_metadata
FROM
	(SELECT
		id,
		name,
		thumbnail,
		key,
		VALUE
	FROM
		installer,
		jsonb_each(version_metadata)
	WHERE
		VALUE -> 'provides' @> ANY (ARRAY [$1]::jsonb[])) installer
GROUP BY
	installer.id,
    installer.name,
	installer.thumbnail;`
	// sorounding the search term in quotes is required for the pq jsonb search
	param := "\"" + providerType + "\""
	args := []interface{}{param}

	installers, err := dbQuery(sql, args)
	if err != nil {
		return nil, err
	}

	return installers, nil
}

// Search searches installers using a full text search on name, description and provides field
func Search(searchTerm string) ([]Installer, error) {
	sql := `
SELECT
	installer.id,
	installer.name,
	installer.thumbnail,
	jsonb_object_agg(installer.key, installer.value) AS version_metadata
FROM
	(SELECT
		id,
		name,
		thumbnail,
		key,
		VALUE,
		to_tsvector(name) || to_tsvector('English', value::text) AS tsvdata
FROM installer,
     jsonb_each(version_metadata)) installer
WHERE installer.tsvdata @@ to_tsquery($1)
GROUP BY installer.id, installer.name, installer.thumbnail;`

	args := []interface{}{searchTerm}

	installers, err := dbQuery(sql, args)
	if err != nil {
		return nil, err
	}

	return installers, nil
}
