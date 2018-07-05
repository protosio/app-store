package installer

import "github.com/nustiueudinastea/protos/daemon"

// Installer represents an application installer, but not a specific version of it.
type Installer struct {
	daemon.InstallerMetadata
	Name      string `json:"name,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

func Save() {

}
