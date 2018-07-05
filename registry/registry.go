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
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

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
	log.Infof("Processing push event for application %s with tag %s ", event.Target.Repository, event.Target.Tag)

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
	log.Info(manifest)

	url = fmt.Sprintf("http://docker-registry:5000/v2/%s/blobs/%s", event.Target.Repository, manifest.Config.Digest.String())
	log.Info(url)
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
		log.Printf("Error unmarshaling image inspect info: %v", err)
		return
	}
	log.Info(imageInfo.Config.Labels)
	metadata, err := daemon.GetMetadata(imageInfo.Config.Labels)
	if err != nil {
		log.Errorf("Could not parse metadata for installer %s(%s): %s", event.Target.Repository, event.Target.Tag, err.Error())
	}
	log.Println(metadata)
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