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
		err := rows.Scan(&installer.Name, &installer.Thumbnail, &installer.VersionMetadata)
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

// Get returns an Installer based on the provided name
func Get(name string) (Installer, bool, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("name", "thumbnail", "version_metadata").From("installer").Where(sq.Eq{"name": name}).Limit(1).ToSql()
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
	sql, args, err := psql.Select("name", "thumbnail", "version_metadata").From("installer").ToSql()
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
	// psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	// sql, args, err := psql.Select("name", "thumbnail", "version_metadata").From("installer").Where("array_length(provides, 1) > 0 AND (?) = ANY(provides)", providerType).ToSql()
	// if err != nil {
	// 	return nil, err
	// }
	sql := `
select
    installer.name,
    installer.thumbnail,
    jsonb_object_agg(installer.key, installer.value) as version_metadata
from
    (
        select
            name,
            thumbnail,
            key,
            value,
            value -> 'provides' as provides
        from
            installer,
            jsonb_each(version_metadata)
        WHERE
            $1 = ANY(provides)
    ) installer
group by
    installer.name,
	installer.thumbnail;`
	args := []interface{}{providerType}

	installers, err := dbQuery(sql, args)
	if err != nil {
		return nil, err
	}

	return installers, nil
}

// Search searches installers using a full text search on name, description and provides field
func Search(searchTerm string) ([]Installer, error) {
	sql := `
SELECT name, thumbnail, version_metadata
FROM (
    SELECT
        name,
        thumbnail,
        version_metadata,
		to_tsvector(name) || to_tsvector('English', version_metadata::text)
	AS document
	FROM installer
	GROUP BY name
    ) installer_search
WHERE installer_search.document @@ to_tsquery($1)
ORDER BY ts_rank(installer_search.document, to_tsquery($1)) DESC;`

	args := []interface{}{searchTerm}

	installers, err := dbQuery(sql, args)
	if err != nil {
		return nil, err
	}

	return installers, nil
}
