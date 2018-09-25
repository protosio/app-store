package installer

import (
	"errors"
	"fmt"

	"encoding/json"

	"github.com/google/go-cmp/cmp"

	"github.com/protosio/app-store/db"
	"github.com/protosio/app-store/util"
	"github.com/protosio/protos/daemon"
)

var log = util.GetLogger()

// Installer represents an application installer, but not a specific versio of it.
type Installer struct {
	Name            string                              `json:"name,omitempty"`
	Thumbnail       string                              `json:"thumbnail,omitempty"`
	VersionMetadata map[string]daemon.InstallerMetadata `json:"versions"`
}

func dbToInstaller(dbinstaller db.Installer) (Installer, error) {
	installer := Installer{}
	installer.Name = dbinstaller.Name
	installer.Thumbnail = dbinstaller.Thumbnail
	err := dbinstaller.VersionMetadata.Unmarshal(&installer.VersionMetadata)
	if err != nil {
		return installer, fmt.Errorf("Failed to JSON unmarshal metadata for %s: %v", installer.Name, err)
	}
	return installer, nil
}

func dbToInstallers(dbInstallers []db.Installer) (map[string]Installer, error) {
	installers := map[string]Installer{}
	for _, dbinstaller := range dbInstallers {
		installerID := util.String2SHA1(dbinstaller.Name)
		installer, err := dbToInstaller(dbinstaller)
		if err != nil {
			return installers, err
		}
		installers[installerID] = installer
	}
	return installers, nil
}

func installerToDB(installer Installer) (db.Installer, error) {
	dbInstaller := db.Installer{
		Name:      installer.Name,
		Thumbnail: installer.Thumbnail,
	}
	jsonMetadata, err := json.Marshal(installer.VersionMetadata)
	if err != nil {
		return dbInstaller, fmt.Errorf("Failed to JSON marshal metadata for %s: %v", installer.Name, err)
	}
	dbInstaller.VersionMetadata = jsonMetadata
	return dbInstaller, nil
}

// Add takes an installer and persists it to the database
func Add(name string, version string, metadata daemon.InstallerMetadata) error {
	dbinstaller, found, err := db.Get(name)
	if err != nil {
		return err
	} else if found {
		installer, err := dbToInstaller(dbinstaller)
		if err != nil {
			return err
		}

		if oldMetadata, ok := installer.VersionMetadata[version]; ok {
			log.Debugf("Version %s for installer %s already in db", version, name)
			if cmp.Equal(oldMetadata, metadata) {
				log.Debugf("No new metadata detected for %s:%s", name, version)
				return nil
			}

			log.Debugf("Detected new metadata for %s:%s. Updating in db", name, version)
			installer.VersionMetadata[version] = metadata
		} else {
			log.Infof("Adding version %s for installer %s", version, name)
			installer.VersionMetadata[version] = metadata
		}

		dbinstaller, err = installerToDB(installer)
		if err != nil {
			return err
		}
		err = db.Update(dbinstaller)
		if err != nil {
			return err
		}
	} else {
		log.Infof("Installer %s not found. Adding it to the database", name)
		installer := Installer{Name: name}
		installer.VersionMetadata[version] = metadata
		dbinstaller, err := installerToDB(installer)
		if err != nil {
			return err
		}
		log.Debugf("Adding installer %v", dbinstaller)
		err = db.Insert(dbinstaller)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAll returns all available installers
func GetAll() (map[string]Installer, error) {
	installers := map[string]Installer{}
	dbinstallers, err := db.GetAll()
	if err != nil {
		return installers, err
	}
	return dbToInstallers(dbinstallers)
}

// Search searches the database for all the installers that match the provides field
func Search(providerType string, general string) (map[string]Installer, error) {
	var installers map[string]Installer
	var dbinstallers []db.Installer
	var err error
	if providerType != "" {
		dbinstallers, err = db.SearchProvider(providerType)
		if err != nil {
			return installers, err
		}
	} else if general != "" {
		dbinstallers, err = db.Search(general)
		if err != nil {
			return installers, err
		}
	} else {
		return installers, errors.New("Either the provider type or the general search string needs to be provided")
	}

	installers, err = dbToInstallers(dbinstallers)
	if err != nil {
		return installers, err
	}

	return installers, nil
}
