package installer

import (
	"errors"

	"github.com/nustiueudinastea/protos/daemon"
	"github.com/protosio/app-store/db"
	"github.com/protosio/app-store/util"
)

var log = util.GetLogger()

// Installer represents an application installer, but not a specific version of it.
type Installer struct {
	daemon.InstallerMetadata
	Name      string   `json:"name,omitempty"`
	Versions  []string `json:"versions"`
	Thumbnail string   `json:"thumbnail,omitempty"`
}

func dbToInstaller(dbinstaller db.Installer) Installer {
	installer := Installer{}
	installer.Name = dbinstaller.Name
	installer.Description = dbinstaller.Description
	installer.Thumbnail = dbinstaller.Thumbnail
	installer.Provides = dbinstaller.Provides
	installer.Versions = dbinstaller.Versions
	return installer
}

func installerToDB(installer Installer) db.Installer {
	return db.Installer{
		Name:        installer.Name,
		Description: installer.Description,
		Thumbnail:   installer.Thumbnail,
		Provides:    installer.Provides,
		Versions:    installer.Versions,
	}
}

// Add takes an installer and persists it to the database
func Add(name string, version string, metadata daemon.InstallerMetadata) error {
	dbinstaller, found := db.Get(name)
	if found {
		log.Infof("Updating installer %s", name)
		dbinstaller.Description = metadata.Description
		if ok, _ := util.StringInSlice(version, dbinstaller.Versions); ok == false {
			dbinstaller.Versions = append(dbinstaller.Versions, version)
		}
		dbinstaller.Provides = metadata.Provides
		log.Debugf("Updating installer %v", dbinstaller)
		err := db.Update(dbinstaller)
		if err != nil {
			return err
		}
	} else {
		log.Infof("Installer %s not found. Adding it to the database", name)
		installer := Installer{
			Name:              name,
			Versions:          []string{version},
			InstallerMetadata: metadata,
		}
		dbinstaller = installerToDB(installer)
		log.Debugf("Adding installer %v", dbinstaller)
		err := db.Insert(dbinstaller)
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
	for _, installer := range dbinstallers {
		installerID := util.String2SHA1(installer.Name)
		installers[installerID] = dbToInstaller(installer)
	}
	return installers, err
}

// Search searches the database for all the installers that match the provides field
func Search(providerType string, general string) (map[string]Installer, error) {
	installers := map[string]Installer{}
	if providerType != "" {
		dbinstallers, err := db.SearchProvider(providerType)
		if err != nil {
			return installers, err
		}
		for _, installer := range dbinstallers {
			installerID := util.String2SHA1(installer.Name)
			installers[installerID] = dbToInstaller(installer)
		}
	} else if general != "" {
		dbinstallers, err := db.Search(general)
		if err != nil {
			return installers, err
		}
		for _, installer := range dbinstallers {
			installerID := util.String2SHA1(installer.Name)
			installers[installerID] = dbToInstaller(installer)
		}
	} else {
		return installers, errors.New("Either the provider type or the general search string needs to be provided")
	}

	return installers, nil
}
