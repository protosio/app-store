package installer

import (
	"errors"
	"fmt"

	"encoding/json"

	"github.com/google/go-cmp/cmp"

	"github.com/protosio/app-store/db"
	"github.com/protosio/app-store/util"
)

var log = util.GetLogger()

// InstallerMetadata holds metadata for the installer
type InstallerMetadata struct {
	Params          []string            `json:"params"`
	Provides        []string            `json:"provides"`
	Requires        []string            `json:"requires"`
	PublicPorts     []util.Port         `json:"publicports"`
	Description     string              `json:"description"`
	PlatformID      string              `json:"platformid"`
	PlatformType    string              `json:"platformtype"`
	PersistancePath string              `json:"persistancepath"`
	Capabilities    []map[string]string `json:"capabilities"`
}

// Installer represents an application installer, but not a specific versio of it.
type Installer struct {
	ID              string                       `json:"id"`
	Name            string                       `json:"name,omitempty"`
	Thumbnail       string                       `json:"thumbnail,omitempty"`
	VersionMetadata map[string]InstallerMetadata `json:"versions"`
}

func dbToInstaller(dbinstaller db.Installer) (Installer, error) {
	installer := Installer{}
	installer.ID = dbinstaller.ID
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
		installer, err := dbToInstaller(dbinstaller)
		if err != nil {
			return installers, err
		}
		installers[installer.ID] = installer
	}
	return installers, nil
}

func installerToDB(installer Installer) (db.Installer, error) {
	dbInstaller := db.Installer{
		ID:        installer.ID,
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
func Add(name string, version string, metadata InstallerMetadata) error {
	id := util.String2SHA1(name)
	dbinstaller, found, err := db.Get(map[string]interface{}{"name": name})
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
			} else {
				log.Debugf("Detected new metadata for %s:%s. Updating in db", name, version)
				installer.VersionMetadata[version] = metadata
			}
		} else {
			log.Infof("Adding version %s for installer %s", version, name)
			installer.VersionMetadata[version] = metadata
		}

		if installer.ID != id {
			log.Infof("Installer id is different, updating to %s", id)
			installer.ID = id
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
		installer := Installer{ID: id, Name: name, VersionMetadata: map[string]InstallerMetadata{}}
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

// Get returns an installer based on its id
func Get(id string) (Installer, error) {
	dbinstaller, found, err := db.Get(map[string]interface{}{"id": id})
	if err != nil {
		return Installer{}, err
	} else if found {
		installer, err := dbToInstaller(dbinstaller)
		if err != nil {
			return Installer{}, err
		}
		return installer, nil
	}
	return Installer{}, fmt.Errorf("Could not find installer %s", id)
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
