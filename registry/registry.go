package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/api/types"
	"github.com/nustiueudinastea/protos/daemon"
	"github.com/protosio/app-store/installer"
	"github.com/protosio/app-store/util"
)

var log = util.GetLogger()

type Target struct {
	MediaType  string `json:"mediaType"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	Repository string `json:"repository"`
	URL        string `json:"url"`
	Tag        string `json:"tag"`
}

type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    Target    `json:"target"`
}

type Events struct {
	Events []Event `json:"events"`
}

func processPushEvent(event Event) {
	if event.Target.Tag == "" {
		log.Errorf("Push event for application %s does not containg a tag. Ignoring", event.Target.Repository)
	}
	log.Infof("Processing push event for application %s with tag %s", event.Target.Repository, event.Target.Tag)

	url := fmt.Sprintf("http://docker-registry:5000/v2/%s/manifests/%s", event.Target.Repository, event.Target.Digest)
	r, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	defer r.Body.Close()

	bodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	var manifest schema2.Manifest
	err = json.Unmarshal(bodyJSON, &manifest)
	if err != nil {
		log.Printf("Error unmarshaling image manifest: %v", err)
		return
	}

	url = fmt.Sprintf("http://docker-registry:5000/v2/%s/blobs/%s", event.Target.Repository, manifest.Config.Digest.String())
	r, err = http.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	defer r.Body.Close()

	bodyJSON, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	var imageInfo types.ImageInspect
	err = json.Unmarshal(bodyJSON, &imageInfo)
	if err != nil {
		log.Errorf("Error unmarshaling image inspect info: %v", err)
		return
	}
	metadata, err := daemon.GetMetadata(imageInfo.Config.Labels)
	if err != nil {
		log.Errorf("Could not parse metadata for installer %s(%s): %s", event.Target.Repository, event.Target.Tag, err.Error())
		return
	}
	err = installer.Add(event.Target.Repository, event.Target.Tag, metadata)
	if err != nil {
		log.Errorf("Could not save installer %s(%s): %s", event.Target.Repository, event.Target.Tag, err.Error())
		return
	}
}

// ProcessEvents takes an events array and process all the events of type "push"
func ProcessEvents(events []Event) {
	for _, event := range events {
		if event.Action != "push" {
			log.Debug("Ignoring event of type " + event.Action)
			continue
		}
		processPushEvent(event)
	}
}
